package request

import (
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/pipeline"
	"iter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

type Buffer interface {
	Get(requestId string, algorithmName string) (*models.CheckpointedRequest, error)
	GetBuffered(host string) (iter.Seq2[*models.CheckpointedRequest, error], error)
	GetNew(host string) (iter.Seq2[*models.CheckpointedRequest, error], error)
	GetTagged(tag string) (iter.Seq2[*models.CheckpointedRequest, error], error)
	Update(checkpoint *models.CheckpointedRequest) error
	GetBufferedEntry(checkpoint *models.CheckpointedRequest) (*models.SubmissionBufferEntry, error)
	Add(requestId string, algorithmName string, request *models.AlgorithmRequest, config *v1.NexusAlgorithmSpec, workgroup *v1.NexusAlgorithmWorkgroupSpec, parent *metav1.OwnerReference, isDryRun bool) error
	Start(submitter pipeline.StageActor[*BufferOutput, types.UID])
}

type BufferConfig struct {
	PayloadStoragePath         string        `mapstructure:"payload-storage-path,omitempty"`
	PayloadValidFor            time.Duration `mapstructure:"payload-valid-for,omitempty"`
	FailureRateBaseDelay       time.Duration `mapstructure:"failure-rate-base-delay,omitempty"`
	FailureRateMaxDelay        time.Duration `mapstructure:"failure-rate-max-delay,omitempty"`
	RateLimitElementsPerSecond int           `mapstructure:"rate-limit-elements-per-second,omitempty"`
	RateLimitElementsBurst     int           `mapstructure:"rate-limit-elements-burst,omitempty"`
	Workers                    int           `mapstructure:"workers,omitempty"`
}

type BufferInput struct {
	Checkpoint        *models.CheckpointedRequest
	ResolvedWorkgroup *v1.NexusAlgorithmWorkgroupSpec
	ResolvedParent    *metav1.OwnerReference
	SerializedPayload *[]byte
	Config            *v1.NexusAlgorithmSpec
	IsDryRun          bool
}

type BufferOutput struct {
	Checkpoint      *models.CheckpointedRequest
	Entry           *models.SubmissionBufferEntry
	Workgroup       *v1.NexusAlgorithmWorkgroupSpec
	ParentReference *metav1.OwnerReference
	IsDryRun        bool
}

func (input *BufferInput) Tags() map[string]string {
	return map[string]string{
		"algorithm": input.Checkpoint.Algorithm,
	}
}

func NewBufferInput(requestId string, algorithmName string, request *models.AlgorithmRequest, config *v1.NexusAlgorithmSpec, workgroup *v1.NexusAlgorithmWorkgroupSpec, parent *metav1.OwnerReference, isDryRun bool) (*BufferInput, error) {
	checkpoint, serializedPayload, err := models.FromAlgorithmRequest(requestId, algorithmName, request, config)

	if err != nil {
		return nil, err
	}

	return &BufferInput{
		Checkpoint:        checkpoint,
		ResolvedWorkgroup: workgroup,
		ResolvedParent:    parent,
		SerializedPayload: &serializedPayload,
		Config:            config,
		IsDryRun:          isDryRun,
	}, nil
}
