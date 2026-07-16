package cassandra

import (
	"os"

	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/store"
	"github.com/aws/aws-sigv4-auth-cassandra-gocql-driver-plugin/sigv4"
	"github.com/gocql/gocql"
	"k8s.io/klog/v2"
)

// KeyspacesConfig defines configuration for gocql needed to connect to AWS Keyspaces
type KeyspacesConfig struct {
	Hosts []string `mapstructure:"hosts"`
	Port  string   `mapstructure:"port"`
	/**
	 * CaPath must contain file generated with commands below:
	 * curl -O https://www.amazontrust.com/repository/AmazonRootCA1.pem
	 * curl -O https://www.amazontrust.com/repository/AmazonRootCA2.pem
	 * curl -O https://www.amazontrust.com/repository/AmazonRootCA3.pem
	 * curl -O https://www.amazontrust.com/repository/AmazonRootCA4.pem
	 * curl -O https://certs.secureserver.net/repository/sf-class2-root.crt
	 * ----------------
	 * cat AmazonRootCA1.pem \
	 * AmazonRootCA2.pem \
	 * AmazonRootCA3.pem \
	 * AmazonRootCA4.pem \
	 * sf-class2-root.crt \
	 * > keyspaces-bundle.pem
	 */
	CaPath string `mapstructure:"ca-path"`
	Region string `mapstructure:"region"`
}

func (k *KeyspacesConfig) getIsolatedKeyspacesAuth() *sigv4.AwsAuthenticator { // coverage-ignore
	awsKeyID := os.Getenv("KEYSPACES_AWS_ACCESS_KEY_ID")
	awsSecret := os.Getenv("KEYSPACES_AWS_SECRET_ACCESS_KEY")

	return &sigv4.AwsAuthenticator{
		Region:          k.Region,
		AccessKeyId:     awsKeyID,
		SecretAccessKey: awsSecret,
	}
}

func NewKeyspacesStore(logger klog.Logger, config *KeyspacesConfig) store.CheckpointStore { // coverage-ignore
	cluster := gocql.NewCluster(config.Hosts...)

	cluster.Authenticator = config.getIsolatedKeyspacesAuth()
	cluster.SslOpts = &gocql.SslOptions{
		CaPath:                 config.CaPath,
		EnableHostVerification: false,
	}

	cluster.Consistency = gocql.LocalQuorum
	cluster.DisableInitialHostLookup = false

	cassandraStore := NewCassandraStore(cluster, logger)

	// Keyspaces do not support indexed tables
	return NewBareCassandraStore(cassandraStore)
}
