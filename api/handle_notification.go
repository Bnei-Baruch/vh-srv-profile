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

func (p *ProfileManager) handleNotificationFetchByID(c *gin.Context) {
	id := c.Param("id")
	notificationId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, err := p.repo.GetNotificationByID(c.Request.Context(), notificationId)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.GetNotificationByID: %w", err))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Fetched!", "data": res})
}

func (p *ProfileManager) handleNotificationCreate(c *gin.Context) {
	var noti repo.Notification
	if err := c.ShouldBindJSON(&noti); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if noti.Slug == nil || noti.Content == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and content are required"})
		return
	}

	if !p.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}

	ID, err := p.repo.CreateNotification(c.Request.Context(), noti)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.CreateNotification: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Created!", "data": ID})
}

func (p *ProfileManager) handleNotificationPatchByID(c *gin.Context) {
	id := c.Param("id")
	notificationId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var noti repo.Notification
	if err := c.ShouldBindJSON(&noti); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !p.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}

	_, err = p.repo.PatchNotification(c.Request.Context(), noti, notificationId)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.PatchNotification: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Updated!", "data": noti})
}

func (p *ProfileManager) handleNotificationSoftDeleteByID(c *gin.Context) {
	id := c.Param("id")
	notificationId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if !p.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}

	err = p.repo.SoftDeleteNotification(c.Request.Context(), notificationId)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.SoftDeleteNotification: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Soft deleted!", "data": ""})
}

func (p *ProfileManager) handleNotificationFetchAll(c *gin.Context) {
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

	res, err := p.repo.GetMultipleNotification(c.Request.Context(), intSkip, intLimit)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetMultipleNotification: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Fetched!", "data": res})
}
