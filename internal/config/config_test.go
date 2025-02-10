package config_test

import (
	"os"
	"testing"

	"github.com/dhaifley/apigo/internal/config"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	os.Clearenv()

	cfg := config.NewDefault()

	cfg.Load([]byte(cfg.String()))

	cfg.Load([]byte(cfg.YAML()))

	if cfg.ServiceName() != config.DefaultServiceName {
		t.Errorf("Expected default config service name, got: %v",
			cfg.ServiceName())
	}
}
