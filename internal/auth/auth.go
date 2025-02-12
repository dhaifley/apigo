// Package auth provides access to authentication services.
package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/dhaifley/apigo/internal/cache"
	"github.com/dhaifley/apigo/internal/config"
	"github.com/dhaifley/apigo/internal/errors"
	"github.com/dhaifley/apigo/internal/logger"
	"github.com/dhaifley/apigo/internal/metric"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/sqldb"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/trace"
)

// Claims values contain token claims information.
type Claims struct {
	AccountID   string `json:"account_id"`
	AccountName string `json:"account_name"`
	UserID      string `json:"user_id"`
	Scopes      string `json:"scopes"`
}

// Service values are used to provide access to authentication services.
type Service struct {
	cfg    *config.Config
	db     sqldb.SQLDB
	cache  cache.Accessor
	log    logger.Logger
	metric metric.Recorder
	tracer trace.Tracer
}

// NewService creates a new authentication service.
func NewService(
	cfg *config.Config,
	db sqldb.SQLDB,
	cache cache.Accessor,
	log logger.Logger,
	metric metric.Recorder,
	tracer trace.Tracer,
) *Service {
	if cfg == nil {
		cfg = config.NewDefault()
	}

	if db == nil || (reflect.ValueOf(db).Kind() == reflect.Ptr &&
		reflect.ValueOf(db).IsNil()) {
		return nil
	}

	if cache != nil && reflect.ValueOf(cache).Kind() == reflect.Ptr &&
		reflect.ValueOf(cache).IsNil() {
		cache = nil
	}

	if log == nil || (reflect.ValueOf(log).Kind() == reflect.Ptr &&
		reflect.ValueOf(log).IsNil()) {
		log = logger.NullLog
	}

	if metric == nil || (reflect.ValueOf(metric).Kind() == reflect.Ptr &&
		reflect.ValueOf(metric).IsNil()) {
		metric = nil
	}

	if tracer == nil || (reflect.ValueOf(tracer).Kind() == reflect.Ptr &&
		reflect.ValueOf(tracer).IsNil()) {
		tracer = nil
	}

	return &Service{
		cfg:    cfg,
		db:     db,
		cache:  cache,
		log:    log,
		metric: metric,
		tracer: tracer,
	}
}

// getAccountSecret retrieves an encryption secret from the database by
// account ID.
func (s *Service) getAccountSecret(ctx context.Context, accountID string,
) ([]byte, error) {
	ctx = context.WithValue(ctx, request.CtxKeyAccountID, accountID)

	base := `SELECT account.secret
	FROM account
	LIMIT 1`

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:   s.db,
		Type: sqldb.QuerySelect,
		Base: base,
	})

	q.Limit = 1

	row, err := q.QueryRow(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase, "")
	}

	var r *string

	if err := row.Scan(&r); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(errors.ErrNotFound,
				"unable to find account secret")
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to select account secret row")
	}

	if r == nil {
		return nil, errors.New(errors.ErrNotFound,
			"account secret not found")
	}

	return []byte(*r), nil
}

// AuthJWT authenticates using a JWT token.
func (s *Service) AuthJWT(ctx context.Context,
	token, tenant string,
) (*Claims, error) {
	res := &Claims{}

	tenantID := ""

	if tenant != "" {
		aCtx := context.WithValue(ctx, request.CtxKeyAccountID, "sys")

		a, err := s.GetAccountByName(aCtx, tenant)
		if err != nil {
			return nil, errors.New(errors.ErrUnauthorized,
				"invalid tenant",
				"token", token,
				"tenant", tenant)
		}

		tenantID = a.AccountID.Value
	}

	tok, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		switch token.Method.(type) {
		case *jwt.SigningMethodHMAC:
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, errors.New(errors.ErrServer,
					"unable to find kid in token headers",
					"token", token)
			}

			return s.getAccountSecret(ctx, kid)
		case *jwt.SigningMethodECDSA:
			key, err := jwt.ParseECPublicKeyFromPEM(
				s.cfg.AuthTokenPublicKey())
			if err != nil {
				return nil, errors.New(errors.ErrServer,
					"unable to parse server token key",
					"token", token)
			}

			return key, nil
		case *jwt.SigningMethodRSA:
			key, err := jwt.ParseRSAPublicKeyFromPEM(
				s.cfg.AuthTokenPublicKey())
			if err != nil {
				return nil, errors.New(errors.ErrServer,
					"unable to parse server token key",
					"token", token)
			}

			if key == nil {
				kid, ok := token.Header["kid"].(string)
				if !ok {
					return nil, errors.New(errors.ErrServer,
						"unable to find kid in token headers",
						"token", token)
				}

				key = s.cfg.AuthTokenJWKSPublicKey(kid)
			}

			if key == nil {
				return nil, errors.New(errors.ErrServer,
					"unable to find public key for token",
					"token", token)
			}

			return key, nil
		default:
			return nil, errors.New(errors.ErrUnauthorized,
				"invalid authentication token signing method",
				"token", token)
		}
	})
	if err != nil {
		s.log.Log(ctx, logger.LvlDebug,
			"unable to parse authentication token",
			"error", err)

		return nil, errors.New(errors.ErrUnauthorized,
			"invalid authentication token",
			"token", token)
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok || !tok.Valid {
		s.log.Log(ctx, logger.LvlDebug,
			"invalid authentication token used",
			"error", err,
			"token", token,
			"tenant", tenant,
			"claims", claims)

		return nil, errors.New(errors.ErrUnauthorized,
			"invalid authentication token",
			"token", token)
	}

	res.AccountID = s.cfg.ServiceName()
	res.AccountName = s.cfg.ServiceName()

	ca, err := request.ContextAccountID(ctx)
	if err != nil || ca != request.SystemAccount {
		ctx = context.WithValue(ctx, request.CtxKeyAccountID, res.AccountID)
		ctx = context.WithValue(ctx, request.CtxKeyScopes,
			request.ScopeSuperuser)

		oa, err := s.GetAccount(ctx, res.AccountID)
		if err != nil && !errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlDebug,
				"unable to retrieve account",
				"error", err,
				"token", token,
				"tenant", tenant,
				"claims", claims,
				"account_id", res.AccountID)

			return nil, err
		}

		if oa == nil {
			s.log.Log(ctx, logger.LvlDebug,
				"valid authentication token used with invalid "+
					"account",
				"error", err,
				"token", token,
				"claims", claims)

			return nil, errors.New(errors.ErrUnauthorized,
				"invalid authentication token",
				"token", token)
		}
	}

	res.Scopes, _ = claims["scopes"].(string)

	sysAdmin := false

	if strings.Contains(res.Scopes, request.ScopeSuperuser) {
		sysAdmin = true

		if aID, err := request.ContextAccountID(ctx); err == nil {
			ctx = context.WithValue(ctx, request.CtxKeyAccountID, aID)

			res.AccountID = aID
		}
	}

	if tenantID != "" && res.AccountID != tenantID && sysAdmin {
		// Cross-tenant requests currently only permitted for system admin.
		res.AccountID = tenantID
	}

	ctx = context.WithValue(ctx, request.CtxKeyAccountID, res.AccountID)

	uID, ok := claims["sub"].(string)
	if !ok {
		s.log.Log(ctx, logger.LvlDebug,
			"unable to get subject from claims",
			"error", err,
			"token", token,
			"tenant", tenant,
			"claims", claims)

		return nil, errors.New(errors.ErrUnauthorized,
			"invalid authentication token",
			"token", token)
	}

	if !request.ValidUserID(uID) {
		s.log.Log(ctx, logger.LvlDebug,
			"invalid subject found in claims",
			"error", err,
			"token", token,
			"tenant", tenant,
			"claims", claims)

		return nil, errors.New(errors.ErrUnauthorized,
			"invalid authentication token",
			"token", token)
	}

	res.UserID = uID

	return res, nil
}

// AuthPassword authenticates using a user password.
func (s *Service) AuthPassword(ctx context.Context,
	userID, password, tenant string,
) error {
	if !request.ValidUserID(userID) {
		return errors.New(errors.ErrInvalidParameter, "invalid user_id",
			"user_id", userID)
	}

	aID := ""

	if tenant != "" {
		aCtx := context.WithValue(ctx, request.CtxKeyAccountID, "sys")

		a, err := s.GetAccountByName(aCtx, tenant)
		if err != nil {
			return errors.New(errors.ErrUnauthorized,
				"invalid tenant",
				"tenant", tenant)
		}

		aID = a.AccountID.Value
	} else {
		aID = s.cfg.ServiceName()
	}

	ctx = context.WithValue(ctx, request.CtxKeyAccountID, aID)

	base := `SELECT password FROM "user"
		WHERE "user".user_id = $1`

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QuerySelect,
		Base:   base,
		Fields: userFields,
		Params: []any{userID},
	})

	q.Limit = 1

	row, err := q.QueryRow(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase, "",
			"user_id", userID)
	}

	hp := new(string)

	if err := row.Scan(&hp); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(errors.ErrNotFound,
				"user not found",
				"user_id", userID)
		}

		return errors.Wrap(err, errors.ErrDatabase,
			"unable to select user password",
			"user_id", userID)
	}

	if hp == nil || *hp == "" {
		return errors.New(errors.ErrUnauthorized,
			"user cannot login",
			"user_id", userID)
	}

	if err := verifyPassword(*hp, password); err != nil {
		return errors.New(errors.ErrUnauthorized,
			"invalid user id or password",
			"user_id", userID)
	}

	return nil
}

// Update periodically updates authentication data.
func (s *Service) Update(ctx context.Context) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)

	if tu, err := uuid.NewRandom(); err == nil {
		ctx = context.WithValue(ctx, request.CtxKeyTraceID, tu.String())
	}

	go func(ctx context.Context) {
		tick := time.NewTimer(0)

		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				if s.db == nil {
					break
				}

				ctx, cancel := request.ContextReplaceTimeout(ctx,
					s.cfg.AuthUpdateInterval())

				if tu, err := uuid.NewRandom(); err == nil {
					ctx = context.WithValue(ctx, request.CtxKeyTraceID,
						tu.String())
				}

				aid := s.cfg.AuthIdentityDomain()
				wkp := s.cfg.AuthTokenWellKnown()

				if aid == "" || wkp == "" {
					cancel()

					break
				}

				wkURL := url.URL{
					Scheme: "https",
					Host:   aid,
					Path:   wkp,
				}

				r, err := http.NewRequestWithContext(ctx, http.MethodGet,
					wkURL.String(), nil)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to create auth well known info request",
						"error", err,
						"url", wkURL.String())

					cancel()

					break
				}

				cli := &http.Client{Timeout: time.Second * 10}

				resp, err := cli.Do(r)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to retrieve auth well known info",
						"error", err)

					cancel()

					break
				}

				wk := map[string]any{}

				err = json.NewDecoder(resp.Body).Decode(&wk)

				if err := resp.Body.Close(); err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to close well known info response body",
						"error", err)
				}

				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to read well known info response body",
						"error", err)

					cancel()

					break
				}

				jwksURI, ok := wk["jwks_uri"].(string)
				if !ok || jwksURI == "" {
					s.log.Log(ctx, logger.LvlError,
						"JWKS URI not found in well known info",
						"error", err)

					cancel()

					break
				}

				rk, err := http.NewRequestWithContext(ctx, http.MethodGet,
					jwksURI, nil)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to create auth well known info request",
						"error", err,
						"url", wkURL.String())

					cancel()

					break
				}

				resp, err = cli.Do(rk)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to retrieve auth JWKS",
						"error", err)

					cancel()

					break
				}

				jwksRes := map[string]any{}

				err = json.NewDecoder(resp.Body).Decode(&jwksRes)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to read JWKS response body",
						"error", err)

					cancel()

					break
				}

				if err := resp.Body.Close(); err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to close JWKS response body",
						"error", err)
				}

				jwksList, ok := jwksRes["keys"].([]any)
				if !ok || len(jwksList) == 0 {
					s.log.Log(ctx, logger.LvlError,
						"keys not found in JWKS data",
						"response", jwksRes)

					cancel()

					break
				}

				jwks := map[string]*rsa.PublicKey{}

				for _, j := range jwksList {
					jm, ok := j.(map[string]any)
					if !ok {
						continue
					}

					alg, ok := jm["alg"].(string)
					if !ok || alg != "RS256" {
						continue
					}

					kid, ok := jm["kid"].(string)
					if !ok || kid == "" {
						continue
					}

					n, ok := jm["n"].(string)
					if !ok || n == "" {
						continue
					}

					e, ok := jm["e"].(string)
					if !ok && e == "" {
						continue
					}

					nb, err := base64.RawURLEncoding.DecodeString(n)
					if err != nil {
						s.log.Log(ctx, logger.LvlError,
							"unable to decode n value in JWKS data",
							"error", err,
							"jwks", jm,
							"n", n)

						continue
					}

					ev := 0

					if e == "AQAB" || e == "AAEAAQ" {
						ev = 65537
					} else {
						eb, err := base64.RawURLEncoding.DecodeString(e)
						if err != nil {
							s.log.Log(ctx, logger.LvlError,
								"unable to decode e value in JWKS data",
								"error", err,
								"jwks", jm,
								"e", e)
						}

						ebi := new(big.Int).SetBytes(eb)

						ev = int(ebi.Int64())
					}

					jwks[kid] = &rsa.PublicKey{
						N: new(big.Int).SetBytes(nb),
						E: ev,
					}
				}

				s.cfg.SetAuthTokenJWKS(jwks)

				cancel()
			}

			tick = time.NewTimer(s.cfg.AuthUpdateInterval())
		}
	}(ctx)

	return cancel
}

// CreateToken is used to create a JWT token that can be used for tokens.
func (s *Service) CreateToken(ctx context.Context,
	userID string,
	expiration int64,
	scopes, tenant string,
) (string, error) {
	accountID := ""

	if tenant != "" {
		aCtx := context.WithValue(ctx, request.CtxKeyAccountID, "sys")

		a, err := s.GetAccountByName(aCtx, tenant)
		if err != nil {
			return "", errors.New(errors.ErrUnauthorized,
				"invalid tenant",
				"tenant", tenant)
		}

		accountID = a.AccountID.Value
	} else {
		accountID = s.cfg.ServiceName()
	}

	if !request.ValidUserID(userID) {
		return "", errors.New(errors.ErrInvalidParameter,
			"invalid user_id",
			"user_id", userID)
	}

	if !request.ValidScopes(scopes) {
		return "", errors.New(errors.ErrInvalidParameter,
			"invalid scopes",
			"scopes", scopes)
	}

	now := time.Now()

	if now.Unix() >= expiration {
		return "", errors.New(errors.ErrInvalidParameter,
			"invalid expiration",
			"expiration", expiration)
	}

	claims := jwt.MapClaims{
		"exp":    expiration,
		"iat":    now.Unix(),
		"nbf":    now.Unix(),
		"iss":    s.cfg.AuthTokenIssuer(),
		"sub":    userID,
		"aud":    []string{s.cfg.ServiceName()},
		"scopes": scopes,
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tok.Header = map[string]any{
		"alg": "HS512",
		"typ": "JWT",
		"kid": accountID,
	}

	secret, err := s.getAccountSecret(ctx, accountID)
	if err != nil {
		return "", err
	}

	authToken, err := tok.SignedString(secret)
	if err != nil {
		return "", errors.New(errors.ErrServer,
			"unable to create token secret")
	}

	return authToken, nil
}
