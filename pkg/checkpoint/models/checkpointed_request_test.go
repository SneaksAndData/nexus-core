package models

import (
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/aws/smithy-go/ptr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/diff"
	"reflect"
	"testing"
	"time"
)

func getFakeRequest() *CheckpointedRequest {
	return &CheckpointedRequest{
		Algorithm:               "test",
		Id:                      "test-id",
		LifecycleStage:          LifecycleStageRunning,
		PayloadUri:              "https://somewhere",
		ResultUri:               "",
		AlgorithmFailureCause:   "",
		AlgorithmFailureDetails: "",
		ReceivedByHost:          "test-host",
		ReceivedAt:              time.Now(),
		SentAt:                  time.Now(),
		AppliedConfiguration: &v1.NexusAlgorithmSpec{
			Container: &v1.NexusAlgorithmContainer{
				Image:              "test.io",
				Registry:           "algorithms/test",
				VersionTag:         "v1.0.0",
				ServiceAccountName: "test-sa",
			},
			ComputeResources: &v1.NexusAlgorithmResources{
				CpuLimit:    "1000m",
				MemoryLimit: "2000Mi",
			},
			WorkgroupRef: &v1.NexusAlgorithmWorkgroupRef{
				Name:  "test-workgroup",
				Group: "nexus-workgroup.io",
				Kind:  "KarpenterWorkgroupV1",
			},
			Command: "python",
			Args:    []string{"job.py", "--sas-uri=%s", "--request-id=%s", "--arg1=true"},
			RuntimeEnvironment: &v1.NexusAlgorithmRuntimeEnvironment{
				EnvironmentVariables:       nil,
				MappedEnvironmentVariables: nil,
				DeadlineSeconds:            ptr.Int32(120),
				MaximumRetries:             ptr.Int32(3),
			},
			ErrorHandlingBehaviour: nil,
			DatadogIntegrationSettings: &v1.NexusDatadogIntegrationSettings{
				MountDatadogSocket: ptr.Bool(true),
			},
		},
		ConfigurationOverrides: nil,
		ContentHash:            "12313",
		LastModified:           time.Now(),
		Tag:                    "abc",
		ApiVersion:             "1.2",
		JobUid:                 "1231",
		Parent:                 &AlgorithmRequestRef{RequestId: "test-parent", AlgorithmName: "test-parent-algorithm"},
		PayloadValidFor:        "86400s",
	}
}

func TestCheckpointedRequest_DeepCopy(t *testing.T) {
	fakeRequest := getFakeRequest()
	deepCopy := fakeRequest.DeepCopy()

	if !reflect.DeepEqual(fakeRequest, deepCopy) || (fakeRequest == deepCopy) {
		t.Errorf("Failed to deep copy a request %s: objects do not match or retained reference", diff.ObjectGoPrintSideBySide(fakeRequest, deepCopy))
	}
	t.Log("CheckpointedRequest.DeepCopy() returns correct result")
}

func TestCheckpointedRequest_ToCqlModel(t *testing.T) {
	fakeRequest := getFakeRequest()
	cqlModel, err := fakeRequest.ToCqlModel()
	if err != nil {
		t.Errorf("Error when creating CQL model: %s", err)
	}
	expectedCqlModel := &CheckpointedRequestCqlModel{
		Algorithm:               fakeRequest.Algorithm,
		Id:                      fakeRequest.Id,
		LifecycleStage:          "RUNNING",
		PayloadUri:              fakeRequest.PayloadUri,
		ResultUri:               fakeRequest.ResultUri,
		AlgorithmFailureCause:   fakeRequest.AlgorithmFailureCause,
		AlgorithmFailureDetails: fakeRequest.AlgorithmFailureDetails,
		ReceivedByHost:          fakeRequest.ReceivedByHost,
		ReceivedAt:              fakeRequest.ReceivedAt,
		SentAt:                  fakeRequest.SentAt,
		AppliedConfiguration:    "{\"container\":{\"image\":\"test.io\",\"registry\":\"algorithms/test\",\"versionTag\":\"v1.0.0\",\"serviceAccountName\":\"test-sa\"},\"computeResources\":{\"cpuLimit\":\"1000m\",\"memoryLimit\":\"2000Mi\"},\"workgroupRef\":{\"name\":\"test-workgroup\",\"group\":\"nexus-workgroup.io\",\"kind\":\"KarpenterWorkgroupV1\"},\"command\":\"python\",\"args\":[\"job.py\",\"--sas-uri=%s\",\"--request-id=%s\",\"--arg1=true\"],\"runtimeEnvironment\":{\"deadlineSeconds\":120,\"maximumRetries\":3},\"datadogIntegrationSettings\":{\"mountDatadogSocket\":true}}",
		ConfigurationOverrides:  "{}",
		ContentHash:             fakeRequest.ContentHash,
		LastModified:            fakeRequest.LastModified,
		Tag:                     fakeRequest.Tag,
		ApiVersion:              fakeRequest.ApiVersion,
		JobUid:                  fakeRequest.JobUid,
		Parent:                  "{}",
		PayloadValidFor:         "86400s",
	}

	if !reflect.DeepEqual(expectedCqlModel, cqlModel) {
		t.Errorf("Failed to convert request to a cql model %s: values do not match", diff.ObjectGoPrintSideBySide(expectedCqlModel, cqlModel))
	}
	t.Log("CheckpointedRequest.ToCqlModel() returns correct result")
}

func TestCheckpointedRequest_FromCqlModel(t *testing.T) {
	fakeRequest := getFakeRequest()
	cqlModel, err := fakeRequest.ToCqlModel()
	if err != nil {
		t.Errorf("Error when creating CQL model: %s", err)
	}
	fakeRequestFromModel, err := cqlModel.FromCqlModel()

	if err != nil {
		t.Errorf("Error when converting a CQL model back to a checkpoint: %s", err)
	}

	if !reflect.DeepEqual(fakeRequest, fakeRequestFromModel) {
		t.Errorf("Failed to deserialize a checkpoint from its cql model %s: values do not match", diff.ObjectGoPrintSideBySide(fakeRequest, fakeRequestFromModel))
	}
	t.Log("CheckpointedRequest.FromCqlModel() returns correct result")
}

func TestCheckpointedRequest_ToV1Job(t *testing.T) {
	fakeRequest := getFakeRequest()
	job := fakeRequest.ToV1Job("v0.0.0", &v1.NexusAlgorithmWorkgroupSpec{
		Description:  "default",
		Capabilities: nil,
		Cluster:      "shard-0",
		Tolerations:  nil,
		Affinity:     nil,
	}, &metav1.OwnerReference{
		APIVersion: v1.SchemeGroupVersion.String(),
		Kind:       "NexusAlgorithmTemplate",
		Name:       "test-parent",
		UID:        "test-uid",
	})

	expectedJob := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-id",
			Labels: map[string]string{
				JobTemplateNameKey:          fakeRequest.Algorithm,
				JobLabelFrameworkVersionKey: "v0.0.0",
				NexusComponentLabel:         JobLabelAlgorithmRun,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: v1.SchemeGroupVersion.String(),
					Kind:       "NexusAlgorithmTemplate",
					Name:       "test-parent",
					UID:        "test-uid",
				},
			},
		},
		Spec: batchv1.JobSpec{
			PodFailurePolicy: &batchv1.PodFailurePolicy{
				Rules: []batchv1.PodFailurePolicyRule{
					{
						Action: "Ignore",
						OnPodConditions: []batchv1.PodFailurePolicyOnPodConditionsPattern{
							{
								Type:   "DisruptionTarget",
								Status: "True",
							},
						},
					},
				},
			},
			BackoffLimit:            ptr.Int32(3),
			TTLSecondsAfterFinished: ptr.Int32(300),
			ActiveDeadlineSeconds:   ptr.Int64(120),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						JobTemplateNameKey:          fakeRequest.Algorithm,
						JobLabelFrameworkVersionKey: "v0.0.0",
						NexusComponentLabel:         JobLabelAlgorithmRun,
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "dsdsocket",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/run/datadog/",
								},
							},
						},
					},
					ServiceAccountName: "test-sa",
					RestartPolicy:      corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:            "test-id",
							Image:           "algorithms/test/test.io:v1.0.0",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"python",
							},
							Args: []string{
								"job.py",
								"--sas-uri=https://somewhere",
								"--request-id=test-id",
								"--arg1=true",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "NEXUS__ALGORITHM_NAME",
									Value: "test",
								},
								{
									Name:  "NEXUS__SHARD_NAME",
									Value: "shard-0",
								},
								{
									Name:  "NEXUS__PARENT_REQUEST_ID",
									Value: "test-parent",
								},
								{
									Name:  "NEXUS__PARENT_ALGORITHM_NAME",
									Value: "test-parent-algorithm",
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1000m"),
									corev1.ResourceMemory: resource.MustParse("2000Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("2000Mi"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dsdsocket",
									MountPath: "/var/run/datadog",
									ReadOnly:  false,
								},
							},
						},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(*expectedJob, job) {
		t.Errorf("Generated job does not match expected job, diff: %s", diff.ObjectGoPrintSideBySide(*expectedJob, job))
	}
}

func TestFromAlgorithmRequest(t *testing.T) {
	checkpoint, payload, err := FromAlgorithmRequest("test-id", "test", &AlgorithmRequest{
		AlgorithmParameters: map[string]interface{}{
			"parameterA": "a",
			"parameterB": "b",
		},
		CustomConfiguration: nil,
		RequestApiVersion:   "",
		Tag:                 "test",
		ParentRequest:       nil,
		PayloadValidFor:     "24h",
	}, &v1.NexusAlgorithmSpec{
		Container:        nil,
		ComputeResources: nil,
		WorkgroupRef: &v1.NexusAlgorithmWorkgroupRef{
			Name:  "test",
			Group: "science.sneaksanddata.com/v1",
			Kind:  "NexusAlgorithmWorkgroup",
		},
		Command:                    "python",
		Args:                       nil,
		RuntimeEnvironment:         nil,
		ErrorHandlingBehaviour:     nil,
		DatadogIntegrationSettings: nil,
	})

	if err != nil {
		t.Errorf("failed to create a checkpoint: %s", err)
	}

	if string(payload) != "{\"parameterA\":\"a\",\"parameterB\":\"b\"}" {
		t.Errorf("serialized payload does not match expected: %s", string(payload))
	}

	if checkpoint == nil {
		t.Errorf("checkpoint is not expected to be nil")
	}

	if checkpoint != nil && checkpoint.Id != "test-id" {
		t.Errorf("checkpoint id does not match expected: %s", checkpoint.Id)
	}
}

func TestCheckpointedRequest_IsFinished(t *testing.T) {
	request := getFakeRequest()

	for _, stage := range []string{LifecycleStageFailed, LifecycleStageCompleted, LifecycleStageDeadlineExceeded, LifecycleStageSchedulingFailed, LifecycleStageCancelled} {
		request.LifecycleStage = stage

		if !request.IsFinished() {
			t.Errorf("a %s request is expected to be finished", stage)
		}
	}

	for _, stage := range []string{LifecycleStageRunning, LifecycleStageBuffered, LifecycleStageNew} {
		request.LifecycleStage = stage

		if request.IsFinished() {
			t.Errorf("a %s request is not expected to be finished", stage)
		}
	}
}
