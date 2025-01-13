package models

import v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"

type ParentAlgorithmRequest struct {
	RequestId     string `json:"requestId"`
	AlgorithmName string `json:"algorithmName"`
}

type AlgorithmRequest struct {
	AlgorithmParameters map[string]interface{}          `json:"algorithmParameters"`
	CustomConfiguration v1.MachineLearningAlgorithmSpec `json:"customConfiguration"`
	MonitoringMetadata  map[string][]string             `json:"monitoringMetadata"`
	RequestApiVersion   string                          `json:"requestApiVersion"`
	Tag                 string                          `json:"tag"`
	ParentRequest       ParentAlgorithmRequest          `json:"parentRequest,omitempty"`
}
