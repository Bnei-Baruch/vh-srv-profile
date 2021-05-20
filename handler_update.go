package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *profileManager) update(c *gin.Context) {
	keycloakID, ok := c.Params.Get("keycloak_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keycloak ID missing"})
		return
	}

	var request userInput
	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	if request.keycloakID != nil {
		err := fmt.Errorf("keycloakID cannot be updated")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	if err := p.updater.updateProfile(c.Request.Context(), keycloakID, request); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while updating user %q: %w", keycloakID, err))
		return
	}

	c.Status(http.StatusCreated)
}
