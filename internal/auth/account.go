package auth

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/dhaifley/apigo/internal/cache"
	"github.com/dhaifley/apigo/internal/errors"
	"github.com/dhaifley/apigo/internal/logger"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/sqldb"
	"github.com/jackc/pgx/v5"
)

const current = "current"

// Account values represent service accounts.
type Account struct {
	AccountID      request.FieldString `json:"account_id"`
	Name           request.FieldString `json:"name"`
	Status         request.FieldString `json:"status"`
	StatusData     request.FieldJSON   `json:"status_data"`
	Repo           request.FieldString `json:"-"`
	RepoStatus     request.FieldString `json:"repo_status"`
	RepoStatusData request.FieldJSON   `json:"repo_status_data"`
	Secret         request.FieldString `json:"-"`
	Data           request.FieldJSON   `json:"data"`
	CreatedAt      request.FieldTime   `json:"created_at"`
	UpdatedAt      request.FieldTime   `json:"updated_at"`
}

// Validate checks that the value contains valid data.
func (a *Account) Validate() error {
	if a.AccountID.Set {
		if !a.AccountID.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"account_id must not be null",
				"account", a)
		}

		if !request.ValidAccountID(a.AccountID.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid account_id",
				"account", a)
		}
	}

	if a.Name.Set {
		if !a.Name.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"name must not be null",
				"account", a)
		}

		if !request.ValidAccountName(a.Name.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid name",
				"account", a)
		}
	}

	if a.Status.Set {
		if !a.Status.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"status must not be null",
				"account", a)
		}

		switch a.Status.Value {
		case request.StatusActive, request.StatusInactive:
		default:
			return errors.New(errors.ErrInvalidRequest,
				"invalid status",
				"account", a)
		}
	}

	if a.RepoStatus.Set {
		if !a.RepoStatus.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"repo_status must not be null",
				"account", a)
		}

		switch a.RepoStatus.Value {
		case request.StatusActive, request.StatusInactive,
			request.StatusError, request.StatusImporting:
		default:
			return errors.New(errors.ErrInvalidRequest,
				"invalid repo_status",
				"account", a)
		}
	}

	return nil
}

// ValidateCreate checks that the value contains valid data for creation.
func (a *Account) ValidateCreate() error {
	if !a.AccountID.Set {
		return errors.New(errors.ErrInvalidRequest,
			"missing account_id",
			"account", a)
	}

	if !a.Name.Set {
		return errors.New(errors.ErrInvalidRequest,
			"missing name",
			"account", a)
	}

	return a.Validate()
}

// ScanDest returns the destination fields for a SQL row scan.
func (a *Account) ScanDest() []any {
	return []any{
		&a.AccountID,
		&a.Name,
		&a.Status,
		&a.StatusData,
		&a.Repo,
		&a.RepoStatus,
		&a.RepoStatusData,
		&a.Secret,
		&a.Data,
		&a.CreatedAt,
		&a.UpdatedAt,
	}
}

// accountFields contain the search fields for accounts.
var accountFields = []*sqldb.Field{{
	Name:  "account_id",
	Type:  sqldb.FieldString,
	Table: "account",
}, {
	Name:    "name",
	Type:    sqldb.FieldString,
	Table:   "account",
	Primary: true,
}, {
	Name:  "status",
	Type:  sqldb.FieldString,
	Table: "account",
}, {
	Name:  "status_data",
	Type:  sqldb.FieldJSON,
	Table: "account",
}, {
	Name:  "repo",
	Type:  sqldb.FieldString,
	Table: "account",
}, {
	Name:  "repo_status",
	Type:  sqldb.FieldString,
	Table: "account",
}, {
	Name:  "repo_status_data",
	Type:  sqldb.FieldString,
	Table: "account",
}, {
	Name:  "secret",
	Type:  sqldb.FieldString,
	Table: "account",
}, {
	Name:  "data",
	Type:  sqldb.FieldJSON,
	Table: "account",
}, {
	Name:  "created_at",
	Type:  sqldb.FieldTime,
	Table: "account",
}, {
	Name:  "updated_at",
	Type:  sqldb.FieldTime,
	Table: "account",
}}

// GetAccount retrieves an account from the database.
func (s *Service) GetAccount(ctx context.Context,
	id string,
) (*Account, error) {
	if id == "" || id == current {
		accountID, err := request.ContextAccountID(ctx)
		if err != nil {
			return nil, errors.New(errors.ErrForbidden,
				"unable to retrieve account id",
				"id", id)
		}

		id = accountID
	}

	if !request.ValidAccountID(id) {
		return nil, errors.New(errors.ErrInvalidParameter,
			"invalid id",
			"id", id)
	}

	var r *Account

	if s.cache != nil {
		ck := cache.KeyAccount(id)

		ci, err := s.cache.Get(ctx, ck)
		if err != nil && !errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to get account cache key",
				"error", err,
				"cache_key", ck,
				"id", id)
		} else if ci != nil {
			buf := bytes.NewBuffer(ci.Value)

			if err := json.NewDecoder(buf).Decode(&r); err != nil {
				s.log.Log(ctx, logger.LvlError,
					"unable to decode account cache value",
					"error", err,
					"cache_key", ck,
					"cache_value", string(ci.Value),
					"id", id)
			}
		}
	}

	if r == nil {
		base := sqldb.SelectFields("account", accountFields, nil, nil) +
			`WHERE account.account_id = $1`

		q := sqldb.NewQuery(&sqldb.QueryOptions{
			DB:     s.db,
			Type:   sqldb.QuerySelect,
			Base:   base,
			Fields: accountFields,
			Params: []any{id},
		})

		q.Limit = 1

		row, err := q.QueryRow(ctx)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase, "",
				"id", id)
		}

		r = &Account{}

		if err := row.Scan(r.ScanDest()...); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New(errors.ErrNotFound,
					"account not found",
					"id", id)
			}

			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to select account row",
				"id", id)
		}

		if s.cache != nil {
			ck := cache.KeyAccount(r.AccountID.Value)

			buf, err := json.Marshal(r)
			if err != nil {
				s.log.Log(ctx, logger.LvlError,
					"unable to encode account cache value",
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
						"unable to set account cache value",
						"error", err,
						"cache_key", ck,
						"cache_value", string(buf),
						"expiration", s.cfg.CacheExpiration(),
						"id", id)
				}
			}
		}
	}

	return r, nil
}

// GetAccountByName retrieves an account from the database by name not ID.
func (s *Service) GetAccountByName(ctx context.Context,
	name string,
) (*Account, error) {
	if name == "" || name == current {
		an, err := request.ContextAccountName(ctx)
		if err != nil {
			return nil, errors.New(errors.ErrForbidden,
				"unable to retrieve account name",
				"name", name)
		}

		name = an
	}

	if !request.ValidAccountName(name) {
		return nil, errors.New(errors.ErrInvalidParameter,
			"invalid account name",
			"name", name)
	}

	var r *Account

	if s.cache != nil {
		ck := cache.KeyAccountName(name)

		ci, err := s.cache.Get(ctx, ck)
		if err != nil && !errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to get account name cache key",
				"error", err,
				"cache_key", ck,
				"name", name)
		} else if ci != nil {
			buf := bytes.NewBuffer(ci.Value)

			if err := json.NewDecoder(buf).Decode(&r); err != nil {
				s.log.Log(ctx, logger.LvlError,
					"unable to decode account name cache value",
					"error", err,
					"cache_key", ck,
					"cache_value", string(ci.Value),
					"name", name)
			}
		}
	}

	if r == nil {
		base := sqldb.SelectFields("account", accountFields, nil, nil) +
			`WHERE account.name = $1`

		q := sqldb.NewQuery(&sqldb.QueryOptions{
			DB:     s.db,
			Type:   sqldb.QuerySelect,
			Base:   base,
			Fields: accountFields,
			Params: []any{name},
		})

		q.Limit = 1

		row, err := q.QueryRow(ctx)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase, "",
				"name", name)
		}

		r = &Account{}

		if err := row.Scan(r.ScanDest()...); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New(errors.ErrNotFound,
					"account not found",
					"name", name)
			}

			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to select account row by name",
				"name", name)
		}

		if s.cache != nil {
			ck := cache.KeyAccountName(r.Name.Value)

			buf, err := json.Marshal(r)
			if err != nil {
				s.log.Log(ctx, logger.LvlError,
					"unable to encode account name cache value",
					"error", err,
					"cache_key", ck,
					"cache_value", r,
					"name", name)
			} else if len(buf) < s.cfg.CacheMaxBytes() {
				if err := s.cache.Set(ctx, &cache.Item{
					Key:        ck,
					Value:      buf,
					Expiration: s.cfg.CacheExpiration(),
				}); err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to set account name cache value",
						"error", err,
						"cache_key", ck,
						"cache_value", string(buf),
						"expiration", s.cfg.CacheExpiration(),
						"name", name)
				}
			}
		}
	}

	return r, nil
}

// CreateAccount inserts a new account in the database.
func (s *Service) CreateAccount(ctx context.Context,
	v *Account,
) (*Account, error) {
	accountID := ""

	if !request.ContextHasScope(ctx, request.ScopeSuperuser) {
		if !request.ContextHasScope(ctx, request.ScopeAccountAdmin) {
			return nil, errors.New(errors.ErrForbidden,
				"unable to create account",
				"account", v)
		} else {
			if aID, err := request.ContextAccountID(ctx); err != nil {
				return nil, errors.New(errors.ErrForbidden,
					"unable to update account",
					"account", v)
			} else {
				accountID = aID
			}
		}
	}

	repo := ""

	if v == nil {
		return nil, errors.New(errors.ErrInvalidRequest,
			"missing account",
			"account", v)
	}

	if accountID != "" {
		v.AccountID = request.FieldString{
			Set: true, Valid: true, Value: accountID,
		}
	}

	if !v.Name.Set {
		v.Name = request.FieldString{
			Set: true, Valid: true, Value: v.AccountID.Value,
		}
	}

	if (!v.Repo.Set || v.Repo.Value == "") && repo != "" {
		v.Repo = request.FieldString{
			Set: true, Valid: true, Value: repo,
		}
	}

	if err := v.ValidateCreate(); err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, request.CtxKeyAccountID, v.AccountID.Value)

	base := `INSERT INTO account () VALUES ()
		ON CONFLICT (account_id) DO UPDATE SET` +
		sqldb.ReturningFields("account", accountFields, nil)

	sets, params := []string{}, []any{}

	request.SetField("account_id", v.AccountID, &sets, &params)
	request.SetField("name", v.Name, &sets, &params)
	request.SetField("status", v.Status, &sets, &params)
	request.SetField("status_data", v.StatusData, &sets, &params)
	request.SetField("repo", v.Repo, &sets, &params)
	request.SetField("repo_status", v.RepoStatus, &sets, &params)
	request.SetField("repo_status_data", v.RepoStatusData, &sets, &params)
	request.SetField("secret", v.Secret, &sets, &params)
	request.SetField("data", v.Data, &sets, &params)

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QueryInsert,
		Base:   base,
		Fields: accountFields,
		Sets:   sets,
		Params: params,
	})

	row, err := q.QueryRow(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase, "",
			"account", v)
	}

	r := &Account{}

	if err := row.Scan(r.ScanDest()...); err != nil {
		if errors.ErrorHas(err, `"account_pkey"`) {
			return nil, errors.Wrap(err, errors.ErrConflict,
				"account already exists",
				"account", v)
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to insert account row",
			"account", v)
	}

	if s.cache != nil {
		ck := cache.KeyAccount(r.AccountID.Value)

		if err := s.cache.Delete(ctx, ck); err != nil &&
			!errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to delete account cache key",
				"error", err,
				"cache_key", ck,
				"account", v)
		}

		ck = cache.KeyAccountName(r.Name.Value)

		if err := s.cache.Delete(ctx, ck); err != nil &&
			!errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to delete account name cache key",
				"error", err,
				"cache_key", ck,
				"account", v)
		}
	}

	return r, nil
}

// AccountRepo values represent an account import repository.
type AccountRepo struct {
	Repo           request.FieldString `json:"repo"`
	RepoStatus     request.FieldString `json:"repo_status"`
	RepoStatusData request.FieldJSON   `json:"repo_status_data"`
}

// GetAccountRepo retrieves the account repository from the database.
func (s *Service) GetAccountRepo(ctx context.Context) (*AccountRepo, error) {
	admin := true

	if !request.ContextHasScope(ctx, request.ScopeSuperuser) &&
		!request.ContextHasScope(ctx, request.ScopeAccountAdmin) {
		admin = false
	}

	base := `SELECT
		account.repo,
		account.repo_status,
		account.repo_status_data
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

	r := &AccountRepo{}

	if err := row.Scan(&r.Repo, &r.RepoStatus, &r.RepoStatusData); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(errors.ErrNotFound,
				"unable to find account repo")
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to select account repo row")
	}

	if !admin {
		r.Repo = request.FieldString{Set: false, Valid: false}
	}

	return r, nil
}

// SetAccountRepo sets the account repository in the database.
func (s *Service) SetAccountRepo(ctx context.Context,
	v *AccountRepo,
) error {
	if !request.ContextHasScope(ctx, request.ScopeSuperuser) &&
		!request.ContextHasScope(ctx, request.ScopeAccountAdmin) {
		return errors.New(errors.ErrForbidden,
			"unable to set account repo",
			"repo", v)
	}

	accountID := ""

	if aID, err := request.ContextAccountID(ctx); err != nil {
		return errors.New(errors.ErrForbidden,
			"unable to get account from context",
			"repo", v)
	} else {
		accountID = aID
	}

	if v == nil {
		return errors.New(errors.ErrInvalidRequest,
			"missing repo")
	}

	if v.RepoStatus.Set {
		if !v.RepoStatus.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"repo_status must not be null",
				"repo", v)
		}

		switch v.RepoStatus.Value {
		case request.StatusActive, request.StatusInactive,
			request.StatusError, request.StatusImporting:
		default:
			return errors.New(errors.ErrInvalidRequest,
				"invalid repo_status",
				"repo", v)
		}
	}

	base := `UPDATE account SET
	WHERE account_id = $1
	RETURNING repo, repo_status, repo_status_data`

	sets, params := []string{}, []any{accountID}

	request.SetField("repo", v.Repo, &sets, &params)
	request.SetField("repo_status", v.RepoStatus, &sets, &params)
	request.SetField("repo_status_data", v.RepoStatusData, &sets, &params)

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QueryUpdate,
		Base:   base,
		Sets:   sets,
		Params: params,
	})

	q.Limit = 1

	row, err := q.QueryRow(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase, "")
	}

	r := &AccountRepo{}

	if err := row.Scan(&r.Repo, &r.RepoStatus, &r.RepoStatusData); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(errors.ErrNotFound,
				"unable to find account to set repo data")
		}

		return errors.Wrap(err, errors.ErrDatabase,
			"unable to set account repo data")
	}

	return nil
}
