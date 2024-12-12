package request

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
	"k8s.io/klog/v2"
	"os"
)

type MetadataStore interface {
	UpsertMetadata(entry *models.SubmissionBufferEntry) error
	ReadMetadata(checkpoint *models.CheckpointedRequest) (*models.SubmissionBufferEntry, error)
}

type CqlStore struct {
	cluster    *gocql.ClusterConfig
	cqlSession gocqlx.Session
	logger     klog.Logger
}

func NewCqlStore(cluster *gocql.ClusterConfig, logger klog.Logger) *CqlStore {
	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		logger.V(0).Error(err, "failed to create CQL session")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	return &CqlStore{
		cluster:    cluster,
		cqlSession: session,
		logger:     logger,
	}
}

func NewAstraCqlStore(logger klog.Logger) *CqlStore {
	bundleBytes := os.Getenv("NEXUS__CQL_ASTRA_BUNDLE")
	if bundleBytes == "" {
		logger.V(0).Info("Astra bundle or credentials are not present in the environment")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	// unpack and bootstrap Astra connection
	return nil
}

func (cqls *CqlStore) UpsertMetadata(entry *models.SubmissionBufferEntry) error {
	return nil
}

func (cqls *CqlStore) ReadMetadata(checkpoint *models.CheckpointedRequest) (*models.SubmissionBufferEntry, error) {
	return nil, nil
}
