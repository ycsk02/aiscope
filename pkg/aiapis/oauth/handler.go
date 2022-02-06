package oauth

import (
	"aiscope/pkg/api"
	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	"aiscope/pkg/apiserver/authentication"
	"aiscope/pkg/apiserver/authentication/oauth"
	"aiscope/pkg/apiserver/authentication/token"
	"aiscope/pkg/apiserver/query"
	"aiscope/pkg/apiserver/request"
	"aiscope/pkg/models/auth"
	"aiscope/pkg/models/iam/im"
	"aiscope/pkg/server/errors"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/form3tech-oss/jwt-go"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"
	"net/http"
	"net/url"
)

const (
	KindTokenReview       = "TokenReview"
	grantTypePassword     = "password"
	grantTypeRefreshToken = "refresh_token"
	grantTypeCode         = "code"
)

type LoginRequest struct {
	Username string `json:"username" description:"username"`
	Password string `json:"password" description:"password"`
}

type handler struct {
	im 					  im.IdentityManagementInterface
	oauthAuthenticator    auth.OAuthAuthenticator
	passwordAuthenticator auth.PasswordAuthenticator
	tokenOperator         auth.TokenManagementInterface
	loginRecorder         auth.LoginRecorder
	options               *authentication.Options
}

func newHandler(tokenOperator auth.TokenManagementInterface,
	oauthAuthenticator auth.OAuthAuthenticator,
	passwordAuthenticator auth.PasswordAuthenticator,
	loginRecorder auth.LoginRecorder,
	options *authentication.Options) *handler {
	return &handler{
		tokenOperator:         tokenOperator,
		oauthAuthenticator:    oauthAuthenticator,
		passwordAuthenticator: passwordAuthenticator,
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

// To obtain an Access Token, an ID Token, and optionally a Refresh Token,
// the RP (Client) sends a Token Request to the Token Endpoint to obtain a Token Response,
// as described in Section 3.2 of OAuth 2.0 [RFC6749], when using the Authorization Code Flow.
// Communication with the Token Endpoint MUST utilize TLS.
func (h *handler) token(req *restful.Request, response *restful.Response) {
	// TODO(hongming) support basic auth
	// https://datatracker.ietf.org/doc/html/rfc6749#section-2.3
	clientID, err := req.BodyParameter("client_id")
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusUnauthorized, oauth.NewInvalidClient(err))
		return
	}
	clientSecret, err := req.BodyParameter("client_secret")
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusUnauthorized, oauth.NewInvalidClient(err))
		return
	}

	client, err := h.options.OAuthOptions.OAuthClient(clientID)
	if err != nil {
		oauthError := oauth.NewInvalidClient(err)
		response.WriteHeaderAndEntity(http.StatusUnauthorized, oauthError)
		return
	}

	if client.Secret != clientSecret {
		oauthError := oauth.NewInvalidClient(fmt.Errorf("invalid client credential"))
		response.WriteHeaderAndEntity(http.StatusUnauthorized, oauthError)
		return
	}

	grantType, err := req.BodyParameter("grant_type")
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidRequest(err))
		return
	}

	switch grantType {
	case grantTypePassword:
		username, _ := req.BodyParameter("username")
		password, _ := req.BodyParameter("password")
		h.passwordGrant(username, password, req, response)
		return
	case grantTypeRefreshToken:
		h.refreshTokenGrant(req, response)
		return
	default:
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.ErrorUnsupportedGrantType)
		return
	}
}

// passwordGrant handle Resource Owner Password Credentials Grant
// for more details: https://datatracker.ietf.org/doc/html/rfc6749#section-4.3
// The resource owner password credentials grant type is suitable in
// cases where the resource owner has a trust relationship with the client,
// such as the device operating system or a highly privileged application.
// The authorization server should take special care when enabling this
// grant type and only allow it when other flows are not viable.
func (h *handler) passwordGrant(username string, password string, req *restful.Request, response *restful.Response) {
	authenticated, provider, err := h.passwordAuthenticator.Authenticate(req.Request.Context(), username, password)
	if err != nil {
		switch err {
		case auth.IncorrectPasswordError:
			requestInfo, _ := request.RequestInfoFrom(req.Request.Context())
			if err := h.loginRecorder.RecordLogin(username, iamv1alpha2.Token, provider, requestInfo.SourceIP, requestInfo.UserAgent, err); err != nil {
				klog.Errorf("Failed to record unsuccessful login attempt for user %s, error: %v", username, err)
			}
			response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant(err))
			return
		case auth.RateLimitExceededError:
			response.WriteHeaderAndEntity(http.StatusTooManyRequests, oauth.NewInvalidGrant(err))
			return
		default:
			response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(err))
			return
		}
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

func (h *handler) refreshTokenGrant(req *restful.Request, response *restful.Response) {
	refreshToken, err := req.BodyParameter("refresh_token")
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidRequest(err))
		return
	}

	verified, err := h.tokenOperator.Verify(refreshToken)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant(err))
		return
	}

	if verified.TokenType != token.RefreshToken {
		err = fmt.Errorf("ivalid token type %v want %v", verified.TokenType, token.RefreshToken)
		response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant(err))
		return
	}

	authenticated := verified.User
	// update token after registration
	if authenticated.GetName() == iamv1alpha2.PreRegistrationUser &&
		authenticated.GetExtra() != nil &&
		len(authenticated.GetExtra()[iamv1alpha2.ExtraIdentityProvider]) > 0 &&
		len(authenticated.GetExtra()[iamv1alpha2.ExtraUID]) > 0 {

		idp := authenticated.GetExtra()[iamv1alpha2.ExtraIdentityProvider][0]
		uid := authenticated.GetExtra()[iamv1alpha2.ExtraUID][0]
		queryParam := query.New()
		queryParam.LabelSelector = labels.SelectorFromSet(labels.Set{
			iamv1alpha2.IdentifyProviderLabel: idp,
			iamv1alpha2.OriginUIDLabel:        uid}).String()
		result, err := h.im.ListUsers(queryParam)
		if err != nil {
			response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(err))
			return
		}
		if len(result.Items) != 1 {
			response.WriteHeaderAndEntity(http.StatusBadRequest, oauth.NewInvalidGrant(fmt.Errorf("authenticated user does not exist")))
			return
		}

		authenticated = &user.DefaultInfo{Name: result.Items[0].(*iamv1alpha2.User).Name}
	}

	result, err := h.issueTokenTo(authenticated)
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(err))
		return
	}

	response.WriteEntity(result)
}

func (h *handler) logout(req *restful.Request, resp *restful.Response) {
	authenticated, ok := request.UserFrom(req.Request.Context())
	if ok {
		if err := h.tokenOperator.RevokeAllUserTokens(authenticated.GetName()); err != nil {
			api.HandleInternalError(resp, req, apierrors.NewInternalError(err))
			return
		}
	}

	postLogoutRedirectURI := req.QueryParameter("post_logout_redirect_uri")
	if postLogoutRedirectURI == "" {
		resp.WriteAsJson(errors.None)
		return
	}

	redirectURL, err := url.Parse(postLogoutRedirectURI)
	if err != nil {
		api.HandleBadRequest(resp, req, fmt.Errorf("invalid logout redirect URI: %s", err))
		return
	}

	state := req.QueryParameter("state")
	if state != "" {
		redirectURL.Query().Add("state", state)
	}

	resp.Header().Set("Content-Type", "text/plain")
	http.Redirect(resp, req.Request, redirectURL.String(), http.StatusFound)
}

// userinfo Endpoint is an OAuth 2.0 Protected Resource that returns Claims about the authenticated End-User.
func (h *handler) userinfo(req *restful.Request, response *restful.Response) {
	authenticated, _ := request.UserFrom(req.Request.Context())
	if authenticated == nil || authenticated.GetName() == user.Anonymous {
		response.WriteHeaderAndEntity(http.StatusUnauthorized, oauth.ErrorLoginRequired)
		return
	}
	detail, err := h.im.DescribeUser(authenticated.GetName())
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, oauth.NewServerError(err))
		return
	}

	result := token.Claims{
		StandardClaims: jwt.StandardClaims{
			Subject: detail.Name,
		},
		Name:              detail.Name,
		Email:             detail.Spec.Email,
		Locale:            detail.Spec.Lang,
		PreferredUsername: detail.Name,
	}
	response.WriteEntity(result)
}

func (h *handler) login(request *restful.Request, response *restful.Response) {
	var loginRequest LoginRequest
	err := request.ReadEntity(&loginRequest)
	if err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}
	h.passwordGrant(loginRequest.Username, loginRequest.Password, request, response)
}