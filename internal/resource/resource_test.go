package resource_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/dhaifley/apid/internal/auth"
	"github.com/dhaifley/apid/internal/cache"
	"github.com/dhaifley/apid/internal/errors"
	"github.com/dhaifley/apid/internal/repo"
	"github.com/dhaifley/apid/internal/request"
	"github.com/dhaifley/apid/internal/resource"
	"github.com/dhaifley/apid/internal/search"
	"github.com/dhaifley/apid/internal/sqldb"
	"github.com/dhaifley/apid/tests/mocks"
	"gopkg.in/yaml.v3"
)

func mockAuthContext() context.Context {
	ctx := context.Background()

	ctx = context.WithValue(ctx, request.CtxKeyAccountID, mocks.TestID)

	ctx = context.WithValue(ctx, request.CtxKeyUserID, mocks.TestUUID)

	ctx = context.WithValue(ctx, request.CtxKeyRoles, []string{
		request.RoleUser,
	})

	return ctx
}

func mockAdminAuthContext() context.Context {
	ctx := context.Background()

	ctx = context.WithValue(ctx, request.CtxKeyAccountID, mocks.TestID)

	ctx = context.WithValue(ctx, request.CtxKeyUserID, mocks.TestUUID)

	ctx = context.WithValue(ctx, request.CtxKeyRoles, []string{
		request.RoleAdmin,
	})

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
	return yaml.Marshal(&mocks.TestResource)
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

func TestGetResources(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := resource.NewService(nil, &mocks.MockResourceDB{}, &mc, nil, nil, nil)

	opts, err := sqldb.ParseFieldOptions(url.Values{
		"user_details": []string{"true"},
	})
	if err != nil {
		t.Fatal(err)
	}

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

	if res[0].ResourceID.Value != mocks.TestResource.ResourceID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestResource.ResourceID.Value, res[0].ResourceID.Value)
	}

	if !mc.WasMissed() {
		t.Error("expected cache miss")
	}

	if !mc.WasSet() {
		t.Error("expected cache set")
	}

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
		t.Errorf("Expected length to be greater than 0")
	}

	if res[0].ResourceID.Value != mocks.TestResource.ResourceID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestResource.ResourceID.Value, res[0].ResourceID.Value)
	}

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
}

func TestGetResource(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := resource.NewService(nil, &mocks.MockResourceDB{}, &mc, nil, nil, nil)

	res, err := svc.GetResource(ctx, mocks.TestResource.ResourceID.Value,
		sqldb.FieldOptions{sqldb.OptUserDetails})
	if err != nil {
		t.Fatal(err)
	}

	if res.ResourceID.Value != mocks.TestResource.ResourceID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestResource.ResourceID.Value, res.ResourceID.Value)
	}

	if !mc.WasMissed() {
		t.Error("expected cache miss")
	}

	if !mc.WasSet() {
		t.Error("expected cache set")
	}

	res, err = svc.GetResource(ctx, mocks.TestResource.ResourceID.Value,
		sqldb.FieldOptions{sqldb.OptUserDetails})
	if err != nil {
		t.Fatal(err)
	}

	if !mc.WasHit() {
		t.Error("expected cache hit")
	}

	if res.ResourceID.Value != mocks.TestResource.ResourceID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestResource.ResourceID.Value, res.ResourceID.Value)
	}
}

func TestCreateResource(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := resource.NewService(nil, &mocks.MockResourceDB{}, &mc, nil, nil, nil)

	res, err := svc.CreateResource(ctx, &mocks.TestResource)
	if err != nil {
		t.Fatal(err)
	}

	if res.ResourceID.Value != mocks.TestResource.ResourceID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestResource.ResourceID.Value, res.ResourceID.Value)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}
}

func TestUpdateResource(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := resource.NewService(nil, &mocks.MockResourceDB{}, &mc, nil, nil, nil)

	res, err := svc.UpdateResource(ctx, &mocks.TestResource)
	if err != nil {
		t.Fatal(err)
	}

	if res.ResourceID.Value != mocks.TestResource.ResourceID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestResource.ResourceID.Value, res.ResourceID.Value)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}
}

func TestDeleteResource(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := resource.NewService(nil, &mocks.MockResourceDB{}, &mc, nil, nil, nil)

	if err := svc.DeleteResource(ctx, mocks.TestUUID); err != nil {
		t.Fatal(err)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}
}

func TestUpdateResourceData(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := resource.NewService(nil, &mocks.MockResourceDB{}, &mc, nil, nil, nil)

	res, err := svc.UpdateResourceData(ctx, map[string]any{
		"resources": []any{
			map[string]any{
				"resource_id": mocks.TestUUID,
				"account_id":  mocks.TestUUID,
				"cleared_on":  int64(1),
			},
		},
	}, mocks.TestID, mocks.TestResource.ResourceID.Value)
	if err != nil {
		t.Fatal(err)
	}

	if res.ResourceID.Value != mocks.TestResource.ResourceID.Value {
		t.Errorf("Expected resource_id: %v, got: %v",
			mocks.TestResource.ResourceID, res.ResourceID.Value)
	}

	if _, ok := res.Data.Value[mocks.TestUUID]; !ok {
		t.Errorf("Expected resource data to contain key: %v, got: %v",
			mocks.TestUUID, res.Data.Value)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}
}

func TestUpdateResourceError(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := resource.NewService(nil, &mocks.MockResourceDB{}, &mc, nil, nil, nil)

	if err := svc.UpdateResourceError(ctx, mocks.TestID, mocks.TestUUID,
		errors.New(errors.ErrServer, "test error")); err != nil {
		t.Fatal(err)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}
}

func TestImportResources(t *testing.T) {
	t.Parallel()

	ctx := mockAdminAuthContext()

	svc := resource.NewService(nil, &mocks.MockResourceDB{},
		nil, nil, nil, nil)

	svc.SetRepoClient(&mockRepoClient{})

	ma := &mockAuthSvc{}

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
}

func TestImportResource(t *testing.T) {
	t.Parallel()

	ctx := mockAdminAuthContext()

	svc := resource.NewService(nil, &mocks.MockResourceDB{},
		nil, nil, nil, nil)

	svc.SetRepoClient(&mockRepoClient{})

	if err := svc.ImportResource(ctx, &mockAuthSvc{},
		mocks.TestUUID); err != nil {
		t.Fatal(err)
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	svc := resource.NewService(nil, &mocks.MockResourceDB{}, nil, nil, nil, nil)

	cancel := svc.Update(ctx, &mockAuthSvc{})

	time.Sleep(time.Second)

	cancel()
}
