package identityprovider

import (
	"errors"
	"fmt"

	"k8s.io/klog"

	"aiscope/pkg/apiserver/authentication/oauth"
)

var (
	oauthProviderFactories   = make(map[string]OAuthProviderFactory)
	genericProviderFactories = make(map[string]GenericProviderFactory)
	identityProviderNotFound = errors.New("identity provider not found")
	oauthProviders           = make(map[string]OAuthProvider)
	genericProviders         = make(map[string]GenericProvider)
)

// Identity represents the account mapped to aiscope
type Identity interface {
	// GetUserID required
	// Identifier for the End-User at the Issuer.
	GetUserID() string
	// GetUsername optional
	// The username which the End-User wishes to be referred to aiscope.
	GetUsername() string
	// GetEmail optional
	GetEmail() string
}

// SetupWithOptions will verify the configuration and initialize the identityProviders
func SetupWithOptions(options []oauth.IdentityProviderOptions) error {
	for _, o := range options {
		if oauthProviders[o.Name] != nil || genericProviders[o.Name] != nil {
			err := fmt.Errorf("duplicate identity provider found: %s, name must be unique", o.Name)
			klog.Error(err)
			return err
		}
		if genericProviderFactories[o.Type] == nil && oauthProviderFactories[o.Type] == nil {
			err := fmt.Errorf("identity provider %s with type %s is not supported", o.Name, o.Type)
			klog.Error(err)
			return err
		}
		if factory, ok := oauthProviderFactories[o.Type]; ok {
			if provider, err := factory.Create(o.Provider); err != nil {
				// don’t return errors, decoupling external dependencies
				klog.Error(fmt.Sprintf("failed to create identity provider %s: %s", o.Name, err))
			} else {
				oauthProviders[o.Name] = provider
				klog.V(4).Infof("create identity provider %s successfully", o.Name)
			}
		}
		if factory, ok := genericProviderFactories[o.Type]; ok {
			if provider, err := factory.Create(o.Provider); err != nil {
				klog.Error(fmt.Sprintf("failed to create identity provider %s: %s", o.Name, err))
			} else {
				genericProviders[o.Name] = provider
				klog.V(4).Infof("create identity provider %s successfully", o.Name)
			}
		}
	}
	return nil
}

// GetGenericProvider returns GenericProvider with given name
func GetGenericProvider(providerName string) (GenericProvider, error) {
	if provider, ok := genericProviders[providerName]; ok {
		return provider, nil
	}
	return nil, identityProviderNotFound
}

// GetOAuthProvider returns OAuthProvider with given name
func GetOAuthProvider(providerName string) (OAuthProvider, error) {
	if provider, ok := oauthProviders[providerName]; ok {
		return provider, nil
	}
	return nil, identityProviderNotFound
}

// RegisterOAuthProvider register OAuthProviderFactory with the specified type
func RegisterOAuthProvider(factory OAuthProviderFactory) {
	oauthProviderFactories[factory.Type()] = factory
}

// RegisterGenericProvider registers GenericProviderFactory with the specified type
func RegisterGenericProvider(factory GenericProviderFactory) {
	genericProviderFactories[factory.Type()] = factory
}
