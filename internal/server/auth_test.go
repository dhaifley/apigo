package server_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dhaifley/apigo/internal/auth"
	"github.com/dhaifley/apigo/internal/errors"
	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/server"
	"github.com/dhaifley/apigo/internal/sqldb"
)

var TestAccount = auth.Account{
	AccountID: request.FieldString{
		Set: true, Valid: true,
		Value: TestID,
	},
	Name: request.FieldString{
		Set: true, Valid: true,
		Value: "testAccount",
	},
	Status: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusActive,
	},
	StatusData: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"last_error": "test",
		},
	},
	Repo: request.FieldString{
		Set: true, Valid: true,
		Value: "test",
	},
	RepoStatus: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusActive,
	},
	RepoStatusData: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"last_error": "test",
		},
	},
	Secret: request.FieldString{
		Set: true, Valid: true,
		Value: "test",
	},
	Data: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"test": "test",
		},
	},
}

var TestUser = auth.User{
	UserID: request.FieldString{
		Set: true, Valid: true,
		Value: TestUUID,
	},
	Email: request.FieldString{
		Set: true, Valid: true,
		Value: "test@apigo.io",
	},
	LastName: request.FieldString{
		Set: true, Valid: true,
		Value: "testLastName",
	},
	FirstName: request.FieldString{
		Set: true, Valid: true,
		Value: "testFirstName",
	},
	Status: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusActive,
	},
	Scopes: request.FieldString{
		Set: true, Valid: true,
		Value: request.ScopeSuperUser,
	},
	Data: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"test": "test",
		},
	},
}

type mockAuthService struct{}

func (m *mockAuthService) AuthJWT(ctx context.Context,
	token, tenant string,
) (*auth.Claims, error) {
	switch token {
	case "test":
		return &auth.Claims{
			AccountID:   TestAccount.AccountID.Value,
			AccountName: TestAccount.Name.Value,
			UserID:      TestUser.UserID.Value,
			Roles:       []string{request.RoleUser},
		}, nil
	case "admin":
		return &auth.Claims{
			AccountID:   TestAccount.AccountID.Value,
			AccountName: TestAccount.Name.Value,
			UserID:      TestUser.UserID.Value,
			Roles:       []string{request.RoleAdmin},
		}, nil
	default:
		return nil, errors.New(errors.ErrForbidden, "invalid auth token")
	}
}

func (m *mockAuthService) GetAccount(ctx context.Context, id string,
) (*auth.Account, error) {
	return &TestAccount, nil
}

func (m *mockAuthService) CreateAccount(ctx context.Context,
	v *auth.Account,
) (*auth.Account, error) {
	return &TestAccount, nil
}

func (m *mockAuthService) GetAccountRepo(ctx context.Context,
) (*auth.AccountRepo, error) {
	return &auth.AccountRepo{
		Repo: request.FieldString{
			Set: true, Valid: true,
			Value: "test://test:test@test/test/test#test",
		},
		RepoStatus: request.FieldString{
			Set: true, Valid: true,
			Value: request.StatusActive,
		},
	}, nil
}

func (m *mockAuthService) SetAccountRepo(ctx context.Context,
	v *auth.AccountRepo,
) error {
	return nil
}

func (m *mockAuthService) GetUser(ctx context.Context,
	id string,
	options sqldb.FieldOptions,
) (*auth.User, error) {
	return &TestUser, nil
}

func (m *mockAuthService) UpdateUser(ctx context.Context, v *auth.User,
) (*auth.User, error) {
	return &TestUser, nil
}

func (m *mockAuthService) Update(ctx context.Context) context.CancelFunc {
	_, cancel := context.WithCancel(ctx)

	return cancel
}

func TestAuth(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	md, _, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svr.SetDB(md)

	svr.SetAuthService(&mockAuthService{})

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		code   int
		resp   string
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/account",
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp:   `"account_id":"` + TestID + `"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequest(http.MethodGet, tt.url, nil)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, tt.w.Code)
			}

			res := tt.w.Body.String()
			if !strings.Contains(res, tt.resp) {
				t.Errorf("Expected body to contain: %v, got: %v", tt.resp, res)
			}
		})
	}
}

func TestGetAccount(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	md, _, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svr.SetDB(md)

	svr.SetAuthService(&mockAuthService{})

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		code   int
		resp   string
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/account",
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp:   `"account_id":"` + TestID + `"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequest(http.MethodGet, tt.url, nil)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, tt.w.Code)
			}

			res := tt.w.Body.String()
			if !strings.Contains(res, tt.resp) {
				t.Errorf("Expected body to contain: %v, got: %v", tt.resp, res)
			}
		})
	}
}

func TestPostAccount(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	md, _, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svr.SetDB(md)

	svr.SetAuthService(&mockAuthService{})

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		body   string
		code   int
		resp   string
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/account",
		header: map[string]string{"Authorization": "admin"},
		body:   `{"account_id":"` + TestID + `"}`,
		code:   http.StatusCreated,
		resp:   `"account_id":"` + TestID + `"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(tt.body)

			r, err := http.NewRequest(http.MethodPost, tt.url, buf)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, tt.w.Code)
			}

			res := tt.w.Body.String()
			if !strings.Contains(res, tt.resp) {
				t.Errorf("Expected body to contain: %v, got: %v", tt.resp, res)
			}
		})
	}
}

func TestGetAccountRepo(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	md, _, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svr.SetDB(md)

	svr.SetAuthService(&mockAuthService{})

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		code   int
		resp   string
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/account/repo",
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp:   `"repo_status":"active"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequest(http.MethodGet, tt.url, nil)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, tt.w.Code)
			}

			res := tt.w.Body.String()
			if !strings.Contains(res, tt.resp) {
				t.Errorf("Expected body to contain: %v, got: %v", tt.resp, res)
			}
		})
	}
}

func TestPostAccountRepo(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	md, _, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svr.SetDB(md)

	svr.SetAuthService(&mockAuthService{})

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		body   string
		code   int
		resp   string
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/account/repo",
		header: map[string]string{"Authorization": "admin"},
		body:   `{"repo":"test://test:test@test/test/test#test"}`,
		code:   http.StatusCreated,
		resp:   "test://",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(tt.body)

			r, err := http.NewRequest(http.MethodPost, tt.url, buf)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, tt.w.Code)
			}

			res := tt.w.Body.String()
			if !strings.Contains(res, tt.resp) {
				t.Errorf("Expected body to contain: %v, got: %v", tt.resp, res)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	md, _, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svr.SetDB(md)

	svr.SetAuthService(&mockAuthService{})

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		code   int
		resp   string
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/user",
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp:   `"user_id":"` + TestUUID + `"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequest(http.MethodGet, tt.url, nil)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, tt.w.Code)
			}

			res := tt.w.Body.String()
			if !strings.Contains(res, tt.resp) {
				t.Errorf("Expected body to contain: %v, got: %v", tt.resp, res)
			}
		})
	}
}

func TestPutUser(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	md, _, err := sqldb.NewMockSQLDB(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svr.SetDB(md)

	svr.SetAuthService(&mockAuthService{})

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		body   string
		code   int
		resp   string
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/user",
		header: map[string]string{"Authorization": "test"},
		body:   `{"user_id":"` + TestUUID + `"}`,
		code:   http.StatusOK,
		resp:   `"user_id":"` + TestUUID + `"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(tt.body)

			r, err := http.NewRequest(http.MethodPut, tt.url, buf)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			svr.Mux(tt.w, r)

			if tt.w.Code != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, tt.w.Code)
			}

			res := tt.w.Body.String()
			if !strings.Contains(res, tt.resp) {
				t.Errorf("Expected body to contain: %v, got: %v", tt.resp, res)
			}
		})
	}
}
