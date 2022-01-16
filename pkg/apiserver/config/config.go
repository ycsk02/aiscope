package config

import (
	"aiscope/pkg/simple/client/k8s"
	"aiscope/pkg/simple/client/ldap"
)

type Config struct {
	KubernetesOptions     *k8s.KubernetesOptions  `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty" mapstructure:"kubernetes"`
	LdapOptions           *ldap.Options           `json:"-,omitempty" yaml:"ldap,omitempty" mapstructure:"ldap"`
}

func New() *Config {
	return &Config{
		KubernetesOptions:		k8s.NewKubernetesOptions(),
		LdapOptions:			ldap.NewOptions(),
	}
}
