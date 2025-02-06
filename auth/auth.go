// Package auth provides access to authentication services.
package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/dhaifley/apid/cache"
	"github.com/dhaifley/apid/config"
	"github.com/dhaifley/apid/errors"
	"github.com/dhaifley/apid/logger"
	"github.com/dhaifley/apid/metric"
	"github.com/dhaifley/apid/request"
	"github.com/dhaifley/apid/sqldb"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/trace"
)

// Claims values contain token claims information.
type Claims struct {
	AccountID   string   `json:"account_id"`
	AccountName string   `json:"account_name"`
	UserID      string   `json:"user_id"`
	TokenID     string   `json:"token_id"`
	Roles       []string `json:"roles"`
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

// verifyToken verifies the existence of an active token in the database by
// token ID.
func (s *Service) verifyToken(ctx context.Context, tokenID string,
) error {
	base := `SELECT token.token_id
	FROM token
	WHERE token.token_id = $1
		AND token.status = 'active'
		AND token.expiration > CURRENT_TIMESTAMP
	LIMIT 1`

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QuerySelect,
		Base:   base,
		Params: []any{tokenID},
	})

	q.Limit = 1

	row, err := q.QueryRow(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"")
	}

	var r string

	if err := row.Scan(&r); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(errors.ErrNotFound,
				"token not found")
		}

		return errors.Wrap(err, errors.ErrDatabase,
			"unable to select token row")
	}

	if r == "" {
		return errors.New(errors.ErrNotFound,
			"token was not found")
	}

	return nil
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
	if err != nil || ca != request.SystemUser {
		ctx = context.WithValue(ctx, request.CtxKeyAccountID, res.AccountID)

		ctx = context.WithValue(ctx, request.CtxKeyRoles, []string{
			request.RoleSystemAdmin,
		})

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
			secret := ""

			if uID, err := uuid.NewRandom(); err != nil {
				return nil, errors.Wrap(err, errors.ErrServer,
					"unable to create account secret",
					"token", token,
					"tenant", tenant,
					"claims", claims,
					"account_id", res.AccountID)
			} else {
				secret = uID.String()
			}

			if _, err := s.CreateAccount(ctx, &Account{
				AccountID: request.FieldString{
					Set: true, Valid: true, Value: res.AccountID,
				},
				Name: request.FieldString{
					Set: true, Valid: true, Value: res.AccountName,
				},
				Secret: request.FieldString{
					Set: true, Valid: true, Value: secret,
				},
			}); err != nil {
				if errors.Has(err, errors.ErrForbidden) {
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

				s.log.Log(ctx, logger.LvlError,
					"unable to create account",
					"error", err,
					"token", token,
					"tenant", tenant,
					"claims", claims,
					"account_id", res.AccountID)

				return nil, err
			}
		}
	}

	role, ok := claims["role"].(string)
	if !ok || len(role) == 0 {
		role = request.RoleUser
	}

	refresh, sysAdmin := false, false

	res.Roles = append(res.Roles, role)

	switch role {
	case request.RoleRefresh:
		refresh = true
	case request.RoleSystemAdmin:
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

	if refresh {
		if tID, ok := claims["token_id"].(string); !ok {
			s.log.Log(ctx, logger.LvlDebug,
				"unable to get token id from claims",
				"token", token,
				"tenant", tenant,
				"claims", claims)

			return nil, errors.New(errors.ErrUnauthorized,
				"invalid authentication token",
				"token", token)
		} else {
			if err := s.verifyToken(ctx, tID); err != nil {
				s.log.Log(ctx, logger.LvlDebug,
					"unable to verify token",
					"error", err,
					"token", token,
					"tenant", tenant,
					"token_id", tID)

				return nil, errors.New(errors.ErrUnauthorized,
					"invalid authentication token",
					"token", token)
			}

			res.TokenID = tID
		}
	}

	if uID, ok := claims["sub"].(string); !ok {
		s.log.Log(ctx, logger.LvlDebug,
			"unable to get subject from claims",
			"error", err,
			"token", token,
			"tenant", tenant,
			"claims", claims)

		return nil, errors.New(errors.ErrUnauthorized,
			"invalid authentication token",
			"token", token)
	} else if !request.ValidUserID(uID) {
		s.log.Log(ctx, logger.LvlDebug,
			"invalid subject found in claims",
			"error", err,
			"token", token,
			"tenant", tenant,
			"claims", claims)

		return nil, errors.New(errors.ErrUnauthorized,
			"invalid authentication token",
			"token", token)
	} else {
		u := &User{
			UserID: request.FieldString{
				Set: true, Valid: true, Value: uID,
			},
			Status: request.FieldString{
				Set: true, Valid: true, Value: request.StatusActive,
			},
		}

		cu, err := request.ContextUserID(ctx)
		if err != nil || cu != request.SystemUser {
			ctx = context.WithValue(ctx, request.CtxKeyUserID, u.UserID.Value)

			ctx = context.WithValue(ctx, request.CtxKeyRoles, []string{
				request.RoleSystemAdmin,
			})

			ou, err := s.GetUser(ctx, u.UserID.Value, nil)
			if err != nil {
				if !errors.Has(err, errors.ErrNotFound) {
					s.log.Log(ctx, logger.LvlError,
						"unable to retrieve user",
						"error", err,
						"token", token,
						"tenant", tenant,
						"claims", claims,
						"user", u)

					return nil, err
				}
			}

			if ou == nil || (ou.Email.Value != u.Email.Value ||
				ou.FirstName.Value != u.FirstName.Value ||
				ou.LastName.Value != u.LastName.Value ||
				ou.Status.Value != u.Status.Value) {
				if _, err := s.CreateUser(ctx, u); err != nil {
					if errors.Has(err, errors.ErrForbidden) {
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

					s.log.Log(ctx, logger.LvlError,
						"unable to create or update user",
						"error", err,
						"token", token,
						"tenant", tenant,
						"claims", claims,
						"user", u)

					return nil, err
				}
			}
		}

		res.UserID = u.UserID.Value
	}

	return res, nil
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

				cli.Transport = &http.Transport{
					Proxy: http.ProxyFromEnvironment,
					DialContext: (&net.Dialer{
						Timeout:   cli.Timeout,
						KeepAlive: 30 * time.Second,
						DualStack: true,
					}).DialContext,
					ForceAttemptHTTP2:     true,
					DisableKeepAlives:     true,
					MaxConnsPerHost:       0,
					MaxIdleConns:          10,
					MaxIdleConnsPerHost:   1,
					IdleConnTimeout:       cli.Timeout,
					TLSHandshakeTimeout:   cli.Timeout,
					ExpectContinueTimeout: cli.Timeout,
				}

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
