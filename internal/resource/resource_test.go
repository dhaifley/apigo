package resource_test

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dhaifley/apigo/internal/auth"
	"github.com/dhaifley/apigo/internal/cache"
	"github.com/dhaifley/apigo/internal/errors"
	"github.com/dhaifley/apigo/internal/repo"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/resource"
	"github.com/dhaifley/apigo/internal/search"
	"github.com/dhaifley/apigo/internal/sqldb"
	"github.com/pashagolub/pgxmock/v4"
	"gopkg.in/yaml.v3"
)

const (
	TestKey  = int64(1)
	TestID   = "1"
	TestUUID = "11223344-5566-7788-9900-aabbccddeeff"
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
	CreatedBy: request.FieldString{
		Set: true, Valid: true,
		Value: TestID,
	},
	CreatedAt: request.FieldTime{
		Set: true, Valid: true,
		Value: 1,
	},
	UpdatedBy: request.FieldString{
		Set: true, Valid: true,
		Value: TestID,
	},
	UpdatedAt: request.FieldTime{
		Set: true, Valid: true,
		Value: 1,
	},
}

func mockAuthContext() context.Context {
	ctx := context.Background()

	ctx = context.WithValue(ctx, request.CtxKeyAccountID, TestID)

	ctx = context.WithValue(ctx, request.CtxKeyUserID, TestID)

	ctx = context.WithValue(ctx, request.CtxKeyScopes, strings.Join([]string{
		request.ScopeAccountRead,
		request.ScopeUserRead,
		request.ScopeResourceRead,
		request.ScopeResourceWrite,
	}, " "))

	return ctx
}

func mockAdminAuthContext() context.Context {
	ctx := context.Background()

	ctx = context.WithValue(ctx, request.CtxKeyAccountID, TestID)

	ctx = context.WithValue(ctx, request.CtxKeyUserID, TestID)

	ctx = context.WithValue(ctx, request.CtxKeyScopes, request.ScopeSuperUser)

	return ctx
}

type mockRepoClient struct{}

func (m *mockRepoClient) List(ctx context.Context, dirPath string,
) ([]repo.Item, error) {
	return []repo.Item{}, nil
}

func (m *mockRepoClient) ListAll(ctx context.Context, dirPath string,
) ([]repo.Item, error) {
	return []repo.Item{}, nil
}

func (m *mockRepoClient) Get(ctx context.Context, filePath string,
) ([]byte, error) {
	return yaml.Marshal(&TestResource)
}

func (m *mockRepoClient) Commit(ctx context.Context) (string, error) {
	return "test", nil
}

type mockAuthSvc struct {
	v *auth.AccountRepo
}

func (m *mockAuthSvc) GetAccountRepo(ctx context.Context,
) (*auth.AccountRepo, error) {
	if m.v != nil {
		return m.v, nil
	}

	return &auth.AccountRepo{
		Repo: request.FieldString{
			Set: true, Valid: true,
			Value: "test://test:test@test/test/test#test",
		},
		RepoStatus: request.FieldString{
			Set: true, Valid: true,
			Value: request.StatusActive,
		},
	}, nil
}

func (m *mockAuthSvc) SetAccountRepo(ctx context.Context,
	v *auth.AccountRepo,
) error {
	m.v = v

	return nil
}

func mockTransaction(mock pgxmock.PgxCommonIface) {
	mock.ExpectBegin()

	mock.ExpectExec("SET app.account_id").
		WillReturnResult(pgxmock.NewResult("SET", 1))
}

func mockAccountCommitHashRows(mock pgxmock.PgxCommonIface) *pgxmock.Rows {
	return mock.NewRows([]string{"resource_commit_hash"}).
		AddRow(&[]string{"test"}[0])
}

func mockResourceKeyRows(mock pgxmock.PgxCommonIface) *pgxmock.Rows {
	return mock.NewRows([]string{"resource_key", "resource_id"}).
		AddRow(TestKey, TestResource.ResourceID.Value)
}

func mockResourceIDRows(mock pgxmock.PgxCommonIface) *pgxmock.Rows {
	return mock.NewRows([]string{"resource_id"}).
		AddRow(TestResource.ResourceID.Value)
}

func mockResourceSummaryRows(mock pgxmock.PgxCommonIface) *pgxmock.Rows {
	return mock.NewRows([]string{"status", "count"}).
		AddRow(request.StatusActive, int64(1))
}

func mockResourceRows(mock pgxmock.PgxCommonIface) *pgxmock.Rows {
	return mock.NewRows([]string{
		"resource_id",
		"name",
		"version",
		"description",
		"status",
		"status_data",
		"key_field",
		"key_regex",
		"clear_condition",
		"clear_after",
		"clear_delay",
		"data",
		"source",
		"commit_hash",
	}).AddRow(
		TestResource.ResourceID.Value,
		TestResource.Name.Value,
		TestResource.Version.Value,
		TestResource.Description.Value,
		TestResource.Status.Value,
		TestResource.StatusData.Value,
		TestResource.KeyField.Value,
		TestResource.KeyRegex.Value,
		TestResource.ClearCondition.Value,
		TestResource.ClearAfter.Value,
		TestResource.ClearDelay.Value,
		TestResource.Data.Value,
		TestResource.Source.Value,
		TestResource.CommitHash.Value,
	)
}

func TestGetResources(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, mc, nil, nil, nil)

	opts, err := sqldb.ParseFieldOptions(url.Values{
		"user_details": []string{"false"},
	})
	if err != nil {
		t.Fatal(err)
	}

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM resource").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockResourceKeyRows(mock))

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM resource").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockResourceRows(mock))

	res, _, err := svc.GetResources(ctx, &search.Query{
		Search: "and(name:*)",
		Size:   10,
		From:   0,
		Order:  "-name",
	}, opts)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) <= 0 {
		t.Errorf("Expected length to be greater than 0")
	}

	if res[0].ResourceID.Value != TestResource.ResourceID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestResource.ResourceID.Value, res[0].ResourceID.Value)
	}

	if !mc.WasMissed() {
		t.Error("expected cache miss")
	}

	if !mc.WasSet() {
		t.Error("expected cache set")
	}

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM resource").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockResourceKeyRows(mock))

	res, _, err = svc.GetResources(ctx, &search.Query{
		Search: "and(name:*)",
		Size:   10,
		From:   0,
		Order:  "-name",
	}, opts)
	if err != nil {
		t.Fatal(err)
	}

	if !mc.WasHit() {
		t.Error("expected cache hit")
	}

	if len(res) <= 0 {
		t.Fatal("Expected length to be greater than 0")
	}

	if res[0].ResourceID.Value != TestResource.ResourceID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestResource.ResourceID.Value, res[0].ResourceID.Value)
	}

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM resource").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockResourceKeyRows(mock))

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM resource").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockResourceSummaryRows(mock))

	_, sum, err := svc.GetResources(ctx, &search.Query{
		Search:  "and(status:*)",
		Size:    10,
		From:    0,
		Order:   "-status",
		Summary: "status",
	}, opts)
	if err != nil {
		t.Fatal(err)
	}

	if len(sum) <= 0 {
		t.Errorf("Expected summary to be greater than 0")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestGetResource(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM resource").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockResourceRows(mock))

	res, err := svc.GetResource(ctx, TestResource.ResourceID.Value, nil)
	if err != nil {
		t.Fatal(err)
	}

	if res.ResourceID.Value != TestResource.ResourceID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestResource.ResourceID.Value, res.ResourceID.Value)
	}

	if !mc.WasMissed() {
		t.Error("expected cache miss")
	}

	if !mc.WasSet() {
		t.Error("expected cache set")
	}

	res, err = svc.GetResource(ctx, TestResource.ResourceID.Value, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !mc.WasHit() {
		t.Error("expected cache hit")
	}

	if res.ResourceID.Value != TestResource.ResourceID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestResource.ResourceID.Value, res.ResourceID.Value)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestCreateResource(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	args := make([]any, 16)

	for i := 0; i < 16; i++ {
		args[i] = pgxmock.AnyArg()
	}

	mock.ExpectQuery("INSERT INTO resource").
		WithArgs(args...).WillReturnRows(mockResourceRows(mock))

	res, err := svc.CreateResource(ctx, &TestResource)
	if err != nil {
		t.Fatal(err)
	}

	if res.ResourceID.Value != TestResource.ResourceID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestResource.ResourceID.Value, res.ResourceID.Value)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestUpdateResource(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	args := make([]any, 16)

	for i := 0; i < 16; i++ {
		args[i] = pgxmock.AnyArg()
	}

	mock.ExpectQuery("UPDATE resource").
		WithArgs(args...).WillReturnRows(mockResourceRows(mock))

	res, err := svc.UpdateResource(ctx, &TestResource)
	if err != nil {
		t.Fatal(err)
	}

	if res.ResourceID.Value != TestResource.ResourceID.Value {
		t.Errorf("Expected id: %v, got: %v",
			TestResource.ResourceID.Value, res.ResourceID.Value)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestDeleteResource(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	args := make([]any, 16)

	for i := 0; i < 16; i++ {
		args[i] = pgxmock.AnyArg()
	}

	mock.ExpectExec("DELETE FROM resource").
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	if err := svc.DeleteResource(ctx, TestUUID); err != nil {
		t.Fatal(err)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestUpdateResourceData(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM resource").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockResourceRows(mock))

	mockTransaction(mock)

	args := make([]any, 16)

	for i := 0; i < 16; i++ {
		args[i] = pgxmock.AnyArg()
	}

	mock.ExpectQuery("UPDATE resource").
		WithArgs(args...).WillReturnRows(mockResourceRows(mock))

	res, err := svc.UpdateResourceData(ctx, map[string]any{
		"resources": []any{
			map[string]any{
				"resource_id": TestUUID,
				"account_id":  TestUUID,
				"cleared_on":  int64(1),
			},
		},
	}, TestID, TestResource.ResourceID.Value)
	if err != nil {
		t.Fatal(err)
	}

	if res.ResourceID.Value != TestResource.ResourceID.Value {
		t.Errorf("Expected resource_id: %v, got: %v",
			TestResource.ResourceID, res.ResourceID.Value)
	}

	if _, ok := res.Data.Value[TestUUID]; !ok {
		t.Errorf("Expected resource data to contain key: %v, got: %v",
			TestUUID, res.Data.Value)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestUpdateResourceError(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := &cache.MockCache{}

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, mc, nil, nil, nil)

	mockTransaction(mock)

	args := make([]any, 5)

	for i := 0; i < 5; i++ {
		args[i] = pgxmock.AnyArg()
	}

	mock.ExpectQuery("UPDATE resource").
		WithArgs(args...).WillReturnRows(mockResourceRows(mock))

	if err := svc.UpdateResourceError(ctx, TestUUID, TestUUID,
		errors.New(errors.ErrServer, "test error")); err != nil {
		t.Fatal(err)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestImportResources(t *testing.T) {
	t.Parallel()

	ctx := mockAdminAuthContext()

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, nil, nil, nil, nil)

	svc.SetRepoClient(&mockRepoClient{})

	ma := &mockAuthSvc{}

	mockTransaction(mock)

	mock.ExpectQuery("SELECT resource_commit_hash FROM account").
		WillReturnRows(mockAccountCommitHashRows(mock))

	mockTransaction(mock)

	mock.ExpectQuery("UPDATE account SET resource_commit_hash").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(mockAccountCommitHashRows(mock))

	mockTransaction(mock)

	mock.ExpectQuery("DELETE FROM resource").WithArgs(pgxmock.AnyArg()).
		WillReturnRows(mockResourceIDRows(mock))

	if err := svc.ImportResources(ctx, true, ma); err != nil {
		t.Fatal(err)
	}

	if ma.v.RepoStatus.Value != request.StatusActive {
		t.Errorf("Expected repo status: %v, got: %v",
			request.StatusActive, ma.v.RepoStatus.Value)
	}

	expData := []string{
		"resources_last_imported",
		"resources_deleted",
		"resources_updated",
	}

	for _, expF := range expData {
		if _, ok := ma.v.RepoStatusData.Value[expF]; !ok {
			t.Errorf("Expected repo status data field: %v", expF)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestImportResource(t *testing.T) {
	t.Parallel()

	ctx := mockAdminAuthContext()

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, nil, nil, nil, nil)

	svc.SetRepoClient(&mockRepoClient{})

	mockTransaction(mock)

	args := make([]any, 16)

	for i := 0; i < 16; i++ {
		args[i] = pgxmock.AnyArg()
	}

	mock.ExpectQuery("INSERT INTO resource").
		WithArgs(args...).WillReturnRows(mockResourceRows(mock))

	if err := svc.ImportResource(ctx, &mockAuthSvc{}, TestUUID); err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := resource.NewService(nil, md, nil, nil, nil, nil)

	cancel := svc.Update(ctx, &mockAuthSvc{})

	time.Sleep(time.Second)

	cancel()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
	}
}
