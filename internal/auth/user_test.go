package auth_test

import (
	"net/url"
	"testing"

	"github.com/dhaifley/apigo/internal/auth"
	"github.com/dhaifley/apigo/internal/cache"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/sqldb"
	"github.com/pashagolub/pgxmock/v4"
)

var TestUser = auth.User{
	UserID: request.FieldString{
		Set: true, Valid: true,
		Value: TestUUID,
	},
	Email: request.FieldString{
		Set: true, Valid: true,
		Value: "test@apigo.io",
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
	Scopes: request.FieldString{
		Set: true, Valid: true,
		Value: request.ScopeSuperUser,
	},
	Data: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"test": "test",
		},
	},
}

func mockUserRows(mock pgxmock.PgxCommonIface) *pgxmock.Rows {
	return mock.NewRows([]string{
		"user_id",
		"email",
		"last_name",
		"first_name",
		"status",
		"scopes",
		"data",
	}).AddRow(
		TestUser.UserID.Value,
		TestUser.Email.Value,
		TestUser.LastName.Value,
		TestUser.FirstName.Value,
		TestUser.Status.Value,
		TestUser.Scopes.Value,
		TestUser.Data.Value,
	)
}

func TestGetUser(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := auth.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	mock.ExpectQuery(`SELECT (.+) FROM "user"`).
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockUserRows(mock))

	opts, err := sqldb.ParseFieldOptions(url.Values{
		"user_details": []string{"false"},
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := svc.GetUser(ctx, "", opts)
	if err != nil {
		t.Fatal(err)
	}

	if res.UserID.Value != TestUser.UserID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestUser.UserID.Value, res.UserID.Value)
	}

	if !mc.WasMissed() {
		t.Error("expected cache miss")
	}

	if !mc.WasSet() {
		t.Error("expected cache set")
	}

	res, err = svc.GetUser(ctx, "", opts)
	if err != nil {
		t.Fatal(err)
	}

	if !mc.WasHit() {
		t.Error("expected cache hit")
	}

	if res.UserID.Value != TestUser.UserID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestUser.UserID.Value, res.UserID.Value)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestCreateUser(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := auth.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	args := make([]any, 9)

	for i := 0; i < 9; i++ {
		args[i] = pgxmock.AnyArg()
	}

	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs(args...).WillReturnRows(mockUserRows(mock))

	res, err := svc.CreateUser(ctx, &TestUser)
	if err != nil {
		t.Fatal(err)
	}

	if res.UserID.Value != TestUser.UserID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestUser.UserID.Value, res.UserID.Value)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := auth.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	args := make([]any, 9)

	for i := 0; i < 9; i++ {
		args[i] = pgxmock.AnyArg()
	}

	mock.ExpectQuery(`UPDATE "user" SET`).
		WithArgs(args...).WillReturnRows(mockUserRows(mock))

	res, err := svc.UpdateUser(ctx, &TestUser)
	if err != nil {
		t.Fatal(err)
	}

	if res.UserID.Value != TestUser.UserID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestUser.UserID.Value, res.UserID.Value)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := auth.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	mock.ExpectExec(`DELETE FROM "user"`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	if err := svc.DeleteUser(ctx, TestUser.UserID.Value); err != nil {
		t.Fatal(err)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}
