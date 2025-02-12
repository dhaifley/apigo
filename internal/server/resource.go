package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/dhaifley/apigo/internal/errors"
	"github.com/dhaifley/apigo/internal/logger"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/resource"
	"github.com/dhaifley/apigo/internal/search"
	"github.com/dhaifley/apigo/internal/sqldb"
	"github.com/go-chi/chi/v5"
)

// ResourceService values are used to perform resource management.
type ResourceService interface {
	GetResources(ctx context.Context,
		query *search.Query,
		options sqldb.FieldOptions,
	) ([]*resource.Resource, []*sqldb.SummaryData, error)
	GetResource(ctx context.Context,
		id string,
		options sqldb.FieldOptions,
	) (*resource.Resource, error)
	CreateResource(ctx context.Context,
		v *resource.Resource,
	) (*resource.Resource, error)
	UpdateResource(ctx context.Context,
		v *resource.Resource,
	) (*resource.Resource, error)
	DeleteResource(ctx context.Context,
		id string,
	) error
	UpdateResourceData(ctx context.Context,
		payload map[string]any,
		accountID, resourceID string,
	) (*resource.Resource, error)
	UpdateResourceError(ctx context.Context,
		accountID, resourceID string,
		resourceError error,
	) error
	ImportResources(ctx context.Context,
		force bool,
		authSvc resource.AuthService,
	) error
	ImportResource(ctx context.Context,
		authSvc resource.AuthService,
		resourceID string,
	) error
	Update(ctx context.Context,
		authSvc resource.AuthService,
	) context.CancelFunc
	GetTags(ctx context.Context) (resource.TagMap, error)
	GetResourceTags(ctx context.Context,
		resourceID string,
	) ([]string, error)
	AddResourceTags(ctx context.Context,
		resourceID string,
		tags []string,
	) ([]string, error)
	DeleteResourceTags(ctx context.Context,
		resourceID string,
		tags []string,
	) error
	CreateTagsMultiAssignment(ctx context.Context,
		v *resource.TagsMultiAssignment,
	) (*resource.TagsMultiAssignment, error)
	DeleteTagsMultiAssignment(ctx context.Context,
		v *resource.TagsMultiAssignment,
	) (*resource.TagsMultiAssignment, error)
}

// SetResourceService sets the get resource service function.
func (s *Server) SetResourceService(svc ResourceService) {
	s.Lock()
	defer s.Unlock()

	s.getResourceService = func(r *http.Request) ResourceService {
		return svc
	}
}

// ResourceHandler performs routing for event type requests.
func (s *Server) ResourceHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(s.dbAvail)

	r.With(s.Stat, s.Trace, s.Auth).Post("/{id}/import", s.PostImportResource)
	r.With(s.Stat, s.Trace, s.Auth).Post("/import", s.PostImportResources)

	r.With(s.Stat, s.Trace).Post(
		"/update/{account_id}/{id}",
		s.PostUpdateResource)

	r.With(s.Stat, s.Trace, s.Auth).Get("/tags", s.GetAllResourceTags)

	r.With(s.Stat, s.Trace, s.Auth).Post("/tags_multi_assignments",
		s.PostTagsMultiAssignment)
	r.With(s.Stat, s.Trace, s.Auth).Post("/tags_multi_assignment",
		s.PostTagsMultiAssignment)
	r.With(s.Stat, s.Trace, s.Auth).Delete("/tags_multi_assignments",
		s.DeleteTagsMultiAssignment)
	r.With(s.Stat, s.Trace, s.Auth).Delete("/tags_multi_assignment",
		s.DeleteTagsMultiAssignment)

	r.With(s.Stat, s.Trace, s.Auth).Get("/{id}/tags",
		s.GetResourceTags)
	r.With(s.Stat, s.Trace, s.Auth).Post("/{id}/tags",
		s.PostResourceTags)
	r.With(s.Stat, s.Trace, s.Auth).Delete("/{id}/tags",
		s.DeleteResourceTags)

	r.With(s.Stat, s.Trace, s.Auth).Get("/", s.SearchResource)
	r.With(s.Stat, s.Trace, s.Auth).Get("/{id}", s.GetResource)
	r.With(s.Stat, s.Trace, s.Auth).Post("/", s.PostResource)
	r.With(s.Stat, s.Trace, s.Auth).Patch("/{id}", s.PutResource)
	r.With(s.Stat, s.Trace, s.Auth).Put("/{id}", s.PutResource)
	r.With(s.Stat, s.Trace, s.Auth).Delete("/{id}", s.DeleteResource)

	return r
}

// SearchResource is the search handler function for resource types.
func (s *Server) SearchResource(w http.ResponseWriter, r *http.Request) {
	svc := s.getResourceService(r)

	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeResourcesRead); err != nil {
		s.error(err, w, r)

		return
	}

	q, err := search.ParseQuery(r.URL.Query())
	if err != nil {
		s.error(err, w, r)

		return
	}

	opts, err := sqldb.ParseFieldOptions(r.URL.Query())
	if err != nil {
		s.error(err, w, r)

		return
	}

	res, sum, err := svc.GetResources(ctx, q, opts)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if q.Summary != "" {
		if err := json.NewEncoder(w).Encode(sum); err != nil {
			s.error(err, w, r)
		}

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// GetResource is the get handler function for resource types.
func (s *Server) GetResource(w http.ResponseWriter, r *http.Request) {
	svc := s.getResourceService(r)

	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeResourcesRead); err != nil {
		s.error(err, w, r)

		return
	}

	id := chi.URLParam(r, "id")

	opts, err := sqldb.ParseFieldOptions(r.URL.Query())
	if err != nil {
		s.error(err, w, r)

		return
	}

	res, err := svc.GetResource(ctx, id, opts)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// PostResource is the post handler function for resource types.
func (s *Server) PostResource(w http.ResponseWriter, r *http.Request) {
	svc := s.getResourceService(r)

	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeResourcesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	req := &resource.Resource{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.error(e, w, r)
		default:
			s.error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	res, err := svc.CreateResource(ctx, req)
	if err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusCreated)

	scheme := "https"
	if strings.Contains(r.Host, "localhost") {
		scheme = "http"
	}

	loc := &url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   r.URL.Path + "/" + res.ResourceID.Value,
	}

	w.Header().Set("Location", loc.String())

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// PutResource is the put handler function for resource types.
func (s *Server) PutResource(w http.ResponseWriter, r *http.Request) {
	svc := s.getResourceService(r)

	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeResourcesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	id := chi.URLParam(r, "id")

	req := &resource.Resource{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.error(e, w, r)
		default:
			s.error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	req.ResourceID = request.FieldString{
		Set: true, Valid: true,
		Value: id,
	}

	res, err := svc.UpdateResource(ctx, req)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// DeleteResource is the delete handler function for resource types.
func (s *Server) DeleteResource(w http.ResponseWriter, r *http.Request) {
	svc := s.getResourceService(r)

	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeResourcesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	id := chi.URLParam(r, "id")

	if err := svc.DeleteResource(ctx, id); err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PostUpdateResource is the post handler function for external systems
// to update the resource data.
func (s *Server) PostUpdateResource(w http.ResponseWriter, r *http.Request) {
	svc := s.getResourceService(r)

	ctx := r.Context()

	accountID := chi.URLParam(r, "account_id")

	resourceID := chi.URLParam(r, "id")

	req := map[string]any{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var dErr *errors.Error

		switch e := err.(type) {
		case *errors.Error:
			dErr = e
		default:
			dErr = errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request")
		}

		aErr := svc.UpdateResourceError(ctx, accountID, resourceID, dErr)
		if aErr != nil && !errors.Has(aErr, errors.ErrNotFound) {
			s.log.Log(ctx, logger.LvlError,
				"unable to set resource error status",
				"error", aErr,
				"account_id", accountID,
				"resource_id", resourceID,
				"resource_error", dErr)
		}

		s.error(dErr, w, r)

		return
	}

	res, err := svc.UpdateResourceData(ctx, req, accountID, resourceID)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// PostImportResources is the post handler used to import resources.
func (s *Server) PostImportResources(w http.ResponseWriter, r *http.Request) {
	svc := s.getResourceService(r)

	aSvc := s.getAuthService(r)

	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeResourcesAdmin); err != nil {
		s.error(err, w, r)

		return
	}

	force := false

	fs := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("force")))
	if fs != "" && fs != "0" && fs != "f" && fs != "false" {
		force = true
	}

	if err := svc.ImportResources(ctx, force, aSvc); err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PostImportResource is the post handler used to import a single resource.
func (s *Server) PostImportResource(w http.ResponseWriter, r *http.Request) {
	svc := s.getResourceService(r)

	aSvc := s.getAuthService(r)

	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeResourcesAdmin); err != nil {
		s.error(err, w, r)

		return
	}

	id := chi.URLParam(r, "id")

	if err := svc.ImportResource(ctx, aSvc, id); err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAllResourceTags is the get handler function for all resource tags.
func (s *Server) GetAllResourceTags(w http.ResponseWriter, r *http.Request) {
	svc := s.getResourceService(r)

	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeResourcesRead); err != nil {
		s.error(err, w, r)

		return
	}

	res, err := svc.GetTags(ctx)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// GetResourceTags is the get handler function for resource tags.
func (s *Server) GetResourceTags(w http.ResponseWriter, r *http.Request) {
	svc := s.getResourceService(r)

	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeResourcesRead); err != nil {
		s.error(err, w, r)

		return
	}

	resourceID := chi.URLParam(r, "id")

	res, err := svc.GetResourceTags(ctx, resourceID)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// PostResourceTags is the post handler function for resource tags.
func (s *Server) PostResourceTags(w http.ResponseWriter, r *http.Request) {
	svc := s.getResourceService(r)

	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeResourcesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	resourceID := chi.URLParam(r, "id")

	tags := []string{}

	if err := json.NewDecoder(r.Body).Decode(&tags); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.error(e, w, r)
		default:
			s.error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	res, err := svc.AddResourceTags(ctx, resourceID, tags)
	if err != nil {
		s.error(err, w, r)

		return
	}

	w.Header().Set("Location", r.URL.String())

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// DeleteResourceTags is the delete handler function for resource tags.
func (s *Server) DeleteResourceTags(w http.ResponseWriter, r *http.Request) {
	svc := s.getResourceService(r)

	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeResourcesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	resourceID := chi.URLParam(r, "id")

	tags := []string{}

	if err := json.NewDecoder(r.Body).Decode(&tags); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.error(e, w, r)
		default:
			s.error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	if err := svc.DeleteResourceTags(ctx, resourceID, tags); err != nil {
		s.error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PostTagsMultiAssignment is the post handler function for resource tags
// multiple assignments.
func (s *Server) PostTagsMultiAssignment(w http.ResponseWriter,
	r *http.Request,
) {
	svc := s.getResourceService(r)

	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeResourcesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	req := &resource.TagsMultiAssignment{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.error(e, w, r)
		default:
			s.error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	res, err := svc.CreateTagsMultiAssignment(ctx, req)
	if err != nil {
		s.error(err, w, r)

		return
	}

	scheme := "https"
	if strings.Contains(r.Host, "localhost") {
		scheme = "http"
	}

	loc := &url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   r.URL.Path,
	}

	w.Header().Set("Location", loc.String())

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}

// DeleteTagsMultiAssignment is the delete handler function for resource tags
// multiple removals.
func (s *Server) DeleteTagsMultiAssignment(w http.ResponseWriter,
	r *http.Request,
) {
	svc := s.getResourceService(r)

	ctx := r.Context()

	if err := s.checkScope(ctx, request.ScopeResourcesWrite); err != nil {
		s.error(err, w, r)

		return
	}

	req := &resource.TagsMultiAssignment{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.error(e, w, r)
		default:
			s.error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	res, err := svc.DeleteTagsMultiAssignment(ctx, req)
	if err != nil {
		s.error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.error(err, w, r)
	}
}
