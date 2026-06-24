package cassandra

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/scylladb/gocqlx/v3/table"
)

// table names
const (
	checkpointTableName         = "nexus.checkpoints"
	checkpointByHostTableName   = "nexus.checkpoints_by_host"
	checkpointByTagTableName    = "nexus.checkpoints_by_tag"
	checkpointPayloadsTableName = "nexus.checkpoint_payloads"
	EncodePrefix                = "b64__"
)

// table metadata
var (
	checkpointColumns = []string{
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
	checkpointByHostColumns = []string{
		"host",
		"lifecycle_stage",
		"algorithm",
		"id",
	}
	checkpointByTagColumns = []string{
		"tag",
		"algorithm",
		"id",
	}
	//checkpointPayloadsColumns = []string{
	//	"algorithm",
	//	"id",
	//	"payload",
	//}
)

// table definitions for goclqx
var (
	CheckpointedRequestTable = table.New(table.Metadata{
		Name:    checkpointTableName,
		Columns: checkpointColumns,
		PartKey: []string{
			"algorithm",
			"id",
		},
		SortKey: []string{},
	})
	CheckpointedRequestTableIndexByHost = table.New(table.Metadata{
		Name:    checkpointTableName,
		Columns: checkpointColumns,
		PartKey: []string{
			"received_by_host",
			"lifecycle_stage",
		},
		SortKey: []string{},
	})
	CheckpointedRequestTableByHost = table.New(table.Metadata{
		Name:    checkpointByHostTableName,
		Columns: checkpointByHostColumns,
		PartKey: []string{
			"host",
			"lifecycle_stage",
		},
		SortKey: []string{"id"},
	})
	CheckpointedRequestTableIndexByTag = table.New(table.Metadata{
		Name:    checkpointTableName,
		Columns: checkpointColumns,
		PartKey: []string{
			"tag",
		},
		SortKey: []string{},
	})
	CheckpointedRequestTableByTag = table.New(table.Metadata{
		Name:    checkpointByTagTableName,
		Columns: checkpointByTagColumns,
		PartKey: []string{
			"tag",
		},
		SortKey: []string{"id"},
	})
	//CheckpointedRequestPayloadTable = table.New(table.Metadata{
	//	Name:    checkpointPayloadsTableName,
	//	Columns: checkpointPayloadsColumns,
	//	PartKey: []string{
	//		"algorithm",
	//		"id",
	//	},
	//	SortKey: []string{},
	//})
)

type CheckpointCassandraModel struct {
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

func ToCassandraModel(request *models.CheckpointedRequest) (*CheckpointCassandraModel, error) {
	parent := []byte("{}")
	serializedOverrides := []byte("{}")
	serializedConfig, err := json.Marshal(request.AppliedConfiguration)

	if err != nil {
		return nil, err
	}

	if request.ConfigurationOverrides != nil {
		serializedOverrides, _ = json.Marshal(request.ConfigurationOverrides)
	}

	if request.Parent != nil {
		parent, err = json.Marshal(request.Parent)
		if err != nil {
			return nil, err
		}
	}

	return &CheckpointCassandraModel{
		Algorithm:               request.Algorithm,
		Id:                      request.Id,
		LifecycleStage:          request.LifecycleStage,
		PayloadUri:              request.PayloadUri,
		ResultUri:               request.ResultUri,
		AlgorithmFailureCause:   request.AlgorithmFailureCause,
		AlgorithmFailureDetails: request.AlgorithmFailureDetails,
		ReceivedByHost:          request.ReceivedByHost,
		ReceivedAt:              request.ReceivedAt,
		SentAt:                  request.SentAt,
		AppliedConfiguration:    fmt.Sprintf("%s%s", EncodePrefix, base64.StdEncoding.EncodeToString(serializedConfig)),
		ConfigurationOverrides:  fmt.Sprintf("%s%s", EncodePrefix, base64.StdEncoding.EncodeToString(serializedOverrides)),
		ContentHash:             request.ContentHash,
		LastModified:            request.LastModified,
		Tag:                     request.Tag,
		ApiVersion:              request.ApiVersion,
		JobUid:                  request.JobUid,
		Parent:                  fmt.Sprintf("%s%s", EncodePrefix, base64.StdEncoding.EncodeToString(parent)),
		PayloadValidFor:         request.PayloadValidFor,
	}, nil
}

func (c *CheckpointCassandraModel) readSerializedSpec(serializedSpec string) (*v1.NexusAlgorithmSpec, error) {
	spec := &v1.NexusAlgorithmSpec{}
	var serializedValue []byte
	var err error

	if serializedSpec == "{}" || serializedSpec == "" {
		return nil, nil
	}

	serializedValue, err = base64.StdEncoding.DecodeString(strings.TrimPrefix(serializedSpec, EncodePrefix))
	if err != nil {
		return nil, err
	}

	if string(serializedValue) == "{}" || string(serializedValue) == "" {
		return nil, nil
	}

	err = json.Unmarshal(serializedValue, spec)

	if err != nil {
		return nil, err
	}

	return spec, nil
}

func (c *CheckpointCassandraModel) getParent() (*models.AlgorithmRequestRef, error) {
	parent := &models.AlgorithmRequestRef{}
	var serializedValue []byte
	var err error

	if c.Parent == "" || c.Parent == "{}" {
		return nil, nil
	}

	// backwards-compatible code: only use b64 decode if it was used to write the value
	if strings.HasPrefix(c.Parent, EncodePrefix) {
		serializedValue, err = base64.StdEncoding.DecodeString(strings.TrimPrefix(c.Parent, EncodePrefix))
		if err != nil {
			return nil, err
		}

		if string(serializedValue) == "{}" || string(serializedValue) == "" {
			return nil, nil
		}
	} else {
		serializedValue = []byte(c.Parent)
	}
	err = json.Unmarshal(serializedValue, parent)

	if err != nil {
		return nil, err
	}

	return parent, nil
}

func (c *CheckpointCassandraModel) FromCassandraModel() (*models.CheckpointedRequest, error) {
	var appliedConfig *v1.NexusAlgorithmSpec
	var overrides *v1.NexusAlgorithmSpec
	var parent *models.AlgorithmRequestRef

	var unmarshalErr error

	// ignore override unmarshal if set to empty object
	overrides, unmarshalErr = c.readSerializedSpec(c.ConfigurationOverrides)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	appliedConfig, unmarshalErr = c.readSerializedSpec(c.AppliedConfiguration)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	parent, unmarshalErr = c.getParent()

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &models.CheckpointedRequest{
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

func (c *CheckpointCassandraModel) ByHostModel() interface{} {
	result := struct {
		Host           string `db:"host"`
		LifecycleStage string `db:"lifecycle_stage"`
		Algorithm      string `db:"algorithm"`
		Id             string `db:"id"`
	}{
		Host:           c.ReceivedByHost,
		LifecycleStage: c.LifecycleStage,
		Algorithm:      c.Algorithm,
		Id:             c.Id,
	}
	return &result
}

func (c *CheckpointCassandraModel) ByTagModel() interface{} {
	result := struct {
		Tag       string `db:"tag"`
		Algorithm string `db:"algorithm"`
		Id        string `db:"id"`
	}{
		Tag:       c.Tag,
		Algorithm: c.Algorithm,
		Id:        c.Id,
	}
	return &result
}
