package request

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-go/v5/statsd"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/payload"
	"github.com/SneaksAndData/nexus-core/pkg/pipeline"
	"github.com/SneaksAndData/nexus-core/pkg/telemetry"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"path"
	"time"
)

type Buffer interface {
	BufferRequest(ctx context.Context, requestId string, request *models.AlgorithmRequest, config *v1.MachineLearningAlgorithmSpec) (*models.SubmissionBufferEntry, error)
}

type BufferConfig struct {
	PayloadStoragePath         string
	PayloadValidFor            time.Duration
	FailureRateBaseDelay       time.Duration
	FailureRateMaxDelay        time.Duration
	RateLimitElementsPerSecond int
	RateLimitElementsBurst     int
	Workers                    int
}

type BufferInput struct {
	Checkpoint        *models.CheckpointedRequest
	SerializedPayload []byte
	Config            *v1.MachineLearningAlgorithmSpec
}

type BufferOutput struct {
	checkpoint *models.CheckpointedRequest
	entry      *models.SubmissionBufferEntry
}

func (input *BufferInput) tags() map[string]string {
	return map[string]string{
		"algorithm": input.Checkpoint.Algorithm,
	}
}

func newBufferInput(requestId string, request *models.AlgorithmRequest, config *v1.MachineLearningAlgorithmSpec) (*BufferInput, error) {
	checkpoint, serializedPayload, err := models.FromAlgorithmRequest(requestId, request, config)

	if err != nil {
		return nil, err
	}

	return &BufferInput{
		Checkpoint:        checkpoint,
		SerializedPayload: serializedPayload,
		Config:            config,
	}, nil
}

type DefaultBuffer struct {
	checkpointStore CheckpointStore
	metadataStore   MetadataStore
	blobStore       payload.BlobStore
	bufferConfig    BufferConfig
	logger          klog.Logger
	metrics         *statsd.Client
	ctx             context.Context
	actor           *pipeline.DefaultPipelineStageActor[*BufferInput, *BufferOutput]
}

func (buffer *DefaultBuffer) Start(submitter pipeline.StageActor[*BufferOutput, types.UID]) {
	buffer.actor = pipeline.NewDefaultPipelineStageActor[*BufferInput, *BufferOutput](
		buffer.bufferConfig.FailureRateBaseDelay,
		buffer.bufferConfig.FailureRateMaxDelay,
		buffer.bufferConfig.RateLimitElementsPerSecond,
		buffer.bufferConfig.RateLimitElementsBurst,
		buffer.bufferConfig.Workers,
		buffer.bufferRequest,
		submitter,
	)

	buffer.actor.Start(buffer.ctx)
}

func (buffer *DefaultBuffer) Add(requestId string, request *models.AlgorithmRequest, config *v1.MachineLearningAlgorithmSpec) error {
	input, err := newBufferInput(requestId, request, config)
	if err != nil {
		return err
	}

	buffer.actor.Receive(input)
	return nil
}

func (buffer *DefaultBuffer) bufferRequest(input *BufferInput) (*BufferOutput, error) {
	telemetry.Increment(buffer.metrics, "incoming_requests", input.tags())

	payloadPath := path.Join(
		buffer.bufferConfig.PayloadStoragePath,
		fmt.Sprintf("algorithm=%s", input.Checkpoint.Algorithm),
		input.Checkpoint.Id)

	if err := buffer.checkpointStore.UpsertCheckpoint(input.Checkpoint); err != nil {
		return nil, err
	}

	// TODO: add parent handling
	if err := buffer.blobStore.SaveTextAsBlob(buffer.ctx, string(input.SerializedPayload), payloadPath); err != nil {
		return nil, err
	}

	bufferedCheckpoint := input.Checkpoint.DeepCopy()
	payloadUri, err := buffer.blobStore.GetBlobUri(buffer.ctx, payloadPath, buffer.bufferConfig.PayloadValidFor)
	if err != nil {
		return nil, err
	}

	bufferedCheckpoint.PayloadUri = payloadUri
	bufferedCheckpoint.LifecycleStage = models.LifecyclestageBuffered
	bufferedEntry := models.FromCheckpoint(bufferedCheckpoint)

	if err := buffer.metadataStore.UpsertMetadata(bufferedEntry); err != nil {
		return nil, err
	}

	telemetry.Increment(buffer.metrics, "checkpoint_requests", input.tags())

	return &BufferOutput{
		checkpoint: bufferedCheckpoint,
		entry:      bufferedEntry,
	}, nil
}