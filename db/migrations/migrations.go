// Package migrations is used to process database migrations.
package migrations

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dhaifley/apigo/internal/config"
	"github.com/dhaifley/apigo/internal/errors"
	"github.com/dhaifley/apigo/internal/logger"
	"github.com/dhaifley/apigo/internal/sqldb"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/bitbucket"
	"github.com/golang-migrate/migrate/v4/source/github"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/stdlib"
)

// Database schema version.
const (
	CurrentVersion = 6
)

// mfs is a file system containing the database migrations.
//
//go:embed *.sql
var mfs embed.FS

// Migrate executes the required database migrations.
func Migrate(cfg *config.Config, log logger.Logger) error {
	ctx := context.Background()

	log.Log(ctx, logger.LvlInfo,
		"applying database migrations...")

	defer func() {
		if quit := os.Getenv("DB_SIDECAR_QUIT"); quit != "" {
			if _, err := http.Post(quit, "application/json",
				bytes.NewBufferString("{}")); err != nil {
				log.Log(ctx, logger.LvlError,
					"unable to shutdown cloud sql sidecar",
					"error", err)
			}
		}
	}()

	sc := sqldb.NewSQLConn(cfg, log, nil, nil)

	sc.SetMode(config.DBModeMigrate)

	if err := sc.Connect(ctx); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to connect to SQL database")
	}

	log.Log(ctx, logger.LvlInfo,
		"checking database connection...")

	err := sc.Ping(ctx)

	retry := 0

	for err != nil && retry < 5 {
		if errors.ErrorHas(err, `database "`+cfg.DBDatabase()+
			`" does not exist`) {
			isc := sqldb.NewSQLConn(cfg, log, nil, nil)

			isc.SetMode(config.DBModeInit)

			if err := isc.Connect(ctx); err != nil {
				return errors.Wrap(err, errors.ErrDatabase,
					"unable to connect to SQL database for initialization")
			}

			log.Log(ctx, logger.LvlInfo,
				"creating database...")

			if _, err := isc.ExecNoTx(ctx,
				`CREATE DATABASE "`+cfg.DBDatabase()+`" WITH OWNER="`+
					cfg.DBMigrateUser()+`"`,
			); err != nil {
				return errors.Wrap(err, errors.ErrDatabase,
					"unable to create database")
			}

			isc.Close()
		} else {
			log.Log(ctx, logger.LvlError,
				"unable to ping database, retrying...",
				"error", err)

			time.Sleep(time.Second)
		}

		retry++

		err = sc.Ping(ctx)
	}

	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to ping database")
	}

	mp := cfg.DBMigrations()

	var source source.Driver

	sourceName := "iofs"

	if strings.HasPrefix(mp, "github") {
		gh := &github.Github{}

		log.Log(ctx, logger.LvlInfo,
			"loading migration github source...")

		source, err = gh.Open(mp)
		if err != nil {
			return errors.Wrap(err, errors.ErrServer,
				"unable to initialize migrations github source")
		}

		sourceName = "github"
	} else if strings.HasPrefix(mp, "bitbucket") {
		bb := &bitbucket.Bitbucket{}

		log.Log(ctx, logger.LvlInfo,
			"loading migration bitbucket source...")

		source, err = bb.Open(mp)
		if err != nil {
			return errors.Wrap(err, errors.ErrServer,
				"unable to initialize migrations bitbucket source")
		}

		sourceName = "bitbucket"
	} else {
		log.Log(ctx, logger.LvlInfo,
			"loading migration file source...")

		source, err = iofs.New(mfs, ".")
		if err != nil {
			return errors.Wrap(err, errors.ErrServer,
				"unable to initialize migrations file source")
		}
	}

	log.Log(ctx, logger.LvlInfo,
		"initializing migration database connection...")

	driver, err := pgx.WithInstance(sql.OpenDB(
		stdlib.GetPoolConnector(sc.Pool())), &pgx.Config{})
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to initialize database migration driver")
	}

	m, err := migrate.NewWithInstance(sourceName, source,
		"postgres", driver)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to create database migration")
	}

	m.Log = migrationLog{log: log}

	ver, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to get database schema version")
	}

	if dirty {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to migrate database after failed migration")
	}

	if ver <= 1 || err != nil {
		log.Log(ctx, logger.LvlInfo,
			"creating database user...")

		if _, err := sc.ExecNoTx(ctx,
			`CREATE USER "`+cfg.DBUser()+`" WITH PASSWORD NULL`,
		); err != nil {
			return errors.Wrap(err, errors.ErrDatabase,
				"unable to create database user")
		}

		if password := cfg.DBPassword(); password != "" {
			if _, err := sc.ExecNoTx(ctx,
				`ALTER USER "`+cfg.DBUser()+`" WITH PASSWORD '`+
					password+`'`); err != nil {
				return errors.Wrap(err, errors.ErrDatabase,
					"unable to set database user password")
			}
		}

		if _, err := sc.ExecNoTx(ctx,
			`GRANT CONNECT ON DATABASE "`+cfg.DBDatabase()+
				`" TO "`+cfg.DBUser()+`"`); err != nil {
			return errors.Wrap(err, errors.ErrDatabase,
				"unable to grant permissions to database user")
		}
	}

	if err := m.Migrate(CurrentVersion); err != nil &&
		!errors.Is(err, migrate.ErrNoChange) {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to complete database migration")
	}

	log.Log(ctx, logger.LvlInfo,
		"granting database permissions...")

	if _, err := sc.ExecNoTx(ctx,
		`GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA "public" TO "`+
			cfg.DBUser()+`"`); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to grant database user sequence privileges")
	}

	if _, err := sc.ExecNoTx(ctx,
		`GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA "public" TO "`+
			cfg.DBUser()+`"`); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to grant database user table privileges")
	}

	return nil
}

// migrationLog values allow the service logger to be used with migrations.
type migrationLog struct {
	log logger.Logger
}

// Printf implements the logger.Printf interface.
func (ml migrationLog) Printf(format string, args ...interface{}) {
	ml.log.Log(context.Background(), logger.LvlInfo,
		fmt.Sprintf(format, args...))
}

// Verbose should return true when verbose logging output is wanted.
func (ml migrationLog) Verbose() bool {
	return false
}
