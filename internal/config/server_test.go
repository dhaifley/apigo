package config_test

import (
	"testing"
	"time"

	"github.com/dhaifley/apid/internal/config"
)

func TestServerConfig(t *testing.T) {
	t.Parallel()

	cfg := config.New("")

	cfg.Load(nil)

	cfg.SetServer(&config.ServerConfig{
		Address:        ":8090",
		Cert:           "test",
		Key:            "test",
		Timeout:        time.Second * 10,
		IdleTimeout:    time.Second * 10,
		Host:           "test.com",
		MaxRequestSize: 10,
	})

	if cfg.ServerAddress() != ":8090" {
		t.Errorf("Expected address: :8090, got: %v", cfg.ServerAddress())
	}

	if cfg.ServerCert() != "test" {
		t.Errorf("Expected cert: test, got: %v", cfg.ServerCert())
	}

	if cfg.ServerKey() != "test" {
		t.Errorf("Expected key: test, got: %v", cfg.ServerKey())
	}

	if cfg.ServerTimeout() != time.Second*10 {
		t.Errorf("Expected timeout: 10s, got: %v", cfg.ServerTimeout())
	}

	if cfg.ServerIdleTimeout() != time.Second*10 {
		t.Errorf("Expected idle timeout: 10s, got: %v",
			cfg.ServerIdleTimeout())
	}

	if cfg.ServerHost() != "test.com" {
		t.Errorf("Expected host: test.com, got: %v", cfg.ServerHost())
	}

	if cfg.ServerMaxRequestSize() != 10 {
		t.Errorf("Expected max request size: 10, got: %v",
			cfg.ServerMaxRequestSize())
	}
}
