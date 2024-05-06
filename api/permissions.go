package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.bbdev.team/vh/vh-srv-profile/api/middleware"
	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

func (p *ProfileManager) HasAnyRole(c *gin.Context, roles ...string) bool {
	claims := c.Request.Context().Value(common.CtxAuthClaims).(*middleware.IDTokenClaims)

	if !claims.HasAnyRole(roles...) {
		c.Status(http.StatusForbidden)
		return false
	}

	return true
}

func (p *ProfileManager) isSubjectOrHasAnyRole(c *gin.Context, keycloakID string, roles ...string) bool {
	claims := c.Request.Context().Value(common.CtxAuthClaims).(*middleware.IDTokenClaims)

	if claims.Sub != keycloakID && !claims.HasAnyRole(roles...) {
		c.Status(http.StatusForbidden)
		return false
	}

	return true
}

func (p *ProfileManager) isUserOrHasAnyRole(c *gin.Context, userID string, roles ...string) bool {
	claims := c.Request.Context().Value(common.CtxAuthClaims).(*middleware.IDTokenClaims)

	if claims.HasAnyRole(roles...) {
		return true
	}

	match, err := p.repo.IsSubjectID(c.Request.Context(), claims.Sub, userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.IsSubjectID: %w", err))
		return false
	}

	if !match {
		c.Status(http.StatusForbidden)
		return false
	}

	return true
}
