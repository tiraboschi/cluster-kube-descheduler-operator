package e2e

import (
	"context"
	"os"
	"strings"
	"time"

	o "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"

	"github.com/openshift/cluster-kube-descheduler-operator/pkg/operator/operatorclient"
	"github.com/openshift/cluster-kube-descheduler-operator/test/e2e/bindata"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"
)

// setupOperator sets up the operator and waits for it to be ready.
func setupOperator() (context.Context, context.CancelFunc, *k8sclient.Clientset, error) {
	if os.Getenv("KUBECONFIG") == "" {
		klog.Errorf("KUBECONFIG environment variable not set")
		os.Exit(1)
	}

	if os.Getenv("OPERATOR_IMAGE") == "" && os.Getenv("OPERAND_IMAGE") == "" {
		if os.Getenv("RELEASE_IMAGE_LATEST") == "" {
			klog.Errorf("RELEASE_IMAGE_LATEST environment variable not set")
			os.Exit(1)
		}

		if os.Getenv("NAMESPACE") == "" {
			klog.Errorf("NAMESPACE environment variable not set")
			os.Exit(1)
		}
	}

	kubeClient := GetKubeClient()
	apiExtClient := GetApiExtensionClient()
	deschClient := GetDeschedulerClient()

	eventRecorder := events.NewKubeRecorder(kubeClient.CoreV1().Events("default"), "test-e2e", &corev1.ObjectReference{}, clock.RealClock{})

	ctx, cancelFnc := context.WithCancel(context.TODO())

	assets := []struct {
		path           string
		readerAndApply func(objBytes []byte) error
	}{
		{
			path: "assets/00_kube-descheduler-operator-crd.yaml",
			readerAndApply: func(objBytes []byte) error {
				_, _, err := resourceapply.ApplyCustomResourceDefinitionV1(ctx, apiExtClient.ApiextensionsV1(), eventRecorder, resourceread.ReadCustomResourceDefinitionV1OrDie(objBytes))
				return err
			},
		},
		{
			path: "assets/01_namespace.yaml",
			readerAndApply: func(objBytes []byte) error {
				_, _, err := resourceapply.ApplyNamespace(ctx, kubeClient.CoreV1(), eventRecorder, resourceread.ReadNamespaceV1OrDie(objBytes))
				return err
			},
		},
		{
			path: "assets/02_serviceaccount.yaml",
			readerAndApply: func(objBytes []byte) error {
				_, _, err := resourceapply.ApplyServiceAccount(ctx, kubeClient.CoreV1(), eventRecorder, resourceread.ReadServiceAccountV1OrDie(objBytes))
				return err
			},
		},
		{
			path: "assets/03_clusterrole.yaml",
			readerAndApply: func(objBytes []byte) error {
				_, _, err := resourceapply.ApplyClusterRole(ctx, kubeClient.RbacV1(), eventRecorder, resourceread.ReadClusterRoleV1OrDie(objBytes))
				return err
			},
		},
		{
			path: "assets/04_clusterrolebinding.yaml",
			readerAndApply: func(objBytes []byte) error {
				_, _, err := resourceapply.ApplyClusterRoleBinding(ctx, kubeClient.RbacV1(), eventRecorder, resourceread.ReadClusterRoleBindingV1OrDie(objBytes))
				return err
			},
		},
		{
			path: "assets/05_deployment.yaml",
			readerAndApply: func(objBytes []byte) error {
				required := resourceread.ReadDeploymentV1OrDie(objBytes)
				// override the operator image with the one built in the CI

				var operator_image, operand_image string

				if os.Getenv("OPERATOR_IMAGE") != "" {
					operator_image = os.Getenv("OPERATOR_IMAGE")
				} else {
					// E.g. RELEASE_IMAGE_LATEST=registry.build03.ci.openshift.org/ci-op-52fj47p4/stable:${component}
					registry := strings.Split(os.Getenv("RELEASE_IMAGE_LATEST"), "/")[0]
					operator_image = registry + "/" + os.Getenv("NAMESPACE") + "/" + "pipeline:cluster-kube-descheduler-operator"
				}

				if os.Getenv("OPERAND_IMAGE") != "" {
					operand_image = os.Getenv("OPERAND_IMAGE")
				} else {
					operand_image = "quay.io/jchaloup/descheduler:v5.3.2-0"
				}

				required.Spec.Template.Spec.Containers[0].Image = operator_image
				// OPERAND_IMAGE env
				for i, env := range required.Spec.Template.Spec.Containers[0].Env {
					if env.Name == "RELATED_IMAGE_OPERAND_IMAGE" {
						required.Spec.Template.Spec.Containers[0].Env[i].Value = operand_image
					} else if env.Name == "RELATED_IMAGE_SOFTTAINTER_IMAGE" {
						required.Spec.Template.Spec.Containers[0].Env[i].Value = operator_image
					}
				}
				_, _, err := resourceapply.ApplyDeployment(
					ctx,
					kubeClient.AppsV1(),
					eventRecorder,
					required,
					1000, // any random high number
				)
				return err
			},
		},
		{
			path: "assets/06_configmap.yaml",
			readerAndApply: func(objBytes []byte) error {
				_, _, err := resourceapply.ApplyConfigMap(ctx, kubeClient.CoreV1(), eventRecorder, resourceread.ReadConfigMapV1OrDie(objBytes))
				return err
			},
		},
	}

	// create required resources, e.g. namespace, crd, roles
	o.Eventually(func() bool {
		for _, asset := range assets {
			klog.Infof("Creating %v", asset.path)
			if err := asset.readerAndApply(bindata.MustAsset(asset.path)); err != nil {
				klog.Errorf("Unable to create %v: %v", asset.path, err)
				return false
			}
		}
		return true
	}).WithTimeout(10*time.Second).WithPolling(1*time.Second).Should(o.BeTrue(), "Unable to create Descheduler operator resources")

	// apply base CR for the operator
	err := operatorConfigsAppliers[baseConf](ctx, deschClient)
	if err != nil {
		klog.Errorf("Unable to apply a CR for Descheduler operator: %v", err)
		os.Exit(1)
	}

	// wait for descheduler operator pod to be running
	deschOpPod, err := waitForPodRunningByNamePrefix(ctx, kubeClient, operatorclient.OperatorNamespace, operatorclient.OperandName+"-operator", "")
	if err != nil {
		klog.Errorf("Unable to wait for the Descheduler operator pod to run")
		os.Exit(1)
	}
	klog.Infof("Descheduler operator pod running in %v", deschOpPod.Name)

	// wait for descheduler pod to be running
	deschPod, err := waitForPodRunningByNamePrefix(ctx, kubeClient, operatorclient.OperatorNamespace, operatorclient.OperandName, operatorclient.OperandName+"-operator")
	if err != nil {
		klog.Errorf("Unable to wait for the Descheduler pod to run")
		os.Exit(1)
	}
	klog.Infof("Descheduler (operand) pod running in %v", deschPod.Name)

	return ctx, cancelFnc, kubeClient, nil
}
