package server_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dhaifley/apid/internal/request"
	"github.com/dhaifley/apid/internal/resource"
	"github.com/dhaifley/apid/internal/search"
	"github.com/dhaifley/apid/internal/server"
	"github.com/dhaifley/apid/internal/sqldb"
	"github.com/dhaifley/apid/tests/mocks"
)

type mockResourceService struct{}

func (m *mockResourceService) GetResources(ctx context.Context,
	query *search.Query,
	options sqldb.FieldOptions,
) ([]*resource.Resource, []*sqldb.SummaryData, error) {
	return []*resource.Resource{&mocks.TestResource}, []*sqldb.SummaryData{{
		"status": mocks.TestResource.Status.Value,
		"count":  1,
	}}, nil
}

func (m *mockResourceService) GetResource(ctx context.Context,
	id string,
	options sqldb.FieldOptions,
) (*resource.Resource, error) {
	return &mocks.TestResource, nil
}

func (m *mockResourceService) CreateResource(ctx context.Context,
	v *resource.Resource,
) (*resource.Resource, error) {
	return &mocks.TestResource, nil
}

func (m *mockResourceService) UpdateResource(ctx context.Context,
	v *resource.Resource,
) (*resource.Resource, error) {
	return &mocks.TestResource, nil
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
	return &mocks.TestResource, nil
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

	svr.SetDB(&mocks.MockResourceDB{})

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
			mocks.TestResource.ResourceID.Value + `"`,
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

	svr.SetDB(&mocks.MockResourceDB{})

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
		url:    basePath + "/resources/" + mocks.TestResource.ResourceID.Value,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp: `"resource_id":"` +
			mocks.TestResource.ResourceID.Value + `"`,
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

	svr.SetDB(&mocks.MockResourceDB{})

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
			"event_id": "` + mocks.TestResource.ResourceID.Value + `",
			"name":"test",
			"status":"` + request.StatusActive + `"
		}`,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusCreated,
		resp: `"resource_id":"` +
			mocks.TestResource.ResourceID.Value + `"`,
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

	svr.SetDB(&mocks.MockResourceDB{})

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
		url:  basePath + "/resources/" + mocks.TestResource.ResourceID.Value,
		body: `{
			"name": "changed",
			"status":"` + request.StatusActive + `"
		}`,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp: `"resource_id":"` +
			mocks.TestResource.ResourceID.Value + `"`,
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

	svr.SetDB(&mocks.MockResourceDB{})

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
		url:    basePath + "/resources/" + mocks.TestResource.ResourceID.Value,
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

	svr.SetDB(&mocks.MockResourceDB{})

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
		url: basePath + "/resources/update/" + mocks.TestID + "/" +
			mocks.TestUUID,
		body: `{
			"resources": [
				{
					"resource_id": "` + mocks.TestUUID + `",
					"account_id": "` + mocks.TestID + `",
					"cleared_on": 1
				}
			]
		}`,
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusOK,
		resp:   `"resource_id":"` + mocks.TestUUID + `"`,
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

	svr.SetDB(&mocks.MockResourceDB{})

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
		header: map[string]string{"Authorization": "test"},
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

	svr.SetDB(&mocks.MockResourceDB{})

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
		header: map[string]string{"Authorization": "test"},
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
