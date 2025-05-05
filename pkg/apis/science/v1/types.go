/*
 * Copyright (c) 2024. ECCO Data & AI Open-Source Project Maintainers.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package v1

import (
	"github.com/SneaksAndData/nexus-core/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"maps"
	"slices"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NexusAlgorithmTemplate is a specification for an AI/ML batch application run: inference, training etc.
type NexusAlgorithmTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NexusAlgorithmSpec   `json:"spec"`
	Status NexusAlgorithmStatus `json:"status"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NexusAlgorithmWorkgroup specifies node tolerations for algorithm pods
type NexusAlgorithmWorkgroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NexusAlgorithmWorkgroupSpec   `json:"spec"`
	Status NexusAlgorithmWorkgroupStatus `json:"status"`
}

// NexusAlgorithmWorkgroupRef contains a reference to the workgroup
type NexusAlgorithmWorkgroupRef struct {
	Name  string `json:"name"`
	Group string `json:"group"`
	Kind  string `json:"kind"`
}

// NexusAlgorithmWorkgroupSpec is a spec for NexusAlgorithmWorkgroup resource
type NexusAlgorithmWorkgroupSpec struct {
	Description    string          `json:"description"`
	CapabilityTags map[string]bool `json:"capabilityTags"`

	Tolerations []corev1.Toleration `json:"tolerations"`
	Affinity    *corev1.Affinity    `json:"affinity"`
}

// NexusAlgorithmResources defines maximum compute resources that should be provisioned for the algorithm
type NexusAlgorithmResources struct {
	CpuLimit        string            `json:"cpuLimit"`
	MemoryLimit     string            `json:"memoryLimit"`
	CustomResources map[string]string `json:"customResources,omitempty"`
}

// NexusAlgorithmSubmissionBehaviour defines submission behaviour: shards or single cluster, plus a workgroup
type NexusAlgorithmSubmissionBehaviour struct {
	ShardClusters []string                    `json:"shardClusters,omitempty"`
	WorkgroupRef  *NexusAlgorithmWorkgroupRef `json:"workgroupRef"`
}

// NexusAlgorithmContainer provides container specification for each run
type NexusAlgorithmContainer struct {
	Image              string `json:"image"`
	Registry           string `json:"registry"`
	VersionTag         string `json:"versionTag"`
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// NexusAlgorithmRuntimeEnvironment defines environment configuration for each run
type NexusAlgorithmRuntimeEnvironment struct {
	EnvironmentVariables       []corev1.EnvVar        `json:"environmentVariables"`
	MappedEnvironmentVariables []corev1.EnvFromSource `json:"mappedEnvironmentVariables,omitempty"`
	Annotations                map[string]string      `json:"annotations,omitempty"`

	DeadlineSeconds *int32 `json:"deadlineSeconds,omitempty"`
	MaximumRetries  *int32 `json:"maximumRetries,omitempty"`
}

// NexusErrorHandlingBehaviour defines error handling behaviours for algorithm exit codes
type NexusErrorHandlingBehaviour struct {
	TransientExitCodes []int32 `json:"transientExitCodes,omitempty"`
	FatalExitCodes     []int32 `json:"fatalExitCodes,omitempty"`
}

// NexusDatadogIntegrationSettings defines settings for Nexus algorithms that use Datadog metrics and logging capabilities
type NexusDatadogIntegrationSettings struct {
	MountDatadogSocket *bool `json:"mountDatadogSocket,omitempty"`
}

// NexusAlgorithmSpec is the spec for a NexusAlgorithmTemplate resource
type NexusAlgorithmSpec struct {
	Container                  *NexusAlgorithmContainer           `json:"container"`
	ComputeResources           *NexusAlgorithmResources           `json:"computeResources,omitempty"`
	SubmissionBehaviour        *NexusAlgorithmSubmissionBehaviour `json:"submissionBehaviour,omitempty"`
	Command                    string                             `json:"command"`
	Args                       []string                           `json:"args,omitempty"`
	RuntimeEnvironment         *NexusAlgorithmRuntimeEnvironment  `json:"runtimeEnvironment,omitempty"`
	ErrorHandlingBehaviour     *NexusErrorHandlingBehaviour       `json:"errorHandlingBehaviour,omitempty"`
	DatadogIntegrationSettings *NexusDatadogIntegrationSettings   `json:"datadogIntegrationSettings,omitempty"`
}

func (spec *NexusAlgorithmSpec) Merge(other *NexusAlgorithmSpec) *NexusAlgorithmSpec {
	cloned := spec.DeepCopy()
	otherCloned := other.DeepCopy()

	cloned.Container = util.CoalescePointer(otherCloned.Container, cloned.Container)
	cloned.ComputeResources = util.CoalescePointer(otherCloned.ComputeResources, cloned.ComputeResources)
	cloned.SubmissionBehaviour = util.CoalescePointer(otherCloned.SubmissionBehaviour, cloned.SubmissionBehaviour)
	cloned.Command = util.CoalesceString(otherCloned.Command, cloned.Command)
	cloned.Args = util.CoalesceCollection[string](otherCloned.Args, cloned.Args)

	cloned.RuntimeEnvironment = util.CoalescePointer(otherCloned.RuntimeEnvironment, cloned.RuntimeEnvironment)
	cloned.ErrorHandlingBehaviour = util.CoalescePointer(otherCloned.ErrorHandlingBehaviour, cloned.ErrorHandlingBehaviour)

	cloned.DatadogIntegrationSettings = util.CoalescePointer(otherCloned.DatadogIntegrationSettings, cloned.DatadogIntegrationSettings)

	return cloned
}

// NewResourceReadyCondition creates a new condition indicating an overall Mla synchronisation success or failure
func NewResourceReadyCondition(transitionTime metav1.Time, status metav1.ConditionStatus, message string) *metav1.Condition {
	return &metav1.Condition{
		LastTransitionTime: transitionTime,
		Type:               "Ready",
		Status:             status,
		Reason:             "AlgorithmReady",
		Message:            message,
	}
}

// NexusAlgorithmWorkgroupStatus is the status for a NexusAlgorithmWorkgroup resource
type NexusAlgorithmWorkgroupStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// NexusAlgorithmStatus is the status for a NexusAlgorithmTemplate resource
type NexusAlgorithmStatus struct {
	SyncedSecrets        []string           `json:"syncedSecrets,omitempty"`
	SyncedConfigurations []string           `json:"syncedConfigurations,omitempty"`
	SyncedToClusters     []string           `json:"syncedToClusters,omitempty"`
	Conditions           []metav1.Condition `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NexusAlgorithmTemplateList is a list of NexusAlgorithmTemplate resources
type NexusAlgorithmTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []NexusAlgorithmTemplate `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NexusAlgorithmWorkgroupList is a list of NexusAlgorithmWorkgroup resources
type NexusAlgorithmWorkgroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []NexusAlgorithmWorkgroup `json:"items"`
}

// GetSecretNames retrieves a list of unique secret names for this MLA
func (template *NexusAlgorithmTemplate) GetSecretNames() []string {
	subset := map[string]bool{}
	for _, ref := range template.Spec.RuntimeEnvironment.MappedEnvironmentVariables {
		if ref.SecretRef != nil {
			subset[ref.SecretRef.Name] = true
		}
	}

	for _, ref := range template.Spec.RuntimeEnvironment.EnvironmentVariables {
		if ref.ValueFrom != nil && ref.ValueFrom.SecretKeyRef != nil {
			subset[ref.ValueFrom.SecretKeyRef.Name] = true
		}
	}

	return slices.Collect(maps.Keys(subset))
}

// GetConfigMapNames retrieves a list of unique config names for this MLA
func (template *NexusAlgorithmTemplate) GetConfigMapNames() []string {
	subset := map[string]bool{}
	for _, ref := range template.Spec.RuntimeEnvironment.MappedEnvironmentVariables {
		if ref.ConfigMapRef != nil {
			subset[ref.ConfigMapRef.Name] = true
		}
	}

	for _, ref := range template.Spec.RuntimeEnvironment.EnvironmentVariables {
		if ref.ValueFrom != nil && ref.ValueFrom.ConfigMapKeyRef != nil {
			subset[ref.ValueFrom.ConfigMapKeyRef.Name] = true
		}
	}

	return slices.Collect(maps.Keys(subset))
}

// GetSecretDiff resolves difference in secret references between two algorithms
func (template *NexusAlgorithmTemplate) GetSecretDiff(other *NexusAlgorithmTemplate) []string {
	return util.GetConfigResolverDiff(template.GetSecretNames, other.GetSecretNames)
}

// GetConfigmapDiff resolves difference in configmap references between two algorithms
func (template *NexusAlgorithmTemplate) GetConfigmapDiff(other *NexusAlgorithmTemplate) []string {
	return util.GetConfigResolverDiff(template.GetConfigMapNames, other.GetConfigMapNames)
}
