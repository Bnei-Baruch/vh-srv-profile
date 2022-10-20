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

type membershipInterface interface {
	getMembershipByID(ctx context.Context, id int) (membershipRes, error)
	getMultipleMembership(ctx context.Context, intSkip int, intLimit int) ([]membershipRes, error)
	patchMembershipByID(ctx context.Context, membership membership, id int) error
	softDeleteMembershipByID(ctx context.Context, id int) error
	cancelMembership(ctx context.Context, body membershipCancellationBody) (error, int, int, int)
}

type membership struct {
	Active *bool      `json:"active"`
	UserID *uuid.UUID `json:"user_id"`
	Type   *string    `json:"type"`
	Month  *int       `json:"month"`
	Year   *int       `json:"year"`
	Expiry *time.Time `json:"expiry"`
}

type membershipCancellationBody struct {
	Email      *string `json:"email"`
	KeycloakID *string `json:"keycloak_id"`
	UserID     *string `json:"user_id"`
}

type membershipRes struct {
	ID *int `json:"id" db:"id"`
	membership
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}

func (p *profileManager) handleMembershipFetchByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	membershipID, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, dbErr := p.membership.getMembershipByID(c.Request.Context(), membershipID)

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

func (p *profileManager) handleMembershipPatchByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	membershipID, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var membership membership

	if err := c.ShouldBindJSON(&membership); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patchErr := p.membership.patchMembershipByID(c.Request.Context(), membership, membershipID)

	if patchErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while updating membership: %w", patchErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Updated!", "data": membership})
}

func (p *profileManager) handleMembershipCancellation(c *gin.Context) {
	var membCancel membershipCancellationBody

	if err := c.ShouldBindJSON(&membCancel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if membCancel.Email == nil && membCancel.KeycloakID == nil && membCancel.UserID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body ( at least one of email, keycloak_id or user_id is required )"})
		return
	}

	if membCancel.KeycloakID != nil {
		_, err := uuid.FromString(*membCancel.KeycloakID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			_ = c.Error(err)
			return
		}
	}

	if membCancel.UserID != nil {
		_, userIDErr := uuid.FromString(*membCancel.UserID)
		if userIDErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": userIDErr.Error()})
			_ = c.Error(userIDErr)
			return
		}
	}

	cancelErr, orderCancelledNum, grantCancelledNum, specialTableDeletedNum := p.membership.cancelMembership(c.Request.Context(), membCancel)

	if cancelErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while updating membership: %w", cancelErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Updated!", "data": gin.H{
		"order_cancelled":       orderCancelledNum,
		"grant_cancelled":       grantCancelledNum,
		"special_table_deleted": specialTableDeletedNum,
	}})
}

func (p *profileManager) handleMembershipSoftDeleteByID(c *gin.Context) {

	id := c.Param("id")

	// convert id to int
	membershipID, err := strconv.Atoi(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var mebership membership

	if err := c.ShouldBindJSON(&mebership); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patchErr := p.membership.softDeleteMembershipByID(c.Request.Context(), membershipID)

	if patchErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while creating membership: %w", patchErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Soft deleted!", "data": ""})
}

func (p *profileManager) handleMembershipFetchAll(c *gin.Context) {

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

	res, err := p.membership.getMultipleMembership(c.Request.Context(), intSkip, intLimit)
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
