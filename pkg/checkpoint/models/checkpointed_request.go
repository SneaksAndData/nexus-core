package models

import (
	"encoding/json"
	"errors"
	"fmt"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
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
	LifecycleStageNew              = "NEW"
	LifecycleStageBuffered         = "BUFFERED"
	LifecycleStageRunning          = "RUNNING"
	LifecycleStageCompleted        = "COMPLETED"
	LifecycleStageFailed           = "FAILED"
	LifecycleStageSchedulingFailed = "SCHEDULING_FAILED"
	LifecycleStageDeadlineExceeded = "DEADLINE_EXCEEDED"
	LifecycleStageCancelled        = "CANCELLED"

	JobTemplateNameKey          = "science.sneaksanddata.com/algorithm-template-name"
	JobLabelFrameworkVersionKey = "science.sneaksanddata.com/nexus-version"
	NexusComponentLabel         = "science.sneaksanddata.com/nexus-component"
	JobLabelAlgorithmRun        = "algorithm-run"
)

type CheckpointedRequest struct {
	Algorithm               string                 `json:"algorithm"`
	Id                      string                 `json:"id"`
	LifecycleStage          string                 `json:"lifecycle_stage"`
	PayloadUri              string                 `json:"payload_uri"`
	ResultUri               string                 `json:"result_uri"`
	AlgorithmFailureCause   string                 `json:"algorithm_failure_cause"`
	AlgorithmFailureDetails string                 `json:"algorithm_failure_details"`
	ReceivedByHost          string                 `json:"received_by_host"`
	ReceivedAt              time.Time              `json:"received_at"`
	SentAt                  time.Time              `json:"sent_at"`
	AppliedConfiguration    *v1.NexusAlgorithmSpec `json:"applied_configuration,omitempty"`
	ConfigurationOverrides  *v1.NexusAlgorithmSpec `json:"configuration_overrides,omitempty"`
	ContentHash             string                 `json:"content_hash"`
	LastModified            time.Time              `json:"last_modified"`
	Tag                     string                 `json:"tag,omitempty"`
	ApiVersion              string                 `json:"api_version"`
	JobUid                  string                 `json:"job_uid,omitempty"`
	Parent                  *AlgorithmRequestRef   `json:"parent,omitempty"`
	PayloadValidFor         string                 `json:"payload_valid_for,omitempty"`
}

type CheckpointedRequestCqlModel struct {
	Algorithm               string
	Id                      string
	LifecycleStage          string
	PayloadUri              string
	ResultUri               string
	AlgorithmFailureCause   string
	AlgorithmFailureDetails string
	ReceivedByHost          string
	ReceivedAt              time.Time
	SentAt                  time.Time
	AppliedConfiguration    string
	ConfigurationOverrides  string
	ContentHash             string
	LastModified            time.Time
	Tag                     string
	ApiVersion              string
	JobUid                  string
	Parent                  string
	PayloadValidFor         string
}

var checkpointColumns = []string{
	"algorithm",
	"id",
	"lifecycle_stage",
	"payload_uri",
	"result_uri",
	"algorithm_failure_cause",
	"algorithm_failure_details",
	"received_by_host",
	"received_at",
	"sent_at",
	"applied_configuration",
	"configuration_overrides",
	"content_hash",
	"last_modified",
	"tag",
	"api_version",
	"job_uid",
	"parent",
	"payload_valid_for",
}

const tableName = "nexus.checkpoints"

var CheckpointedRequestTable = table.New(table.Metadata{
	Name:    tableName,
	Columns: checkpointColumns,
	PartKey: []string{
		"algorithm",
		"id",
	},
	SortKey: []string{},
})

var CheckpointedRequestTableIndexByHost = table.New(table.Metadata{
	Name:    tableName,
	Columns: checkpointColumns,
	PartKey: []string{
		"received_by_host",
		"lifecycle_stage",
	},
	SortKey: []string{},
})

var CheckpointedRequestTableIndexByTag = table.New(table.Metadata{
	Name:    tableName,
	Columns: checkpointColumns,
	PartKey: []string{
		"tag",
	},
	SortKey: []string{},
})

func (c *CheckpointedRequest) ToCqlModel() (*CheckpointedRequestCqlModel, error) {
	parent := []byte("{}")
	serializedOverrides := []byte("{}")
	serializedConfig, err := json.Marshal(c.AppliedConfiguration)

	if err != nil {
		return nil, err
	}

	if c.ConfigurationOverrides != nil {
		serializedOverrides, _ = json.Marshal(c.ConfigurationOverrides)
	}

	if c.Parent != nil {
		parent, err = json.Marshal(c.Parent)
		if err != nil {
			return nil, err
		}
	}

	return &CheckpointedRequestCqlModel{
		Algorithm:               c.Algorithm,
		Id:                      c.Id,
		LifecycleStage:          c.LifecycleStage,
		PayloadUri:              c.PayloadUri,
		ResultUri:               c.ResultUri,
		AlgorithmFailureCause:   c.AlgorithmFailureCause,
		AlgorithmFailureDetails: c.AlgorithmFailureDetails,
		ReceivedByHost:          c.ReceivedByHost,
		ReceivedAt:              c.ReceivedAt,
		SentAt:                  c.SentAt,
		AppliedConfiguration:    string(serializedConfig),
		ConfigurationOverrides:  string(serializedOverrides),
		ContentHash:             c.ContentHash,
		LastModified:            c.LastModified,
		Tag:                     c.Tag,
		ApiVersion:              c.ApiVersion,
		JobUid:                  c.JobUid,
		Parent:                  string(parent),
		PayloadValidFor:         c.PayloadValidFor,
	}, nil
}

func (c *CheckpointedRequest) PayloadValidityPeriod() *time.Duration {
	if c.PayloadValidFor == "" {
		return nil
	}

	result, _ := time.ParseDuration(c.PayloadValidFor)
	return &result
}

func (c *CheckpointedRequestCqlModel) FromCqlModel() (*CheckpointedRequest, error) {
	appliedConfig := &v1.NexusAlgorithmSpec{}
	var overrides *v1.NexusAlgorithmSpec
	var parent *AlgorithmRequestRef

	var unmarshalErr error

	// ignore override unmarshal if set to empty object
	if c.ConfigurationOverrides == "{}" || c.ConfigurationOverrides == "" {
		unmarshalErr = json.Unmarshal([]byte(c.AppliedConfiguration), appliedConfig)
	} else { // coverage-ignore
		overrides = &v1.NexusAlgorithmSpec{}
		unmarshalErr = errors.Join(json.Unmarshal([]byte(c.AppliedConfiguration), appliedConfig), json.Unmarshal([]byte(c.ConfigurationOverrides), overrides))
	}

	if c.Parent != "" && c.Parent != "{}" {
		parent = &AlgorithmRequestRef{}
		unmarshalErr = errors.Join(unmarshalErr, json.Unmarshal([]byte(c.Parent), parent))
	}

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &CheckpointedRequest{
		Algorithm:               c.Algorithm,
		Id:                      c.Id,
		LifecycleStage:          c.LifecycleStage,
		PayloadUri:              c.PayloadUri,
		ResultUri:               c.ResultUri,
		AlgorithmFailureCause:   c.AlgorithmFailureCause,
		AlgorithmFailureDetails: c.AlgorithmFailureDetails,
		ReceivedByHost:          c.ReceivedByHost,
		ReceivedAt:              c.ReceivedAt,
		SentAt:                  c.SentAt,
		AppliedConfiguration:    appliedConfig,
		ConfigurationOverrides:  overrides,
		ContentHash:             c.ContentHash,
		LastModified:            c.LastModified,
		Tag:                     c.Tag,
		ApiVersion:              c.ApiVersion,
		JobUid:                  c.JobUid,
		Parent:                  parent,
		PayloadValidFor:         c.PayloadValidFor,
	}, nil
}

func FromAlgorithmRequest(requestId string, algorithmName string, request *AlgorithmRequest, config *v1.NexusAlgorithmSpec) (*CheckpointedRequest, []byte, error) {
	hostname, _ := os.Hostname()
	serializedPayload, err := json.Marshal(request.AlgorithmParameters)

	if err != nil {
		return nil, nil, err
	}

	// check time.Duration
	if request.PayloadValidFor != "" {
		_, err = time.ParseDuration(request.PayloadValidFor)

		if err != nil {
			return nil, nil, err
		}
	}

	return &CheckpointedRequest{
		Algorithm:              algorithmName,
		Id:                     requestId,
		LifecycleStage:         LifecycleStageNew,
		ReceivedByHost:         hostname,
		ReceivedAt:             time.Now(),
		LastModified:           time.Now(),
		ConfigurationOverrides: request.CustomConfiguration,
		Tag:                    request.Tag,
		JobUid:                 "",
		Parent:                 request.ParentRequest,
		ApiVersion:             request.RequestApiVersion,
		AppliedConfiguration:   config.Merge(request.CustomConfiguration),
		PayloadValidFor:        request.PayloadValidFor,
	}, serializedPayload, nil
}

func (c *CheckpointedRequest) DeepCopy() *CheckpointedRequest {
	return &CheckpointedRequest{
		Algorithm:               c.Algorithm,
		Id:                      c.Id,
		LifecycleStage:          c.LifecycleStage,
		PayloadUri:              c.PayloadUri,
		ResultUri:               c.ResultUri,
		AlgorithmFailureCause:   c.AlgorithmFailureCause,
		AlgorithmFailureDetails: c.AlgorithmFailureDetails,
		ReceivedByHost:          c.ReceivedByHost,
		ReceivedAt:              c.ReceivedAt,
		SentAt:                  c.SentAt,
		AppliedConfiguration:    c.AppliedConfiguration.DeepCopy(),
		ConfigurationOverrides:  c.ConfigurationOverrides.DeepCopy(),
		ContentHash:             c.ContentHash,
		LastModified:            c.LastModified,
		Tag:                     c.Tag,
		ApiVersion:              c.ApiVersion,
		JobUid:                  c.JobUid,
		Parent:                  c.Parent.DeepCopy(),
		PayloadValidFor:         c.PayloadValidFor,
	}
}

func defaultFailurePolicy() *batchv1.PodFailurePolicy {
	return &batchv1.PodFailurePolicy{
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
}

func (c *CheckpointedRequest) ToV1Job(appVersion string, workgroup *v1.NexusAlgorithmWorkgroupSpec, parent *metav1.OwnerReference) batchv1.Job {
	owners := []metav1.OwnerReference{}
	if parent != nil {
		owners = append(owners, *parent)
	}
	jobResourceLimits := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(c.AppliedConfiguration.ComputeResources.CpuLimit),
		corev1.ResourceMemory: resource.MustParse(c.AppliedConfiguration.ComputeResources.MemoryLimit),
	}

	jobResourceRequests := jobResourceLimits.DeepCopy()
	cpuLimit := jobResourceRequests[corev1.ResourceCPU]
	jobResourceRequests[corev1.ResourceCPU] = resource.MustParse(fmt.Sprintf("%.0fm", float64(cpuLimit.MilliValue())*0.1))

	for customResourceKey, customResourceValue := range c.AppliedConfiguration.ComputeResources.CustomResources {
		jobResourceLimits[corev1.ResourceName(customResourceKey)] = resource.MustParse(customResourceValue)
		jobResourceRequests[corev1.ResourceName(customResourceKey)] = resource.MustParse(customResourceValue)
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

	jobVolumes := []corev1.Volume{}
	jobVolumeMounts := []corev1.VolumeMount{}

	jobPodFailurePolicy := defaultFailurePolicy()

	if *c.AppliedConfiguration.DatadogIntegrationSettings.MountDatadogSocket {
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

	if c.AppliedConfiguration.ErrorHandlingBehaviour != nil { // coverage-ignore
		if c.AppliedConfiguration.ErrorHandlingBehaviour.FatalExitCodes != nil {
			jobPodFailurePolicy.Rules = append(jobPodFailurePolicy.Rules, batchv1.PodFailurePolicyRule{
				Action: batchv1.PodFailurePolicyActionFailJob,
				OnExitCodes: &batchv1.PodFailurePolicyOnExitCodesRequirement{
					Operator: batchv1.PodFailurePolicyOnExitCodesOpIn,
					Values:   c.AppliedConfiguration.ErrorHandlingBehaviour.FatalExitCodes,
				},
				OnPodConditions: make([]batchv1.PodFailurePolicyOnPodConditionsPattern, 0),
			})
		}

		if c.AppliedConfiguration.ErrorHandlingBehaviour.TransientExitCodes != nil {
			jobPodFailurePolicy.Rules = append(jobPodFailurePolicy.Rules, batchv1.PodFailurePolicyRule{
				Action: batchv1.PodFailurePolicyActionIgnore,
				OnExitCodes: &batchv1.PodFailurePolicyOnExitCodesRequirement{
					Operator: batchv1.PodFailurePolicyOnExitCodesOpIn,
					Values:   c.AppliedConfiguration.ErrorHandlingBehaviour.TransientExitCodes,
				},
				OnPodConditions: make([]batchv1.PodFailurePolicyOnPodConditionsPattern, 0),
			})
		}
	}

	parentInfo := []corev1.EnvVar{}

	if c.Parent != nil {
		parentInfo = []corev1.EnvVar{
			{
				Name:  "NEXUS__PARENT_REQUEST_ID",
				Value: c.Parent.RequestId,
			},
			{
				Name:  "NEXUS__PARENT_ALGORITHM_NAME",
				Value: c.Parent.AlgorithmName,
			},
		}
	}

	jobEnv := append(c.AppliedConfiguration.RuntimeEnvironment.EnvironmentVariables, []corev1.EnvVar{
		{
			Name:  "NEXUS__ALGORITHM_NAME",
			Value: c.Algorithm,
		},
		{
			Name:  "NEXUS__SHARD_NAME",
			Value: workgroup.Cluster,
		},
	}...)

	return batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: batchv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: c.Id,
			Labels: map[string]string{
				JobTemplateNameKey:          c.Algorithm,
				JobLabelFrameworkVersionKey: appVersion,
				NexusComponentLabel:         JobLabelAlgorithmRun,
			},
			Annotations:     c.AppliedConfiguration.RuntimeEnvironment.Annotations,
			OwnerReferences: owners,
		},
		Spec: batchv1.JobSpec{
			PodFailurePolicy: jobPodFailurePolicy,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						JobTemplateNameKey:          c.Algorithm,
						JobLabelFrameworkVersionKey: appVersion,
						NexusComponentLabel:         JobLabelAlgorithmRun,
					},
					Annotations: c.AppliedConfiguration.RuntimeEnvironment.Annotations,
				},
				Spec: corev1.PodSpec{
					Volumes: jobVolumes,
					Containers: []corev1.Container{
						{
							Name:            c.Id,
							Image:           fmt.Sprintf("%s/%s:%s", c.AppliedConfiguration.Container.Registry, c.AppliedConfiguration.Container.Image, c.AppliedConfiguration.Container.VersionTag),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Resources: corev1.ResourceRequirements{
								Requests: jobResourceRequests,
								Limits:   jobResourceLimits,
							},
							Command:      strings.Split(c.AppliedConfiguration.Command, " "),
							Args:         jobArgs,
							Env:          append(jobEnv, parentInfo...),
							EnvFrom:      c.AppliedConfiguration.RuntimeEnvironment.MappedEnvironmentVariables,
							VolumeMounts: jobVolumeMounts,
						},
					},
					Affinity:           workgroup.Affinity,
					Tolerations:        workgroup.Tolerations,
					ServiceAccountName: c.AppliedConfiguration.Container.ServiceAccountName,
					RestartPolicy:      "Never",
				},
			},
			BackoffLimit:            c.AppliedConfiguration.RuntimeEnvironment.MaximumRetries,
			ActiveDeadlineSeconds:   ptr.Int64(int64(*c.AppliedConfiguration.RuntimeEnvironment.DeadlineSeconds)),
			TTLSecondsAfterFinished: ptr.Int32(300),
		},
	}
}

func (c *CheckpointedRequest) IsFinished() bool {
	switch c.LifecycleStage {
	case LifecycleStageFailed, LifecycleStageCompleted, LifecycleStageDeadlineExceeded, LifecycleStageSchedulingFailed, LifecycleStageCancelled:
		return true
	case LifecycleStageRunning:
		return false
	default:
		return false
	}
}
