package common

import "fmt"

var (
	ErrProfileNotFound = fmt.Errorf("no profile found for keycloak id")
	ErrUserNotFound    = fmt.Errorf("no profile found")
	ErrNotFound        = fmt.Errorf("not found")
)
