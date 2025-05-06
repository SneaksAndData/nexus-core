package resolvers

import (
	"context"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	nexusclientset "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetWorkgroupByRef(template *v1.NexusAlgorithmTemplate, client nexusclientset.Interface, resourceNamespace string) (*v1.NexusAlgorithmWorkgroupSpec, error) {
	workgroup, err := client.ScienceV1().NexusAlgorithmWorkgroups(resourceNamespace).Get(context.TODO(), template.Spec.SubmissionBehaviour.WorkgroupRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return &workgroup.Spec, err
}
