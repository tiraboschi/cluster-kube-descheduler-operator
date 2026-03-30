package e2e

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	o "github.com/onsi/gomega"

	descv1 "github.com/openshift/cluster-kube-descheduler-operator/pkg/apis/descheduler/v1"
	deschclient "github.com/openshift/cluster-kube-descheduler-operator/pkg/generated/clientset/versioned"
	ssscheme "github.com/openshift/cluster-kube-descheduler-operator/pkg/generated/clientset/versioned/scheme"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"

	"github.com/openshift/cluster-kube-descheduler-operator/pkg/operator/operatorclient"
	"github.com/openshift/cluster-kube-descheduler-operator/test/e2e/bindata"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resource/resourceread"
)

const (
	baseConf                       = "base"
	kubeVirtRelieveAndMigrateConf  = "devKubeVirtRelieveAndMigrate"
	kubeVirtLabelKey               = "kubevirt.io/schedulable"
	kubeVirtLabelValue             = "true"
	workersLabelSelector           = "node-role.kubernetes.io/worker="
	EXPERIMENTAL_DISABLE_PSI_CHECK = "EXPERIMENTAL_DISABLE_PSI_CHECK"
)

var operatorConfigsAppliers = map[string]func(context.Context, *deschclient.Clientset) error{
	baseConf:                      operatorConfigsApplier("assets/07_descheduler-operator.cr.yaml"),
	kubeVirtRelieveAndMigrateConf: operatorConfigsApplier("assets/07_descheduler-operator.cr.devKubeVirtRelieveAndMigrate.yaml"),
}

func operatorConfigsApplier(path string) func(context.Context, *deschclient.Clientset) error {
	return func(ctx context.Context, deschClient *deschclient.Clientset) error {
		requiredObj, err := runtime.Decode(ssscheme.Codecs.UniversalDecoder(descv1.SchemeGroupVersion), bindata.MustAsset(path))
		if err != nil {
			klog.Errorf("Unable to decode %v: %v", path, err)
			return err
		}
		requiredDesch := requiredObj.(*descv1.KubeDescheduler)
		existingDesch, err := deschClient.KubedeschedulersV1().KubeDeschedulers(requiredDesch.Namespace).Get(ctx, requiredDesch.Name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				_, err = deschClient.KubedeschedulersV1().KubeDeschedulers(requiredDesch.Namespace).Create(ctx, requiredDesch, metav1.CreateOptions{})
				return err
			} else {
				return err
			}
		}
		requiredDesch.Spec.DeepCopyInto(&existingDesch.Spec)
		existingDesch.ObjectMeta.Annotations = requiredDesch.ObjectMeta.Annotations
		existingDesch.ObjectMeta.Labels = requiredDesch.ObjectMeta.Labels
		_, err = deschClient.KubedeschedulersV1().KubeDeschedulers(requiredDesch.Namespace).Update(ctx, existingDesch, metav1.UpdateOptions{})
		// retry once on conflicts
		if apierrors.IsConflict(err) {
			existingDesch, err = deschClient.KubedeschedulersV1().KubeDeschedulers(requiredDesch.Namespace).Get(ctx, requiredDesch.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			requiredDesch.Spec.DeepCopyInto(&existingDesch.Spec)
			existingDesch.ObjectMeta.Annotations = requiredDesch.ObjectMeta.Annotations
			existingDesch.ObjectMeta.Labels = requiredDesch.ObjectMeta.Labels
			_, err = deschClient.KubedeschedulersV1().KubeDeschedulers(requiredDesch.Namespace).Update(ctx, existingDesch, metav1.UpdateOptions{})
		}
		return err
	}
}

// setupOperator sets up the operator and waits for it to be ready.
func setupOperator(t testing.TB) (context.Context, context.CancelFunc, *k8sclient.Clientset, error) {
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

// cleanupTestNamespace deletes the test namespace.
func cleanupTestNamespace(t testing.TB, ctx context.Context, kubeClient *k8sclient.Clientset, testNamespace string) {
	if testNamespace == "" {
		return
	}
	klog.Infof("Cleaning up test namespace: %s", testNamespace)
	if err := kubeClient.CoreV1().Namespaces().Delete(ctx, testNamespace, metav1.DeleteOptions{}); err != nil {
		t.Fatalf("Failed to delete namespace %s: %v", testNamespace, err)
	}
	o.Eventually(func() bool {
		_, err := kubeClient.CoreV1().Namespaces().Get(ctx, testNamespace, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true
		}
		return false
	}, time.Minute, 1*time.Second).Should(o.BeTrue(), "namespace not deleted after timeout")
}

func waitForPodRunningByNamePrefix(ctx context.Context, kubeClient *k8sclient.Clientset, namespace, nameprefix, excludedprefix string) (*v1.Pod, error) {
	var expectedPod *corev1.Pod
	// Wait until the expected pod is running
	o.Eventually(func() bool {
		klog.Infof("Listing pods...")
		podItems, err := kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			klog.Errorf("Unable to list pods: %v", err)
			return false
		}
		for _, pod := range podItems.Items {
			if !strings.HasPrefix(pod.Name, nameprefix) || (excludedprefix != "" && strings.HasPrefix(pod.Name, excludedprefix)) {
				continue
			}
			klog.Infof("Checking pod: %v, phase: %v, deletionTS: %v\n", pod.Name, pod.Status.Phase, pod.GetDeletionTimestamp())
			if pod.Status.Phase == corev1.PodRunning && pod.GetDeletionTimestamp() == nil {
				expectedPod = pod.DeepCopy()
				return true
			}
		}
		return false
	}).WithTimeout(1 * time.Minute).WithPolling(5 * time.Second).Should(o.BeTrue())
	return expectedPod, nil
}

func applyKubeVirtNodeLabel(ctx context.Context, kubeClient *k8sclient.Clientset) error {
	nodes, err := kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: workersLabelSelector})
	if err != nil {
		return err
	}
	for _, node := range nodes.Items {
		if node.Labels[kubeVirtLabelKey] != kubeVirtLabelValue {
			node.Labels[kubeVirtLabelKey] = kubeVirtLabelValue
			uerr := updateNodeAndRetryOnConflicts(ctx, kubeClient, &node, metav1.UpdateOptions{})
			if uerr != nil {
				return uerr
			}
		}
	}
	return nil
}

func dropKubeVirtNodeLabel(ctx context.Context, kubeClient *k8sclient.Clientset) error {
	nodes, err := kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: workersLabelSelector})
	if err != nil {
		return err
	}
	for _, node := range nodes.Items {
		if node.Labels[kubeVirtLabelKey] == kubeVirtLabelValue {
			delete(node.Labels, kubeVirtLabelKey)
			uerr := updateNodeAndRetryOnConflicts(ctx, kubeClient, &node, metav1.UpdateOptions{})
			if uerr != nil {
				return uerr
			}
		}
	}
	return nil
}

func updateNodeAndRetryOnConflicts(ctx context.Context, kubeClient *k8sclient.Clientset, node *corev1.Node, opts metav1.UpdateOptions) error {
	uNode, uerr := kubeClient.CoreV1().Nodes().Update(ctx, node, opts)
	if uerr != nil {
		if apierrors.IsConflict(uerr) {
			if uNode.Name == "" {
				uNode, uerr = kubeClient.CoreV1().Nodes().Get(ctx, node.Name, metav1.GetOptions{})
				if uerr != nil {
					return uerr
				}
			}
			node.Spec.DeepCopyInto(&uNode.Spec)
			uNode.ObjectMeta.Labels = node.ObjectMeta.Labels
			uNode.ObjectMeta.Annotations = node.ObjectMeta.Annotations
			_, err := kubeClient.CoreV1().Nodes().Update(ctx, uNode, opts)
			return err
		}
		return uerr
	}
	return nil
}

func updateDeploymentAndRetryOnConflicts(ctx context.Context, kubeClient *k8sclient.Clientset, deployment *appsv1.Deployment, opts metav1.UpdateOptions) error {
	uDeployment, uerr := kubeClient.AppsV1().Deployments(deployment.Namespace).Update(ctx, deployment, opts)
	if uerr != nil {
		if apierrors.IsConflict(uerr) {
			if uDeployment.Name == "" {
				uDeployment, uerr = kubeClient.AppsV1().Deployments(uDeployment.Namespace).Get(ctx, uDeployment.Name, metav1.GetOptions{})
				if uerr != nil {
					return uerr
				}
			}
			deployment.Spec.DeepCopyInto(&uDeployment.Spec)
			uDeployment.ObjectMeta.Labels = deployment.ObjectMeta.Labels
			uDeployment.ObjectMeta.Annotations = deployment.ObjectMeta.Annotations
			_, err := kubeClient.AppsV1().Deployments(deployment.Namespace).Update(ctx, deployment, opts)
			return err
		}
		return uerr
	}
	return nil
}

func mockPSIEnv(ctx context.Context, kubeClient *k8sclient.Clientset) error {
	operatorDeployment, err := kubeClient.AppsV1().Deployments(operatorclient.OperatorNamespace).Get(ctx, operatorclient.OperandName+"-operator", metav1.GetOptions{})
	if err != nil {
		return err
	}
	operatorDeployment.Spec.Template.Spec.Containers[0].Env = append(
		operatorDeployment.Spec.Template.Spec.Containers[0].Env,
		v1.EnvVar{
			Name:  EXPERIMENTAL_DISABLE_PSI_CHECK,
			Value: "true",
		})
	return updateDeploymentAndRetryOnConflicts(ctx, kubeClient, operatorDeployment, metav1.UpdateOptions{})
}

func unmockPSIEnv(ctx context.Context, kubeClient *k8sclient.Clientset, prevDisablePSIcheck string, foundDisablePSIcheck bool) error {
	operatorDeployment, err := kubeClient.AppsV1().Deployments(operatorclient.OperatorNamespace).Get(ctx, operatorclient.OperandName+"-operator", metav1.GetOptions{})
	if err != nil {
		return err
	}
	var envVars []v1.EnvVar
	for _, e := range operatorDeployment.Spec.Template.Spec.Containers[0].Env {
		if e.Name != EXPERIMENTAL_DISABLE_PSI_CHECK {
			envVars = append(envVars, e)
		}
	}
	if foundDisablePSIcheck {
		operatorDeployment.Spec.Template.Spec.Containers[0].Env = append(
			envVars,
			v1.EnvVar{
				Name:  EXPERIMENTAL_DISABLE_PSI_CHECK,
				Value: prevDisablePSIcheck,
			})
	}
	operatorDeployment.Spec.Template.Spec.Containers[0].Env = envVars
	return updateDeploymentAndRetryOnConflicts(ctx, kubeClient, operatorDeployment, metav1.UpdateOptions{})
}

// setupSoftTainterController sets up the common prerequisites for soft tainter tests.
// It collects cleanup functions as it progresses and runs them if any step fails.
// Returns a cleanup function that should be deferred by the caller.
func setupSoftTainterController(ctx context.Context, t testing.TB, kubeClient *k8sclient.Clientset, deschClient *deschclient.Clientset) func() {
	var cleanups []func()

	// Helper to run all cleanups in reverse order (LIFO, like defer)
	runCleanups := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}

	// label all the nodes to mock a KubeVirt deployment
	if err := applyKubeVirtNodeLabel(ctx, kubeClient); err != nil {
		t.Fatalf("Failed applying KubeVirt node label: %v", err)
	}
	cleanups = append(cleanups, func() {
		if err := dropKubeVirtNodeLabel(ctx, kubeClient); err != nil {
			t.Fatalf("Failed reverting KubeVirt node label: %v", err)
		}
	})

	// patch the operator deployment to mock PSI
	prevDisablePSIcheck, foundDisablePSIcheck := os.LookupEnv(EXPERIMENTAL_DISABLE_PSI_CHECK)
	if err := mockPSIEnv(ctx, kubeClient); err != nil {
		runCleanups()
		t.Fatalf("Failed mocking PSI path enviromental variable to the operator depoyment: %v", err)
	}
	cleanups = append(cleanups, func() {
		if err := unmockPSIEnv(ctx, kubeClient, prevDisablePSIcheck, foundDisablePSIcheck); err != nil {
			t.Fatalf("Failed PSI path enviromental variable: %v", err)
		}
	})

	// wait for descheduler operator pod to be running
	deschOpPod, err := waitForPodRunningByNamePrefix(ctx, kubeClient, operatorclient.OperatorNamespace, operatorclient.OperandName, operatorclient.OperandName+"-operator")
	if err != nil {
		runCleanups()
		t.Fatalf("Unable to wait for the Descheduler operator pod to run")
	}
	klog.Infof("Descheduler pod running in %v", deschOpPod.Name)

	// apply devKubeVirtRelieveAndMigrate CR for the operator
	if err := operatorConfigsAppliers[kubeVirtRelieveAndMigrateConf](ctx, deschClient); err != nil {
		runCleanups()
		t.Fatalf("Unable to apply a CR for Descheduler operator: %v", err)
	}
	klog.Infof("Descheduler operator is now configured with devKubeVirtRelieveAndMigrate profile")
	cleanups = append(cleanups, func() {
		if err := operatorConfigsAppliers[baseConf](ctx, deschClient); err != nil {
			t.Fatalf("Failed restoring base profile: %v", err)
		}
	})

	// wait for softtainter pod to be running
	stPod, err := waitForPodRunningByNamePrefix(ctx, kubeClient, operatorclient.OperatorNamespace, operatorclient.SoftTainterOperandName, "")
	if err != nil {
		runCleanups()
		t.Fatalf("Unable to wait for the softtainter pod to run")
	}
	klog.Infof("SoftTainter pod running in %v", stPod.Name)

	// Return cleanup function that runs all cleanups in reverse order
	return runCleanups
}
