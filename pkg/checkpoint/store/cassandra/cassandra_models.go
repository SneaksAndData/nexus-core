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
	checkpointTableName       = "%s.checkpoints"
	checkpointByHostTableName = "%s.checkpoints_by_host"
	checkpointByTagTableName  = "%s.checkpoints_by_tag"
	payloadBufferTable        = "%s.payload_buffer"
	EncodePrefix              = "b64__"
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
	payloadBufferColumns = []string{
		"algorithm",
		"id",
		"payload_content",
	}
)

// table definitions for goclqx

func CheckpointedRequestTable(keyspace string) *table.Table {
	return table.New(table.Metadata{
		Name:    fmt.Sprintf(checkpointTableName, keyspace),
		Columns: checkpointColumns,
		PartKey: []string{
			"algorithm",
			"id",
		},
		SortKey: []string{},
	})
}

func CheckpointedRequestTableIndexByHost(keyspace string) *table.Table {
	return table.New(table.Metadata{
		Name:    fmt.Sprintf(checkpointTableName, keyspace),
		Columns: checkpointColumns,
		PartKey: []string{
			"received_by_host",
			"lifecycle_stage",
		},
		SortKey: []string{},
	})
}

func CheckpointedRequestTableByHost(keyspace string) *table.Table {
	return table.New(table.Metadata{
		Name:    fmt.Sprintf(checkpointByHostTableName, keyspace),
		Columns: checkpointByHostColumns,
		PartKey: []string{
			"host",
			"lifecycle_stage",
		},
		SortKey: []string{"id"},
	})
}

func CheckpointedRequestTableIndexByTag(keyspace string) *table.Table {
	return table.New(table.Metadata{
		Name:    fmt.Sprintf(checkpointTableName, keyspace),
		Columns: checkpointColumns,
		PartKey: []string{
			"tag",
		},
		SortKey: []string{},
	})
}

func CheckpointedRequestTableByTag(keyspace string) *table.Table {
	return table.New(table.Metadata{
		Name:    fmt.Sprintf(checkpointByTagTableName, keyspace),
		Columns: checkpointByTagColumns,
		PartKey: []string{
			"tag",
		},
		SortKey: []string{"id"},
	})
}

func PayloadBufferTable(keyspace string) *table.Table {
	return table.New(table.Metadata{
		Name:    fmt.Sprintf(payloadBufferTable, keyspace),
		Columns: payloadBufferColumns,
		PartKey: []string{
			"algorithm",
			"id",
		},
		SortKey: []string{},
	})
}

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
