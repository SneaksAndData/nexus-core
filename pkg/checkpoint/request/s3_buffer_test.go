/*
*

	 IMPORTANT: Tests use data points for test-resources/checkpoints.cql. If you update that file, make sure to refresh test expectations.
	*
*/
package request

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
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
