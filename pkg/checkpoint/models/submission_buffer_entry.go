package models

import (
	"encoding/json"
	"fmt"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/buildmeta"
	"github.com/scylladb/gocqlx/v3/table"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Cluster   string `json:"cluster"`
	Template  string `json:"template,omitempty"`
}

// SubmissionTemplate returns a Kubernetes Job object generated for the algorithm request
func (sbe *SubmissionBufferEntry) SubmissionTemplate() (*batchv1.Job, error) {
	result := &batchv1.Job{}
	err := json.Unmarshal([]byte(sbe.Template), result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func FromCheckpoint(checkpoint *CheckpointedRequest, resolvedWorkgroup *v1.NexusAlgorithmWorkgroupSpec, resolvedParent *metav1.OwnerReference) *SubmissionBufferEntry {
	jobTemplate, _ := json.Marshal(checkpoint.ToV1Job(fmt.Sprintf("%s-%s", buildmeta.AppVersion, buildmeta.BuildNumber), resolvedWorkgroup, resolvedParent))

	return &SubmissionBufferEntry{
		Algorithm: checkpoint.Algorithm,
		Id:        checkpoint.Id,
		Cluster:   resolvedWorkgroup.Cluster,
		Template:  string(jobTemplate),
	}
}
