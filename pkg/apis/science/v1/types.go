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

// MachineLearningAlgorithm is a specification for a MachineLearningAlgorithm resource
type MachineLearningAlgorithm struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineLearningAlgorithmSpec   `json:"spec"`
	Status MachineLearningAlgorithmStatus `json:"status"`
}

// MachineLearningAlgorithmSpec is the spec for a MachineLearningAlgorithm resource
type MachineLearningAlgorithmSpec struct {
	ImageRegistry        string                 `json:"imageRegistry"`
	ImageRepository      string                 `json:"imageRepository"`
	ImageTag             string                 `json:"imageTag"`
	DeadlineSeconds      *int32                 `json:"deadlineSeconds,omitempty"`
	MaximumRetries       *int32                 `json:"maximumRetries,omitempty"`
	Env                  []corev1.EnvVar        `json:"env,omitempty"`
	EnvFrom              []corev1.EnvFromSource `json:"envFrom,omitempty"`
	CpuLimit             string                 `json:"cpuLimit"`
	MemoryLimit          string                 `json:"memoryLimit"`
	WorkgroupHost        string                 `json:"workgroupHost"`
	Workgroup            string                 `json:"workgroup"`
	AdditionalWorkgroups map[string]string      `json:"additionalWorkgroups,omitempty"`
	MonitoringParameters []string               `json:"monitoringParameters,omitempty"`
	CustomResources      map[string]string      `json:"customResources,omitempty"`
	SpeculativeAttempts  *int32                 `json:"speculativeAttempts,omitempty"`
	TransientExitCodes   []int32                `json:"transientExitCodes,omitempty"`
	FatalExitCodes       []int32                `json:"fatalExitCodes,omitempty"`
	Command              string                 `json:"command"`
	Args                 []string               `json:"args,omitempty"`
	MountDatadogSocket   bool                   `json:"mountDatadogSocket,omitempty"`
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

// MachineLearningAlgorithmStatus is the status for a MachineLearningAlgorithm resource
type MachineLearningAlgorithmStatus struct {
	SyncedSecrets        []string           `json:"syncedSecrets,omitempty"`
	SyncedConfigurations []string           `json:"syncedConfigurations,omitempty"`
	SyncedToClusters     []string           `json:"syncedToClusters,omitempty"`
	Conditions           []metav1.Condition `json:"conditions,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineLearningAlgorithmList is a list of MachineLearningAlgorithm resources
type MachineLearningAlgorithmList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MachineLearningAlgorithm `json:"items"`
}

// GetSecretNames retrieves a list of unique secret names for this MLA
func (mla *MachineLearningAlgorithm) GetSecretNames() []string {
	subset := map[string]bool{}
	for _, ref := range mla.Spec.EnvFrom {
		if ref.SecretRef != nil {
			subset[ref.SecretRef.Name] = true
		}
	}

	for _, ref := range mla.Spec.Env {
		if ref.ValueFrom != nil && ref.ValueFrom.SecretKeyRef != nil {
			subset[ref.ValueFrom.SecretKeyRef.Name] = true
		}
	}

	return slices.Collect(maps.Keys(subset))
}

// GetConfigMapNames retrieves a list of unique config names for this MLA
func (mla *MachineLearningAlgorithm) GetConfigMapNames() []string {
	subset := map[string]bool{}
	for _, ref := range mla.Spec.EnvFrom {
		if ref.ConfigMapRef != nil {
			subset[ref.ConfigMapRef.Name] = true
		}
	}

	for _, ref := range mla.Spec.Env {
		if ref.ValueFrom != nil && ref.ValueFrom.ConfigMapKeyRef != nil {
			subset[ref.ValueFrom.ConfigMapKeyRef.Name] = true
		}
	}

	return slices.Collect(maps.Keys(subset))
}

// Int32Ptr converts int32 type to int32 pointer type
// Method from sample-controller
func Int32Ptr(i int32) *int32 { return &i }

// GetSecretDiff resolves difference in secret references between two algorithms
func (mla *MachineLearningAlgorithm) GetSecretDiff(other *MachineLearningAlgorithm) []string {
	return util.GetConfigResolverDiff(mla.GetSecretNames, other.GetSecretNames)
}

// GetConfigmapDiff resolves difference in configmap references between two algorithms
func (mla *MachineLearningAlgorithm) GetConfigmapDiff(other *MachineLearningAlgorithm) []string {
	return util.GetConfigResolverDiff(mla.GetConfigMapNames, other.GetConfigMapNames)
}
