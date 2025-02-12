package tests_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/dhaifley/apigo"
)

func TestTests(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}

	t.Parallel()

	svc := apigo.New()

	ctx := context.Background()

	errCh := make(chan error, 1)

	go func(ctx context.Context, errCh chan error) {
		if err := svc.Start(ctx); err != nil {
			t.Error("server error", "error", err)
		}
	}(ctx, errCh)

	t.Cleanup(func() {
		svc.Close(ctx)
	})

	time.Sleep(time.Second)

	tests := []struct {
		name   string
		url    string
		header map[string]string
		code   int
		resp   string
	}{{
		name:   "unauthorized",
		url:    "http://localhost:8080/api/v1/account",
		header: map[string]string{"Authorization": "test"},
		code:   http.StatusUnauthorized,
		resp:   `"Unauthorized"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := http.NewRequest(http.MethodGet, tt.url, nil)
			if err != nil {
				t.Fatal("Failed to initialize request", err)
			}

			for th, tv := range tt.header {
				r.Header.Set(th, tv)
			}

			res, err := http.DefaultClient.Do(r)
			if err != nil {
				t.Errorf("Unexpected client error: %v", err)
			}

			if res.StatusCode != tt.code {
				t.Errorf("Code expected: %v, got: %v", tt.code, res.StatusCode)
			}

			defer res.Body.Close()

			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("Unexpected response error: %v", err)
			}

			if !strings.Contains(string(b), tt.resp) {
				t.Errorf("Expected body to contain: %v, got: %v",
					tt.resp, string(b))
			}
		})
	}
}
