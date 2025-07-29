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
	"github.com/SneaksAndData/nexus-core/pkg/util"
	s3credentials "github.com/aws/aws-sdk-go-v2/credentials"
	"iter"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type S3BufferConfig struct {
	BufferConfig    *BufferConfig `mapstructure:"buffer-config,omitempty"`
	AccessKeyID     string        `mapstructure:"access-key-id,omitempty"`
	SecretAccessKey string        `mapstructure:"secret-access-key,omitempty"`
	Region          string        `mapstructure:"region,omitempty"`
	Endpoint        string        `mapstructure:"endpoint,omitempty"`
}

type DefaultBuffer struct {
	checkpointStore CheckpointStore
	metadataStore   MetadataStore
	blobStore       payload.BlobStore
	config          *S3BufferConfig
	logger          *klog.Logger
	metrics         *statsd.Client
	ctx             context.Context
	actor           *pipeline.DefaultPipelineStageActor[*BufferInput, *BufferOutput]
	name            string
	tags            map[string]string
}

// NewAstraS3Buffer creates a default buffer that uses Astra DB for checkpointing and S3-compatible storage for payload persistence
func NewAstraS3Buffer(ctx context.Context, config *S3BufferConfig, astraConfig *AstraBundleConfig, metricTags map[string]string) *DefaultBuffer {
	logger := klog.FromContext(ctx)

	cqlStore := NewAstraCqlStore(logger, astraConfig)
	return &DefaultBuffer{
		checkpointStore: cqlStore,
		metadataStore:   cqlStore,
		blobStore:       payload.NewS3PayloadStore(ctx, logger, s3credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, ""), config.Endpoint, config.Region),
		config:          config,
		logger:          &logger,
		metrics:         ctx.Value(telemetry.MetricsClientContextKey).(*statsd.Client),
		ctx:             ctx,
		actor:           nil,
		name:            "default_astradb_s3",
		tags:            metricTags,
	}
}

// NewScyllaS3Buffer creates a default buffer that uses Astra DB for checkpointing and S3-compatible storage for payload persistence
func NewScyllaS3Buffer(ctx context.Context, config *S3BufferConfig, scyllaConfig *ScyllaCqlStoreConfig, metricTags map[string]string) *DefaultBuffer {
	logger := klog.FromContext(ctx)

	cqlStore := NewScyllaCqlStore(logger, scyllaConfig)
	return &DefaultBuffer{
		checkpointStore: cqlStore,
		metadataStore:   cqlStore,
		blobStore:       payload.NewS3PayloadStore(ctx, logger, s3credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, ""), config.Endpoint, config.Region),
		config:          config,
		logger:          &logger,
		metrics:         ctx.Value(telemetry.MetricsClientContextKey).(*statsd.Client),
		ctx:             ctx,
		actor:           nil,
		name:            "default_scylladb_s3",
		tags:            metricTags,
	}
}

func (buffer *DefaultBuffer) Start(submitter pipeline.StageActor[*BufferOutput, types.UID]) {
	buffer.actor = pipeline.NewDefaultPipelineStageActor[*BufferInput, *BufferOutput](
		buffer.name,
		buffer.tags,
		buffer.config.BufferConfig.FailureRateBaseDelay,
		buffer.config.BufferConfig.FailureRateMaxDelay,
		buffer.config.BufferConfig.RateLimitElementsPerSecond,
		buffer.config.BufferConfig.RateLimitElementsBurst,
		buffer.config.BufferConfig.Workers,
		buffer.bufferRequest,
		submitter,
	)

	buffer.actor.Start(buffer.ctx)
}

func (buffer *DefaultBuffer) Add(requestId string, algorithmName string, request *models.AlgorithmRequest, config *v1.NexusAlgorithmSpec, workgroup *v1.NexusAlgorithmWorkgroupSpec) error {
	input, err := NewBufferInput(requestId, algorithmName, request, config, workgroup)
	if err != nil {
		return err
	}

	buffer.actor.Receive(input)
	return nil
}

func (buffer *DefaultBuffer) Get(requestId string, algorithmName string) (*models.CheckpointedRequest, error) {
	return buffer.checkpointStore.ReadCheckpoint(algorithmName, requestId)
}

func (buffer *DefaultBuffer) GetBuffered(host string) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	return buffer.checkpointStore.ReadBufferedCheckpointsByHost(host)
}

func (buffer *DefaultBuffer) GetTagged(tag string) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	return buffer.checkpointStore.ReadCheckpointsByTag(tag)
}

func (buffer *DefaultBuffer) Update(checkpoint *models.CheckpointedRequest) error {
	return buffer.checkpointStore.UpsertCheckpoint(checkpoint)
}

func (buffer *DefaultBuffer) GetBufferedEntry(checkpoint *models.CheckpointedRequest) (*models.SubmissionBufferEntry, error) {
	return buffer.metadataStore.ReadMetadata(checkpoint)
}

func (buffer *DefaultBuffer) bufferRequest(input *BufferInput) (*BufferOutput, error) {
	telemetry.Increment(buffer.metrics, "incoming_requests", input.Tags())
	buffer.logger.V(0).Info("persisting payload", "request", input.Checkpoint.Id, "algorithm", input.Checkpoint.Algorithm)

	payloadPath := fmt.Sprintf("%s/%s/%s",
		buffer.config.BufferConfig.PayloadStoragePath,
		fmt.Sprintf("algorithm=%s", input.Checkpoint.Algorithm),
		input.Checkpoint.Id)

	if err := buffer.checkpointStore.UpsertCheckpoint(input.Checkpoint); err != nil {
		return nil, err
	}

	// TODO: add parent handling
	if err := buffer.blobStore.SaveTextAsBlob(buffer.ctx, string(*input.SerializedPayload), payloadPath); err != nil {
		return nil, err
	}

	bufferedCheckpoint := input.Checkpoint.DeepCopy()
	payloadValidity := *util.CoalescePointer(input.Checkpoint.PayloadValidityPeriod(), &buffer.config.BufferConfig.PayloadValidFor)
	payloadUri, err := buffer.blobStore.GetBlobUri(buffer.ctx, payloadPath, payloadValidity)
	if err != nil {
		return nil, err
	}

	bufferedCheckpoint.PayloadUri = payloadUri
	bufferedCheckpoint.PayloadValidFor = payloadValidity.String()
	bufferedCheckpoint.LifecycleStage = models.LifecycleStageBuffered
	bufferedEntry := models.FromCheckpoint(bufferedCheckpoint, input.ResolvedWorkgroup)

	if err := buffer.metadataStore.UpsertMetadata(bufferedEntry); err != nil {
		return nil, err
	}

	if err := buffer.checkpointStore.UpsertCheckpoint(bufferedCheckpoint); err != nil {
		return nil, err
	}

	telemetry.Increment(buffer.metrics, "checkpoint_requests", input.Tags())

	return &BufferOutput{
		Checkpoint: bufferedCheckpoint,
		Entry:      bufferedEntry,
		Workgroup:  input.ResolvedWorkgroup,
	}, nil
}
