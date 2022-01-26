package v1alpha2

import (
	"aiscope/pkg/api"
	experimentv1alpha2 "aiscope/pkg/apis/experiment/v1alpha2"
	"aiscope/pkg/apiserver/runtime"
	"aiscope/pkg/constants"
	model "aiscope/pkg/models/experiment"
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"net/http"
)

func AddToContainer(container *restful.Container, ep model.Interface) error {
	ws := runtime.NewWebService(experimentv1alpha2.SchemeGroupVersion)
	handler := newHandler(ep)

	// trackingservers
	ws.Route(ws.POST("/namespaces/{namespace}/trackingservers").
		To(handler.CreateTrackingServer).
		Doc("Create a tracking server in the specified namespace.").
		Reads(experimentv1alpha2.TrackingServer{}).
		Param(ws.PathParameter("namespace", "namespace")).
		Returns(http.StatusOK, api.StatusOK, experimentv1alpha2.TrackingServer{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ExperimentTrackingServerTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/trackingservers").
		To(handler.ListTrackingServer).
		Param(ws.PathParameter("namespace", "namespace")).
		Doc("List the trackingservers of the specified namespace for the current user").
		Returns(http.StatusOK, api.StatusOK, api.ListResult{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ExperimentTrackingServerTag}))
	ws.Route(ws.GET("/namespaces/{namespace}/trackingservers/{trackingserver}").
		To(handler.DescribeTrackingServer).
		Doc("Retrieve trackingserver details.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("trackingserver", "trackingserver name")).
		Returns(http.StatusOK, api.StatusOK, experimentv1alpha2.TrackingServer{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ExperimentTrackingServerTag}))

	container.Add(ws)
	return nil
}
