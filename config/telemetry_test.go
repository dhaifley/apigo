package config_test

import (
	"testing"
	"time"

	"github.com/dhaifley/apid/config"
)

func TestTelemetryConfig(t *testing.T) {
	t.Parallel()

	exp := "test"

	cfg := config.NewDefault()

	cfg.SetTelemetry(&config.TelemetryConfig{
		MetricAddress:  exp,
		MetricInterval: time.Second,
		MetricVersion:  exp,
		TraceAddress:   exp,
	})

	if cfg.MetricAddress() != exp {
		t.Errorf("Expected metric address: %v, got: %v",
			exp, cfg.MetricAddress())
	}

	if cfg.MetricInterval() != time.Second {
		t.Errorf("Expected metric interval: 1s, got: %v",
			cfg.MetricInterval())
	}

	if cfg.MetricVersion() != exp {
		t.Errorf("Expected metric version: %v, got: %v",
			exp, cfg.MetricVersion())
	}

	if cfg.TraceAddress() != exp {
		t.Errorf("Expected trace address: %v, got: %v",
			exp, cfg.TraceAddress())
	}
}
