package payload

import (
	"context"
)

// RequestPayloadStore defines behaviours for a service responsible for serving client payloads to algorithms
type RequestPayloadStore interface {
	Persist(ctx context.Context, payload string, requestId string, templateName string) error
	Retrieve(ctx context.Context, requestId string, templateName string) ([]byte, error)
}
