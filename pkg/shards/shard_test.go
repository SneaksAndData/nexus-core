package shards

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	ktesting "k8s.io/klog/v2/ktesting"
	"reflect"
	nexusv1 "science.sneaksanddata.com/nexus-core/pkg/apis/science/v1"
	"science.sneaksanddata.com/nexus-core/pkg/generated/clientset/versioned/fake"
	informers "science.sneaksanddata.com/nexus-core/pkg/generated/informers/externalversions"
	"testing"
	"time"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t *testing.T

	nexusClient *fake.Clientset
	kubeClient  *k8sfake.Clientset

	mlaLister    []*nexusv1.MachineLearningAlgorithm
	secretLister []*corev1.Secret
	configLister []*corev1.ConfigMap

	// actions to expect on the Kubernetes API
	//kubeActions  []core.Action
	nexusActions []core.Action

	// Objects from here preloaded into NewSimpleFake for controller and a shard.
	kubeObjects  []runtime.Object
	nexusObjects []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t

	f.mlaLister = []*nexusv1.MachineLearningAlgorithm{}
	f.secretLister = []*corev1.Secret{}
	f.configLister = []*corev1.ConfigMap{}

	return f
}

type FakeInformers struct {
	nexusInformers informers.SharedInformerFactory
	k8sInformers   kubeinformers.SharedInformerFactory
}

type ApiFixture struct {
	mlaListResults       []*nexusv1.MachineLearningAlgorithm
	secretListResults    []*corev1.Secret
	configMapListResults []*corev1.ConfigMap

	existingCoreObjects []runtime.Object
	existingMlaObjects  []runtime.Object
}

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {
	ret := []core.Action{}
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "machinelearningalgorithms") ||
				action.Matches("watch", "machinelearningalgorithms") ||
				action.Matches("list", "configmaps") ||
				action.Matches("watch", "configmaps") ||
				action.Matches("list", "secrets") ||
				action.Matches("watch", "secrets")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

// checkAction verifies that expected and actual actions are equal and both have
// same attached resources
func checkAction(expected, actual core.Action, t *testing.T) {
	if !(expected.Matches(actual.GetVerb(), actual.GetResource().Resource) && actual.GetSubresource() == expected.GetSubresource()) {
		t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expected, actual)
		return
	}

	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		t.Errorf("Action has wrong type. Expected: %t. Got: %t", expected, actual)
		return
	}

	switch a := actual.(type) {
	case core.CreateActionImpl:
		e, _ := expected.(core.CreateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.UpdateActionImpl:
		e, _ := expected.(core.UpdateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()
		switch expObject.(type) {
		case *nexusv1.MachineLearningAlgorithm:
			// avoid issues with time drift
			currentTime := metav1.Now()
			expCopy := expObject.DeepCopyObject().(*nexusv1.MachineLearningAlgorithm)
			for ix := range expCopy.Status.Conditions {
				expCopy.Status.Conditions[ix].LastTransitionTime = currentTime
			}

			objCopy := object.DeepCopyObject().(*nexusv1.MachineLearningAlgorithm)
			for ix := range objCopy.Status.Conditions {
				objCopy.Status.Conditions[ix].LastTransitionTime = currentTime
			}

			if !reflect.DeepEqual(expCopy, objCopy) {
				t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
					a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expCopy, objCopy))
			}
		default:
			if !reflect.DeepEqual(expObject, object) {
				t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
					a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
			}
		}
	case core.PatchActionImpl:
		e, _ := expected.(core.PatchActionImpl)
		expPatch := e.GetPatch()
		patch := a.GetPatch()
		if !reflect.DeepEqual(expPatch, patch) {
			t.Errorf("Action %s %s has wrong patch\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expPatch, patch))
		}
	case core.DeleteActionImpl:
		e, _ := expected.(core.DeleteActionImpl)
		if e.GetName() != a.GetName() || e.GetNamespace() != a.GetNamespace() {
			t.Errorf("Action %s targets wrong resource %s/%s, should target %s/%s", a.GetVerb(), a.GetNamespace(), a.GetName(), e.GetNamespace(), e.GetName())
		}
	default:
		t.Errorf("Uncaptured Action %s %s, you should explicitly add a case to capture it",
			actual.GetVerb(), actual.GetResource().Resource)
	}
}

func (f *fixture) checkActions(expected []core.Action, actual []core.Action) {
	actions := filterInformerActions(actual)
	for i, action := range actions {
		if len(expected) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(expected), actions[i:])
			break
		}

		expectedAction := expected[i]
		checkAction(expectedAction, action, f.t)
	}
	if len(expected) > len(actions) {
		f.t.Errorf("%d additional expected actions:%+v", len(expected)-len(actions), expected[len(actions):])
	}
}

// configure adds necessary mock return results for Kubernetes API calls for the respective listers
// and adds existing objects to the respective containers
func (f *fixture) configure(apiFixture *ApiFixture) *fixture {
	f.mlaLister = append(f.mlaLister, apiFixture.mlaListResults...)
	f.secretLister = append(f.secretLister, apiFixture.secretListResults...)
	f.configLister = append(f.configLister, apiFixture.configMapListResults...)
	f.nexusObjects = append(f.nexusObjects, apiFixture.existingMlaObjects...)
	f.kubeObjects = append(f.kubeObjects, apiFixture.existingCoreObjects...)

	return f
}

func (f *fixture) newShard() (*Shard, *FakeInformers) {
	f.nexusClient = fake.NewSimpleClientset(f.nexusObjects...)
	f.kubeClient = k8sfake.NewSimpleClientset(f.kubeObjects...)

	nexusInf := informers.NewSharedInformerFactory(f.nexusClient, noResyncPeriodFunc())
	kubeInf := kubeinformers.NewSharedInformerFactory(f.kubeClient, noResyncPeriodFunc())

	newShard := NewShard(
		"test-controller-cluster",
		"shard0",
		f.kubeClient,
		f.nexusClient,
		nexusInf.Science().V1().MachineLearningAlgorithms(),
		kubeInf.Core().V1().Secrets(),
		kubeInf.Core().V1().ConfigMaps())

	newShard.MlaSynced = alwaysReady
	newShard.SecretsSynced = alwaysReady
	newShard.ConfigMapsSynced = alwaysReady

	for _, d := range f.mlaLister {
		_ = nexusInf.Science().V1().MachineLearningAlgorithms().Informer().GetIndexer().Add(d)
	}

	for _, d := range f.secretLister {
		_ = kubeInf.Core().V1().Secrets().Informer().GetIndexer().Add(d)
	}

	for _, d := range f.configLister {
		_ = kubeInf.Core().V1().ConfigMaps().Informer().GetIndexer().Add(d)
	}

	return newShard, &FakeInformers{
		nexusInformers: nexusInf,
		k8sInformers:   kubeInf,
	}
}

//func getRef(mla *nexusv1.MachineLearningAlgorithm) cache.ObjectName {
//	ref := cache.MetaObjectToName(mla)
//	return ref
//}

func newShardMla() *nexusv1.MachineLearningAlgorithm {
	mla := &nexusv1.MachineLearningAlgorithm{
		TypeMeta: metav1.TypeMeta{APIVersion: nexusv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-algorithms",
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				"science.sneaksanddata.com/controller-app":      "nexus-configuration-controller",
				"science.sneaksanddata.com/configuration-owner": "test-controller-cluster",
			},
		},
		Spec: nexusv1.MachineLearningAlgorithmSpec{
			ImageRegistry:   "test.io",
			ImageRepository: "algorithms/test",
			ImageTag:        "v1.0.0",
			DeadlineSeconds: nexusv1.Int32Ptr(120),
			MaximumRetries:  nexusv1.Int32Ptr(3),
			Env:             make([]corev1.EnvVar, 0),
			EnvFrom: []corev1.EnvFromSource{
				{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "test-secret"},
					},
				},
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "test-cfg1"},
					},
				},
			},
			CpuLimit:             "1000m",
			MemoryLimit:          "2000Mi",
			WorkgroupHost:        "test-cluster.io",
			Workgroup:            "default",
			AdditionalWorkgroups: map[string]string{},
			MonitoringParameters: []string{},
			CustomResources:      map[string]string{},
			SpeculativeAttempts:  nexusv1.Int32Ptr(0),
			TransientExitCodes:   []int32{},
			FatalExitCodes:       []int32{},
			Command:              "python",
			Args:                 []string{"job.py", "--request-id 111-222-333 --arg1 true"},
			MountDatadogSocket:   true,
		},
	}

	return mla
}

// TestCreateMachineLearningAlgorithm tests that resource creation action happens correctly
func TestCreateMachineLearningAlgorithm(t *testing.T) {
	f := newFixture(t)
	mla := newShardMla()
	_, ctx := ktesting.NewTestContext(t)
	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(ctx.Done())
	testInformers.nexusInformers.Start(ctx.Done())

	f = f.configure(
		&ApiFixture{
			mlaListResults:       []*nexusv1.MachineLearningAlgorithm{},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{},
			existingCoreObjects:  []runtime.Object{},
			existingMlaObjects:   []runtime.Object{},
		},
	)

	f.nexusActions = append(f.nexusActions, core.NewCreateAction(schema.GroupVersionResource{Resource: "machinelearningalgorithms"}, mla.Namespace, mla))
	_, _ = shard.CreateMachineLearningAlgorithm(mla.Name, mla.Namespace, mla.Spec, "test")

	f.checkActions(f.nexusActions, f.nexusClient.Actions())
	t.Log("Shard client created a new MLA resource")
}
