package resource

import (
	"bytes"
	"context"
	"strings"

	"github.com/dhaifley/apid/internal/cache"
	"github.com/dhaifley/apid/internal/errors"
	"github.com/dhaifley/apid/internal/logger"
	"github.com/dhaifley/apid/internal/request"
	"github.com/dhaifley/apid/internal/search"
	"github.com/dhaifley/apid/internal/sqldb"
)

// TagMap values contain a map of tag values keyed by tag categories.
type TagMap map[string][]string

// GetTags retrieves all resource tags and tag values.
func (s *Service) GetTags(ctx context.Context,
) (TagMap, error) {
	if _, err := request.ContextAuthUser(ctx); err != nil {
		return nil, err
	}

	base := `SELECT tag_obj.tag_key || ':' || tag_obj.tag_val AS tag
		FROM tag_obj
		WHERE tag_obj.status = '` + request.StatusActive + `'
			AND tag_obj.tag_type = 'resource'`

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:   s.db,
		Type: sqldb.QuerySelect,
		Base: base,
	})

	rows, err := q.Query(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase, "")
	}

	defer rows.Close()

	res := TagMap{}

	for rows.Next() {
		select {
		case <-ctx.Done():
			return nil, errors.Context(ctx)
		default:
		}

		r := ""

		if err = rows.Scan(&r); err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to select resource tags row")
		}

		parts := strings.SplitN(r, ":", 2)

		if len(parts) < 1 {
			continue
		}

		if _, ok := res[parts[0]]; !ok {
			res[parts[0]] = []string{}
		}

		if len(parts) > 1 && parts[1] != "" {
			res[parts[0]] = append(res[parts[0]], parts[1])
		}
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to select resource tags rows")
	}

	return res, nil
}

// GetResourceTags retrieves all tags assigned to an resource by ID.
func (s *Service) GetResourceTags(ctx context.Context,
	resourceID string,
) ([]string, error) {
	if _, err := request.ContextAuthUser(ctx); err != nil {
		return nil, err
	}

	base := `SELECT tag_obj.tag_key || ':' || tag_obj.tag_val AS tag
		FROM tag_obj
		WHERE tag_obj.status = '` + request.StatusActive + `'
			AND tag_obj.tag_type = 'resource'
			AND tag_obj_id = $1`

	q := sqldb.NewQuery(&sqldb.QueryOptions{
		DB:     s.db,
		Type:   sqldb.QuerySelect,
		Base:   base,
		Params: []any{resourceID},
	})

	rows, err := q.Query(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase, "")
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
				"unable to select resource tag row")
		}

		res = append(res, r)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to select resource tag rows")
	}

	return res, nil
}

// AddResourceTags adds tags for a specified resource by ID, creating the tags if
// they do not already exist.
func (s *Service) AddResourceTags(ctx context.Context,
	resourceID string,
	tags []string,
) ([]string, error) {
	userID, err := request.ContextAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	res := []string{}

	tagsMap := map[string]struct{}{}

	for _, tag := range tags {
		if _, ok := tagsMap[tag]; !ok {
			tagsMap[tag] = struct{}{}

			tagEls := strings.SplitN(tag, ":", 2)

			tagKey, tagVal := "", ""

			if len(tagEls) > 0 {
				tagKey = tagEls[0]
			}

			if len(tagEls) > 1 {
				tagVal = tagEls[1]
			}

			base := `WITH insert_tag AS (
				INSERT INTO tag (tag_key, tag_val, created_by, updated_by)
				VALUES ($1, $2,
					(SELECT user_key FROM "user" WHERE user_id = $3),
					(SELECT user_key FROM "user" WHERE user_id = $3))
				ON CONFLICT (account_id, tag_key, tag_val) DO UPDATE SET
					status = '` + request.StatusActive + `',
					updated_at = CURRENT_TIMESTAMP,
					updated_by = (SELECT user_key FROM "user" WHERE user_id = $3)
				RETURNING tag_key, tag_val
			) INSERT INTO tag_obj (tag_type, tag_obj_id, tag_key, tag_val,
				created_by, updated_by) SELECT
					'resource' AS tag_type,
					$4::TEXT AS tag_obj_id,
					it.tag_key AS tag_key,
					it.tag_val AS tag_val,
					(SELECT user_key FROM "user" WHERE user_id = $3) AS created_by,
					(SELECT user_key FROM "user" WHERE user_id = $3) AS updated_by
				FROM insert_tag it
			ON CONFLICT (account_id, tag_type, tag_obj_id,
					tag_key, tag_val) DO UPDATE SET
				status = '` + request.StatusActive + `',
				updated_at = CURRENT_TIMESTAMP,
				updated_by = (SELECT user_key FROM "user" WHERE user_id = $3)
			RETURNING tag_key || ':' || tag_val AS tag`

			q := sqldb.NewQuery(&sqldb.QueryOptions{
				DB:     s.db,
				Type:   sqldb.QueryInsert,
				Base:   base,
				Params: []any{tagKey, tagVal, userID, resourceID},
			})

			rows, err := q.Query(ctx)
			if err != nil {
				return nil, errors.Wrap(err, errors.ErrDatabase, "")
			}

			defer rows.Close()

			for rows.Next() {
				select {
				case <-ctx.Done():
					return nil, errors.Context(ctx)
				default:
				}

				r := ""

				if err = rows.Scan(&r); err != nil {
					return nil, errors.Wrap(err, errors.ErrDatabase,
						"unable to insert resource tag row")
				}

				res = append(res, r)
			}

			if err := rows.Err(); err != nil {
				return nil, errors.Wrap(err, errors.ErrDatabase,
					"unable to insert resource tag rows")
			}
		}
	}

	if s.cache != nil && len(res) > 0 {
		ck := cache.KeyResource(resourceID)

		if err := s.cache.Delete(ctx, ck); err != nil &&
			!errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to delete resource cache key",
				"error", err,
				"cache_key", ck)
		}
	}

	return res, nil
}

// DeleteResourceTags removes tags for a specified resource by ID, removing the tags
// if they are no longer used by any objects.
func (s *Service) DeleteResourceTags(ctx context.Context,
	resourceID string,
	tags []string,
) error {
	if _, err := request.ContextAuthUser(ctx); err != nil {
		return err
	}

	res := []string{}

	for _, tag := range tags {
		tagEls := strings.SplitN(tag, ":", 2)

		tagKey, tagVal := "", ""

		if len(tagEls) > 0 {
			tagKey = tagEls[0]
		}

		if len(tagEls) > 1 {
			tagVal = tagEls[1]
		}

		base := `DELETE FROM tag_obj WHERE tag_type = 'resource'
			AND tag_obj_id = $1
			AND tag_key = $2
			AND tag_val = $3
		RETURNING tag_key || ':' || tag_val AS tag`

		q := sqldb.NewQuery(&sqldb.QueryOptions{
			DB:     s.db,
			Type:   sqldb.QueryDelete,
			Base:   base,
			Params: []any{resourceID, tagKey, tagVal},
		})

		rows, err := q.Query(ctx)
		if err != nil {
			return errors.Wrap(err, errors.ErrDatabase, "")
		}

		defer rows.Close()

		for rows.Next() {
			select {
			case <-ctx.Done():
				return errors.Context(ctx)
			default:
			}

			r := ""

			if err = rows.Scan(&r); err != nil {
				return errors.Wrap(err, errors.ErrDatabase,
					"unable to delete resource tag object row",
					"resource_id", resourceID)
			}

			res = append(res, r)
		}

		if err := rows.Err(); err != nil {
			return errors.Wrap(err, errors.ErrDatabase,
				"unable to delete resource tag object rows",
				"resource_id", resourceID)
		}

		base = `DELETE FROM tag WHERE tag_key = $1
			AND tag_val = $2
			AND NOT EXISTS(SELECT tag_key FROM tag_obj WHERE tag_key = $1
				AND tag_val = $2)
		RETURNING tag_key || ':' || tag_val AS tag`

		q = sqldb.NewQuery(&sqldb.QueryOptions{
			DB:     s.db,
			Type:   sqldb.QueryDelete,
			Base:   base,
			Params: []any{tagKey, tagVal},
		})

		rows, err = q.Query(ctx)
		if err != nil {
			return errors.Wrap(err, errors.ErrDatabase, "")
		}

		defer rows.Close()

		for rows.Next() {
			select {
			case <-ctx.Done():
				return errors.Context(ctx)
			default:
			}

			r := ""

			if err = rows.Scan(&r); err != nil {
				return errors.Wrap(err, errors.ErrDatabase,
					"unable to delete resource tag row",
					"resource_id", resourceID)
			}

			res = append(res, r)
		}

		if err := rows.Err(); err != nil {
			return errors.Wrap(err, errors.ErrDatabase,
				"unable to delete resource tag rows",
				"resource_id", resourceID)
		}
	}

	if s.cache != nil && len(res) > 0 {
		ck := cache.KeyResource(resourceID)

		if err := s.cache.Delete(ctx, ck); err != nil &&
			!errors.Has(err, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to delete resource cache key",
				"error", err,
				"cache_key", ck)
		}
	}

	return nil
}

// TagsMultiAssignment values represent assignment, or removal, of tags to
// multiple resources using an resource selector.
type TagsMultiAssignment struct {
	Tags             request.FieldStringArray `json:"tags"`
	ResourceSelector request.FieldString      `json:"resource_selector"`
}

// Validate checks that the value contains valid data.
func (tma *TagsMultiAssignment) Validate() error {
	if tma.Tags.Set && !tma.Tags.Valid {
		return errors.New(errors.ErrInvalidRequest,
			"tags must not be null",
			"tags_multi_assignment", tma)
	}

	if tma.ResourceSelector.Set {
		if !tma.ResourceSelector.Valid {
			return errors.New(errors.ErrInvalidRequest,
				"resource_selector must not be null",
				"tags_multi_assignment", tma)
		}

		p := search.NewParser(bytes.NewBufferString(tma.ResourceSelector.Value))

		if _, err := p.Parse(); err != nil {
			return errors.Wrap(err, errors.ErrInvalidRequest,
				"invalid resource_selector",
				"tags_multi_assignment", tma)
		}
	}

	return nil
}

// ValidateCreate checks that the value contains valid data for creation.
func (tma *TagsMultiAssignment) ValidateCreate() error {
	if !tma.Tags.Set {
		return errors.New(errors.ErrInvalidRequest,
			"missing tags",
			"tags_multi_assignment", tma)
	}

	if !tma.ResourceSelector.Set {
		return errors.New(errors.ErrInvalidRequest,
			"missing resource_selector",
			"tags_multi_assignment", tma)
	}

	return tma.Validate()
}

// CreateTagsMultiAssignment creates resource tags using an resource selector.
func (s *Service) CreateTagsMultiAssignment(ctx context.Context,
	v *TagsMultiAssignment,
) (*TagsMultiAssignment, error) {
	if _, err := request.ContextAuthUser(ctx); err != nil {
		return nil, err
	}

	if v == nil {
		return nil, errors.New(errors.ErrInvalidRequest,
			"missing tags_multi_assignment",
			"tags_multi_assignment", v)
	}

	if err := v.ValidateCreate(); err != nil {
		return nil, err
	}

	resources, _, err := s.GetResources(ctx, &search.Query{
		Search: v.ResourceSelector.Value,
		Size:   10000,
	}, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to get resource selector resources",
			"tags_multi_assignment", v)
	}

	for _, a := range resources {
		if _, err := s.AddResourceTags(ctx, a.ResourceID.Value,
			v.Tags.Value); err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to create resource selector tags",
				"tags_multi_assignment", v)
		}
	}

	return v, nil
}

// DeleteTagsMultiAssignment deletes resource tags an resource selector.
func (s *Service) DeleteTagsMultiAssignment(ctx context.Context,
	v *TagsMultiAssignment,
) (*TagsMultiAssignment, error) {
	if _, err := request.ContextAuthUser(ctx); err != nil {
		return nil, err
	}

	if v == nil {
		return nil, errors.New(errors.ErrInvalidRequest,
			"missing tags_multi_assignment",
			"tags_multi_assignment", v)
	}

	if err := v.ValidateCreate(); err != nil {
		return nil, err
	}

	resources, _, err := s.GetResources(ctx, &search.Query{
		Search: v.ResourceSelector.Value,
		Size:   10000,
	}, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabase,
			"unable to get resource selector resources",
			"tags_multi_assignment", v)
	}

	for _, a := range resources {
		if err := s.DeleteResourceTags(ctx, a.ResourceID.Value,
			v.Tags.Value); err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabase,
				"unable to delete resource selector tags",
				"tags_multi_assignment", v)
		}
	}

	return v, nil
}
