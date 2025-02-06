package sqldb_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/dhaifley/apid/internal/cache"
	"github.com/dhaifley/apid/internal/errors"
	"github.com/dhaifley/apid/internal/request"
	"github.com/dhaifley/apid/internal/search"
	"github.com/dhaifley/apid/internal/sqldb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestParseFieldOptions(t *testing.T) {
	t.Parallel()

	options, err := sqldb.ParseFieldOptions(url.Values{
		"user_details":       []string{"true"},
		"config_assignments": []string{"true"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !options.Contains(sqldb.OptUserDetails) {
		t.Errorf("Expected: %v, got: %v", sqldb.OptUserDetails, options)
	}
}

func TestUserFields(t *testing.T) {
	t.Parallel()

	f := sqldb.UserFields("test")

	exp := 14

	if len(f) != exp {
		t.Errorf("Expected length: %v, got: %v", exp, len(f))
	}
}

func TestSelectFields(t *testing.T) {
	t.Parallel()

	v := sqldb.SelectFields("test", sqldb.UserFields("test"), nil,
		[]sqldb.FieldOption{sqldb.OptUserDetails})

	exp := `SELECT
	EXTRACT(epoch FROM test.created_at)::BIGINT AS test_created_at,
	create_user.user_id AS create_user_user_id,
	create_user.email AS create_user_email,
	create_user.last_name AS create_user_last_name,
	create_user.first_name AS create_user_first_name,
	create_user.status AS create_user_status,
	create_user.data AS create_user_data,
	EXTRACT(epoch FROM test.updated_at)::BIGINT AS test_updated_at,
	update_user.user_id AS update_user_user_id,
	update_user.email AS update_user_email,
	update_user.last_name AS update_user_last_name,
	update_user.first_name AS update_user_first_name,
	update_user.status AS update_user_status,
	update_user.data AS update_user_data
FROM test
LEFT JOIN "user" create_user ON (create_user.user_key = test.created_by)
LEFT JOIN "user" update_user ON (update_user.user_key = test.updated_by)
`

	if v != exp {
		t.Errorf("Expected: %v, got: %v", exp, v)
	}

	v = sqldb.SelectFields("test", []*sqldb.Field{{
		Name:  "status",
		Table: "test",
		Type:  sqldb.FieldString,
	}}, &search.Query{
		Summary: "status",
	}, nil)

	exp = `SELECT
	test.status AS test_status,
	COUNT(*) AS count
FROM test
`

	if v != exp {
		t.Errorf("Expected: %v, got: %v", exp, v)
	}
}

func TestSearchFields(t *testing.T) {
	t.Parallel()

	exp := "test"

	v := sqldb.SearchFields(exp, append([]*sqldb.Field{{
		Name:  exp + "_key",
		Table: exp,
		Type:  sqldb.FieldInt,
	}}, sqldb.UserFields(exp)...))

	exp = `SELECT
	test.test_key AS test_test_key
FROM test
LEFT JOIN "user" create_user ON (create_user.user_key = test.created_by)
LEFT JOIN "user" update_user ON (update_user.user_key = test.updated_by)
`

	if v != exp {
		t.Errorf("Expected: %v, got: %v", exp, v)
	}
}

func TestReturningFields(t *testing.T) {
	t.Parallel()

	exp := "test"

	v := sqldb.ReturningFields(exp, sqldb.UserFields(exp),
		[]sqldb.FieldOption{sqldb.OptUserDetails})

	exp = "\n" + `RETURNING
	EXTRACT(epoch FROM test.created_at)::BIGINT AS test_created_at,
	(SELECT create_user_1.user_id AS create_user_1_user_id FROM "user" ` +
		`create_user_1 WHERE create_user_1.user_key = test.created_by LIMIT 1),
	(SELECT create_user_2.email AS create_user_2_email FROM "user" ` +
		`create_user_2 WHERE create_user_2.create_user_key = ` +
		`test.create_user_key LIMIT 1),
	(SELECT create_user_3.last_name AS create_user_3_last_name FROM ` +
		`"user" create_user_3 WHERE create_user_3.create_user_key = ` +
		`test.create_user_key LIMIT 1),
	(SELECT create_user_4.first_name AS create_user_4_first_name FROM ` +
		`"user" create_user_4 WHERE create_user_4.create_user_key = ` +
		`test.create_user_key LIMIT 1),
	(SELECT create_user_5.status AS create_user_5_status FROM ` +
		`"user" create_user_5 WHERE create_user_5.create_user_key = ` +
		`test.create_user_key LIMIT 1),
	(SELECT create_user_6.data AS create_user_6_data FROM ` +
		`"user" create_user_6 WHERE create_user_6.create_user_key = ` +
		`test.create_user_key LIMIT 1),
	EXTRACT(epoch FROM test.updated_at)::BIGINT AS test_updated_at,
	(SELECT update_user_8.user_id AS update_user_8_user_id FROM ` +
		`"user" update_user_8 WHERE update_user_8.user_key = ` +
		`test.updated_by LIMIT 1),
	(SELECT update_user_9.email AS update_user_9_email FROM ` +
		`"user" update_user_9 WHERE update_user_9.update_user_key = ` +
		`test.update_user_key LIMIT 1),
	(SELECT update_user_10.last_name AS update_user_10_last_name FROM ` +
		`"user" update_user_10 WHERE update_user_10.update_user_key = ` +
		`test.update_user_key LIMIT 1),
	(SELECT update_user_11.first_name AS update_user_11_first_name FROM ` +
		`"user" update_user_11 WHERE update_user_11.update_user_key = ` +
		`test.update_user_key LIMIT 1),
	(SELECT update_user_12.status AS update_user_12_status FROM ` +
		`"user" update_user_12 WHERE update_user_12.update_user_key = ` +
		`test.update_user_key LIMIT 1),
	(SELECT update_user_13.data AS update_user_13_data FROM ` +
		`"user" update_user_13 WHERE update_user_13.update_user_key = ` +
		`test.update_user_key LIMIT 1)` + "\n"

	if v != exp {
		t.Errorf("Expected: %v, got: %v", exp, v)
	}
}

type mockTx struct {
	db *mockDB
}

func (m *mockTx) Commit(ctx context.Context) error {
	return nil
}

func (m *mockTx) Exec(ctx context.Context,
	q string, args ...any,
) (sqldb.SQLResult, error) {
	return m.db.Exec(ctx, q, args...)
}

func (m *mockTx) Query(ctx context.Context,
	q string, args ...any,
) (sqldb.SQLRows, error) {
	return m.db.Query(ctx, q, args...)
}

func (m *mockTx) QueryRow(ctx context.Context,
	q string, args ...any,
) sqldb.SQLRow {
	return m.db.QueryRow(ctx, q, args...)
}

func (m *mockTx) Rollback(ctx context.Context) error {
	return nil
}

func (m *mockTx) CloseTx(ctx context.Context, err error) error {
	return nil
}

type mockDB struct{}

func (m *mockDB) BeginTx(ctx context.Context,
	opts pgx.TxOptions,
) (sqldb.SQLTX, error) {
	return &mockTx{db: m}, nil
}

func (m *mockDB) Close() {
	return
}

func (m *mockDB) Exec(ctx context.Context,
	q string, args ...any,
) (sqldb.SQLResult, error) {
	return &mockResult{}, nil
}

func (m *mockDB) Query(ctx context.Context,
	q string, args ...any,
) (sqldb.SQLRows, error) {
	switch {
	default:
		return nil, errors.New(errors.ErrDatabase, "invalid query")
	}
}

func (m *mockDB) QueryRow(ctx context.Context,
	q string, args ...any,
) sqldb.SQLRow {
	switch {
	case strings.Contains(q, `FROM "user"`+"\n"):
		return &mockUserDataRow{}
	default:
		return &mockRowError{}
	}
}

func (m *mockDB) Ping(ctx context.Context) error {
	return nil
}

func (m *mockDB) Stat() *pgxpool.Stat {
	return &pgxpool.Stat{}
}

type mockResult struct{}

func (m *mockResult) RowsAffected() int64 {
	return 1
}

type mockRowError struct{}

func (m *mockRowError) Scan(dest ...any) error {
	return errors.New(errors.ErrDatabase, "database error")
}

var testUserData = sqldb.UserData{
	UserID: request.FieldString{
		Set: true, Valid: true,
		Value: testID,
	},
	Email: request.FieldString{
		Set: true, Valid: true,
		Value: "test@test.com",
	},
	LastName: request.FieldString{
		Set: true, Valid: true,
		Value: "testLastName",
	},
	FirstName: request.FieldString{
		Set: true, Valid: true,
		Value: "testFirstName",
	},
	Status: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusActive,
	},
	Data: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"testData": "testData",
		},
	},
}

type mockUserDataRow struct{}

func (m *mockUserDataRow) Scan(dest ...any) error {
	n := 0

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = testUserData.UserID
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = testUserData.Email
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = testUserData.LastName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = testUserData.FirstName
		n++
	}

	if v, ok := dest[n].(*request.FieldString); ok {
		*v = testUserData.Status
		n++
	}

	if v, ok := dest[n].(*request.FieldJSON); ok {
		*v = testUserData.Data
	}

	return nil
}

func TestUserDataScanDest(t *testing.T) {
	t.Parallel()

	ud := sqldb.UserData{}

	res := ud.ScanDest()

	exp := 6

	if len(res) != exp {
		t.Errorf("Expected length: %v, got:, %v", exp, len(res))
	}
}

func TestGetUserDetails(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	res, err := sqldb.GetUserDetails(ctx, testID, &mockDB{}, &mc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if res.UserID.Value != testUserData.UserID.Value {
		t.Errorf("Expected user_id: %v, got: %v",
			testUserData.UserID.Value, res.UserID.Value)
	}

	if !mc.WasMissed() {
		t.Error("expected cache miss")
	}

	if !mc.WasSet() {
		t.Error("expected cache set")
	}

	res, err = sqldb.GetUserDetails(ctx, testID, &mockDB{}, &mc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !mc.WasHit() {
		t.Error("expected cache hit")
	}

	if res.UserID.Value != testUserData.UserID.Value {
		t.Errorf("Expected user_id: %v, got: %v",
			testUserData.UserID.Value, res.UserID.Value)
	}
}
