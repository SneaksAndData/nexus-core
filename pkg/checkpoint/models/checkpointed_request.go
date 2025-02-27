package models

import (
	"encoding/json"
	"fmt"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/util"
	"github.com/aws/smithy-go/ptr"
	"github.com/scylladb/gocqlx/v3/table"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"
	"time"
)

type LifecycleStage string

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

type ClientErrorCode string

const (
	CB000  ClientErrorCode = "Scheduling timeout."                                            // client-facing code for buffer stage errors
	CAJ011 ClientErrorCode = "Execution timed out."                                           // client-facing code for deadline exceed due to running over time
	CAJ012 ClientErrorCode = "Job timed out waiting to be scheduled. Please try again later." // lost submissions and other scheduling errors that cause infinite hang in a buffering stage
	CAJ000 ClientErrorCode = "Execution cancelled by %s, reason: %s"                          // job has been gracefully cancelled via API
)

func (ce ClientErrorCode) ErrorName() string {
	switch ce {
	case CB000:
		return "CB000"
	case CAJ011:
		return "CAJ011"
	case CAJ012:
		return "CAJ012"
	case CAJ000:
		return "CAJ000"
	default:
		return "UNDEFINED"
	}
}

func (ce ClientErrorCode) ErrorMessage() string {
	return string(ce)
}

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

type CheckpointedRequestCqlModel struct {
	Algorithm               string
	Id                      string
	LifecycleStage          string
	PayloadUri              string
	ResultUri               string
	AlgorithmFailureCode    string
	AlgorithmFailureCause   string
	AlgorithmFailureDetails string
	ReceivedByHost          string
	ReceivedAt              time.Time
	SentAt                  time.Time
	AppliedConfiguration    string
	ConfigurationOverrides  string
	MonitoringMetadata      map[string][]string
	ContentHash             string
	LastModified            time.Time
	Tag                     string
	ApiVersion              string
	JobUid                  string
	ParentJob               string
}

var CheckpointedRequestTable = table.New(table.Metadata{
	Name: "nexus.checkpoints",
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

func (c *CheckpointedRequest) ToCqlModel() *CheckpointedRequestCqlModel {
	serializedConfig, _ := json.Marshal(c.AppliedConfiguration)
	serializedOverrides, _ := json.Marshal(c.ConfigurationOverrides)

	return &CheckpointedRequestCqlModel{
		Algorithm:               c.Algorithm,
		Id:                      c.Id,
		LifecycleStage:          c.LifecycleStage,
		PayloadUri:              c.PayloadUri,
		ResultUri:               c.ResultUri,
		AlgorithmFailureCode:    c.AlgorithmFailureCode,
		AlgorithmFailureCause:   c.AlgorithmFailureCause,
		AlgorithmFailureDetails: c.AlgorithmFailureDetails,
		ReceivedByHost:          c.ReceivedByHost,
		ReceivedAt:              c.ReceivedAt,
		SentAt:                  c.SentAt,
		AppliedConfiguration:    string(serializedConfig),
		ConfigurationOverrides:  string(serializedOverrides),
		MonitoringMetadata:      c.MonitoringMetadata,
		ContentHash:             c.ContentHash,
		LastModified:            c.LastModified,
		Tag:                     c.Tag,
		ApiVersion:              c.ApiVersion,
		JobUid:                  c.JobUid,
		ParentJob:               "", // TODO: fixme
	}
}

func (c *CheckpointedRequestCqlModel) FromCqlModel() *CheckpointedRequest {
	var appliedConfig v1.MachineLearningAlgorithmSpec
	var overrides v1.MachineLearningAlgorithmSpec
	_ = json.Unmarshal([]byte(c.AppliedConfiguration), &appliedConfig)
	_ = json.Unmarshal([]byte(c.ConfigurationOverrides), &overrides)

	return &CheckpointedRequest{
		Algorithm:               c.Algorithm,
		Id:                      c.Id,
		LifecycleStage:          c.LifecycleStage,
		PayloadUri:              c.PayloadUri,
		ResultUri:               c.ResultUri,
		AlgorithmFailureCode:    c.AlgorithmFailureCode,
		AlgorithmFailureCause:   c.AlgorithmFailureCause,
		AlgorithmFailureDetails: c.AlgorithmFailureDetails,
		ReceivedByHost:          c.ReceivedByHost,
		ReceivedAt:              c.ReceivedAt,
		SentAt:                  c.SentAt,
		AppliedConfiguration:    appliedConfig,
		ConfigurationOverrides:  overrides,
		MonitoringMetadata:      c.MonitoringMetadata,
		ContentHash:             c.ContentHash,
		LastModified:            c.LastModified,
		Tag:                     c.Tag,
		ApiVersion:              c.ApiVersion,
		JobUid:                  c.JobUid,
		ParentJob:               ParentJobReference{},
	}
}

func FromAlgorithmRequest(requestId string, algorithmName string, request *AlgorithmRequest, config *v1.MachineLearningAlgorithmSpec) (*CheckpointedRequest, []byte, error) {
	hostname, _ := os.Hostname()
	serializedPayload, err := json.Marshal(request.AlgorithmParameters)

	if err != nil {
		return nil, nil, err
	}

	return &CheckpointedRequest{
		Algorithm:              algorithmName,
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
	}, serializedPayload, nil
}

func (c *CheckpointedRequest) DeepCopy() *CheckpointedRequest {
	return &CheckpointedRequest{
		Algorithm:               c.Algorithm,
		Id:                      c.Id,
		LifecycleStage:          c.LifecycleStage,
		PayloadUri:              c.PayloadUri,
		ResultUri:               c.ResultUri,
		AlgorithmFailureCode:    c.AlgorithmFailureCode,
		AlgorithmFailureCause:   c.AlgorithmFailureCause,
		AlgorithmFailureDetails: c.AlgorithmFailureDetails,
		ReceivedByHost:          c.ReceivedByHost,
		ReceivedAt:              c.ReceivedAt,
		SentAt:                  c.SentAt,
		AppliedConfiguration:    *c.AppliedConfiguration.DeepCopy(),
		ConfigurationOverrides:  *c.ConfigurationOverrides.DeepCopy(),
		MonitoringMetadata:      c.MonitoringMetadata,
		ContentHash:             c.ContentHash,
		LastModified:            c.LastModified,
		Tag:                     c.Tag,
		ApiVersion:              c.ApiVersion,
		JobUid:                  c.JobUid,
		ParentJob:               ParentJobReference{},
	}
}

func (c *CheckpointedRequest) ToV1Job() batchv1.Job {
	appVersion := util.CoalesceString(os.Getenv("APPLICATION_VERSION"), "0.0.0")
	jobNodeTaint := util.CoalesceString(os.Getenv("NEXUS__JOB_NODE_CLASS"), "kubernetes.sneaksanddata.io/node-group")
	jobNodeTaintValue := util.CoalesceString(c.AppliedConfiguration.Workgroup, os.Getenv("NEXUS__JOB_NODE_CLASS_VALUE"))

	jobResourceList := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(c.AppliedConfiguration.CpuLimit),
		corev1.ResourceMemory: resource.MustParse(c.AppliedConfiguration.MemoryLimit),
	}

	for customResourceKey, customResourceValue := range c.AppliedConfiguration.CustomResources {
		jobResourceList[corev1.ResourceName(customResourceKey)] = resource.MustParse(customResourceValue)
	}

	jobArgs := []string{}

	for _, argValue := range c.AppliedConfiguration.Args {
		if strings.Contains(argValue, "sas-uri") {
			jobArgs = append(jobArgs, fmt.Sprintf(argValue, c.PayloadUri))
		} else if strings.Contains(argValue, "request-id") {
			jobArgs = append(jobArgs, fmt.Sprintf(argValue, c.Id))
		} else {
			jobArgs = append(jobArgs, argValue)
		}
	}

	jobAnnotations := map[string]string{
		"science.sneaksanddata.com/algorithm-template-name": c.Algorithm,
	}

	jobTolerations := []corev1.Toleration{
		{
			Operator: corev1.TolerationOpEqual,
			Key:      jobNodeTaint,
			Value:    jobNodeTaintValue,
		},
	}

	jobVolumes := []corev1.Volume{}
	jobVolumeMounts := []corev1.VolumeMount{}
	jobPodFailurePolicy := &batchv1.PodFailurePolicy{
		Rules: []batchv1.PodFailurePolicyRule{
			{
				Action:      batchv1.PodFailurePolicyActionIgnore,
				OnExitCodes: nil,
				OnPodConditions: []batchv1.PodFailurePolicyOnPodConditionsPattern{
					{
						Type:   corev1.DisruptionTarget,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
	}

	if *c.AppliedConfiguration.MountDatadogSocket {
		jobVolumes = append(jobVolumes, corev1.Volume{
			Name: "dsdsocket",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/var/run/datadog/",
				},
			},
		})
		jobVolumeMounts = append(jobVolumeMounts, corev1.VolumeMount{
			Name:      "dsdsocket",
			MountPath: "/var/run/datadog",
		})
	}

	if c.MonitoringMetadata != nil {
		for annotationKey, annotationValue := range c.MonitoringMetadata {
			jobAnnotations[annotationKey] = strings.Join(annotationValue, ",")
		}
	}

	if c.AppliedConfiguration.AdditionalWorkgroups != nil {
		for adwKey, adwValue := range c.AppliedConfiguration.AdditionalWorkgroups {
			jobTolerations = append(jobTolerations, corev1.Toleration{
				Operator: corev1.TolerationOpEqual,
				Key:      adwKey,
				Value:    adwValue,
			})
		}
	}

	if c.AppliedConfiguration.FatalExitCodes != nil {
		jobPodFailurePolicy.Rules = append(jobPodFailurePolicy.Rules, batchv1.PodFailurePolicyRule{
			Action: batchv1.PodFailurePolicyActionFailJob,
			OnExitCodes: &batchv1.PodFailurePolicyOnExitCodesRequirement{
				Operator: batchv1.PodFailurePolicyOnExitCodesOpIn,
				Values:   c.AppliedConfiguration.FatalExitCodes,
			},
			OnPodConditions: make([]batchv1.PodFailurePolicyOnPodConditionsPattern, 0),
		})
	}

	if c.AppliedConfiguration.TransientExitCodes != nil {
		jobPodFailurePolicy.Rules = append(jobPodFailurePolicy.Rules, batchv1.PodFailurePolicyRule{
			Action: batchv1.PodFailurePolicyActionIgnore,
			OnExitCodes: &batchv1.PodFailurePolicyOnExitCodesRequirement{
				Operator: batchv1.PodFailurePolicyOnExitCodesOpIn,
				Values:   c.AppliedConfiguration.TransientExitCodes,
			},
			OnPodConditions: make([]batchv1.PodFailurePolicyOnPodConditionsPattern, 0),
		})
	}

	return batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: batchv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: c.Id,
			Labels: map[string]string{
				"science.sneaksanddata.com/nexus-version": appVersion,
				"app.kubernetes.io/component":             "algorithm-run",
			},
		},
		Spec: batchv1.JobSpec{
			PodFailurePolicy: jobPodFailurePolicy,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"science.sneaksanddata.com/nexus-version": appVersion,
						"app.kubernetes.io/component":             "algorithm-run",
					},
					Annotations: map[string]string{
						"science.sneaksanddata.com/algorithm-template-name": c.Algorithm,
					},
				},
				Spec: corev1.PodSpec{
					Volumes: jobVolumes,
					Containers: []corev1.Container{
						{
							Name:            c.Id,
							Image:           fmt.Sprintf("%s/%s:%s", c.AppliedConfiguration.ImageRegistry, c.AppliedConfiguration.ImageRepository, c.AppliedConfiguration.ImageTag),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Resources: corev1.ResourceRequirements{
								Requests: jobResourceList,
								Limits:   jobResourceList,
							},
							Command: strings.Split(" ", c.AppliedConfiguration.Command),
							Args:    jobArgs,
							Env: append(c.AppliedConfiguration.Env, []corev1.EnvVar{
								{
									Name:  "NEXUS__ALGORITHM_NAME",
									Value: c.Algorithm,
								},
								{
									Name:  "NEXUS__BOUND_WORKGROUP",
									Value: c.AppliedConfiguration.WorkgroupHost,
								},
							}...),
							EnvFrom:      c.AppliedConfiguration.EnvFrom,
							VolumeMounts: jobVolumeMounts,
						},
					},
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      jobNodeTaint,
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{jobNodeTaintValue},
											},
										},
									},
								},
							},
						},
					},
					Tolerations:        jobTolerations,
					ServiceAccountName: c.AppliedConfiguration.ServiceAccountName,
					RestartPolicy:      "Never",
				},
			},
			BackoffLimit:            c.AppliedConfiguration.MaximumRetries,
			ActiveDeadlineSeconds:   ptr.Int64(int64(*c.AppliedConfiguration.DeadlineSeconds)),
			TTLSecondsAfterFinished: ptr.Int32(300),
		},
	}
}

func (c *CheckpointedRequest) IsFinished() bool {
	switch c.LifecycleStage {
	case LifecyclestageFailed, LifecyclestageCompleted, LifecyclestageDeadlineExceeded, LifecyclestageScheduleTimeout, LifecyclestageCancelled:
		return true
	default:
		return false
	}
}

func (c *CheckpointedRequest) AsCAJ011() *CheckpointedRequest {
	result := c.DeepCopy()
	result.LifecycleStage = LifecyclestageDeadlineExceeded
	result.AlgorithmFailureCode = CAJ011.ErrorName()
	result.AlgorithmFailureCause = CAJ011.ErrorMessage()

	return result
}

func (c *CheckpointedRequest) AsCAJ012() *CheckpointedRequest {
	result := c.DeepCopy()
	result.LifecycleStage = LifecyclestageScheduleTimeout
	result.AlgorithmFailureCode = CAJ012.ErrorName()
	result.AlgorithmFailureCause = CAJ012.ErrorMessage()

	return result
}

func (c *CheckpointedRequest) AsCB000() *CheckpointedRequest {
	result := c.DeepCopy()
	result.LifecycleStage = LifecyclestageFailed
	result.AlgorithmFailureCode = CB000.ErrorName()
	result.AlgorithmFailureCause = CB000.ErrorMessage()

	return result
}

func (c *CheckpointedRequest) AsCancelled(request CancellationRequest) *CheckpointedRequest {
	result := c.DeepCopy()
	result.LifecycleStage = LifecyclestageCancelled
	result.AlgorithmFailureCode = CAJ000.ErrorName()
	result.AlgorithmFailureCause = fmt.Sprintf(CAJ000.ErrorMessage(), request.Initiator, request.Reason)

	return result
}
