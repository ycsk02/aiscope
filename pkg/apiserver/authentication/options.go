package authentication

import (
	"errors"
	"time"

	"github.com/spf13/pflag"

	"aiscope/pkg/apiserver/authentication/identityprovider"
	_ "aiscope/pkg/apiserver/authentication/identityprovider/ldap"
	"aiscope/pkg/apiserver/authentication/oauth"
)

type Options struct {
	// AuthenticateRateLimiter defines under which circumstances we will block user.
	// A user will be blocked if his/her failed login attempt reaches AuthenticateRateLimiterMaxTries in
	// AuthenticateRateLimiterDuration for about AuthenticateRateLimiterDuration. For example,
	//   AuthenticateRateLimiterMaxTries: 5
	//   AuthenticateRateLimiterDuration: 10m
	// A user will be blocked for 10m if he/she logins with incorrect credentials for at least 5 times in 10m.
	AuthenticateRateLimiterMaxTries int           `json:"authenticateRateLimiterMaxTries" yaml:"authenticateRateLimiterMaxTries"`
	AuthenticateRateLimiterDuration time.Duration `json:"authenticateRateLimiterDuration" yaml:"authenticateRateLimiterDuration"`
	// Token verification maximum time difference, default to 10s.
	// You should consider allowing a clock skew when checking the time-based values.
	// This should be values of a few seconds, and we don’t recommend using more than 30 seconds for this purpose,
	// as this would rather indicate problems with the server, rather than a common clock skew.
	MaximumClockSkew time.Duration `json:"maximumClockSkew" yaml:"maximumClockSkew"`
	// retention login history, records beyond this amount will be deleted
	LoginHistoryRetentionPeriod time.Duration `json:"loginHistoryRetentionPeriod" yaml:"loginHistoryRetentionPeriod"`
	// retention login history, records beyond this amount will be deleted
	// LoginHistoryMaximumEntries restricts for all aiscope accounts and must be greater than AuthenticateRateLimiterMaxTries
	LoginHistoryMaximumEntries int `json:"loginHistoryMaximumEntries" yaml:"loginHistoryMaximumEntries"`
	// allow multiple users login from different location at the same time
	MultipleLogin bool `json:"multipleLogin" yaml:"multipleLogin"`
	// secret to sign jwt token
	JwtSecret string `json:"-" yaml:"jwtSecret"`
	// OAuthOptions defines options needed for integrated oauth plugins
	OAuthOptions *oauth.Options `json:"oauthOptions" yaml:"oauthOptions"`
	// KubectlImage is the image address we use to create kubectl pod for users who have admin access to the cluster.
	KubectlImage string `json:"kubectlImage" yaml:"kubectlImage"`
}

func NewOptions() *Options {
	return &Options{
		AuthenticateRateLimiterMaxTries: 5,
		AuthenticateRateLimiterDuration: time.Minute * 30,
		MaximumClockSkew:                10 * time.Second,
		LoginHistoryRetentionPeriod:     time.Hour * 24 * 7,
		LoginHistoryMaximumEntries:      100,
		OAuthOptions:                    oauth.NewOptions(),
		MultipleLogin:                   false,
		JwtSecret:                       "",
		KubectlImage:                    "aiscope/kubectl:v1.0.0",
	}
}

func (options *Options) Validate() []error {
	var errs []error
	if len(options.JwtSecret) == 0 {
		errs = append(errs, errors.New("JWT secret MUST not be empty"))
	}
	if options.AuthenticateRateLimiterMaxTries > options.LoginHistoryMaximumEntries {
		errs = append(errs, errors.New("authenticateRateLimiterMaxTries MUST not be greater than loginHistoryMaximumEntries"))
	}
	if err := identityprovider.SetupWithOptions(options.OAuthOptions.IdentityProviders); err != nil {
		errs = append(errs, err)
	}
	return errs
}

func (options *Options) AddFlags(fs *pflag.FlagSet, s *Options) {
	fs.IntVar(&options.AuthenticateRateLimiterMaxTries, "authenticate-rate-limiter-max-retries", s.AuthenticateRateLimiterMaxTries, "")
	fs.DurationVar(&options.AuthenticateRateLimiterDuration, "authenticate-rate-limiter-duration", s.AuthenticateRateLimiterDuration, "")
	fs.BoolVar(&options.MultipleLogin, "multiple-login", s.MultipleLogin, "Allow multiple login with the same account, disable means only one user can login at the same time.")
	fs.StringVar(&options.JwtSecret, "jwt-secret", s.JwtSecret, "Secret to sign jwt token, must not be empty.")
	fs.DurationVar(&options.LoginHistoryRetentionPeriod, "login-history-retention-period", s.LoginHistoryRetentionPeriod, "login-history-retention-period defines how long login history should be kept.")
	fs.IntVar(&options.LoginHistoryMaximumEntries, "login-history-maximum-entries", s.LoginHistoryMaximumEntries, "login-history-maximum-entries defines how many entries of login history should be kept.")
	fs.DurationVar(&options.OAuthOptions.AccessTokenMaxAge, "access-token-max-age", s.OAuthOptions.AccessTokenMaxAge, "access-token-max-age control the lifetime of access tokens, 0 means no expiration.")
	fs.StringVar(&s.KubectlImage, "kubectl-image", s.KubectlImage, "Setup the image used by kubectl terminal pod")
	fs.DurationVar(&options.MaximumClockSkew, "maximum-clock-skew", s.MaximumClockSkew, "The maximum time difference between the system clocks of the ks-apiserver that issued a JWT and the ks-apiserver that verified the JWT.")
}
