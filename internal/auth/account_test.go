package auth_test

import (
	"context"
	"testing"

	"github.com/dhaifley/apigo/internal/auth"
	"github.com/dhaifley/apigo/internal/cache"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/sqldb"
	"github.com/pashagolub/pgxmock/v4"
)

var TestAccount = auth.Account{
	AccountID: request.FieldString{
		Set: true, Valid: true,
		Value: TestID,
	},
	Name: request.FieldString{
		Set: true, Valid: true,
		Value: "testAccount",
	},
	Status: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusActive,
	},
	StatusData: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"last_error": "test",
		},
	},
	Repo: request.FieldString{
		Set: true, Valid: true,
		Value: "test",
	},
	RepoStatus: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusActive,
	},
	RepoStatusData: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"last_error": "test",
		},
	},
	Secret: request.FieldString{
		Set: true, Valid: true,
		Value: "test",
	},
	Data: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"test": "test",
		},
	},
}

func mockTransaction(mock pgxmock.PgxCommonIface) {
	mock.ExpectBegin()

	mock.ExpectExec("SET app.account_id").
		WillReturnResult(pgxmock.NewResult("SET", 1))
}

func mockAccountRows(mock pgxmock.PgxCommonIface) *pgxmock.Rows {
	return mock.NewRows([]string{
		"account_id",
		"name",
		"status",
		"status_data",
		"repo",
		"repo_status",
		"repo_status_data",
		"secret",
		"data",
		"created_at",
		"updated_at",
	}).AddRow(
		TestAccount.AccountID.Value,
		TestAccount.Name.Value,
		TestAccount.Status.Value,
		TestAccount.StatusData.Value,
		TestAccount.Repo.Value,
		TestAccount.RepoStatus.Value,
		TestAccount.RepoStatusData.Value,
		TestAccount.Secret.Value,
		TestAccount.Data.Value,
		TestAccount.CreatedAt.Value,
		TestAccount.UpdatedAt.Value,
	)
}

func mockAccountRepoRows(mock pgxmock.PgxCommonIface) *pgxmock.Rows {
	return mock.NewRows([]string{
		"repo",
		"repo_status",
		"repo_status_data",
	}).AddRow(
		TestAccount.Repo.Value,
		TestAccount.RepoStatus.Value,
		TestAccount.RepoStatusData.Value,
	)
}

func TestGetAccount(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := auth.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM account").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockAccountRows(mock))

	res, err := svc.GetAccount(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	if res.AccountID.Value != TestAccount.AccountID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestAccount.AccountID.Value, res.AccountID.Value)
	}

	if !mc.WasMissed() {
		t.Error("expected cache miss")
	}

	if !mc.WasSet() {
		t.Error("expected cache set")
	}

	res, err = svc.GetAccount(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	if !mc.WasHit() {
		t.Error("expected cache hit")
	}

	if res.AccountID.Value != TestAccount.AccountID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestAccount.AccountID.Value, res.AccountID.Value)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestGetAccountByName(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := auth.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM account").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockAccountRows(mock))

	res, err := svc.GetAccountByName(ctx, TestAccount.Name.Value)
	if err != nil {
		t.Fatal(err)
	}

	if res.AccountID.Value != TestAccount.AccountID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestAccount.AccountID.Value, res.AccountID.Value)
	}

	if !mc.WasMissed() {
		t.Error("expected cache miss")
	}

	if !mc.WasSet() {
		t.Error("expected cache set")
	}

	res, err = svc.GetAccountByName(ctx, TestAccount.Name.Value)
	if err != nil {
		t.Fatal(err)
	}

	if !mc.WasHit() {
		t.Error("expected cache hit")
	}

	if res.AccountID.Value != TestAccount.AccountID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestAccount.AccountID.Value, res.AccountID.Value)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestCreateAccount(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	ctx = context.WithValue(ctx, request.CtxKeyScopes, request.ScopeSuperUser)

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

	mock.ExpectQuery("INSERT INTO account").
		WithArgs(args...).WillReturnRows(mockAccountRows(mock))

	res, err := svc.CreateAccount(ctx, &TestAccount)
	if err != nil {
		t.Fatal(err)
	}

	if res.AccountID.Value != TestAccount.AccountID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestAccount.AccountID.Value, res.AccountID.Value)
	}

	exp := "test"

	if v, ok := res.Data.Value["test"]; !ok || v != exp {
		t.Errorf("Expected data value: %v, got: %v", exp, v)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestGetAccountRepo(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := auth.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	mock.ExpectQuery("SELECT(.+)account\\.repo(.*)FROM account").
		WillReturnRows(mockAccountRepoRows(mock))

	res, err := svc.GetAccountRepo(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if res.Repo.Value != TestAccount.Repo.Value {
		t.Errorf("Expected repo: %v, got: %v",
			TestAccount.Repo.Value, res)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestSetAccountRepo(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := auth.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	mock.ExpectQuery("UPDATE account SET").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(mockAccountRepoRows(mock))

	if err := svc.SetAccountRepo(ctx, &auth.AccountRepo{
		Repo: request.FieldString{Set: true, Valid: true, Value: "test"},
	}); err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}
