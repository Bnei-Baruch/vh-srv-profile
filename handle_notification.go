package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type notificationInterface interface {
	getNotificationByID(ctx context.Context, id int) (notificationRes, error)
	createNotification(ctx context.Context, noti notification) (int, error)
	getMultipleNotification(ctx context.Context, intSkip int, intLimit int) ([]notificationRes, error)
	patchNotification(ctx context.Context, noti notification, id int) (int, error)
	softDeleteNotification(ctx context.Context, id int) error
}

type notification struct {
	Slug    *string `json:"slug"`
	Content *string `json:"content"`
}

func (p *profileManager) handleNotificationFetchByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	notificationId, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, dbErr := p.notification.getNotificationByID(c.Request.Context(), notificationId)

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

func (p *profileManager) handleNotificationCreate(c *gin.Context) {

	var noti notification

	if err := c.ShouldBindJSON(&noti); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if noti.Slug == nil || noti.Content == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug and content are required"})
		return
	}

	ID, dbErr := p.notification.createNotification(c.Request.Context(), noti)

	if dbErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while creating notification: %w", dbErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Created!", "data": ID})
}

func (p *profileManager) handleNotificationPatchByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	notificationId, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var noti notification

	if err := c.ShouldBindJSON(&noti); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, patchErr := p.notification.patchNotification(c.Request.Context(), noti, notificationId)

	if patchErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while patching notification: %w", patchErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Updated!", "data": noti})
}

func (p *profileManager) handleNotificationSoftDelete(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	notificationId, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	softDelErr := p.notification.softDeleteNotification(c.Request.Context(), notificationId)

	if softDelErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while soft deleting notification: %w", softDelErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Soft deleted!", "data": ""})
}

func (p *profileManager) handleNotificationFetchAll(c *gin.Context) {

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

	res, err := p.notification.getMultipleNotification(c.Request.Context(), intSkip, intLimit)
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
