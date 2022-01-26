package resource

import (
	"aiscope/pkg/api"
	experimentv1alpha2 "aiscope/pkg/apis/experiment/v1alpha2"
	"aiscope/pkg/apiserver/query"
	"aiscope/pkg/informers"
	"aiscope/pkg/models/resources/v1alpha2"
	"aiscope/pkg/models/resources/v1alpha2/namespace"
	"aiscope/pkg/models/resources/v1alpha2/trackingserver"
	"aiscope/pkg/server/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var ErrResourceNotSupported = errors.New("resource is not supported")

type ResourceGetter struct {
	clusterResourceGetters    map[schema.GroupVersionResource]v1alpha2.Interface
	namespacedResourceGetters map[schema.GroupVersionResource]v1alpha2.Interface
}

func NewResourceGetter(factory informers.InformerFactory) *ResourceGetter {
	namespacedResourceGetters := make(map[schema.GroupVersionResource]v1alpha2.Interface)
	clusterResourceGetters := make(map[schema.GroupVersionResource]v1alpha2.Interface)

	clusterResourceGetters[schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}] = namespace.New(factory.KubernetesSharedInformerFactory())
	namespacedResourceGetters[experimentv1alpha2.SchemeGroupVersion.WithResource(experimentv1alpha2.ResourcePluralTrackingServer)] = trackingserver.New(factory.AIScopeSharedInformerFactory())

	return &ResourceGetter{
		namespacedResourceGetters: namespacedResourceGetters,
		clusterResourceGetters:    clusterResourceGetters,
	}
}

func (r *ResourceGetter) TryResource(clusterScope bool, resource string) v1alpha2.Interface {
	if clusterScope {
		for k, v := range r.clusterResourceGetters {
			if k.Resource == resource {
				return v
			}
		}
	}
	for k, v := range r.namespacedResourceGetters {
		if k.Resource == resource {
			return v
		}
	}
	return nil
}

func (r *ResourceGetter) Get(resource, namespace, name string) (runtime.Object, error) {
	clusterScope := namespace == ""
	getter := r.TryResource(clusterScope, resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}
	return getter.Get(namespace, name)
}

func (r *ResourceGetter) List(resource, namespace string, query *query.Query) (*api.ListResult, error) {
	clusterScope := namespace == ""
	getter := r.TryResource(clusterScope, resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}
	return getter.List(namespace, query)
}
