package request

import (
	models2 "github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
)

type MetadataStore interface {
	UpsertMetadata(entry *models2.SubmissionBufferEntry) error
	ReadMetadata(checkpoint *models2.CheckpointedRequest) (*models2.SubmissionBufferEntry, error)
}

type CqlStore struct {
	cluster    *gocql.ClusterConfig
	cqlSession *gocqlx.Session
}

func (cqls *CqlStore) UpsertMetadata(entry *models2.SubmissionBufferEntry) error {
	return nil
}

func (cqls *CqlStore) ReadMetadata(checkpoint *models2.CheckpointedRequest) (*models2.SubmissionBufferEntry, error) {
	return nil, nil
}
