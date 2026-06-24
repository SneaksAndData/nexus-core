package store

import (
	"iter"

	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
)

// CheckpointStore defines behaviours for checkpointing operations
type CheckpointStore interface {
	UpsertCheckpoint(checkpoint *models.CheckpointedRequest) error
	ReadCheckpoint(algorithm string, id string) (*models.CheckpointedRequest, error)
	ReadCheckpointsByHost(host string, lifecycleStage models.LifecycleStage) (iter.Seq2[*models.CheckpointedRequest, error], error)
	ReadCheckpointsByTag(requestTag string) (iter.Seq2[*models.CheckpointedRequest, error], error)
	UpsertMetadata(entry *models.SubmissionBufferEntry) error
	ReadMetadata(checkpoint *models.CheckpointedRequest) (*models.SubmissionBufferEntry, error)
}
