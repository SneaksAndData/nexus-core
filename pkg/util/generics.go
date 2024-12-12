package util

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"slices"
)

type deepCloneable[T corev1.Secret | corev1.ConfigMap] interface {
	DeepCopy() *T
	GetOwnerReferences() []metav1.OwnerReference
	SetOwnerReferences(references []metav1.OwnerReference)
	GetNamespace() string
}

func removeOwner[T corev1.Secret | corev1.ConfigMap](ctx context.Context, obj deepCloneable[T], ownerUID types.UID, k8sApiRef kubernetes.Interface, fieldManager string) error {
	updateOwners := func(obj interface{}, newOwners []metav1.OwnerReference) error {
		switch obj.(type) {
		case *corev1.Secret:
			objRef := obj.(*corev1.Secret)
			objRef.SetOwnerReferences(newOwners)
			_, err := k8sApiRef.CoreV1().Secrets(objRef.GetNamespace()).Update(ctx, objRef, metav1.UpdateOptions{FieldManager: fieldManager})
			if err != nil {
				return err
			}
		case *corev1.ConfigMap:
			objRef := obj.(*corev1.ConfigMap)
			objRef.SetOwnerReferences(newOwners)
			_, err := k8sApiRef.CoreV1().ConfigMaps(objRef.GetNamespace()).Update(ctx, objRef, metav1.UpdateOptions{FieldManager: fieldManager})
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown type: %T", obj)
		}

		return nil
	}

	objCopy := obj.DeepCopy()
	newReferences := []metav1.OwnerReference{}
	for _, ownerRef := range obj.GetOwnerReferences() {
		if ownerRef.UID != ownerUID {
			newReferences = append(newReferences, ownerRef)
		}
	}

	return updateOwners(objCopy, newReferences)
}

func GetConfigResolverDiff(configResolver func() []string, otherConfigResolver func() []string) []string {
	result := make([]string, 0)
	thisConfigs := configResolver()
	otherConfigs := otherConfigResolver()

	slices.Sort(thisConfigs)
	slices.Sort(otherConfigs)

	if !reflect.DeepEqual(thisConfigs, otherConfigs) {
		for _, configName := range thisConfigs {
			if !slices.Contains(otherConfigs, configName) {
				result = append(result, configName)
			}
		}
	}

	return result
}
