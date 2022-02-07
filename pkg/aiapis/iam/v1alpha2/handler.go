package v1alpha2

import (
	"aiscope/pkg/api"
	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	"aiscope/pkg/apiserver/query"
	"aiscope/pkg/models/iam/im"
	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/endpoints/request"
)

type iamHandler struct {
	im         im.IdentityManagementInterface
}

func newIAMHandler(im im.IdentityManagementInterface) *iamHandler {
	return &iamHandler{
		im:         im,
	}
}

func (h *iamHandler) CreateUser(req *restful.Request, resp *restful.Response) {
	var user iamv1alpha2.User
	err := req.ReadEntity(&user)
	if err != nil {
		api.HandleBadRequest(resp, req, err)
		return
	}

	operator, ok := request.UserFrom(req.Request.Context())
	if ok && operator.GetName() == iamv1alpha2.PreRegistrationUser {
		extra := operator.GetExtra()
		// The token used for registration must contain additional information
		if len(extra[iamv1alpha2.ExtraIdentityProvider]) != 1 || len(extra[iamv1alpha2.ExtraUID]) != 1 {
			err = errors.NewBadRequest("invalid registration token")
			api.HandleBadRequest(resp, req, err)
			return
		}
		if user.Labels == nil {
			user.Labels = make(map[string]string)
		}
		user.Labels[iamv1alpha2.IdentifyProviderLabel] = extra[iamv1alpha2.ExtraIdentityProvider][0]
		user.Labels[iamv1alpha2.OriginUIDLabel] = extra[iamv1alpha2.ExtraUID][0]
		// default role
	}

	created, err := h.im.CreateUser(&user)
	if err != nil {
		api.HandleError(resp, req, err)
		return
	}

	// ensure encrypted password will not be output
	created.Spec.EncryptedPassword = ""

	resp.WriteEntity(created)
}


func (h *iamHandler) ListUsers(request *restful.Request, response *restful.Response) {
	queryParam := query.ParseQueryParameter(request)
	result, err := h.im.ListUsers(queryParam)
	if err != nil {
		api.HandleInternalError(response, request, err)
		return
	}
	// for i, item := range result.Items {
	// 	user := item.(*iamv1alpha2.User)
	// 	user = user.DeepCopy()
	// 	globalRole, err := h.im.GetGlobalRoleOfUser(user.Name)
	// 	// ignore not found error
	// 	if err != nil && !errors.IsNotFound(err) {
	// 		api.HandleInternalError(response, request, err)
	// 		return
	// 	}
	// 	if globalRole != nil {
	// 		user = appendGlobalRoleAnnotation(user, globalRole.Name)
	// 	}
	// 	result.Items[i] = user
	// }
	response.WriteEntity(result)
}

func appendGlobalRoleAnnotation(user *iamv1alpha2.User, globalRole string) *iamv1alpha2.User {
	if user.Annotations == nil {
		user.Annotations = make(map[string]string, 0)
	}
	user.Annotations[iamv1alpha2.GlobalRoleAnnotation] = globalRole
	return user
}

