package auth

import (
	iamv1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	"aiscope/pkg/apiserver/authentication"
	"aiscope/pkg/apiserver/authentication/identityprovider"
	"aiscope/pkg/apiserver/authentication/oauth"
	aiscope "aiscope/pkg/client/clientset/versioned"
	iamv1alpha2listers "aiscope/pkg/client/listers/iam/v1alpha2"
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authuser "k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"
	"net/http"
)

type oauthAuthenticator struct {
	aiClient   aiscope.Interface
	userGetter *userGetter
	options    *authentication.Options
}

func NewOAuthAuthenticator(aiClient aiscope.Interface,
	userLister iamv1alpha2listers.UserLister,
	options *authentication.Options) OAuthAuthenticator {
	authenticator := &oauthAuthenticator{
		aiClient:   aiClient,
		userGetter: &userGetter{userLister: userLister},
		options:    options,
	}
	return authenticator
}

func (o *oauthAuthenticator) Authenticate(_ context.Context, provider string, req *http.Request) (authuser.Info, string, error) {
	providerOptions, err := o.options.OAuthOptions.IdentityProviderOptions(provider)
	// identity provider not registered
	if err != nil {
		klog.Error(err)
		return nil, "", err
	}
	oauthIdentityProvider, err := identityprovider.GetOAuthProvider(providerOptions.Name)
	if err != nil {
		klog.Error(err)
		return nil, "", err
	}
	authenticated, err := oauthIdentityProvider.IdentityExchangeCallback(req)
	if err != nil {
		klog.Error(err)
		return nil, "", err
	}

	user, err := o.userGetter.findMappedUser(providerOptions.Name, authenticated.GetUserID())
	if user == nil && providerOptions.MappingMethod == oauth.MappingMethodLookup {
		klog.Error(err)
		return nil, "", err
	}

	// the user will automatically create and mapping when login successful.
	if user == nil && providerOptions.MappingMethod == oauth.MappingMethodAuto {
		if !providerOptions.DisableLoginConfirmation {
			return preRegistrationUser(providerOptions.Name, authenticated), providerOptions.Name, nil
		}
		user, err = o.aiClient.IamV1alpha2().Users().Create(context.Background(), mappedUser(providerOptions.Name, authenticated), metav1.CreateOptions{})
		if err != nil {
			return nil, providerOptions.Name, err
		}
	}

	if user != nil {
		return &authuser.DefaultInfo{Name: user.GetName()}, providerOptions.Name, nil
	}

	return nil, "", errors.NewNotFound(iamv1alpha2.Resource("user"), authenticated.GetUsername())
}
