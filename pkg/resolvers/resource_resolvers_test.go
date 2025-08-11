package resolvers

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2/ktesting"
	"testing"
	"time"
)

type fixture struct {
	kubeClient kubernetes.Interface

	//kubeObjects  []runtime.Object
	k8sFactory   kubeinformers.SharedInformerFactory
	k8sInformers map[string]cache.SharedIndexInformer
}

func (f *fixture) populateJobs(jobs []batchv1.Job) *fixture {
	for _, job := range jobs {
		_ = f.k8sInformers["Job"].GetIndexer().Add(&job)
	}

	return f
}

func (f *fixture) populatePods(pods []corev1.Pod) *fixture {
	for _, pod := range pods {
		_ = f.k8sInformers["Pod"].GetIndexer().Add(&pod)
	}

	return f
}

var noResyncPeriodFunc = func() time.Duration { return 0 }

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.k8sInformers = map[string]cache.SharedIndexInformer{}

	f.kubeClient = fake.NewClientset()

	f.k8sFactory = kubeinformers.NewSharedInformerFactoryWithOptions(f.kubeClient, noResyncPeriodFunc(), kubeinformers.WithNamespace("nexus"))

	f.k8sInformers["Job"] = f.k8sFactory.Batch().V1().Jobs().Informer()
	f.k8sInformers["Pod"] = f.k8sFactory.Core().V1().Pods().Informer()

	return f
}

func Test_IsNexusRunEvent_Job(t *testing.T) {
	_, ctx := ktesting.NewTestContext(t)
	testFixture := newFixture(t).populateJobs([]batchv1.Job{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job",
				Namespace: "nexus",
				Labels: map[string]string{
					models.NexusComponentLabel: models.JobLabelAlgorithmRun,
				},
			},
		},
	},
	)
	testFixture.k8sFactory.Start(ctx.Done())

	evt := &corev1.Event{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Event",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-event",
			Namespace: "nexus",
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Job",
			Namespace: "nexus",
			Name:      "test-job",
		},
		Reason:         "test",
		Message:        "test",
		Source:         corev1.EventSource{},
		FirstTimestamp: metav1.Time{},
		LastTimestamp:  metav1.Time{},
		Count:          1,
		EventTime:      metav1.MicroTime{},
	}

	isNexusRun, err := IsNexusRunEvent(evt, "nexus", testFixture.k8sInformers)

	if err != nil {
		t.Errorf("error checking an Event to be emitted from Nexus: %v", err)
	}

	if !isNexusRun {
		t.Errorf("expected a test event to be emitted from Nexus")
	}
}

func Test_IsNexusRunEvent_Pod(t *testing.T) {
	_, ctx := ktesting.NewTestContext(t)
	testFixture := newFixture(t).populatePods([]corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-1",
				Namespace: "nexus",
				Labels: map[string]string{
					models.NexusComponentLabel: models.JobLabelAlgorithmRun,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod-2",
				Namespace: "nexus",
			},
		},
	},
	)
	testFixture.k8sFactory.Start(ctx.Done())

	evt := &corev1.Event{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Event",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-event",
			Namespace: "nexus",
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "Pod",
			Namespace: "nexus",
			Name:      "test-pod-2",
		},
		Reason:         "test",
		Message:        "test",
		Source:         corev1.EventSource{},
		FirstTimestamp: metav1.Time{},
		LastTimestamp:  metav1.Time{},
		Count:          1,
		EventTime:      metav1.MicroTime{},
	}

	isNexusRun, err := IsNexusRunEvent(evt, "nexus", testFixture.k8sInformers)

	if err != nil {
		t.Errorf("error checking an Event to be emitted from Nexus: %v", err)
	}

	if isNexusRun {
		t.Errorf("expected a test to be NOT emitted from Nexus")
	}
}
