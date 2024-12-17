package models

import (
	"encoding/json"
	"github.com/scylladb/gocqlx/v3/table"
	batchv1 "k8s.io/api/batch/v1"
)

var SubmissionBufferTable = table.New(table.Metadata{
	Name: "submission_buffer",
	Columns: []string{
		"algorithm",
		"id",
		"template",
	},
	PartKey: []string{
		"algorithm",
		"id",
	},
	SortKey: []string{},
})

type SubmissionBufferEntry struct {
	Algorithm string `json:"algorithm"`
	Id        string `json:"id"`
	Template  string `json:"template,omitempty"`
}

// SubmissionTemplate returns a Kubernetes Job object generated for the algorithm request
func (sbe *SubmissionBufferEntry) SubmissionTemplate() (*batchv1.Job, error) {
	var result *batchv1.Job
	err := json.Unmarshal([]byte(sbe.Template), result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func FromCheckpoint(checkpoint *CheckpointedRequest) *SubmissionBufferEntry {
	return &SubmissionBufferEntry{
		Algorithm: checkpoint.Algorithm,
		Id:        checkpoint.Id,
		Template:  checkpoint.ToV1Job(),
	}
}
