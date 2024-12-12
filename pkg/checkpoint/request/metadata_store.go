package request

type MetadataStore interface {
	UpsertMetadata(key string, value []byte) error
}
