package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
)

func (p *ProfileManager) handleGrantFetchByID(c *gin.Context) {
	id := c.Param("id")

	// convert id to int
	grantID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, dbErr := p.repo.GetGrantByID(c.Request.Context(), grantID)
	if dbErr != nil {
		if errors.Is(dbErr, common.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": dbErr.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while getting users: %w", dbErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Fetched!", "data": res})
}

func (p *ProfileManager) handleGrantFetchAll(c *gin.Context) {

	skip := c.Query("skip")
	limit := c.Query("limit")
	cancelled := c.Query("cancelled")
	userID := c.Query("user_id")
	grantType := c.Query("type")
	createdAt := c.Query("created_at")
	var boolCancelled *bool

	// check if cancelled is boolean
	if cancelled != "" {
		val, err := strconv.ParseBool(cancelled)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter [cancelled]"})
			return
		}
		boolCancelled = utils.PointerBool(val)
	}

	if skip == "" {
		skip = "0"
	}
	if limit == "" {
		limit = "10"
	}

	// String conversion to int
	intSkip, serr := strconv.Atoi(skip)
	if serr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid skip value! Accepted value is INTEGER"})
		return
	}

	// String conversion to int
	intLimit, lerr := strconv.Atoi(limit)
	if lerr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value! Accepted value is INTEGER"})
		return
	}

	res, err := p.repo.GetMultipleGrant(c.Request.Context(), intSkip, intLimit, boolCancelled, userID, grantType, createdAt)
	if err != nil {
		if errors.Is(err, common.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while getting users: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Fetched!", "data": res})
}
