// Package request contains functionality related to contexts and requests.
package request

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/dhaifley/apigo/internal/errors"
)

// ContextKey values are used to index context data.
type ContextKey int

const (
	// CtxKeyService is used to select the service name from a context.
	CtxKeyService ContextKey = iota

	// CtxKeyBuildTime is used to select the build time from a context.
	CtxKeyBuildTime

	// CtxKeyCommitID is used to select the commit ID of the running server
	// from a context.
	CtxKeyCommitID

	// CtxKeyRequestURL is used to select the request URL from a context.
	CtxKeyRequestURL

	// CtxKeyRequestBody is used to select the request body from a context.
	CtxKeyRequestBody

	// CtxKeyTraceID is used to select the trace ID from a context.
	CtxKeyTraceID

	// CtxKeySpanID is used to select the tracing span ID from a context.
	CtxKeySpanID

	// CtxKeyRemote is used to select a remote address from a context.
	CtxKeyRemote

	// CtxKeyJWT is used to select the authentication token from a context.
	CtxKeyJWT

	// CtxKeyScopes is used to select the authorization scopes from a context.
	CtxKeyScopes

	// CtxKeyAccountID is used to select the account id from a context.
	CtxKeyAccountID

	// CtxKeyAccountName is used to select the account name from a context.
	CtxKeyAccountName

	// CtxKeyUserID is used to select the user id from a context.
	CtxKeyUserID
)

// ContextService extracts the service name from the context.
func ContextService(ctx context.Context) (string, error) {
	service, ok := ctx.Value(CtxKeyService).(string)
	if !ok {
		return "", errors.New(errors.ErrContext,
			"unable to extract service name from context")
	}

	return service, nil
}

// ContextRemote extracts the remote address from the context.
func ContextRemote(ctx context.Context) (string, error) {
	addr, ok := ctx.Value(CtxKeyRemote).(string)
	if !ok {
		return "", errors.New(errors.ErrContext,
			"unable to extract remote address from context")
	}

	return addr, nil
}

// ContextTraceID extracts the trace id from the context.
func ContextTraceID(ctx context.Context) (string, error) {
	tID, ok := ctx.Value(CtxKeyTraceID).(string)
	if !ok {
		return "", errors.New(errors.ErrContext,
			"unable to extract trace ID from context")
	}

	return tID, nil
}

// ContextRequestURL retrieves the current request URL from the context.
func ContextRequestURL(ctx context.Context) (*url.URL, error) {
	url, ok := ctx.Value(CtxKeyRequestURL).(*url.URL)
	if !ok {
		return nil, errors.New(errors.ErrContext,
			"unable to retrieve request URL from context")
	}

	return url, nil
}

// ContextRequestBody retrieves the current request body from the context.
func ContextRequestBody(ctx context.Context) (string, error) {
	body, ok := ctx.Value(CtxKeyRequestBody).(string)
	if !ok {
		return "", errors.New(errors.ErrContext,
			"unable to retrieve request body from context")
	}

	return body, nil
}

// ContextJWT extracts the authentication token from the context.
func ContextJWT(ctx context.Context) (string, error) {
	token, ok := ctx.Value(CtxKeyJWT).(string)
	if !ok {
		return "", errors.New(errors.ErrContext,
			"unable to extract authentication token from context")
	}

	return token, nil
}

// ContextScopes extracts the authorization scopes from the context.
func ContextScopes(ctx context.Context) (string, error) {
	scopes, ok := ctx.Value(CtxKeyScopes).(string)
	if !ok {
		return "", errors.New(errors.ErrContext,
			"unable to extract authorization scopes from context")
	}

	return scopes, nil
}

// ContextHasScope tests whether the context contains a specified authorization
// scope.
func ContextHasScope(ctx context.Context, scope string) bool {
	scopes, ok := ctx.Value(CtxKeyScopes).(string)
	if !ok {
		return false
	}

	return strings.Contains(scopes, scope) ||
		strings.Contains(scopes, ScopeSuperUser)
}

// ContextAccountID extracts the account id from the context.
func ContextAccountID(ctx context.Context) (string, error) {
	id, ok := ctx.Value(CtxKeyAccountID).(string)
	if !ok {
		return "", errors.New(errors.ErrContext,
			"unable to extract account id from context")
	}

	return id, nil
}

// ContextAccountName extracts the account name from the context.
func ContextAccountName(ctx context.Context) (string, error) {
	id, ok := ctx.Value(CtxKeyAccountName).(string)
	if !ok {
		return "", errors.New(errors.ErrContext,
			"unable to extract account name from context")
	}

	return id, nil
}

// ContextUserID extracts the user id from the context.
func ContextUserID(ctx context.Context) (string, error) {
	id, ok := ctx.Value(CtxKeyUserID).(string)
	if !ok {
		return "", errors.New(errors.ErrContext,
			"unable to extract user id from context")
	}

	return id, nil
}

// ContextReplaceTimeout creates a copy of an existing context but with a new
// timeout.
func ContextReplaceTimeout(ctx context.Context,
	d time.Duration,
) (context.Context, context.CancelFunc) {
	newCtx, newCancel := context.WithTimeout(context.Background(), d)

	newCtx = context.WithValue(newCtx, CtxKeyService, ctx.Value(CtxKeyService))
	newCtx = context.WithValue(newCtx, CtxKeyBuildTime,
		ctx.Value(CtxKeyBuildTime))
	newCtx = context.WithValue(newCtx, CtxKeyCommitID,
		ctx.Value(CtxKeyCommitID))
	newCtx = context.WithValue(newCtx, CtxKeyRequestURL,
		ctx.Value(CtxKeyRequestURL))
	newCtx = context.WithValue(newCtx, CtxKeyRequestBody,
		ctx.Value(CtxKeyRequestBody))
	newCtx = context.WithValue(newCtx, CtxKeyTraceID, ctx.Value(CtxKeyTraceID))
	newCtx = context.WithValue(newCtx, CtxKeySpanID, ctx.Value(CtxKeySpanID))
	newCtx = context.WithValue(newCtx, CtxKeyRemote, ctx.Value(CtxKeyRemote))
	newCtx = context.WithValue(newCtx, CtxKeyJWT, ctx.Value(CtxKeyJWT))
	newCtx = context.WithValue(newCtx, CtxKeyScopes, ctx.Value(CtxKeyScopes))
	newCtx = context.WithValue(newCtx, CtxKeyAccountID,
		ctx.Value(CtxKeyAccountID))
	newCtx = context.WithValue(newCtx, CtxKeyAccountName,
		ctx.Value(CtxKeyAccountName))
	newCtx = context.WithValue(newCtx, CtxKeyUserID, ctx.Value(CtxKeyUserID))

	return newCtx, newCancel
}
