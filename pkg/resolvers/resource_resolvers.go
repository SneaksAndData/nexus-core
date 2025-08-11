package resolvers

import (
	"context"
	"fmt"
	nexusv1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	nexusclientset "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

// GetWorkgroupByRef reads NexusWorkgroup definition using the reference from the template
func GetWorkgroupByRef(template *nexusv1.NexusAlgorithmTemplate, client nexusclientset.Interface, resourceNamespace string) (*nexusv1.NexusAlgorithmWorkgroupSpec, error) { // coverage-ignore
	workgroup, err := client.ScienceV1().NexusAlgorithmWorkgroups(resourceNamespace).Get(context.TODO(), template.Spec.WorkgroupRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &workgroup.Spec, err
}

func isNexusRun(objMeta metav1.ObjectMeta) bool {
	if !metav1.HasLabel(objMeta, models.NexusComponentLabel) {
		return false
	}

	return objMeta.Labels[models.NexusComponentLabel] == models.JobLabelAlgorithmRun
}

// GetInformerObjectKey generates a key that can be used to look up a cached object representation from a respective informer
func GetInformerObjectKey(resourceNamespace string, resourceName string) string {
	return fmt.Sprintf("%s/%s", resourceNamespace, resourceName)
}

// GetCachedObject retrieves a cached resource from informer cache
func GetCachedObject[T any](objectName string, resourceNamespace string, informer cache.SharedIndexInformer) (*T, error) {
	resource, exists, err := informer.GetStore().GetByKey(GetInformerObjectKey(resourceNamespace, objectName))
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	return resource.(*T), nil
}

// IsNexusRunEvent checks if an API server event is related to a Nexus run.
func IsNexusRunEvent(event *corev1.Event, resourceNamespace string, informers map[string]cache.SharedIndexInformer) (bool, error) {
	// ignore anything outside Nexus execution scope
	if event.InvolvedObject.Namespace != resourceNamespace {
		return false, nil
	}

	switch event.InvolvedObject.Kind {
	case "Job":
		job, err := GetCachedObject[batchv1.Job](event.InvolvedObject.Name, resourceNamespace, informers[event.InvolvedObject.Kind])

		if job == nil {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return isNexusRun(job.ObjectMeta), nil
	case "Pod":
		pod, err := GetCachedObject[corev1.Pod](event.InvolvedObject.Name, resourceNamespace, informers[event.InvolvedObject.Kind])

		if pod == nil {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return isNexusRun(pod.ObjectMeta), nil
	default:
		return false, nil
	}
}
