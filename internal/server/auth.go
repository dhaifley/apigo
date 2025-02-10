package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/dhaifley/apigo/internal/auth"
	"github.com/dhaifley/apigo/internal/errors"
	"github.com/dhaifley/apigo/internal/logger"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/search"
	"github.com/dhaifley/apigo/internal/sqldb"
	"github.com/go-chi/chi/v5"
)

// AuthService values are used to perform authentication functions.
type AuthService interface {
	AuthJWT(ctx context.Context,
		token, tenant string,
	) (*auth.Claims, error)
	GetAccount(ctx context.Context,
		id string,
	) (*auth.Account, error)
	CreateAccount(ctx context.Context,
		v *auth.Account,
	) (*auth.Account, error)
	GetAccountRepo(ctx context.Context) (*auth.AccountRepo, error)
	SetAccountRepo(ctx context.Context,
		v *auth.AccountRepo,
	) error
	GetUser(ctx context.Context,
		id string,
		options sqldb.FieldOptions,
	) (*auth.User, error)
	UpdateUser(ctx context.Context,
		v *auth.User,
	) (*auth.User, error)
	Update(ctx context.Context,
	) context.CancelFunc
	GetTokens(ctx context.Context,
		query *search.Query,
		options sqldb.FieldOptions,
	) ([]*auth.Token, []*sqldb.SummaryData, error)
	GetToken(ctx context.Context,
		token string,
		options sqldb.FieldOptions,
	) (*auth.Token, error)
	CreateToken(ctx context.Context,
		v *auth.Token,
	) (*auth.Token, error)
	DeleteToken(ctx context.Context,
		token string,
	) error
}

// Auth wraps an http handler with authentication verification.
func (s *Server) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		svc := s.getAuthService(r)

		ctx := r.Context()

		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

		if token == "" {
			cookie, err := r.Cookie("x-api-key")
			if err != nil && !errors.Is(err, http.ErrNoCookie) {
				s.log.Log(ctx, slog.LevelWarn,
					"invalid authentication cookie received",
					"error", err,
					"cookies", r.Cookies(),
					"request", r)
			} else if cookie != nil {
				token = strings.TrimPrefix(cookie.Value, "Bearer ")
			}
		}

		if token == "" {
			if _, pw, ok := r.BasicAuth(); ok {
				token = pw
			}
		}

		tenant := r.Header.Get("securitytenant")

		claims, err := svc.AuthJWT(ctx, token, tenant)
		if err != nil {
			if e, ok := err.(*errors.Error); ok {
				s.Error(e, w, r)

				return
			}

			s.Error(errors.New(errors.ErrForbidden,
				"unauthenticated request"), w, r)

			return
		}

		if tenant != "" {
			s.log.Log(ctx, logger.LvlInfo,
				"cross-tenant request authorized",
				"error", err,
				"token", token,
				"tenant", tenant,
				"claims", claims,
				"request_method", r.Method,
				"request_url", r.URL.String(),
				"request_headers", r.Header,
				"request_remote", r.RemoteAddr)
		}

		ctx = context.WithValue(ctx, request.CtxKeyJWT, token)

		ctx = context.WithValue(ctx, request.CtxKeyAccountID, claims.AccountID)

		ctx = context.WithValue(ctx, request.CtxKeyAccountName,
			claims.AccountName)

		ctx = context.WithValue(ctx, request.CtxKeyRoles, claims.Roles)

		if claims.UserID != "" {
			ctx = context.WithValue(ctx, request.CtxKeyUserID, claims.UserID)
		}

		if claims.TokenID != "" {
			ctx = context.WithValue(ctx, request.CtxKeyTokenID, claims.TokenID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AccountHandler performs routing for account requests.
func (s *Server) AccountHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(s.DBAvail)

	r.With(s.Stat, s.Trace, s.Auth).Get("/repo", s.GetAccountRepo)
	r.With(s.Stat, s.Trace, s.Auth).Post("/repo", s.PostAccountRepo)

	r.With(s.Stat, s.Trace, s.Auth).Get("/", s.GetAccount)
	r.With(s.Stat, s.Trace, s.Auth).Post("/", s.PostAccount)

	return r
}

// GetAccount is the get handler function for accounts.
func (s *Server) GetAccount(w http.ResponseWriter, r *http.Request) {
	svc := s.getAuthService(r)

	ctx := r.Context()

	res, err := svc.GetAccount(ctx, "")
	if err != nil {
		s.Error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.Error(err, w, r)
	}
}

// PostAccount is the post handler function for accounts.
func (s *Server) PostAccount(w http.ResponseWriter, r *http.Request) {
	svc := s.getAuthService(r)

	ctx := r.Context()

	req := &auth.Account{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.Error(e, w, r)
		default:
			s.Error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	res, err := svc.CreateAccount(ctx, req)
	if err != nil {
		s.Error(err, w, r)

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
		s.Error(err, w, r)
	}
}

// GetAccountRepo is the get handler function for account repos.
func (s *Server) GetAccountRepo(w http.ResponseWriter, r *http.Request) {
	svc := s.getAuthService(r)

	ctx := r.Context()

	res, err := svc.GetAccountRepo(ctx)
	if err != nil {
		s.Error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.Error(err, w, r)
	}
}

// PostAccountRepo is the post handler function for account repos.
func (s *Server) PostAccountRepo(w http.ResponseWriter, r *http.Request) {
	svc := s.getAuthService(r)

	ctx := r.Context()

	req := &auth.AccountRepo{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.Error(e, w, r)
		default:
			s.Error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	if err := svc.SetAccountRepo(ctx, req); err != nil {
		s.Error(err, w, r)

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

	if err := json.NewEncoder(w).Encode(req); err != nil {
		s.Error(err, w, r)
	}
}

// UserHandler performs routing for user requests.
func (s *Server) UserHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(s.DBAvail)

	r.With(s.Stat, s.Trace, s.Auth).Get("/", s.GetUser)
	r.With(s.Stat, s.Trace, s.Auth).Patch("/", s.PutUser)
	r.With(s.Stat, s.Trace, s.Auth).Put("/", s.PutUser)

	return r
}

// TokenHandler performs routing for token requests.
func (s *Server) TokenHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(s.DBAvail)

	r.With(s.Stat, s.Trace, s.Auth).Get("/", s.SearchToken)
	r.With(s.Stat, s.Trace, s.Auth).Get("/{token_id}", s.GetToken)
	r.With(s.Stat, s.Trace, s.Auth).Post("/", s.PostToken)
	r.With(s.Stat, s.Trace, s.Auth).Patch("/{token_id}", s.PutToken)
	r.With(s.Stat, s.Trace, s.Auth).Put("/{token_id}", s.PutToken)
	r.With(s.Stat, s.Trace, s.Auth).Delete("/{token_id}", s.DeleteToken)

	return r
}

// GetUser is the get handler function for users.
func (s *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	svc := s.getAuthService(r)

	ctx := r.Context()

	opts, err := sqldb.ParseFieldOptions(r.URL.Query())
	if err != nil {
		s.Error(err, w, r)

		return
	}

	res, err := svc.GetUser(ctx, "", opts)
	if err != nil {
		s.Error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.Error(err, w, r)
	}
}

// PutUser is the put handler function for users.
func (s *Server) PutUser(w http.ResponseWriter, r *http.Request) {
	svc := s.getAuthService(r)

	ctx := r.Context()

	req := &auth.User{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.Error(e, w, r)
		default:
			s.Error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	req.UserID = request.FieldString{
		Set: true, Valid: true, Value: "",
	}

	res, err := svc.UpdateUser(ctx, req)
	if err != nil {
		s.Error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.Error(err, w, r)
	}
}

// SearchToken is the search handler function for tokens.
func (s *Server) SearchToken(w http.ResponseWriter,
	r *http.Request,
) {
	svc := s.getAuthService(r)

	ctx := r.Context()

	q, err := search.ParseQuery(r.URL.Query())
	if err != nil {
		s.Error(err, w, r)

		return
	}

	opts, err := sqldb.ParseFieldOptions(r.URL.Query())
	if err != nil {
		s.Error(err, w, r)

		return
	}

	res, sum, err := svc.GetTokens(ctx, q, opts)
	if err != nil {
		s.Error(err, w, r)

		return
	}

	if q.Summary != "" {
		if err := json.NewEncoder(w).Encode(sum); err != nil {
			s.Error(err, w, r)
		}

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.Error(err, w, r)
	}
}

// GetToken is the get handler function for tokens.
func (s *Server) GetToken(w http.ResponseWriter,
	r *http.Request,
) {
	svc := s.getAuthService(r)

	ctx := r.Context()

	tID := chi.URLParam(r, "token_id")

	opts, err := sqldb.ParseFieldOptions(r.URL.Query())
	if err != nil {
		s.Error(err, w, r)

		return
	}

	res, err := svc.GetToken(ctx, tID, opts)
	if err != nil {
		s.Error(err, w, r)

		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.Error(err, w, r)
	}
}

// PostToken is the post handler function for tokens.
func (s *Server) PostToken(w http.ResponseWriter,
	r *http.Request,
) {
	svc := s.getAuthService(r)

	ctx := r.Context()

	req := &auth.Token{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		switch e := err.(type) {
		case *errors.Error:
			s.Error(e, w, r)
		default:
			s.Error(errors.Wrap(err, errors.ErrInvalidRequest,
				"unable to decode request"), w, r)
		}

		return
	}

	res, err := svc.CreateToken(ctx, req)
	if err != nil {
		s.Error(err, w, r)

		return
	}

	scheme := "https"
	if strings.Contains(r.Host, "localhost") {
		scheme = "http"
	}

	loc := &url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   r.URL.Path + "/" + res.TokenID.Value,
	}

	w.Header().Set("Location", loc.String())

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		s.Error(err, w, r)
	}
}

// PutToken is the put handler function for tokens.
func (s *Server) PutToken(w http.ResponseWriter,
	r *http.Request,
) {
	s.Error(errors.New(errors.ErrInvalidRequest,
		"unable to update  tokens: they are read only"), w, r)

	return
}

// DeleteToken is the delete handler function for tokens.
func (s *Server) DeleteToken(w http.ResponseWriter,
	r *http.Request,
) {
	svc := s.getAuthService(r)

	ctx := r.Context()

	tID := chi.URLParam(r, "token_id")

	if err := svc.DeleteToken(ctx, tID); err != nil {
		s.Error(err, w, r)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
