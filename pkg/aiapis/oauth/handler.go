package oauth

import (
	"aiscope/pkg/api"
	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	"aiscope/pkg/apiserver/authentication/oauth"
	"aiscope/pkg/apiserver/authentication/token"
	"aiscope/pkg/apiserver/request"
	"aiscope/pkg/authentication"
	"aiscope/pkg/models/auth"
	"fmt"
	"github.com/emicklei/go-restful"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"
	"net/http"
)

type handler struct {
	oauthAuthenticator    auth.OAuthAuthenticator
	tokenOperator         auth.TokenManagementInterface
	loginRecorder         auth.LoginRecorder
	options               *authentication.Options
}

func newHandler(tokenOperator auth.TokenManagementInterface,
	oauthAuthenticator auth.OAuthAuthenticator,
	loginRecorder auth.LoginRecorder,
	options *authentication.Options) *handler {
	return &handler{
		tokenOperator:         tokenOperator,
		oauthAuthenticator:    oauthAuthenticator,
		loginRecorder:         loginRecorder,
		options:               options,
	}
}

func (h *handler) oauthCallback(req *restful.Request, response *restful.Response) {
	provider := req.PathParameter("callback")
	authenticated, provider, err := h.oauthAuthenticator.Authenticate(req.Request.Context(), provider, req.Request)
	if err != nil {
		api.HandleUnauthorized(response, req, apierrors.NewUnauthorized(fmt.Sprintf("Unauthorized: %s", err)))
		return
	}

	result, err := h.issueTokenTo(authenticated)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(err))
		return
	}

	requestInfo, _ := request.RequestInfoFrom(req.Request.Context())
	if err = h.loginRecorder.RecordLogin(authenticated.GetName(), iamv1alpha2.Token, provider, requestInfo.SourceIP, requestInfo.UserAgent, nil); err != nil {
		klog.Errorf("Failed to record successful login for user %s, error: %v", authenticated.GetName(), err)
	}

	response.WriteEntity(result)
}

func (h *handler) issueTokenTo(user user.Info) (*oauth.Token, error) {
	accessToken, err := h.tokenOperator.IssueTo(&token.IssueRequest{
		User:      user,
		Claims:    token.Claims{TokenType: token.AccessToken},
		ExpiresIn: h.options.OAuthOptions.AccessTokenMaxAge,
	})
	if err != nil {
		return nil, err
	}
	refreshToken, err := h.tokenOperator.IssueTo(&token.IssueRequest{
		User:      user,
		Claims:    token.Claims{TokenType: token.RefreshToken},
		ExpiresIn: h.options.OAuthOptions.AccessTokenMaxAge + h.options.OAuthOptions.AccessTokenInactivityTimeout,
	})
	if err != nil {
		return nil, err
	}

	result := oauth.Token{
		AccessToken: accessToken,
		// The OAuth 2.0 token_type response parameter value MUST be Bearer,
		// as specified in OAuth 2.0 Bearer Token Usage [RFC6750]
		TokenType:    "Bearer",
		RefreshToken: refreshToken,
		ExpiresIn:    int(h.options.OAuthOptions.AccessTokenMaxAge.Seconds()),
	}
	return &result, nil
}
