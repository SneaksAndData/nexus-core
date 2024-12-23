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

// FakeMachineLearningAlgorithms implements MachineLearningAlgorithmInterface
type FakeMachineLearningAlgorithms struct {
	Fake *FakeScienceV1
	ns   string
}

var machinelearningalgorithmsResource = v1.SchemeGroupVersion.WithResource("machinelearningalgorithms")

var machinelearningalgorithmsKind = v1.SchemeGroupVersion.WithKind("MachineLearningAlgorithm")

// Get takes name of the machineLearningAlgorithm, and returns the corresponding machineLearningAlgorithm object, and an error if there is any.
func (c *FakeMachineLearningAlgorithms) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.MachineLearningAlgorithm, err error) {
	emptyResult := &v1.MachineLearningAlgorithm{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(machinelearningalgorithmsResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.MachineLearningAlgorithm), err
}

// List takes label and field selectors, and returns the list of MachineLearningAlgorithms that match those selectors.
func (c *FakeMachineLearningAlgorithms) List(ctx context.Context, opts metav1.ListOptions) (result *v1.MachineLearningAlgorithmList, err error) {
	emptyResult := &v1.MachineLearningAlgorithmList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(machinelearningalgorithmsResource, machinelearningalgorithmsKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.MachineLearningAlgorithmList{ListMeta: obj.(*v1.MachineLearningAlgorithmList).ListMeta}
	for _, item := range obj.(*v1.MachineLearningAlgorithmList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested machineLearningAlgorithms.
func (c *FakeMachineLearningAlgorithms) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(machinelearningalgorithmsResource, c.ns, opts))

}

// Create takes the representation of a machineLearningAlgorithm and creates it.  Returns the server's representation of the machineLearningAlgorithm, and an error, if there is any.
func (c *FakeMachineLearningAlgorithms) Create(ctx context.Context, machineLearningAlgorithm *v1.MachineLearningAlgorithm, opts metav1.CreateOptions) (result *v1.MachineLearningAlgorithm, err error) {
	emptyResult := &v1.MachineLearningAlgorithm{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(machinelearningalgorithmsResource, c.ns, machineLearningAlgorithm, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.MachineLearningAlgorithm), err
}

// Update takes the representation of a machineLearningAlgorithm and updates it. Returns the server's representation of the machineLearningAlgorithm, and an error, if there is any.
func (c *FakeMachineLearningAlgorithms) Update(ctx context.Context, machineLearningAlgorithm *v1.MachineLearningAlgorithm, opts metav1.UpdateOptions) (result *v1.MachineLearningAlgorithm, err error) {
	emptyResult := &v1.MachineLearningAlgorithm{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(machinelearningalgorithmsResource, c.ns, machineLearningAlgorithm, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.MachineLearningAlgorithm), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMachineLearningAlgorithms) UpdateStatus(ctx context.Context, machineLearningAlgorithm *v1.MachineLearningAlgorithm, opts metav1.UpdateOptions) (result *v1.MachineLearningAlgorithm, err error) {
	emptyResult := &v1.MachineLearningAlgorithm{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(machinelearningalgorithmsResource, "status", c.ns, machineLearningAlgorithm, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.MachineLearningAlgorithm), err
}

// Delete takes name of the machineLearningAlgorithm and deletes it. Returns an error if one occurs.
func (c *FakeMachineLearningAlgorithms) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(machinelearningalgorithmsResource, c.ns, name, opts), &v1.MachineLearningAlgorithm{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMachineLearningAlgorithms) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(machinelearningalgorithmsResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.MachineLearningAlgorithmList{})
	return err
}

// Patch applies the patch and returns the patched machineLearningAlgorithm.
func (c *FakeMachineLearningAlgorithms) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.MachineLearningAlgorithm, err error) {
	emptyResult := &v1.MachineLearningAlgorithm{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(machinelearningalgorithmsResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.MachineLearningAlgorithm), err
}
