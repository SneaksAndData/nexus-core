package payload

import (
	"context"
	"time"
)

// BlobStore is a generic blob store functionality required by Nexus
type BlobStore interface {
	SaveTextAsBlob(ctx context.Context, text string, blobPath string) error
	GetBlobUri(ctx context.Context, blobPath string, validFor time.Duration) (string, error)
}
