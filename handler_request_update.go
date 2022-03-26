package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type updateRequestStorage interface {
	updateRequest(ctx context.Context, keycloakID uuid.UUID, toUpdate newRequest) error
}

func (p *profileManager) updateRequest(c *gin.Context) {
	keycloakIDString, ok := c.Params.Get("keycloak_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keycloak ID missing"})
		return
	}
	keycloakID, err := uuid.FromString(keycloakIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var request newRequest
	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	if *request.Status != "REQUESTED" && *request.Status != "APPROVED" && *request.Status != "DENIED" {
		err := fmt.Errorf("invalid status: %s", *request.Status)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	if err := p.requestUpdater.updateRequest(c.Request.Context(), keycloakID, request); err != nil {
		if errors.Is(err, errNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while updating user %q: %w", keycloakID, err))
		return
	}

	c.Status(http.StatusOK)
}
