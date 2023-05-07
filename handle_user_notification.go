package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type userNotificationInterface interface {
	getUserNotificationByID(ctx context.Context, id int) (userNotificationRes, error)
	createUserNotification(ctx context.Context, noti userNotification) error
	getMultipleUserNotification(ctx context.Context, intSkip int, intLimit int) ([]userNotificationRes, error)
	patchUserNotification(ctx context.Context, noti userNotification, id int) (int, error)
	softDeleteUserNotification(ctx context.Context, id int) error
}

type userNotification struct {
	UserID         *string    `json:"user_id"`
	NotificationID *int       `json:"notification_id"`
	Active         *bool      `json:"active"`
	SeenAt         *time.Time `json:"seen_at"`
}

type userNotificationRes struct {
	ID *int `json:"id"`
	userNotification
	Slug      *string                 `json:"slug"`
	Content   *map[string]interface{} `json:"content"`
	CreatedAt *time.Time              `json:"created_at"`
	UpdatedAt *time.Time              `json:"updated_at"`
	DeletedAt *time.Time              `json:"deleted_at"`
}

func (p *profileManager) handleUserNotificationFetchByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	userNotificationId, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, dbErr := p.userNotification.getUserNotificationByID(c.Request.Context(), userNotificationId)

	if dbErr != nil {
		if errors.Is(dbErr, errNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": dbErr.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while getting users: %w", dbErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Fetched!", "data": res})
}

func (p *profileManager) handleUserNotificationCreate(c *gin.Context) {

	var noti userNotification

	if err := c.ShouldBindJSON(&noti); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if noti.UserID == nil || noti.NotificationID == nil || noti.Active == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id, notification_id and active are required"})
		return
	}

	dbErr := p.userNotification.createUserNotification(c.Request.Context(), noti)

	if dbErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while creating userNotification: %w", dbErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Created!", "data": noti})
}

func (p *profileManager) handleUserNotificationPatchByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	userNotificationId, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var noti userNotification

	if err := c.ShouldBindJSON(&noti); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, patchErr := p.userNotification.patchUserNotification(c.Request.Context(), noti, userNotificationId)

	if patchErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while patching userNotification: %w", patchErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Updated!", "data": noti})
}

func (p *profileManager) handleUserNotificationSoftDelete(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	userNotificationId, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	softDelErr := p.userNotification.softDeleteUserNotification(c.Request.Context(), userNotificationId)

	if softDelErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while soft deleting userNotification: %w", softDelErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Soft deleted!", "data": ""})
}

func (p *profileManager) handleUserNotificationFetchAll(c *gin.Context) {

	skip := c.Query("skip")
	limit := c.Query("limit")

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

	res, err := p.userNotification.getMultipleUserNotification(c.Request.Context(), intSkip, intLimit)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while getting users: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Fetched!", "data": res})
}
