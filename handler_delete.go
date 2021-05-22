package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type deleteStorage interface {
	deleteProfile(ctx context.Context, keycloakID uuid.UUID) error
}

func (p *profileManager) delete(c *gin.Context) {
	keycloakIDString, ok := c.Params.Get("keycloak_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keycloak ID missing"})
		return
	}
	keycloakID, err := uuid.FromString(keycloakIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	if err := p.deleter.deleteProfile(c.Request.Context(), keycloakID); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while getting user %q: %w", keycloakIDString, err))
		return
	}
}
