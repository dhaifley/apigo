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

func TestAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}

	t.Parallel()

	data := map[string]any{}

	dataLock := sync.Mutex{}

	tests := []struct {
		name   string
		url    string
		method string
		header map[string]string
		body   map[string]any
		resp   func(t *testing.T, res *http.Response)
	}{{
		name:   "unauthorized",
		url:    "http://localhost:8080/api/v1/account",
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

			dataLock.Unlock()
		},
	}, {
		name:   "get account",
		url:    "http://localhost:8080/api/v1/account",
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

			expB := `"account_id":"`

			if !strings.Contains(string(b), expB) {
				t.Errorf("Expected body to contain: %v, got: %v",
					expB, string(b))
			}
		},
	}, {
		name:   "post account",
		url:    "http://localhost:8080/api/v1/account",
		method: http.MethodPost,
		body: map[string]any{
			"account_id": "test-account",
			"name":       "test-account",
			"status":     "active",
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

			expB := `"account_id":"`

			if !strings.Contains(string(b), expB) {
				t.Errorf("Expected body to contain: %v, got: %v",
					expB, string(b))
			}
		},
	}, {
		name:   "create account repo",
		url:    "http://localhost:8080/api/v1/account/repo",
		method: http.MethodPost,
		body: map[string]any{
			"repo":        "https://github.com/test/test.git",
			"repo_status": "active",
			"repo_status_data": map[string]any{
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

			if _, ok := m["repo_status"].(string); !ok {
				t.Errorf("Expected repo_status field in response: %v", m)
			}
		},
	}, {
		name:   "get account repo",
		url:    "http://localhost:8080/api/v1/account/repo",
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

			if _, ok := m["repo_status"].(string); !ok {
				t.Errorf("Expected repo_status field in response: %v", m)
			}
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}

			if len(tt.body) > 0 {
				if ct, ok := tt.header["Content-Type"]; ok {
					if !strings.Contains("json", ct) {
						form := url.Values{}

						for k, v := range tt.body {
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
