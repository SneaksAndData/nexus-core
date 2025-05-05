package shards

import (
	nexusv1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned/fake"
	informers "github.com/SneaksAndData/nexus-core/pkg/generated/informers/externalversions"
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

	mlaLister    []*nexusv1.NexusAlgorithmTemplate
	secretLister []*corev1.Secret
	configLister []*corev1.ConfigMap

	// actions to expect on the Kubernetes API
	kubeActions  []core.Action
	nexusActions []core.Action

	// Objects from here preloaded into NewSimpleFake for controller and a shard.
	kubeObjects  []runtime.Object
	nexusObjects []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t

	f.mlaLister = []*nexusv1.NexusAlgorithmTemplate{}
	f.secretLister = []*corev1.Secret{}
	f.configLister = []*corev1.ConfigMap{}

	return f
}

type FakeInformers struct {
	nexusInformers informers.SharedInformerFactory
	k8sInformers   kubeinformers.SharedInformerFactory
}

type ApiFixture struct {
	mlaListResults       []*nexusv1.NexusAlgorithmTemplate
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
		case *nexusv1.NexusAlgorithmTemplate:
			// avoid issues with time drift
			currentTime := metav1.Now()
			expCopy := expObject.DeepCopyObject().(*nexusv1.NexusAlgorithmTemplate)
			for ix := range expCopy.Status.Conditions {
				expCopy.Status.Conditions[ix].LastTransitionTime = currentTime
			}

			objCopy := object.DeepCopyObject().(*nexusv1.NexusAlgorithmTemplate)
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

func fakeReferenceLabels() map[string]string {
	return map[string]string{
		"science.sneaksanddata.com/controller-app":      "nexus-configuration-controller",
		"science.sneaksanddata.com/configuration-owner": "test-controller-cluster",
	}
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

func newShardMla() *nexusv1.NexusAlgorithmTemplate {
	mla := &nexusv1.NexusAlgorithmTemplate{
		TypeMeta: metav1.TypeMeta{APIVersion: nexusv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-algorithms",
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				"science.sneaksanddata.com/controller-app":      "nexus-configuration-controller",
				"science.sneaksanddata.com/configuration-owner": "test-controller-cluster",
			},
		},
		Spec: nexusv1.NexusAlgorithmSpec{
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
			MountDatadogSocket:   nexusv1.BoolPtr(true),
		},
	}

	return mla
}

// TestCreateMachineLearningAlgorithm tests that resource creation action happens correctly
func TestShard_CreateMachineLearningAlgorithm(t *testing.T) {
	f := newFixture(t)
	mla := newShardMla()
	_, ctx := ktesting.NewTestContext(t)

	f = f.configure(
		&ApiFixture{
			mlaListResults:       []*nexusv1.NexusAlgorithmTemplate{},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{},
			existingCoreObjects:  []runtime.Object{},
			existingMlaObjects:   []runtime.Object{},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(ctx.Done())
	testInformers.nexusInformers.Start(ctx.Done())

	f.nexusActions = append(f.nexusActions, core.NewCreateAction(schema.GroupVersionResource{Resource: "machinelearningalgorithms"}, mla.Namespace, mla))
	_, _ = shard.CreateMachineLearningAlgorithm(mla.Name, mla.Namespace, mla.Spec, "test")

	f.checkActions(f.nexusActions, f.nexusClient.Actions())
	t.Log("Shard client created a new MLA resource")
}

// TestUpdateMachineLearningAlgorithm tests that resource update action happens correctly
func TestShard_UpdateMachineLearningAlgorithm(t *testing.T) {
	f := newFixture(t)
	mla := newShardMla()
	updatedMla := mla.DeepCopy()
	updatedMla.Spec.CpuLimit = "2000m"

	_, ctx := ktesting.NewTestContext(t)

	f = f.configure(
		&ApiFixture{
			mlaListResults:       []*nexusv1.NexusAlgorithmTemplate{mla},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{},
			existingCoreObjects:  []runtime.Object{},
			existingMlaObjects:   []runtime.Object{mla},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(ctx.Done())
	testInformers.nexusInformers.Start(ctx.Done())

	f.nexusActions = append(f.nexusActions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "machinelearningalgorithms"}, mla.Namespace, updatedMla))
	_, _ = shard.UpdateMachineLearningAlgorithm(mla, updatedMla.Spec, "test")

	f.checkActions(f.nexusActions, f.nexusClient.Actions())
	t.Log("Shard client update an existing MLA resource")
}

// TestDeleteMachineLearningAlgorithm tests that resource delete action happens correctly
func TestShard_DeleteMachineLearningAlgorithm(t *testing.T) {
	f := newFixture(t)
	mla := newShardMla()

	_, ctx := ktesting.NewTestContext(t)

	f = f.configure(
		&ApiFixture{
			mlaListResults:       []*nexusv1.NexusAlgorithmTemplate{mla},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{},
			existingCoreObjects:  []runtime.Object{},
			existingMlaObjects:   []runtime.Object{mla},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(ctx.Done())
	testInformers.nexusInformers.Start(ctx.Done())

	f.nexusActions = append(f.nexusActions, core.NewDeleteAction(schema.GroupVersionResource{Resource: "machinelearningalgorithms"}, mla.Namespace, mla.Name))
	_ = shard.DeleteMachineLearningAlgorithm(mla)

	f.checkActions(f.nexusActions, f.nexusClient.Actions())
	t.Log("Shard client deleted an existing MLA resource")
}

// TestShard_CreateSecret tests that resource delete action happens correctly
func TestShard_CreateSecret(t *testing.T) {
	f := newFixture(t)
	mla := newShardMla()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: mla.Namespace,
		},
		Data: map[string][]byte{
			"key": []byte("value"),
		},
		StringData: map[string]string{},
	}

	_, ctx := ktesting.NewTestContext(t)

	expectedSecret := secret.DeepCopy()
	expectedSecret.ObjectMeta.Labels = fakeReferenceLabels()
	expectedSecret.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: nexusv1.SchemeGroupVersion.String(),
			Kind:       "NexusAlgorithmTemplate",
			Name:       mla.Name,
			UID:        mla.UID,
		},
	}

	f = f.configure(
		&ApiFixture{
			mlaListResults:       []*nexusv1.NexusAlgorithmTemplate{mla},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{},
			existingCoreObjects:  []runtime.Object{},
			existingMlaObjects:   []runtime.Object{mla},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(ctx.Done())
	testInformers.nexusInformers.Start(ctx.Done())

	f.kubeActions = append(f.kubeActions, core.NewCreateAction(schema.GroupVersionResource{Resource: "secrets", Version: "v1"}, mla.Namespace, expectedSecret))
	_, _ = shard.CreateSecret(mla, secret, "test")

	f.checkActions(f.kubeActions, f.kubeClient.Actions())
	t.Log("Shard client created a Secret from controller cluster secret")
}

// TestShard_CreateConfigMap tests that resource delete action happens correctly
func TestShard_CreateConfigMap(t *testing.T) {
	f := newFixture(t)
	mla := newShardMla()
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: mla.Namespace,
		},
		Data: map[string]string{
			"key": "value",
		},
	}

	_, ctx := ktesting.NewTestContext(t)

	expectedConfigMap := configMap.DeepCopy()
	expectedConfigMap.ObjectMeta.Labels = fakeReferenceLabels()
	expectedConfigMap.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: nexusv1.SchemeGroupVersion.String(),
			Kind:       "NexusAlgorithmTemplate",
			Name:       mla.Name,
			UID:        mla.UID,
		},
	}

	f = f.configure(
		&ApiFixture{
			mlaListResults:       []*nexusv1.NexusAlgorithmTemplate{mla},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{},
			existingCoreObjects:  []runtime.Object{},
			existingMlaObjects:   []runtime.Object{mla},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(ctx.Done())
	testInformers.nexusInformers.Start(ctx.Done())

	f.kubeActions = append(f.kubeActions, core.NewCreateAction(schema.GroupVersionResource{Resource: "configmaps", Version: "v1"}, mla.Namespace, expectedConfigMap))
	_, _ = shard.CreateConfigMap(mla, expectedConfigMap, "test")

	f.checkActions(f.kubeActions, f.kubeClient.Actions())
	t.Log("Shard client created a configmap from controller cluster configmap")
}

// TestShard_UpdateConfigMap tests that config update action happens correctly
func TestShard_UpdateConfigMap(t *testing.T) {
	f := newFixture(t)
	mla := newShardMla()

	_, ctx := ktesting.NewTestContext(t)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: mla.Namespace,
			Labels:    fakeReferenceLabels(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: nexusv1.SchemeGroupVersion.String(),
					Kind:       "NexusAlgorithmTemplate",
					Name:       mla.Name,
					UID:        mla.UID,
				},
			},
		},
		Data: map[string]string{
			"key": "value",
		},
	}

	expectedConfigMap := configMap.DeepCopy()
	expectedConfigMap.Data = map[string]string{
		"key": "new-value",
	}

	f = f.configure(
		&ApiFixture{
			mlaListResults:       []*nexusv1.NexusAlgorithmTemplate{mla},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{configMap},
			existingCoreObjects:  []runtime.Object{},
			existingMlaObjects:   []runtime.Object{mla},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(ctx.Done())
	testInformers.nexusInformers.Start(ctx.Done())

	f.kubeActions = append(f.kubeActions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "configmaps", Version: "v1"}, mla.Namespace, expectedConfigMap))
	_, _ = shard.UpdateConfigMap(configMap, expectedConfigMap.Data, nil, "test")

	f.checkActions(f.kubeActions, f.kubeClient.Actions())
	t.Log("Shard client updated an existing configmap")
}

// TestShard_UpdateSecret tests that secret update action happens correctly
func TestShard_UpdateSecret(t *testing.T) {
	f := newFixture(t)
	mla := newShardMla()

	_, ctx := ktesting.NewTestContext(t)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: mla.Namespace,
			Labels:    fakeReferenceLabels(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: nexusv1.SchemeGroupVersion.String(),
					Kind:       "NexusAlgorithmTemplate",
					Name:       mla.Name,
					UID:        mla.UID,
				},
			},
		},
		Data: map[string][]byte{
			"key": []byte("value"),
		},
		StringData: map[string]string{},
	}

	expectedSecret := secret.DeepCopy()
	expectedSecret.Data = map[string][]byte{
		"key": []byte("new-value"),
	}

	f = f.configure(
		&ApiFixture{
			mlaListResults:       []*nexusv1.NexusAlgorithmTemplate{mla},
			secretListResults:    []*corev1.Secret{secret},
			configMapListResults: []*corev1.ConfigMap{},
			existingCoreObjects:  []runtime.Object{secret},
			existingMlaObjects:   []runtime.Object{mla},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(ctx.Done())
	testInformers.nexusInformers.Start(ctx.Done())

	f.kubeActions = append(f.kubeActions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "secrets", Version: "v1"}, mla.Namespace, expectedSecret))
	_, _ = shard.UpdateSecret(secret, expectedSecret.Data, nil, "test")

	f.checkActions(f.kubeActions, f.kubeClient.Actions())
	t.Log("Shard client updated an existing secret")
}

// TestShard_DereferenceSecret tests that secret update action happens correctly
func TestShard_DereferenceSecret(t *testing.T) {
	f := newFixture(t)
	mla := newShardMla()

	_, ctx := ktesting.NewTestContext(t)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: mla.Namespace,
			Labels:    fakeReferenceLabels(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: nexusv1.SchemeGroupVersion.String(),
					Kind:       "NexusAlgorithmTemplate",
					Name:       mla.Name,
					UID:        mla.UID,
				},
			},
		},
		Data: map[string][]byte{
			"key": []byte("value"),
		},
		StringData: map[string]string{},
	}

	dereferencedSecret := secret.DeepCopy()
	dereferencedSecret.OwnerReferences = []metav1.OwnerReference{}

	f = f.configure(
		&ApiFixture{
			mlaListResults:       []*nexusv1.NexusAlgorithmTemplate{mla},
			secretListResults:    []*corev1.Secret{secret},
			configMapListResults: []*corev1.ConfigMap{},
			existingCoreObjects:  []runtime.Object{secret},
			existingMlaObjects:   []runtime.Object{mla},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(ctx.Done())
	testInformers.nexusInformers.Start(ctx.Done())

	f.kubeActions = append(f.kubeActions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "secrets", Version: "v1"}, mla.Namespace, dereferencedSecret))
	f.kubeActions = append(f.kubeActions, core.NewDeleteAction(schema.GroupVersionResource{Resource: "secrets", Version: "v1"}, mla.Namespace, "test-secret"))
	_ = shard.DereferenceSecret(secret, mla, "test")

	f.checkActions(f.kubeActions, f.kubeClient.Actions())
	t.Log("Shard client dereferenced an existing secret")
}

// TestShard_DereferenceConfigMap tests that secret update action happens correctly
func TestShard_DereferenceConfigMap(t *testing.T) {
	f := newFixture(t)
	mla := newShardMla()

	_, ctx := ktesting.NewTestContext(t)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cfg1",
			Namespace: mla.Namespace,
			Labels:    fakeReferenceLabels(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: nexusv1.SchemeGroupVersion.String(),
					Kind:       "NexusAlgorithmTemplate",
					Name:       mla.Name,
					UID:        mla.UID,
				},
			},
		},
		Data: map[string]string{
			"key": "value",
		},
	}

	dereferencedConfigMap := configMap.DeepCopy()
	dereferencedConfigMap.OwnerReferences = []metav1.OwnerReference{}

	f = f.configure(
		&ApiFixture{
			mlaListResults:       []*nexusv1.NexusAlgorithmTemplate{mla},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{configMap},
			existingCoreObjects:  []runtime.Object{configMap},
			existingMlaObjects:   []runtime.Object{mla},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(ctx.Done())
	testInformers.nexusInformers.Start(ctx.Done())

	f.kubeActions = append(f.kubeActions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "configmaps", Version: "v1"}, mla.Namespace, dereferencedConfigMap))
	f.kubeActions = append(f.kubeActions, core.NewDeleteAction(schema.GroupVersionResource{Resource: "configmaps", Version: "v1"}, mla.Namespace, "test-cfg1"))
	_ = shard.DereferenceConfigMap(configMap, mla, "test")

	f.checkActions(f.kubeActions, f.kubeClient.Actions())
	t.Log("Shard client dereferenced an existing configMap")
}
