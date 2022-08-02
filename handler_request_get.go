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

type readMultipleRequestStorage interface {
	getMultipleRequest(ctx context.Context, intSkip int, intLimit int, kcid string, status string, name string) ([]requestResponse, error)
}
type requestResponse struct {
	ID            *int       `json:"id"`
	RequestName   *string    `json:"name" db:"name"`
	KeycloakID    *string    `json:"keycloak_id" db:"keycloak_id"`
	Status        *string    `json:"status" db:"status"`
	EventSlug     *string    `json:"event_slug" db:"event_slug"`
	Type          *string    `json:"type" db:"type"`
	RequestNote   *string    `json:"request_note" db:"request_note"`
	RejectionNote *string    `json:"rejection_note" db:"rejection_note"`
	CreatedAt     *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at" db:"updated_at"`
}

func (p *profileManager) getRequest(c *gin.Context) {

	// Fetching all the query strings if present in url
	skip := c.Query("skip")
	limit := c.Query("limit")
	kcid := c.Query("kcid")
	status := c.Query("status")
	name := c.Query("name")
	// fetch all the users based on parameters provided

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

	res, err := p.fetchRequests.getMultipleRequest(c.Request.Context(), intSkip, intLimit, kcid, status, name)
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
