package sqldb_test

import (
	"context"
	"testing"
	"time"

	"github.com/dhaifley/apigo/internal/config"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/sqldb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	testID   = "1"
	testUUID = "11223344-5566-7788-9900-aabbccddeeff"
)

func mockAuthContext() context.Context {
	ctx := context.Background()

	ctx = context.WithValue(ctx, request.CtxKeyAccountID, testID)

	return ctx
}

type mockSQLTrans struct{}

func (m *mockSQLTrans) Commit(ctx context.Context) error {
	return nil
}

func (m *mockSQLTrans) Exec(ctx context.Context,
	q string, args ...any,
) (sqldb.SQLResult, error) {
	return nil, nil
}

func (m *mockSQLTrans) Query(ctx context.Context,
	q string, args ...any,
) (sqldb.SQLRows, error) {
	return nil, nil
}

func (m *mockSQLTrans) QueryRow(ctx context.Context,
	q string, args ...any,
) sqldb.SQLRow {
	return nil
}

func (m *mockSQLTrans) Rollback(ctx context.Context) error {
	return nil
}

func (m *mockSQLTrans) CloseTx(ctx context.Context, err error) error {
	return nil
}

type mockSQLConn struct{}

func (m *mockSQLConn) BeginTx(ctx context.Context,
	opts pgx.TxOptions,
) (sqldb.SQLTX, error) {
	return nil, nil
}

func (m *mockSQLConn) Close() {
	return
}

func (m *mockSQLConn) Exec(ctx context.Context,
	q string, args ...any,
) (sqldb.SQLResult, error) {
	return nil, nil
}

func (m *mockSQLConn) Query(ctx context.Context,
	q string, args ...any,
) (sqldb.SQLRows, error) {
	return nil, nil
}

func (m *mockSQLConn) QueryRow(ctx context.Context,
	q string, args ...any,
) sqldb.SQLRow {
	return nil
}

func (m *mockSQLConn) Ping(ctx context.Context) error {
	return nil
}

func (m *mockSQLConn) Stat() *pgxpool.Stat {
	return &pgxpool.Stat{}
}

func TestNewSQLConn(t *testing.T) {
	t.Parallel()

	cfg := config.NewDefault()
	cfg.SetService(&config.ServiceConfig{Name: "test"})

	sc := sqldb.NewSQLConn(cfg, nil, nil, nil)

	if sc.Svc() != "test" {
		t.Errorf("Expected service: test, got: %v", sc.Svc())
	}
}

func TestConnect(t *testing.T) {
	t.Parallel()

	cfg := config.NewDefault()
	cfg.SetService(&config.ServiceConfig{Name: "test"})
	cfg.SetDB(&config.DBConfig{
		Conn: "postgres://test@test:5432" +
			"/test?sslmode=disable&binary_parameters=yes",
		Type: "postgres",
	})

	sc := sqldb.NewSQLConn(cfg, nil, nil, nil)

	defer sc.Close()

	ctx := context.Background()

	err := sc.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = sc.Test()
	if err == nil {
		t.Fatal("Expected connection error, got: nil")
	}
}

func TestReconnect(t *testing.T) {
	t.Parallel()

	cfg := config.NewDefault()
	cfg.SetService(&config.ServiceConfig{Name: "test"})
	cfg.SetDB(&config.DBConfig{
		Conn: "postgres://test@test:5432" +
			"/test?sslmode=disable&binary_parameters=yes",
		Type: "postgres",
	})

	sc := sqldb.NewSQLConn(cfg, nil, nil, nil)

	defer sc.Close()

	ctx := context.Background()

	err := sc.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = sc.Reconnect(mockAuthContext())
	if err == nil {
		t.Fatal("Expected error reconnecting")
	}
}

func TestBeginTx(t *testing.T) {
	t.Parallel()

	cfg := config.NewDefault()
	cfg.SetService(&config.ServiceConfig{Name: "test"})
	cfg.SetDB(&config.DBConfig{
		Conn: "postgres://test@test:5432" +
			"/test?sslmode=disable&binary_parameters=yes",
		Type: "postgres",
	})

	sc := sqldb.NewSQLConn(cfg, nil, nil, nil)

	defer sc.Close()

	ctx := context.Background()

	err := sc.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(mockAuthContext(), 5*time.Second)
	defer cancel()

	_, err = sc.BeginTx(ctx, pgx.TxOptions{})
	if err == nil {
		t.Fatal("Expected error, got: nil")
	}
}

func TestOperations(t *testing.T) {
	t.Parallel()

	cfg := config.NewDefault()
	cfg.SetService(&config.ServiceConfig{Name: "test"})
	cfg.SetDB(&config.DBConfig{
		Conn: "postgres://test@test:5432" +
			"/test?sslmode=disable&binary_parameters=yes",
		Type: "postgres",
	})

	sc := sqldb.NewSQLConn(cfg, nil, nil, nil)

	defer sc.Close()

	ctx := context.Background()

	err := sc.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(mockAuthContext(), 5*time.Second)
	defer cancel()

	if _, err = sc.Exec(ctx, "SELECT 1"); err == nil {
		t.Fatal("Expected error, got: nil")
	}

	if _, err = sc.Query(ctx, "SELECT 1"); err == nil {
		t.Fatal("Expected error, got: nil")
	}

	row := sc.QueryRow(ctx, "SELECT 1")

	if err = row.Scan(); err == nil {
		t.Fatal("Expected error, got: nil")
	}

	if err = sc.Ping(ctx); err == nil {
		t.Fatal("Expected error, got: nil")
	}
}

func TestStat(t *testing.T) {
	t.Parallel()

	cfg := config.NewDefault()
	cfg.SetService(&config.ServiceConfig{Name: "test"})
	cfg.SetDB(&config.DBConfig{
		Conn: "postgres://test@test:5432" +
			"/test?sslmode=disable&binary_parameters=yes",
		Type: "postgres",
	})

	sc := sqldb.NewSQLConn(cfg, nil, nil, nil)

	defer sc.Close()

	ctx := context.Background()

	err := sc.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	stat := sc.Stat()
	if stat.TotalConns() != 0 {
		t.Errorf("Expected total connections: 2, got: %v",
			stat.TotalConns())
	}
}

func TestMonitor(t *testing.T) {
	t.Parallel()

	cfg := config.NewDefault()
	cfg.SetService(&config.ServiceConfig{Name: "test"})
	cfg.SetDB(&config.DBConfig{
		Monitor: 1 * time.Second,
	})

	sc := sqldb.NewSQLConn(cfg, nil, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sc.Monitor(ctx)

	// Let monitor run for a bit
	time.Sleep(2 * time.Second)
}

func TestSQLConnLogFunctions(t *testing.T) {
	t.Parallel()

	sc := sqldb.NewSQLConn(nil, nil, nil, nil)
	ctx := context.Background()

	// Test all log levels
	sc.LogErrorf(ctx, "test", nil, "message %s", "test")
	sc.LogWarnf(ctx, "test", "message %s", "test")
	sc.LogInfof(ctx, "test", "message %s", "test")
	sc.LogDebugf(ctx, "test", "message %s", "test")
}

func TestSQLConnSetMode(t *testing.T) {
	t.Parallel()

	sc := sqldb.NewSQLConn(nil, nil, nil, nil)

	mode := 1
	sc.SetMode(mode)

	if sc.Mode() != mode {
		t.Errorf("Expected mode %d, got %d", mode, sc.Mode())
	}
}

func TestSQLConnAccessors(t *testing.T) {
	t.Parallel()

	cfg := config.NewDefault()

	cfg.SetService(&config.ServiceConfig{Name: "test"})

	sc := sqldb.NewSQLConn(cfg, nil, nil, nil)

	if sc.DB() != nil {
		t.Error("Expected nil DB")
	}

	if sc.Pool() != nil {
		t.Error("Expected nil Pool")
	}

	if sc.Svc() != "test" {
		t.Errorf("Expected service 'test', got %s", sc.Svc())
	}

	if sc.Log() == nil {
		t.Error("Expected non-nil Logger")
	}

	if sc.Metric() != nil {
		t.Error("Expected nil Metric")
	}

	if sc.Tracer() != nil {
		t.Error("Expected nil Tracer")
	}
}
