package auth_test

import (
	"net/url"
	"testing"

	"github.com/dhaifley/apid/internal/auth"
	"github.com/dhaifley/apid/internal/cache"
	"github.com/dhaifley/apid/internal/sqldb"
	"github.com/dhaifley/apid/tests/mocks"
)

func TestGetUser(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := auth.NewService(nil, &mocks.MockAuthDB{}, &mc, nil, nil, nil)

	opts, err := sqldb.ParseFieldOptions(url.Values{
		"user_details": []string{"true"},
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := svc.GetUser(ctx, "", opts)
	if err != nil {
		t.Fatal(err)
	}

	if res.UserID.Value != mocks.TestUser.UserID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestUser.UserID.Value, res.UserID.Value)
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

	if res.UserID.Value != mocks.TestUser.UserID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestUser.UserID.Value, res.UserID.Value)
	}
}

func TestCreateUser(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := auth.NewService(nil, &mocks.MockAuthDB{}, &mc, nil, nil, nil)

	res, err := svc.CreateUser(ctx, &mocks.TestUser)
	if err != nil {
		t.Fatal(err)
	}

	if res.UserID.Value != mocks.TestUser.UserID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestUser.UserID.Value, res.UserID.Value)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}
}

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := auth.NewService(nil, &mocks.MockAuthDB{}, &mc, nil, nil, nil)

	res, err := svc.UpdateUser(ctx, &mocks.TestUser)
	if err != nil {
		t.Fatal(err)
	}

	if res.UserID.Value != mocks.TestUser.UserID.Value {
		t.Errorf("Expected id: %v, got: %v",
			mocks.TestUser.UserID.Value, res.UserID.Value)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}
}

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := auth.NewService(nil, &mocks.MockAuthDB{}, &mc, nil, nil, nil)

	if err := svc.DeleteUser(ctx, mocks.TestUUID); err != nil {
		t.Fatal(err)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}
}
