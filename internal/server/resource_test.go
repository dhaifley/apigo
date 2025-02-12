package server_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/internal/resource"
	"github.com/dhaifley/apigo/internal/search"
	"github.com/dhaifley/apigo/internal/server"
	"github.com/dhaifley/apigo/internal/sqldb"
)

var TestResource = resource.Resource{
	ResourceID: request.FieldString{
		Set: true, Valid: true,
		Value: TestUUID,
	},
	Name: request.FieldString{
		Set: true, Valid: true,
		Value: "testName",
	},
	Version: request.FieldString{
		Set: true, Valid: true,
		Value: "1",
	},
	Description: request.FieldString{
		Set: true, Valid: true,
		Value: "testDescription",
	},
	Status: request.FieldString{
		Set: true, Valid: true,
		Value: request.StatusNew,
	},
	StatusData: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			"last_error": "testError",
		},
	},
	KeyField: request.FieldString{
		Set: true, Valid: true,
		Value: "resource_id",
	},
	KeyRegex: request.FieldString{
		Set: true, Valid: true,
		Value: ".*",
	},
	ClearCondition: request.FieldString{
		Set: true, Valid: true,
		Value: "gt(cleared_on:0)",
	},
	ClearAfter: request.FieldInt64{
		Set: true, Valid: true,
		Value: int64(time.Hour.Seconds()),
	},
	ClearDelay: request.FieldInt64{
		Set: true, Valid: true,
		Value: 0,
	},
	Data: request.FieldJSON{
		Set: true, Valid: true,
		Value: map[string]any{
			TestUUID: map[string]any{
				"test":        "testData",
				"resource_id": TestUUID,
				"array": []map[string]any{{
					"status": "testStatus",
				}},
			},
		},
	},
	Source: request.FieldString{
		Set: true, Valid: true,
		Value: "testSource",
	},
	CommitHash: request.FieldString{
		Set: true, Valid: true,
		Value: "testHash",
	},
	CreatedBy: request.FieldString{
		Set: true, Valid: true,
		Value: TestID,
	},
	CreatedAt: request.FieldTime{
		Set: true, Valid: true,
		Value: 1,
	},
	UpdatedBy: request.FieldString{
		Set: true, Valid: true,
		Value: TestID,
	},
	UpdatedAt: request.FieldTime{
		Set: true, Valid: true,
		Value: 1,
	},
}

type mockResourceService struct{}

func (m *mockResourceService) GetResources(ctx context.Context,
	query *search.Query,
	options sqldb.FieldOptions,
) ([]*resource.Resource, []*sqldb.SummaryData, error) {
	return []*resource.Resource{&TestResource}, []*sqldb.SummaryData{{
		"status": TestResource.Status.Value,
		"count":  1,
	}}, nil
}

func (m *mockResourceService) GetResource(ctx context.Context,
	id string,
	options sqldb.FieldOptions,
) (*resource.Resource, error) {
	return &TestResource, nil
}

func (m *mockResourceService) CreateResource(ctx context.Context,
	v *resource.Resource,
) (*resource.Resource, error) {
	return &TestResource, nil
}

func (m *mockResourceService) UpdateResource(ctx context.Context,
	v *resource.Resource,
) (*resource.Resource, error) {
	return &TestResource, nil
}

func (m *mockResourceService) DeleteResource(ctx context.Context,
	id string,
) error {
	return nil
}

func (m *mockResourceService) UpdateResourceData(ctx context.Context,
	payload map[string]any,
	accountID, resourceID string,
) (*resource.Resource, error) {
	return &TestResource, nil
}

func (m *mockResourceService) UpdateResourceError(ctx context.Context,
	accountID, resourceID string,
	resourceError error,
) error {
	return nil
}

func (m *mockResourceService) ImportResources(ctx context.Context,
	force bool,
	authSvc resource.AuthService,
) error {
	return nil
}

func (m *mockResourceService) ImportResource(ctx context.Context,
	authSvc resource.AuthService,
	resourceID string,
) error {
	return nil
}

func (m *mockResourceService) Update(ctx context.Context,
	authSvc resource.AuthService,
) context.CancelFunc {
	_, cancel := context.WithCancel(ctx)

	return cancel
}

func (m *mockResourceService) GetTags(ctx context.Context,
) (resource.TagMap, error) {
	return resource.TagMap{"test": []string{"test"}}, nil
}

func (m *mockResourceService) GetResourceTags(ctx context.Context,
	resourceID string,
) ([]string, error) {
	return []string{"test:test"}, nil
}

func (m *mockResourceService) AddResourceTags(ctx context.Context,
	resourceID string,
	tags []string,
) ([]string, error) {
	return tags, nil
}

func (m *mockResourceService) DeleteResourceTags(ctx context.Context,
	resourceID string,
	tags []string,
) error {
	return nil
}

func (m *mockResourceService) CreateTagsMultiAssignment(
	ctx context.Context,
	v *resource.TagsMultiAssignment,
) (*resource.TagsMultiAssignment, error) {
	return v, nil
}

func (m *mockResourceService) DeleteTagsMultiAssignment(
	ctx context.Context,
	v *resource.TagsMultiAssignment,
) (*resource.TagsMultiAssignment, error) {
	return v, nil
}

func TestSearchResource(t *testing.T) {
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

	svr.SetResourceService(&mockResourceService{})

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
		url:    basePath + "/resources",
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp: `"resource_id":"` +
			TestResource.ResourceID.Value + `"`,
	}, {
		name:   "summary",
		w:      httptest.NewRecorder(),
		url:    basePath + "/resources",
		query:  `?search=and(name:*)&summary=status`,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp:   `"count":1`,
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

func TestGetResource(t *testing.T) {
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

	svr.SetResourceService(&mockResourceService{})

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
		url:    basePath + "/resources/" + TestResource.ResourceID.Value,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp: `"resource_id":"` +
			TestResource.ResourceID.Value + `"`,
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

func TestPostResource(t *testing.T) {
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

	svr.SetResourceService(&mockResourceService{})

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
		url:  basePath + "/resources",
		body: `{
			"event_id": "` + TestResource.ResourceID.Value + `",
			"name":"test",
			"status":"` + request.StatusActive + `"
		}`,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusCreated,
		resp: `"resource_id":"` +
			TestResource.ResourceID.Value + `"`,
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

func TestPutResource(t *testing.T) {
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

	svr.SetResourceService(&mockResourceService{})

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
		url:  basePath + "/resources/" + TestResource.ResourceID.Value,
		body: `{
			"name": "changed",
			"status":"` + request.StatusActive + `"
		}`,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp: `"resource_id":"` +
			TestResource.ResourceID.Value + `"`,
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

func TestDeleteResource(t *testing.T) {
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

	svr.SetResourceService(&mockResourceService{})

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		code   int
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/resources/" + TestResource.ResourceID.Value,
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

func TestPostUpdateResources(t *testing.T) {
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

	svr.SetResourceService(&mockResourceService{})

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
		url: basePath + "/resources/update/" + TestID + "/" +
			TestUUID,
		body: `{
			"resources": [
				{
					"resource_id": "` + TestUUID + `",
					"account_id": "` + TestID + `",
					"cleared_on": 1
				}
			]
		}`,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp:   `"resource_id":"` + TestUUID + `"`,
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

func TestPostImportResources(t *testing.T) {
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

	svr.SetResourceService(&mockResourceService{})

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		code   int
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/resources/import",
		header: map[string]string{"Authorization": "admin"},
		code:   http.StatusNoContent,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequest(http.MethodPost, tt.url, nil)
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

func TestPostImportResource(t *testing.T) {
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

	svr.SetResourceService(&mockResourceService{})

	tests := []struct {
		name   string
		w      *httptest.ResponseRecorder
		url    string
		header map[string]string
		code   int
	}{{
		name:   "success",
		w:      httptest.NewRecorder(),
		url:    basePath + "/resources/1/import",
		header: map[string]string{"Authorization": "admin"},
		code:   http.StatusNoContent,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequest(http.MethodPost, tt.url, nil)
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
