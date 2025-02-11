package auth_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dhaifley/apigo/internal/auth"
	"github.com/dhaifley/apigo/internal/config"
	"github.com/dhaifley/apigo/internal/logger"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/sqldb"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pashagolub/pgxmock/v4"
)

func mockAccountSecretRows(mock pgxmock.PgxCommonIface) *pgxmock.Rows {
	return mock.NewRows([]string{
		"secret",
	}).AddRow(
		&TestAccount.Secret.Value,
	)
}

func mockAuthContext() context.Context {
	ctx := context.Background()

	ctx = context.WithValue(ctx, request.CtxKeyAccountID, TestID)

	ctx = context.WithValue(ctx, request.CtxKeyAccountName, TestName)

	ctx = context.WithValue(ctx, request.CtxKeyUserID, TestUUID)

	ctx = context.WithValue(ctx, request.CtxKeyRoles, []string{
		request.RoleSystemAdmin,
	})

	return ctx
}

func TestAuthJWT(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cfg := config.NewDefault()

	md, mock, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc := auth.NewService(cfg, md, nil, logger.New(logger.LogOutStdout, logger.LogFmtJSON, logger.LvlDebug), nil, nil)

	now := time.Now()

	expr := now.Add(cfg.AuthTokenExpiresIn())

	claims := jwt.MapClaims{
		"exp":   expr.Unix(),
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
		"iss":   cfg.AuthTokenIssuer(),
		"sub":   TestUser.UserID.Value,
		"aud":   []string{cfg.ServiceName()},
		"email": TestUser.Email.Value,
		"role":  request.RoleAdmin,
	}

	signMethod := jwt.SigningMethodHS512

	signKey := []byte(TestAccount.Secret.Value)

	tok := jwt.NewWithClaims(signMethod, claims)

	tok.Header = map[string]any{
		"alg": "HS512",
		"kid": TestID,
	}

	authToken, err := tok.SignedString(signKey)
	if err != nil {
		t.Fatal(err)
	}

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM account").
		WillReturnRows(mockAccountSecretRows(mock))

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM account").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockAccountRows(mock))

	mockTransaction(mock)

	mock.ExpectQuery(`SELECT (.+) FROM "user"`).
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockUserRows(mock))

	mockTransaction(mock)

	args := make([]any, 4)

	for i := 0; i < 4; i++ {
		args[i] = pgxmock.AnyArg()
	}

	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs(args...).WillReturnRows(mockUserRows(mock))

	c, err := svc.AuthJWT(ctx, authToken, "")
	if err != nil {
		t.Fatal(err)
	}

	if c.UserID != TestUser.UserID.Value {
		t.Errorf("Expected claim user_id: %v, got: %v",
			TestUser.UserID.Value, c.UserID)
	}

	claims = jwt.MapClaims{
		"exp":      expr.Unix(),
		"iat":      now.Unix(),
		"nbf":      now.Unix(),
		"iss":      cfg.AuthTokenIssuer(),
		"sub":      TestUUID,
		"aud":      []string{cfg.ServiceName()},
		"role":     request.RoleRefresh,
		"token_id": TestUUID,
	}

	tok = jwt.NewWithClaims(signMethod, claims)

	tok.Header = map[string]any{
		"alg": "HS512",
		"kid": TestID,
	}

	authToken, err = tok.SignedString(signKey)
	if err != nil {
		t.Fatal(err)
	}

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM account").
		WillReturnRows(mockAccountSecretRows(mock))

	mockTransaction(mock)

	mock.ExpectQuery("SELECT (.+) FROM account").
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockAccountRows(mock))

	mockTransaction(mock)

	mock.ExpectQuery(`SELECT (.+) FROM "user"`).
		WithArgs(pgxmock.AnyArg()).WillReturnRows(mockUserRows(mock))

	mockTransaction(mock)

	args = make([]any, 4)

	for i := 0; i < 4; i++ {
		args[i] = pgxmock.AnyArg()
	}

	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs(args...).WillReturnRows(mockUserRows(mock))

	c, err = svc.AuthJWT(ctx, authToken, "")
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet database expectations: %v", err)
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
