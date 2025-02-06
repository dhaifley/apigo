package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/dhaifley/apid/internal/cache"
	"github.com/dhaifley/apid/internal/errors"
	"github.com/dhaifley/apid/internal/logger"
	"github.com/dhaifley/apid/internal/request"
	"github.com/dhaifley/apid/internal/search"
	"github.com/dhaifley/apid/internal/sqldb"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Token values represent tokens used for service access.
type Token struct {
	TokenID       request.FieldString `json:"token_id"`
	Status        request.FieldString `json:"status"`
	Expiration    request.FieldTime   `json:"expiration"`
	Secret        *string             `json:"secret,omitempty"`
	CreatedAt     request.FieldTime   `json:"created_at"`
	CreatedBy     request.FieldString `json:"created_by"`
	CreatedByUser *sqldb.UserData     `json:"created_by_user,omitempty"`
	UpdatedAt     request.FieldTime   `json:"updated_at"`
	UpdatedBy     request.FieldString `json:"updated_by"`
	UpdatedByUser *sqldb.UserData     `json:"updated_by_user,omitempty"`
}

// Validate checks that the value contains valid data.
func (t *Token) Validate() error {
	if t.TokenID.Set {
		if !t.TokenID.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"token_id must not be null",
				"token", t)
		}

		if !request.ValidTokenID(t.TokenID.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid token_id",
				"token", t)
		}
	}

	if t.Status.Set {
		if !t.Status.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"status must not be null",
				"token", t)
		}

		switch t.Status.Value {
		case request.StatusActive, request.StatusInactive:
		default:
			return errors.New(errors.ErrInvalidRequest,
				"invalid status",
				"token", t)
		}
	}

	if t.Expiration.Set {
		if !t.Expiration.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"expiration must not be null",
				"token", t)
		}

		now := time.Now().Unix()

		if t.Expiration.Value <= now {
			return errors.New(errors.ErrInvalidRequest,
				"invalid expiration: must be greater than current time",
				"token", t)
		}
	}

	return nil
}

// ValidateCreate checks that the value contains valid data for creation.
func (t *Token) ValidateCreate() error {
	if t.TokenID.Set {
		return errors.New(errors.ErrInvalidRequest,
			"token_id must not be set when creating",
			"token", t)
	}

	return t.Validate()
}

// ScanDest returns the destination fields for a SQL row scan.
func (t *Token) ScanDest(options sqldb.FieldOptions) []any {
	dest := []any{
		&t.TokenID,
		&t.Status,
		&t.Expiration,
	}

	if options != nil && options.Contains(sqldb.OptUserDetails) {
		t.CreatedByUser = &sqldb.UserData{}

		t.UpdatedByUser = &sqldb.UserData{}

		dest = append(dest,
			&t.CreatedAt,
			&t.CreatedBy,
			&t.CreatedByUser.Email,
			&t.CreatedByUser.LastName,
			&t.CreatedByUser.FirstName,
			&t.CreatedByUser.Status,
			&t.CreatedByUser.Data,
			&t.UpdatedAt,
			&t.UpdatedBy,
			&t.UpdatedByUser.Email,
			&t.UpdatedByUser.LastName,
			&t.UpdatedByUser.FirstName,
			&t.UpdatedByUser.Status,
			&t.UpdatedByUser.Data,
		)
	} else {
		dest = append(dest,
			&t.CreatedAt,
			&t.CreatedBy,
			&t.UpdatedAt,
			&t.UpdatedBy,
		)
	}

	return dest
}

// tokenFields contain the search fields for tokens.
var tokenFields = append([]*sqldb.Field{{
	Name:    "token_id",
	Type:    sqldb.FieldString,
	Table:   "token",
	Primary: true,
}, {
	Name:  "status",
	Type:  sqldb.FieldString,
	Table: "token",
}, {
	Name:  "expiration",
	Type:  sqldb.FieldTime,
	Table: "token",
}}, sqldb.UserFields("token")...)

// GetTokens retrieves tokens based on a search query.
func (s *Service) GetTokens(ctx context.Context,
	query *search.Query,
	options sqldb.FieldOptions,
) ([]*Token, []*sqldb.SummaryData, error) {
	if _, err := request.ContextAuthUser(ctx); err != nil {
		return nil, nil, err
	}

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QuerySelect,
		Base:   sqldb.SearchFields("token", tokenFields),
		Search: query.NoSummary(),
		Fields: tokenFields,
	})

	rows, err := q.Query(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, errors.ErrDatabase, "",
			"search", query)
	}

	keys, cacheKeys := []string{}, []string{}

	index := map[string]int{}

	for rows.Next() {
		select {
		case <-ctx.Done():
			rows.Close()

			return nil, nil, errors.Context(ctx)
		default:
		}

		k := ""

		if err = rows.Scan(&k); err != nil {
			rows.Close()

			return nil, nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to select token key",
				"search", query)
		}

		key := cache.KeyToken(k)

		cacheKeys = append(cacheKeys, key)
		index[key] = len(keys)
		keys = append(keys, k)
	}

	if err := rows.Err(); err != nil {
		rows.Close()

		return nil, nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to select token key rows",
			"search", query)
	}

	rows.Close()

	res := make([]*Token, len(index))

	sum := []*sqldb.SummaryData{}

	if len(res) == 0 {
		return res, sum, nil
	}

	if s.cache != nil && query != nil && query.Summary == "" {
		found := false

		cMap, err := s.cache.GetMulti(ctx, cacheKeys...)
		if err != nil && !errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to get token cache keys",
				"error", err,
				"cache_keys", cacheKeys,
				"search", query)
		} else {
			for ck, ci := range cMap {
				if ci == nil {
					continue
				}

				var v *Token

				buf := bytes.NewBuffer(ci.Value)

				if err := json.NewDecoder(buf).Decode(&v); err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to decode token cache value",
						"error", err,
						"cache_key", ck,
						"cache_value", string(ci.Value),
						"search", query)
				}

				if options.Contains(sqldb.OptUserDetails) {
					if v.CreatedBy.Valid {
						ud, err := sqldb.GetUserDetails(ctx, v.CreatedBy.Value,
							s.db, s.cache, s.log, s.cfg)
						if err != nil {
							return nil, nil, errors.Wrap(err,
								errors.ErrDatabase,
								"unable to select created_by user details",
								"search", query,
								"user_id", v.CreatedBy.Value)
						}

						v.CreatedByUser = ud
					}

					if v.UpdatedBy.Valid {
						ud, err := sqldb.GetUserDetails(ctx, v.UpdatedBy.Value,
							s.db, s.cache, s.log, s.cfg)
						if err != nil {
							return nil, nil, errors.Wrap(err,
								errors.ErrDatabase,
								"unable to select updated_by user details",
								"search", query,
								"user_id", v.UpdatedBy.Value)
						}

						v.UpdatedByUser = ud
					}
				}

				res[index[ck]] = v
				keys[index[ck]] = ""
				found = true
			}
		}

		if found {
			newKeys := make([]string, 0, len(keys))

			for _, k := range keys {
				if k != "" {
					newKeys = append(newKeys, k)
				}
			}

			keys = newKeys
		}
	}

	if len(keys) > 0 {
		base := sqldb.SelectFields("token",
			tokenFields, query, options) +
			`WHERE token.token_id = ANY($1::TEXT[])`

		q = sqldb.NewQuery(&sqldb.QueryOptions{
			DB:     s.db,
			Type:   sqldb.QuerySelect,
			Base:   base,
			Fields: tokenFields,
			Params: []any{keys},
		})

		if query != nil && query.Summary != "" {
			q.Search = &search.Query{Summary: query.Summary}
		}

		rows, err = q.Query(ctx)
		if err != nil {
			return nil, nil, errors.Wrap(err, errors.ErrDatabase, "",
				"search", query)
		}

		defer rows.Close()

		for rows.Next() {
			select {
			case <-ctx.Done():
				return nil, nil, errors.Context(ctx)
			default:
			}

			r := &Token{}

			sr := &sqldb.SummaryData{}

			if query != nil && query.Summary != "" {
				if err = rows.Scan(sr.ScanDest(tokenFields,
					query)...); err != nil {
					return nil, nil, errors.Wrap(err, errors.ErrDatabase,
						"unable to select token summary row",
						"search", query)
				}

				sum = append(sum, sr)

				continue
			}

			if err = rows.Scan(r.ScanDest(options)...); err != nil {
				return nil, nil, errors.Wrap(err, errors.ErrDatabase,
					"unable to select token row",
					"search", query)
			}

			var cbu *sqldb.UserData

			if r.CreatedByUser != nil {
				r.CreatedByUser.UserID = r.CreatedBy
				cbu = r.CreatedByUser
				r.CreatedByUser = nil
			}

			var ubu *sqldb.UserData

			if r.UpdatedByUser != nil {
				r.UpdatedByUser.UserID = r.UpdatedBy
				ubu = r.UpdatedByUser
				r.UpdatedByUser = nil
			}

			if s.cache != nil {
				ck := cache.KeyToken(r.TokenID.Value)

				buf, err := json.Marshal(r)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to encode token cache value",
						"error", err,
						"cache_key", ck,
						"cache_value", r,
						"search", query)
				} else if len(buf) < s.cfg.CacheMaxBytes() {
					if err := s.cache.Set(ctx, &cache.Item{
						Key:        ck,
						Value:      buf,
						Expiration: s.cfg.CacheExpiration(),
					}); err != nil {
						s.log.Log(ctx, logger.LvlError,
							"unable to set token cache value",
							"error", err,
							"cache_key", ck,
							"cache_value", string(buf),
							"expiration", s.cfg.CacheExpiration(),
							"search", query)
					}
				}
			}

			if cbu != nil {
				r.CreatedByUser = cbu
			}

			if ubu != nil {
				r.UpdatedByUser = ubu
			}

			res[index[cache.KeyToken(r.TokenID.Value)]] = r
		}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to select token rows",
			"search", query)
	}

	if len(sum) > 0 {
		res = []*Token{}
	}

	return res, sum, nil
}

// GetToken retrieves a token from the database.
func (s *Service) GetToken(ctx context.Context,
	id string,
	options sqldb.FieldOptions,
) (*Token, error) {
	if _, err := request.ContextAuthUser(ctx); err != nil {
		return nil, err
	}

	if !request.ValidTokenID(id) {
		return nil, errors.New(errors.ErrInvalidParameter,
			"invalid token_id",
			"id", id)
	}

	var r *Token

	if s.cache != nil {
		ck := cache.KeyToken(id)

		ci, err := s.cache.Get(ctx, ck)
		if err != nil && !errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to get token cache key",
				"error", err,
				"cache_key", ck,
				"id", id)
		} else if ci != nil {
			buf := bytes.NewBuffer(ci.Value)

			if err := json.NewDecoder(buf).Decode(&r); err != nil {
				s.log.Log(ctx, logger.LvlError,
					"unable to decode token cache value",
					"error", err,
					"cache_key", ck,
					"cache_value", string(ci.Value),
					"id", id)
			} else {
				if options.Contains(sqldb.OptUserDetails) {
					if r.CreatedBy.Valid {
						ud, err := sqldb.GetUserDetails(ctx, r.CreatedBy.Value,
							s.db, s.cache, s.log, s.cfg)
						if err != nil {
							return nil, errors.Wrap(err, errors.ErrDatabase,
								"unable to select created_by user details",
								"id", id,
								"user_id", r.CreatedBy.Value)
						}

						r.CreatedByUser = ud
					}

					if r.UpdatedBy.Valid {
						ud, err := sqldb.GetUserDetails(ctx, r.UpdatedBy.Value,
							s.db, s.cache, s.log, s.cfg)
						if err != nil {
							return nil, errors.Wrap(err, errors.ErrDatabase,
								"unable to select updated_by user details",
								"id", id,
								"user_id", r.UpdatedBy.Value)
						}

						r.UpdatedByUser = ud
					}
				}
			}
		}
	}

	if r == nil {
		base := sqldb.SelectFields("token",
			tokenFields, nil, options) +
			`WHERE token.token_id = $1`

		q := sqldb.NewQuery(&sqldb.QueryOptions{
			DB:     s.db,
			Type:   sqldb.QuerySelect,
			Base:   base,
			Fields: tokenFields,
			Params: []any{id},
		})

		q.Limit = 1

		row, err := q.QueryRow(ctx)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase, "",
				"id", id)
		}

		r = &Token{}

		if err := row.Scan(r.ScanDest(options)...); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New(errors.ErrNotFound,
					"token not found",
					"id", id)
			}

			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to select token row",
				"id", id)
		}

		var cbu *sqldb.UserData

		if r.CreatedByUser != nil {
			r.CreatedByUser.UserID = r.CreatedBy
			cbu = r.CreatedByUser
			r.CreatedByUser = nil
		}

		var ubu *sqldb.UserData

		if r.UpdatedByUser != nil {
			r.UpdatedByUser.UserID = r.UpdatedBy
			ubu = r.UpdatedByUser
			r.UpdatedByUser = nil
		}

		if s.cache != nil {
			ck := cache.KeyToken(r.TokenID.Value)

			buf, err := json.Marshal(r)
			if err != nil {
				s.log.Log(ctx, logger.LvlError,
					"unable to encode token cache value",
					"error", err,
					"cache_key", ck,
					"cache_value", r,
					"id", id)
			} else if len(buf) < s.cfg.CacheMaxBytes() {
				if err := s.cache.Set(ctx, &cache.Item{
					Key:        ck,
					Value:      buf,
					Expiration: s.cfg.CacheExpiration(),
				}); err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to set token cache value",
						"error", err,
						"cache_key", ck,
						"cache_value", string(buf),
						"expiration", s.cfg.CacheExpiration(),
						"id", id)
				}
			}
		}

		if cbu != nil {
			r.CreatedByUser = cbu
		}

		if ubu != nil {
			r.UpdatedByUser = ubu
		}
	}

	return r, nil
}

// createSecret is used to create a JWT secret that can be used for tokens.
func (s *Service) createSecret(ctx context.Context,
	expiration int64,
	tokenID string,
) (string, error) {
	now := time.Now()

	aID, err := request.ContextAccountID(ctx)
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"exp":      expiration,
		"iat":      now.Unix(),
		"nbf":      now.Unix(),
		"iss":      s.cfg.AuthTokenIssuer(),
		"sub":      "0",
		"aud":      []string{s.cfg.ServiceName()},
		"role":     request.RoleRefresh,
		"token_id": tokenID,
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tok.Header = map[string]any{
		"alg": "HS512",
		"typ": "JWT",
		"kid": aID,
	}

	secret, err := s.getAccountSecret(ctx, aID)
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

// CreateToken inserts a new token in the database.
func (s *Service) CreateToken(ctx context.Context,
	v *Token,
) (*Token, error) {
	userID, err := request.ContextAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, errors.New(errors.ErrInvalidRequest,
			"missing token",
			"token", v)
	}

	if err := v.ValidateCreate(); err != nil {
		return nil, err
	}

	secret := ""

	tok, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrServer,
			"unable to create token uuid",
			"token", v)
	}

	v.TokenID = request.FieldString{
		Set: true, Valid: true, Value: tok.String(),
	}

	if !v.Expiration.Valid {
		v.Expiration = request.FieldTime{
			Set: true, Valid: true,
			Value: time.Now().Add(time.Hour * 24 * 7).Unix(),
		}
	}

	secret, err = s.createSecret(ctx, v.Expiration.Value, tok.String())
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrServer,
			"unable to create token secret key",
			"token", v)
	}

	base := `INSERT INTO token () VALUES () ` +
		sqldb.ReturningFields("token", tokenFields, nil)

	sets, params := []string{}, []any{}

	request.SetField("token_id", v.TokenID, &sets, &params)
	request.SetField("status", v.Status, &sets, &params)
	request.SetField("expiration", v.Expiration, &sets, &params)
	request.SetField("created_by", request.FieldString{
		Set: true, Valid: true, Value: userID,
	}, &sets, &params)
	request.SetField("updated_by", request.FieldString{
		Set: true, Valid: true, Value: userID,
	}, &sets, &params)

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QueryInsert,
		Base:   base,
		Fields: tokenFields,
		Sets:   sets,
		Params: params,
	})

	row, err := q.QueryRow(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase, "",
			"token", v)
	}

	r := &Token{}

	if err := row.Scan(r.ScanDest(nil)...); err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to insert token row",
			"token", v)
	}

	r.Secret = &secret

	if s.cache != nil {
		ck := cache.KeyToken(r.TokenID.Value)

		if err := s.cache.Delete(ctx, ck); err != nil &&
			!errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to delete token cache key",
				"error", err,
				"cache_key", ck,
				"token", v)
		}
	}

	return r, nil
}

// DeleteToken deletes a token from the database.
func (s *Service) DeleteToken(ctx context.Context,
	id string,
) error {
	if _, err := request.ContextAuthUser(ctx); err != nil {
		return err
	}

	rt, err := s.GetToken(ctx, id, nil)
	if err != nil {
		return err
	}

	if s.cache != nil {
		defer func(ck string) {
			if err := s.cache.Delete(ctx, ck); err != nil &&
				!errors.Has(err, errors.ErrNotFound) {
				s.log.Log(ctx, logger.LvlError,
					"unable to delete token cache key",
					"error", err,
					"cache_key", ck,
					"id", id)
			}
		}(cache.KeyToken(id))
	}

	base := `DELETE FROM token
		WHERE token.token_id = $1`

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QueryDelete,
		Base:   base,
		Fields: tokenFields,
		Params: []any{rt.TokenID.Value},
	})

	res, err := q.Exec(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase, "",
			"id", id)
	}

	if n := res.RowsAffected(); n == 0 {
		return errors.New(errors.ErrNotFound,
			"token not found",
			"id", id)
	}

	return nil
}
