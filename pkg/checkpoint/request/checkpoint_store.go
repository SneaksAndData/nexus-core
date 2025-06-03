package request

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"iter"
	"time"
)

type CheckpointStore interface {
	UpsertCheckpoint(checkpoint *models.CheckpointedRequest) error
	ReadCheckpoint(algorithm string, id string) (*models.CheckpointedRequest, error)
	ReadBufferedCheckpointsByHost(host string) (iter.Seq[*models.CheckpointedRequest], error)
	ReadCheckpointsByTag(requestTag string) (iter.Seq[*models.CheckpointedRequest], error)
}

func (cqls *CqlStore) UpsertCheckpoint(checkpoint *models.CheckpointedRequest) error {
	cloned := checkpoint.DeepCopy()

	cloned.LastModified = time.Now()
	if serialized, err := cloned.ToCqlModel(); err == nil {
		cqls.logger.V(0).Info("upserting checkpoint", "checkpoint", serialized)

		var query = cqls.cqlSession.Query(models.CheckpointedRequestTable.Insert()).BindStruct(*serialized).Strict()
		cqls.logger.V(0).Info("executing query", "query", query.String())

		if err := query.ExecRelease(); err != nil {
			cqls.logger.V(1).Error(err, "error when inserting a checkpoint", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
			return err
		}

		return nil
	} else {
		cqls.logger.V(0).Error(err, "error when preparing an insert of a checkpoint", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
		return err
	}
}

func (cqls *CqlStore) ReadCheckpoint(algorithm string, id string) (*models.CheckpointedRequest, error) {
	result := &models.CheckpointedRequestCqlModel{
		Algorithm: algorithm,
		Id:        id,
	}

	var query = cqls.cqlSession.Query(models.CheckpointedRequestTable.Get()).BindStruct(*result)
	if err := query.GetRelease(result); err != nil {
		cqls.logger.V(1).Error(err, "error when reading a checkpoint", "algorithm", algorithm, "id", id)
		return nil, err
	}

	return result.FromCqlModel()
}

func (cqls *CqlStore) ReadBufferedCheckpointsByHost(host string) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	predicate := &models.CheckpointedRequestCqlModel{
		ReceivedByHost: host,
		LifecycleStage: models.LifecycleStageBuffered,
	}
	queryResult := []*models.CheckpointedRequestCqlModel{}

	var query = cqls.cqlSession.Query(models.CheckpointedRequestTableIndexByHost.Get()).BindStruct(*predicate)
	if err := query.SelectRelease(&queryResult); err != nil {
		cqls.logger.V(1).Error(err, "error when reading buffered checkpoints", "host", host)
		return nil, err
	}

	return func(yield func(*models.CheckpointedRequest, error) bool) {
		for _, model := range queryResult {
			if !yield(model.FromCqlModel()) {
				return
			}
		}
	}, nil
}

func (cqls *CqlStore) ReadCheckpointsByTag(requestTag string) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	predicate := &models.CheckpointedRequestCqlModel{
		Tag: requestTag,
	}
	queryResult := []*models.CheckpointedRequestCqlModel{}

	var query = cqls.cqlSession.Query(models.CheckpointedRequestTableIndexByTag.Get()).BindStruct(*predicate)
	if err := query.SelectRelease(&queryResult); err != nil {
		cqls.logger.V(1).Error(err, "error when reading checkpoints by a tag", "tag", requestTag)
		return nil, err
	}

	return func(yield func(*models.CheckpointedRequest, error) bool) {
		for _, model := range queryResult {
			if !yield(model.FromCqlModel()) {
				return
			}
		}
	}, nil
}
