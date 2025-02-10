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
	"github.com/dhaifley/apigo/internal/search"
	"github.com/dhaifley/apigo/internal/server"
	"github.com/dhaifley/apigo/internal/sqldb"
	"github.com/dhaifley/apigo/tests/mocks"
)

type mockAuthService struct{}

func (m *mockAuthService) AuthJWT(ctx context.Context,
	token, tenant string,
) (*auth.Claims, error) {
	switch token {
	case "test":
		return &auth.Claims{
			AccountID:   mocks.TestAccount.AccountID.Value,
			AccountName: mocks.TestAccount.Name.Value,
			UserID:      mocks.TestUser.UserID.Value,
			Roles:       []string{request.RoleUser},
		}, nil
	case "admin":
		return &auth.Claims{
			AccountID:   mocks.TestAccount.AccountID.Value,
			AccountName: mocks.TestAccount.Name.Value,
			UserID:      mocks.TestUser.UserID.Value,
			Roles:       []string{request.RoleAdmin},
		}, nil
	default:
		return nil, errors.New(errors.ErrForbidden, "invalid auth token")
	}
}

func (m *mockAuthService) GetAccount(ctx context.Context, id string,
) (*auth.Account, error) {
	return &mocks.TestAccount, nil
}

func (m *mockAuthService) CreateAccount(ctx context.Context,
	v *auth.Account,
) (*auth.Account, error) {
	return &mocks.TestAccount, nil
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
	return &mocks.TestUser, nil
}

func (m *mockAuthService) UpdateUser(ctx context.Context, v *auth.User,
) (*auth.User, error) {
	return &mocks.TestUser, nil
}

func (m *mockAuthService) Update(ctx context.Context) context.CancelFunc {
	_, cancel := context.WithCancel(ctx)

	return cancel
}

func (m *mockAuthService) GetTokens(ctx context.Context,
	query *search.Query,
	options sqldb.FieldOptions,
) ([]*auth.Token, []*sqldb.SummaryData, error) {
	return []*auth.Token{&mocks.TestToken},
		[]*sqldb.SummaryData{{
			"status": mocks.TestToken.Status.Value,
			"count":  1,
		}}, nil
}

func (m *mockAuthService) GetToken(ctx context.Context,
	token string,
	options sqldb.FieldOptions,
) (*auth.Token, error) {
	return &mocks.TestToken, nil
}

func (m *mockAuthService) CreateToken(ctx context.Context,
	token *auth.Token,
) (*auth.Token, error) {
	return &mocks.TestToken, nil
}

func (m *mockAuthService) DeleteToken(ctx context.Context,
	token string,
) error {
	return nil
}

func TestAuth(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svr.SetDB(&mocks.MockAuthDB{})

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
		resp:   `"account_id":"` + mocks.TestID + `"`,
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

	svr.SetDB(&mocks.MockAuthDB{})

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
		resp:   `"account_id":"` + mocks.TestID + `"`,
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

	svr.SetDB(&mocks.MockAuthDB{})

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
		body:   `{"account_id":"` + mocks.TestID + `"}`,
		code:   http.StatusCreated,
		resp:   `"account_id":"` + mocks.TestID + `"`,
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

	svr.SetDB(&mocks.MockAuthDB{})

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

	svr.SetDB(&mocks.MockAuthDB{})

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

	svr.SetDB(&mocks.MockAuthDB{})

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
		resp:   `"user_id":"` + mocks.TestUUID + `"`,
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

	svr.SetDB(&mocks.MockAuthDB{})

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
		body:   `{"user_id":"` + mocks.TestUUID + `"}`,
		code:   http.StatusOK,
		resp:   `"user_id":"` + mocks.TestUUID + `"`,
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

func TestSearchToken(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svr.SetDB(&mocks.MockAuthDB{})

	svr.SetAuthService(&mockAuthService{})

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		query  string
		header map[string]string
		code   int
		resp   string
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/tokens",
		query:  `?search=and(token:*)&from=0&size=10&order=token&dir=asc`,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp:   `"token_id":"` + mocks.TestToken.TokenID.Value + `"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u := tt.url + tt.query

			r, err := http.NewRequest(http.MethodGet, u, nil)
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

func TestGetToken(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svr.SetDB(&mocks.MockAuthDB{})

	svr.SetAuthService(&mockAuthService{})

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		code   int
		resp   string
	}{{
		name: "success",
		w:    httptest.NewRecorder(),
		url: basePath + "/tokens/" +
			mocks.TestToken.TokenID.Value,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp:   `"token_id":"` + mocks.TestToken.TokenID.Value + `"`,
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

func TestPostToken(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svr.SetDB(&mocks.MockAuthDB{})

	svr.SetAuthService(&mockAuthService{})

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		body   string
		header map[string]string
		code   int
		resp   string
	}{{
		name: "success",
		w:    httptest.NewRecorder(),
		url:  basePath + "/tokens",
		body: `{
			"token": "` + mocks.TestToken.TokenID.Value + `",
			"status":"` + request.StatusActive + `",
			"expiration": 1000
		}`,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusCreated,
		resp:   `"token_id":"` + mocks.TestToken.TokenID.Value + `"`,
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

func TestDeleteToken(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	svr.SetDB(&mocks.MockAuthDB{})

	svr.SetAuthService(&mockAuthService{})

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		code   int
	}{{
		name: "success",
		w:    httptest.NewRecorder(),
		url: basePath + "/tokens/" +
			mocks.TestToken.TokenID.Value,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusNoContent,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequest(http.MethodDelete, tt.url, nil)
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
		})
	}
}
