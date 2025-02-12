package resource

import (
	"bytes"
	"context"
	"encoding/json"
	"math/rand/v2"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dhaifley/apigo/internal/auth"
	"github.com/dhaifley/apigo/internal/cache"
	"github.com/dhaifley/apigo/internal/config"
	"github.com/dhaifley/apigo/internal/errors"
	"github.com/dhaifley/apigo/internal/logger"
	"github.com/dhaifley/apigo/internal/metric"
	"github.com/dhaifley/apigo/internal/repo"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/search"
	"github.com/dhaifley/apigo/internal/sqldb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/yaml.v3"
)

// AuthService values are used to access authentication services.
type AuthService interface {
	GetAccountRepo(ctx context.Context) (*auth.AccountRepo, error)
	SetAccountRepo(ctx context.Context, v *auth.AccountRepo) error
}

// Service values are used to provide functionality for managing telemetry
// rules.
type Service struct {
	cfg           *config.Config
	db            sqldb.SQLDB
	cache         cache.Accessor
	log           logger.Logger
	metric        metric.Recorder
	tracer        trace.Tracer
	getRepoClient func(repoURL string) (repo.Client, error)
}

// NewService creates a new service.
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

	s := &Service{
		cfg:    cfg,
		db:     db,
		cache:  cache,
		log:    log,
		metric: metric,
		tracer: tracer,
	}

	s.getRepoClient = func(repoURL string) (repo.Client, error) {
		return repo.NewClient(repoURL, s.metric, s.tracer)
	}

	return s
}

// SetRepoClient sets the git repository client to be used for imports.
func (s *Service) SetRepoClient(cli repo.Client) {
	s.getRepoClient = func(repoURL string) (repo.Client, error) {
		return cli, nil
	}
}

// Resource values represent individual external resource conditions.
type Resource struct {
	ResourceID     request.FieldString `json:"resource_id"`
	Name           request.FieldString `json:"name"`
	Version        request.FieldString `json:"version"`
	Description    request.FieldString `json:"description"`
	Status         request.FieldString `json:"status"`
	StatusData     request.FieldJSON   `json:"status_data"`
	KeyField       request.FieldString `json:"key_field"`
	KeyRegex       request.FieldString `json:"key_regex"`
	ClearCondition request.FieldString `json:"clear_condition"`
	ClearAfter     request.FieldInt64  `json:"clear_after"`
	ClearDelay     request.FieldInt64  `json:"clear_delay"`
	Data           request.FieldJSON   `json:"data"`
	Source         request.FieldString `json:"source"`
	CommitHash     request.FieldString `json:"commit_hash"`
	CreatedAt      request.FieldTime   `json:"created_at"`
	CreatedBy      request.FieldString `json:"created_by"`
	UpdatedAt      request.FieldTime   `json:"updated_at"`
	UpdatedBy      request.FieldString `json:"updated_by"`
}

// Validate checks that the value contains valid data.
func (r *Resource) Validate(cfg *config.Config) error {
	if r.ResourceID.Set {
		if !r.ResourceID.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"resource_id must not be null",
				"resource", r)
		}

		if !request.ValidResourceID(r.ResourceID.Value) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid resource_id",
				"resource", r)
		}
	}

	if r.Name.Set && !r.Name.Valid {
		return errors.New(errors.ErrInvalidRequest,
			"name must not be null",
			"resource", r)
	}

	if r.KeyField.Set && !r.KeyField.Valid {
		return errors.New(errors.ErrInvalidRequest,
			"key_field must not be null",
			"resource", r)
	}

	if r.ClearCondition.Set && r.ClearCondition.Valid {
		p := search.NewParser(bytes.NewBufferString(r.ClearCondition.Value))

		if _, err := p.Parse(); err != nil {
			return errors.Wrap(err, errors.ErrInvalidRequest,
				"invalid clear_condition",
				"resource", r)
		}
	}

	if r.ClearAfter.Set {
		if !r.ClearAfter.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"clear_after must not be null",
				"resource", r)
		}

		if r.ClearAfter.Value < 0 ||
			r.ClearAfter.Value > int64(cfg.ResourceDataRetention().Seconds()) {
			return errors.New(errors.ErrInvalidRequest,
				"invalid clear_after",
				"resource", r)
		}
	}

	if r.ClearDelay.Set {
		if !r.ClearDelay.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"clear_delay must not be null",
				"resource", r)
		}

		if r.ClearDelay.Value < 0 || r.ClearDelay.Value > 60*60 {
			return errors.New(errors.ErrInvalidRequest,
				"invalid clear_delay",
				"resource", r)
		}
	}

	if r.Status.Set {
		if !r.Status.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"status must not be null",
				"resource", r)
		}

		switch r.Status.Value {
		case request.StatusNew, request.StatusActive, request.StatusInactive,
			request.StatusError:
		default:
			return errors.New(errors.ErrInvalidRequest,
				"invalid status",
				"resource", r)
		}
	}

	return nil
}

// ValidateCreate checks that the value contains valid data for creation.
func (r *Resource) ValidateCreate(cfg *config.Config) error {
	if !r.Name.Set {
		return errors.New(errors.ErrInvalidRequest,
			"missing name",
			"resource", r)
	}

	if !r.KeyField.Set {
		return errors.New(errors.ErrInvalidRequest,
			"missing key_field",
			"resource", r)
	}

	return r.Validate(cfg)
}

// ScanDest returns the destination fields for a SQL row scan.
func (r *Resource) ScanDest(options sqldb.FieldOptions) []any {
	dest := []any{
		&r.ResourceID,
		&r.Name,
		&r.Version,
		&r.Description,
		&r.Status,
		&r.StatusData,
		&r.KeyField,
		&r.KeyRegex,
		&r.ClearCondition,
		&r.ClearAfter,
		&r.ClearDelay,
		&r.Data,
		&r.Source,
		&r.CommitHash,
	}

	if options != nil && options.Contains(sqldb.OptUserDetails) {
		dest = append(dest,
			&r.CreatedAt,
			&r.CreatedBy,
			&r.UpdatedAt,
			&r.UpdatedBy,
		)
	}

	return dest
}

// resourceFields contain the search fields for resources.
var resourceFields = []*sqldb.Field{{
	Name:   "resource_key",
	Type:   sqldb.FieldInt,
	Table:  "resource",
	Hidden: true,
}, {
	Name:  "resource_id",
	Type:  sqldb.FieldString,
	Table: "resource",
}, {
	Name:    "name",
	Type:    sqldb.FieldString,
	Table:   "resource",
	Primary: true,
}, {
	Name:  "version",
	Type:  sqldb.FieldString,
	Table: "resource",
}, {
	Name:  "description",
	Type:  sqldb.FieldString,
	Table: "resource",
}, {
	Name:  "status",
	Type:  sqldb.FieldString,
	Table: "resource",
}, {
	Name:  "status_data",
	Type:  sqldb.FieldJSON,
	Table: "resource",
}, {
	Name:  "key_field",
	Type:  sqldb.FieldString,
	Table: "resource",
}, {
	Name:  "key_regex",
	Type:  sqldb.FieldString,
	Table: "resource",
}, {
	Name:  "clear_condition",
	Type:  sqldb.FieldString,
	Table: "resource",
}, {
	Name:  "clear_after",
	Type:  sqldb.FieldInt,
	Table: "resource",
}, {
	Name:  "clear_delay",
	Type:  sqldb.FieldInt,
	Table: "resource",
}, {
	Name:  "data",
	Type:  sqldb.FieldJSON,
	Table: "resource",
}, {
	Name:  "source",
	Type:  sqldb.FieldString,
	Table: "resource",
}, {
	Name:  "commit_hash",
	Type:  sqldb.FieldString,
	Table: "resource",
}, {
	Name:   "created_at",
	Type:   sqldb.FieldTime,
	Option: "user_details",
	Table:  "resource",
}, {
	Name:   "created_by",
	Type:   sqldb.FieldString,
	Option: "user_details",
	Table:  `"user"`,
}, {
	Name:   "updated_at",
	Type:   sqldb.FieldTime,
	Option: "user_details",
	Table:  "resource",
}, {
	Name:   "updated_by",
	Type:   sqldb.FieldString,
	Option: "user_details",
	Table:  `"user"`,
}}

// GetResources retrieves resources based on a search query.
func (s *Service) GetResources(ctx context.Context,
	query *search.Query,
	options sqldb.FieldOptions,
) ([]*Resource, []*sqldb.SummaryData, error) {
	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QuerySelect,
		Base:   sqldb.SearchFields("resource", resourceFields),
		Search: query.NoSummary(),
		Fields: resourceFields,
	})

	rows, err := q.Query(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, errors.ErrDatabase, "",
			"search", query)
	}

	keys, cacheKeys := []int64{}, []string{}

	index := map[string]int{}

	for rows.Next() {
		select {
		case <-ctx.Done():
			rows.Close()

			return nil, nil, errors.Context(ctx)
		default:
		}

		k, id := int64(0), ""

		if err = rows.Scan(&k, &id); err != nil {
			rows.Close()

			return nil, nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to select resource key",
				"search", query)
		}

		key := cache.KeyResource(id)

		cacheKeys = append(cacheKeys, key)
		index[key] = len(keys)
		keys = append(keys, k)
	}

	if err := rows.Err(); err != nil {
		rows.Close()

		return nil, nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to select resource key rows",
			"search", query)
	}

	rows.Close()

	res := make([]*Resource, len(index))

	sum := []*sqldb.SummaryData{}

	if len(res) == 0 {
		return res, sum, nil
	}

	if s.cache != nil && query != nil && query.Summary == "" {
		found := false

		cMap, err := s.cache.GetMulti(ctx, cacheKeys...)
		if err != nil && !errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to get resource cache keys",
				"error", err,
				"cache_keys", cacheKeys,
				"search", query)
		} else {
			for ck, ci := range cMap {
				if ci == nil {
					continue
				}

				var v *Resource

				buf := bytes.NewBuffer(ci.Value)

				if err := json.NewDecoder(buf).Decode(&v); err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to decode resource cache value",
						"error", err,
						"cache_key", ck,
						"cache_value", string(ci.Value),
						"search", query)
				}

				res[index[ck]] = v
				keys[index[ck]] = 0
				found = true
			}
		}

		if found {
			newKeys := make([]int64, 0, len(keys))

			for _, k := range keys {
				if k != 0 {
					newKeys = append(newKeys, k)
				}
			}

			keys = newKeys
		}
	}

	if len(keys) > 0 {
		base := sqldb.SelectFields("resource", resourceFields, query, options) +
			`WHERE resource.resource_key = ANY($1::BIGINT[])`

		q = sqldb.NewQuery(&sqldb.QueryOptions{
			DB:     s.db,
			Type:   sqldb.QuerySelect,
			Base:   base,
			Fields: resourceFields,
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

			r := &Resource{}

			sr := &sqldb.SummaryData{}

			if query != nil && query.Summary != "" {
				if err = rows.Scan(sr.ScanDest(resourceFields,
					query)...); err != nil {
					return nil, nil, errors.Wrap(err, errors.ErrDatabase,
						"unable to select resource summary row",
						"search", query)
				}

				sum = append(sum, sr)

				continue
			}

			if err = rows.Scan(r.ScanDest(options)...); err != nil {
				return nil, nil, errors.Wrap(err, errors.ErrDatabase,
					"unable to select resource row",
					"search", query)
			}

			if s.cache != nil {
				ck := cache.KeyResource(r.ResourceID.Value)

				buf, err := json.Marshal(r)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to encode resource cache value",
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
							"unable to set resource cache value",
							"error", err,
							"cache_key", ck,
							"cache_value", string(buf),
							"expiration", s.cfg.CacheExpiration(),
							"search", query)
					}
				}
			}

			res[index[cache.KeyResource(r.ResourceID.Value)]] = r
		}

		if err := rows.Err(); err != nil {
			return nil, nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to select resource rows",
				"search", query)
		}
	}

	if len(sum) > 0 {
		res = []*Resource{}
	}

	return res, sum, nil
}

// GetResource retrieves a single resource by ID.
func (s *Service) GetResource(ctx context.Context,
	id string,
	options sqldb.FieldOptions,
) (*Resource, error) {
	var r *Resource

	if s.cache != nil {
		ck := cache.KeyResource(id)

		ci, err := s.cache.Get(ctx, ck)
		if err != nil && !errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to get resource cache key",
				"error", err,
				"cache_key", ck,
				"id", id)
		} else if ci != nil {
			buf := bytes.NewBuffer(ci.Value)

			if err := json.NewDecoder(buf).Decode(&r); err != nil {
				s.log.Log(ctx, logger.LvlError,
					"unable to decode resource cache value",
					"error", err,
					"cache_key", ck,
					"cache_value", string(ci.Value),
					"id", id)
			}
		}
	}

	if r == nil {
		base := sqldb.SelectFields("resource", resourceFields, nil, options) +
			`WHERE resource.resource_id = $1`

		q := sqldb.NewQuery(&sqldb.QueryOptions{
			DB:     s.db,
			Type:   sqldb.QuerySelect,
			Base:   base,
			Fields: resourceFields,
			Params: []any{id},
		})

		q.Limit = 1

		row, err := q.QueryRow(ctx)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase, "", "id", id)
		}

		r = &Resource{}

		if err := row.Scan(r.ScanDest(options)...); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New(errors.ErrNotFound,
					"resource not found",
					"id", id)
			}

			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to select resource row",
				"id", id)
		}

		if s.cache != nil {
			ck := cache.KeyResource(r.ResourceID.Value)

			buf, err := json.Marshal(r)
			if err != nil {
				s.log.Log(ctx, logger.LvlError,
					"unable to encode resource cache value",
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
						"unable to set resource cache value",
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

// CreateResource creates a new resource.
func (s *Service) CreateResource(ctx context.Context,
	v *Resource,
) (*Resource, error) {
	userID, err := request.ContextUserID(ctx)
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, errors.New(errors.ErrInvalidRequest,
			"missing resource",
			"resource", v)
	}

	if err := v.ValidateCreate(s.cfg); err != nil {
		return nil, err
	}

	if v.ResourceID.Value == "" {
		uID, err := uuid.NewRandom()
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrServer,
				"unable to create ID for resource")
		}

		v.ResourceID = request.FieldString{
			Set: true, Valid: true, Value: uID.String(),
		}
	}

	base := `INSERT INTO resource () VALUES ()` +
		sqldb.ReturningFields("resource", resourceFields, nil)

	sets, params := []string{}, []any{}

	request.SetField("resource_id", v.ResourceID, &sets, &params)
	request.SetField("name", v.Name, &sets, &params)
	request.SetField("version", v.Version, &sets, &params)
	request.SetField("description", v.Description, &sets, &params)
	request.SetField("status", v.Status, &sets, &params)
	request.SetField("status_data", v.StatusData, &sets, &params)
	request.SetField("key_field", v.KeyField, &sets, &params)
	request.SetField("key_regex", v.KeyRegex, &sets, &params)
	request.SetField("clear_condition", v.ClearCondition, &sets, &params)
	request.SetField("clear_after", v.ClearAfter, &sets, &params)
	request.SetField("clear_delay", v.ClearDelay, &sets, &params)
	request.SetField("data", v.Data, &sets, &params)
	request.SetField("source", v.Source, &sets, &params)
	request.SetField("commit_hash", v.CommitHash, &sets, &params)
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
		Fields: resourceFields,
		Sets:   sets,
		Params: params,
	})

	row, err := q.QueryRow(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase, "", "resource", v)
	}

	r := &Resource{}

	if err := row.Scan(r.ScanDest(nil)...); err != nil {
		if errors.ErrorHas(err, `"resource_account_id_resource_id_key"`) {
			return nil, errors.New(errors.ErrConflict,
				"invalid resource_id: already in use by another resource",
				"resource", v)
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to insert resource row",
			"resource", v)
	}

	if s.cache != nil {
		ck := cache.KeyResource(r.ResourceID.Value)

		if err := s.cache.Delete(ctx, ck); err != nil &&
			!errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to delete resource cache key",
				"error", err,
				"cache_key", ck,
				"resource", v)
		}
	}

	return r, nil
}

// UpdateResource updates an resource.
func (s *Service) UpdateResource(ctx context.Context,
	v *Resource,
) (*Resource, error) {
	userID, err := request.ContextUserID(ctx)
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, errors.New(errors.ErrInvalidRequest,
			"missing resource",
			"resource", v)
	}

	if !v.ResourceID.Set || !v.ResourceID.Valid {
		return nil, errors.New(errors.ErrInvalidRequest,
			"missing resource_id",
			"resource", v)
	}

	if err := v.Validate(s.cfg); err != nil {
		return nil, err
	}

	base := `UPDATE resource SET
		WHERE resource.resource_id = $1` +
		sqldb.ReturningFields("resource", resourceFields, nil)

	sets, params := []string{}, []any{v.ResourceID.Value}

	request.SetField("name", v.Name, &sets, &params)
	request.SetField("version", v.Version, &sets, &params)
	request.SetField("description", v.Description, &sets, &params)
	request.SetField("status", v.Status, &sets, &params)
	request.SetField("status_data", v.StatusData, &sets, &params)
	request.SetField("key_field", v.KeyField, &sets, &params)
	request.SetField("key_regex", v.KeyRegex, &sets, &params)
	request.SetField("clear_condition", v.ClearCondition, &sets, &params)
	request.SetField("clear_after", v.ClearAfter, &sets, &params)
	request.SetField("clear_delay", v.ClearDelay, &sets, &params)
	request.SetField("data", v.Data, &sets, &params)
	request.SetField("source", v.Source, &sets, &params)
	request.SetField("commit_hash", v.CommitHash, &sets, &params)
	request.SetField("updated_at", request.FieldTime{
		Set: true, Valid: true, Value: time.Now().Unix(),
	}, &sets, &params)

	if userID == request.SystemUser {
		request.SetField("updated_by", request.FieldString{
			Set: true, Valid: false,
		}, &sets, &params)
	} else {
		request.SetField("updated_by", request.FieldString{
			Set: true, Valid: true, Value: userID,
		}, &sets, &params)
	}

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QueryUpdate,
		Base:   base,
		Fields: resourceFields,
		Sets:   sets,
		Params: params,
	})

	row, err := q.QueryRow(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase, "",
			"resource", v)
	}

	r := &Resource{}

	if err := row.Scan(r.ScanDest(nil)...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(errors.ErrNotFound,
				"resource not found",
				"resource", v)
		}

		if errors.ErrorHas(err, `"resource_account_id_resource_id_key"`) {
			return nil, errors.New(errors.ErrConflict,
				"invalid resource_id: already in use by another resource",
				"resource", v)
		}

		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to update resource row",
			"resource", v)
	}

	if s.cache != nil {
		ck := cache.KeyResource(r.ResourceID.Value)

		if err := s.cache.Delete(ctx, ck); err != nil &&
			!errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to delete resource cache key",
				"error", err,
				"cache_key", ck,
				"resource", v)
		}
	}

	return r, nil
}

// DeleteResource deletes an resource.
func (s *Service) DeleteResource(ctx context.Context,
	id string,
) error {
	if s.cache != nil {
		defer func(ck string) {
			if err := s.cache.Delete(ctx, ck); err != nil &&
				!errors.Has(err, errors.ErrNotFound) {
				s.log.Log(ctx, logger.LvlError,
					"unable to delete resource cache key",
					"error", err,
					"cache_key", ck,
					"id", id)
			}
		}(cache.KeyResource(id))
	}

	base := `DELETE FROM resource
		WHERE resource.resource_id = $1`

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QueryDelete,
		Base:   base,
		Fields: resourceFields,
		Params: []any{id},
	})

	res, err := q.Exec(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase, "", "id", id)
	}

	if n := res.RowsAffected(); n == 0 {
		return errors.New(errors.ErrNotFound, "resource not found",
			"id", id)
	}

	return nil
}

// findResourceData is used to create a keyed map of resource data values from
// an existing resource and an resource update payload.
func findResourceData(payload map[string]any,
	resource *Resource,
) (map[string]any, []string, error) {
	if resource.KeyField.Value == "" {
		return nil, nil, errors.New(errors.ErrInvalidRequest,
			"unable to extract resource data: missing key field",
			"resource", resource,
			"payload", payload)
	}

	resourceData := map[string]any{}

	clears := []string{}

	resources, ok := payload["resources"].([]any)
	if !ok {
		resources = []any{payload}
	}

	for _, ad := range resources {
		am, ok := ad.(map[string]any)
		if !ok {
			continue
		}

		kv, ok := am[resource.KeyField.Value]
		if !ok {
			continue
		}

		key := ""

		switch kt := kv.(type) {
		case string:
			key = kt
		default:
			if b, err := json.Marshal(kt); err != nil {
				continue
			} else {
				key = string(b)
			}
		}

		if resource.KeyRegex.Value != "" {
			re, err := regexp.Compile(resource.KeyRegex.Value)
			if err != nil {
				return nil, nil, errors.Wrap(err, errors.ErrInvalidRequest,
					"invalid resource key_regex",
					"resource", resource,
					"payload", payload)
			}

			key = re.FindString(key)
		}

		if key != "" {
			am["ts"] = time.Now().Unix()

			cleared := false

			if resource.ClearCondition.Value != "" {
				p := search.NewParser(bytes.NewBufferString(
					resource.ClearCondition.Value))

				ast, err := p.Parse()
				if err != nil {
					return nil, nil, errors.Wrap(err, errors.ErrInvalidRequest,
						"invalid resource clear_condition",
						"resource", resource,
						"payload", payload)
				}

				cleared, err = ast.Eval(
					func(node *search.QueryNode) (bool, error) {
						getValInt64 := func(cat, val string) (int64, error) {
							r, err := strconv.ParseInt(val, 10, 64)
							if err != nil {
								return 0, errors.Wrap(err,
									errors.ErrInvalidRequest,
									"invalid condition value for category",
									"category", cat,
									"value", val)
							}

							return r, nil
						}

						getValFloat64 := func(cat string,
							val string,
						) (float64, error) {
							r, err := strconv.ParseFloat(val, 64)
							if err != nil {
								return 0, errors.Wrap(err,
									errors.ErrInvalidRequest,
									"invalid condition value for category",
									"category", cat,
									"value", val)
							}

							return r, nil
						}

						op := node.Comp

						cat := strings.TrimSpace(node.Cat)

						val := strings.TrimSpace(node.Val)

						valRegExp := node.ValRE

						var valRE *regexp.Regexp

						if valRegExp == "" && strings.Contains(val, "*") {
							valRegExp = strings.ReplaceAll(val, "*", ".*")
						}

						if valRegExp != "" {
							val = valRegExp

							re, err := regexp.Compile(val)
							if err != nil {
								return false, errors.Wrap(err,
									errors.ErrInvalidRequest,
									"invalid condition value "+
										"regular expression",
									"value", val)
							}

							valRE = re
						}

						parts := strings.Split(cat, ".")

						if parts[0] == "true" {
							return true, nil
						}

						res := false

						var v any = am

						for i := 0; i < len(parts); i++ {
							key := parts[i]

							index := -1

							if strings.HasSuffix(key, "]") {
								startIdx := strings.LastIndex(key, "[")

								endIdx := strings.LastIndex(key, "]")

								if startIdx > 0 && endIdx > 0 &&
									startIdx < endIdx {
									idx := key[startIdx+1 : endIdx]

									iv, err := strconv.ParseInt(idx, 10, 64)
									if err == nil {
										index = int(iv)

										key = key[:startIdx]
									}
								}
							}

							switch vt := v.(type) {
							case map[string]any:
								vv, ok := vt[key]
								if !ok {
									return false, nil
								}

								switch vvt := vv.(type) {
								case []any:
									if index < 0 || index >= len(vvt) {
										return false, nil
									}

									v = vvt[index]
								case []map[string]any:
									if index < 0 || index >= len(vvt) {
										return false, nil
									}

									v = vvt[index]
								default:
									v = vvt
								}
							default:
								return false, nil
							}
						}

						switch vt := v.(type) {
						case nil:
							if val == "null" || val == "" {
								res = true
							}
						case float64:
							l := vt

							r, err := getValFloat64(cat, val)
							if err != nil {
								return false, err
							}

							switch op {
							case search.OpMatch:
								if l == r {
									res = true
								}
							case search.OpGT:
								if l > r {
									res = true
								}
							case search.OpGTE:
								if l >= r {
									res = true
								}
							case search.OpLT:
								if l < r {
									res = true
								}
							case search.OpLTE:
								if l <= r {
									res = true
								}
							default:
								return false, errors.New(
									errors.ErrInvalidRequest,
									"invalid condition operator for category",
									"category", cat,
									"operator", op)
							}
						case int64:
							l := vt

							r, err := getValInt64(cat, val)
							if err != nil {
								return false, err
							}

							switch op {
							case search.OpMatch:
								if l == r {
									res = true
								}
							case search.OpGT:
								if l > r {
									res = true
								}
							case search.OpGTE:
								if l >= r {
									res = true
								}
							case search.OpLT:
								if l < r {
									res = true
								}
							case search.OpLTE:
								if l <= r {
									res = true
								}
							default:
								return false, errors.New(
									errors.ErrInvalidRequest,
									"invalid condition operator for category",
									"category", cat,
									"operator", op)
							}
						case string:
							if valRE != nil {
								if valRE.MatchString(vt) {
									res = true
								}
							} else {
								res, err = filepath.Match(val, vt)
								if err != nil {
									return false, errors.Wrap(err,
										errors.ErrInvalidRequest,
										"invalid value pattern for category",
										"category", cat,
										"value", val,
										"operator", op)
								}
							}
						default:
							return false, nil
						}

						return res, nil
					})
				if err != nil {
					return nil, nil, errors.Wrap(err, errors.ErrInvalidRequest,
						"unable to evaluate resource clear_condition",
						"resource", resource,
						"payload", payload)
				}

				if cleared {
					clears = append(clears, key)
				}
			}

			if !cleared {
				resourceData[key] = am
			}
		}
	}

	return resourceData, clears, nil
}

// UpdateResourceData allows external systems to update resource data.
func (s *Service) UpdateResourceData(ctx context.Context,
	payload map[string]any,
	accountID, resourceID string,
) (*Resource, error) {
	ctx = context.WithValue(ctx, request.CtxKeyUserID, request.SystemUser)
	ctx = context.WithValue(ctx, request.CtxKeyScopes, request.ScopeSuperuser)
	ctx = context.WithValue(ctx, request.CtxKeyAccountID, accountID)

	r, err := s.GetResource(ctx, resourceID, nil)
	if err != nil {
		return nil, err
	}

	if r.Status.Value == request.StatusInactive {
		return nil, errors.New(errors.ErrInvalidRequest,
			"unable to update resource data for inactive resource",
			"payload", payload,
			"resource", r)
	}

	resourceData, clears, err := findResourceData(payload, r)
	if err != nil {
		r.Status = request.FieldString{
			Set: true, Valid: true, Value: request.StatusError,
		}

		r.StatusData = request.FieldJSON{
			Set: true, Valid: true, Value: map[string]any{
				"last_error": err.Error(),
			},
		}

		if _, err := s.UpdateResource(ctx, r); err != nil {
			s.log.Log(ctx, logger.LvlError,
				"unable to update resource error status",
				"error", err,
				"resource", r)
		}

		return nil, err
	}

	if !r.Data.Set || !r.Data.Valid || len(r.Data.Value) == 0 {
		r.Data = request.FieldJSON{
			Set: true, Valid: true, Value: map[string]any{},
		}
	}

	for k, v := range resourceData {
		r.Data.Value[k] = v
	}

	// Remove any cleared resources.
	for _, key := range clears {
		delete(r.Data.Value, key)
	}

	// Prune any data older than the clear_after setting.
	newData := map[string]any{}

	oldTS := time.Now().Add(0 -
		(time.Second * time.Duration(r.ClearAfter.Value))).Unix()

	for k, v := range r.Data.Value {
		if vv, ok := v.(map[string]any); ok {
			if pts, ok := vv["ts"]; ok {
				switch tv := pts.(type) {
				case int64:
					if tv >= oldTS {
						newData[k] = v
					}
				case float64:
					if int64(tv) >= oldTS {
						newData[k] = v
					}
				case string:
					if iv, err := strconv.ParseInt(tv, 10, 64); err == nil {
						if iv >= oldTS {
							newData[k] = v
						}
					}
				}
			}
		}
	}

	r.Data.Value = newData

	if len(r.Data.Value) == 0 {
		r.Data.Valid = false
	}

	r.Status = request.FieldString{
		Set: true, Valid: true, Value: request.StatusActive,
	}

	res, err := s.UpdateResource(ctx, r)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// UpdateResourceError allows external systems to update resource error status.
func (s *Service) UpdateResourceError(ctx context.Context,
	accountID, resourceID string,
	resourceError error,
) error {
	ctx = context.WithValue(ctx, request.CtxKeyUserID, request.SystemUser)

	ctx = context.WithValue(ctx, request.CtxKeyScopes, request.ScopeSuperuser)

	ctx = context.WithValue(ctx, request.CtxKeyAccountID, accountID)

	if resourceError != nil {
		if _, err := s.UpdateResource(ctx, &Resource{
			ResourceID: request.FieldString{
				Set: true, Valid: true, Value: resourceID,
			},
			Status: request.FieldString{
				Set: true, Valid: true, Value: request.StatusError,
			},
			StatusData: request.FieldJSON{
				Set: true, Valid: true, Value: map[string]any{
					"last_error": resourceError.Error(),
				},
			},
		}); err != nil {
			return errors.Wrap(err, errors.ErrDatabase,
				"unable to update resource error status",
				"account_id", accountID,
				"resource_id", resourceID,
				"error", resourceError)
		}
	}

	return nil
}

// ImportResource loads and updates a single resource from the import
// repository.
func (s *Service) ImportResource(ctx context.Context,
	authSvc AuthService,
	resourceID string,
) error {
	ctx = context.WithValue(ctx, request.CtxKeyUserID, request.SystemUser)
	ctx = context.WithValue(ctx, request.CtxKeyScopes, request.ScopeSuperuser)

	ar, err := authSvc.GetAccountRepo(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to get account repository")
	}

	cli, err := s.getRepoClient(ar.Repo.Value)
	if err != nil {
		return errors.Wrap(err, errors.ErrImport,
			"unable to create repository client")
	}

	newHash, err := cli.Commit(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrImport,
			"unable to get repository commit hash")
	}

	vb, err := cli.Get(ctx, "resources/"+resourceID+".yaml")
	if err != nil {
		return errors.Wrap(err, errors.ErrImport,
			"unable to get resource repository file",
			"resource_id", resourceID)
	}

	a := &Resource{}

	if err := yaml.Unmarshal(vb, &a); err != nil {
		return errors.Wrap(err,
			errors.ErrImport,
			"invalid repository resource contents",
			"resource_id", resourceID,
			"contents", string(vb))
	}

	a.ResourceID = request.FieldString{
		Set: true, Valid: true, Value: resourceID,
	}

	a.Version = request.FieldString{
		Set: true, Valid: true, Value: newHash,
	}

	a.Status = request.FieldString{
		Set: true, Valid: true, Value: request.StatusActive,
	}

	a.Source = request.FieldString{
		Set: true, Valid: true, Value: "git",
	}

	a.CommitHash = request.FieldString{
		Set: true, Valid: true, Value: newHash,
	}

	if _, err := s.CreateResource(ctx, a); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to create imported resource",
			"resource", a)
	}

	return nil
}

// ImportResources loads and updates resource data.
func (s *Service) ImportResources(ctx context.Context,
	force bool,
	authSvc AuthService,
) error {
	ctx = context.WithValue(ctx, request.CtxKeyUserID, request.SystemUser)
	ctx = context.WithValue(ctx, request.CtxKeyScopes, request.ScopeSuperuser)

	ar, err := authSvc.GetAccountRepo(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to get account repository")
	}

	if !force && ar.RepoStatus.Value == request.StatusImporting {
		if pli, ok := ar.RepoStatusData.Value["resources_last_imported"]; ok {
			if i, ok := pli.(int64); ok && i > time.Now().Unix()-120 {
				return errors.Wrap(err, errors.ErrImport,
					"unable to import resources, another import in progress")
			}
		}
	}

	ar.RepoStatus = request.FieldString{
		Set: true, Valid: true, Value: request.StatusImporting,
	}

	dm := ar.RepoStatusData.Value

	if dm == nil {
		dm = map[string]any{}
	}

	dm["resources_last_imported"] = time.Now().Unix()

	ar.RepoStatusData = request.FieldJSON{
		Set: true, Valid: true, Value: dm,
	}

	if err := authSvc.SetAccountRepo(ctx, ar); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to set account repository status")
	}

	updated, deleted, uErr := s.updateResources(ctx, ar, force)

	ar, err = authSvc.GetAccountRepo(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to get account repository")
	}

	ar.RepoStatus = request.FieldString{
		Set: true, Valid: true, Value: request.StatusActive,
	}

	dm = ar.RepoStatusData.Value

	if dm == nil {
		dm = map[string]any{}
	}

	dm["resources_updated"] = updated

	dm["resources_deleted"] = deleted

	if uErr != nil {
		ar.RepoStatus.Value = request.StatusError

		dm["resources_last_error"] = uErr.Error()
	} else {
		delete(dm, "resources_last_error")
	}

	ar.RepoStatusData = request.FieldJSON{
		Set: true, Valid: true, Value: dm,
	}

	if err := authSvc.SetAccountRepo(ctx, ar); err != nil {
		return errors.Wrap(err, errors.ErrDatabase,
			"unable to set account repository status")
	}

	if uErr != nil {
		return uErr
	}

	return nil
}

// updateResources updates the resources based on the contents of the account
// import repository.
func (s *Service) updateResources(ctx context.Context,
	ar *auth.AccountRepo,
	force bool,
) (int, int, error) {
	ctx, cancel := request.ContextReplaceTimeout(ctx, s.cfg.ServerTimeout())

	defer cancel()

	cli, err := s.getRepoClient(ar.Repo.Value)
	if err != nil {
		return 0, 0, errors.Wrap(err, errors.ErrImport,
			"unable to create repository client")
	}

	newHash, err := cli.Commit(ctx)
	if err != nil {
		return 0, 0, errors.Wrap(err, errors.ErrImport,
			"unable to get repository commit hash")
	}

	ch, err := s.getAccountResourceCommitHash(ctx)
	if err != nil {
		return 0, 0, errors.Wrap(err, errors.ErrImport,
			"unable to get account commit_hash")
	}

	if !force && ch == newHash {
		s.log.Log(ctx, logger.LvlDebug,
			"resource import completed, commit unchanged",
			"updated", 0,
			"deleted", 0)

		return 0, 0, nil
	}

	res, err := cli.ListAll(ctx, "resources/")
	if err != nil {
		return 0, 0, errors.Wrap(err, errors.ErrImport,
			"unable to list repository path",
			"path", "resources/")
	}

	updated := 0

	errs := errors.New(errors.ErrImport,
		"unable to import resources")

	for _, i := range res {
		if i.Type == "file" || i.Type == "commit_file" {
			ctx, cancel := request.ContextReplaceTimeout(ctx,
				s.cfg.ServerTimeout())

			defer cancel()

			resourceID := strings.TrimPrefix(strings.TrimPrefix(i.Path, "/"),
				"resources/")

			ext := filepath.Ext(resourceID)

			resourceID = strings.TrimSuffix(resourceID, ext)

			a, err := s.GetResource(ctx, resourceID, nil)
			if err != nil && !errors.Has(err, errors.ErrNotFound) {
				errs.Errors = append(errs.Errors, errors.Wrap(err,
					errors.ErrDatabase,
					"unable to get current resource",
					"resource_id", resourceID))

				continue
			}

			if a != nil && (!force && a.Version.Value == i.Commit) {
				if a.CommitHash.Value != newHash {
					a.CommitHash = request.FieldString{
						Set: true, Valid: true, Value: newHash,
					}

					if _, err := s.UpdateResource(ctx, a); err != nil {
						errs.Errors = append(errs.Errors, errors.Wrap(err,
							errors.ErrDatabase,
							"unable to update repository resource",
							"resource", a))

						continue
					}

					updated++
				}

				continue
			}

			vb, err := cli.Get(ctx, "resources/"+resourceID+ext)
			if err != nil {
				errs.Errors = append(errs.Errors, errors.Wrap(err,
					errors.ErrImport,
					"unable to get resource repository file",
					"resource_id", resourceID))

				continue
			}

			m := map[string]any{}

			if err := yaml.Unmarshal(vb, &m); err != nil {
				errs.Errors = append(errs.Errors, errors.Wrap(err,
					errors.ErrImport,
					"unable to parse resource repository file",
					"resource_id", resourceID))

				continue
			}

			vmb, err := json.Marshal(&m)
			if err != nil {
				errs.Errors = append(errs.Errors, errors.Wrap(err,
					errors.ErrImport,
					"unable to format resource repository file map",
					"resource_id", resourceID))

				continue
			}

			if err := json.Unmarshal(vmb, &a); err != nil {
				errs.Errors = append(errs.Errors, errors.Wrap(err,
					errors.ErrImport,
					"invalid repository resource contents",
					"resource_id", resourceID,
					"contents", string(vmb)))

				continue
			}

			a.ResourceID = request.FieldString{
				Set: true, Valid: true, Value: resourceID,
			}

			a.Version = request.FieldString{
				Set: true, Valid: true, Value: newHash,
			}

			a.Status = request.FieldString{
				Set: true, Valid: true, Value: request.StatusActive,
			}

			a.Source = request.FieldString{
				Set: true, Valid: true, Value: "git",
			}

			a.CommitHash = request.FieldString{
				Set: true, Valid: true, Value: newHash,
			}

			if _, err := s.CreateResource(ctx, a); err != nil {
				errs.Errors = append(errs.Errors, errors.Wrap(err,
					errors.ErrDatabase,
					"unable to create imported resource",
					"resource", a))

				continue
			}

			updated++
		}
	}

	if len(errs.Errors) > 0 {
		s.log.Log(ctx, logger.LvlWarn,
			"unable to complete resource import",
			"updated", updated,
			"errors", errs.Errors)

		return updated, 0, errs
	}

	ctx, cancel = request.ContextReplaceTimeout(ctx, s.cfg.ServerTimeout())

	defer cancel()

	deleted := 0

	if newHash != "" {
		err := s.setAccountResourceCommitHash(ctx, newHash)
		if err != nil {
			errs.Errors = append(errs.Errors, errors.Wrap(err,
				errors.ErrDatabase,
				"unable to set account resource_commit_hash"))
		} else {
			deleted, err = s.deleteResources(ctx, newHash)
			if err != nil {
				errs.Errors = append(errs.Errors, errors.Wrap(err,
					errors.ErrDatabase,
					"unable to delete removed repository resources",
					"commit_hash", newHash))
			}
		}
	}

	if len(errs.Errors) > 0 {
		s.log.Log(ctx, logger.LvlWarn,
			"unable to complete resource import",
			"updated", updated,
			"deleted", deleted,
			"errors", errs.Errors)

		return updated, deleted, errs
	}

	s.log.Log(ctx, logger.LvlInfo,
		"resource import completed",
		"updated", updated,
		"deleted", deleted)

	return updated, deleted, nil
}

// deleteResources deletes all resources not in the specified list.
func (s *Service) deleteResources(ctx context.Context,
	commit string,
) (int, error) {
	base := `DELETE FROM resource
		WHERE source = 'git' AND commit_hash <> $1::TEXT
		RETURNING resource_id`

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QueryDelete,
		Base:   base,
		Fields: resourceFields,
		Params: []any{commit},
	})

	rows, err := q.Query(ctx)
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDatabase, "",
			"commit_hash", commit)
	}

	defer rows.Close()

	count := 0

	for rows.Next() {
		select {
		case <-ctx.Done():
			return 0, errors.Context(ctx)
		default:
		}

		dID := ""

		if err := rows.Scan(&dID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				break
			}

			return count, errors.Wrap(err, errors.ErrDatabase,
				"unable to select deleted resource_id",
				"commit_hash", commit)
		}

		if s.cache != nil && dID != "" {
			ck := cache.KeyResource(dID)

			if err := s.cache.Delete(ctx, ck); err != nil &&
				!errors.Has(err, errors.ErrNotFound) {
				s.log.Log(ctx, logger.LvlError,
					"unable to delete resource cache key",
					"error", err,
					"cache_key", ck,
					"resource_id", dID)
			}
		}

		count++
	}

	return count, nil
}

// getAccountResourceCommitHash retrieves the current account commit hash.
func (s *Service) getAccountResourceCommitHash(ctx context.Context,
) (string, error) {
	base := `SELECT resource_commit_hash FROM account
		LIMIT 1`

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:   s.db,
		Type: sqldb.QuerySelect,
		Base: base,
		Fields: []*sqldb.Field{{
			Name:  "resource_commit_hash",
			Type:  sqldb.FieldString,
			Table: "account",
		}},
	})

	row, err := q.QueryRow(ctx)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrDatabase, "")
	}

	var ch *string

	if err := row.Scan(&ch); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return "", errors.Wrap(err, errors.ErrDatabase,
				"unable to select account resource_commit_hash")
		}
	}

	if ch == nil {
		return "", nil
	}

	return *ch, nil
}

// setAccountResourceCommitHash sets the current account commit hash.
func (s *Service) setAccountResourceCommitHash(ctx context.Context,
	commit string,
) error {
	base := `UPDATE account SET resource_commit_hash = $1
	RETURNING resource_commit_hash`

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QueryUpdate,
		Base:   base,
		Params: []any{commit},
		Fields: []*sqldb.Field{{
			Name:  "resource_commit_hash",
			Type:  sqldb.FieldString,
			Table: "account",
		}},
	})

	row, err := q.QueryRow(ctx)
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabase, "")
	}

	var ch *string

	if err := row.Scan(&ch); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return errors.Wrap(err, errors.ErrDatabase,
				"unable to set account resource_commit_hash")
		}
	}

	return nil
}

// Update periodically imports resources data.
func (s *Service) Update(ctx context.Context,
	authSvc AuthService,
) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)

	go func(ctx context.Context) {
		tick := time.NewTimer(0)

		adj := time.Duration(0)

		retries := 0

		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				accounts, err := s.getAllAccounts(ctx)
				if err != nil {
					s.log.Log(ctx, logger.LvlError,
						"unable to get accounts to update resources",
						"error", err)

					break
				}

				var wg sync.WaitGroup

				for _, aID := range accounts {
					wg.Add(1)

					go func(ctx context.Context, accountID string) {
						ctx = context.WithValue(ctx, request.CtxKeyAccountID,
							accountID)
						ctx = context.WithValue(ctx, request.CtxKeyUserID,
							request.SystemUser)
						ctx = context.WithValue(ctx, request.CtxKeyScopes,
							request.ScopeSuperuser)

						if tu, err := uuid.NewRandom(); err == nil {
							ctx = context.WithValue(ctx, request.CtxKeyTraceID,
								tu.String())
						}

						if err := s.ImportResources(ctx, false,
							authSvc); err != nil {
							lvl := logger.LvlError

							if errors.ErrorHas(err,
								"another import in progress") {
								lvl = logger.LvlDebug
							}

							s.log.Log(ctx, lvl,
								"unable to import resources",
								"error", err)

							adj = s.cfg.ImportInterval()*
								time.Duration(retries) +
								time.Duration(float64(
									s.cfg.ImportInterval())*rand.Float64())

							retries++

							if retries > 10 {
								retries = 10
							}
						} else {
							retries = 0
						}

						wg.Done()
					}(ctx, aID)
				}

				wg.Wait()
			}

			tick = time.NewTimer(s.cfg.ImportInterval() + adj)

			adj = 0
		}
	}(ctx)

	return cancel
}

// getAllAccounts retrieves a list of all active account ID's.
func (s *Service) getAllAccounts(ctx context.Context) ([]string, error) {
	ctx = context.WithValue(ctx, request.CtxKeyAccountID, request.SystemAccount)

	base := `SELECT account.account_id
	FROM account
	WHERE status = '` + request.StatusActive + `'`

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:   s.db,
		Type: sqldb.QuerySelect,
		Base: base,
	})

	q.Limit = 10000

	rows, err := q.Query(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"")
	}

	defer rows.Close()

	res := []string{}

	for rows.Next() {
		select {
		case <-ctx.Done():
			return nil, errors.Context(ctx)
		default:
		}

		r := ""

		if err = rows.Scan(&r); err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to select account row")
		}

		res = append(res, r)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to select account rows")
	}

	return res, nil
}
