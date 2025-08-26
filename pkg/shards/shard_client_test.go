package shards

import (
	"context"
	nexuscore "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	"github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned/fake"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/ktesting"
	"reflect"
	"testing"
)

type clientFixture struct {
	t *testing.T

	testContext context.Context
	kubeClient  kubernetes.Interface
	nexusClient nexuscore.Interface
}

func newClientFixture(t *testing.T, existingObjects []runtime.Object) *clientFixture {
	f := &clientFixture{}
	f.t = t
	_, f.testContext = ktesting.NewTestContext(t)

	f.kubeClient = k8sfake.NewClientset(existingObjects...)
	f.nexusClient = fake.NewClientset()
	return f
}

func (f *clientFixture) newShardClient(name string, namespace string) *ShardClient {
	return NewShardClient(f.kubeClient, f.nexusClient, name, namespace, klog.FromContext(f.testContext))
}

func TestShardClient_SendJob(t *testing.T) {
	f := newClientFixture(t, []runtime.Object{})
	client := f.newShardClient("test", metav1.NamespaceDefault)

	job, err := client.SendJob(metav1.NamespaceDefault, &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: batchv1.SchemeGroupVersion.String(),
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: batchv1.JobSpec{},
	})

	if err != nil {
		t.Fatal(err)
	}

	if job.Name != "test" {
		t.Errorf("job name is %s, want %s", job.Name, "test")
	}
}

func TestShardClient_DeleteJob(t *testing.T) {
	f := newClientFixture(t, []runtime.Object{})
	client := f.newShardClient("test", metav1.NamespaceDefault)
	err := client.DeleteJob(metav1.NamespaceDefault, "test", metav1.DeletePropagationForeground)

	if err == nil {
		t.Errorf("method DeleteJob should have failed (job does not exist) but didn't")
		t.FailNow()
	}

	if !errors.IsNotFound(err) {
		t.Errorf("unexpected error (should be NotFound): %s", err)
	}
}

func TestShardClient_FindJob(t *testing.T) {
	jobs := []*batchv1.Job{
		{
			TypeMeta: metav1.TypeMeta{
				APIVersion: batchv1.SchemeGroupVersion.String(),
				Kind:       "Job",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: metav1.NamespaceDefault,
			},
			Spec: batchv1.JobSpec{},
		},
		{
			TypeMeta: metav1.TypeMeta{
				APIVersion: batchv1.SchemeGroupVersion.String(),
				Kind:       "Job",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-2",
				Namespace: metav1.NamespaceDefault,
			},
			Spec: batchv1.JobSpec{},
		},
	}

	runtimeObjs := []runtime.Object{}
	for _, job := range jobs {
		runtimeObjs = append(runtimeObjs, job)
	}

	f := newClientFixture(t, runtimeObjs)
	client := f.newShardClient("test", metav1.NamespaceDefault)

	// no cache hit for job 'test'
	jobTest, err := client.FindJob("test", metav1.NamespaceDefault)

	if err != nil {
		t.Fatal(err)
	}

	// expecting cache hit for job 'test'
	jobTest2, err := client.FindJob("test", metav1.NamespaceDefault)

	if !reflect.DeepEqual(jobTest, jobTest2) {
		t.Errorf("different jobs returned from API and cache: %s", diff.ObjectGoPrintSideBySide(jobTest, jobTest2))
		t.FailNow()
	}
}
