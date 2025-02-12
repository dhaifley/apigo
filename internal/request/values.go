package request

import (
	"strings"

	"github.com/google/uuid"
)

// Valid status field values.
const (
	StatusActive       = "active"
	StatusInactive     = "inactive"
	StatusNew          = "new"
	StatusError        = "error"
	StatusPending      = "pending"
	StatusCanceled     = "canceled"
	StatusRemove       = "remove"
	StatusRemoving     = "removing"
	StatusModified     = "modified"
	StatusBusy         = "busy"
	StatusAlerting     = "alerting"
	StatusRecovered    = "recovered"
	StatusRunning      = "running"
	StatusStopped      = "stopped"
	StatusStopping     = "stopping"
	StatusFailed       = "failed"
	StatusSuccess      = "success"
	StatusMaintenance  = "maintenance"
	StatusActivating   = "activating"
	StatusDeactivating = "deactivating"
	StatusDisconnected = "disconnected"
	StatusImporting    = "importing"
)

// Valid system entities.
const (
	DefaultAccount = "default"
	SystemAccount  = "sys"
	SystemUser     = "sys"
)

// Valid scopes.
const (
	ScopeSuperUser     = "superuser"
	ScopeAccountRead   = "account:read"
	ScopeAccountWrite  = "account:write"
	ScopeAccountAdmin  = "account:admin"
	ScopeUserRead      = "user:read"
	ScopeUserWrite     = "user:write"
	ScopeUserAdmin     = "user:admin"
	ScopeResourceRead  = "resource:read"
	ScopeResourceWrite = "resource:write"
	ScopeResourceAdmin = "resource:admin"
)

// Scopes is a slice of all valid scopes.
var Scopes = []string{
	ScopeSuperUser,
	ScopeAccountRead,
	ScopeAccountWrite,
	ScopeAccountAdmin,
	ScopeUserRead,
	ScopeUserWrite,
	ScopeUserAdmin,
	ScopeResourceRead,
	ScopeResourceWrite,
	ScopeResourceAdmin,
}

// ValidAccountID checks whether a string is a valid account ID.
func ValidAccountID(id string) bool {
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"1234567890-_/\\:."

	if len(id) == 0 {
		return false
	}

	for _, r := range id {
		if !strings.ContainsRune(validChars, r) {
			return false
		}
	}

	return true
}

// ValidAccountName checks whether a string is a valid account name.
func ValidAccountName(name string) bool {
	return ValidAccountID(name)
}

// ValidUserID checks whether a string is a valid user ID.
func ValidUserID(id string) bool {
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"1234567890!#$%&'*+-/=?^_`" + `{|}~"(),:;<>@[\].@`

	if len(id) == 0 {
		return false
	}

	for _, r := range id {
		if !strings.ContainsRune(validChars, r) {
			return false
		}
	}

	return true
}

// ValidResourceID checks whether a string is a valid external resource ID.
func ValidResourceID(id string) bool {
	if _, err := uuid.Parse(id); err != nil {
		return false
	}

	return true
}

// ValidScope checks whether a string is a valid scope.
func ValidScope(scope string) bool {
	for _, s := range Scopes {
		if scope == s {
			return true
		}
	}

	return false
}

// ValidScopes checks whether a string is a valid scope.
func ValidScopes(scopes string) bool {
	s := strings.Split(scopes, " ")

	for _, scope := range s {
		if !ValidScope(scope) {
			return false
		}
	}

	return true
}
