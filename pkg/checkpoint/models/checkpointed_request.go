package models

import "time"

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
	Algorithm               string                 `json:"algorithm"`
	Id                      string                 `json:"id"`
	LifecycleStage          string                 `json:"lifecycle_stage"`
	PayloadUri              string                 `json:"payload_uri"`
	ResultUri               string                 `json:"result_uri"`
	AlgorithmFailureCode    string                 `json:"algorithm_failure_code"`
	AlgorithmFailureCause   string                 `json:"algorithm_failure_cause"`
	AlgorithmFailureDetails string                 `json:"algorithm_failure_details"`
	ReceivedByHost          string                 `json:"received_by_host"`
	ReceivedAt              time.Time              `json:"received_at"`
	SentAt                  time.Time              `json:"sent_at"`
	AppliedConfiguration    AlgorithmConfiguration `json:"applied_configuration"`
	ConfigurationOverrides  AlgorithmConfiguration `json:"configuration_overrides"`
	MonitoringMetadata      map[string][]string    `json:"monitoring_metadata"`
	ContentHash             string                 `json:"content_hash"`
	LastModified            time.Time              `json:"last_modified"`
	Tag                     string                 `json:"tag"`
	ApiVersion              string                 `json:"api_version"`
	JobUid                  string                 `json:"job_uid"`
	ParentJob               ParentJobReference     `json:"parent_job"`
}
