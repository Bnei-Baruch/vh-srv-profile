package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

func (p *ProfileManager) hardDelete(c *gin.Context) {
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

	if !p.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}

	if err := p.repo.HardDeleteProfile(c.Request.Context(), keycloakID); err != nil {
		if errors.Is(err, common.ErrProfileNotFound) {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.HardDeleteProfile: %w", err))
		return
	}

	c.Status(http.StatusOK)
}
