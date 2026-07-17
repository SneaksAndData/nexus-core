/*
*

	 IMPORTANT: Tests use data points for test-resources/checkpoints.cql. If you update that file, make sure to refresh test expectations.
	*
*/
package request

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/payload"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/store/cassandra"
	"github.com/aws/smithy-go/ptr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2/ktesting"
)

type fixture struct {
	buffer *DefaultBuffer
}

func newIndexedCassandraConfig() *cassandra.ScyllaConfig {
	return &cassandra.ScyllaConfig{
		Hosts:            []string{"127.0.0.1"},
		IndexesSupported: true,
	}
}

func newBareCassandraConfig() *cassandra.ScyllaConfig {
	return &cassandra.ScyllaConfig{
		Hosts:            []string{"127.0.0.1"},
		IndexesSupported: false,
	}
}

func newFixture(t *testing.T, config *cassandra.ScyllaConfig) *fixture {
	_, ctx := ktesting.NewTestContext(t)
	f := &fixture{}
	f.buffer = NewScyllaS3Buffer(ctx, &S3BufferConfig{
		BufferConfig: &BufferConfig{
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
		RequestPayloadProxyConfiguration: &payload.RequestPayloadProxyConfiguration{
			TenantId:          "test-tenant",
			ServePathTemplate: "/data/v1/payloads/%s/%s",
			SignSecret:        "test-secret",
		},
	}, config, map[string]string{})

	return f
}

func waitForBuffer(t *testing.T, f *fixture) {
	err := wait.PollUntilContextTimeout(t.Context(), 1*time.Second, 5*time.Second, true, func(ctx context.Context) (done bool, err error) {
		if f.buffer.actor != nil {
			return f.buffer.actor.IsRunning(), nil
		}

		return false, nil
	})
	if err != nil {
		t.Fatalf("error while waiting for buffer to startup %s", err)
	}
}

func waitForBuffered(t *testing.T, f *fixture, expectedRequestId string, expectedTemplateName string, timeout time.Duration) {
	err := wait.PollUntilContextTimeout(t.Context(), 1*time.Second, timeout, true, func(ctx context.Context) (done bool, err error) {
		checkpoint, err := f.buffer.Get(expectedRequestId, expectedTemplateName)

		if err != nil {
			return true, fmt.Errorf("error when reading an expected checkpoint entry: %v", err)
		}

		if checkpoint == nil {
			return false, nil
		}

		if checkpoint.LifecycleStage == models.LifecycleStageBuffered {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		t.Fatalf("error when waiting for buffer operation result %s", err)
	}

	checkpoint, err := f.buffer.Get(expectedRequestId, expectedTemplateName)

	if err != nil {
		t.Fatalf("error when verifying buffer operation result %s", err)
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

func TestDefaultBuffer_Get(t *testing.T) {
	cases := []struct {
		name    string
		fixture *fixture
	}{
		{"get from buffer with indexed Cassandra store", newFixture(t, newIndexedCassandraConfig())},
		{"get from buffer with bare Cassandra store", newFixture(t, newBareCassandraConfig())},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkpoint, err := tc.fixture.buffer.Get("f47ac10b-58cc-4372-a567-0e02b2c3d479", "test-algorithm")

			if err != nil {
				t.Fatalf("error when reading a checkpoint: %v", err)
			}

			if checkpoint != nil && checkpoint.JobUid != "d94c16c8-2c1e-4f3a-85d1-2d9c3b7f0a24" && checkpoint.LifecycleStage != models.LifecycleStageNew {
				t.Fatalf("expected LifecycleStage = NEW, but got %s and expected JobUid = d94c16c8-2c1e-4f3a-85d1-2d9c3b7f0a24, but got %s", checkpoint.LifecycleStage, checkpoint.JobUid)
			}
		})
	}
}

func TestDefaultBuffer_GetBuffered(t *testing.T) {
	cases := []struct {
		name    string
		fixture *fixture
	}{
		{"get from buffer by host (buffered) with indexed Cassandra store", newFixture(t, newIndexedCassandraConfig())},
		{"get from buffer by host (buffered) with bare Cassandra store", newFixture(t, newBareCassandraConfig())},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkpoints, err := tc.fixture.buffer.GetBuffered("host123")

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
		})
	}
}

func TestDefaultBuffer_GetBuffered_DoesNotExist(t *testing.T) {
	cases := []struct {
		name    string
		fixture *fixture
	}{
		{"get buffered, not existing with indexed Cassandra store", newFixture(t, newIndexedCassandraConfig())},
		{"get buffered, not existing with bare Cassandra store", newFixture(t, newBareCassandraConfig())},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkpoint, err := tc.fixture.buffer.Get("something", "not-existing")

			if err != nil {
				t.Fatalf("error when reading a non-existent checkpoint: %v", err)
			}

			if checkpoint != nil {
				t.Fatalf("checkpoint should be nil, but got %s", checkpoint.Id)
			}
		})
	}
}

func TestDefaultBuffer_GetBufferedByHost_DoesNotExist(t *testing.T) {
	cases := []struct {
		name    string
		fixture *fixture
	}{
		{"get from buffer by host (buffered, not existing) with indexed Cassandra store", newFixture(t, newIndexedCassandraConfig())},
		{"get from buffer by host (buffered, not existing) with bare Cassandra store", newFixture(t, newBareCassandraConfig())},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkpoints, err := tc.fixture.buffer.GetBuffered("non-existent-host")

			if err != nil {
				t.Fatalf("error when reading a non-existent checkpoint: %v", err)
			}

			if checkpoints == nil {
				t.Fatal("checkpoints should not be nil")
			}

			result := []*models.CheckpointedRequest{}

			for checkpoint, err := range checkpoints {
				if err != nil {
					t.Fatalf("error when deserializing a buffered checkpoint: %v", err)
				}

				result = append(result, checkpoint)
			}

			if len(result) > 0 {
				t.Fatalf("checkpoints should be empty, but got %d", len(result))
			}
		})
	}
}

func TestDefaultBuffer_GetNew(t *testing.T) {
	cases := []struct {
		name    string
		fixture *fixture
	}{
		{"get from buffer by host (new) with indexed Cassandra store", newFixture(t, newIndexedCassandraConfig())},
		{"get from buffer by host (new) with bare Cassandra store", newFixture(t, newBareCassandraConfig())},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkpoints, err := tc.fixture.buffer.GetNew("host123")

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
		})
	}
}

func TestDefaultBuffer_GetTagged(t *testing.T) {
	cases := []struct {
		name    string
		fixture *fixture
	}{
		{"get tagged with indexed Cassandra store", newFixture(t, newIndexedCassandraConfig())},
		{"get tagged with bare Cassandra store", newFixture(t, newBareCassandraConfig())},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkpoints, err := tc.fixture.buffer.GetTagged("running_tag")

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
		})
	}
}

func TestDefaultBuffer_GetMetadata(t *testing.T) {
	cases := []struct {
		name    string
		fixture *fixture
	}{
		{"get metadata with indexed Cassandra store", newFixture(t, newIndexedCassandraConfig())},
		{"get metadata with bare Cassandra store", newFixture(t, newBareCassandraConfig())},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			metadataEntry, err := tc.fixture.buffer.GetBufferedEntry(&models.CheckpointedRequest{Id: "123e4567-e89b-12d3-a456-426614174000", Algorithm: "test-algorithm"})

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
		})
	}
}

func TestDefaultBuffer_Add_Retrieve(t *testing.T) {
	cases := []struct {
		name              string
		fixture           *fixture
		serializationMode v1.PayloadSerializationMode
	}{
		{"add to buffer that uses indexed Cassandra store, with S3 serialization", newFixture(t, newIndexedCassandraConfig()), v1.SERIALIZE_TO_S3},
		{"add to buffer that uses bare Cassandra store, with S3 serialization", newFixture(t, newBareCassandraConfig()), v1.SERIALIZE_TO_S3},
		{"add to buffer that uses indexed Cassandra store, without S3 serialization", newFixture(t, newIndexedCassandraConfig()), v1.SERIALIZE_TO_BACKEND},
		{"add to buffer that uses bare Cassandra store, without S3 serialization", newFixture(t, newBareCassandraConfig()), v1.SERIALIZE_TO_BACKEND},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var testPayload = map[string]interface{}{
				"parameterA": "a",
				"parameterB": "b",
			}

			go tc.fixture.buffer.Start(nil)

			waitForBuffer(t, tc.fixture)

			err := tc.fixture.buffer.Add("new-id", "test-algorithm-v2", &models.AlgorithmRequest{
				AlgorithmParameters: testPayload,
				CustomConfiguration: nil,
				RequestApiVersion:   "",
				Tag:                 "",
				ParentRequest: &models.AlgorithmRequestRef{
					RequestId:     "test-parent",
					AlgorithmName: "test-algorithm-v2",
				},
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
				PayloadConfiguration: &v1.NexusAlgorithmPayloadConfiguration{
					PayloadValidFor:          "24h",
					PayloadSerializationMode: tc.serializationMode,
				},
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

			waitForBuffered(t, tc.fixture, "new-id", "test-algorithm-v2", 5*time.Second)

			retrieved, err := tc.fixture.buffer.GetPersisted("new-id", "test-algorithm-v2")

			if err != nil {
				t.Fatalf("error when retrieving checkpoint: %v", err)
			}

			var storedPayload map[string]interface{}
			err = json.Unmarshal(retrieved, &storedPayload)
			if err != nil {
				t.Fatalf("error when unmarshalling stored payload: %v", err)
			}

			if !reflect.DeepEqual(storedPayload, testPayload) {
				t.Fatalf("stored payload is not equal to the test payload %v", diff.ObjectGoPrintSideBySide(storedPayload, testPayload))
			}
		})
	}
}
