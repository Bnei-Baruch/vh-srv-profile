package keycloak

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestClaimsValid_UnexpiredTokenIsValid(t *testing.T) {
	claims := jwt.MapClaims{"exp": float64(time.Now().Add(time.Minute).Unix())}
	c := &ServiceClient{claims: &claims}

	assert.NoError(t, c.claimsValid())
}

func TestClaimsValid_ExpiredTokenIsInvalid(t *testing.T) {
	claims := jwt.MapClaims{"exp": float64(time.Now().Add(-time.Minute).Unix())}
	c := &ServiceClient{claims: &claims}

	assert.ErrorIs(t, c.claimsValid(), jwt.ErrTokenExpired)
}
