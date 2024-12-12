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

package shards

import (
	"context"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	clientset "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	nexusinformers "github.com/SneaksAndData/nexus-core/pkg/generated/informers/externalversions/science/v1"
	nexuslisters "github.com/SneaksAndData/nexus-core/pkg/generated/listers/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type Shard struct {
	OwnerName           string
	Name                string
	kubernetesclientset kubernetes.Interface
	nexusclientset      clientset.Interface
	// SecretLister is a Secret lister in this Shard
	SecretLister  corelisters.SecretLister
	SecretsSynced cache.InformerSynced

	// ConfigMapLister is a ConfigMap lister in this Shard
	ConfigMapLister  corelisters.ConfigMapLister
	ConfigMapsSynced cache.InformerSynced

	// MlaLister is a MachineLearningAlgorithm lister in this Shard
	MlaLister nexuslisters.MachineLearningAlgorithmLister
	MlaSynced cache.InformerSynced
}

// NewShard creates a new Shard instance. File name in *kubeConfigPath* will be used as the Shard's name
// in case of more than a single Shard make sure their kubeconfig files are named differently.
func NewShard(
	ownerName string,
	name string,
	kubeClient kubernetes.Interface,
	nexusClient clientset.Interface,

	mlainformer nexusinformers.MachineLearningAlgorithmInformer,
	secretinformer coreinformers.SecretInformer,
	configmapinformer coreinformers.ConfigMapInformer,
) *Shard {
	return &Shard{
		OwnerName:           ownerName,
		Name:                name,
		kubernetesclientset: kubeClient,
		nexusclientset:      nexusClient,

		SecretLister:  secretinformer.Lister(),
		SecretsSynced: secretinformer.Informer().HasSynced,

		ConfigMapLister:  configmapinformer.Lister(),
		ConfigMapsSynced: configmapinformer.Informer().HasSynced,

		MlaLister: mlainformer.Lister(),
		MlaSynced: mlainformer.Informer().HasSynced,
	}
}

func (shard *Shard) GetReferenceLabels() map[string]string {
	return map[string]string{
		"science.sneaksanddata.com/controller-app":      "nexus-configuration-controller",
		"science.sneaksanddata.com/configuration-owner": shard.OwnerName,
	}
}

func (shard *Shard) CreateMachineLearningAlgorithm(mlaName string, mlaNamespace string, mlaSpec v1.MachineLearningAlgorithmSpec, fieldManager string) (*v1.MachineLearningAlgorithm, error) {
	newMla := &v1.MachineLearningAlgorithm{
		TypeMeta: metav1.TypeMeta{APIVersion: v1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mlaName,
			Namespace: mlaNamespace,
			Labels:    shard.GetReferenceLabels(),
		},
		Spec: *mlaSpec.DeepCopy(),
	}

	return shard.nexusclientset.ScienceV1().MachineLearningAlgorithms(mlaNamespace).Create(context.TODO(), newMla, metav1.CreateOptions{FieldManager: fieldManager})
}

// UpdateMachineLearningAlgorithm updates the MLA in this shard in case it drifts from the one in the controller cluster
func (shard *Shard) UpdateMachineLearningAlgorithm(mla *v1.MachineLearningAlgorithm, mlaSpec v1.MachineLearningAlgorithmSpec, fieldManager string) (*v1.MachineLearningAlgorithm, error) {
	newMla := mla.DeepCopy()
	newMla.Spec = *mlaSpec.DeepCopy()

	return shard.nexusclientset.ScienceV1().MachineLearningAlgorithms(newMla.Namespace).Update(context.TODO(), newMla, metav1.UpdateOptions{FieldManager: fieldManager})
}

// DeleteMachineLearningAlgorithm removes the MLA from this shard
func (shard *Shard) DeleteMachineLearningAlgorithm(mla *v1.MachineLearningAlgorithm) error {
	return shard.nexusclientset.ScienceV1().MachineLearningAlgorithms(mla.Namespace).Delete(context.TODO(), mla.Name, metav1.DeleteOptions{})
}

// CreateSecret creates a new Secret for a MachineLearningAlgorithm resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func (shard *Shard) CreateSecret(mla *v1.MachineLearningAlgorithm, secret *corev1.Secret, fieldManager string) (*corev1.Secret, error) {
	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Namespace: mla.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: v1.SchemeGroupVersion.String(),
					Kind:       "MachineLearningAlgorithm",
					Name:       mla.Name,
					UID:        mla.UID,
				},
			},
			Labels: shard.GetReferenceLabels(),
		},
		Data:       secret.Data,
		StringData: secret.StringData,
	}

	return shard.kubernetesclientset.CoreV1().Secrets(mla.Namespace).Create(context.TODO(), newSecret, metav1.CreateOptions{FieldManager: fieldManager})
}

// CreateConfigMap creates a new ConfigMap for a MachineLearningAlgorithm resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func (shard *Shard) CreateConfigMap(mla *v1.MachineLearningAlgorithm, configMap *corev1.ConfigMap, fieldManager string) (*corev1.ConfigMap, error) {
	newConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMap.Name,
			Namespace: mla.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: v1.SchemeGroupVersion.String(),
					Kind:       "MachineLearningAlgorithm",
					Name:       mla.Name,
					UID:        mla.UID,
				},
			},
			Labels: shard.GetReferenceLabels(),
		},
		Data: configMap.Data,
	}

	return shard.kubernetesclientset.CoreV1().ConfigMaps(mla.Namespace).Create(context.TODO(), newConfigMap, metav1.CreateOptions{FieldManager: fieldManager})
}

// UpdateSecret updates the secret with new content
func (shard *Shard) UpdateSecret(secret *corev1.Secret, newData map[string][]byte, newOwner *v1.MachineLearningAlgorithm, fieldManager string) (*corev1.Secret, error) {
	updatedSecret := secret.DeepCopy()
	if newData != nil {
		updatedSecret.Data = newData
	}
	if newOwner != nil {
		updatedSecret.OwnerReferences = append(updatedSecret.OwnerReferences, metav1.OwnerReference{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "MachineLearningAlgorithm",
			Name:       newOwner.Name,
			UID:        newOwner.UID,
		})
	}
	return shard.kubernetesclientset.CoreV1().Secrets(updatedSecret.Namespace).Update(context.TODO(), updatedSecret, metav1.UpdateOptions{FieldManager: fieldManager})
}

// DereferenceSecret removes provided algorithm as the owner of the secret, and optionally removes the secret if it has no owners
func (shard *Shard) DereferenceSecret(secret *corev1.Secret, mla *v1.MachineLearningAlgorithm, fieldManager string) error {
	remainingOwners, err := util.RemoveOwner[corev1.Secret](context.TODO(), secret, mla.UID, shard.kubernetesclientset, fieldManager)
	if err != nil {
		return err
	}
	// delete the secret if there are no remaining owners
	if remainingOwners == 0 {
		return shard.kubernetesclientset.CoreV1().Secrets(secret.Namespace).Delete(context.TODO(), secret.Name, metav1.DeleteOptions{})
	}

	return nil
}

// UpdateConfigMap updates the config map with new content
func (shard *Shard) UpdateConfigMap(configMap *corev1.ConfigMap, newData map[string]string, newOwner *v1.MachineLearningAlgorithm, fieldManager string) (*corev1.ConfigMap, error) {
	updatedConfigMap := configMap.DeepCopy()
	if newData != nil {
		updatedConfigMap.Data = newData
	}
	if newOwner != nil {
		updatedConfigMap.OwnerReferences = append(updatedConfigMap.OwnerReferences, metav1.OwnerReference{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "MachineLearningAlgorithm",
			Name:       newOwner.Name,
			UID:        newOwner.UID,
		})
	}
	return shard.kubernetesclientset.CoreV1().ConfigMaps(updatedConfigMap.Namespace).Update(context.TODO(), updatedConfigMap, metav1.UpdateOptions{FieldManager: fieldManager})
}

// DereferenceConfigMap removes provided algorithm as the owner of the secret, and optionally removes the secret if it has no owners
func (shard *Shard) DereferenceConfigMap(configMap *corev1.ConfigMap, mla *v1.MachineLearningAlgorithm, fieldManager string) error {
	remainingOwners, err := util.RemoveOwner[corev1.ConfigMap](context.TODO(), configMap, mla.UID, shard.kubernetesclientset, fieldManager)
	if err != nil {
		return err
	}
	// delete the secret if there are no remaining owners
	if remainingOwners == 0 {
		return shard.kubernetesclientset.CoreV1().ConfigMaps(configMap.Namespace).Delete(context.TODO(), configMap.Name, metav1.DeleteOptions{})
	}

	return nil
}
