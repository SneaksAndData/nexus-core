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
	"github.com/DataDog/datadog-go/v5/statsd"
	"time"
)

type contextKey string

const (
	// MetricsNamespace sets the datadog metrics namespace to use
	MetricsNamespace = "ncc"

	MetricsClientContextKey contextKey = "metrics"
)

// WithStatsd enriches the context with a statsd client if it can be instantiated
func WithStatsd(ctx context.Context) context.Context { // coverage-ignore
	statsdClient, err := statsd.New("", statsd.WithNamespace(MetricsNamespace))
	if err == nil {
		return context.WithValue(ctx, MetricsClientContextKey, statsdClient)
	}

	return ctx
}

// Gauge reports a GAUGE metric using best-effort approach
func Gauge(metrics *statsd.Client, name string, value float64, tags []string, rate float64) { // coverage-ignore
	_ = metrics.Gauge(name, value, tags, rate)
}

// GaugeDuration reports a GAUGE metric corresponding to a duration of an operation that started at a specified time, in milliseconds
func GaugeDuration(metrics *statsd.Client, name string, startedAt time.Time, tags []string, rate float64) { // coverage-ignore
	duration := time.Since(startedAt).Milliseconds()
	_ = metrics.Gauge(name, float64(duration), tags, rate)
}
