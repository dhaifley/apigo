package auth_test

import (
	"context"
	"testing"

	"github.com/dhaifley/apid/internal/auth"
	"github.com/dhaifley/apid/internal/cache"
	"github.com/dhaifley/apid/internal/request"
	"github.com/dhaifley/apid/tests/mocks"
)

func TestGetAccount(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := auth.NewService(nil, &mocks.MockAuthDB{}, &mc, nil, nil, nil)

	res, err := svc.GetAccount(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	if res.AccountID.Value != mocks.TestAccount.AccountID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestAccount.AccountID.Value, res.AccountID.Value)
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

	if res.AccountID.Value != mocks.TestAccount.AccountID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestAccount.AccountID.Value, res.AccountID.Value)
	}
}

func TestGetAccountByName(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := auth.NewService(nil, &mocks.MockAuthDB{}, &mc, nil, nil, nil)

	res, err := svc.GetAccountByName(ctx, mocks.TestAccount.Name.Value)
	if err != nil {
		t.Fatal(err)
	}

	if res.AccountID.Value != mocks.TestAccount.AccountID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestAccount.AccountID.Value, res.AccountID.Value)
	}

	if !mc.WasMissed() {
		t.Error("expected cache miss")
	}

	if !mc.WasSet() {
		t.Error("expected cache set")
	}

	res, err = svc.GetAccountByName(ctx, mocks.TestAccount.Name.Value)
	if err != nil {
		t.Fatal(err)
	}

	if !mc.WasHit() {
		t.Error("expected cache hit")
	}

	if res.AccountID.Value != mocks.TestAccount.AccountID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestAccount.AccountID.Value, res.AccountID.Value)
	}
}

func TestCreateAccount(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	ctx = context.WithValue(ctx, request.CtxKeyRoles,
		[]string{request.RoleSystemAdmin})

	mc := cache.MockCache{}

	svc := auth.NewService(nil, &mocks.MockAuthDB{}, &mc, nil, nil, nil)

	res, err := svc.CreateAccount(ctx, &mocks.TestAccount)
	if err != nil {
		t.Fatal(err)
	}

	if res.AccountID.Value != mocks.TestAccount.AccountID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestAccount.AccountID.Value, res.AccountID.Value)
	}

	exp := "test"

	if v, ok := res.Data.Value["test"]; !ok || v != exp {
		t.Errorf("Expected data value: %v, got: %v", exp, v)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}
}

func TestGetAccountRepo(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := auth.NewService(nil, &mocks.MockAuthDB{}, &mc, nil, nil, nil)

	res, err := svc.GetAccountRepo(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if res.Repo.Value != mocks.TestAccount.Repo.Value {
		t.Errorf("Expected repo: %v, got: %v",
			mocks.TestAccount.Repo.Value, res)
	}
}

func TestSetAccountRepo(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := auth.NewService(nil, &mocks.MockAuthDB{}, &mc, nil, nil, nil)

	if err := svc.SetAccountRepo(ctx, &auth.AccountRepo{
		Repo: request.FieldString{Set: true, Valid: true, Value: "test"},
	}); err != nil {
		t.Fatal(err)
	}
}
