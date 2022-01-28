package config

import (
	"aiscope/pkg/apiserver/authentication"
	"aiscope/pkg/simple/client/cache"
	"aiscope/pkg/simple/client/k8s"
	"aiscope/pkg/simple/client/ldap"
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

const (
	// DefaultConfigurationName is the default name of configuration
	defaultConfigurationName = "aiscope"

	// DefaultConfigurationPath the default location of the configuration file
	defaultConfigurationPath = "/etc/aiscope"
)

type Config struct {
	KubernetesOptions     *k8s.KubernetesOptions  `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty" mapstructure:"kubernetes"`
	LdapOptions           *ldap.Options           `json:"-,omitempty" yaml:"ldap,omitempty" mapstructure:"ldap"`
	RedisOptions          *cache.Options          `json:"redis,omitempty" yaml:"redis,omitempty" mapstructure:"redis"`
	AuthenticationOptions *authentication.Options `json:"authentication,omitempty" yaml:"authentication,omitempty" mapstructure:"authentication"`
}

func New() *Config {
	return &Config{
		KubernetesOptions:		k8s.NewKubernetesOptions(),
		LdapOptions:			ldap.NewOptions(),
		RedisOptions: 			cache.NewRedisOptions(),
		AuthenticationOptions:  authentication.NewOptions(),
	}
}

// TryLoadFromDisk loads configuration from default location after server startup
// return nil error if configuration file not exists
func TryLoadFromDisk() (*Config, error) {
	viper.SetConfigName(defaultConfigurationName)
	viper.AddConfigPath(defaultConfigurationPath)

	// Load from current working directory, only used for debugging
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Load from Environment variables
	viper.SetEnvPrefix("aiscope")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, err
		} else {
			return nil, fmt.Errorf("error parsing configuration file %s", err)
		}
	}

	conf := New()

	if err := viper.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
