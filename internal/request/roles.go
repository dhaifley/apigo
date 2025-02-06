package request

import (
	"context"

	"github.com/dhaifley/apid/internal/errors"
)

// Valid authorization roles values.
const (
	RoleSystemAdmin = "system_admin"
	RoleAdmin       = "admin"
	RoleUser        = "user"
	RoleRefresh     = "refresh"
)

// ContextAuthUser provides a way to check a request context for sufficient
// user grants and return the userID.
func ContextAuthUser(ctx context.Context) (string, error) {
	if !ContextHasRole(ctx, RoleUser) &&
		!ContextHasRole(ctx, RoleAdmin) &&
		!ContextHasRole(ctx, RoleSystemAdmin) {
		return "", errors.New(errors.ErrForbidden,
			"unauthorized request")
	}

	if _, err := ContextAccountID(ctx); err != nil {
		return "", errors.New(errors.ErrForbidden,
			"unable to retrieve account id")
	}

	userID, err := ContextUserID(ctx)
	if err != nil {
		return "", errors.New(errors.ErrForbidden,
			"unable to retrieve user id")
	}

	return userID, nil
}

// ContextAuthRefresh provides a way to check a request context for sufficient
// refresh grants and return the userID of the token.
func ContextAuthRefresh(ctx context.Context) (string, error) {
	if !ContextHasRole(ctx, RoleRefresh) &&
		!ContextHasRole(ctx, RoleSystemAdmin) {
		return "", errors.New(errors.ErrForbidden,
			"unauthorized request")
	}

	if _, err := ContextAccountID(ctx); err != nil {
		return "", errors.New(errors.ErrForbidden,
			"unable to retrieve account id")
	}

	userID, err := ContextUserID(ctx)
	if err != nil {
		return "", errors.New(errors.ErrForbidden,
			"unable to retrieve user id")
	}

	return userID, nil
}

// ContextAuthAdmin provides a way to check a request context for sufficient
// administrative grants and return the userID.
func ContextAuthAdmin(ctx context.Context) (string, error) {
	if !ContextHasRole(ctx, RoleAdmin) &&
		!ContextHasRole(ctx, RoleSystemAdmin) {
		return "", errors.New(errors.ErrForbidden,
			"unauthorized request")
	}

	if _, err := ContextAccountID(ctx); err != nil {
		return "", errors.New(errors.ErrForbidden,
			"unable to retrieve account id")
	}

	userID, err := ContextUserID(ctx)
	if err != nil {
		return "", errors.New(errors.ErrForbidden,
			"unable to retrieve user id")
	}

	return userID, nil
}

// ContextAuthSysAdmin provides a way to check a request context for sufficient
// system administrative grants and return the userID.
func ContextAuthSysAdmin(ctx context.Context) (string, error) {
	if !ContextHasRole(ctx, RoleSystemAdmin) {
		return "", errors.New(errors.ErrForbidden,
			"unauthorized request")
	}

	if _, err := ContextAccountID(ctx); err != nil {
		return "", errors.New(errors.ErrForbidden,
			"unable to retrieve account id")
	}

	userID, err := ContextUserID(ctx)
	if err != nil {
		return "", errors.New(errors.ErrForbidden,
			"unable to retrieve user id")
	}

	return userID, nil
}
