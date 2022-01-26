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
	mimePatch := []string{restful.MIME_JSON, runtime.MimeMergePatchJson, runtime.MimeJsonPatchJson}

	ws := runtime.NewWebService(experimentv1alpha2.SchemeGroupVersion)
	handler := newHandler(ep)

	// trackingservers
	ws.Route(ws.POST("/namespaces/{namespace}/trackingservers").
		To(handler.CreateTrackingServer).
		Reads(experimentv1alpha2.TrackingServer{}).
		Param(ws.PathParameter("namespace", "namespace")).
		Doc("Create a trackingserver in the specified namespace.").
		Returns(http.StatusOK, api.StatusOK, experimentv1alpha2.TrackingServer{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ExperimentTrackingServerTag}))
	ws.Route(ws.PUT("/namespaces/{namespace}/trackingservers/{trackingserver}").
		To(handler.UpdateTrackingServer).
		Doc("Update trackingserver in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("trackingserver", "trackingserver name")).
		Reads(experimentv1alpha2.TrackingServer{}).
		Returns(http.StatusOK, api.StatusOK, experimentv1alpha2.TrackingServer{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ExperimentTrackingServerTag}))
	ws.Route(ws.PATCH("/namespaces/{namespace}/trackingservers/{trackingserver}").
		To(handler.PatchTrackingServer).
		Consumes(mimePatch...).
		Doc("Update trackingserver in the specified namespace.").
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("trackingserver", "trackingserver name")).
		Reads(experimentv1alpha2.TrackingServer{}).
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
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("trackingserver", "trackingserver name")).
		Doc("Retrieve trackingserver details.").
		Returns(http.StatusOK, api.StatusOK, experimentv1alpha2.TrackingServer{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ExperimentTrackingServerTag}))
	ws.Route(ws.DELETE("/namespaces/{namespace}/trackingservers/{trackingserver}").
		To(handler.DeleteTrackingServer).
		Param(ws.PathParameter("namespace", "namespace")).
		Param(ws.PathParameter("trackingserver", "trackingserver name")).
		Doc("Delete trackingserver under namespace.").
		Returns(http.StatusOK, api.StatusOK, experimentv1alpha2.TrackingServer{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.ExperimentTrackingServerTag}))

	container.Add(ws)
	return nil
}
