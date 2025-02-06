package config

import (
	"os"
	"time"
)

const (
	KeyMetricAddress  = "metric/address"
	KeyMetricInterval = "metric/interval"
	KeyMetricVersion  = "metric/version"
	KeyTraceAddress   = "trace/address"

	DefaultMetricAddress  = ""
	DefaultMetricInterval = time.Second * 60
	DefaultMetricVersion  = "v0.1.0"
	DefaultTraceAddress   = ""
)

// TelemetryConfig values represent telemetry configuration data.
type TelemetryConfig struct {
	MetricAddress  string        `json:"metric_address,omitempty"  yaml:"metric_address,omitempty"`
	MetricInterval time.Duration `json:"metric_interval,omitempty" yaml:"metric_interval,omitempty"`
	MetricVersion  string        `json:"metric_version,omitempty"  yaml:"metric_version,omitempty"`
	TraceAddress   string        `json:"trace_address,omitempty"   yaml:"trace_address,omitempty"`
}

// Load reads configuration data from environment variables and applies defaults
// for any missing or invalid configuration data.
func (c *TelemetryConfig) Load() {
	if v := os.Getenv(ReplaceEnv(KeyMetricAddress)); v != "" {
		c.MetricAddress = v
	}

	if c.MetricAddress == "" {
		c.MetricAddress = DefaultMetricAddress
	}

	if v := os.Getenv(ReplaceEnv(KeyMetricInterval)); v != "" {
		v, err := time.ParseDuration(v)
		if err != nil {
			v = DefaultMetricInterval
		}

		c.MetricInterval = v
	}

	if c.MetricInterval == 0 {
		c.MetricInterval = DefaultMetricInterval
	}

	if v := os.Getenv(ReplaceEnv(KeyMetricVersion)); v != "" {
		c.MetricVersion = v
	}

	if c.MetricVersion == "" {
		c.MetricVersion = DefaultMetricVersion
	}

	if v := os.Getenv(ReplaceEnv(KeyTraceAddress)); v != "" {
		c.TraceAddress = v
	}

	if c.TraceAddress == "" {
		c.TraceAddress = DefaultTraceAddress
	}
}

// MetricAddress returns the address of the collector where metrics data is
// sent.
func (c *Config) MetricAddress() string {
	c.RLock()
	defer c.RUnlock()

	if c.telemetry == nil {
		return DefaultMetricAddress
	}

	return c.telemetry.MetricAddress
}

// MetricInterval returns the periodic interval at which the service will
// update metrics data.
func (c *Config) MetricInterval() time.Duration {
	c.RLock()
	defer c.RUnlock()

	if c.telemetry == nil {
		return DefaultMetricInterval
	}

	return c.telemetry.MetricInterval
}

// MetricVersion returns the version of the metrics instrumentation data
// emitted by this service.
func (c *Config) MetricVersion() string {
	c.RLock()
	defer c.RUnlock()

	if c.telemetry == nil {
		return DefaultMetricVersion
	}

	return c.telemetry.MetricVersion
}

// TraceAddress returns the address of the collector where traces data is sent.
func (c *Config) TraceAddress() string {
	c.RLock()
	defer c.RUnlock()

	if c.telemetry == nil {
		return DefaultTraceAddress
	}

	return c.telemetry.TraceAddress
}
