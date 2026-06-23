package util_test

import (
	"time"

	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/aws/smithy-go/ptr"
)

func GetFakeRequest(hasParent bool) *models.CheckpointedRequest {
	var parentRef *models.AlgorithmRequestRef

	if hasParent {
		parentRef = &models.AlgorithmRequestRef{RequestId: "test-parent", AlgorithmName: "test-parent-algorithm"}
	}

	return &models.CheckpointedRequest{
		Algorithm:               "test",
		Id:                      "test-id",
		LifecycleStage:          models.LifecycleStageRunning,
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
		Parent:                 parentRef,
		PayloadValidFor:        "86400s",
	}
}
