package informers

import (
	"k8s.io/client-go/kubernetes"
	"time"
	k8sinformers "k8s.io/client-go/informers"
)

const defaultResync = 600 * time.Second

// InformerFactory is a group all shared informer factories which aiscope needed
// callers should check if the return value is nil
type InformerFactory interface {
	KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory

	// Start shared informer factory one by one if they are not nil
	Start(stopCh <-chan struct{})
}

type informerFactories struct {
	informerFactory              k8sinformers.SharedInformerFactory
}

func NewInformerFactories(client kubernetes.Interface) InformerFactory {
	factory := &informerFactories{}

	if client != nil {
		factory.informerFactory = k8sinformers.NewSharedInformerFactory(client, defaultResync)
	}

	return factory
}

func (f *informerFactories) KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory {
	return f.informerFactory
}

func (f *informerFactories) Start(stopCh <-chan struct{}) {
	if f.informerFactory != nil {
		f.informerFactory.Start(stopCh)
	}
}