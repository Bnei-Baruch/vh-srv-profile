package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

func (p *ProfileManager) createRequest(c *gin.Context) {
	var request repo.NewRequest

	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.KeycloakId == nil || request.RequestName == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required field, keycloak_id and name are mandatory"})
		return
	}

	if !p.isSubjectOrHasAnyRole(c, *request.KeycloakId, common.RoleAnyAdmin...) {
		return
	}

	if err := p.repo.CreateRequest(c.Request.Context(), request); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.CreateRequest: %w", err))
		return
	}

	c.Status(http.StatusCreated)
}

func (p *ProfileManager) concludeRequest(c *gin.Context) {
	var conclusion repo.RequestConclusion

	if err := c.ShouldBind(&conclusion); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reqID, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("malformed request id: %v", err.Error())})
		return
	}

	if !p.HasAnyRole(c, common.RoleAnyAdmin...) {
		return
	}

	if err := p.repo.ConcludeRequest(c.Request.Context(), reqID, conclusion); err != nil {
		if errors.Is(err, common.ErrNotFound) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.ConcludeRequest: %w", err))
		}

		return
	}

	c.Status(http.StatusOK)
}
