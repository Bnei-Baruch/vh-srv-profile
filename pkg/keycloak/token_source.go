package keycloak

import (
	"fmt"
	"strings"
)

type TokenSource interface {
	Token() (string, error)
}

// AuthHeaderTokenSource will strip the token from the header and reuse it forever
func AuthHeaderTokenSource(authHeader string) TokenSource {
	parts := strings.Split(authHeader, " ")
	token := parts[1]
	if token == "" {
		return authHeaderTokenSource{token: "", err: fmt.Errorf("malformed auth header: %s", authHeader)}
	}
	return authHeaderTokenSource{token: token, err: nil}
}

type authHeaderTokenSource struct {
	token string
	err   error
}

func (s authHeaderTokenSource) Token() (string, error) {
	return s.token, s.err
}
