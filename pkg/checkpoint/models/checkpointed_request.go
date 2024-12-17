package models

import (
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/scylladb/gocqlx/v3/table"
	batchv1 "k8s.io/api/batch/v1"
	"os"
	"time"
)

const (
	LifecyclestageNew              = "NEW"
	LifecyclestageBuffered         = "BUFFERED"
	LifecyclestageRunning          = "RUNNING"
	LifecyclestageCompleted        = "COMPLETED"
	LifecyclestageFailed           = "FAILED"
	LifecyclestageScheduleTimeout  = "SCHEDULING_TIMEOUT"
	LifecyclestageDeadlineExceeded = "DEADLINE_EXCEEDED"
	LifecyclestageCancelled        = "CANCELLED"
)

type CheckpointedRequest struct {
	Algorithm               string                          `json:"algorithm"`
	Id                      string                          `json:"id"`
	LifecycleStage          string                          `json:"lifecycle_stage"`
	PayloadUri              string                          `json:"payload_uri"`
	ResultUri               string                          `json:"result_uri"`
	AlgorithmFailureCode    string                          `json:"algorithm_failure_code"`
	AlgorithmFailureCause   string                          `json:"algorithm_failure_cause"`
	AlgorithmFailureDetails string                          `json:"algorithm_failure_details"`
	ReceivedByHost          string                          `json:"received_by_host"`
	ReceivedAt              time.Time                       `json:"received_at"`
	SentAt                  time.Time                       `json:"sent_at"`
	AppliedConfiguration    v1.MachineLearningAlgorithmSpec `json:"applied_configuration"`
	ConfigurationOverrides  v1.MachineLearningAlgorithmSpec `json:"configuration_overrides"`
	MonitoringMetadata      map[string][]string             `json:"monitoring_metadata"`
	ContentHash             string                          `json:"content_hash"`
	LastModified            time.Time                       `json:"last_modified"`
	Tag                     string                          `json:"tag"`
	ApiVersion              string                          `json:"api_version"`
	JobUid                  string                          `json:"job_uid"`
	ParentJob               ParentJobReference              `json:"parent_job"`
}

var CheckpointedRequestTable = table.New(table.Metadata{
	Name: "checkpoints",
	Columns: []string{
		"algorithm",
		"id",
		"lifecycle_stage",
		"payload_uri",
		"result_uri",
		"algorithm_failure_code",
		"algorithm_failure_cause",
		"algorithm_failure_details",
		"received_by_host",
		"received_at",
		"sent_at",
		"applied_configuration",
		"configuration_overrides",
		"monitoring_metadata",
		"content_hash",
		"last_modified",
		"tag",
		"api_version",
		"parent_job",
	},
	PartKey: []string{
		"algorithm",
		"id",
	},
	SortKey: []string{},
})

func FromAlgorithmRequest(requestId string, request *AlgorithmRequest, config *v1.MachineLearningAlgorithmSpec) *CheckpointedRequest {
	hostname, _ := os.Hostname()
	return &CheckpointedRequest{
		Algorithm:              request.AlgorithmName,
		Id:                     requestId,
		LifecycleStage:         LifecyclestageNew,
		ReceivedByHost:         hostname,
		ReceivedAt:             time.Now(),
		LastModified:           time.Now(),
		ConfigurationOverrides: request.CustomConfiguration,
		Tag:                    request.Tag,
		JobUid:                 "",
		ParentJob:              ParentJobReference{}, // TODO: add support for parent job
		MonitoringMetadata:     request.MonitoringMetadata,
		ApiVersion:             request.RequestApiVersion,
		AppliedConfiguration:   *config,
	}
}

// TODO: implement
func (req *CheckpointedRequest) ToV1Job() batchv1.Job {
	return batchv1.Job{}
}
