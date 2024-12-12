package models

import (
	"encoding/json"
	"github.com/scylladb/gocqlx/v3/table"
	batchv1 "k8s.io/api/batch/v1"
)

var submissionBufferTable = table.New(table.Metadata{
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
	Algorithm string
	Id        string
	Template  string
}

func (sbe *SubmissionBufferEntry) submissionTemplate() (*batchv1.Job, error) {
	var result *batchv1.Job
	err := json.Unmarshal([]byte(sbe.Template), result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
