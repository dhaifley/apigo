package mocks

import (
	"context"
	"strings"

	"github.com/dhaifley/apid/errors"
	"github.com/dhaifley/apid/request"
	"github.com/dhaifley/apid/sqldb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MockResult struct{}

func (m *MockResult) RowsAffected() int64 {
	return 1
}

type MockRowError struct{}

func (m *MockRowError) Scan(dest ...any) error {
	return errors.New(errors.ErrDatabase, "database error")
}

type MockNoRows struct{}

func (m *MockNoRows) Scan(dest ...any) error {
	return pgx.ErrNoRows
}

type MockUUIDRow struct{}

func (m *MockUUIDRow) Scan(dest ...any) error {
	n := 0

	if v, ok := dest[n].(*string); ok {
		*v = TestUUID
		n++

		if len(dest) <= n {
			return nil
		}
	}

	if v, ok := dest[n].(**string); ok {
		*v = new(string)
		**v = TestUUID
		n++

		if len(dest) <= n {
			return nil
		}
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = request.FieldString{Set: true, Valid: true, Value: TestUUID}
		n++

		if len(dest) <= n {
			return nil
		}
	}

	return nil
}

type MockTx struct {
	db sqldb.SQLDB
}

func (m *MockTx) Commit(ctx context.Context) error {
	return nil
}

func (m *MockTx) Exec(ctx context.Context,
	q string, args ...any,
) (sqldb.SQLResult, error) {
	return m.db.Exec(ctx, q, args...)
}

func (m *MockTx) Query(ctx context.Context,
	q string, args ...any,
) (sqldb.SQLRows, error) {
	return m.db.Query(ctx, q, args...)
}

func (m *MockTx) QueryRow(ctx context.Context,
	q string, args ...any,
) sqldb.SQLRow {
	return m.db.QueryRow(ctx, q, args...)
}

func (m *MockTx) Rollback(ctx context.Context) error {
	return nil
}

func (m *MockTx) CloseTx(ctx context.Context, err error) error {
	return nil
}

type MockAuthDB struct{}

func (m *MockAuthDB) BeginTx(ctx context.Context,
	opts pgx.TxOptions,
) (sqldb.SQLTX, error) {
	return &MockTx{db: m}, nil
}

func (m *MockAuthDB) Close() {
	return
}

func (m *MockAuthDB) Exec(ctx context.Context,
	q string, args ...any,
) (sqldb.SQLResult, error) {
	return &MockResult{}, nil
}

func (m *MockAuthDB) Query(ctx context.Context,
	q string, args ...any,
) (sqldb.SQLRows, error) {
	switch {
	case strings.Contains(q, "FROM token\n"):
		return &MockTokenRows{}, nil
	default:
		return nil, errors.New(errors.ErrDatabase, "invalid query")
	}
}

func (m *MockAuthDB) QueryRow(ctx context.Context,
	q string, args ...any,
) sqldb.SQLRow {
	switch {
	case strings.Contains(q, "SELECT account.account_id"):
		return &MockNoRows{}
	case strings.Contains(q, "FROM account"),
		strings.Contains(q, "INSERT INTO account"),
		strings.Contains(q, "UPDATE account SET"):
		return &MockAccountRow{}
	case strings.Contains(q, "FROM token\n"),
		strings.Contains(q, `DELETE FROM token`),
		strings.Contains(q, "INSERT INTO token"),
		strings.Contains(q, "UPDATE token SET"):
		return &MockTokenRow{}
	case strings.Contains(q, `WHERE "user".user_id = $1`),
		strings.Contains(q, `FROM "user"`+"\n"),
		strings.Contains(q, `DELETE FROM "user"`),
		strings.Contains(q, `INSERT INTO "user"`),
		strings.Contains(q, `UPDATE "user"`):
		return &MockUserRow{}
	default:
		return &MockRowError{}
	}
}

func (m *MockAuthDB) Ping(ctx context.Context) error {
	return nil
}

func (m *MockAuthDB) Stat() *pgxpool.Stat {
	return &pgxpool.Stat{}
}
