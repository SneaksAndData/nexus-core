package cassandra

import (
	"context"
	"encoding/base64"

	"github.com/SneaksAndData/nexus-core/pkg/util"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
	"k8s.io/klog/v2"
)

type CheckpointCassandraStore struct {
	cluster    *gocql.ClusterConfig
	cqlSession gocqlx.Session
	logger     klog.Logger
}

// NewCassandraStore creates a generic connected CheckpointCassandraStore (Apache Cassandra/Scylla)
func NewCassandraStore(cluster *gocql.ClusterConfig, logger klog.Logger) *CheckpointCassandraStore {
	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil { // coverage-ignore
		logger.V(0).Error(err, "failed to create CQL session")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	return &CheckpointCassandraStore{
		cluster:    cluster,
		cqlSession: session,
		logger:     logger,
	}
}

func (s *CheckpointCassandraStore) SavePayload(ctx context.Context, payload string, requestId string, templateName string) error {
	compressedPayload, err := util.CompressString(payload)

	if err != nil { // coverage-ignore
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

	var insertQuery = PayloadBufferTable.InsertQueryContext(ctx, s.cqlSession).BindStruct(insertValues)

	if err := insertQuery.ExecRelease(); err != nil { // coverage-ignore
		s.logger.V(1).Error(err, "error when persisting payload to the checkpoint store", "algorithm", templateName, "id", requestId)
		return err
	}

	return nil
}
