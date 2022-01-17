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
	v1alpha2 "aiscope/pkg/apis/iam/v1alpha2"
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeGroupBindings implements GroupBindingInterface
type FakeGroupBindings struct {
	Fake *FakeIamV1alpha2
}

var groupbindingsResource = schema.GroupVersionResource{Group: "iam", Version: "v1alpha2", Resource: "groupbindings"}

var groupbindingsKind = schema.GroupVersionKind{Group: "iam", Version: "v1alpha2", Kind: "GroupBinding"}

// Get takes name of the groupBinding, and returns the corresponding groupBinding object, and an error if there is any.
func (c *FakeGroupBindings) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.GroupBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(groupbindingsResource, name), &v1alpha2.GroupBinding{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.GroupBinding), err
}

// List takes label and field selectors, and returns the list of GroupBindings that match those selectors.
func (c *FakeGroupBindings) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.GroupBindingList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(groupbindingsResource, groupbindingsKind, opts), &v1alpha2.GroupBindingList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha2.GroupBindingList{ListMeta: obj.(*v1alpha2.GroupBindingList).ListMeta}
	for _, item := range obj.(*v1alpha2.GroupBindingList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested groupBindings.
func (c *FakeGroupBindings) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(groupbindingsResource, opts))
}

// Create takes the representation of a groupBinding and creates it.  Returns the server's representation of the groupBinding, and an error, if there is any.
func (c *FakeGroupBindings) Create(ctx context.Context, groupBinding *v1alpha2.GroupBinding, opts v1.CreateOptions) (result *v1alpha2.GroupBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(groupbindingsResource, groupBinding), &v1alpha2.GroupBinding{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.GroupBinding), err
}

// Update takes the representation of a groupBinding and updates it. Returns the server's representation of the groupBinding, and an error, if there is any.
func (c *FakeGroupBindings) Update(ctx context.Context, groupBinding *v1alpha2.GroupBinding, opts v1.UpdateOptions) (result *v1alpha2.GroupBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(groupbindingsResource, groupBinding), &v1alpha2.GroupBinding{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.GroupBinding), err
}

// Delete takes name of the groupBinding and deletes it. Returns an error if one occurs.
func (c *FakeGroupBindings) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(groupbindingsResource, name), &v1alpha2.GroupBinding{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeGroupBindings) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(groupbindingsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha2.GroupBindingList{})
	return err
}

// Patch applies the patch and returns the patched groupBinding.
func (c *FakeGroupBindings) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.GroupBinding, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(groupbindingsResource, name, pt, data, subresources...), &v1alpha2.GroupBinding{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha2.GroupBinding), err
}
