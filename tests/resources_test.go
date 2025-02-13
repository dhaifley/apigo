package integration_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
)

func TestResources(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}

	data := map[string]any{}

	dataLock := sync.Mutex{}

	tests := []struct {
		name   string
		url    string
		method string
		header map[string]string
		body   any
		resp   func(t *testing.T, res *http.Response)
	}{{
		name:   "unauthorized",
		url:    "http://localhost:8080/api/v1/resources",
		method: http.MethodGet,
		header: map[string]string{"Authorization": "test"},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusUnauthorized

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			expB := `"Unauthorized"`

			if !strings.Contains(string(b), expB) {
				t.Errorf("Expected body to contain: %v, got: %v",
					expB, string(b))
			}
		},
	}, {
		name:   "password login",
		url:    "http://localhost:8080/api/v1/login/token",
		method: http.MethodPost,
		header: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		body: map[string]any{
			"username": "admin",
			"password": "admin",
			"scope":    "superuser",
		},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusOK

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			m := map[string]any{}

			json.Unmarshal(b, &m)
			if err != nil {
				t.Errorf("Unexpected error decoding response: %v", err)
			}

			at, ok := m["access_token"].(string)
			if !ok {
				t.Errorf("Unexpected response: %v", m)
			}

			if len(at) < 8 {
				t.Errorf("Expected access token, got: %v", at)
			}

			dataLock.Lock()
			data["access_token"] = at
			data["resource_id"] = ""
			dataLock.Unlock()
		},
	}, {
		name:   "search resources empty",
		url:    "http://localhost:8080/api/v1/resources?search=name:test",
		method: http.MethodGet,
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusOK

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			var resources []map[string]any
			if err := json.Unmarshal(b, &resources); err != nil {
				t.Errorf("Unexpected error decoding response: %v", err)
			}

			if len(resources) != 0 {
				t.Errorf("Expected empty resources array, got: %v", resources)
			}
		},
	}, {
		name:   "create resource",
		url:    "http://localhost:8080/api/v1/resources",
		method: http.MethodPost,
		body: map[string]any{
			"name":        "Test Resource",
			"version":     "1",
			"description": "A test resource",
			"status":      "active",
			"key_field":   "resource_id",
			"data": map[string]any{
				"test": "test",
			},
		},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusCreated

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			m := map[string]any{}

			if err := json.Unmarshal(b, &m); err != nil {
				t.Errorf("Unexpected error decoding response: %v", err)
			}

			resourceID, ok := m["resource_id"].(string)
			if !ok {
				t.Errorf("Expected resource_id in response: %v", m)
			}

			dataLock.Lock()
			data["resource_id"] = resourceID
			dataLock.Unlock()
		},
	}, {
		name:   "get resource",
		url:    "http://localhost:8080/api/v1/resources/{{resource_id}}",
		method: http.MethodGet,
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusOK

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			m := map[string]any{}

			if err := json.Unmarshal(b, &m); err != nil {
				t.Errorf("Unexpected error decoding response: %v", err)
			}

			if _, ok := m["resource_id"].(string); !ok {
				t.Errorf("Expected resource_id in response: %v", m)
			}
		},
	}, {
		name:   "patch resource",
		url:    "http://localhost:8080/api/v1/resources/{{resource_id}}",
		method: http.MethodPatch,
		body: map[string]any{
			"description": "Updated test resource",
			"data": map[string]any{
				"updated": true,
			},
		},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusOK

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			m := map[string]any{}

			if err := json.Unmarshal(b, &m); err != nil {
				t.Errorf("Unexpected error decoding response: %v", err)
			}

			desc, ok := m["description"].(string)
			if !ok || desc != "Updated test resource" {
				t.Errorf("Expected updated description in response: %v", m)
			}
		},
	}, {
		name:   "put resource",
		url:    "http://localhost:8080/api/v1/resources/{{resource_id}}",
		method: http.MethodPut,
		body: map[string]any{
			"name":        "Test Resource Updated",
			"version":     "2",
			"description": "A fully updated test resource",
			"status":      "active",
			"key_field":   "resource_id",
			"data": map[string]any{
				"test": "updated",
			},
		},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusOK

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			m := map[string]any{}

			if err := json.Unmarshal(b, &m); err != nil {
				t.Errorf("Unexpected error decoding response: %v", err)
			}

			version, ok := m["version"].(string)
			if !ok || version != "2" {
				t.Errorf("Expected updated version in response: %v", m)
			}
		},
	}, {
		name:   "get resource tags",
		url:    "http://localhost:8080/api/v1/resources/{{resource_id}}/tags",
		method: http.MethodGet,
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusOK

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			var tags []string
			if err := json.Unmarshal(b, &tags); err != nil {
				t.Errorf("Unexpected error decoding response: %v", err)
			}
		},
	}, {
		name:   "create resource tags",
		url:    "http://localhost:8080/api/v1/resources/{{resource_id}}/tags",
		method: http.MethodPost,
		body:   []string{"test:tag1", "test:tag2"},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusCreated

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			var tags []string
			if err := json.Unmarshal(b, &tags); err != nil {
				t.Errorf("Unexpected error decoding response: %v", err)
			}

			if len(tags) != 2 {
				t.Errorf("Expected 2 tags, got: %v", len(tags))
			}
		},
	}, {
		name:   "delete resource tags",
		url:    "http://localhost:8080/api/v1/resources/{{resource_id}}/tags",
		method: http.MethodDelete,
		body:   []string{"test:tag1", "test:tag2"},
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusNoContent

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}
		},
	}, {
		name:   "delete resource",
		url:    "http://localhost:8080/api/v1/resources/{{resource_id}}",
		method: http.MethodDelete,
		resp: func(t *testing.T, res *http.Response) {
			expC := http.StatusNoContent

			if res.StatusCode != expC {
				t.Errorf("Status code expected: %v, got: %v",
					expC, res.StatusCode)
			}
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if strings.Contains(tt.url, "{{resource_id}}") {
				dataLock.Lock()
				resourceID, _ := data["resource_id"].(string)
				dataLock.Unlock()

				tt.url = strings.ReplaceAll(tt.url, "{{resource_id}}",
					resourceID)
			}

			buf := &bytes.Buffer{}

			if tt.body != nil {
				if ct, ok := tt.header["Content-Type"]; ok {
					if !strings.Contains("json", ct) {
						if bm, ok := tt.body.(map[string]any); ok {
							form := url.Values{}

							for k, v := range bm {
								switch vv := v.(type) {
								case string:
									form.Add(k, vv)
								default:
									b, err := json.Marshal(vv)
									if err != nil {
										t.Error(err)
									}

									form.Add(k, string(b))
								}
							}

							buf = bytes.NewBufferString(form.Encode())
						}
					}
				}

				if buf.Len() == 0 {
					b, err := json.Marshal(tt.body)
					if err != nil {
						t.Error(err)
					}

					buf = bytes.NewBuffer(b)
				}
			}

			var br io.Reader

			if buf.Len() > 0 {
				br = buf
			}

			r, err := http.NewRequest(tt.method, tt.url, br)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			dataLock.Lock()

			at, _ := data["access_token"].(string)

			dataLock.Unlock()

			if at != "" {
				if tt.header == nil {
					tt.header = map[string]string{}
				}

				tt.header["Authorization"] = "Bearer " + at
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			res, err := http.DefaultClient.Do(r)
			if err != nil {
				t.Errorf("Unexpected client error: %v", err)
			}

			defer res.Body.Close()

			tt.resp(t, res)
		})
	}
}
