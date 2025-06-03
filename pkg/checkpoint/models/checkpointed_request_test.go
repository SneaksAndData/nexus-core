package models

import (
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/aws/smithy-go/ptr"
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
			Args:    []string{"job.py", "--request-id 111-222-333 --arg1 true"},
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
		ParentJob:              nil,
		PayloadValidFor:        128,
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
		AppliedConfiguration:    "{\"container\":{\"image\":\"test.io\",\"registry\":\"algorithms/test\",\"versionTag\":\"v1.0.0\",\"serviceAccountName\":\"test-sa\"},\"computeResources\":{\"cpuLimit\":\"1000m\",\"memoryLimit\":\"2000Mi\"},\"workgroupRef\":{\"name\":\"test-workgroup\",\"group\":\"nexus-workgroup.io\",\"kind\":\"KarpenterWorkgroupV1\"},\"command\":\"python\",\"args\":[\"job.py\",\"--request-id 111-222-333 --arg1 true\"],\"runtimeEnvironment\":{\"deadlineSeconds\":120,\"maximumRetries\":3},\"errorHandlingBehaviour\":{},\"datadogIntegrationSettings\":{\"mountDatadogSocket\":true}}",
		ConfigurationOverrides:  "{}",
		ContentHash:             fakeRequest.ContentHash,
		LastModified:            fakeRequest.LastModified,
		Tag:                     fakeRequest.Tag,
		ApiVersion:              fakeRequest.ApiVersion,
		JobUid:                  fakeRequest.JobUid,
		ParentJob:               "",
		PayloadValidFor:         "128ns",
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
