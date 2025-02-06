package mocks

import (
	"context"
	"strings"
	"time"

	"github.com/dhaifley/apid/errors"
	"github.com/dhaifley/apid/request"
	"github.com/dhaifley/apid/resource"
	"github.com/dhaifley/apid/sqldb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var TestResource = resource.Resource{
	ResourceID: request.FieldString{
		Set: true, Valid: true,
		Value: TestUUID,
	},
	Name: request.FieldString{
		Set: true, Valid: true,
		Value: "testName",
	},
	Version: request.FieldString{
		Set: true, Valid: true,
		Value: "1",
	},
	Description: request.FieldString{
		Set: true, Valid: true,
		Value: "testDescription",
	},
	Status: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusNew,
	},
	StatusData: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"last_error": "testError",
		},
	},
	KeyField: request.FieldString{
		Set: true, Valid: true,
		Value: "resource_id",
	},
	KeyRegex: request.FieldString{
		Set: true, Valid: true,
		Value: ".*",
	},
	ClearCondition: request.FieldString{
		Set: true, Valid: true,
		Value: "gt(cleared_on:0)",
	},
	ClearAfter: request.FieldInt64{
		Set: true, Valid: true,
		Value: int64(time.Hour.Seconds()),
	},
	ClearDelay: request.FieldInt64{
		Set: true, Valid: true,
		Value: 0,
	},
	Data: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			TestUUID: map[string]any{
				"test":        "testData",
				"resource_id": TestUUID,
				"array": []map[string]any{{
					"status": "testStatus",
				}},
			},
		},
	},
	Source: request.FieldString{
		Set: true, Valid: true,
		Value: "testSource",
	},
	CommitHash: request.FieldString{
		Set: true, Valid: true,
		Value: "testHash",
	},
	CreatedByUser: &sqldb.UserData{},
	UpdatedByUser: &sqldb.UserData{},
}

type MockResourceRow struct{}

func (m *MockResourceRow) Scan(dest ...any) error {
	n := 0

	if v, ok := dest[n].(*string); ok {
		*v = TestResource.Status.Value
		n++

		if len(dest) <= n {
			return nil
		}
	}

	if v, ok := dest[n].(*int64); ok {
		*v = TestKey
		n++

		if len(dest) <= n {
			return nil
		}
	}

	if v, ok := dest[n].(*string); ok {
		*v = TestResource.ResourceID.Value
		n++

		if len(dest) <= n {
			return nil
		}
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.ResourceID
		n++

		if len(dest) <= n {
			return nil
		}
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.Name
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.Version
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.Description
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.Status
		n++
	}

	if v, ok := dest[n].(*request.FieldJSON); ok {
		*v = request.FieldJSON{
			Set: true, Valid: true, Value: map[string]any{},
		}

		for k, vv := range TestResource.StatusData.Value {
			(*v).Value[k] = vv
		}

		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.KeyField
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.KeyRegex
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.ClearCondition
		n++
	}

	if v, ok := dest[n].(*request.FieldInt64); ok {
		*v = TestResource.ClearAfter
		n++
	}

	if v, ok := dest[n].(*request.FieldInt64); ok {
		*v = TestResource.ClearDelay
		n++
	}

	if v, ok := dest[n].(*request.FieldJSON); ok {
		*v = request.FieldJSON{
			Set: true, Valid: true, Value: map[string]any{},
		}

		for k, vv := range TestResource.Data.Value {
			(*v).Value[k] = vv
		}

		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.Source
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.CommitHash
		n++
	}

	if v, ok := dest[n].(*request.FieldTime); ok {
		*v = TestResource.CreatedAt
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.CreatedBy
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.CreatedByUser.UserID
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.CreatedByUser.Email
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.CreatedByUser.LastName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.CreatedByUser.FirstName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.CreatedByUser.Status
		n++
	}

	if v, ok := dest[n].(*request.FieldJSON); ok {
		*v = TestResource.CreatedByUser.Data
		n++
	}

	if v, ok := dest[n].(*request.FieldTime); ok {
		*v = TestResource.UpdatedAt
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.UpdatedBy
		n++
	}

	if len(dest) <= n {
		return nil
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.UpdatedByUser.UserID
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.UpdatedByUser.Email
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.UpdatedByUser.LastName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.UpdatedByUser.FirstName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = TestResource.UpdatedByUser.Status
		n++
	}

	if v, ok := dest[n].(*request.FieldJSON); ok {
		*v = TestResource.UpdatedByUser.Data
	}

	return nil
}

type MockResourceRows struct {
	row int
}

func (m *MockResourceRows) Err() error {
	return nil
}

func (m *MockResourceRows) Close() {
	return
}

func (m *MockResourceRows) Next() bool {
	m.row++

	return m.row <= 1
}

func (m *MockResourceRows) Scan(dest ...interface{}) error {
	r := &MockResourceRow{}

	return r.Scan(dest...)
}

type MockResourceDB struct{}

func (m *MockResourceDB) BeginTx(ctx context.Context,
	opts pgx.TxOptions,
) (sqldb.SQLTX, error) {
	return &MockTx{db: m}, nil
}

func (m *MockResourceDB) Close() {
	return
}

func (m *MockResourceDB) Exec(ctx context.Context,
	q string, args ...any,
) (sqldb.SQLResult, error) {
	return &MockResult{}, nil
}

func (m *MockResourceDB) Query(ctx context.Context,
	q string, args ...any,
) (sqldb.SQLRows, error) {
	switch {
	case strings.Contains(q, "FROM token\n"):
		return &MockTokenRows{}, nil
	case strings.Contains(q, "FROM resource\n"):
		return &MockResourceRows{}, nil
	case strings.Contains(q, "FROM tag_obj\n"),
		strings.Contains(q, "DELETE FROM tag_obj"),
		strings.Contains(q, "INSERT INTO tag_obj"),
		strings.Contains(q, "UPDATE tag_obj SET"):
		return &MockTagRows{}, nil
	case strings.Contains(q, "FROM tag\n"),
		strings.Contains(q, "DELETE FROM tag"),
		strings.Contains(q, "INSERT INTO tag"),
		strings.Contains(q, "UPDATE tag SET"):
		return &MockTagRows{}, nil
	default:
		return nil, errors.New(errors.ErrDatabase, "invalid query")
	}
}

func (m *MockResourceDB) QueryRow(ctx context.Context,
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
	case strings.Contains(q, "FROM resource\n"),
		strings.Contains(q, `DELETE FROM resource`),
		strings.Contains(q, "INSERT INTO resource"),
		strings.Contains(q, "UPDATE resource SET"):
		return &MockResourceRow{}
	case strings.Contains(q, `WHERE "user".user_id = $1`),
		strings.Contains(q, `FROM "user"`+"\n"),
		strings.Contains(q, `DELETE FROM "user"`),
		strings.Contains(q, `INSERT INTO "user"`),
		strings.Contains(q, `UPDATE "user"`):
		return &MockUserRow{}
	case strings.Contains(q, "FROM tag_obj\n"),
		strings.Contains(q, "DELETE FROM tag_obj"),
		strings.Contains(q, "INSERT INTO tag_obj"),
		strings.Contains(q, "UPDATE tag_obj SET"):
		return &MockTagRow{}
	case strings.Contains(q, "FROM tag\n"),
		strings.Contains(q, "DELETE FROM tag"),
		strings.Contains(q, "INSERT INTO tag"),
		strings.Contains(q, "UPDATE tag SET"):
		return &MockTagRow{}
	default:
		return &MockRowError{}
	}
}

func (m *MockResourceDB) Ping(ctx context.Context) error {
	return nil
}

func (m *MockResourceDB) Stat() *pgxpool.Stat {
	return &pgxpool.Stat{}
}
