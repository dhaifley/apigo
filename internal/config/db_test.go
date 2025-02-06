package config_test

import (
	"testing"
	"time"

	"github.com/dhaifley/apid/internal/config"
)

func TestDatabaseConfig(t *testing.T) {
	t.Parallel()

	exp := "test"

	cfg := config.New("")

	cfg.Load(nil)

	cfg.SetDB(&config.DBConfig{
		User:            exp,
		Password:        "te:st",
		Database:        "api-db",
		MigrateUser:     "migrate",
		MigratePassword: "te:st",
		MigrateDatabase: "postgres",
		Instance:        exp,
		PrivateIP:       "1.1.1.1",
		Port:            "5432",
		Host:            "test-host",
		MaxConns:        10,
		Monitor:         time.Second * 10,
		Type:            exp,
		SSLMode:         "enable",
		DefaultSize:     10,
		MaxSize:         100,
		Migrations:      exp,
	})

	if cfg.DBInstance() != exp {
		t.Errorf("Expected instance: %v, got: %v", exp, cfg.DBInstance())
	}

	if cfg.DBMaxConns() != 10 {
		t.Errorf("Expected max connections: 10, got: %v", cfg.DBMaxConns())
	}

	expC := "test://test:te%3Ast@1.1.1.1:5432/api-db" +
		"?sslmode=enable"

	if cfg.DBConn(config.DBModeNormal) != expC {
		t.Errorf("Expected connection string: %v, got: %v",
			expC, cfg.DBConn(config.DBModeNormal))
	}

	expC = "test://migrate:te%3Ast@1.1.1.1:5432/api-db" +
		"?sslmode=enable"

	if cfg.DBConn(config.DBModeMigrate) != expC {
		t.Errorf("Expected connection string: %v, got: %v",
			expC, cfg.DBConn(config.DBModeMigrate))
	}

	expC = "test://migrate:te%3Ast@1.1.1.1:5432/postgres" +
		"?sslmode=enable"

	if cfg.DBConn(config.DBModeInit) != expC {
		t.Errorf("Expected connection string: %v, got: %v",
			expC, cfg.DBConn(config.DBModeInit))
	}

	if cfg.DBType() != exp {
		t.Errorf("Expected type: %v, got: %v", exp, cfg.DBType())
	}

	if cfg.DBMonitor() != (10 * time.Second) {
		t.Errorf("Expected monitor: 10s, got: %v", cfg.DBMonitor())
	}

	if cfg.DBDefaultSize() != 10 {
		t.Errorf("Expected default size: 10, got: %v", cfg.DBDefaultSize())
	}

	if cfg.DBMaxSize() != 100 {
		t.Errorf("Expected max size: 100, got: %v", cfg.DBMaxSize())
	}

	if cfg.DBUser() != exp {
		t.Errorf("Expected user: %v, got: %v", exp, cfg.DBUser())
	}

	exp = "migrate"

	if cfg.DBMigrateUser() != exp {
		t.Errorf("Expected user: %v, got: %v", exp, cfg.DBUser())
	}

	exp = "te:st"

	if cfg.DBPassword() != exp {
		t.Errorf("Expected password: %v, got: %v", exp, cfg.DBPassword())
	}

	if cfg.DBMigratePassword() != exp {
		t.Errorf("Expected password: %v, got: %v", exp, cfg.DBPassword())
	}

	exp = "1.1.1.1"

	if cfg.DBPrivateIP() != exp {
		t.Errorf("Expected private ip: %v, got: %v", exp, cfg.DBPrivateIP())
	}

	exp = "test-host"

	if cfg.DBHost() != exp {
		t.Errorf("Expected host: %v, got: %v", exp, cfg.DBHost())
	}

	exp = "5432"

	if cfg.DBPort() != exp {
		t.Errorf("Expected port: %v, got: %v", exp, cfg.DBPort())
	}

	exp = "enable"

	if cfg.DBSSLMode() != exp {
		t.Errorf("Expected ssl mode: %v, got: %v", exp, cfg.DBSSLMode())
	}

	exp = "api-db"

	if cfg.DBDatabase() != exp {
		t.Errorf("Expected database: %v, got: %v", exp, cfg.DBDatabase())
	}

	exp = "test"

	if cfg.DBMigrations() != exp {
		t.Errorf("Expected migrations: %v, got: %v", exp, cfg.DBMigrations())
	}

	cfg.SetDB(&config.DBConfig{Conn: exp})

	if cfg.DBConn(config.DBModeNormal) != exp {
		t.Errorf("Expected connection string: %v, got: %v", exp,
			cfg.DBConn(config.DBModeNormal))
	}
}
