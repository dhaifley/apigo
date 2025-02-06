package migrations_test

import (
	"os"
	"testing"

	"github.com/dhaifley/apid/config"
	"github.com/dhaifley/apid/db/migrations"
	"github.com/dhaifley/apid/logger"
)

func TestMigrations(t *testing.T) {
	cfg := config.NewDefault()

	host := os.Getenv("POSTGRES_HOST")

	if host == "" || os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("skipping integration test")
	}

	cfg.SetDB(&config.DBConfig{
		User:            "api-db-user",
		Password:        "api",
		Database:        "api-db",
		MigrateUser:     "postgres",
		MigratePassword: "postgres",
		MigrateDatabase: "postgres",
		Host:            host,
		Port:            "5432",
	})

	log := logger.New(cfg.LogOut(), cfg.LogFormat(), cfg.LogLevel())

	if err := migrations.Migrate(cfg, log); err != nil {
		t.Fatal(err)
	}
}
