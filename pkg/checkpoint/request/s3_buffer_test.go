/*
*

	 IMPORTANT: Tests use data points for test-resources/checkpoints.cql. If you update that file, make sure to refresh test expectations.
	*
*/
package request

import (
	"testing"
	"time"

	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/store/cassandra"
	"github.com/aws/smithy-go/ptr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2/ktesting"
)

type fixture struct {
	buffer *DefaultBuffer
}

func newFixture(t *testing.T) *fixture {
	_, ctx := ktesting.NewTestContext(t)
	f := &fixture{}
	f.buffer = NewScyllaS3Buffer(ctx, &S3BufferConfig{
		BufferConfig: &BufferConfig{
			PayloadValidFor:            time.Hour,
			FailureRateBaseDelay:       time.Second,
			FailureRateMaxDelay:        time.Second * 2,
			RateLimitElementsPerSecond: 10,
			RateLimitElementsBurst:     100,
			Workers:                    2,
		},
		AccessKeyID:        "minioadmin",
		SecretAccessKey:    "minioadmin",
		Region:             "us-east-1",
		Endpoint:           "http://localhost:9000",
		PayloadStoragePath: "s3a://nexus",
	}, &cassandra.ScyllaConfig{
		Hosts: []string{"127.0.0.1"},
	}, map[string]string{})

	return f
}

func TestDefaultBuffer_Get(t *testing.T) {
	f := newFixture(t)

	checkpoint, err := f.buffer.Get("f47ac10b-58cc-4372-a567-0e02b2c3d479", "test-algorithm")

	if err != nil {
		t.Fatalf("error when reading a checkpoint: %v", err)
	}

	if checkpoint != nil && checkpoint.JobUid != "d94c16c8-2c1e-4f3a-85d1-2d9c3b7f0a24" && checkpoint.LifecycleStage != models.LifecycleStageNew {
		t.Fatalf("expected LifecycleStage = NEW, but got %s and expected JobUid = d94c16c8-2c1e-4f3a-85d1-2d9c3b7f0a24, but got %s", checkpoint.LifecycleStage, checkpoint.JobUid)
	}
}

func TestDefaultBuffer_GetBuffered(t *testing.T) {
	f := newFixture(t)
	checkpoints, err := f.buffer.GetBuffered("host123")

	if err != nil {
		t.Fatalf("error when reading buffered checkpoints by host: %v", err)
	}

	result := []*models.CheckpointedRequest{}

	for checkpoint, err := range checkpoints {
		if err != nil {
			t.Fatalf("error when deserializing a buffered checkpoint: %v", err)
		}

		result = append(result, checkpoint)
	}

	if len(result) != 1 {
		t.Fatalf("expected only one checkpoint, but got %d", len(result))
	}

	if result[0].Id != "2c7b6e8d-cc3c-4b5b-a3f6-5d7b9e2c7f2a" {
		t.Fatalf("Only a checkpoint with id 2c7b6e8d-cc3c-4b5b-a3f6-5d7b9e2c7f2a should be BUFFERED for host123, but found %s", result[0].Id)
	}
}

func TestDefaultBuffer_GetNew(t *testing.T) {
	f := newFixture(t)
	checkpoints, err := f.buffer.GetNew("host123")

	if err != nil {
		t.Fatalf("error when reading NEW checkpoints by host: %v", err)
	}

	result := []*models.CheckpointedRequest{}

	for checkpoint, err := range checkpoints {
		if err != nil {
			t.Fatalf("error when deserializing a NEW checkpoint: %v", err)
		}

		result = append(result, checkpoint)
	}

	if len(result) != 1 {
		t.Fatalf("expected only one checkpoint, but got %d", len(result))
	}

	if result[0].Id != "f47ac10b-58cc-4372-a567-0e02b2c3d479" {
		t.Fatalf("Only a checkpoint with id f47ac10b-58cc-4372-a567-0e02b2c3d479 should be NEW for host123, but found %s", result[0].Id)
	}
}

func TestDefaultBuffer_GetTagged(t *testing.T) {
	f := newFixture(t)

	checkpoints, err := f.buffer.GetTagged("running_tag")

	if err != nil {
		t.Fatalf("error when reading checkpoints by tag: %v", err)
	}

	result := []*models.CheckpointedRequest{}

	for checkpoint, err := range checkpoints {
		if err != nil {
			t.Fatalf("error when deserializing a checkpoint: %v", err)
		}

		result = append(result, checkpoint)
	}

	if len(result) != 1 {
		t.Fatalf("expected only one checkpoint, but got %d", len(result))
	}

	if result[0].Id != "8a0c8aa9-fc9d-4b2e-9c5c-2c8d7f1e7a3f" {
		t.Fatalf("Only a checkpoint with id 8a0c8aa9-fc9d-4b2e-9c5c-2c8d7f1e7a3f should be RUNNING with tag running_tag, but found %s", result[0].Id)
	}
}

func TestDefaultBuffer_GetMetadata(t *testing.T) {
	f := newFixture(t)

	metadataEntry, err := f.buffer.GetBufferedEntry(&models.CheckpointedRequest{Id: "123e4567-e89b-12d3-a456-426614174000", Algorithm: "test-algorithm"})

	if err != nil {
		t.Fatalf("error when reading checkpoint metadata: %v", err)
	}

	if metadataEntry == nil {
		t.Fatalf("checkpoint metadata entry must not be nil")
	}

	job, err := metadataEntry.SubmissionTemplate()

	if err != nil {
		t.Fatalf("error when deserializing checkpoint metadata: %v", err)
	}

	if job == nil {
		t.Fatalf("checkpoint metadata job must not be nil")
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
		ParentRequest: &models.AlgorithmRequestRef{
			RequestId:     "test-parent",
			AlgorithmName: "test-algorithm-v2",
		},
		PayloadValidFor: "24h",
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
	}, &metav1.OwnerReference{
		APIVersion: v1.SchemeGroupVersion.String(),
		Kind:       "NexusAlgorithmTemplate",
		Name:       "test-parent",
		UID:        "test-parent-uid",
	}, false)

	if err != nil {
		t.Fatalf("error when buffering checkpoint: %v", err)
	}

	time.Sleep(time.Second * 5)

	checkpoint, err := f.buffer.Get("new-id", "test-algorithm-v2")

	if err != nil {
		t.Fatalf("error when reading an expected checkpoint entry: %v", err)
	}

	if checkpoint == nil {
		t.Fatalf("expected checkpoint not found in the buffer after calling Add")
	}

	if checkpoint.LifecycleStage != models.LifecycleStageBuffered {
		t.Fatalf("lifecycle stage should be Buffered, but is %s", checkpoint.LifecycleStage)
	}

	if checkpoint.Parent == nil {
		t.Fatalf("parent should not be nil")
	}

	if checkpoint.Parent.RequestId != "test-parent" {
		t.Fatalf("parent request id should be test-parent-uid, but is %s", checkpoint.Parent.RequestId)
	}
}
