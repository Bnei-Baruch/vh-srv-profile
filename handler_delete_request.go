package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type deleteRequestStorage interface {
	deleteRequest(ctx context.Context, keycloakID int) error
}

func (p *profileManager) deleteRequest(c *gin.Context) {
	id, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keycloak ID missing"})
		return
	}
	// String conversion to int
	intID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id! Accepted value is INTEGER"})
		return
	}

	if err := p.requestDeleter.deleteRequest(c.Request.Context(), intID); err != nil {
		if errors.Is(err, errProfileNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while getting request %d: %w", intID, err))
		return
	}
}
