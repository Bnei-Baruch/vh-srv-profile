package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type operationInterface interface {
	createOperation(ctx context.Context, opr operationReq) (int, error)
}

type operationReq struct {
	ID            *int    `json:"id"`
	NewEmail      *string `json:"new_email"`
	NewKeycloakID *string `json:"new_keycloak_id"`
	OldKeycloakID *string `json:"old_keycloak_id"`
	Input         *string `json:"input"`
	Type          *string `json:"type"`
	Output        *string `json:"output"`
	Status        *string `json:"status"`
	Revert        *string `json:"revert"`
	RevertOutput  *string `json:"revert_output"`
}

func (p *profileManager) handleOperationCreate(c *gin.Context) {

	var opr operationReq

	if err := c.ShouldBindJSON(&opr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if opr.Type == nil || *opr.Type != "email_update" ||
		opr.NewEmail == nil || opr.NewKeycloakID == nil || opr.OldKeycloakID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
		return
	}

	ID, dbErr := p.operation.createOperation(c.Request.Context(), opr)

	if dbErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while creating grant: %w", dbErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Created!", "data": ID})
}
