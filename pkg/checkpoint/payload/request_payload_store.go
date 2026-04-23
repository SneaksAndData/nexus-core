package payload

import (
	"context"
	"time"
)

// RequestPayloadStore defines behaviours for a service responsible for serving client payloads to algorithms
type RequestPayloadStore interface {
	Persist(ctx context.Context, payload string, requestId string, templateName string) error
	GenerateUrl(ctx context.Context, requestId string, templateName string, validFor time.Duration) (string, error)
}
