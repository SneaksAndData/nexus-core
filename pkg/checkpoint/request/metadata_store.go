package request

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
)

type MetadataStore interface {
	UpsertMetadata(entry *models.SubmissionBufferEntry) error
	ReadMetadata(checkpoint *models.CheckpointedRequest) (*models.SubmissionBufferEntry, error)
}

func (cqls *CqlStore) UpsertMetadata(entry *models.SubmissionBufferEntry) error {
	var query = cqls.cqlSession.Query(models.SubmissionBufferTable.Insert()).BindStruct(*entry)
	if err := query.ExecRelease(); err != nil { // coverage-ignore
		cqls.logger.V(1).Error(err, "error when inserting buffered checkpoint metadata", "algorithm", entry.Algorithm, "id", entry.Id)
		return err
	}

	return nil
}

func (cqls *CqlStore) ReadMetadata(checkpoint *models.CheckpointedRequest) (*models.SubmissionBufferEntry, error) {
	result := &models.SubmissionBufferEntry{
		Algorithm: checkpoint.Algorithm,
		Id:        checkpoint.Id,
	}

	var query = cqls.cqlSession.Query(models.SubmissionBufferTable.Get()).BindStruct(*result)
	if err := query.GetRelease(result); err != nil { // coverage-ignore
		cqls.logger.V(1).Error(err, "error when reading a buffered checkpoint metadata", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
		return nil, err
	}

	return result, nil
}
