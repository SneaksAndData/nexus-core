package models

import (
	"encoding/json"
	"fmt"
	"github.com/SneaksAndData/nexus-core/pkg/buildmeta"
	"github.com/scylladb/gocqlx/v3/table"
	batchv1 "k8s.io/api/batch/v1"
)

var SubmissionBufferTable = table.New(table.Metadata{
	Name: "nexus.submission_buffer",
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
	// TODO: job node taint and taint value should be moved to CRD
	// this change will be implemented once receiver is functional, to avoid adding extra complexity to initial testing
	jobTemplate, _ := json.Marshal(checkpoint.ToV1Job("kubernetes.sneaksanddata.com/service-node-group", checkpoint.AppliedConfiguration.Workgroup, fmt.Sprintf("%s-%s", buildmeta.AppVersion, buildmeta.BuildNumber)))

	return &SubmissionBufferEntry{
		Algorithm: checkpoint.Algorithm,
		Id:        checkpoint.Id,
		Template:  string(jobTemplate),
	}
}
