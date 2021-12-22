package config

import "aiscope/pkg/simple/client/k8s"

type Config struct {
	KubernetesOptions     *k8s.KubernetesOptions  `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty" mapstructure:"kubernetes"`
}

func New() *Config {
	return &Config{
		KubernetesOptions:  k8s.NewKubernetesOptions(),
	}
}
