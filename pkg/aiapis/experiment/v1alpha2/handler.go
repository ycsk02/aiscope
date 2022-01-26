package v1alpha2

import (
	"aiscope/pkg/api"
	experimentv1alpha2 "aiscope/pkg/apis/experiment/v1alpha2"
	"aiscope/pkg/apiserver/query"
	"aiscope/pkg/constants"
	model "aiscope/pkg/models/experiment"
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

func (h *handler) resolveNamespace(workspace string) string {
	return fmt.Sprintf(constants.TenantDevopsNamespaceFormat, workspace)
}

func (h *handler) CreateTrackingServer(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	var trackingserver *experimentv1alpha2.TrackingServer
	if err := req.ReadEntity(&trackingserver); err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	created, err := h.ep.CreateOrUpdateTrackingServer(namespace, trackingserver)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	resp.WriteEntity(created)
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
