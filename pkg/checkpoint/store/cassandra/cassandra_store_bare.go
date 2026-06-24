package cassandra

import (
	"iter"
	"time"

	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/store"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3/qb"
)

type BareCassandraStore struct {
	cassandraStore *CheckpointCassandraStore
}

func NewBareCassandraStore(store *CheckpointCassandraStore) store.CheckpointStore {
	return &BareCassandraStore{
		cassandraStore: store,
	}
}

func (bcs *BareCassandraStore) UpsertCheckpoint(checkpoint *models.CheckpointedRequest) error {
	cloned := checkpoint.DeepCopy()
	// atomically update all tables
	updateBatch := bcs.cassandraStore.cqlSession.NewBatch(gocql.LoggedBatch)

	cloned.LastModified = time.Now()
	if serialized, err := ToCassandraModel(cloned); err == nil {
		bcs.cassandraStore.logger.V(1).Info("upserting checkpoint", "checkpoint", serialized)

		// checkpoint upsert consists of:
		// 1 - actual checkpoint update/insert
		// 2 - update/insert into the by_hosts table
		// 3 - update/insert into the by_tag table
		// 4 (optional) - update/insert into the payload_buffer table

		// 1
		var upsertQuery = bcs.cassandraStore.cqlSession.Query(CheckpointedRequestTable.Insert()).Strict()

		// 2
		var upsertByHostQuery = bcs.cassandraStore.cqlSession.Query(CheckpointedRequestTableByHost.Insert()).Strict()

		// 3
		var upsertByTagQuery = bcs.cassandraStore.cqlSession.Query(CheckpointedRequestTableByTag.Insert()).Strict()

		// TODO: implement 4
		//var upsertPayloadQuery = bcs.cassandraStore.cqlSession.Query(CheckpointedRequestPayloadTable.Insert()).Strict()

		bcs.cassandraStore.logger.V(1).Info("adding main query to batch", "query", upsertQuery.String())

		if bindErr := updateBatch.BindStruct(upsertQuery, *serialized); bindErr != nil {
			bcs.cassandraStore.logger.V(0).Error(err, "error when preparing a checkpoint insert", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
			return bindErr
		}

		if bindErr := updateBatch.BindStruct(upsertByHostQuery, serialized.ByHostModel()); bindErr != nil {
			bcs.cassandraStore.logger.V(0).Error(err, "error when preparing a checkpoint insert", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
			return bindErr
		}

		if bindErr := updateBatch.BindStruct(upsertByTagQuery, serialized.ByTagModel()); bindErr != nil {
			bcs.cassandraStore.logger.V(0).Error(err, "error when preparing a checkpoint insert", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
			return bindErr
		}

		if execErr := bcs.cassandraStore.cqlSession.ExecuteBatch(updateBatch); execErr != nil {
			bcs.cassandraStore.logger.V(0).Error(err, "error when inserting a checkpoint", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
			return execErr
		}

		return nil
	} else {
		bcs.cassandraStore.logger.V(0).Error(err, "error when preparing an insert of a checkpoint", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
		return err
	}
}

func (bcs *BareCassandraStore) ReadCheckpoint(algorithm string, id string) (*models.CheckpointedRequest, error) {
	result := &CheckpointCassandraModel{
		Algorithm: algorithm,
		Id:        id,
	}

	var query = bcs.cassandraStore.cqlSession.Query(CheckpointedRequestTable.Get()).BindStruct(*result)
	if err := query.GetRelease(result); err != nil { // coverage-ignore
		bcs.cassandraStore.logger.V(1).Error(err, "error when reading a checkpoint", "algorithm", algorithm, "id", id)
		return nil, err
	}

	return result.FromCassandraModel()
}

func (bcs *BareCassandraStore) ReadCheckpointsByHost(host string, lifecycleStage models.LifecycleStage) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	byHostResults := &[]*struct {
		Host           string `db:"host"`
		LifecycleStage string `db:"lifecycle_stage"`
		Algorithm      string `db:"algorithm"`
		Id             string `db:"id"`
	}{}

	var byHostQuery = CheckpointedRequestTableByHost.SelectQuery(bcs.cassandraStore.cqlSession).BindMap(qb.M{
		"host":            host,
		"lifecycle_stage": lifecycleStage,
	})
	if err := byHostQuery.SelectRelease(byHostResults); err != nil { // coverage-ignore
		bcs.cassandraStore.logger.V(1).Error(err, "error when reading a checkpoint by host table", "host", host, "lifecycleStage", lifecycleStage)
		return nil, err
	}

	return func(yield func(*models.CheckpointedRequest, error) bool) {
		for _, byHostResult := range *byHostResults {
			if !yield(bcs.ReadCheckpoint(byHostResult.Algorithm, byHostResult.Id)) {
				return
			}
		}
	}, nil
}

func (bcs *BareCassandraStore) ReadCheckpointsByTag(requestTag string) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	byTagResults := &[]*models.CheckpointedRequest{}
	var byTagQuery = CheckpointedRequestTableByTag.SelectQuery(bcs.cassandraStore.cqlSession).BindMap(qb.M{
		"tag": requestTag,
	})
	if err := byTagQuery.SelectRelease(byTagResults); err != nil { // coverage-ignore
		bcs.cassandraStore.logger.V(1).Error(err, "error when reading a checkpoint by tag table", "tag", requestTag)
		return nil, err
	}

	return func(yield func(*models.CheckpointedRequest, error) bool) {
		for _, byTagResult := range *byTagResults {
			if !yield(byTagResult, nil) {
				return
			}
		}
	}, nil
}

func (bcs *BareCassandraStore) UpsertMetadata(entry *models.SubmissionBufferEntry) error {
	var query = bcs.cassandraStore.cqlSession.Query(models.SubmissionBufferTable.Insert()).BindStruct(*entry)
	if err := query.ExecRelease(); err != nil { // coverage-ignore
		bcs.cassandraStore.logger.V(1).Error(err, "error when inserting buffered checkpoint metadata", "algorithm", entry.Algorithm, "id", entry.Id)
		return err
	}

	return nil
}

func (bcs *BareCassandraStore) ReadMetadata(checkpoint *models.CheckpointedRequest) (*models.SubmissionBufferEntry, error) {
	result := &models.SubmissionBufferEntry{
		Algorithm: checkpoint.Algorithm,
		Id:        checkpoint.Id,
	}

	var query = bcs.cassandraStore.cqlSession.Query(models.SubmissionBufferTable.Get()).BindStruct(*result)
	if err := query.GetRelease(result); err != nil { // coverage-ignore
		bcs.cassandraStore.logger.V(1).Error(err, "error when reading a buffered checkpoint metadata", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
		return nil, err
	}

	return result, nil
}
