package config_test

import (
	"testing"
	"time"

	"github.com/dhaifley/apid/config"
)

func TestServiceConfig(t *testing.T) {
	t.Parallel()

	cfg := config.New("test name")

	cfg.Load(nil)

	cfg.SetService(&config.ServiceConfig{
		Name:           "test name",
		Maintenance:    true,
		ImportInterval: time.Second,
	})

	if cfg.ServiceName() != "test name" {
		t.Errorf("Expected name: test name, got: %v", cfg.ServiceName())
	}

	if cfg.ServiceMaintenance() != true {
		t.Errorf("Expected maintenance: true, got: %v",
			cfg.ServiceMaintenance())
	}

	if cfg.ImportInterval() != time.Second {
		t.Errorf("Expected import interval: 1s, got: %v", cfg.ImportInterval())
	}
}
