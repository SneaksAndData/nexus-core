package util

import (
	"context"
	"encoding/json"
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

func DeepClone[T any](obj T) (*T, error) {
	serialized, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	cloned := new(T)
	if err = json.Unmarshal(serialized, *cloned); err != nil {
		return nil, err
	}

	return cloned, nil
}

func Coalesce[T any](v1 T, v2 T) T {
	if v1 == nil {
		return v2
	}

	return v1
}

func CoalesceCollection[T any, C []T | map[string]T](v1 C, v2 C) C {
	if v1 == nil || len(v1) == 0 {
		return v2
	}

	return v1
}
