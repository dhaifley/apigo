package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/mail"
	"time"

	"github.com/dhaifley/apid/cache"
	"github.com/dhaifley/apid/errors"
	"github.com/dhaifley/apid/logger"
	"github.com/dhaifley/apid/request"
	"github.com/dhaifley/apid/sqldb"
	"github.com/jackc/pgx/v5"
)

// User values represent service users.
type User struct {
	UserID        request.FieldString `json:"user_id"`
	Email         request.FieldString `json:"email"`
	LastName      request.FieldString `json:"last_name"`
	FirstName     request.FieldString `json:"first_name"`
	Status        request.FieldString `json:"status"`
	Data          request.FieldJSON   `json:"data"`
	CreatedAt     request.FieldTime   `json:"created_at"`
	CreatedBy     request.FieldString `json:"created_by"`
	CreatedByUser *sqldb.UserData     `json:"created_by_user,omitempty"`
	UpdatedAt     request.FieldTime   `json:"updated_at"`
	UpdatedBy     request.FieldString `json:"updated_by"`
	UpdatedByUser *sqldb.UserData     `json:"updated_by_user,omitempty"`
}

// Validate checks that the value contains valid data.
func (u *User) Validate() error {
	if u.UserID.Set {
		if !u.UserID.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"user_id must not be null",
				"user", u)
		}

		if !request.ValidUserID(u.UserID.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid user_id",
				"user", u)
		}
	}

	if u.Status.Set {
		if !u.Status.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"status must not be null",
				"user", u)
		}

		switch u.Status.Value {
		case request.StatusActive, request.StatusInactive:
		default:
			return errors.New(errors.ErrInvalidRequest,
				"invalid status",
				"user", u)
		}
	}

	if u.Email.Set && u.Email.Valid {
		if _, err := mail.ParseAddress(u.Email.Value); err != nil {
			return errors.New(errors.ErrInvalidRequest,
				"invalid email",
				"user", u)
		}
	}

	return nil
}

// ValidateCreate checks that the value contains valid data for creation.
func (u *User) ValidateCreate() error {
	if !u.UserID.Set {
		return errors.New(errors.ErrInvalidRequest,
			"missing user_id",
			"user", u)
	}

	return u.Validate()
}

// ScanDest returns the destination fields for a SQL row scan.
func (u *User) ScanDest(options sqldb.FieldOptions) []any {
	dest := []any{
		&u.UserID,
		&u.Email,
		&u.LastName,
		&u.FirstName,
		&u.Status,
		&u.Data,
	}

	if options != nil && options.Contains(sqldb.OptUserDetails) {
		u.CreatedByUser = &sqldb.UserData{}

		u.UpdatedByUser = &sqldb.UserData{}

		dest = append(dest,
			&u.CreatedAt,
			&u.CreatedBy,
			&u.CreatedByUser.Email,
			&u.CreatedByUser.LastName,
			&u.CreatedByUser.FirstName,
			&u.CreatedByUser.Status,
			&u.CreatedByUser.Data,
			&u.UpdatedAt,
			&u.UpdatedBy,
			&u.UpdatedByUser.Email,
			&u.UpdatedByUser.LastName,
			&u.UpdatedByUser.FirstName,
			&u.UpdatedByUser.Status,
			&u.UpdatedByUser.Data,
		)
	} else {
		dest = append(dest,
			&u.CreatedAt,
			&u.CreatedBy,
			&u.UpdatedAt,
			&u.UpdatedBy,
		)
	}

	return dest
}

// userFields contain the search fields for users.
var userFields = append([]*sqldb.Field{{
	Name:   "user_key",
	Type:   sqldb.FieldInt,
	Table:  `"user"`,
	Hidden: true,
}, {
	Name:  "user_id",
	Type:  sqldb.FieldString,
	Table: `"user"`,
}, {
	Name:    "email",
	Type:    sqldb.FieldString,
	Table:   `"user"`,
	Primary: true,
}, {
	Name:  "first_name",
	Type:  sqldb.FieldString,
	Table: `"user"`,
}, {
	Name:  "last_name",
	Type:  sqldb.FieldString,
	Table: `"user"`,
}, {
	Name:  "status",
	Type:  sqldb.FieldString,
	Table: `"user"`,
}, {
	Name:  "data",
	Type:  sqldb.FieldJSON,
	Table: `"user"`,
}}, sqldb.UserFields(`"user"`)...)

// GetUser retrieves a user from the database.
func (s *Service) GetUser(ctx context.Context,
	id string,
	options sqldb.FieldOptions,
) (*User, error) {
	userID, err := request.ContextAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	if id == "" || id == current {
		id = userID
	} else if id != userID {
		if _, err := request.ContextAuthSysAdmin(ctx); err != nil {
			return nil, errors.New(errors.ErrNotFound, "user not found")
		}
	}

	if !request.ValidUserID(id) {
		return nil, errors.New(errors.ErrInvalidParameter, "invalid id",
			"id", id)
	}

	var r *User

	if s.cache != nil {
		ck := cache.KeyUser(id)

		ci, err := s.cache.Get(ctx, ck)
		if err != nil && !errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to get user cache key",
				"error", err,
				"cache_key", ck,
				"id", id)
		} else if ci != nil {
			buf := bytes.NewBuffer(ci.Value)

			if err := json.NewDecoder(buf).Decode(&r); err != nil {
				s.log.Log(ctx, logger.LvlError,
					"unable to decode user cache value",
					"error", err,
					"cache_key", ck,
					"cache_value", string(ci.Value),
					"id", id)
			}
		}
	}

	if r == nil {
		base := sqldb.SelectFields(`"user"`, userFields, nil, options) +
			`WHERE "user".user_id = $1`

		q := sqldb.NewQuery(&sqldb.QueryOptions{
			DB:     s.db,
			Type:   sqldb.QuerySelect,
			Base:   base,
			Fields: userFields,
			Params: []any{id},
		})

		q.Limit = 1

		row, err := q.QueryRow(ctx)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase, "",
				"id", id)
		}

		r = &User{}

		if err := row.Scan(r.ScanDest(options)...); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New(errors.ErrNotFound,
					"user not found",
					"id", id)
			}

			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to select user row",
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
			ck := cache.KeyUser(r.UserID.Value)

			buf, err := json.Marshal(r)
			if err != nil {
				s.log.Log(ctx, logger.LvlError,
					"unable to encode user cache value",
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
						"unable to set user cache value",
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

// CreateUser inserts a new user in the database.
func (s *Service) CreateUser(ctx context.Context,
	v *User,
) (*User, error) {
	if v == nil {
		return nil, errors.New(errors.ErrInvalidRequest,
			"missing user",
			"user", v)
	}

	if err := v.ValidateCreate(); err != nil {
		return nil, err
	}

	base := `INSERT INTO "user" () VALUES ()
		ON CONFLICT (user_id) DO UPDATE SET` +
		sqldb.ReturningFields(`"user"`, userFields, nil)

	sets, params := []string{}, []any{}

	request.SetField("user_id", v.UserID, &sets, &params)
	request.SetField("email", v.Email, &sets, &params)
	request.SetField("last_name", v.LastName, &sets, &params)
	request.SetField("first_name", v.FirstName, &sets, &params)
	request.SetField("status", v.Status, &sets, &params)
	request.SetField("data", v.Data, &sets, &params)

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QueryInsert,
		Base:   base,
		Fields: userFields,
		Sets:   sets,
		Params: params,
	})

	row, err := q.QueryRow(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase, "",
			"user", v)
	}

	r := &User{}

	if err := row.Scan(r.ScanDest(nil)...); err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to insert user row",
			"user", v)
	}

	if s.cache != nil {
		ck := cache.KeyUser(r.UserID.Value)

		if err := s.cache.Delete(ctx, ck); err != nil &&
			!errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to delete user cache key",
				"error", err,
				"cache_key", ck,
				"user", v)
		}
	}

	return r, nil
}

// UpdateUser updates a user in the database.
func (s *Service) UpdateUser(ctx context.Context,
	v *User,
) (*User, error) {
	userID, err := request.ContextUserID(ctx)
	if err != nil {
		userID = "unknown"
	}

	if v == nil {
		return nil, errors.New(errors.ErrInvalidRequest,
			"missing user",
			"user", v)
	}

	if !v.UserID.Set || !v.UserID.Valid {
		return nil, errors.New(errors.ErrInvalidRequest,
			"missing user_id",
			"user", v)
	}

	if v.UserID.Value == "" || v.UserID.Value == current {
		v.UserID.Value = userID
	}

	if err := v.Validate(); err != nil {
		return nil, err
	}

	base := `UPDATE "user" SET
		WHERE "user".user_id = $1` +
		sqldb.ReturningFields(`"user"`, userFields, nil)

	sets, params := []string{}, []any{v.UserID.Value}

	request.SetField("email", v.Email, &sets, &params)
	request.SetField("last_name", v.LastName, &sets, &params)
	request.SetField("first_name", v.FirstName, &sets, &params)
	request.SetField("status", v.Status, &sets, &params)
	request.SetField("data", v.Data, &sets, &params)
	request.SetField("updated_at", request.FieldTime{
		Set: true, Valid: true, Value: time.Now().Unix(),
	}, &sets, &params)

	if userID != "unknown" {
		request.SetField("updated_by", request.FieldString{
			Set: true, Valid: true, Value: userID,
		}, &sets, &params)
	}

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QueryUpdate,
		Base:   base,
		Fields: userFields,
		Sets:   sets,
		Params: params,
	})

	row, err := q.QueryRow(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase, "",
			"user", v)
	}

	r := &User{}

	if err := row.Scan(r.ScanDest(nil)...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(errors.ErrNotFound,
				"user not found",
				"user", v)
		}

		if errors.ErrorHas(err, `"user_pkey"`) {
			return nil, errors.New(errors.ErrConflict,
				"invalid user_id: in use by another user",
				"user", v)
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to update user row",
			"user", v)
	}

	if s.cache != nil {
		ck := cache.KeyUser(r.UserID.Value)

		if err := s.cache.Delete(ctx, ck); err != nil &&
			!errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to delete user cache key",
				"error", err,
				"cache_key", ck,
				"user", v)
		}
	}

	return r, nil
}

// DeleteUser deletes a user from the database.
func (s *Service) DeleteUser(ctx context.Context,
	id string,
) error {
	if s.cache != nil {
		defer func(ck string) {
			if err := s.cache.Delete(ctx, ck); err != nil &&
				!errors.Has(err, errors.ErrNotFound) {
				s.log.Log(ctx, logger.LvlError,
					"unable to delete user cache key",
					"error", err,
					"cache_key", ck,
					"id", id)
			}
		}(cache.KeyUser(id))
	}

	base := `DELETE FROM "user"
		WHERE "user".user_id = $1`

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QueryDelete,
		Base:   base,
		Fields: userFields,
		Params: []any{id},
	})

	res, err := q.Exec(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase, "",
			"id", id)
	}

	if n := res.RowsAffected(); n == 0 {
		return errors.New(errors.ErrNotFound,
			"user not found",
			"id", id)
	}

	return nil
}
