package models

import (
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"time"
)

type AlgorithmRequestRef struct {
	RequestId     string `json:"requestId" binding:"required"`
	AlgorithmName string `json:"algorithmName" binding:"required"`
}

type AlgorithmRequest struct {
	AlgorithmParameters map[string]interface{} `json:"algorithmParameters" binding:"required"`
	CustomConfiguration *v1.NexusAlgorithmSpec `json:"customConfiguration,omitempty"`
	RequestApiVersion   string                 `json:"requestApiVersion,omitempty"`
	Tag                 string                 `json:"tag,omitempty"`
	ParentRequest       *AlgorithmRequestRef   `json:"parentRequest,omitempty"`
	PayloadValidFor     time.Duration          `json:"payloadValidFor,omitempty"`
}
