package common

import "errors"

var (
	ErrProfileNotFound = errors.New("no profile found for keycloak id")
	ErrUserNotFound    = errors.New("no profile found")
	ErrNotFound        = errors.New("not found")
	ErrSpouseConflict  = errors.New("one or both users already have a spouse that is not the intended new spouse")
	ErrSpouseSelf      = errors.New("cannot set spouse to self")
)
