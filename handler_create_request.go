package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createRequestStorage interface {
	createRequest(ctx context.Context, request newRequest) error
}

func (p *profileManager) createRequest(c *gin.Context) {
	var request newRequest

	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	if request.KeycloakId == nil || request.RequestName == nil || request.Status == nil {
		err := fmt.Errorf("missing a required field in provided request: %#v", request)
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

	if err := p.requestCreator.createRequest(c.Request.Context(), request); err != nil {
		_ = c.Error(fmt.Errorf("error while creating request %q: %w", *request.KeycloakId, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}
