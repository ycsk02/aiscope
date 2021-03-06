/*
Copyright 2020 The AIScope Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

    https://vectorcloud.io
*/

// Code generated by client-gen. DO NOT EDIT.

package versioned

import (
	experimentv1alpha2 "aiscope/pkg/client/clientset/versioned/typed/experiment/v1alpha2"
	iamv1alpha2 "aiscope/pkg/client/clientset/versioned/typed/iam/v1alpha2"
	tenantv1alpha2 "aiscope/pkg/client/clientset/versioned/typed/tenant/v1alpha2"
	"fmt"

	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	ExperimentV1alpha2() experimentv1alpha2.ExperimentV1alpha2Interface
	IamV1alpha2() iamv1alpha2.IamV1alpha2Interface
	TenantV1alpha2() tenantv1alpha2.TenantV1alpha2Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	experimentV1alpha2 *experimentv1alpha2.ExperimentV1alpha2Client
	iamV1alpha2        *iamv1alpha2.IamV1alpha2Client
	tenantV1alpha2     *tenantv1alpha2.TenantV1alpha2Client
}

// ExperimentV1alpha2 retrieves the ExperimentV1alpha2Client
func (c *Clientset) ExperimentV1alpha2() experimentv1alpha2.ExperimentV1alpha2Interface {
	return c.experimentV1alpha2
}

// IamV1alpha2 retrieves the IamV1alpha2Client
func (c *Clientset) IamV1alpha2() iamv1alpha2.IamV1alpha2Interface {
	return c.iamV1alpha2
}

// TenantV1alpha2 retrieves the TenantV1alpha2Client
func (c *Clientset) TenantV1alpha2() tenantv1alpha2.TenantV1alpha2Interface {
	return c.tenantV1alpha2
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.experimentV1alpha2, err = experimentv1alpha2.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.iamV1alpha2, err = iamv1alpha2.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.tenantV1alpha2, err = tenantv1alpha2.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.experimentV1alpha2 = experimentv1alpha2.NewForConfigOrDie(c)
	cs.iamV1alpha2 = iamv1alpha2.NewForConfigOrDie(c)
	cs.tenantV1alpha2 = tenantv1alpha2.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.experimentV1alpha2 = experimentv1alpha2.New(c)
	cs.iamV1alpha2 = iamv1alpha2.New(c)
	cs.tenantV1alpha2 = tenantv1alpha2.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
