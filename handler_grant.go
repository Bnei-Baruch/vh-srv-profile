package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type grantInterface interface {
	getGrantByID(ctx context.Context, id int) (grantRes, error)
	createGrant(ctx context.Context, grant grantAndGrantMembership) (int, error)
	patchGrant(ctx context.Context, grant grantAndGrantMembership, id int) error
	createGrantMembership(ctx context.Context, grant grantAndGrantMembership) (int, error)
	patchGrantMembership(ctx context.Context, grant grantAndGrantMembership, grantId int) (int, error)
}

type grantMembeship struct {
	GrantID    *int       `json:"grant_id"`
	UserID     *uuid.UUID `json:"user_id"`
	Month      *int       `json:"nb_months"`
	MonthsUsed *int       `json:"months_used"`
	MonthsLeft *int       `json:"months_left"`
}

type grantMembeshipRes struct {
	grantMembeship
	ID        *int       `json:"id"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeleteAt  *time.Time `json:"delete_at"`
}

type grantAndGrantMembership struct {
	grant
	grantMembeship
}

type grant struct {
	Amount   *int    `json:"amount" db:"amount" validate:"required"`
	Currency *string `json:"currency" db:"currency" validate:"required"`
	Type     *string `json:"type" db:"type" validate:"required"`
	Loaned   *int    `json:"loaned" db:"loaned"`
	Granted  *int    `json:"granted" db:"granted"`
	Repayed  *int    `json:"repayed" db:"repayed"`
}

type grantRes struct {
	ID *int `json:"id" db:"id"`
	grant
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}

func (p *profileManager) handleGrantFetchByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	grantID, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, dbErr := p.grant.getGrantByID(c.Request.Context(), grantID)

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

func (p *profileManager) handleGrantCreate(c *gin.Context) {

	var grant grantAndGrantMembership

	if err := c.ShouldBindJSON(&grant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Checking if the grant type is not nil and if it is not equal to membership.
	// As of now we only have membership grant type.
	if grant.Type == nil || *grant.Type != "membership" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid grant type"})
		return
	}

	ID, dbErr := p.grant.createGrant(c.Request.Context(), grant)

	if dbErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while creating grant: %w", dbErr))
		return
	}

	grant.GrantID = &ID

	if *grant.Type == "membership" {
		_, grantMembErr := p.grant.createGrantMembership(c.Request.Context(), grant)

		if grantMembErr != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("error while creating grant: %w", grantMembErr))
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Created!", "data": ID})
}

func (p *profileManager) handleGrantPatchByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	grantID, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var grant grantAndGrantMembership

	if err := c.ShouldBindJSON(&grant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patchErr := p.grant.patchGrant(c.Request.Context(), grant, grantID)

	if patchErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while creating grant: %w", patchErr))
		return
	}

	grant.GrantID = &grantID

	if grant.UserID != nil || grant.Month != nil || grant.MonthsUsed != nil || grant.MonthsLeft != nil {
		_, grantMembErr := p.grant.patchGrantMembership(c.Request.Context(), grant, grantID)

		if grantMembErr != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("error while creating grant: %w", grantMembErr))
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Updated!", "data": ""})
}
