package namespace

import (
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"

	"aiscope/pkg/api"
	"aiscope/pkg/apiserver/query"
	"aiscope/pkg/models/resources/v1alpha2"
)

type namespacesGetter struct {
	informers informers.SharedInformerFactory
}

func New(informers informers.SharedInformerFactory) v1alpha2.Interface {
	return &namespacesGetter{informers: informers}
}

func (n namespacesGetter) Get(_, name string) (runtime.Object, error) {
	return n.informers.Core().V1().Namespaces().Lister().Get(name)
}

func (n namespacesGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	ns, err := n.informers.Core().V1().Namespaces().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, item := range ns {
		result = append(result, item)
	}

	return v1alpha2.DefaultList(result, query, n.compare, n.filter), nil
}

func (n namespacesGetter) filter(item runtime.Object, filter query.Filter) bool {
	namespace, ok := item.(*v1.Namespace)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(string(namespace.Status.Phase), string(filter.Value)) == 0
	default:
		return v1alpha2.DefaultObjectMetaFilter(namespace.ObjectMeta, filter)
	}
}

func (n namespacesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftNs, ok := left.(*v1.Namespace)
	if !ok {
		return false
	}

	rightNs, ok := right.(*v1.Namespace)
	if !ok {
		return true
	}
	return v1alpha2.DefaultObjectMetaCompare(leftNs.ObjectMeta, rightNs.ObjectMeta, field)
}
