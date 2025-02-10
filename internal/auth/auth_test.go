package auth_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dhaifley/apigo/internal/auth"
	"github.com/dhaifley/apigo/internal/config"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/tests/mocks"
	"github.com/golang-jwt/jwt/v5"
)

func mockAuthContext() context.Context {
	ctx := context.Background()

	ctx = context.WithValue(ctx, request.CtxKeyAccountID, mocks.TestID)

	ctx = context.WithValue(ctx, request.CtxKeyAccountName, mocks.TestName)

	ctx = context.WithValue(ctx, request.CtxKeyUserID, mocks.TestUUID)

	ctx = context.WithValue(ctx, request.CtxKeyRoles, []string{
		request.RoleSystemAdmin,
	})

	return ctx
}

func TestAuthJWT(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cfg := config.NewDefault()

	svc := auth.NewService(cfg, &mocks.MockAuthDB{}, nil, nil, nil, nil)

	now := time.Now()

	expr := now.Add(cfg.AuthTokenExpiresIn())

	claims := jwt.MapClaims{
		"exp":   expr.Unix(),
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
		"iss":   cfg.AuthTokenIssuer(),
		"sub":   mocks.TestUser.UserID.Value,
		"aud":   []string{cfg.ServiceName()},
		"email": mocks.TestUser.Email.Value,
		"role":  request.RoleAdmin,
	}

	signMethod := jwt.SigningMethodHS512

	signKey := []byte(mocks.TestUUID)

	tok := jwt.NewWithClaims(signMethod, claims)

	tok.Header = map[string]any{
		"alg": "HS512",
		"kid": mocks.TestID,
	}

	authToken, err := tok.SignedString(signKey)
	if err != nil {
		t.Fatal(err)
	}

	c, err := svc.AuthJWT(ctx, authToken, "")
	if err != nil {
		t.Fatal(err)
	}

	if c.UserID != mocks.TestUser.UserID.Value {
		t.Errorf("Expected claim user_id: %v, got: %v",
			mocks.TestUser.UserID.Value, c.UserID)
	}

	claims = jwt.MapClaims{
		"exp":      expr.Unix(),
		"iat":      now.Unix(),
		"nbf":      now.Unix(),
		"iss":      cfg.AuthTokenIssuer(),
		"sub":      mocks.TestUUID,
		"aud":      []string{cfg.ServiceName()},
		"role":     request.RoleRefresh,
		"token_id": mocks.TestUUID,
	}

	tok = jwt.NewWithClaims(signMethod, claims)

	tok.Header = map[string]any{
		"alg": "HS512",
		"kid": mocks.TestID,
	}

	authToken, err = tok.SignedString(signKey)
	if err != nil {
		t.Fatal(err)
	}

	c, err = svc.AuthJWT(ctx, authToken, "")
	if err != nil {
		t.Fatal(err)
	}
}

// TestAuthCreateJWT is used to test creation if a JWT.
func TestAuthCreateJWT(t *testing.T) {
	now := time.Now()

	claims := jwt.MapClaims{
		"exp":  now.AddDate(1, 0, 0).Unix(),
		"iat":  now.Unix(),
		"nbf":  now.Unix(),
		"iss":  "api",
		"sub":  "0",
		"aud":  "api",
		"role": request.RoleAdmin,
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)

	pkb, err := os.ReadFile("../../certs/tls.key")
	if err != nil || len(pkb) == 0 {
		t.Error("invalid key", string(pkb))
		t.Error(err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(pkb)
	if err != nil {
		t.Error(err)
	}

	authToken, err := tok.SignedString(key)
	if err != nil {
		t.Error(err)
	}

	t.Log(authToken)
}
