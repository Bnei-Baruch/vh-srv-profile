package common

import "errors"

var (
	ErrProfileNotFound = errors.New("no profile found for keycloak id")
	ErrUserNotFound    = errors.New("no profile found")
	ErrNotFound        = errors.New("not found")
	ErrSpouseConflict  = errors.New("one or both users already have a spouse that is not the intended new spouse")
	ErrSpouseSelf      = errors.New("cannot set spouse to self")

	// ErrOperationNotFound: no operation_trace row for this email pair.
	ErrOperationNotFound = errors.New("operation not found")
	// ErrOperationReverted: the latest operation for this email pair was already reverted.
	ErrOperationReverted = errors.New("operation already reverted")
	// ErrOperationNoMatch: no user has the email and keycloak id the rename
	// (or its revert) starts from, so nothing was changed.
	ErrOperationNoMatch = errors.New("no user matches this email and keycloak id")
)
