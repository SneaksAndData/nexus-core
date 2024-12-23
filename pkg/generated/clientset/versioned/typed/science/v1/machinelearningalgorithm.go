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

package v1

import (
	"context"

	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	scheme "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// MachineLearningAlgorithmsGetter has a method to return a MachineLearningAlgorithmInterface.
// A group's client should implement this interface.
type MachineLearningAlgorithmsGetter interface {
	MachineLearningAlgorithms(namespace string) MachineLearningAlgorithmInterface
}

// MachineLearningAlgorithmInterface has methods to work with MachineLearningAlgorithm resources.
type MachineLearningAlgorithmInterface interface {
	Create(ctx context.Context, machineLearningAlgorithm *v1.MachineLearningAlgorithm, opts metav1.CreateOptions) (*v1.MachineLearningAlgorithm, error)
	Update(ctx context.Context, machineLearningAlgorithm *v1.MachineLearningAlgorithm, opts metav1.UpdateOptions) (*v1.MachineLearningAlgorithm, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, machineLearningAlgorithm *v1.MachineLearningAlgorithm, opts metav1.UpdateOptions) (*v1.MachineLearningAlgorithm, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.MachineLearningAlgorithm, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.MachineLearningAlgorithmList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.MachineLearningAlgorithm, err error)
	MachineLearningAlgorithmExpansion
}

// machineLearningAlgorithms implements MachineLearningAlgorithmInterface
type machineLearningAlgorithms struct {
	*gentype.ClientWithList[*v1.MachineLearningAlgorithm, *v1.MachineLearningAlgorithmList]
}

// newMachineLearningAlgorithms returns a MachineLearningAlgorithms
func newMachineLearningAlgorithms(c *ScienceV1Client, namespace string) *machineLearningAlgorithms {
	return &machineLearningAlgorithms{
		gentype.NewClientWithList[*v1.MachineLearningAlgorithm, *v1.MachineLearningAlgorithmList](
			"machinelearningalgorithms",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1.MachineLearningAlgorithm { return &v1.MachineLearningAlgorithm{} },
			func() *v1.MachineLearningAlgorithmList { return &v1.MachineLearningAlgorithmList{} }),
	}
}
