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
	membershipID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if !p.HasAnyRole(c, common.RoleAnyAdmin...) {
		return
	}

	res, dbErr := p.repo.GetMembershipByID(c.Request.Context(), membershipID)
	if dbErr != nil {
		if errors.Is(dbErr, common.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetMembershipByID: %w", dbErr))
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

	if !p.isSubjectOrHasAnyRole(c, kcID, common.RoleAnyAdmin...) {
		return
	}

	res, dbErr := p.repo.GetMembershipByKCID(c.Request.Context(), kcID)
	if dbErr != nil {
		if errors.Is(dbErr, common.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetMembershipByKCID: %w", dbErr))
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

	if !p.isUserOrHasAnyRole(c, userID, common.RoleAnyAdmin...) {
		return
	}

	res, dbErr := p.repo.GetMembershipByUserID(c.Request.Context(), userID)
	if dbErr != nil {
		if errors.Is(dbErr, common.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetMembershipByUserID: %w", dbErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Fetched!", "data": res})
}

func (p *ProfileManager) handleMembershipPatchByID(c *gin.Context) {
	id := c.Param("id")
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

	if !p.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}

	_, patchErr := p.repo.PatchMembershipByID(c.Request.Context(), membership, membershipID)
	if patchErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.PatchMembershipByID: %w", patchErr))
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

	if membCancel.KeycloakID == nil && membCancel.UserID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body (at least one of keycloak_id or user_id is required)"})
		return
	}

	if membCancel.KeycloakID != nil && *membCancel.KeycloakID != "" {
		_, err := uuid.FromString(*membCancel.KeycloakID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("malformed keycloak_id: %v", err)})
			return
		}

		if !p.isSubjectOrHasAnyRole(c, *membCancel.KeycloakID, common.RoleRoot, common.RoleAdmin) {
			return
		}
	}

	if membCancel.UserID != nil && *membCancel.UserID != "" {
		_, userIDErr := uuid.FromString(*membCancel.UserID)
		if userIDErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("malformed user_id: %v", userIDErr)})
			return
		}

		if !p.isUserOrHasAnyRole(c, *membCancel.UserID, common.RoleRoot, common.RoleAdmin) {
			return
		}
	}

	err := p.repo.CancelMembership(c.Request.Context(), membCancel)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.CancelMembership: %w", err))
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

	if evalbody.UserID == nil && evalbody.KeycloakID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body ( at least one of keycloak_id or user_id is required )"})
		return
	}

	if !p.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}

	userMembershipRes, evaluateErr := p.repo.EvaluateMembershipByUserID(c.Request.Context(), evalbody)
	if evaluateErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.EvaluateMembershipByUserID: %w", evaluateErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Evaluated!", "data": userMembershipRes})
}

func (p *ProfileManager) handleMembershipSoftDeleteByID(c *gin.Context) {
	id := c.Param("id")
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

	if !p.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}

	patchErr := p.repo.SoftDeleteMembershipByID(c.Request.Context(), membershipID)
	if patchErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.SoftDeleteMembershipByID: %w", patchErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Soft deleted!", "data": ""})
}

func (p *ProfileManager) handleMembershipFetchAll(c *gin.Context) {
	skip := c.Query("skip")
	limit := c.Query("limit")
	userID := c.Query("user_id")

	if skip == "" {
		skip = "0"
	}
	if limit == "" {
		limit = "10"
	}

	intSkip, serr := strconv.Atoi(skip)
	if serr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid skip value! Accepted value is INTEGER"})
		return
	}

	intLimit, lerr := strconv.Atoi(limit)
	if lerr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value! Accepted value is INTEGER"})
		return
	}

	if !p.HasAnyRole(c, common.RoleAnyAdmin...) {
		return
	}

	res, err := p.repo.GetMultipleMembership(c.Request.Context(), intSkip, intLimit, userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetMultipleMembership(: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Fetched!", "data": res})
}
