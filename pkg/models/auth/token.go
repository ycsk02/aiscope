package auth

import (
	"errors"
	"fmt"
	"time"

	"aiscope/pkg/apiserver/authentication"

	"k8s.io/klog/v2"

	"aiscope/pkg/apiserver/authentication/token"
	"aiscope/pkg/simple/client/cache"
)

// TokenManagementInterface Cache issued token, support revocation of tokens after issuance
type TokenManagementInterface interface {
	// Verify the given token and returns token.VerifiedResponse
	Verify(token string) (*token.VerifiedResponse, error)
	// IssueTo issue a token for the specified user
	IssueTo(request *token.IssueRequest) (string, error)
	// Revoke revoke the specified token
	Revoke(token string) error
	// RevokeAllUserTokens revoke all user tokens
	RevokeAllUserTokens(username string) error
	// Keys hold encryption and signing keys.
	Keys() *token.Keys
}

type tokenOperator struct {
	issuer  token.Issuer
	options *authentication.Options
	cache   cache.Interface
}

func (t tokenOperator) Revoke(token string) error {
	pattern := fmt.Sprintf("aiscope:user:*:token:%s", token)
	if keys, err := t.cache.Keys(pattern); err != nil {
		klog.Error(err)
		return err
	} else if len(keys) > 0 {
		if err := t.cache.Del(keys...); err != nil {
			klog.Error(err)
			return err
		}
	}
	return nil
}

func NewTokenOperator(cache cache.Interface, issuer token.Issuer, options *authentication.Options) TokenManagementInterface {
	operator := &tokenOperator{
		issuer:  issuer,
		options: options,
		cache:   cache,
	}
	return operator
}

func (t *tokenOperator) Verify(tokenStr string) (*token.VerifiedResponse, error) {
	response, err := t.issuer.Verify(tokenStr)
	if err != nil {
		return nil, err
	}
	if t.options.OAuthOptions.AccessTokenMaxAge == 0 ||
		response.TokenType == token.StaticToken {
		return response, nil
	}
	if err := t.tokenCacheValidate(response.User.GetName(), tokenStr); err != nil {
		return nil, err
	}
	return response, nil
}

func (t *tokenOperator) IssueTo(request *token.IssueRequest) (string, error) {
	tokenStr, err := t.issuer.IssueTo(request)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	if request.ExpiresIn > 0 {
		if err = t.cacheToken(request.User.GetName(), tokenStr, request.ExpiresIn); err != nil {
			klog.Error(err)
			return "", err
		}
	}
	return tokenStr, nil
}

// RevokeAllUserTokens revoke all user tokens in the cache
func (t *tokenOperator) RevokeAllUserTokens(username string) error {
	pattern := fmt.Sprintf("aiscope:user:%s:token:*", username)
	if keys, err := t.cache.Keys(pattern); err != nil {
		klog.Error(err)
		return err
	} else if len(keys) > 0 {
		if err := t.cache.Del(keys...); err != nil {
			klog.Error(err)
			return err
		}
	}
	return nil
}

func (t *tokenOperator) Keys() *token.Keys {
	return t.issuer.Keys()
}

// tokenCacheValidate verify that the token is in the cache
func (t *tokenOperator) tokenCacheValidate(username, token string) error {
	key := fmt.Sprintf("aiscope:user:%s:token:%s", username, token)
	if exist, err := t.cache.Exists(key); err != nil {
		return err
	} else if !exist {
		err = errors.New("token not found in cache")
		klog.V(4).Info(fmt.Errorf("%s: %s", err, token))
		return err
	}
	return nil
}

// cacheToken cache the token for a period of time
func (t *tokenOperator) cacheToken(username, token string, duration time.Duration) error {
	key := fmt.Sprintf("aiscope:user:%s:token:%s", username, token)
	if err := t.cache.Set(key, token, duration); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}
