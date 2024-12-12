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

func RemoveOwner[T corev1.Secret | corev1.ConfigMap](ctx context.Context, obj deepCloneable[T], ownerUID types.UID, k8sApiRef kubernetes.Interface, fieldManager string) (int, error) {
	updateOwners := func(obj interface{}, newOwners []metav1.OwnerReference) error {
		switch obj := obj.(type) {
		case *corev1.Secret:
			obj.SetOwnerReferences(newOwners)
			_, err := k8sApiRef.CoreV1().Secrets(obj.GetNamespace()).Update(ctx, obj, metav1.UpdateOptions{FieldManager: fieldManager})
			if err != nil {
				return err
			}

			return nil
		case *corev1.ConfigMap:
			obj.SetOwnerReferences(newOwners)
			_, err := k8sApiRef.CoreV1().ConfigMaps(obj.GetNamespace()).Update(ctx, obj, metav1.UpdateOptions{FieldManager: fieldManager})
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

	return len(newReferences), updateOwners(objCopy, newReferences)
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
