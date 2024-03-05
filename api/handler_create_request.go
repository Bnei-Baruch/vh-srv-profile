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
		_ = c.Error(err)
		return
	}

	if request.KeycloakId == nil || request.RequestName == nil {
		err := fmt.Errorf("missing a required field in provided request: %#v", request)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	if err := p.repo.CreateRequest(c.Request.Context(), request); err != nil {
		_ = c.Error(fmt.Errorf("error while creating request %q: %w", *request.KeycloakId, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

func (p *ProfileManager) concludeRequest(c *gin.Context) {
	var conclusion repo.RequestConclusion

	if err := c.ShouldBind(&conclusion); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	reqID, err := strconv.Atoi(c.Params.ByName("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprint("malformed request id: ", err.Error())})
		_ = c.Error(err)
		return
	}

	if err := p.repo.ConcludeRequest(c.Request.Context(), reqID, conclusion); err != nil {
		if errors.Is(err, common.ErrNotFound) {
			c.Status(http.StatusNotFound)
		} else {
			_ = c.Error(fmt.Errorf("error concluding request %d: %w", reqID, err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		return
	}

	c.Status(http.StatusOK)
}
