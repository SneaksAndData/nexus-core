/*
Copyright 2024-2026 ECCO Data & AI Open-Source Project Maintainers.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeNexusAlgorithmWorkgroups implements NexusAlgorithmWorkgroupInterface
type FakeNexusAlgorithmWorkgroups struct {
	Fake *FakeScienceV1
	ns   string
}

var nexusalgorithmworkgroupsResource = v1.SchemeGroupVersion.WithResource("nexusalgorithmworkgroups")

var nexusalgorithmworkgroupsKind = v1.SchemeGroupVersion.WithKind("NexusAlgorithmWorkgroup")

// Get takes name of the nexusAlgorithmWorkgroup, and returns the corresponding nexusAlgorithmWorkgroup object, and an error if there is any.
func (c *FakeNexusAlgorithmWorkgroups) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.NexusAlgorithmWorkgroup, err error) {
	emptyResult := &v1.NexusAlgorithmWorkgroup{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(nexusalgorithmworkgroupsResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.NexusAlgorithmWorkgroup), err
}

// List takes label and field selectors, and returns the list of NexusAlgorithmWorkgroups that match those selectors.
func (c *FakeNexusAlgorithmWorkgroups) List(ctx context.Context, opts metav1.ListOptions) (result *v1.NexusAlgorithmWorkgroupList, err error) {
	emptyResult := &v1.NexusAlgorithmWorkgroupList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(nexusalgorithmworkgroupsResource, nexusalgorithmworkgroupsKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.NexusAlgorithmWorkgroupList{ListMeta: obj.(*v1.NexusAlgorithmWorkgroupList).ListMeta}
	for _, item := range obj.(*v1.NexusAlgorithmWorkgroupList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested nexusAlgorithmWorkgroups.
func (c *FakeNexusAlgorithmWorkgroups) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(nexusalgorithmworkgroupsResource, c.ns, opts))

}

// Create takes the representation of a nexusAlgorithmWorkgroup and creates it.  Returns the server's representation of the nexusAlgorithmWorkgroup, and an error, if there is any.
func (c *FakeNexusAlgorithmWorkgroups) Create(ctx context.Context, nexusAlgorithmWorkgroup *v1.NexusAlgorithmWorkgroup, opts metav1.CreateOptions) (result *v1.NexusAlgorithmWorkgroup, err error) {
	emptyResult := &v1.NexusAlgorithmWorkgroup{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(nexusalgorithmworkgroupsResource, c.ns, nexusAlgorithmWorkgroup, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.NexusAlgorithmWorkgroup), err
}

// Update takes the representation of a nexusAlgorithmWorkgroup and updates it. Returns the server's representation of the nexusAlgorithmWorkgroup, and an error, if there is any.
func (c *FakeNexusAlgorithmWorkgroups) Update(ctx context.Context, nexusAlgorithmWorkgroup *v1.NexusAlgorithmWorkgroup, opts metav1.UpdateOptions) (result *v1.NexusAlgorithmWorkgroup, err error) {
	emptyResult := &v1.NexusAlgorithmWorkgroup{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(nexusalgorithmworkgroupsResource, c.ns, nexusAlgorithmWorkgroup, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.NexusAlgorithmWorkgroup), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeNexusAlgorithmWorkgroups) UpdateStatus(ctx context.Context, nexusAlgorithmWorkgroup *v1.NexusAlgorithmWorkgroup, opts metav1.UpdateOptions) (result *v1.NexusAlgorithmWorkgroup, err error) {
	emptyResult := &v1.NexusAlgorithmWorkgroup{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(nexusalgorithmworkgroupsResource, "status", c.ns, nexusAlgorithmWorkgroup, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.NexusAlgorithmWorkgroup), err
}

// Delete takes name of the nexusAlgorithmWorkgroup and deletes it. Returns an error if one occurs.
func (c *FakeNexusAlgorithmWorkgroups) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(nexusalgorithmworkgroupsResource, c.ns, name, opts), &v1.NexusAlgorithmWorkgroup{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeNexusAlgorithmWorkgroups) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(nexusalgorithmworkgroupsResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.NexusAlgorithmWorkgroupList{})
	return err
}

// Patch applies the patch and returns the patched nexusAlgorithmWorkgroup.
func (c *FakeNexusAlgorithmWorkgroups) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.NexusAlgorithmWorkgroup, err error) {
	emptyResult := &v1.NexusAlgorithmWorkgroup{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(nexusalgorithmworkgroupsResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.NexusAlgorithmWorkgroup), err
}
