//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineLearningAlgorithm) DeepCopyInto(out *MachineLearningAlgorithm) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineLearningAlgorithm.
func (in *MachineLearningAlgorithm) DeepCopy() *MachineLearningAlgorithm {
	if in == nil {
		return nil
	}
	out := new(MachineLearningAlgorithm)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MachineLearningAlgorithm) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineLearningAlgorithmList) DeepCopyInto(out *MachineLearningAlgorithmList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MachineLearningAlgorithm, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineLearningAlgorithmList.
func (in *MachineLearningAlgorithmList) DeepCopy() *MachineLearningAlgorithmList {
	if in == nil {
		return nil
	}
	out := new(MachineLearningAlgorithmList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MachineLearningAlgorithmList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineLearningAlgorithmSpec) DeepCopyInto(out *MachineLearningAlgorithmSpec) {
	*out = *in
	if in.DeadlineSeconds != nil {
		in, out := &in.DeadlineSeconds, &out.DeadlineSeconds
		*out = new(int32)
		**out = **in
	}
	if in.MaximumRetries != nil {
		in, out := &in.MaximumRetries, &out.MaximumRetries
		*out = new(int32)
		**out = **in
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]corev1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.EnvFrom != nil {
		in, out := &in.EnvFrom, &out.EnvFrom
		*out = make([]corev1.EnvFromSource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AdditionalWorkgroups != nil {
		in, out := &in.AdditionalWorkgroups, &out.AdditionalWorkgroups
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.MonitoringParameters != nil {
		in, out := &in.MonitoringParameters, &out.MonitoringParameters
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.CustomResources != nil {
		in, out := &in.CustomResources, &out.CustomResources
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.SpeculativeAttempts != nil {
		in, out := &in.SpeculativeAttempts, &out.SpeculativeAttempts
		*out = new(int32)
		**out = **in
	}
	if in.TransientExitCodes != nil {
		in, out := &in.TransientExitCodes, &out.TransientExitCodes
		*out = make([]int32, len(*in))
		copy(*out, *in)
	}
	if in.FatalExitCodes != nil {
		in, out := &in.FatalExitCodes, &out.FatalExitCodes
		*out = make([]int32, len(*in))
		copy(*out, *in)
	}
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.MountDatadogSocket != nil {
		in, out := &in.MountDatadogSocket, &out.MountDatadogSocket
		*out = new(bool)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineLearningAlgorithmSpec.
func (in *MachineLearningAlgorithmSpec) DeepCopy() *MachineLearningAlgorithmSpec {
	if in == nil {
		return nil
	}
	out := new(MachineLearningAlgorithmSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineLearningAlgorithmStatus) DeepCopyInto(out *MachineLearningAlgorithmStatus) {
	*out = *in
	if in.SyncedSecrets != nil {
		in, out := &in.SyncedSecrets, &out.SyncedSecrets
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SyncedConfigurations != nil {
		in, out := &in.SyncedConfigurations, &out.SyncedConfigurations
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SyncedToClusters != nil {
		in, out := &in.SyncedToClusters, &out.SyncedToClusters
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineLearningAlgorithmStatus.
func (in *MachineLearningAlgorithmStatus) DeepCopy() *MachineLearningAlgorithmStatus {
	if in == nil {
		return nil
	}
	out := new(MachineLearningAlgorithmStatus)
	in.DeepCopyInto(out)
	return out
}
