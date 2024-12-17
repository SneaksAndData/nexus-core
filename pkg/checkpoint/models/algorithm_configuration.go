package models

import (
	"github.com/SneaksAndData/nexus-core/pkg/util"
	v1 "k8s.io/api/core/v1"
)

// AlgorithmConfiguration is a static algorithm configuration used to configure algorithm job.
type AlgorithmConfiguration struct {
	ImageRegistry        string             `json:"imageRegistry"`
	ImageRepository      string             `json:"imageRepository"`
	ImageTag             string             `json:"imageTag"`
	DeadlineSeconds      int32              `json:"deadlineSeconds"`
	MaximumRetries       int32              `json:"maximumRetries"`
	Env                  []v1.EnvVar        `json:"env"`
	EnvFrom              []v1.EnvFromSource `json:"envFrom"`
	CpuLimit             string             `json:"cpuLimit"`
	MemoryLimit          string             `json:"memoryLimit"`
	WorkgroupHost        string             `json:"workgroupHost"`
	Workgroup            string             `json:"workgroup"`
	AdditionalWorkgroups map[string]string  `json:"additionalWorkgroups"`
	MonitoringParameters []string           `json:"monitoringParameters"`
	CustomResources      map[string]string  `json:"customResources"`
	SpeculativeAttempts  int32              `json:"speculativeAttempts"`
	TransientExitCodes   []int32            `json:"transientExitCodes"`
	FatalExitCodes       []int32            `json:"fatalExitCodes"`
	Command              string             `json:"command"`
	Args                 []string           `json:"args"`
	MountDatadogSocket   bool               `json:"mountDatadogSocket"`
}

func Merge(other *AlgorithmConfiguration) *AlgorithmConfiguration {
	cloned, _ := util.DeepClone(*other)

	cloned.ImageTag = util.Coalesce(other.ImageTag, cloned.ImageTag)
	cloned.DeadlineSeconds = util.Coalesce(other.DeadlineSeconds, cloned.DeadlineSeconds)
	cloned.MaximumRetries = util.Coalesce(other.MaximumRetries, cloned.MaximumRetries)
	cloned.Env = util.CoalesceArray(other.Env, cloned.Env)
	cloned.EnvFrom = util.CoalesceArray(other.EnvFrom, cloned.EnvFrom)
	cloned.CpuLimit = util.Coalesce(other.CpuLimit, cloned.CpuLimit)
	cloned.MemoryLimit = util.Coalesce(other.MemoryLimit, cloned.MemoryLimit)
	cloned.WorkgroupHost = util.Coalesce(other.WorkgroupHost, cloned.WorkgroupHost)
	cloned.Workgroup = util.Coalesce(other.Workgroup, cloned.Workgroup)

	return cloned
}
