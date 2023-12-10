package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

func (p *ProfileManager) handleMembershipFetchByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	membershipID, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, dbErr := p.repo.GetMembershipByID(c.Request.Context(), membershipID)

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

func (p *ProfileManager) handleMembershipFetchByKCID(c *gin.Context) {
	kcID := c.Param("kcid")
	if kcID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid kcid"})
		return
	}

	res, dbErr := p.repo.GetMembershipByKCID(c.Request.Context(), kcID)

	if dbErr != nil {
		if errors.Is(dbErr, common.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": dbErr.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while getting membership: %w", dbErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Fetched!", "data": res})
}

func (p *ProfileManager) handleMembershipFetchByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	res, dbErr := p.repo.GetMembershipByUserID(c.Request.Context(), userID)

	if dbErr != nil {
		if errors.Is(dbErr, common.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": dbErr.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while getting membership: %w", dbErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Fetched!", "data": res})
}

func (p *ProfileManager) handleMembershipPatchByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	membershipID, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var membership repo.Membership

	if err := c.ShouldBindJSON(&membership); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, patchErr := p.repo.PatchMembershipByID(c.Request.Context(), membership, membershipID)

	if patchErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while updating membership: %w", patchErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Updated!", "data": membership})
}

func (p *ProfileManager) handleMembershipCancellation(c *gin.Context) {
	var membCancel repo.EmailKeycloakAndUserIDBody

	if err := c.ShouldBindJSON(&membCancel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if membCancel.Email == nil && membCancel.KeycloakID == nil && membCancel.UserID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body ( at least one of email, keycloak_id or user_id is required )"})
		return
	}

	if membCancel.KeycloakID != nil && *membCancel.KeycloakID != "" {
		_, err := uuid.FromString(*membCancel.KeycloakID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			_ = c.Error(err)
			return
		}
	}

	if membCancel.UserID != nil && *membCancel.UserID != "" {
		_, userIDErr := uuid.FromString(*membCancel.UserID)
		if userIDErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": userIDErr.Error()})
			_ = c.Error(userIDErr)
			return
		}
	}

	err := p.repo.CancelMembership(c.Request.Context(), membCancel)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while updating membership: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Cancelled!"})
}

func (p *ProfileManager) handleMembershipEvaluationByUserID(c *gin.Context) {

	var evalbody repo.EmailKeycloakAndUserIDBody

	if err := c.ShouldBindJSON(&evalbody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if evalbody.UserID == nil && evalbody.KeycloakID == nil && evalbody.Email == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body ( at least one of email, keycloak_id or user_id is required )"})
		return
	}

	userMembershipRes, evaluateErr := p.repo.EvaluateMembershipByUserID(c.Request.Context(), evalbody)

	if evaluateErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while evaluating membership: %w", evaluateErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Evaluated!", "data": userMembershipRes})
}

func (p *ProfileManager) handleMembershipSoftDeleteByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	membershipID, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var mebership repo.Membership

	if err := c.ShouldBindJSON(&mebership); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patchErr := p.repo.SoftDeleteMembershipByID(c.Request.Context(), membershipID)

	if patchErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while creating membership: %w", patchErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Soft deleted!", "data": ""})
}

func (p *ProfileManager) handleMembershipFetchAll(c *gin.Context) {

	skip := c.Query("skip")
	limit := c.Query("limit")
	month := c.Query("month")
	year := c.Query("year")
	userID := c.Query("user_id")

	var (
		monthInt     int
		yearInt      int
		monthYearErr error
	)

	// month and year to int
	if month != "" {
		monthInt, monthYearErr = strconv.Atoi(month)
		if monthYearErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month"})
			return
		}
	}

	if year != "" {
		yearInt, monthYearErr = strconv.Atoi(year)
		if monthYearErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
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

	res, err := p.repo.GetMultipleMembership(c.Request.Context(), intSkip, intLimit, monthInt, yearInt, userID)
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
