package request

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/payload"
	"k8s.io/klog/v2"
	"path"
	"time"
)

type Buffer interface {
	BufferRequest(ctx context.Context, requestId string, request *models.AlgorithmRequest, config *v1.MachineLearningAlgorithmSpec) (*models.SubmissionBufferEntry, error)
}

type PayloadStorageConfig struct {
	PayloadStoragePath string
	PayloadValidFor    time.Duration
}

type DefaultBuffer struct {
	checkpointStore CheckpointStore
	metadataStore   MetadataStore
	blobStore       payload.BlobStore
	payloadConfig   PayloadStorageConfig
	logger          klog.Logger
}

func (buffer *DefaultBuffer) BufferRequest(ctx context.Context, requestId string, request *models.AlgorithmRequest, config *v1.MachineLearningAlgorithmSpec) (*models.SubmissionBufferEntry, error) {
	checkpoint := models.FromAlgorithmRequest(requestId, request, config)
	// TODO: add metric
	//metricService.Increment(DeclaredMetrics.INCOMING_REQUESTS,
	//	new SortedDictionary<string, string>
	//{ { nameof(CheckpointedRequest.Algorithm), chk.Algorithm } });
	err := buffer.checkpointStore.UpsertCheckpoint(checkpoint)
	if err != nil {
		// TODO: log properly
		return nil, err
	}
	// TODO: add parent handling
	payloadPath := path.Join(
		buffer.payloadConfig.PayloadStoragePath,
		fmt.Sprintf("algorithm=%s", request.AlgorithmName),
		requestId)
	serializedPayload, err := json.Marshal(request.AlgorithmParameters)
	err = buffer.blobStore.SaveTextAsBlob(
		ctx,
		string(serializedPayload),
		payloadPath)
	if err != nil {
		// TODO: log properly
		return nil, err
	}
	checkpoint.PayloadUri, err = buffer.blobStore.GetBlobUri(ctx, payloadPath, buffer.payloadConfig.PayloadValidFor)
	if err != nil {
		// TODO: log properly
		return nil, err
	}
	checkpoint.LifecycleStage = models.LifecyclestageBuffered
	bufferedEntry := models.FromCheckpoint(checkpoint)
	err = buffer.metadataStore.UpsertMetadata(bufferedEntry)

	if err != nil {
		// TODO: log properly
		return nil, err
	}

	return bufferedEntry, nil
}
