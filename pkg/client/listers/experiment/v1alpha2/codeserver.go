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

// CodeServerLister helps list CodeServers.
// All objects returned here must be treated as read-only.
type CodeServerLister interface {
	// List lists all CodeServers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.CodeServer, err error)
	// CodeServers returns an object that can list and get CodeServers.
	CodeServers(namespace string) CodeServerNamespaceLister
	CodeServerListerExpansion
}

// codeServerLister implements the CodeServerLister interface.
type codeServerLister struct {
	indexer cache.Indexer
}

// NewCodeServerLister returns a new CodeServerLister.
func NewCodeServerLister(indexer cache.Indexer) CodeServerLister {
	return &codeServerLister{indexer: indexer}
}

// List lists all CodeServers in the indexer.
func (s *codeServerLister) List(selector labels.Selector) (ret []*v1alpha2.CodeServer, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.CodeServer))
	})
	return ret, err
}

// CodeServers returns an object that can list and get CodeServers.
func (s *codeServerLister) CodeServers(namespace string) CodeServerNamespaceLister {
	return codeServerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// CodeServerNamespaceLister helps list and get CodeServers.
// All objects returned here must be treated as read-only.
type CodeServerNamespaceLister interface {
	// List lists all CodeServers in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.CodeServer, err error)
	// Get retrieves the CodeServer from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha2.CodeServer, error)
	CodeServerNamespaceListerExpansion
}

// codeServerNamespaceLister implements the CodeServerNamespaceLister
// interface.
type codeServerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all CodeServers in the indexer for a given namespace.
func (s codeServerNamespaceLister) List(selector labels.Selector) (ret []*v1alpha2.CodeServer, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.CodeServer))
	})
	return ret, err
}

// Get retrieves the CodeServer from the indexer for a given namespace and name.
func (s codeServerNamespaceLister) Get(name string) (*v1alpha2.CodeServer, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha2.Resource("codeserver"), name)
	}
	return obj.(*v1alpha2.CodeServer), nil
}
