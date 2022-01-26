package trackingserver

import (
	"aiscope/pkg/api"
	experimentv1alpha2 "aiscope/pkg/apis/experiment/v1alpha2"
	"aiscope/pkg/apiserver/query"
	informers "aiscope/pkg/client/informers/externalversions"
	"aiscope/pkg/models/resources/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
)

type trackingserverGetter struct {
	sharedInformers informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha2.Interface {
	return &trackingserverGetter{sharedInformers: sharedInformers}
}

func (g *trackingserverGetter) Get(namespace, name string) (runtime.Object, error) {
	return g.sharedInformers.Experiment().V1alpha2().TrackingServers().Lister().TrackingServers(namespace).Get(name)
}

func (g *trackingserverGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	trackingservers, err := g.sharedInformers.Experiment().V1alpha2().TrackingServers().Lister().TrackingServers(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, ts := range trackingservers {
		result = append(result, ts)
	}
	return v1alpha2.DefaultList(result, query, g.compare, g.filter), nil
}

func (g *trackingserverGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftTrackingServer, ok := left.(*experimentv1alpha2.TrackingServer)
	if !ok {
		return false
	}
	rightTrackingServer, ok := right.(*experimentv1alpha2.TrackingServer)
	if !ok {
		return false
	}
	return v1alpha2.DefaultObjectMetaCompare(leftTrackingServer.ObjectMeta, rightTrackingServer.ObjectMeta, field)
}

func (g *trackingserverGetter) filter(object runtime.Object, filter query.Filter) bool {
	trackingserver, ok := object.(*experimentv1alpha2.TrackingServer)

	if !ok {
		return false
	}

	return v1alpha2.DefaultObjectMetaFilter(trackingserver.ObjectMeta, filter)
}
