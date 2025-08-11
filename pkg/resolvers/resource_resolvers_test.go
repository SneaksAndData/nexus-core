package resolvers

import (
	"fmt"
	nexusv1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	nexusclient "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"testing"
)

type fixture struct {
	client nexusclient.Interface
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	kubeCfg, err := clientcmd.BuildConfigFromFlags("", "../../test-resources/kubecfg/controller")
	if err != nil {
		t.Fatal(fmt.Errorf("error building kubeConfig: %v", err))
	}

	f.client, err = nexusclient.NewForConfig(kubeCfg)

	if err != nil {
		t.Fatal(fmt.Errorf("error building Nexus client: %v", err))
	}

	return f
}

func Test_GetExistingWorkgroupByRef(t *testing.T) {
	testFixture := newFixture(t)
	wg, err := GetWorkgroupByRef(&nexusv1.NexusAlgorithmTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NexusAlgorithm",
			APIVersion: nexusv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello-world",
			Namespace: "nexus",
		},
		Spec: nexusv1.NexusAlgorithmSpec{
			WorkgroupRef: &nexusv1.NexusAlgorithmWorkgroupRef{
				Name:  "default",
				Group: "science.sneaksanddata.com/v1",
				Kind:  "NexusAlgorithmWorkgroup",
			},
		},
		Status: nexusv1.NexusAlgorithmStatus{},
	}, testFixture.client, "nexus")

	if err != nil {
		t.Errorf("error reading algorithm workgroup: %v", err)
	}

	if wg == nil {
		t.Error("expected NexusAlgorithm workgroup to be returned, but received nil")
	}

	if wg != nil && wg.Cluster != "kind-nexus-shard-0" {
		t.Errorf("expected NexusAlgorithm workgroup to target kind-nexus-shard-0 cluster, but received %s", wg.Cluster)
	}
}
