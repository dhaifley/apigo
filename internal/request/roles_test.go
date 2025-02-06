package request_test

import (
	"context"
	"testing"

	"github.com/dhaifley/apid/internal/request"
)

func TestContextAuthUser(t *testing.T) {
	t.Parallel()

	exp := "test"

	ctx := context.WithValue(context.Background(), request.CtxKeyAccountID, exp)
	ctx = context.WithValue(ctx, request.CtxKeyUserID, exp)
	ctx = context.WithValue(ctx, request.CtxKeyRoles,
		[]string{request.RoleUser})

	val, err := request.ContextAuthUser(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if val != exp {
		t.Errorf("Expected value: %v, got: %v", exp, val)
	}
}

func TestContextAuthRefresh(t *testing.T) {
	t.Parallel()

	exp := "test"

	ctx := context.WithValue(context.Background(), request.CtxKeyAccountID, exp)
	ctx = context.WithValue(ctx, request.CtxKeyUserID, exp)
	ctx = context.WithValue(ctx, request.CtxKeyRoles,
		[]string{request.RoleRefresh})

	uID, err := request.ContextAuthRefresh(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if uID != exp {
		t.Errorf("Expected user ID: %v, got: %v", exp, uID)
	}
}

func TestContextAuthAdmin(t *testing.T) {
	t.Parallel()

	exp := "test"

	ctx := context.WithValue(context.Background(), request.CtxKeyAccountID, exp)
	ctx = context.WithValue(ctx, request.CtxKeyUserID, exp)
	ctx = context.WithValue(ctx, request.CtxKeyRoles,
		[]string{request.RoleAdmin})

	val, err := request.ContextAuthAdmin(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if val != exp {
		t.Errorf("Expected value: %v, got: %v", exp, val)
	}
}

func TestContextAuthSysAdmin(t *testing.T) {
	t.Parallel()

	exp := "test"

	ctx := context.WithValue(context.Background(), request.CtxKeyAccountID, exp)
	ctx = context.WithValue(ctx, request.CtxKeyUserID, exp)
	ctx = context.WithValue(ctx, request.CtxKeyRoles,
		[]string{request.RoleSystemAdmin})

	val, err := request.ContextAuthSysAdmin(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if val != exp {
		t.Errorf("Expected value: %v, got: %v", exp, val)
	}
}
