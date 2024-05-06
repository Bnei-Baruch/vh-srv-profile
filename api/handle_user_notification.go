package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

func (p *ProfileManager) handleUserNotificationFetchByID(c *gin.Context) {
	id := c.Param("id")
	userNotificationId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if !p.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}

	res, dbErr := p.repo.GetUserNotificationByID(c.Request.Context(), userNotificationId)
	if dbErr != nil {
		if errors.Is(dbErr, common.ErrNotFound) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.GetUserNotificationByID: %w", dbErr))
		}
		return
	}

	if !p.isUserOrHasAnyRole(c, *res.UserID, common.RoleAnyAdmin...) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Fetched!", "data": res})
}

func (p *ProfileManager) handleUserNotificationCreate(c *gin.Context) {
	var noti repo.UserNotification
	if err := c.ShouldBindJSON(&noti); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if noti.UserID == nil || noti.NotificationID == nil || noti.Active == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id, notification_id and active are required"})
		return
	}

	if !p.isUserOrHasAnyRole(c, *noti.UserID, common.RoleAnyAdmin...) {
		return
	}

	dbErr := p.repo.CreateUserNotification(c.Request.Context(), noti)
	if dbErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.CreateUserNotification: %w", dbErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Created!", "data": noti})
}

func (p *ProfileManager) handleUserNotificationPatchByID(c *gin.Context) {
	id := c.Param("id")
	userNotificationId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var noti repo.UserNotification
	if err := c.ShouldBindJSON(&noti); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification, err := p.repo.GetUserNotificationByID(c.Request.Context(), userNotificationId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.GetUserNotificationByID: %w", err))
		}
		return
	}

	if !p.isUserOrHasAnyRole(c, *notification.UserID, common.RoleRoot, common.RoleAdmin) {
		return
	}

	_, err = p.repo.PatchUserNotification(c.Request.Context(), noti, userNotificationId)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.PatchUserNotification: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Updated!", "data": noti})
}

func (p *ProfileManager) handleUserNotificationSoftDeleteByID(c *gin.Context) {
	id := c.Param("id")
	userNotificationId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	notification, err := p.repo.GetUserNotificationByID(c.Request.Context(), userNotificationId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.GetUserNotificationByID: %w", err))
		}
		return
	}

	if !p.isUserOrHasAnyRole(c, *notification.UserID, common.RoleRoot, common.RoleAdmin) {
		return
	}

	softDelErr := p.repo.SoftDeleteUserNotification(c.Request.Context(), userNotificationId)
	if softDelErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.SoftDeleteUserNotification: %w", softDelErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Soft deleted!", "data": ""})
}

func (p *ProfileManager) handleUserNotificationFetchAll(c *gin.Context) {
	skip := c.Query("skip")
	limit := c.Query("limit")

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

	if !p.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}

	res, err := p.repo.GetMultipleUserNotification(c.Request.Context(), intSkip, intLimit)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetMultipleUserNotification: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Fetched!", "data": res})
}
