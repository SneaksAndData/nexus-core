package request

import (
	"context"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"time"
)

type Buffer interface {
	BufferRequest(ctx context.Context, requestId string, request *models.AlgorithmRequest, config *v1.NexusAlgorithmSpec) (*models.SubmissionBufferEntry, error)
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
	SerializedPayload *[]byte
	Config            *v1.NexusAlgorithmSpec
}

type BufferOutput struct {
	Checkpoint *models.CheckpointedRequest
	Entry      *models.SubmissionBufferEntry
	Workgroup  *v1.NexusAlgorithmWorkgroupSpec
}

func (input *BufferInput) Tags() map[string]string {
	return map[string]string{
		"algorithm": input.Checkpoint.Algorithm,
	}
}

func NewBufferInput(requestId string, algorithmName string, request *models.AlgorithmRequest, config *v1.NexusAlgorithmSpec, workgroup *v1.NexusAlgorithmWorkgroupSpec) (*BufferInput, error) {
	checkpoint, serializedPayload, err := models.FromAlgorithmRequest(requestId, algorithmName, request, config)

	if err != nil {
		return nil, err
	}

	return &BufferInput{
		Checkpoint:        checkpoint,
		ResolvedWorkgroup: workgroup,
		SerializedPayload: &serializedPayload,
		Config:            config,
	}, nil
}
