package cassandra

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/payload"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/store"
	"github.com/gocql/gocql"
	"k8s.io/klog/v2"
)

// ScyllaConfig defines configuration for gocql needed to connect to ScyllaDB
type ScyllaConfig struct {
	Hosts            []string `mapstructure:"hosts"`
	Port             string   `mapstructure:"port"`
	User             string   `mapstructure:"user"`
	Password         string   `mapstructure:"password"`
	LocalDC          string   `mapstructure:"local-dc"`
	IndexesSupported bool     `mapstructure:"indexes-supported"`
}

func NewScyllaStore(logger klog.Logger, config *ScyllaConfig, payloadProxyConfiguration *payload.RequestPayloadProxyConfiguration) store.CheckpointStore {
	cluster := gocql.NewCluster(config.Hosts...)
	fallback := gocql.RoundRobinHostPolicy()
	if config.LocalDC != "" { // coverage-ignore
		fallback = gocql.DCAwareRoundRobinPolicy(config.LocalDC)
	}

	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(fallback)
	if config.LocalDC != "" { // coverage-ignore
		cluster.Consistency = gocql.LocalQuorum
	}

	if config.Password != "" { // coverage-ignore
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: config.User,
			Password: config.Password,
		}
	}

	cassandraStore := NewCassandraStore(cluster, logger)

	if config.IndexesSupported {
		return NewIndexedCassandraStore(cassandraStore, payloadProxyConfiguration)
	}
	return NewBareCassandraStore(cassandraStore, payloadProxyConfiguration)
}
