package configurations

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"os"
	"strings"
)

const (
	EnvPrefix = "NEXUS_" // varnames will be NEXUS__MY_ENV_VAR or NEXUS__SECTION1__SECTION2__MY_ENV_VAR
)

func configExists(configPath string) (bool, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func LoadConfig[T any](ctx context.Context) T {
	logger := klog.FromContext(ctx)
	customViper := viper.NewWithOptions(viper.KeyDelimiter("__"))
	customViper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	localConfig := fmt.Sprintf("appconfig.%s.yaml", strings.ToLower(os.Getenv("APPLICATION_ENVIRONMENT")))

	if exists, err := configExists(localConfig); exists {
		customViper.SetConfigFile(fmt.Sprintf("appconfig.%s.yaml", strings.ToLower(os.Getenv("APPLICATION_ENVIRONMENT"))))
	} else if err != nil {
		logger.Error(err, "could not locate application configuration file")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	} else {
		customViper.SetConfigFile("appconfig.yaml")
	}

	customViper.SetEnvPrefix(EnvPrefix)
	customViper.AllowEmptyEnv(true)
	customViper.AutomaticEnv()

	if err := customViper.ReadInConfig(); err != nil {
		logger.Error(err, "error loading application config from appconfig.yaml")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	var appConfig T
	err := customViper.Unmarshal(&appConfig)

	if err != nil {
		logger.Error(err, "error loading application config from appconfig.yaml")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	return appConfig
}
