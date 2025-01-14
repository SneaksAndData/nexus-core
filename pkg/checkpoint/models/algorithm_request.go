package models

import v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"

type ParentAlgorithmRequest struct {
	RequestId     string `json:"requestId" binding:"required"`
	AlgorithmName string `json:"algorithmName" binding:"required"`
}

type AlgorithmRequest struct {
	AlgorithmParameters map[string]interface{}          `json:"algorithmParameters" binding:"required"`
	CustomConfiguration v1.MachineLearningAlgorithmSpec `json:"customConfiguration,omitempty"`
	MonitoringMetadata  map[string][]string             `json:"monitoringMetadata,omitempty"`
	RequestApiVersion   string                          `json:"requestApiVersion,omitempty"`
	Tag                 string                          `json:"tag,omitempty"`
	ParentRequest       ParentAlgorithmRequest          `json:"parentRequest,omitempty"`
}
