package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type updateRequestStorage interface {
	updateRequest(ctx context.Context, keycloakID int, toUpdate newRequest) error
}

func (p *profileManager) updateRequest(c *gin.Context) {
	id, ok := c.Params.Get("id")

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID missing"})
		return
	}

	// String conversion to int
	intID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id! Accepted value is INTEGER"})
		return
	}

	var request newRequest
	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	if *request.Status != "REQUESTED" && *request.Status != "APPROVED" && *request.Status != "DENIED" {
		err := fmt.Errorf("invalid status: %s", *request.Status)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	if request.Type != nil {
		if *request.Type != "hhticket" && *request.Type != "hhmembership" && *request.Type != "arvut" {
			err := fmt.Errorf("invalid type: %s", *request.Type)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			_ = c.Error(err)
			return
		}
	}

	if err := p.requestUpdater.updateRequest(c.Request.Context(), intID, request); err != nil {
		if errors.Is(err, errNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while updating request %q: %w", intID, err))
		return
	}

	c.Status(http.StatusOK)
}
