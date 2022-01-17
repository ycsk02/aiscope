package informers

import (
	"aiscope/pkg/client/clientset/versioned"
	aiscopeinformers "aiscope/pkg/client/informers/externalversions"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"time"
)

const defaultResync = 600 * time.Second

// InformerFactory is a group all shared informer factories which aiscope needed
// callers should check if the return value is nil
type InformerFactory interface {
	KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory
	AIScopeSharedInformerFactory() aiscopeinformers.SharedInformerFactory

	// Start shared informer factory one by one if they are not nil
	Start(stopCh <-chan struct{})
}

type informerFactories struct {
	informerFactory              k8sinformers.SharedInformerFactory
	aiInformerFactory            aiscopeinformers.SharedInformerFactory
}

func NewInformerFactories(client kubernetes.Interface, aiClient versioned.Interface) InformerFactory {
	factory := &informerFactories{}

	if client != nil {
		factory.informerFactory = k8sinformers.NewSharedInformerFactory(client, defaultResync)
	}

	if aiClient != nil {
		factory.aiInformerFactory = aiscopeinformers.NewSharedInformerFactory(aiClient, defaultResync)
	}

	return factory
}

func (f *informerFactories) KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory {
	return f.informerFactory
}

func (f *informerFactories) AIScopeSharedInformerFactory() aiscopeinformers.SharedInformerFactory {
	return f.aiInformerFactory
}

func (f *informerFactories) Start(stopCh <-chan struct{}) {
	if f.informerFactory != nil {
		f.informerFactory.Start(stopCh)
	}
}
