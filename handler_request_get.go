package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type readMultipleRequestStorage interface {
	getMultipleRequest(ctx context.Context, intSkip int, intLimit int, kcid string, status string, name string) ([]requestResponse, error)
}
type requestResponse struct {
	ID            *int    `json:"id"`
	RequestName   *string `json:"request_name"`
	KeycloakID    *string `json:"keycloak_id"`
	Status        *string `json:"status"`
	RequestNote   *string `json:"request_note"`
	RejectionNote *string `json:"rejection_note"`
	CreatedAt     *string `json:"created_at"`
	UpdatedAt     *string `json:"updated_at"`
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
	intSkip, err := strconv.Atoi(skip)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid skip value! Accepted value is INTEGER"})
		return
	}

	// String conversion to int
	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value! Accepted value is INTEGER"})
		return
	}

	request, err := p.fetchRequests.getMultipleRequest(c.Request.Context(), intSkip, intLimit, kcid, status, name)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while getting users: %w", err))
		return
	}

	c.JSON(http.StatusOK, request)
}
