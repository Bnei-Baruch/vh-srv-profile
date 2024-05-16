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
	email := c.Query("email")
	typeFilter := c.Query("type")
	orderByCreatedAt := c.Query("o_created_at")

	if orderByCreatedAt != "" && orderByCreatedAt != "asc" && orderByCreatedAt != "desc" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid value for o_created_at"})
		return
	}

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

	res, err := p.repo.GetMultipleRequest(c.Request.Context(), intSkip, intLimit, kcid, status, name, email, typeFilter, orderByCreatedAt)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetMultipleRequest: %w", err))
		return
	}
	totalCount := len(res)
	if totalCount > 0 {
		totalCount, err = p.repo.GetMultipleRequestCount(c.Request.Context(), kcid, status, name, email, typeFilter, orderByCreatedAt)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.GetMultipleRequestCount: %w", err))
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Fetched!", "totalCount": totalCount, "data": res})
}
