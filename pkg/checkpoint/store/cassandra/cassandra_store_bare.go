package cassandra

import (
	"iter"
	"time"

	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/gocql/gocql"
)

type BareCassandraStore struct {
	cassandraStore *CheckpointCassandraStore
}

func (bcs *BareCassandraStore) UpsertCheckpoint(checkpoint *models.CheckpointedRequest) error {
	cloned := checkpoint.DeepCopy()
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

		// 4 - to be added
		// ...

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
	//TODO implement me
	panic("implement me")
}

func (bcs *BareCassandraStore) ReadCheckpointsByHost(host string, lifecycleStage models.LifecycleStage) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	//TODO implement me
	panic("implement me")
}

func (bcs *BareCassandraStore) ReadCheckpointsByTag(requestTag string) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	//TODO implement me
	panic("implement me")
}

func (bcs *BareCassandraStore) UpsertMetadata(entry *models.SubmissionBufferEntry) error {
	//TODO implement me
	panic("implement me")
}

func (bcs *BareCassandraStore) ReadMetadata(checkpoint *models.CheckpointedRequest) (*models.SubmissionBufferEntry, error) {
	//TODO implement me
	panic("implement me")
}
