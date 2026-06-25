package payload

import (
	"context"

	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
)

// RequestPayloadStore defines behaviours for a service responsible for serving client payloads to algorithms
type RequestPayloadStore interface {
	Persist(ctx context.Context, payload string, requestId string, templateName string) error
	GenerateUrl(ctx context.Context, checkpoint *models.CheckpointedRequest) (string, error)
}
