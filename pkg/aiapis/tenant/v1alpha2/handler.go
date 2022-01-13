package v1alpha2

import (
	"aiscope/pkg/api"
	tenantv1alpha2 "aiscope/pkg/apis/tenant/v1alpha2"
	"aiscope/pkg/apiserver/query"
	aiscope "aiscope/pkg/client/clientset/versioned"
	"aiscope/pkg/models/tenant"
	"fmt"
	"github.com/emicklei/go-restful"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type tenantHandler struct {
	tenant       tenant.Interface
}

func newTenantHandler(aiclient aiscope.Interface, k8sclient kubernetes.Interface) *tenantHandler {
	return &tenantHandler{
		tenant: tenant.NewOperator(aiclient, k8sclient),
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

func (h *tenantHandler) CreateNamespace(request *restful.Request, response *restful.Response) {
	workspace := request.PathParameter("workspace")
	var namespace corev1.Namespace

	err := request.ReadEntity(&namespace)

	if err != nil {
		klog.Error(err)
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.tenant.CreateNamespace(workspace, &namespace)

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

func (h *tenantHandler) ListNamespaces(req *restful.Request, resp *restful.Response) {
	workspace := req.PathParameter("workspace")
	queryParam := query.ParseQueryParameter(req)

	var workspaceMember user.Info
	if username := req.PathParameter("workspacemember"); username != "" {
		workspaceMember = &user.DefaultInfo{
			Name: username,
		}
	} else {
		requestUser, ok := request.UserFrom(req.Request.Context())
		if !ok {
			err := fmt.Errorf("cannot obtain user info")
			klog.Errorln(err)
			api.HandleForbidden(resp, nil, err)
			return
		}
		workspaceMember = requestUser
	}

	result, err := h.tenant.ListNamespaces(workspaceMember, workspace, queryParam)
	if err != nil {
		api.HandleInternalError(resp, nil, err)
		return
	}

	resp.WriteEntity(result)
}
