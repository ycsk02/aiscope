package options

import (
	"aiscope/pkg/simple/client/k8s"
	ldapclient "aiscope/pkg/simple/client/ldap"
	"k8s.io/client-go/tools/leaderelection"
	"time"
)

type AIScopeControllerManagerOptions struct {
	KubernetesOptions     *k8s.KubernetesOptions
	LdapOptions           *ldapclient.Options
	LeaderElect           bool
	LeaderElection        *leaderelection.LeaderElectionConfig
}

func NewAIScopeControllerManagerOptions() *AIScopeControllerManagerOptions {
	s := &AIScopeControllerManagerOptions{
		KubernetesOptions:     k8s.NewKubernetesOptions(),
		LdapOptions:           ldapclient.NewOptions(),
		LeaderElection: &leaderelection.LeaderElectionConfig{
			LeaseDuration: 30 * time.Second,
			RenewDeadline: 15 * time.Second,
			RetryPeriod:   5 * time.Second,
		},
		LeaderElect:         false,
	}

	return s
}
