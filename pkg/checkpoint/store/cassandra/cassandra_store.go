package cassandra

import (
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
