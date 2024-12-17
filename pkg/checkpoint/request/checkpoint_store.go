package request

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/util"
	"github.com/scylladb/gocqlx/v3/qb"
	"time"
)

type CheckpointStore interface {
	UpsertCheckpoint(checkpoint *models.CheckpointedRequest) error
	ReadCheckpoint(algorithm string, id string) (*models.CheckpointedRequest, error)
	ReadCheckpoints(requestTag string) ([]models.CheckpointedRequest, error)
}

func (cqls *CqlStore) UpsertCheckpoint(checkpoint *models.CheckpointedRequest) error {
	cloned, err := util.DeepClone(*checkpoint)
	if err != nil {
		cqls.logger.V(1).Error(err, "Error when serializing a checkpoint", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
		return err
	}

	cloned.LastModified = time.Now()

	var query = cqls.cqlSession.Query(models.CheckpointedRequestTable.Insert()).BindStruct(*checkpoint)
	if err := query.ExecRelease(); err != nil {
		cqls.logger.V(1).Error(err, "Error when inserting a checkpoint", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
		return err
	}

	return nil
}

func (cqls *CqlStore) ReadCheckpoint(algorithm string, id string) (*models.CheckpointedRequest, error) {
	result := &models.CheckpointedRequest{
		Algorithm: algorithm,
		Id:        id,
	}

	var query = cqls.cqlSession.Query(models.CheckpointedRequestTable.Get()).BindStruct(*result)
	if err := query.GetRelease(result); err != nil {
		cqls.logger.V(1).Error(err, "Error when reading a checkpoint", "algorithm", algorithm, "id", id)
		return nil, err
	}

	return result, nil
}

func (cqls *CqlStore) ReadCheckpoints(requestTag string) ([]models.CheckpointedRequest, error) {
	var result []models.CheckpointedRequest
	query := cqls.cqlSession.Query(models.CheckpointedRequestTable.Select()).BindMap(qb.M{
		"tag": requestTag,
	})
	if err := query.SelectRelease(&result); err != nil {
		cqls.logger.V(1).Error(err, "Error when reading checkpoints by tag", "tag", requestTag)
		return nil, err
	}

	return result, nil
}
