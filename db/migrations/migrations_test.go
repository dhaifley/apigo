package migrations_test

import (
	"testing"

	"github.com/dhaifley/apigo/db/migrations"
	"github.com/dhaifley/apigo/internal/config"
	"github.com/dhaifley/apigo/internal/logger"
)

func TestMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}

	cfg := config.NewDefault()

	log := logger.New(cfg.LogOut(), cfg.LogFormat(), cfg.LogLevel())

	if err := migrations.Migrate(cfg, log); err != nil {
		t.Fatal(err)
	}
}
