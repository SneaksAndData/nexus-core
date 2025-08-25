package request

import (
	"context"
	"github.com/DataDog/datadog-go/v5/statsd"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/pipeline"
	"github.com/SneaksAndData/nexus-core/pkg/telemetry"
	"iter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"time"
)

type MemoryPassthroughBuffer struct {
	Checkpoints     []*models.CheckpointedRequest
	BufferedEntries []*models.SubmissionBufferEntry
	logger          *klog.Logger
	metrics         *statsd.Client
	ctx             context.Context
	actor           *pipeline.DefaultPipelineStageActor[*BufferInput, *BufferOutput]
	name            string
	tags            map[string]string
}

// NewMemoryPassthroughBuffer creates a default buffer that does not persist payloads. This buffer persists ALL information in app memory and is ONLY intended to use in tests. DO NOT USE THIS IN PRODUCTION.
// Some methods in this buffer will behave different from Cassandra buffers, for example Update replaces the checkpoint instead of updating its properties.
func NewMemoryPassthroughBuffer(ctx context.Context, metricTags map[string]string) *MemoryPassthroughBuffer {
	logger := klog.FromContext(ctx)
	return &MemoryPassthroughBuffer{
		Checkpoints: []*models.CheckpointedRequest{},
		logger:      &logger,
		metrics:     telemetry.GetClient(ctx),
		ctx:         ctx,
		actor:       nil,
		name:        "default_memory_passthrough",
		tags:        metricTags,
	}
}

func (buffer *MemoryPassthroughBuffer) Get(requestId string, algorithmName string) (*models.CheckpointedRequest, error) {
	for _, checkpoint := range buffer.Checkpoints {
		if checkpoint.Id == requestId && checkpoint.Algorithm == algorithmName {
			return checkpoint, nil
		}
	}

	return nil, nil
}

func (buffer *MemoryPassthroughBuffer) GetBuffered(host string) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	return func(yield func(*models.CheckpointedRequest, error) bool) {
		for _, checkpoint := range buffer.Checkpoints {
			if checkpoint.ReceivedByHost == host {
				yield(checkpoint, nil)
			}
		}
	}, nil
}

func (buffer *MemoryPassthroughBuffer) GetTagged(tag string) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	return func(yield func(*models.CheckpointedRequest, error) bool) {
		for _, checkpoint := range buffer.Checkpoints {
			if checkpoint.Tag == tag {
				yield(checkpoint, nil)
			}
		}
	}, nil
}

func (buffer *MemoryPassthroughBuffer) Update(checkpoint *models.CheckpointedRequest) error {
	var checkpointToUpdate int
	for index, bufferedCheckpoint := range buffer.Checkpoints {
		if checkpoint.Id == bufferedCheckpoint.Id && checkpoint.Algorithm == bufferedCheckpoint.Algorithm {
			checkpointToUpdate = index
			break
		}
	}

	buffer.Checkpoints[checkpointToUpdate] = checkpoint

	return nil
}

func (buffer *MemoryPassthroughBuffer) GetBufferedEntry(checkpoint *models.CheckpointedRequest) (*models.SubmissionBufferEntry, error) {
	for _, entry := range buffer.BufferedEntries {
		if checkpoint.Id == entry.Id && checkpoint.Algorithm == entry.Algorithm {
			return entry, nil
		}
	}

	return nil, nil
}

func (buffer *MemoryPassthroughBuffer) Add(requestId string, algorithmName string, request *models.AlgorithmRequest, config *v1.NexusAlgorithmSpec, workgroup *v1.NexusAlgorithmWorkgroupSpec, parent *metav1.OwnerReference) error {
	input, err := NewBufferInput(requestId, algorithmName, request, config, workgroup, parent)
	if err != nil {
		return err
	}

	buffer.actor.Receive(input)
	return nil
}

func (buffer *MemoryPassthroughBuffer) bufferRequest(input *BufferInput) (*BufferOutput, error) {
	bufferedCheckpoint := input.Checkpoint.DeepCopy()
	bufferedCheckpoint.LifecycleStage = models.LifecycleStageBuffered
	bufferedCheckpoint.PayloadUri = "memory"
	bufferedCheckpoint.PayloadValidFor = "24h"
	entry := models.FromCheckpoint(bufferedCheckpoint, input.ResolvedWorkgroup, input.ResolvedParent)

	buffer.Checkpoints = append(buffer.Checkpoints, input.Checkpoint)
	buffer.BufferedEntries = append(buffer.BufferedEntries, entry)

	return &BufferOutput{
		Checkpoint: bufferedCheckpoint,
		Entry:      entry,
		Workgroup:  input.ResolvedWorkgroup,
	}, nil
}

func (buffer *MemoryPassthroughBuffer) Start(submitter pipeline.StageActor[*BufferOutput, types.UID]) {
	buffer.actor = pipeline.NewDefaultPipelineStageActor[*BufferInput, *BufferOutput](
		buffer.name,
		buffer.tags,
		time.Second,
		time.Second*2,
		10,
		10,
		2,
		buffer.bufferRequest,
		submitter,
	)

	buffer.actor.Start(buffer.ctx, nil)
}
