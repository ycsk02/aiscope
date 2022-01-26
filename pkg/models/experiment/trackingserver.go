package experiment

import (
	"aiscope/pkg/api"
	experimentv1alpha2 "aiscope/pkg/apis/experiment/v1alpha2"
	"aiscope/pkg/apiserver/query"
	"context"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func (o *Operator) CreateOrUpdateTrackingServer(namespace string, trackingserver *experimentv1alpha2.TrackingServer) (*experimentv1alpha2.TrackingServer, error) {
	var created *experimentv1alpha2.TrackingServer
	var err error

	trackingserver.Namespace = namespace

	if trackingserver.ResourceVersion != "" {
		created, err = o.aiclient.ExperimentV1alpha2().TrackingServers(namespace).Update(context.Background(), trackingserver, metav1.UpdateOptions{})
	} else {
		created, err = o.aiclient.ExperimentV1alpha2().TrackingServers(namespace).Create(context.Background(), trackingserver, metav1.CreateOptions{})
	}

	return created, err
}

func (o *Operator) PatchTrackingServer(namespace string, trackingserver *experimentv1alpha2.TrackingServer) (*experimentv1alpha2.TrackingServer, error) {
	data, err := json.Marshal(trackingserver)
	if err != nil {
		return nil, err
	}

	return o.aiclient.ExperimentV1alpha2().TrackingServers(namespace).Patch(context.Background(), trackingserver.Name, types.MergePatchType, data, metav1.PatchOptions{})
}

func (o *Operator) DeleteTrackingServer(namespace string, name string) error {
	return o.aiclient.ExperimentV1alpha2().TrackingServers(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func (o *Operator) ListTrackingServers(namespace string, queryParam *query.Query) (*api.ListResult, error) {
	result, err := o.resourceGetter.List(experimentv1alpha2.ResourcePluralTrackingServer, namespace, queryParam)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return result, nil
}

func (o *Operator) DescribeTrackingServer(namespace, name string) (*experimentv1alpha2.TrackingServer, error) {
	obj, err := o.resourceGetter.Get(experimentv1alpha2.ResourcePluralTrackingServer, namespace, name)
	if err != nil {
		return nil, err
	}
	result := obj.(*experimentv1alpha2.TrackingServer)
	return result, nil
}
