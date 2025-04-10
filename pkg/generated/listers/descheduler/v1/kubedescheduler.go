// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	deschedulerv1 "github.com/openshift/cluster-kube-descheduler-operator/pkg/apis/descheduler/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
)

// KubeDeschedulerLister helps list KubeDeschedulers.
// All objects returned here must be treated as read-only.
type KubeDeschedulerLister interface {
	// List lists all KubeDeschedulers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*deschedulerv1.KubeDescheduler, err error)
	// KubeDeschedulers returns an object that can list and get KubeDeschedulers.
	KubeDeschedulers(namespace string) KubeDeschedulerNamespaceLister
	KubeDeschedulerListerExpansion
}

// kubeDeschedulerLister implements the KubeDeschedulerLister interface.
type kubeDeschedulerLister struct {
	listers.ResourceIndexer[*deschedulerv1.KubeDescheduler]
}

// NewKubeDeschedulerLister returns a new KubeDeschedulerLister.
func NewKubeDeschedulerLister(indexer cache.Indexer) KubeDeschedulerLister {
	return &kubeDeschedulerLister{listers.New[*deschedulerv1.KubeDescheduler](indexer, deschedulerv1.Resource("kubedescheduler"))}
}

// KubeDeschedulers returns an object that can list and get KubeDeschedulers.
func (s *kubeDeschedulerLister) KubeDeschedulers(namespace string) KubeDeschedulerNamespaceLister {
	return kubeDeschedulerNamespaceLister{listers.NewNamespaced[*deschedulerv1.KubeDescheduler](s.ResourceIndexer, namespace)}
}

// KubeDeschedulerNamespaceLister helps list and get KubeDeschedulers.
// All objects returned here must be treated as read-only.
type KubeDeschedulerNamespaceLister interface {
	// List lists all KubeDeschedulers in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*deschedulerv1.KubeDescheduler, err error)
	// Get retrieves the KubeDescheduler from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*deschedulerv1.KubeDescheduler, error)
	KubeDeschedulerNamespaceListerExpansion
}

// kubeDeschedulerNamespaceLister implements the KubeDeschedulerNamespaceLister
// interface.
type kubeDeschedulerNamespaceLister struct {
	listers.ResourceIndexer[*deschedulerv1.KubeDescheduler]
}
