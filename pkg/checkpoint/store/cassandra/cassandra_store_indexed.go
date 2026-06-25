package cassandra

import (
	"context"
	"encoding/base64"
	"fmt"
	"iter"
	"time"

	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/payload"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/store"
	"github.com/SneaksAndData/nexus-core/pkg/urlsign"
	"github.com/SneaksAndData/nexus-core/pkg/util"
)

type IndexedCassandraStore struct {
	cassandraStore            *CheckpointCassandraStore
	payloadProxyConfiguration *payload.RequestPayloadProxyConfiguration
}

func NewIndexedCassandraStore(store *CheckpointCassandraStore, payloadProxyConfiguration *payload.RequestPayloadProxyConfiguration) store.CheckpointStore {
	return &IndexedCassandraStore{
		cassandraStore:            store,
		payloadProxyConfiguration: payloadProxyConfiguration,
	}
}

func (ics *IndexedCassandraStore) UpsertCheckpoint(checkpoint *models.CheckpointedRequest) error {
	cloned := checkpoint.DeepCopy()

	cloned.LastModified = time.Now()
	if serialized, err := ToCassandraModel(cloned); err == nil {
		ics.cassandraStore.logger.V(1).Info("upserting checkpoint", "checkpoint", serialized)

		var query = ics.cassandraStore.cqlSession.Query(CheckpointedRequestTable.Insert()).BindStruct(*serialized).Strict()
		ics.cassandraStore.logger.V(1).Info("executing query", "query", query.String())

		if err := query.ExecRelease(); err != nil { // coverage-ignore
			ics.cassandraStore.logger.V(0).Error(err, "error when inserting a checkpoint", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
			return err
		}

		return nil
	} else {
		ics.cassandraStore.logger.V(0).Error(err, "error when preparing an insert of a checkpoint", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
		return err
	}
}

func (ics *IndexedCassandraStore) ReadCheckpoint(algorithm string, id string) (*models.CheckpointedRequest, error) {
	result := &CheckpointCassandraModel{
		Algorithm: algorithm,
		Id:        id,
	}

	var query = ics.cassandraStore.cqlSession.Query(CheckpointedRequestTable.Get()).BindStruct(*result)
	if err := query.GetRelease(result); err != nil { // coverage-ignore
		ics.cassandraStore.logger.V(1).Error(err, "error when reading a checkpoint", "algorithm", algorithm, "id", id)
		return nil, err
	}

	return result.FromCassandraModel()
}

func (ics *IndexedCassandraStore) ReadCheckpointsByHost(host string, lifecycleStage models.LifecycleStage) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	predicate := &CheckpointCassandraModel{
		ReceivedByHost: host,
		LifecycleStage: string(lifecycleStage),
	}
	queryResult := []*CheckpointCassandraModel{}

	var query = ics.cassandraStore.cqlSession.Query(CheckpointedRequestTableIndexByHost.GetBuilder().AllowFiltering().ToCql()).BindStruct(*predicate)
	if err := query.SelectRelease(&queryResult); err != nil { // coverage-ignore
		ics.cassandraStore.logger.V(1).Error(err, "error when reading buffered checkpoints", "host", host)
		return nil, err
	}

	return func(yield func(*models.CheckpointedRequest, error) bool) {
		for _, model := range queryResult {
			if !yield(model.FromCassandraModel()) {
				return
			}
		}
	}, nil
}

func (ics *IndexedCassandraStore) ReadCheckpointsByTag(requestTag string) (iter.Seq2[*models.CheckpointedRequest, error], error) {
	predicate := &CheckpointCassandraModel{
		Tag: requestTag,
	}
	queryResult := []*CheckpointCassandraModel{}

	var query = ics.cassandraStore.cqlSession.Query(CheckpointedRequestTableIndexByTag.Get()).BindStruct(*predicate)
	if err := query.SelectRelease(&queryResult); err != nil { // coverage-ignore
		ics.cassandraStore.logger.V(1).Error(err, "error when reading checkpoints by a tag", "tag", requestTag)
		return nil, err
	}

	return func(yield func(*models.CheckpointedRequest, error) bool) {
		for _, model := range queryResult {
			if !yield(model.FromCassandraModel()) {
				return
			}
		}
	}, nil
}

func (ics *IndexedCassandraStore) UpsertMetadata(entry *models.SubmissionBufferEntry) error {
	var query = ics.cassandraStore.cqlSession.Query(models.SubmissionBufferTable.Insert()).BindStruct(*entry)
	if err := query.ExecRelease(); err != nil { // coverage-ignore
		ics.cassandraStore.logger.V(1).Error(err, "error when inserting buffered checkpoint metadata", "algorithm", entry.Algorithm, "id", entry.Id)
		return err
	}

	return nil
}

func (ics *IndexedCassandraStore) ReadMetadata(checkpoint *models.CheckpointedRequest) (*models.SubmissionBufferEntry, error) {
	result := &models.SubmissionBufferEntry{
		Algorithm: checkpoint.Algorithm,
		Id:        checkpoint.Id,
	}

	var query = ics.cassandraStore.cqlSession.Query(models.SubmissionBufferTable.Get()).BindStruct(*result)
	if err := query.GetRelease(result); err != nil { // coverage-ignore
		ics.cassandraStore.logger.V(1).Error(err, "error when reading a buffered checkpoint metadata", "algorithm", checkpoint.Algorithm, "id", checkpoint.Id)
		return nil, err
	}

	return result, nil
}

func (ics *IndexedCassandraStore) Persist(ctx context.Context, payload string, requestId string, templateName string) error {
	if ics.payloadProxyConfiguration == nil {
		return fmt.Errorf("no payload proxy configuration provided, unable to persist checkpoint payload for %s/%s", templateName, requestId)
	}

	compressedPayload, err := util.CompressString(payload)

	if err != nil {
		return err
	}

	insertValues := &struct {
		Algorithm      string `db:"algorithm"`
		Id             string `db:"id"`
		PayloadContent string `db:"payload_content"`
	}{
		Algorithm:      templateName,
		Id:             requestId,
		PayloadContent: base64.StdEncoding.EncodeToString(compressedPayload),
	}

	var insertQuery = PayloadBufferTable.InsertQueryContext(ctx, ics.cassandraStore.cqlSession).BindStruct(insertValues)

	if err := insertQuery.ExecRelease(); err != nil { // coverage-ignore
		ics.cassandraStore.logger.V(1).Error(err, "error when persisting payload to the checkpoint store", "algorithm", templateName, "id", requestId)
		return err
	}

	return nil
}

func (ics *IndexedCassandraStore) GenerateUrl(_ context.Context, checkpoint *models.CheckpointedRequest) (string, error) {
	if ics.payloadProxyConfiguration == nil {
		return "", fmt.Errorf("no payload proxy configuration provided, unable to generate a proxy url for %s/%s", checkpoint.Algorithm, checkpoint.Id)
	}
	// Nexus URL signer ignores hostname, thus use localhost for simplicity
	baseUrl := checkpoint.GetProxyUrlToSign("localhost", ics.payloadProxyConfiguration.ServePathTemplate)
	signed, err := urlsign.Sign(baseUrl, ics.payloadProxyConfiguration.TenantId, checkpoint.PayloadValidityPeriod(), checkpoint.ContentHash, ics.payloadProxyConfiguration.SignSecret)

	if err != nil {
		return "", err
	}

	return signed.Url.String(), nil
}
