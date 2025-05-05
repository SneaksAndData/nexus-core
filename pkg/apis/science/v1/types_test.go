package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/diff"
	"reflect"
	"slices"
	"testing"
)

func newFakeMla(withConfigs bool) *NexusAlgorithmTemplate {
	var envFrom []corev1.EnvFromSource
	var env []corev1.EnvVar

	if withConfigs {
		envFrom = []corev1.EnvFromSource{
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
			{
				ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: "test-cfg2"},
				},
			},
			{
				ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: "test-cfg2"},
				},
			},
		}
		env = []corev1.EnvVar{
			{
				Name: "TEST_VAR1",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "test-secret"},
					},
				},
			},
			{
				Name: "TEST_VAR2",
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "test-cfg3"},
					},
				},
			},
		}
	}

	mla := &NexusAlgorithmTemplate{
		TypeMeta: metav1.TypeMeta{APIVersion: SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-algorithms",
			Namespace: metav1.NamespaceDefault,
			Labels:    map[string]string{"nexus/algorithm-class": "mip"},
			UID:       types.UID("123"),
		},
		Spec: NexusAlgorithmSpec{
			ImageRegistry:        "test.io",
			ImageRepository:      "algorithms/test",
			ImageTag:             "v1.0.0",
			DeadlineSeconds:      Int32Ptr(120),
			MaximumRetries:       Int32Ptr(3),
			Env:                  env,
			EnvFrom:              envFrom,
			CpuLimit:             "1000m",
			MemoryLimit:          "2000Mi",
			WorkgroupHost:        "test-cluster.io",
			Workgroup:            "default",
			AdditionalWorkgroups: map[string]string{},
			MonitoringParameters: []string{},
			CustomResources:      map[string]string{},
			SpeculativeAttempts:  Int32Ptr(0),
			TransientExitCodes:   []int32{},
			FatalExitCodes:       []int32{},
			Command:              "python",
			Args:                 []string{"job.py", "--request-id 111-222-333 --arg1 true"},
			MountDatadogSocket:   BoolPtr(true),
		},
	}

	mla.Status = NexusAlgorithmStatus{
		SyncedSecrets:        []string{"test-secret"},
		SyncedConfigurations: []string{"test-config"},
		SyncedToClusters:     []string{"test-cluster"},
		Conditions: []metav1.Condition{
			*NewResourceReadyCondition(metav1.Now(), metav1.ConditionTrue, "Success"),
		},
	}

	return mla
}

func TestMachineLearningAlgorithm_GetSecretNames(t *testing.T) {
	secretsNames := newFakeMla(true).GetSecretNames()
	expectedSecretNames := []string{
		"test-secret",
	}

	slices.Sort(secretsNames)
	slices.Sort(expectedSecretNames)

	if !reflect.DeepEqual(expectedSecretNames, secretsNames) {
		t.Errorf("Incorrect secrets %s returned for the algorithm", diff.ObjectGoPrintSideBySide(expectedSecretNames, secretsNames))
	}
	t.Log("GetSecretNames returns correct result")
}

func TestMachineLearningAlgorithm_GetConfigMapNames(t *testing.T) {
	configMapNames := newFakeMla(true).GetConfigMapNames()
	expectedConfigMapNames := []string{
		"test-cfg1",
		"test-cfg2",
		"test-cfg3",
	}

	slices.Sort(configMapNames)
	slices.Sort(expectedConfigMapNames)

	if !reflect.DeepEqual(expectedConfigMapNames, configMapNames) {
		t.Errorf("Incorrect configmaps %s returned for the algorithm", diff.ObjectGoPrintSideBySide(expectedConfigMapNames, configMapNames))
	}
	t.Log("GetConfigMapNames returns correct result")
}

func TestMachineLearningAlgorithm_GetSecretNamesNames_Empty(t *testing.T) {
	secretsNames := newFakeMla(false).GetSecretNames()

	if secretsNames != nil {
		t.Errorf("GetSecretNames returns non-empty result when algorithm has no secret references")
	}
	t.Log("GetSecretNames returns correct result when algorithm has no secret references")
}

func TestMachineLearningAlgorithm_GetConfigMapNames_Empty(t *testing.T) {
	configMapNames := newFakeMla(false).GetConfigMapNames()

	if configMapNames != nil {
		t.Errorf("GetConfigMapNames returns non-empty result when algorithm has no configmap references")
	}
	t.Log("GetConfigMapNames returns correct result when algorithm has no config references")
}

func TestMachineLearningAlgorithm_GetConfigMapDiff(t *testing.T) {
	mla1 := newFakeMla(true)
	mla2 := mla1.DeepCopy()
	// remove test-secret and test-cfg3 references
	// however, test-secret is also referenced in EnvFrom, so we should only have test-cfg3 reported as diff
	mla2.Spec.Env = []corev1.EnvVar{}
	diffs := mla1.GetConfigmapDiff(mla2)

	if !reflect.DeepEqual([]string{"test-cfg3"}, diffs) {
		t.Errorf("Incorrect difference %s returned", diff.ObjectGoPrintSideBySide(diffs, diffs))
	}
	t.Log("GetConfigMapDiff evaluates difference in configmap references correctly")
}

func TestMachineLearningAlgorithm_GetSecretDiff(t *testing.T) {
	mla1 := newFakeMla(true)
	mla2 := mla1.DeepCopy()
	// remove test-secret and test-cfg3 references
	// however, test-secret is also referenced in EnvFrom, so we should only have test-cfg3 reported as diff
	mla2.Spec.Env = []corev1.EnvVar{}
	mla2.Spec.EnvFrom = []corev1.EnvFromSource{}
	diffs := mla1.GetSecretDiff(mla2)

	if !reflect.DeepEqual([]string{"test-secret"}, diffs) {
		t.Errorf("Incorrect difference %s returned", diff.ObjectGoPrintSideBySide(diffs, diffs))
	}
	t.Log("GetSecretDiff evaluates difference in secret references correctly")
}
