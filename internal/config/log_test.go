package config_test

import (
	"log/slog"
	"testing"

	"github.com/dhaifley/apid/internal/config"
)

func TestLogConfig(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}

	cfg.SetLog(&config.LogConfig{
		Level:  config.LogLvlDebug,
		Out:    config.LogOutStdout,
		Format: config.LogFmtText,
	})

	cfg.Load(nil)

	if cfg.LogLevel() != slog.LevelDebug {
		t.Errorf("Expected log level: %v, got: %v",
			slog.LevelDebug, cfg.LogLevel())
	}

	if cfg.LogOut() != config.LogOutStdout {
		t.Errorf("Expected log out: %v, got: %v",
			config.LogFmtText, cfg.LogOut())
	}

	if cfg.LogFormat() != config.LogFmtText {
		t.Errorf("Expected log format: %v, got: %v",
			config.LogFmtText, cfg.LogFormat())
	}
}
