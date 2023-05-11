package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type operationInterface interface {
	performOperation(ctx context.Context, opr operationReq) (int, error)
	revertOperation(ctx context.Context, id int) error
}

type operationReq struct {
	ID            *int    `json:"id"`
	NewEmail      *string `json:"new_email"`
	OldEmail      *string `json:"old_email"`
	NewKeycloakID *string `json:"new_keycloak_id"`
	OldKeycloakID *string `json:"old_keycloak_id"`
	Input         *string `json:"input"`
	Type          *string `json:"type"`
	Output        *string `json:"output"`
	Status        *string `json:"status"`
	Revert        *string `json:"revert"`
	ToRevert      *bool   `json:"to_revert"`
}

type operationTrace struct {
	ID     *int    `json:"id"`
	Input  *string `json:"input"`
	Output *string `json:"output"`
	Type   *string `json:"type"`
	Status *string `json:"status"`
	Revert *string `json:"revert"`
}

// handleOperationRevert
func (p *profileManager) handleOperationRevert(c *gin.Context) {

	idStr, ok := c.Params.Get("id")

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id missing"})
		return
	}

	// convert id to int
	id, err := strconv.Atoi(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	revertErr := p.operation.revertOperation(c.Request.Context(), id)

	if revertErr != nil {
		_ = c.Error(fmt.Errorf("error while reverting operation: %w", revertErr))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Reverted!"})
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

	if opr.NewKeycloakID == nil || opr.OldKeycloakID == nil {
		if opr.NewKeycloakID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "new keycloak id missing"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "old keycloak id missing"})
		}
	}

	// check if both keycloak ids are same

	if *opr.NewKeycloakID == *opr.OldKeycloakID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "both keycloak ids are same"})
		return
	}

	ID, dbErr := p.operation.performOperation(c.Request.Context(), opr)

	if dbErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while creating grant: %w", dbErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Created!", "data": ID})
}
