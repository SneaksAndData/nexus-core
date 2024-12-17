package payload

import "time"

type BlobStore interface {
	SaveTextAsBlob(text string) (string, error)
	GetBlobUri(blobPath string, validFor time.Duration) (string, error)
}
