package request_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/dhaifley/apid/internal/request"
)

func TestContextService(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(),
		request.CtxKeyService, "test")

	svc, err := request.ContextService(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if svc != "test" {
		t.Errorf("Expected service: test, got: %v", svc)
	}
}

func TestContextRemote(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(),
		request.CtxKeyRemote, "1.1.1.1")

	addr, err := request.ContextRemote(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if addr != "1.1.1.1" {
		t.Errorf("Expected IP: 1.1.1.1, got: %v", addr)
	}
}

func TestContextTraceID(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(),
		request.CtxKeyTraceID, "test")

	tID, err := request.ContextTraceID(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if tID != "test" {
		t.Errorf("Expected trace ID: test, got: %v", tID)
	}
}

func TestContextRequestURL(t *testing.T) {
	t.Parallel()

	r, err := url.Parse("http://test.com")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.WithValue(context.Background(), request.CtxKeyRequestURL, r)

	u, err := request.ContextRequestURL(ctx)
	if err != nil {
		t.Fatal(err)
	}

	exp := "test.com"
	if u.Host != exp {
		t.Errorf("Expected host: %v, got: %v", exp, u.Host)
	}
}

func TestContextRequestBody(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(),
		request.CtxKeyRequestBody, "test")

	b, err := request.ContextRequestBody(ctx)
	if err != nil {
		t.Fatal(err)
	}

	exp := "test"
	if b != exp {
		t.Errorf("Expected body: %v, got: %v", exp, b)
	}
}

func TestContextJWT(t *testing.T) {
	t.Parallel()

	exp := "test"

	ctx := context.WithValue(context.Background(), request.CtxKeyJWT, exp)

	val, err := request.ContextJWT(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if val != exp {
		t.Errorf("Expected value: %v, got: %v", exp, val)
	}
}

func TestContextAccountID(t *testing.T) {
	t.Parallel()

	exp := "test"

	ctx := context.WithValue(context.Background(), request.CtxKeyAccountID, exp)

	val, err := request.ContextAccountID(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if val != exp {
		t.Errorf("Expected value: %v, got: %v", exp, val)
	}
}

func TestContextAccountName(t *testing.T) {
	t.Parallel()

	exp := "test"

	ctx := context.WithValue(context.Background(),
		request.CtxKeyAccountName, exp)

	val, err := request.ContextAccountName(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if val != exp {
		t.Errorf("Expected value: %v, got: %v", exp, val)
	}
}

func TestContextUserID(t *testing.T) {
	t.Parallel()

	exp := "test"

	ctx := context.WithValue(context.Background(), request.CtxKeyUserID, exp)

	val, err := request.ContextUserID(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if val != exp {
		t.Errorf("Expected value: %v, got: %v", exp, val)
	}
}

func TestContextTokenID(t *testing.T) {
	t.Parallel()

	exp := "test"

	ctx := context.WithValue(context.Background(), request.CtxKeyTokenID, exp)

	val, err := request.ContextTokenID(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if val != exp {
		t.Errorf("Expected value: %v, got: %v", exp, val)
	}
}
