/*
*

	 IMPORTANT: Tests use data points for test-resources/checkpoints.cql. If you update that file, make sure to refresh test expectations.
	*
*/
package request

import (
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/aws/smithy-go/ptr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2/ktesting"
	"testing"
	"time"
)

type fixture struct {
	buffer *DefaultBuffer
}

func newFixture(t *testing.T) *fixture {
	_, ctx := ktesting.NewTestContext(t)
	f := &fixture{}
	f.buffer = NewScyllaS3Buffer(ctx, &S3BufferConfig{
		BufferConfig: &BufferConfig{
			PayloadStoragePath:         "s3a://nexus",
			PayloadValidFor:            time.Hour,
			FailureRateBaseDelay:       time.Second,
			FailureRateMaxDelay:        time.Second * 2,
			RateLimitElementsPerSecond: 10,
			RateLimitElementsBurst:     100,
			Workers:                    2,
		},
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		Region:          "us-east-1",
		Endpoint:        "http://localhost:9000",
	}, &ScyllaCqlStoreConfig{
		Hosts: []string{"127.0.0.1"},
	}, map[string]string{})

	return f
}

func TestDefaultBuffer_Get(t *testing.T) {
	f := newFixture(t)

	checkpoint, err := f.buffer.Get("f47ac10b-58cc-4372-a567-0e02b2c3d479", "test-algorithm")

	if err != nil {
		t.Errorf("error when reading a checkpoint: %v", err)
		t.FailNow()
	}

	if checkpoint != nil && checkpoint.JobUid != "d94c16c8-2c1e-4f3a-85d1-2d9c3b7f0a24" && checkpoint.LifecycleStage != models.LifecycleStageNew {
		t.Errorf("expected LifecycleStage = NEW, but got %s and expected JobUid = d94c16c8-2c1e-4f3a-85d1-2d9c3b7f0a24, but got %s", checkpoint.LifecycleStage, checkpoint.JobUid)
	}
}

func TestDefaultBuffer_GetBuffered(t *testing.T) {
	f := newFixture(t)
	checkpoints, err := f.buffer.GetBuffered("host123")

	if err != nil {
		t.Errorf("error when reading buffered checkpoints by host: %v", err)
		t.FailNow()
	}

	result := []*models.CheckpointedRequest{}

	for checkpoint, err := range checkpoints {
		if err != nil {
			t.Errorf("error when deserializing a buffered checkpoint: %v", err)
			t.FailNow()
		}

		result = append(result, checkpoint)
	}

	if len(result) != 1 {
		t.Errorf("expected only one checkpoint, but got %d", len(result))
		t.FailNow()
	}

	if result[0].Id != "2c7b6e8d-cc3c-4b5b-a3f6-5d7b9e2c7f2a" {
		t.Errorf("Only a checkpoint with id 2c7b6e8d-cc3c-4b5b-a3f6-5d7b9e2c7f2a should be BUFFERED for host123, but found %s", result[0].Id)
	}
}

func TestDefaultBuffer_GetTagged(t *testing.T) {
	f := newFixture(t)

	checkpoints, err := f.buffer.GetTagged("running_tag")

	if err != nil {
		t.Errorf("error when reading checkpoints by tag: %v", err)
		t.FailNow()
	}

	result := []*models.CheckpointedRequest{}

	for checkpoint, err := range checkpoints {
		if err != nil {
			t.Errorf("error when deserializing a checkpoint: %v", err)
			t.FailNow()
		}

		result = append(result, checkpoint)
	}

	if len(result) != 1 {
		t.Errorf("expected only one checkpoint, but got %d", len(result))
		t.FailNow()
	}

	if result[0].Id != "8a0c8aa9-fc9d-4b2e-9c5c-2c8d7f1e7a3f" {
		t.Errorf("Only a checkpoint with id 8a0c8aa9-fc9d-4b2e-9c5c-2c8d7f1e7a3f should be RUNNING with tag running_tag, but found %s", result[0].Id)
	}
}

func TestDefaultBuffer_GetMetadata(t *testing.T) {
	f := newFixture(t)

	metadataEntry, err := f.buffer.GetBufferedEntry(&models.CheckpointedRequest{Id: "123e4567-e89b-12d3-a456-426614174000", Algorithm: "test-algorithm"})

	if err != nil {
		t.Errorf("error when reading checkpoint metadata: %v", err)
		t.FailNow()
	}

	if metadataEntry == nil {
		t.Errorf("checkpoint metadata entry must not be nil")
		t.FailNow()
	}

	job, err := metadataEntry.SubmissionTemplate()

	if err != nil {
		t.Errorf("error when deserializing checkpoint metadata: %v", err)
	}

	if job == nil {
		t.Errorf("checkpoint metadata job must not be nil")
	}
}

func TestDefaultBuffer_Add(t *testing.T) {
	f := newFixture(t)

	go f.buffer.Start(nil)

	time.Sleep(1 * time.Second)

	err := f.buffer.Add("new-id", "test-algorithm-v2", &models.AlgorithmRequest{
		AlgorithmParameters: map[string]interface{}{
			"parameterA": "a",
			"parameterB": "b",
		},
		CustomConfiguration: nil,
		RequestApiVersion:   "",
		Tag:                 "",
		ParentRequest:       nil,
		PayloadValidFor:     "24h",
	}, &v1.NexusAlgorithmSpec{
		Container: &v1.NexusAlgorithmContainer{
			Image:              "test-image",
			Registry:           "test",
			VersionTag:         "v1.2.3",
			ServiceAccountName: "test-sa",
		},
		ComputeResources: &v1.NexusAlgorithmResources{
			CpuLimit:        "1000m",
			MemoryLimit:     "2000Mi",
			CustomResources: nil,
		},
		WorkgroupRef: &v1.NexusAlgorithmWorkgroupRef{
			Name:  "default",
			Group: "science.sneaksanddata.com/v1",
			Kind:  "NexusAlgorithmWorkgroup",
		},
		Command: "python",
		Args:    []string{"main.py", "--request-id=%s", "--sas-uri=%s"},
		RuntimeEnvironment: &v1.NexusAlgorithmRuntimeEnvironment{
			EnvironmentVariables: []corev1.EnvVar{
				{
					Name:  "TEST_ENV_VAR",
					Value: "TEST_VALUE",
				},
			},
			MappedEnvironmentVariables: []corev1.EnvFromSource{
				{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-secret",
						},
					},
				},
			},
			Annotations:     nil,
			DeadlineSeconds: ptr.Int32(300),
			MaximumRetries:  ptr.Int32(10),
		},
		ErrorHandlingBehaviour:     nil,
		DatadogIntegrationSettings: &v1.NexusDatadogIntegrationSettings{MountDatadogSocket: ptr.Bool(true)},
	}, &v1.NexusAlgorithmWorkgroupSpec{
		Description:  "test",
		Capabilities: nil,
		Cluster:      "test-shard",
		Tolerations:  nil,
		Affinity:     nil,
	})

	if err != nil {
		t.Errorf("error when buffering checkpoint: %v", err)
		t.FailNow()
	}

	time.Sleep(time.Second * 5)

	checkpoint, err := f.buffer.Get("new-id", "test-algorithm-v2")

	if err != nil {
		t.Errorf("error when reading an expected checkpoint entry: %v", err)
		t.FailNow()
	}

	if checkpoint == nil {
		t.Errorf("expected checkpoint not found in the buffer after calling Add")
	}

	f.buffer.ctx.Done()

	time.Sleep(1 * time.Second)
}
