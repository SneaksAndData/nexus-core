package request

import (
	"context"
	"iter"

	"github.com/DataDog/datadog-go/v5/statsd"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/payload"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/store"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/store/cassandra"
	"github.com/SneaksAndData/nexus-core/pkg/pipeline"
	"github.com/SneaksAndData/nexus-core/pkg/telemetry"
	"github.com/SneaksAndData/nexus-core/pkg/util"
	s3credentials "github.com/aws/aws-sdk-go-v2/credentials"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type S3BufferConfig struct {
	BufferConfig       *BufferConfig `mapstructure:"buffer-config,omitempty"`
	AccessKeyID        string        `mapstructure:"access-key-id,omitempty"`
	SecretAccessKey    string        `mapstructure:"secret-access-key,omitempty"`
	Region             string        `mapstructure:"region,omitempty"`
	Endpoint           string        `mapstructure:"endpoint,omitempty"`
	PayloadStoragePath string        `mapstructure:"payload-storage-path,omitempty"`
}

func (c *S3BufferConfig) GetStaticCredentialsProvider() s3credentials.StaticCredentialsProvider {
	return s3credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, "")
}

type DefaultBuffer struct {
	checkpointStore store.CheckpointStore
	blobStore       payload.RequestPayloadStore
	config          *S3BufferConfig
	logger          *klog.Logger
	metrics         *statsd.Client
	ctx             context.Context
	actor           *pipeline.DefaultPipelineStageActor[*BufferInput, *BufferOutput]
	name            string
	tags            map[string]string
}

// NewAstraS3Buffer creates a default buffer that uses Astra DB for checkpointing and S3-compatible storage for payload persistence
func NewAstraS3Buffer(ctx context.Context, config *S3BufferConfig, astraConfig *cassandra.AstraBundleConfig, metricTags map[string]string) *DefaultBuffer { // coverage-ignore
	logger := klog.FromContext(ctx)

	cqlStore := cassandra.NewAstraStore(logger, astraConfig)
	return &DefaultBuffer{
		checkpointStore: cqlStore,
		blobStore: payload.NewS3PayloadStore(
			ctx, logger,
			config.GetStaticCredentialsProvider(),
			config.Endpoint,
			config.Region,
			config.PayloadStoragePath),
		config:  config,
		logger:  &logger,
		metrics: ctx.Value(telemetry.MetricsClientContextKey).(*statsd.Client),
		ctx:     ctx,
		actor:   nil,
		name:    "default_astradb_s3",
		tags:    metricTags,
	}
}

// NewScyllaS3Buffer creates a default buffer that uses ScyllaDb for checkpointing and S3-compatible storage for payload persistence
func NewScyllaS3Buffer(ctx context.Context, config *S3BufferConfig, scyllaConfig *cassandra.ScyllaConfig, metricTags map[string]string) *DefaultBuffer {
	logger := klog.FromContext(ctx)

	cqlStore := cassandra.NewScyllaStore(logger, scyllaConfig)
	return &DefaultBuffer{
		checkpointStore: cqlStore,
		blobStore:       payload.NewS3PayloadStore(ctx, logger, config.GetStaticCredentialsProvider(), config.Endpoint, config.Region, config.PayloadStoragePath),
		config:          config,
		logger:          &logger,
		metrics:         telemetry.GetClient(ctx),
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

	buffer.actor.Start(buffer.ctx, nil)
}

func (buffer *DefaultBuffer) Add(requestId string, algorithmName string, request *models.AlgorithmRequest, config *v1.NexusAlgorithmSpec, workgroup *v1.NexusAlgorithmWorkgroupSpec, parent *metav1.OwnerReference, isDryRun bool) error {
	input, err := NewBufferInput(requestId, algorithmName, request, config, workgroup, parent, isDryRun)
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
	return buffer.checkpointStore.ReadCheckpointsByHost(host, models.LifecycleStageBuffered)
}

func (buffer *DefaultBuffer) GetNew(host string) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	return buffer.checkpointStore.ReadCheckpointsByHost(host, models.LifecycleStageNew)
}

func (buffer *DefaultBuffer) GetTagged(tag string) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	return buffer.checkpointStore.ReadCheckpointsByTag(tag)
}

func (buffer *DefaultBuffer) Update(checkpoint *models.CheckpointedRequest) error {
	return buffer.checkpointStore.UpsertCheckpoint(checkpoint)
}

func (buffer *DefaultBuffer) GetBufferedEntry(checkpoint *models.CheckpointedRequest) (*models.SubmissionBufferEntry, error) {
	return buffer.checkpointStore.ReadMetadata(checkpoint)
}

func (buffer *DefaultBuffer) bufferRequest(input *BufferInput) (*BufferOutput, error) {
	telemetry.Increment(buffer.metrics, "incoming_requests", input.Tags())
	buffer.logger.V(0).Info("persisting payload", "request", input.Checkpoint.Id, "algorithm", input.Checkpoint.Algorithm)

	if err := buffer.checkpointStore.UpsertCheckpoint(input.Checkpoint); err != nil {
		return nil, err
	}

	if err := buffer.blobStore.Persist(buffer.ctx, string(*input.SerializedPayload), input.Checkpoint.Id, input.Checkpoint.Algorithm); err != nil {
		return nil, err
	}

	bufferedCheckpoint := input.Checkpoint.DeepCopy()
	payloadValidity := *util.CoalescePointer(input.Checkpoint.PayloadValidityPeriod(), &buffer.config.BufferConfig.PayloadValidFor)
	payloadUri, err := buffer.blobStore.GenerateUrl(buffer.ctx, input.Checkpoint.Id, input.Checkpoint.Algorithm, payloadValidity)
	if err != nil {
		return nil, err
	}

	bufferedCheckpoint.PayloadUri = payloadUri
	bufferedCheckpoint.PayloadValidFor = payloadValidity.String()
	bufferedCheckpoint.LifecycleStage = models.LifecycleStageBuffered
	bufferedEntry := models.FromCheckpoint(bufferedCheckpoint, input.ResolvedWorkgroup, input.ResolvedParent)

	if err := buffer.checkpointStore.UpsertMetadata(bufferedEntry); err != nil {
		return nil, err
	}

	if err := buffer.checkpointStore.UpsertCheckpoint(bufferedCheckpoint); err != nil {
		return nil, err
	}

	telemetry.Increment(buffer.metrics, "checkpoint_requests", input.Tags())

	return &BufferOutput{
		Checkpoint:      bufferedCheckpoint,
		Entry:           bufferedEntry,
		Workgroup:       input.ResolvedWorkgroup,
		ParentReference: input.ResolvedParent,
		IsDryRun:        input.IsDryRun,
	}, nil
}
