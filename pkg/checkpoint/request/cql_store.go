package request

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v3"
	"io"
	"k8s.io/klog/v2"
)

type CqlStore struct {
	cluster    *gocql.ClusterConfig
	cqlSession gocqlx.Session
	logger     klog.Logger
}

type AstraBundleConfig struct {
	SecureConnectionBundleBase64 string `mapstructure:"secure-connection-bundle-base64"`
	GatewayUser                  string `mapstructure:"gateway-user"`
	GatewayPassword              string `mapstructure:"gateway-password"`
}

type AstraCqlStoreConfig struct {
	GatewayHost string
	GatewayPort string
	GatewayUser string
	GatewayPass string
	TlsConfig   *tls.Config
}

func getContent(zipFile *zip.File) ([]byte, error) {
	handle, err := zipFile.Open()
	if err != nil {
		return nil, err
	}

	defer func(handle io.ReadCloser) {
		_ = handle.Close()
	}(handle)

	return io.ReadAll(handle)
}

func NewAstraCqlStoreConfig(logger klog.Logger, config *AstraBundleConfig) *AstraCqlStoreConfig {
	zipReader, err := zip.NewReader(bytes.NewReader([]byte(config.SecureConnectionBundleBase64)), int64(len(config.SecureConnectionBundleBase64)))
	if err != nil {
		logger.V(0).Error(err, "Astra bundle content cannot be read")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	bundleFiles := map[string][]byte{}

	// Read all the files from the bundle
	for _, zipFile := range zipReader.File {
		bundleFileContent, err := getContent(zipFile)
		if err != nil {
			logger.V(0).Error(err, "Error when unpacking the bundle")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}

		bundleFiles[zipFile.Name] = bundleFileContent
	}

	cert, err := tls.X509KeyPair(bundleFiles["cert"], bundleFiles["key"])
	if err != nil {
		logger.V(0).Error(err, "Unable to instantiate X509KeyPair certificate from the bundle")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(bundleFiles["ca.crt"])
	var gatewayConfig map[string]string

	err = json.Unmarshal(bundleFiles["config.json"], &gatewayConfig)
	if err != nil {
		logger.V(0).Error(err, "Error parsing connection configuration json")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	return &AstraCqlStoreConfig{
		GatewayHost: gatewayConfig["host"],
		GatewayPort: gatewayConfig["cql_port"],
		GatewayPass: config.GatewayPassword,
		GatewayUser: config.GatewayUser,
		TlsConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		},
	}

}

// NewCqlStore creates a generic connected CqlStore (Apache Cassandra/Scylla)
func NewCqlStore(cluster *gocql.ClusterConfig, logger klog.Logger) *CqlStore {
	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		logger.V(0).Error(err, "failed to create CQL session")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	return &CqlStore{
		cluster:    cluster,
		cqlSession: session,
		logger:     logger,
	}
}

// NewAstraCqlStore creates a CqlStore connected to DataStax AstraDB serverless instance
func NewAstraCqlStore(logger klog.Logger, bundle *AstraBundleConfig) *CqlStore {
	config := NewAstraCqlStoreConfig(logger, bundle)
	cluster := gocql.NewCluster(config.GatewayHost)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.GatewayUser,
		Password: config.GatewayPass,
	}
	cluster.Hosts = []string{config.GatewayHost + ":" + config.GatewayPort}
	cluster.SslOpts = &gocql.SslOptions{
		Config:                 config.TlsConfig,
		EnableHostVerification: false,
	}
	cluster.Consistency = gocql.LocalQuorum
	return NewCqlStore(cluster, logger)
}
