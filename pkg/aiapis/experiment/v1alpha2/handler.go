package v1alpha2

import (
	"aiscope/pkg/api"
	experimentv1alpha2 "aiscope/pkg/apis/experiment/v1alpha2"
	"aiscope/pkg/apiserver/query"
	model "aiscope/pkg/models/experiment"
	servererr "aiscope/pkg/server/errors"
	"fmt"
	"github.com/emicklei/go-restful"
)

type handler struct {
	ep      model.Interface
}

func newHandler(ep model.Interface) *handler {
	return &handler{
		ep:         ep,
	}
}

func (h *handler) CreateTrackingServer(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	var trackingserver *experimentv1alpha2.TrackingServer
	if err := request.ReadEntity(&trackingserver); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.ep.CreateOrUpdateTrackingServer(namespace, trackingserver)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(created)
}

func (h *handler) UpdateTrackingServer(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	trackingserverName := request.PathParameter("trackingserver")

	var trackingserver experimentv1alpha2.TrackingServer
	err := request.ReadEntity(&trackingserver)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if trackingserverName != trackingserver.Name {
		err := fmt.Errorf("the name of the object (%s) does not match the name on the URL (%s)", trackingserver.Name, trackingserverName)
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.ep.CreateOrUpdateTrackingServer(namespace, &trackingserver)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(updated)
}

func (h *handler) PatchTrackingServer(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	trackingserverName := request.PathParameter("trackingserver")

	var trackingserver experimentv1alpha2.TrackingServer
	err := request.ReadEntity(&trackingserver)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	trackingserver.Name = trackingserverName
	patched, err := h.ep.PatchTrackingServer(namespace, &trackingserver)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(patched)
}

func (h *handler) ListTrackingServer(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	queryParam := query.ParseQueryParameter(request)

	result, err := h.ep.ListTrackingServers(namespace, queryParam)
	if err != nil {
		api.HandleError(response, nil, err)
	}

	response.WriteEntity(result)
}

func (h *handler) DescribeTrackingServer(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	trackingserverName := request.PathParameter("trackingserver")

	trackingserver, err := h.ep.DescribeTrackingServer(namespace, trackingserverName)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(trackingserver)
}

func (h *handler) DeleteTrackingServer(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	trackingserverName := request.PathParameter("trackingserver")

	err := h.ep.DeleteTrackingServer(namespace, trackingserverName)
	if err != nil {
		api.HandleError(response, request, err)
		return
	}

	response.WriteEntity(servererr.None)
}
