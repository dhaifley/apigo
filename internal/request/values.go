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

// ValidTokenID checks whether a string is a valid token ID.
func ValidTokenID(id string) bool {
	validChars := "1234567890abcdef-"

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
