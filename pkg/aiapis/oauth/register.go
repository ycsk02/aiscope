package oauth

import (
	"aiscope/pkg/api"
	"aiscope/pkg/apiserver/authentication"
	"aiscope/pkg/apiserver/authentication/oauth"
	"aiscope/pkg/constants"
	"aiscope/pkg/models/auth"
	"aiscope/pkg/models/iam/im"
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"net/http"
)

func AddToContainer(container *restful.Container, im im.IdentityManagementInterface,
	tokenOperator auth.TokenManagementInterface,
	oauth2Authenticator auth.OAuthAuthenticator,
	loginRecorder auth.LoginRecorder,
	options *authentication.Options) error {
	ws := &restful.WebService{}
	ws.Path("/oauth").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	handler := newHandler(tokenOperator, oauth2Authenticator, loginRecorder, options)

	ws.Route(ws.GET("/callback/{callback}").
		Doc("OAuth callback API, the path param callback is config by identity provider").
		Param(ws.QueryParameter("access_token", "The access token issued by the authorization server.").
			Required(true)).
		Param(ws.QueryParameter("token_type", "The type of the token issued as described in [RFC6479] Section 7.1. "+
			"Value is case insensitive.").Required(true)).
		Param(ws.QueryParameter("expires_in", "The lifetime in seconds of the access token.  For "+
			"example, the value \"3600\" denotes that the access token will "+
			"expire in one hour from the time the response was generated."+
			"If omitted, the authorization server SHOULD provide the "+
			"expiration time via other means or document the default value.")).
		Param(ws.QueryParameter("scope", "if identical to the scope requested by the client;"+
			"otherwise, REQUIRED.  The scope of the access token as described by [RFC6479] Section 3.3.").Required(false)).
		Param(ws.QueryParameter("state", "if the \"state\" parameter was present in the client authorization request."+
			"The exact value received from the client.").Required(true)).
		To(handler.oauthCallback).
		Returns(http.StatusOK, api.StatusOK, oauth.Token{}).
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.AuthenticationTag}))
	container.Add(ws)

	return nil
}
