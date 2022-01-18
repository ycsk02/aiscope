/*
Copyright 2020 The AIScope Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

    https://vectorcloud.io
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha2

import (
	experimentv1alpha2 "aiscope/pkg/apis/experiment/v1alpha2"
	versioned "aiscope/pkg/client/clientset/versioned"
	internalinterfaces "aiscope/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha2 "aiscope/pkg/client/listers/experiment/v1alpha2"
	"context"
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// JupyterNotebookInformer provides access to a shared informer and lister for
// JupyterNotebooks.
type JupyterNotebookInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha2.JupyterNotebookLister
}

type jupyterNotebookInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewJupyterNotebookInformer constructs a new informer for JupyterNotebook type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewJupyterNotebookInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredJupyterNotebookInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredJupyterNotebookInformer constructs a new informer for JupyterNotebook type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredJupyterNotebookInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ExperimentV1alpha2().JupyterNotebooks(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ExperimentV1alpha2().JupyterNotebooks(namespace).Watch(context.TODO(), options)
			},
		},
		&experimentv1alpha2.JupyterNotebook{},
		resyncPeriod,
		indexers,
	)
}

func (f *jupyterNotebookInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredJupyterNotebookInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *jupyterNotebookInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&experimentv1alpha2.JupyterNotebook{}, f.defaultInformer)
}

func (f *jupyterNotebookInformer) Lister() v1alpha2.JupyterNotebookLister {
	return v1alpha2.NewJupyterNotebookLister(f.Informer().GetIndexer())
}
