package server_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dhaifley/apigo/internal/server"
)

func TestGetHealthCheck(t *testing.T) {
	t.Parallel()

	svr, err := server.NewServer(nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		w    *httptest.ResponseRecorder
		url  string
		code int
		resp string
	}{{
		name: "success",
		w:    httptest.NewRecorder(),
		url:  basePath + "/health",
		code: http.StatusOK,
		resp: `{"service":"api","health":200}`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequest(http.MethodGet, tt.url, nil)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
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
