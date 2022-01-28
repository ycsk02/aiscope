package identityprovider

import (
	"net/http"

	"aiscope/pkg/apiserver/authentication/oauth"
)

type OAuthProvider interface {
	// IdentityExchangeCallback handle oauth callback, exchange identity from remote server
	IdentityExchangeCallback(req *http.Request) (Identity, error)
}

type OAuthProviderFactory interface {
	// Type unique type of the provider
	Type() string
	// Create Apply the dynamic options
	Create(options oauth.DynamicOptions) (OAuthProvider, error)
}
