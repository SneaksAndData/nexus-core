package cassandra

import (
	"iter"
	"time"

	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
)

func (ccs *CheckpointCassandraStore) UpsertCheckpoint(checkpoint *models.CheckpointedRequest) error {
	cloned := checkpoint.DeepCopy()

	cloned.LastModified = time.Now()
	if serialized, err := cloned.ToCqlModel(); err == nil {
		ccs.logger.V(0).Info("upserting checkpoint", "checkpoint", serialized)

		var query = ccs.cqlSession.Query(models.CheckpointedRequestTable.Insert()).BindStruct(*serialized).Strict()
		ccs.logger.V(0).Info("executing query", "query", query.String())

		if err := query.ExecRelease(); err != nil { // coverage-ignore
			ccs.logger.V(1).Error(err, "error when inserting a checkpoint", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
			return err
		}

		return nil
	} else {
		ccs.logger.V(0).Error(err, "error when preparing an insert of a checkpoint", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
		return err
	}
}

func (ccs *CheckpointCassandraStore) ReadCheckpoint(algorithm string, id string) (*models.CheckpointedRequest, error) {
	result := &models.CheckpointedRequestCqlModel{
		Algorithm: algorithm,
		Id:        id,
	}

	var query = ccs.cqlSession.Query(models.CheckpointedRequestTable.Get()).BindStruct(*result)
	if err := query.GetRelease(result); err != nil { // coverage-ignore
		ccs.logger.V(1).Error(err, "error when reading a checkpoint", "algorithm", algorithm, "id", id)
		return nil, err
	}

	return result.FromCqlModel()
}

func (ccs *CheckpointCassandraStore) ReadCheckpointsByHost(host string, lifecycleStage models.LifecycleStage) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	predicate := &models.CheckpointedRequestCqlModel{
		ReceivedByHost: host,
		LifecycleStage: string(lifecycleStage),
	}
	queryResult := []*models.CheckpointedRequestCqlModel{}

	var query = ccs.cqlSession.Query(models.CheckpointedRequestTableIndexByHost.GetBuilder().AllowFiltering().ToCql()).BindStruct(*predicate)
	if err := query.SelectRelease(&queryResult); err != nil { // coverage-ignore
		ccs.logger.V(1).Error(err, "error when reading buffered checkpoints", "host", host)
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

func (ccs *CheckpointCassandraStore) ReadCheckpointsByTag(requestTag string) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	predicate := &models.CheckpointedRequestCqlModel{
		Tag: requestTag,
	}
	queryResult := []*models.CheckpointedRequestCqlModel{}

	var query = ccs.cqlSession.Query(models.CheckpointedRequestTableIndexByTag.Get()).BindStruct(*predicate)
	if err := query.SelectRelease(&queryResult); err != nil { // coverage-ignore
		ccs.logger.V(1).Error(err, "error when reading checkpoints by a tag", "tag", requestTag)
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

func (ccs *CheckpointCassandraStore) UpsertMetadata(entry *models.SubmissionBufferEntry) error {
	var query = ccs.cqlSession.Query(models.SubmissionBufferTable.Insert()).BindStruct(*entry)
	if err := query.ExecRelease(); err != nil { // coverage-ignore
		ccs.logger.V(1).Error(err, "error when inserting buffered checkpoint metadata", "algorithm", entry.Algorithm, "id", entry.Id)
		return err
	}

	return nil
}

func (ccs *CheckpointCassandraStore) ReadMetadata(checkpoint *models.CheckpointedRequest) (*models.SubmissionBufferEntry, error) {
	result := &models.SubmissionBufferEntry{
		Algorithm: checkpoint.Algorithm,
		Id:        checkpoint.Id,
	}

	var query = ccs.cqlSession.Query(models.SubmissionBufferTable.Get()).BindStruct(*result)
	if err := query.GetRelease(result); err != nil { // coverage-ignore
		ccs.logger.V(1).Error(err, "error when reading a buffered checkpoint metadata", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
		return nil, err
	}

	return result, nil
}
