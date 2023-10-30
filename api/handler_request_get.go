package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

func (p *ProfileManager) getRequests(c *gin.Context) {

	// Fetching all the query strings if present in url
	skip := c.Query("skip")
	limit := c.Query("limit")
	kcid := c.Query("kcid")
	status := c.Query("status")
	name := c.Query("name")
	typeFilter := c.Query("type")
	orderByCreatedAt := c.Query("o_created_at")

	if orderByCreatedAt != "" && orderByCreatedAt != "asc" && orderByCreatedAt != "desc" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value for o_created_at"})
		return
	}
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

	res, err := p.repo.GetMultipleRequest(c.Request.Context(), intSkip, intLimit, kcid, status, name, typeFilter, orderByCreatedAt)
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
