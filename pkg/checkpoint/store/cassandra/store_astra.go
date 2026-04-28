package cassandra

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"io"
	"strconv"

	"github.com/gocql/gocql"
	"k8s.io/klog/v2"
)

type AstraBundleConfig struct {
	SecureConnectionBundleBase64 string `mapstructure:"secure-connection-bundle-base64"`
	GatewayUser                  string `mapstructure:"gateway-user"`
	GatewayPassword              string `mapstructure:"gateway-password"`
}

// CheckpointStoreAstraConfig defines configuration for gocql needed to connect to AstraDB
type CheckpointStoreAstraConfig struct {
	GatewayHost string
	GatewayPort string
	GatewayUser string
	GatewayPass string
	TlsConfig   *tls.Config
}

func getContent(zipFile *zip.File) ([]byte, error) { // coverage-ignore
	handle, err := zipFile.Open()
	if err != nil {
		return nil, err
	}

	defer func(handle io.ReadCloser) {
		_ = handle.Close()
	}(handle)

	return io.ReadAll(handle)
}

func NewAstraCqlStoreConfig(logger klog.Logger, config *AstraBundleConfig) *CheckpointStoreAstraConfig { // coverage-ignore
	bundleBytes, err := base64.StdEncoding.DecodeString(config.SecureConnectionBundleBase64)
	if err != nil {
		logger.V(0).Error(err, "bundle value is not a valid base64-encoded string")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(bundleBytes), int64(len(bundleBytes)))
	if err != nil {
		logger.V(0).Error(err, "bundle cannot be unpacked")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	bundleFiles := map[string][]byte{}

	// Read all the files from the bundle
	for _, zipFile := range zipReader.File {
		bundleFileContent, err := getContent(zipFile)
		if err != nil {
			logger.V(0).Error(err, "error when unpacking the bundle")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}

		bundleFiles[zipFile.Name] = bundleFileContent
	}

	cert, err := tls.X509KeyPair(bundleFiles["cert"], bundleFiles["key"])
	if err != nil {
		logger.V(0).Error(err, "unable to instantiate X509KeyPair certificate from the bundle")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(bundleFiles["ca.crt"])
	var gatewayConfig map[string]any

	err = json.Unmarshal(bundleFiles["config.json"], &gatewayConfig)
	if err != nil {
		logger.V(0).Error(err, "error parsing connection configuration json")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	return &CheckpointStoreAstraConfig{
		GatewayHost: gatewayConfig["host"].(string),
		GatewayPort: strconv.Itoa(int(gatewayConfig["cql_port"].(float64))),
		GatewayPass: config.GatewayPassword,
		GatewayUser: config.GatewayUser,
		TlsConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
			ServerName:   gatewayConfig["host"].(string),
		},
	}

}

// NewAstraStore creates a CqlStore connected to DataStax AstraDB serverless instance
func NewAstraStore(logger klog.Logger, bundle *AstraBundleConfig) *CheckpointCassandraStore { // coverage-ignore
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
	return NewCassandraStore(cluster, logger)
}
