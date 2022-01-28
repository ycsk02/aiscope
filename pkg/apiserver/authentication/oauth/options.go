package oauth

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type GrantHandlerType string
type MappingMethod string
type IdentityProviderType string

const (
	// GrantHandlerAuto auto-approves client authorization grant requests
	GrantHandlerAuto GrantHandlerType = "auto"
	// GrantHandlerPrompt prompts the user to approve new client authorization grant requests
	GrantHandlerPrompt GrantHandlerType = "prompt"
	// GrantHandlerDeny auto-denies client authorization grant requests
	GrantHandlerDeny GrantHandlerType = "deny"
	// MappingMethodAuto  The default value.
	// The user will automatically create and mapping when login successful.
	// Fails if a user with that username is already mapped to another identity.
	MappingMethodAuto MappingMethod = "auto"
	// MappingMethodLookup Looks up an existing identity, user identity mapping, and user, but does not automatically
	// provision users or identities. Using this method requires you to manually provision users.
	MappingMethodLookup MappingMethod = "lookup"
	// MappingMethodMixed  A user entity can be mapped with multiple identifyProvider.
	// not supported yet.
	MappingMethodMixed MappingMethod = "mixed"

	DefaultIssuer string = "aiscope"
)

var (
	ErrorClientNotFound        = errors.New("the OAuth client was not found")
	ErrorProviderNotFound      = errors.New("the identity provider was not found")
	ErrorRedirectURLNotAllowed = errors.New("redirect URL is not allowed")
)

type Options struct {
	// An Issuer Identifier is a case-sensitive URL using the https scheme that contains scheme,
	// host, and optionally, port number and path components and no query or fragment components.
	Issuer string `json:"issuer,omitempty" yaml:"issuer,omitempty"`

	// RSA private key file used to sign the id token
	SignKey string `json:"signKey,omitempty" yaml:"signKey"`

	// Raw RSA private key. Base64 encoded PEM file
	SignKeyData string `json:"-,omitempty" yaml:"signKeyData"`

	// Register identity providers.
	IdentityProviders []IdentityProviderOptions `json:"identityProviders,omitempty" yaml:"identityProviders,omitempty"`

	// Register additional OAuth clients.
	Clients []Client `json:"clients,omitempty" yaml:"clients,omitempty"`

	// AccessTokenMaxAgeSeconds  control the lifetime of access tokens. The default lifetime is 24 hours.
	// 0 means no expiration.
	AccessTokenMaxAge time.Duration `json:"accessTokenMaxAge" yaml:"accessTokenMaxAge"`

	// Inactivity timeout for tokens
	// The value represents the maximum amount of time that can occur between
	// consecutive uses of the token. Tokens become invalid if they are not
	// used within this temporal window. The user will need to acquire a new
	// token to regain access once a token times out.
	// This value needs to be set only if the default set in configuration is
	// not appropriate for this client. Valid values are:
	// - 0: Tokens for this client never time out
	// - X: Tokens time out if there is no activity
	// The current minimum allowed value for X is 5 minutes
	AccessTokenInactivityTimeout time.Duration `json:"accessTokenInactivityTimeout" yaml:"accessTokenInactivityTimeout"`
}

// DynamicOptions accept dynamic configuration, the type of key MUST be string
type DynamicOptions map[string]interface{}

func (o DynamicOptions) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(desensitize(o))
	return data, err
}

var (
	sensitiveKeys = [...]string{"password", "secret"}
)

// isSensitiveData returns whether the input string contains sensitive information
func isSensitiveData(key string) bool {
	for _, v := range sensitiveKeys {
		if strings.Contains(strings.ToLower(key), v) {
			return true
		}
	}
	return false
}

// desensitize returns the desensitized data
func desensitize(data map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})
	for k, v := range data {
		if isSensitiveData(k) {
			continue
		}
		switch v.(type) {
		case map[interface{}]interface{}:
			output[k] = desensitize(convert(v.(map[interface{}]interface{})))
		default:
			output[k] = v
		}
	}
	return output
}

// convert returns formatted data
func convert(m map[interface{}]interface{}) map[string]interface{} {
	output := make(map[string]interface{})
	for k, v := range m {
		switch k.(type) {
		case string:
			output[k.(string)] = v
		}
	}
	return output
}

type IdentityProviderOptions struct {
	// The provider name.
	Name string `json:"name" yaml:"name"`

	// Defines how new identities are mapped to users when they login. Allowed values are:
	//  - auto:   The default value.The user will automatically create and mapping when login successful.
	//            Fails if a user with that user name is already mapped to another identity.
	//  - lookup: Looks up an existing identity, user identity mapping, and user, but does not automatically
	//            provision users or identities. Using this method requires you to manually provision users.
	//  - mixed:  A user entity can be mapped with multiple identifyProvider.
	MappingMethod MappingMethod `json:"mappingMethod" yaml:"mappingMethod"`

	// DisableLoginConfirmation means that when the user login successfully,
	// reconfirm the account information is not required.
	// Username from IDP must math [a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	DisableLoginConfirmation bool `json:"disableLoginConfirmation" yaml:"disableLoginConfirmation"`

	// The type of identify provider
	// OpenIDIdentityProvider LDAPIdentityProvider GitHubIdentityProvider
	Type string `json:"type" yaml:"type"`

	// The options of identify provider
	Provider DynamicOptions `json:"provider" yaml:"provider"`
}

type Token struct {
	// AccessToken is the token that authorizes and authenticates
	// the requests.
	AccessToken string `json:"access_token"`

	// TokenType is the type of token.
	// The Type method returns either this or "Bearer", the default.
	TokenType string `json:"token_type,omitempty"`

	// RefreshToken is a token that's used by the application
	// (as opposed to the user) to refresh the access token
	// if it expires.
	RefreshToken string `json:"refresh_token,omitempty"`

	// ID Token value associated with the authenticated session.
	IDToken string `json:"id_token,omitempty"`

	// ExpiresIn is the optional expiration second of the access token.
	ExpiresIn int `json:"expires_in,omitempty"`
}

type Client struct {
	// The name of the OAuth client is used as the client_id parameter when making requests to <master>/oauth/authorize
	// and <master>/oauth/token.
	Name string `json:"name" yaml:"name,omitempty"`

	// Secret is the unique secret associated with a client
	Secret string `json:"-" yaml:"secret,omitempty"`

	// RespondWithChallenges indicates whether the client wants authentication needed responses made
	// in the form of challenges instead of redirects
	RespondWithChallenges bool `json:"respondWithChallenges,omitempty" yaml:"respondWithChallenges,omitempty"`

	// RedirectURIs is the valid redirection URIs associated with a client
	RedirectURIs []string `json:"redirectURIs,omitempty" yaml:"redirectURIs,omitempty"`

	// GrantMethod determines how to handle grants for this client. If no method is provided, the
	// cluster default grant handling method will be used. Valid grant handling methods are:
	//  - auto:   always approves grant requests, useful for trusted clients
	//  - prompt: prompts the end user for approval of grant requests, useful for third-party clients
	//  - deny:   always denies grant requests, useful for black-listed clients
	GrantMethod GrantHandlerType `json:"grantMethod,omitempty" yaml:"grantMethod,omitempty"`

	// ScopeRestrictions describes which scopes this client can request.  Each requested scope
	// is checked against each restriction.  If any restriction matches, then the scope is allowed.
	// If no restriction matches, then the scope is denied.
	ScopeRestrictions []string `json:"scopeRestrictions,omitempty" yaml:"scopeRestrictions,omitempty"`

	// AccessTokenMaxAge overrides the default access token max age for tokens granted to this client.
	AccessTokenMaxAge *time.Duration `json:"accessTokenMaxAge,omitempty" yaml:"accessTokenMaxAge,omitempty"`

	// AccessTokenInactivityTimeout overrides the default token
	// inactivity timeout for tokens granted to this client.
	AccessTokenInactivityTimeout *time.Duration `json:"accessTokenInactivityTimeout,omitempty" yaml:"accessTokenInactivityTimeout,omitempty"`
}

func (o *Options) IdentityProviderOptions(name string) (*IdentityProviderOptions, error) {
	for _, found := range o.IdentityProviders {
		if found.Name == name {
			return &found, nil
		}
	}
	return nil, ErrorProviderNotFound
}

func NewOptions() *Options {
	return &Options{
		Issuer:                       DefaultIssuer,
		IdentityProviders:            make([]IdentityProviderOptions, 0),
		Clients:                      make([]Client, 0),
		AccessTokenMaxAge:            time.Hour * 2,
		AccessTokenInactivityTimeout: time.Hour * 2,
	}
}
