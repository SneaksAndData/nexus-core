/*
 * Copyright (c) 2024. ECCO Data & AI Open-Source Project Maintainers.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package telemetry

import (
	"context"
	"errors"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	slogdatadog "github.com/samber/slog-datadog/v2"
	slogmulti "github.com/samber/slog-multi"
	"log/slog"
	"os"
	"time"
)

type DatadogLoggerConfiguration struct {
	Endpoint    string
	ApiKey      string
	Hostname    string
	ServiceName string
}

func NewDatadogLoggerConfiguration() (*DatadogLoggerConfiguration, error) { // coverage-ignore
	if apiKey, endpoint, hostname, serviceName := os.Getenv("DATADOG__API_KEY"), os.Getenv("DATADOG__ENDPOINT"), os.Getenv("DATADOG__APPLICATION_HOST"), os.Getenv("DATADOG__SERVICE_NAME"); apiKey != "" && endpoint != "" && hostname != "" && serviceName != "" {
		return &DatadogLoggerConfiguration{
			Endpoint:    endpoint,
			ApiKey:      apiKey,
			Hostname:    hostname,
			ServiceName: serviceName,
		}, nil
	}

	return nil, errors.New("datadog API Key, Endpoint or Hostname and Service Name is not set")
}

func newDatadogClient(endpoint string, apiKey string, rootCtx context.Context) (*datadog.APIClient, context.Context) { // coverage-ignore
	ctx := datadog.NewDefaultContext(rootCtx)
	ctx = context.WithValue(
		ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{"apiKeyAuth": {Key: apiKey}},
	)
	ctx = context.WithValue(
		ctx,
		datadog.ContextServerVariables,
		map[string]string{"site": endpoint},
	)
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)

	return apiClient, ctx
}

func parseSLogLevel(levelText string) slog.Level { // coverage-ignore
	var level slog.Level
	var err = level.UnmarshalText([]byte(levelText))
	if err != nil {
		return slog.LevelInfo
	}

	return level
}

func ConfigureLogger(ctx context.Context, globalTags map[string]string, logLevel string) (*slog.Logger, error) { // coverage-ignore
	slogLevel := parseSLogLevel(logLevel)
	loggerConfig, err := NewDatadogLoggerConfiguration()
	// in case DD logger cannot be configured, use text handler and return error, so we can warn the user they are not getting DD logs recorded
	if err != nil {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel})), err
	}
	apiClient, ctx := newDatadogClient(loggerConfig.Endpoint, loggerConfig.ApiKey, ctx)
	return slog.New(
		slogmulti.Fanout(
			slogdatadog.Option{Level: slogLevel, Client: apiClient, Context: ctx, Timeout: 5 * time.Second, Hostname: loggerConfig.Hostname, Service: loggerConfig.ServiceName, GlobalTags: globalTags}.NewDatadogHandler(), // first send to Datadog handler
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel}), // then to second handler: stdout
		),
	), nil
}
