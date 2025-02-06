package auth_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/dhaifley/apid/internal/auth"
	"github.com/dhaifley/apid/internal/cache"
	"github.com/dhaifley/apid/internal/request"
	"github.com/dhaifley/apid/internal/search"
	"github.com/dhaifley/apid/internal/sqldb"
	"github.com/dhaifley/apid/tests/mocks"
)

func TestGetTokens(t *testing.T) {
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

	res, _, err := svc.GetTokens(ctx, &search.Query{
		Search: "and(token_id:*)",
		Size:   10,
		From:   0,
		Order:  "token_id",
	}, opts)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) <= 0 {
		t.Fatalf("Expected results length: >0, got: %v", len(res))
	}

	if res[0].TokenID.Value != mocks.TestToken.TokenID.Value {
		t.Errorf("Expected token: %v, got: %v",
			mocks.TestToken.TokenID.Value, res[0].TokenID.Value)
	}

	if !mc.WasMissed() {
		t.Error("expected cache miss")
	}

	if !mc.WasSet() {
		t.Error("expected cache set")
	}

	res, _, err = svc.GetTokens(ctx, &search.Query{
		Search: "and(token_id:*)",
		Size:   10,
		From:   0,
		Order:  "token_id",
	}, opts)
	if err != nil {
		t.Fatal(err)
	}

	if !mc.WasHit() {
		t.Error("expected cache hit")
	}

	if len(res) <= 0 {
		t.Fatalf("Expected results length: >0, got: %v", len(res))
	}

	if res[0].TokenID.Value != mocks.TestToken.TokenID.Value {
		t.Errorf("Expected token: %v, got: %v",
			mocks.TestToken.TokenID.Value, res[0].TokenID.Value)
	}

	_, sum, err := svc.GetTokens(ctx, &search.Query{
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

func TestGetToken(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := auth.NewService(nil, &mocks.MockAuthDB{}, &mc, nil, nil, nil)

	res, err := svc.GetToken(ctx, mocks.TestToken.TokenID.Value,
		sqldb.FieldOptions{sqldb.OptUserDetails})
	if err != nil {
		t.Fatal(err)
	}

	if res.TokenID.Value != mocks.TestToken.TokenID.Value {
		t.Errorf("Expected token: %v, got: %v",
			mocks.TestToken.TokenID.Value, res.TokenID.Value)
	}

	if !mc.WasMissed() {
		t.Error("expected cache miss")
	}

	if !mc.WasSet() {
		t.Error("expected cache set")
	}

	res, err = svc.GetToken(ctx, mocks.TestToken.TokenID.Value,
		sqldb.FieldOptions{sqldb.OptUserDetails})
	if err != nil {
		t.Fatal(err)
	}

	if !mc.WasHit() {
		t.Error("expected cache hit")
	}

	if res.TokenID.Value != mocks.TestToken.TokenID.Value {
		t.Errorf("Expected token: %v, got: %v",
			mocks.TestToken.TokenID.Value, res.TokenID.Value)
	}
}

func TestCreateToken(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := auth.NewService(nil, &mocks.MockAuthDB{}, &mc,
		nil, nil, nil)

	res, err := svc.CreateToken(ctx, &auth.Token{
		Status: request.FieldString{
			Set: true, Valid: true,
			Value: request.StatusActive,
		},
		Expiration: request.FieldTime{
			Set: true, Valid: true,
			Value: time.Now().Unix() + 100,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if res.TokenID.Value != mocks.TestToken.TokenID.Value {
		t.Errorf("Expected token: %v, got: %v",
			mocks.TestToken.TokenID.Value, res.TokenID.Value)
	}

	if res.Secret == nil || len(*res.Secret) < 14 {
		t.Errorf("Expected secret, got: %v", res.Secret)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}
}

func TestDeleteToken(t *testing.T) {
	t.Parallel()

	ctx := mockAuthContext()

	mc := cache.MockCache{}

	svc := auth.NewService(nil, &mocks.MockAuthDB{}, &mc,
		nil, nil, nil)

	if err := svc.DeleteToken(ctx,
		mocks.TestToken.TokenID.Value); err != nil {
		t.Fatal(err)
	}

	if !mc.WasDeleted() {
		t.Error("expected cache delete")
	}
}
