package shards

import (
	"context"
	nexusv1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned/fake"
	informers "github.com/SneaksAndData/nexus-core/pkg/generated/informers/externalversions"
	"github.com/aws/smithy-go/ptr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/klog/v2/ktesting"
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

	testContext context.Context

	nexusClient *fake.Clientset
	kubeClient  *k8sfake.Clientset

	templateLister  []*nexusv1.NexusAlgorithmTemplate
	workgroupLister []*nexusv1.NexusAlgorithmWorkgroup
	secretLister    []*corev1.Secret
	configLister    []*corev1.ConfigMap
	jobLister       []*batchv1.Job

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
	_, f.testContext = ktesting.NewTestContext(t)

	f.templateLister = []*nexusv1.NexusAlgorithmTemplate{}
	f.workgroupLister = []*nexusv1.NexusAlgorithmWorkgroup{}
	f.secretLister = []*corev1.Secret{}
	f.configLister = []*corev1.ConfigMap{}
	f.jobLister = []*batchv1.Job{}

	return f
}

type FakeInformers struct {
	nexusInformers informers.SharedInformerFactory
	k8sInformers   kubeinformers.SharedInformerFactory
}

type ApiFixture struct {
	templateListResults  []*nexusv1.NexusAlgorithmTemplate
	workgroupListResults []*nexusv1.NexusAlgorithmWorkgroup
	secretListResults    []*corev1.Secret
	configMapListResults []*corev1.ConfigMap
	jobListResults       []*batchv1.Job

	existingKubeObjects  []runtime.Object
	existingNexusObjects []runtime.Object
}

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {
	ret := []core.Action{}
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "nexusalgorithmtemplates") ||
				action.Matches("watch", "nexusalgorithmtemplates") ||
				action.Matches("list", "nexusalgorithmworkgroups") ||
				action.Matches("watch", "nexusalgorithmworkgroups") ||
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
	if !expected.Matches(actual.GetVerb(), actual.GetResource().Resource) || (actual.GetSubresource() != expected.GetSubresource()) {
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
	f.templateLister = append(f.templateLister, apiFixture.templateListResults...)
	f.workgroupLister = append(f.workgroupLister, apiFixture.workgroupListResults...)
	f.secretLister = append(f.secretLister, apiFixture.secretListResults...)
	f.configLister = append(f.configLister, apiFixture.configMapListResults...)
	f.jobLister = append(f.jobLister, apiFixture.jobListResults...)
	f.nexusObjects = append(f.nexusObjects, apiFixture.existingNexusObjects...)
	f.kubeObjects = append(f.kubeObjects, apiFixture.existingKubeObjects...)

	return f
}

func fakeReferenceLabels() map[string]string {
	return map[string]string{
		"science.sneaksanddata.com/controller-app":      "nexus-configuration-controller",
		"science.sneaksanddata.com/configuration-owner": "test-controller-cluster",
	}
}

func (f *fixture) newShard() (*Shard, *FakeInformers) {
	f.nexusClient = fake.NewClientset(f.nexusObjects...)
	f.kubeClient = k8sfake.NewClientset(f.kubeObjects...)

	nexusInf := informers.NewSharedInformerFactory(f.nexusClient, noResyncPeriodFunc())
	kubeInf := kubeinformers.NewSharedInformerFactory(f.kubeClient, noResyncPeriodFunc())

	newShard := NewShard(
		"test-controller-cluster",
		"shard0",
		f.kubeClient,
		f.nexusClient,
		nexusInf.Science().V1().NexusAlgorithmTemplates(),
		nexusInf.Science().V1().NexusAlgorithmWorkgroups(),
		kubeInf.Core().V1().Secrets(),
		kubeInf.Core().V1().ConfigMaps(),
		kubeInf.Batch().V1().Jobs().Informer())

	newShard.TemplateSynced = alwaysReady
	newShard.WorkgroupSynced = alwaysReady
	newShard.SecretsSynced = alwaysReady
	newShard.ConfigMapsSynced = alwaysReady

	for _, d := range f.templateLister {
		_ = nexusInf.Science().V1().NexusAlgorithmTemplates().Informer().GetIndexer().Add(d)
	}

	for _, d := range f.workgroupLister {
		_ = nexusInf.Science().V1().NexusAlgorithmWorkgroups().Informer().GetIndexer().Add(d)
	}

	for _, d := range f.secretLister {
		_ = kubeInf.Core().V1().Secrets().Informer().GetIndexer().Add(d)
	}

	for _, d := range f.configLister {
		_ = kubeInf.Core().V1().ConfigMaps().Informer().GetIndexer().Add(d)
	}

	for _, d := range f.jobLister {
		_ = kubeInf.Batch().V1().Jobs().Informer().GetIndexer().Add(d)
	}

	return newShard, &FakeInformers{
		nexusInformers: nexusInf,
		k8sInformers:   kubeInf,
	}
}

func newTemplateOnShard() *nexusv1.NexusAlgorithmTemplate {
	template := &nexusv1.NexusAlgorithmTemplate{
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
			Container: &nexusv1.NexusAlgorithmContainer{
				Image:      "algorithms/test",
				Registry:   "test.io",
				VersionTag: "v1.0.0",
			},
			ComputeResources: &nexusv1.NexusAlgorithmResources{
				CpuLimit:    "1000m",
				MemoryLimit: "2000Mi",
			},
			WorkgroupRef: &nexusv1.NexusAlgorithmWorkgroupRef{
				Name:  "default",
				Group: nexusv1.SchemeGroupVersion.String(),
				Kind:  "NexusAlgorithmWorkgroup",
			},
			Command: "python",
			Args:    []string{"job.py", "--request-id 111-222-333 --arg1 true"},
			RuntimeEnvironment: &nexusv1.NexusAlgorithmRuntimeEnvironment{
				EnvironmentVariables: make([]corev1.EnvVar, 0),
				MappedEnvironmentVariables: []corev1.EnvFromSource{
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
				DeadlineSeconds: ptr.Int32(120),
				MaximumRetries:  ptr.Int32(3),
			},
			ErrorHandlingBehaviour: &nexusv1.NexusErrorHandlingBehaviour{
				TransientExitCodes: []int32{},
				FatalExitCodes:     []int32{},
			},
			DatadogIntegrationSettings: &nexusv1.NexusDatadogIntegrationSettings{
				MountDatadogSocket: ptr.Bool(true),
			},
		},
	}

	return template
}

func newWorkgroupOnShard() *nexusv1.NexusAlgorithmWorkgroup {
	workgroup := &nexusv1.NexusAlgorithmWorkgroup{
		TypeMeta: metav1.TypeMeta{APIVersion: nexusv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: metav1.NamespaceDefault,
			Labels: map[string]string{
				"science.sneaksanddata.com/controller-app":      "nexus-configuration-controller",
				"science.sneaksanddata.com/configuration-owner": "test-controller-cluster",
			},
		},
		Spec: nexusv1.NexusAlgorithmWorkgroupSpec{
			Description:  "test workgroup",
			Capabilities: map[string]bool{},
			Cluster:      "shard0",
			Tolerations:  []corev1.Toleration{},
			Affinity:     &corev1.Affinity{},
		},
	}

	return workgroup
}

// TestShard_CreateNexusAlgorithmTemplate tests that resource creation action happens correctly
func TestShard_CreateNexusAlgorithmTemplate(t *testing.T) {
	f := newFixture(t)
	template := newTemplateOnShard()

	f = f.configure(
		&ApiFixture{
			templateListResults:  []*nexusv1.NexusAlgorithmTemplate{},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{},
			existingKubeObjects:  []runtime.Object{},
			existingNexusObjects: []runtime.Object{},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(f.testContext.Done())
	testInformers.nexusInformers.Start(f.testContext.Done())

	f.nexusActions = append(f.nexusActions, core.NewCreateAction(schema.GroupVersionResource{Resource: "nexusalgorithmtemplates"}, template.Namespace, template))
	_, _ = shard.CreateTemplate(template.Name, template.Namespace, template.Spec, "test")

	f.checkActions(f.nexusActions, f.nexusClient.Actions())
	t.Log("Shard client created a new Template resource")
}

// TestUpdateNexusAlgorithmTemplate tests that resource update action happens correctly
func TestShard_UpdateNexusAlgorithmTemplate(t *testing.T) {
	f := newFixture(t)
	template := newTemplateOnShard()
	updatedTemplate := template.DeepCopy()
	updatedTemplate.Spec.ComputeResources.CpuLimit = "2000m"

	f = f.configure(
		&ApiFixture{
			templateListResults:  []*nexusv1.NexusAlgorithmTemplate{template},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{},
			existingKubeObjects:  []runtime.Object{},
			existingNexusObjects: []runtime.Object{template},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(f.testContext.Done())
	testInformers.nexusInformers.Start(f.testContext.Done())

	f.nexusActions = append(f.nexusActions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "nexusalgorithmtemplates"}, template.Namespace, updatedTemplate))
	_, _ = shard.UpdateTemplate(template, updatedTemplate.Spec, "test")

	f.checkActions(f.nexusActions, f.nexusClient.Actions())
	t.Log("Shard client update an existing Template resource")
}

// TestDeleteNexusAlgorithmTemplate tests that resource delete action happens correctly
func TestShard_DeleteNexusAlgorithmTemplate(t *testing.T) {
	f := newFixture(t)
	template := newTemplateOnShard()

	f = f.configure(
		&ApiFixture{
			templateListResults:  []*nexusv1.NexusAlgorithmTemplate{template},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{},
			existingKubeObjects:  []runtime.Object{},
			existingNexusObjects: []runtime.Object{template},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(f.testContext.Done())
	testInformers.nexusInformers.Start(f.testContext.Done())

	f.nexusActions = append(f.nexusActions, core.NewDeleteAction(schema.GroupVersionResource{Resource: "nexusalgorithmtemplates"}, template.Namespace, template.Name))
	_ = shard.DeleteTemplate(template)

	f.checkActions(f.nexusActions, f.nexusClient.Actions())
	t.Log("Shard client deleted an existing Template resource")
}

// TestShard_CreateSecret tests that resource delete action happens correctly
func TestShard_CreateSecret(t *testing.T) {
	f := newFixture(t)
	template := newTemplateOnShard()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: template.Namespace,
		},
		Data: map[string][]byte{
			"key": []byte("value"),
		},
		StringData: map[string]string{},
	}

	expectedSecret := secret.DeepCopy()
	expectedSecret.Labels = fakeReferenceLabels()
	expectedSecret.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: nexusv1.SchemeGroupVersion.String(),
			Kind:       "NexusAlgorithmTemplate",
			Name:       template.Name,
			UID:        template.UID,
		},
	}

	f = f.configure(
		&ApiFixture{
			templateListResults:  []*nexusv1.NexusAlgorithmTemplate{template},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{},
			existingKubeObjects:  []runtime.Object{},
			existingNexusObjects: []runtime.Object{template},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(f.testContext.Done())
	testInformers.nexusInformers.Start(f.testContext.Done())

	f.kubeActions = append(f.kubeActions, core.NewCreateAction(schema.GroupVersionResource{Resource: "secrets", Version: "v1"}, template.Namespace, expectedSecret))
	_, _ = shard.CreateSecret(template, secret, "test")

	f.checkActions(f.kubeActions, f.kubeClient.Actions())
	t.Log("Shard client created a Secret from controller cluster secret")
}

// TestShard_CreateConfigMap tests that resource delete action happens correctly
func TestShard_CreateConfigMap(t *testing.T) {
	f := newFixture(t)
	template := newTemplateOnShard()
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: template.Namespace,
		},
		Data: map[string]string{
			"key": "value",
		},
	}

	expectedConfigMap := configMap.DeepCopy()
	expectedConfigMap.Labels = fakeReferenceLabels()
	expectedConfigMap.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: nexusv1.SchemeGroupVersion.String(),
			Kind:       "NexusAlgorithmTemplate",
			Name:       template.Name,
			UID:        template.UID,
		},
	}

	f = f.configure(
		&ApiFixture{
			templateListResults:  []*nexusv1.NexusAlgorithmTemplate{template},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{},
			existingKubeObjects:  []runtime.Object{},
			existingNexusObjects: []runtime.Object{template},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(f.testContext.Done())
	testInformers.nexusInformers.Start(f.testContext.Done())

	f.kubeActions = append(f.kubeActions, core.NewCreateAction(schema.GroupVersionResource{Resource: "configmaps", Version: "v1"}, template.Namespace, expectedConfigMap))
	_, _ = shard.CreateConfigMap(template, expectedConfigMap, "test")

	f.checkActions(f.kubeActions, f.kubeClient.Actions())
	t.Log("Shard client created a configmap from controller cluster configmap")
}

// TestShard_UpdateConfigMap tests that config update action happens correctly
func TestShard_UpdateConfigMap(t *testing.T) {
	f := newFixture(t)
	template := newTemplateOnShard()

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: template.Namespace,
			Labels:    fakeReferenceLabels(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: nexusv1.SchemeGroupVersion.String(),
					Kind:       "NexusAlgorithmTemplate",
					Name:       template.Name,
					UID:        template.UID,
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
			templateListResults:  []*nexusv1.NexusAlgorithmTemplate{template},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{configMap},
			existingKubeObjects:  []runtime.Object{},
			existingNexusObjects: []runtime.Object{template},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(f.testContext.Done())
	testInformers.nexusInformers.Start(f.testContext.Done())

	f.kubeActions = append(f.kubeActions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "configmaps", Version: "v1"}, template.Namespace, expectedConfigMap))
	_, _ = shard.UpdateConfigMap(configMap, expectedConfigMap.Data, nil, "test")

	f.checkActions(f.kubeActions, f.kubeClient.Actions())
	t.Log("Shard client updated an existing configmap")
}

// TestShard_UpdateSecret tests that secret update action happens correctly
func TestShard_UpdateSecret(t *testing.T) {
	f := newFixture(t)
	template := newTemplateOnShard()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: template.Namespace,
			Labels:    fakeReferenceLabels(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: nexusv1.SchemeGroupVersion.String(),
					Kind:       "NexusAlgorithmTemplate",
					Name:       template.Name,
					UID:        template.UID,
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
			templateListResults:  []*nexusv1.NexusAlgorithmTemplate{template},
			secretListResults:    []*corev1.Secret{secret},
			configMapListResults: []*corev1.ConfigMap{},
			existingKubeObjects:  []runtime.Object{secret},
			existingNexusObjects: []runtime.Object{template},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(f.testContext.Done())
	testInformers.nexusInformers.Start(f.testContext.Done())

	f.kubeActions = append(f.kubeActions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "secrets", Version: "v1"}, template.Namespace, expectedSecret))
	_, _ = shard.UpdateSecret(secret, expectedSecret.Data, nil, "test")

	f.checkActions(f.kubeActions, f.kubeClient.Actions())
	t.Log("Shard client updated an existing secret")
}

// TestShard_DereferenceSecret tests that secret update action happens correctly
func TestShard_DereferenceSecret(t *testing.T) {
	f := newFixture(t)
	template := newTemplateOnShard()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: template.Namespace,
			Labels:    fakeReferenceLabels(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: nexusv1.SchemeGroupVersion.String(),
					Kind:       "NexusAlgorithmTemplate",
					Name:       template.Name,
					UID:        template.UID,
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
			templateListResults:  []*nexusv1.NexusAlgorithmTemplate{template},
			secretListResults:    []*corev1.Secret{secret},
			configMapListResults: []*corev1.ConfigMap{},
			existingKubeObjects:  []runtime.Object{secret},
			existingNexusObjects: []runtime.Object{template},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(f.testContext.Done())
	testInformers.nexusInformers.Start(f.testContext.Done())

	f.kubeActions = append(f.kubeActions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "secrets", Version: "v1"}, template.Namespace, dereferencedSecret))
	f.kubeActions = append(f.kubeActions, core.NewDeleteAction(schema.GroupVersionResource{Resource: "secrets", Version: "v1"}, template.Namespace, "test-secret"))
	_ = shard.DereferenceSecret(secret, template, "test")

	f.checkActions(f.kubeActions, f.kubeClient.Actions())
	t.Log("Shard client dereferenced an existing secret")
}

// TestShard_DereferenceConfigMap tests that secret update action happens correctly
func TestShard_DereferenceConfigMap(t *testing.T) {
	f := newFixture(t)
	template := newTemplateOnShard()

	_, ctx := ktesting.NewTestContext(t)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cfg1",
			Namespace: template.Namespace,
			Labels:    fakeReferenceLabels(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: nexusv1.SchemeGroupVersion.String(),
					Kind:       "NexusAlgorithmTemplate",
					Name:       template.Name,
					UID:        template.UID,
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
			templateListResults:  []*nexusv1.NexusAlgorithmTemplate{template},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{configMap},
			existingKubeObjects:  []runtime.Object{configMap},
			existingNexusObjects: []runtime.Object{template},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(ctx.Done())
	testInformers.nexusInformers.Start(ctx.Done())

	f.kubeActions = append(f.kubeActions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "configmaps", Version: "v1"}, template.Namespace, dereferencedConfigMap))
	f.kubeActions = append(f.kubeActions, core.NewDeleteAction(schema.GroupVersionResource{Resource: "configmaps", Version: "v1"}, template.Namespace, "test-cfg1"))
	_ = shard.DereferenceConfigMap(configMap, template, "test")

	f.checkActions(f.kubeActions, f.kubeClient.Actions())
	t.Log("Shard client dereferenced an existing configMap")
}

// TestShard_CreateNexusAlgorithmWorkgroup tests that workgroup creation action happens correctly
func TestShard_CreateNexusAlgorithmWorkgroup(t *testing.T) {
	f := newFixture(t)
	workgroup := newWorkgroupOnShard()
	_, ctx := ktesting.NewTestContext(t)

	f = f.configure(
		&ApiFixture{
			templateListResults:  []*nexusv1.NexusAlgorithmTemplate{},
			workgroupListResults: []*nexusv1.NexusAlgorithmWorkgroup{},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{},
			existingKubeObjects:  []runtime.Object{},
			existingNexusObjects: []runtime.Object{},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(ctx.Done())
	testInformers.nexusInformers.Start(ctx.Done())

	f.nexusActions = append(f.nexusActions, core.NewCreateAction(schema.GroupVersionResource{Resource: "nexusalgorithmworkgroups"}, workgroup.Namespace, workgroup))
	_, _ = shard.CreateWorkgroup(workgroup.Name, workgroup.Namespace, workgroup.Spec, "test")

	f.checkActions(f.nexusActions, f.nexusClient.Actions())
	t.Log("Shard client created a new Workgroup resource")
}

// TestShard_UpdateNexusAlgorithmWorkgroup tests that workgroup update action works correctly
func TestShard_UpdateNexusAlgorithmWorkgroup(t *testing.T) {
	f := newFixture(t)
	workgroup := newWorkgroupOnShard()
	updatedWorkgroup := workgroup.DeepCopy()
	updatedWorkgroup.Spec.Tolerations = append(updatedWorkgroup.Spec.Tolerations, corev1.Toleration{
		Key:      "key",
		Operator: corev1.TolerationOpEqual,
		Value:    "value",
		Effect:   corev1.TaintEffectNoSchedule,
	})

	_, ctx := ktesting.NewTestContext(t)

	f = f.configure(
		&ApiFixture{
			templateListResults:  []*nexusv1.NexusAlgorithmTemplate{},
			workgroupListResults: []*nexusv1.NexusAlgorithmWorkgroup{workgroup},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{},
			existingKubeObjects:  []runtime.Object{},
			existingNexusObjects: []runtime.Object{workgroup},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(ctx.Done())
	testInformers.nexusInformers.Start(ctx.Done())

	f.nexusActions = append(f.nexusActions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "nexusalgorithmworkgroups"}, workgroup.Namespace, updatedWorkgroup))
	_, _ = shard.UpdateWorkgroup(workgroup, updatedWorkgroup.Spec, "test")

	f.checkActions(f.nexusActions, f.nexusClient.Actions())
	t.Log("Shard client updated an existing Workgroup resource")
}

// TestShard_DeleteNexusAlgorithmWorkgroup tests that workgroup delete action happens correctly
func TestShard_DeleteNexusAlgorithmWorkgroup(t *testing.T) {
	f := newFixture(t)
	workgroup := newWorkgroupOnShard()

	_, ctx := ktesting.NewTestContext(t)

	f = f.configure(
		&ApiFixture{
			templateListResults:  []*nexusv1.NexusAlgorithmTemplate{},
			workgroupListResults: []*nexusv1.NexusAlgorithmWorkgroup{workgroup},
			secretListResults:    []*corev1.Secret{},
			configMapListResults: []*corev1.ConfigMap{},
			existingKubeObjects:  []runtime.Object{},
			existingNexusObjects: []runtime.Object{workgroup},
		},
	)

	shard, testInformers := f.newShard()
	testInformers.k8sInformers.Start(ctx.Done())
	testInformers.nexusInformers.Start(ctx.Done())

	f.nexusActions = append(f.nexusActions, core.NewDeleteAction(schema.GroupVersionResource{Resource: "nexusalgorithmworkgroups"}, workgroup.Namespace, workgroup.Name))
	_ = shard.DeleteWorkgroup(workgroup)

	f.checkActions(f.nexusActions, f.nexusClient.Actions())
	t.Log("Shard client deleted an existing Workgroup resource")
}
