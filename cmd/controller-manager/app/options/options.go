package options

import (
	"aiscope/pkg/simple/client/k8s"
	"k8s.io/client-go/tools/leaderelection"
	"time"
)

type AIScopeControllerManagerOptions struct {
	KubernetesOptions     *k8s.KubernetesOptions
	LeaderElect           bool
	LeaderElection        *leaderelection.LeaderElectionConfig
}

func NewAIScopeControllerManagerOptions() *AIScopeControllerManagerOptions {
	s := &AIScopeControllerManagerOptions{
		KubernetesOptions:     k8s.NewKubernetesOptions(),
		LeaderElection: &leaderelection.LeaderElectionConfig{
			LeaseDuration: 30 * time.Second,
			RenewDeadline: 15 * time.Second,
			RetryPeriod:   5 * time.Second,
		},
		LeaderElect:         false,
	}

	return s
}
