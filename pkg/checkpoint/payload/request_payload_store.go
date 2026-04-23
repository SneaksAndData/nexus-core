package payload

import (
	"context"
	"time"
)

// RequestPayloadStore defines behaviours for a service responsible for serving client payloads to algorithms
type RequestPayloadStore interface {
	Persist(ctx context.Context, text string, blobPath string) error
	GenerateUrl(ctx context.Context, blobPath string, validFor time.Duration) (string, error)
}
