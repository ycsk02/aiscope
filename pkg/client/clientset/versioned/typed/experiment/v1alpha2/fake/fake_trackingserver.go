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

package fake

import (
	v1alpha2 "aiscope/pkg/apis/experiment/v1alpha2"
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeTrackingServers implements TrackingServerInterface
type FakeTrackingServers struct {
	Fake *FakeExperimentV1alpha2
	ns   string
}

var trackingserversResource = schema.GroupVersionResource{Group: "experiment", Version: "v1alpha2", Resource: "trackingservers"}

var trackingserversKind = schema.GroupVersionKind{Group: "experiment", Version: "v1alpha2", Kind: "TrackingServer"}

// Get takes name of the trackingServer, and returns the corresponding trackingServer object, and an error if there is any.
func (c *FakeTrackingServers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.TrackingServer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(trackingserversResource, c.ns, name), &v1alpha2.TrackingServer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.TrackingServer), err
}

// List takes label and field selectors, and returns the list of TrackingServers that match those selectors.
func (c *FakeTrackingServers) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.TrackingServerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(trackingserversResource, trackingserversKind, c.ns, opts), &v1alpha2.TrackingServerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha2.TrackingServerList{ListMeta: obj.(*v1alpha2.TrackingServerList).ListMeta}
	for _, item := range obj.(*v1alpha2.TrackingServerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested trackingServers.
func (c *FakeTrackingServers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(trackingserversResource, c.ns, opts))

}

// Create takes the representation of a trackingServer and creates it.  Returns the server's representation of the trackingServer, and an error, if there is any.
func (c *FakeTrackingServers) Create(ctx context.Context, trackingServer *v1alpha2.TrackingServer, opts v1.CreateOptions) (result *v1alpha2.TrackingServer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(trackingserversResource, c.ns, trackingServer), &v1alpha2.TrackingServer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.TrackingServer), err
}

// Update takes the representation of a trackingServer and updates it. Returns the server's representation of the trackingServer, and an error, if there is any.
func (c *FakeTrackingServers) Update(ctx context.Context, trackingServer *v1alpha2.TrackingServer, opts v1.UpdateOptions) (result *v1alpha2.TrackingServer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(trackingserversResource, c.ns, trackingServer), &v1alpha2.TrackingServer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.TrackingServer), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeTrackingServers) UpdateStatus(ctx context.Context, trackingServer *v1alpha2.TrackingServer, opts v1.UpdateOptions) (*v1alpha2.TrackingServer, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(trackingserversResource, "status", c.ns, trackingServer), &v1alpha2.TrackingServer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.TrackingServer), err
}

// Delete takes name of the trackingServer and deletes it. Returns an error if one occurs.
func (c *FakeTrackingServers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(trackingserversResource, c.ns, name), &v1alpha2.TrackingServer{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeTrackingServers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(trackingserversResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha2.TrackingServerList{})
	return err
}

// Patch applies the patch and returns the patched trackingServer.
func (c *FakeTrackingServers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.TrackingServer, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(trackingserversResource, c.ns, name, pt, data, subresources...), &v1alpha2.TrackingServer{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.TrackingServer), err
}