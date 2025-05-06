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
	"time"
)

type Buffer interface {
	BufferRequest(ctx context.Context, requestId string, request *models.AlgorithmRequest, config *v1.NexusAlgorithmSpec) (*models.SubmissionBufferEntry, error)
}

type BufferConfig struct {
	PayloadStoragePath         string        `mapstructure:"payload-storage-path"`
	PayloadValidFor            time.Duration `mapstructure:"payload-valid-for"`
	FailureRateBaseDelay       time.Duration `mapstructure:"failure-rate-base-delay"`
	FailureRateMaxDelay        time.Duration `mapstructure:"failure-rate-max-delay"`
	RateLimitElementsPerSecond int           `mapstructure:"rate-limit-elements-per-second"`
	RateLimitElementsBurst     int           `mapstructure:"rate-limit-elements-burst"`
	Workers                    int           `mapstructure:"workers"`
}

type BufferInput struct {
	Checkpoint        *models.CheckpointedRequest
	ResolvedWorkgroup *v1.NexusAlgorithmWorkgroupSpec
	ResolvedShardName string
	SerializedPayload []byte
	Config            *v1.NexusAlgorithmSpec
}

type BufferOutput struct {
	Checkpoint *models.CheckpointedRequest
	Entry      *models.SubmissionBufferEntry
}

func (input *BufferInput) tags() map[string]string {
	return map[string]string{
		"algorithm": input.Checkpoint.Algorithm,
	}
}

func newBufferInput(requestId string, algorithmName string, request *models.AlgorithmRequest, config *v1.NexusAlgorithmSpec) (*BufferInput, error) {
	checkpoint, serializedPayload, err := models.FromAlgorithmRequest(requestId, algorithmName, request, config)

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
	bufferConfig    *BufferConfig
	logger          *klog.Logger
	metrics         *statsd.Client
	ctx             context.Context
	actor           *pipeline.DefaultPipelineStageActor[*BufferInput, *BufferOutput]
	name            string
	tags            map[string]string
}

// NewDefaultBuffer creates a default buffer that uses Astra DB for checkpointing and S3-compatible storage for payload persistence
func NewDefaultBuffer(ctx context.Context, config *BufferConfig, astraConfig *AstraBundleConfig, metricTags map[string]string) *DefaultBuffer {
	logger := klog.FromContext(ctx)

	cqlStore := NewAstraCqlStore(logger, astraConfig)
	return &DefaultBuffer{
		checkpointStore: cqlStore,
		metadataStore:   cqlStore,
		blobStore:       payload.NewS3PayloadStore(ctx, logger),
		bufferConfig:    config,
		logger:          &logger,
		metrics:         ctx.Value(telemetry.MetricsClientContextKey).(*statsd.Client),
		ctx:             ctx,
		actor:           nil,
		name:            "default_cassandra_s3",
		tags:            metricTags,
	}
}

func (buffer *DefaultBuffer) Start(submitter pipeline.StageActor[*BufferOutput, types.UID]) {
	buffer.actor = pipeline.NewDefaultPipelineStageActor[*BufferInput, *BufferOutput](
		buffer.name,
		buffer.tags,
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

func (buffer *DefaultBuffer) Add(requestId string, algorithmName string, request *models.AlgorithmRequest, config *v1.NexusAlgorithmSpec) error {
	input, err := newBufferInput(requestId, algorithmName, request, config)
	if err != nil {
		return err
	}

	buffer.actor.Receive(input)
	return nil
}

func (buffer *DefaultBuffer) Get(requestId string, algorithmName string) (*models.CheckpointedRequest, error) {
	return buffer.checkpointStore.ReadCheckpoint(algorithmName, requestId)
}

func (buffer *DefaultBuffer) bufferRequest(input *BufferInput) (*BufferOutput, error) {
	telemetry.Increment(buffer.metrics, "incoming_requests", input.tags())
	buffer.logger.V(4).Info("Persisting payload", "request", input.Checkpoint.Id, "algorithm", input.Checkpoint.Algorithm)

	payloadPath := fmt.Sprintf("%s/%s/%s",
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
	bufferedEntry := models.FromCheckpoint(bufferedCheckpoint, input.ResolvedWorkgroup)

	if err := buffer.metadataStore.UpsertMetadata(bufferedEntry); err != nil {
		return nil, err
	}

	telemetry.Increment(buffer.metrics, "checkpoint_requests", input.tags())

	return &BufferOutput{
		Checkpoint: bufferedCheckpoint,
		Entry:      bufferedEntry,
	}, nil
}
