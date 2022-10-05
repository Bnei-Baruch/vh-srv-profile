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
	getMultipleGrant(ctx context.Context, intSkip int, intLimit int) ([]grantRes, error)
	createGrant(ctx context.Context, grant grantAndGrantMembership) (int, error)
	patchGrant(ctx context.Context, grant grant, id int) error
	softDeleteGrantByID(ctx context.Context, id int) error
	createGrantMembership(ctx context.Context, grant grantAndGrantMembership) (int, error)
	patchGrantMembership(ctx context.Context, grant grantMembeship, grantId int) (int, error)
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

	var grant grant

	if err := c.ShouldBindJSON(&grant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patchErr := p.grant.patchGrant(c.Request.Context(), grant, grantID)

	if patchErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while updating grant: %w", patchErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Updated!", "data": grant})
}

func (p *profileManager) handleGrantSoftDeleteByID(c *gin.Context) {

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

	patchErr := p.grant.softDeleteGrantByID(c.Request.Context(), grantID)

	if patchErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while creating grant: %w", patchErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Soft deleted!", "data": ""})
}

func (p *profileManager) handleGrantFetchAll(c *gin.Context) {

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

	res, err := p.grant.getMultipleGrant(c.Request.Context(), intSkip, intLimit)
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
