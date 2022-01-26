package experiment

import (
	"aiscope/pkg/api"
	experimentv1alpha2 "aiscope/pkg/apis/experiment/v1alpha2"
	"aiscope/pkg/apiserver/query"
	aiscope "aiscope/pkg/client/clientset/versioned"
	"aiscope/pkg/informers"
	resourcesv1alpha2 "aiscope/pkg/models/resources/v1alpha2/resource"
)

type Interface interface {
	CreateOrUpdateTrackingServer(namespace string, trackingserver *experimentv1alpha2.TrackingServer) (*experimentv1alpha2.TrackingServer, error)
	DeleteTrackingServer(namespace string, name string) error
	ListTrackingServers(namespace string, queryParam *query.Query) (*api.ListResult, error)
	DescribeTrackingServer(namespace, trackingserver string) (*experimentv1alpha2.TrackingServer, error)
}

type Operator struct {
	aiclient          aiscope.Interface
	resourceGetter    *resourcesv1alpha2.ResourceGetter
}

func New(aiclient aiscope.Interface, informers informers.InformerFactory) Interface {
	return &Operator{
		aiclient:           aiclient,
		resourceGetter:     resourcesv1alpha2.NewResourceGetter(informers),
	}
}


