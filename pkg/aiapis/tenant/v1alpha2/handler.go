package v1alpha2

import (
	"aiscope/pkg/api"
	tenantv1alpha2 "aiscope/pkg/apis/tenant/v1alpha2"
	aiscope "aiscope/pkg/client/clientset/versioned"
	"aiscope/pkg/models/tenant"
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
)

type tenantHandler struct {
	tenant       tenant.Interface
}

func newTenantHandler(aiclient aiscope.Interface) *tenantHandler {
	return &tenantHandler{
		tenant: tenant.NewOperator(aiclient),
	}
}

func (h *tenantHandler) CreateWorkspace(request *restful.Request, response *restful.Response) {
	var workspace tenantv1alpha2.Workspace

	err := request.ReadEntity(&workspace)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.tenant.CreateWorkspace(&workspace)

	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(response, request, err)
			return
		}
		api.HandleBadRequest(response, request, err)
		return
	}

	response.WriteEntity(created)
}
