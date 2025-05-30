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

// NexusAlgorithmTemplatesGetter has a method to return a NexusAlgorithmTemplateInterface.
// A group's client should implement this interface.
type NexusAlgorithmTemplatesGetter interface {
	NexusAlgorithmTemplates(namespace string) NexusAlgorithmTemplateInterface
}

// NexusAlgorithmTemplateInterface has methods to work with NexusAlgorithmTemplate resources.
type NexusAlgorithmTemplateInterface interface {
	Create(ctx context.Context, nexusAlgorithmTemplate *v1.NexusAlgorithmTemplate, opts metav1.CreateOptions) (*v1.NexusAlgorithmTemplate, error)
	Update(ctx context.Context, nexusAlgorithmTemplate *v1.NexusAlgorithmTemplate, opts metav1.UpdateOptions) (*v1.NexusAlgorithmTemplate, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, nexusAlgorithmTemplate *v1.NexusAlgorithmTemplate, opts metav1.UpdateOptions) (*v1.NexusAlgorithmTemplate, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.NexusAlgorithmTemplate, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.NexusAlgorithmTemplateList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.NexusAlgorithmTemplate, err error)
	NexusAlgorithmTemplateExpansion
}

// nexusAlgorithmTemplates implements NexusAlgorithmTemplateInterface
type nexusAlgorithmTemplates struct {
	*gentype.ClientWithList[*v1.NexusAlgorithmTemplate, *v1.NexusAlgorithmTemplateList]
}

// newNexusAlgorithmTemplates returns a NexusAlgorithmTemplates
func newNexusAlgorithmTemplates(c *ScienceV1Client, namespace string) *nexusAlgorithmTemplates {
	return &nexusAlgorithmTemplates{
		gentype.NewClientWithList[*v1.NexusAlgorithmTemplate, *v1.NexusAlgorithmTemplateList](
			"nexusalgorithmtemplates",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1.NexusAlgorithmTemplate { return &v1.NexusAlgorithmTemplate{} },
			func() *v1.NexusAlgorithmTemplateList { return &v1.NexusAlgorithmTemplateList{} }),
	}
}
