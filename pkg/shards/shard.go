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
	kubernetesClientSet kubernetes.Interface
	nexusClientSet      clientset.Interface
	// SecretLister is a Secret lister in this Shard
	SecretLister  corelisters.SecretLister
	SecretsSynced cache.InformerSynced

	// ConfigMapLister is a ConfigMap lister in this Shard
	ConfigMapLister  corelisters.ConfigMapLister
	ConfigMapsSynced cache.InformerSynced

	// TemplateLister is a NexusAlgorithmTemplate lister in this Shard
	TemplateLister nexuslisters.NexusAlgorithmTemplateLister
	TemplateSynced cache.InformerSynced

	// WorkgroupLister is a NexusAlgorithmWorkgroup list in this Shard
	WorkgroupLister nexuslisters.NexusAlgorithmWorkgroupLister
	WorkgroupSynced cache.InformerSynced
}

// NewShard creates a new Shard instance. File name in *kubeConfigPath* will be used as the Shard's name
// in case of more than a single Shard make sure their kubeconfig files are named differently.
func NewShard(
	ownerName string,
	name string,
	kubeClient kubernetes.Interface,
	nexusClient clientset.Interface,

	templateInformer nexusinformers.NexusAlgorithmTemplateInformer,
	workgroupInformer nexusinformers.NexusAlgorithmWorkgroupInformer,
	secretInformer coreinformers.SecretInformer,
	configmapInformer coreinformers.ConfigMapInformer,
) *Shard {

	return &Shard{
		OwnerName:           ownerName,
		Name:                name,
		kubernetesClientSet: kubeClient,
		nexusClientSet:      nexusClient,

		SecretLister:  secretInformer.Lister(),
		SecretsSynced: secretInformer.Informer().HasSynced,

		ConfigMapLister:  configmapInformer.Lister(),
		ConfigMapsSynced: configmapInformer.Informer().HasSynced,

		TemplateLister: templateInformer.Lister(),
		TemplateSynced: templateInformer.Informer().HasSynced,

		WorkgroupLister: workgroupInformer.Lister(),
		WorkgroupSynced: workgroupInformer.Informer().HasSynced,
	}
}

func (shard *Shard) GetReferenceLabels() map[string]string {
	return map[string]string{
		"science.sneaksanddata.com/controller-app":      "nexus-configuration-controller",
		"science.sneaksanddata.com/configuration-owner": shard.OwnerName,
	}
}

// CreateTemplate creates a new template resource using the provided spec
func (shard *Shard) CreateTemplate(templateName string, templateNamespace string, templateSpec v1.NexusAlgorithmSpec, fieldManager string) (*v1.NexusAlgorithmTemplate, error) {
	newTemplate := &v1.NexusAlgorithmTemplate{
		TypeMeta: metav1.TypeMeta{APIVersion: v1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      templateName,
			Namespace: templateNamespace,
			Labels:    shard.GetReferenceLabels(),
		},
		Spec: *templateSpec.DeepCopy(),
	}

	return shard.nexusClientSet.ScienceV1().NexusAlgorithmTemplates(templateNamespace).Create(context.TODO(), newTemplate, metav1.CreateOptions{FieldManager: fieldManager})
}

// UpdateTemplate updates the Template in this shard in case it drifts from the one in the controller cluster
func (shard *Shard) UpdateTemplate(template *v1.NexusAlgorithmTemplate, templateSpec v1.NexusAlgorithmSpec, fieldManager string) (*v1.NexusAlgorithmTemplate, error) {
	newTemplate := template.DeepCopy()
	newTemplate.Spec = *templateSpec.DeepCopy()

	return shard.nexusClientSet.ScienceV1().NexusAlgorithmTemplates(newTemplate.Namespace).Update(context.TODO(), newTemplate, metav1.UpdateOptions{FieldManager: fieldManager})
}

// DeleteTemplate removes the Template from this shard
func (shard *Shard) DeleteTemplate(template *v1.NexusAlgorithmTemplate) error {
	return shard.nexusClientSet.ScienceV1().NexusAlgorithmTemplates(template.Namespace).Delete(context.TODO(), template.Name, metav1.DeleteOptions{})
}

// CreateWorkgroup creates a new workgroup resource using the provided spec
func (shard *Shard) CreateWorkgroup(workgroupName string, workgroupNamespace string, workgroupSpec v1.NexusAlgorithmWorkgroupSpec, fieldManager string) (*v1.NexusAlgorithmWorkgroup, error) {
	newWorkgroup := &v1.NexusAlgorithmWorkgroup{
		TypeMeta: metav1.TypeMeta{APIVersion: v1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      workgroupName,
			Namespace: workgroupNamespace,
			Labels:    shard.GetReferenceLabels(),
		},
		Spec: *workgroupSpec.DeepCopy(),
	}

	return shard.nexusClientSet.ScienceV1().NexusAlgorithmWorkgroups(workgroupNamespace).Create(context.TODO(), newWorkgroup, metav1.CreateOptions{FieldManager: fieldManager})
}

// UpdateWorkgroup updates the workgroup in this shard in case it drifts from the one in the controller cluster
func (shard *Shard) UpdateWorkgroup(workgroup *v1.NexusAlgorithmWorkgroup, workgroupSpec v1.NexusAlgorithmWorkgroupSpec, fieldManager string) (*v1.NexusAlgorithmWorkgroup, error) {
	newWorkgroup := workgroup.DeepCopy()
	newWorkgroup.Spec = *workgroupSpec.DeepCopy()

	return shard.nexusClientSet.ScienceV1().NexusAlgorithmWorkgroups(newWorkgroup.Namespace).Update(context.TODO(), newWorkgroup, metav1.UpdateOptions{FieldManager: fieldManager})
}

// DeleteWorkgroup removes the Workgroup from this shard
func (shard *Shard) DeleteWorkgroup(workgroup *v1.NexusAlgorithmWorkgroup) error {
	return shard.nexusClientSet.ScienceV1().NexusAlgorithmWorkgroups(workgroup.Namespace).Delete(context.TODO(), workgroup.Name, metav1.DeleteOptions{})
}

// CreateSecret creates a new Secret for a NexusAlgorithmTemplate resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func (shard *Shard) CreateSecret(template *v1.NexusAlgorithmTemplate, secret *corev1.Secret, fieldManager string) (*corev1.Secret, error) {
	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Namespace: template.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: v1.SchemeGroupVersion.String(),
					Kind:       "NexusAlgorithmTemplate",
					Name:       template.Name,
					UID:        template.UID,
				},
			},
			Labels: shard.GetReferenceLabels(),
		},
		Data:       secret.Data,
		StringData: secret.StringData,
	}

	return shard.kubernetesClientSet.CoreV1().Secrets(template.Namespace).Create(context.TODO(), newSecret, metav1.CreateOptions{FieldManager: fieldManager})
}

// CreateConfigMap creates a new ConfigMap for a NexusAlgorithmTemplate resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func (shard *Shard) CreateConfigMap(template *v1.NexusAlgorithmTemplate, configMap *corev1.ConfigMap, fieldManager string) (*corev1.ConfigMap, error) {
	newConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMap.Name,
			Namespace: template.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: v1.SchemeGroupVersion.String(),
					Kind:       "NexusAlgorithmTemplate",
					Name:       template.Name,
					UID:        template.UID,
				},
			},
			Labels: shard.GetReferenceLabels(),
		},
		Data: configMap.Data,
	}

	return shard.kubernetesClientSet.CoreV1().ConfigMaps(template.Namespace).Create(context.TODO(), newConfigMap, metav1.CreateOptions{FieldManager: fieldManager})
}

// UpdateSecret updates the secret with new content
func (shard *Shard) UpdateSecret(secret *corev1.Secret, newData map[string][]byte, newOwner *v1.NexusAlgorithmTemplate, fieldManager string) (*corev1.Secret, error) {
	updatedSecret := secret.DeepCopy()
	if newData != nil {
		updatedSecret.Data = newData
	}
	if newOwner != nil {
		updatedSecret.OwnerReferences = append(updatedSecret.OwnerReferences, metav1.OwnerReference{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "NexusAlgorithmTemplate",
			Name:       newOwner.Name,
			UID:        newOwner.UID,
		})
	}
	return shard.kubernetesClientSet.CoreV1().Secrets(updatedSecret.Namespace).Update(context.TODO(), updatedSecret, metav1.UpdateOptions{FieldManager: fieldManager})
}

// DereferenceSecret removes provided algorithm as the owner of the secret, and optionally removes the secret if it has no owners
func (shard *Shard) DereferenceSecret(secret *corev1.Secret, template *v1.NexusAlgorithmTemplate, fieldManager string) error {
	remainingOwners, err := util.RemoveOwner[corev1.Secret](context.TODO(), secret, template.UID, shard.kubernetesClientSet, fieldManager)
	if err != nil {
		return err
	}
	// delete the secret if there are no remaining owners
	if remainingOwners == 0 {
		return shard.kubernetesClientSet.CoreV1().Secrets(secret.Namespace).Delete(context.TODO(), secret.Name, metav1.DeleteOptions{})
	}

	return nil
}

// UpdateConfigMap updates the config map with new content
func (shard *Shard) UpdateConfigMap(configMap *corev1.ConfigMap, newData map[string]string, newOwner *v1.NexusAlgorithmTemplate, fieldManager string) (*corev1.ConfigMap, error) {
	updatedConfigMap := configMap.DeepCopy()
	if newData != nil {
		updatedConfigMap.Data = newData
	}
	if newOwner != nil {
		updatedConfigMap.OwnerReferences = append(updatedConfigMap.OwnerReferences, metav1.OwnerReference{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "NexusAlgorithmTemplate",
			Name:       newOwner.Name,
			UID:        newOwner.UID,
		})
	}
	return shard.kubernetesClientSet.CoreV1().ConfigMaps(updatedConfigMap.Namespace).Update(context.TODO(), updatedConfigMap, metav1.UpdateOptions{FieldManager: fieldManager})
}

// DereferenceConfigMap removes provided algorithm as the owner of the secret, and optionally removes the secret if it has no owners
func (shard *Shard) DereferenceConfigMap(configMap *corev1.ConfigMap, template *v1.NexusAlgorithmTemplate, fieldManager string) error {
	remainingOwners, err := util.RemoveOwner[corev1.ConfigMap](context.TODO(), configMap, template.UID, shard.kubernetesClientSet, fieldManager)
	if err != nil {
		return err
	}
	// delete the secret if there are no remaining owners
	if remainingOwners == 0 {
		return shard.kubernetesClientSet.CoreV1().ConfigMaps(configMap.Namespace).Delete(context.TODO(), configMap.Name, metav1.DeleteOptions{})
	}

	return nil
}
