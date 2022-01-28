package identityprovider

import (
	"aiscope/pkg/apiserver/authentication/oauth"
)

type GenericProvider interface {
	// Authenticate from remote server
	Authenticate(username string, password string) (Identity, error)
}

type GenericProviderFactory interface {
	// Type unique type of the provider
	Type() string
	// Apply the dynamic options from aiscope-config
	Create(options oauth.DynamicOptions) (GenericProvider, error)
}
