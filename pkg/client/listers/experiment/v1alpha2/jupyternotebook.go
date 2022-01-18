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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha2

import (
	v1alpha2 "aiscope/pkg/apis/experiment/v1alpha2"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// JupyterNotebookLister helps list JupyterNotebooks.
// All objects returned here must be treated as read-only.
type JupyterNotebookLister interface {
	// List lists all JupyterNotebooks in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.JupyterNotebook, err error)
	// JupyterNotebooks returns an object that can list and get JupyterNotebooks.
	JupyterNotebooks(namespace string) JupyterNotebookNamespaceLister
	JupyterNotebookListerExpansion
}

// jupyterNotebookLister implements the JupyterNotebookLister interface.
type jupyterNotebookLister struct {
	indexer cache.Indexer
}

// NewJupyterNotebookLister returns a new JupyterNotebookLister.
func NewJupyterNotebookLister(indexer cache.Indexer) JupyterNotebookLister {
	return &jupyterNotebookLister{indexer: indexer}
}

// List lists all JupyterNotebooks in the indexer.
func (s *jupyterNotebookLister) List(selector labels.Selector) (ret []*v1alpha2.JupyterNotebook, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.JupyterNotebook))
	})
	return ret, err
}

// JupyterNotebooks returns an object that can list and get JupyterNotebooks.
func (s *jupyterNotebookLister) JupyterNotebooks(namespace string) JupyterNotebookNamespaceLister {
	return jupyterNotebookNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// JupyterNotebookNamespaceLister helps list and get JupyterNotebooks.
// All objects returned here must be treated as read-only.
type JupyterNotebookNamespaceLister interface {
	// List lists all JupyterNotebooks in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.JupyterNotebook, err error)
	// Get retrieves the JupyterNotebook from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha2.JupyterNotebook, error)
	JupyterNotebookNamespaceListerExpansion
}

// jupyterNotebookNamespaceLister implements the JupyterNotebookNamespaceLister
// interface.
type jupyterNotebookNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all JupyterNotebooks in the indexer for a given namespace.
func (s jupyterNotebookNamespaceLister) List(selector labels.Selector) (ret []*v1alpha2.JupyterNotebook, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.JupyterNotebook))
	})
	return ret, err
}

// Get retrieves the JupyterNotebook from the indexer for a given namespace and name.
func (s jupyterNotebookNamespaceLister) Get(name string) (*v1alpha2.JupyterNotebook, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha2.Resource("jupyternotebook"), name)
	}
	return obj.(*v1alpha2.JupyterNotebook), nil
}
