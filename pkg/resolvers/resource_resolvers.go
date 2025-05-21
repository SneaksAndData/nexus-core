package resolvers

import (
	"context"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	nexusclientset "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetWorkgroupByRef reads NexusWorkgroup definition using the reference from the template
func GetWorkgroupByRef(template *v1.NexusAlgorithmTemplate, client nexusclientset.Interface, resourceNamespace string) (*v1.NexusAlgorithmWorkgroupSpec, error) {
	workgroup, err := client.ScienceV1().NexusAlgorithmWorkgroups(resourceNamespace).Get(context.TODO(), template.Spec.WorkgroupRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &workgroup.Spec, err
}

func isNexusRun(objMeta *metav1.ObjectMeta) bool {
	if !metav1.HasLabel(*objMeta, models.NexusComponentLabel) {
		return false
	}

	return objMeta.Labels[models.NexusComponentLabel] == models.JobLabelAlgorithmRun
}

// IsNexusRunEvent checks if an API server event is related to a Nexus run.
func IsNexusRunEvent(event *corev1.Event, client kubernetes.Interface, resourceNamespace string) (bool, error) {
	// ignore anything outside Nexus execution scope
	if event.InvolvedObject.Namespace != resourceNamespace {
		return false, nil
	}

	switch event.InvolvedObject.Kind {
	case "Job":
		job, err := client.BatchV1().Jobs(resourceNamespace).Get(context.TODO(), event.InvolvedObject.Name, metav1.GetOptions{})

		if err != nil {
			return false, err
		}

		return isNexusRun(&job.ObjectMeta), nil
	case "Pod":
		pod, err := client.CoreV1().Pods(resourceNamespace).Get(context.TODO(), event.InvolvedObject.Name, metav1.GetOptions{})

		if err != nil {
			return false, err
		}

		return isNexusRun(&pod.ObjectMeta), nil
	default:
		return false, nil
	}
}
