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

func (p *ProfileManager) handleGrantFetchByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	grantID, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, dbErr := p.repo.GetGrantByIDAndUserID(c.Request.Context(), grantID, "")

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

func (p *ProfileManager) handleGrantCreate(c *gin.Context) {

	var grant repo.GrantAndGrantMembership

	if err := c.ShouldBindJSON(&grant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Checking if the grant type is not nil and if it is not equal to membership.
	// As of now we only have membership grant type.
	if grant.Type == nil || *grant.Type != "hhmembership" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid grant type"})
		return
	}

	ID, dbErr := p.repo.CreateGrant(c.Request.Context(), grant)

	if dbErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while creating grant: %w", dbErr))
		return
	}

	grant.GrantID = &ID

	if *grant.Type == "hhmembership" {

		grant.MonthsLeft = grant.Month
		grant.MonthsUsed = new(int)

		_, grantMembErr := p.repo.CreateGrantMembership(c.Request.Context(), grant)

		if grantMembErr != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("error while creating grant: %w", grantMembErr))
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Created!", "data": ID})
}

func (p *ProfileManager) handleGrantPatchByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	grantID, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var grant repo.Grant

	if err := c.ShouldBindJSON(&grant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, patchErr := p.repo.PatchGrant(c.Request.Context(), grant, grantID, 0)

	if patchErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while updating grant: %w", patchErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Updated!", "data": grant})
}

func (p *ProfileManager) handleGrantSoftDeleteByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	grantID, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var grant repo.GrantAndGrantMembership

	if err := c.ShouldBindJSON(&grant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patchErr := p.repo.SoftDeleteGrantByID(c.Request.Context(), grantID)

	if patchErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while creating grant: %w", patchErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Soft deleted!", "data": ""})
}

func (p *ProfileManager) handleGrantFetchAll(c *gin.Context) {

	skip := c.Query("skip")
	limit := c.Query("limit")
	cancelled := c.Query("cancelled")
	userID := c.Query("user_id")
	grantType := c.Query("type")
	createdAt := c.Query("created_at")
	var boolCancelled *bool
	var boolCancelledErr error

	// check if cancelled is boolean
	if cancelled != "" {
		*boolCancelled, boolCancelledErr = strconv.ParseBool(cancelled)
		if boolCancelledErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cancelled"})
			return
		}
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
