package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/diff"
	"reflect"
	"testing"
)

func newMla(withConfigs bool) *MachineLearningAlgorithm {
	var envFrom []corev1.EnvFromSource

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
	}

	mla := &MachineLearningAlgorithm{
		TypeMeta: metav1.TypeMeta{APIVersion: SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-algorithms",
			Namespace: metav1.NamespaceDefault,
			Labels:    map[string]string{"nexus/algorithm-class": "mip"},
			UID:       types.UID("123"),
		},
		Spec: MachineLearningAlgorithmSpec{
			ImageRegistry:        "test.io",
			ImageRepository:      "algorithms/test",
			ImageTag:             "v1.0.0",
			DeadlineSeconds:      Int32Ptr(120),
			MaximumRetries:       Int32Ptr(3),
			Env:                  make([]corev1.EnvVar, 0),
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
			MountDatadogSocket:   true,
		},
	}

	mla.Status = MachineLearningAlgorithmStatus{
		SyncedSecrets:        []string{"test-secret"},
		SyncedConfigurations: []string{"test-config"},
		SyncedToClusters:     []string{"test-cluster"},
		Conditions: []metav1.Condition{
			*NewResourceReadyCondition(metav1.Now(), metav1.ConditionTrue, "Success"),
		},
	}

	return mla
}

func TestMachineLearningAlgorithm_GetSecretNamesNames(t *testing.T) {
	secretsNames := newMla(true).GetSecretNames()
	expectedSecretNames := []string{
		"test-secret",
	}

	if !reflect.DeepEqual(expectedSecretNames, secretsNames) {
		t.Errorf("Incorrect secrets %s returned for the algorithm", diff.ObjectGoPrintSideBySide(expectedSecretNames, secretsNames))
	}
	t.Log("GetSecretNames returns correct result")
}

func TestMachineLearningAlgorithm_GetSecretNamesNames_Empty(t *testing.T) {
	secretsNames := newMla(false).GetSecretNames()

	if secretsNames != nil {
		t.Errorf("GetSecretNames returns non-empty result when algorithm has no secret references")
	}
	t.Log("GetSecretNames returns correct result when algorithm has no secret references")
}

func TestMachineLearningAlgorithm_GetConfigMapNames_Empty(t *testing.T) {
	configMapNames := newMla(false).GetConfigMapNames()

	if configMapNames != nil {
		t.Errorf("GetConfigMapNames returns non-empty result when algorithm has no configmap references")
	}
	t.Log("GetConfigMapNames returns correct result when algorithm has no config references")
}
