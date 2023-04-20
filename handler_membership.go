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
	getMembershipByUserID(ctx context.Context, userID string, authHeader string) (userMembershipRes, error)
	getMultipleMembership(ctx context.Context, intSkip int, intLimit int, month int, year int, userID string) ([]membershipRes, error)
	patchMembershipByID(ctx context.Context, membership membership, id int) (int, error)
	softDeleteMembershipByID(ctx context.Context, id int) error
	cancelMembership(ctx context.Context, body emailKeycloakAndUserIDBody, authHeader string) (int, int, int, error)
	getAutomaticMembershipByMembershipID(ctx context.Context, membershipID int) (membershipAutomatic, error)
	evaluateMembershipByUserID(ctx context.Context, evalbody emailKeycloakAndUserIDBody, authHeader string) (userMembershipRes, error)
}

type membership struct {
	Active *bool      `json:"active"`
	UserID *uuid.UUID `json:"user_id"`
	Type   *string    `json:"type"`
	Month  *int       `json:"month"`
	Year   *int       `json:"year"`
	Expiry *time.Time `json:"expiry"`
}

type membershipAutomatic struct {
	ID           *int       `json:"id"`
	OrderID      *int       `json:"order_id"`
	PaymentID    *int       `json:"payment_id"`
	MembershipID *int       `json:"membership_id"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

type membershipSpecial struct {
	ID           *int       `json:"id"`
	MembershipID *int       `json:"membership_id"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

type membershipManual struct {
	ID           *int       `json:"id"`
	OrderID      *int       `json:"order_id"`
	PaymentID    *int       `json:"payment_id"`
	MembershipID *int       `json:"membership_id"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
	Quantity     *int       `json:"quantity`
}

type membershipHelpHaver struct {
	ID           *int       `json:"id"`
	GrantID      *int       `json:"grant_id"`
	MembershipID *int       `json:"membership_id"`
	NbMonths     *int       `json:"nb_months,omitempty"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

type emailKeycloakAndUserIDBody struct {
	Email      *string `json:"email"`
	KeycloakID *string `json:"keycloak_id"`
	UserID     *string `json:"user_id"`
}

type userNotification struct {
	UserID         *string    `json:"user_id"`
	NotificationID *int       `json:"notification_id"`
	Active         *bool      `json:"active"`
	SeenAt         *time.Time `json:"seen_at"`
}

type userNotificationRes struct {
	ID *int `json:"id"`
	userNotification
	Slug      *string                 `json:"slug"`
	Content   *map[string]interface{} `json:"content"`
	CreatedAt *time.Time              `json:"created_at"`
	UpdatedAt *time.Time              `json:"updated_at"`
	DeletedAt *time.Time              `json:"deleted_at"`
}

type notificationRes struct {
	ID        *int       `json:"id"`
	Slug      *string    `json:"slug"`
	Content   *string    `json:"content"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type userMembershipNotification struct {
	Slug    *string                 `json:"slug"`
	Content *map[string]interface{} `json:"content,omitempty"`
}

type userMembershipRes struct {
	membershipRes
	Notifications []userMembershipNotification `json:"notifications,omitempty"`
	Details       struct {
		Payment struct {
			Date          *time.Time `json:"date,omitempty"`
			Amount        *int       `json:"amount,omitempty"`
			Currency      *string    `json:"currency,omitempty"`
			PaymentMethod *string    `json:"payment_method,omitempty"`
			Status        *string    `json:"status,omitempty"`
		} `json:"payment,omitempty"`
		Automatic struct {
			OrderID   *int `json:"order_id,omitempty"`
			PaymentID *int `json:"payment_id,omitempty"`
		} `json:"automatic"`
		Special struct {
			ApprovedBy *string `json:"approved_by,omitempty"`
			Type       *string `json:"type,omitempty"`
		} `json:"special,omitempty"`
		HelpHaver struct {
			CreatedAt *time.Time `json:"created_at,omitempty"`
			NbMonths  *int       `json:"nb_months,omitempty"`
		} `json:"help_haver,omitempty"`
	} `json:"details,omitempty"`
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

func (p *profileManager) handleMembershipFetchByUserID(c *gin.Context) {

	userID := c.Param("user_id")

	authHeader := c.GetHeader("Authorization")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	res, dbErr := p.membership.getMembershipByUserID(c.Request.Context(), userID, authHeader)

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

	_, patchErr := p.membership.patchMembershipByID(c.Request.Context(), membership, membershipID)

	if patchErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while updating membership: %w", patchErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Updated!", "data": membership})
}

func (p *profileManager) handleMembershipCancellation(c *gin.Context) {
	var membCancel emailKeycloakAndUserIDBody

	if err := c.ShouldBindJSON(&membCancel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if membCancel.Email == nil && membCancel.KeycloakID == nil && membCancel.UserID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body ( at least one of email, keycloak_id or user_id is required )"})
		return
	}

	if membCancel.KeycloakID != nil && *membCancel.KeycloakID != "" {
		_, err := uuid.FromString(*membCancel.KeycloakID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			_ = c.Error(err)
			return
		}
	}

	if membCancel.UserID != nil && *membCancel.UserID != "" {
		_, userIDErr := uuid.FromString(*membCancel.UserID)
		if userIDErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": userIDErr.Error()})
			_ = c.Error(userIDErr)
			return
		}
	}

	authHeader := c.GetHeader("Authorization")

	orderCancelledNum, grantCancelledNum, specialTableDeletedNum, cancelErr := p.membership.cancelMembership(c.Request.Context(), membCancel, authHeader)

	if cancelErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while updating membership: %w", cancelErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Cancelled!", "data": gin.H{
		"order_cancelled": orderCancelledNum,
		"grant_cancelled": grantCancelledNum,
		"special_deleted": specialTableDeletedNum,
	}})
}

func (p *profileManager) handleMembershipEvaluationByUserID(c *gin.Context) {

	var evalbody emailKeycloakAndUserIDBody

	if err := c.ShouldBindJSON(&evalbody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if evalbody.UserID == nil && evalbody.KeycloakID == nil && evalbody.Email == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body ( at least one of email, keycloak_id or user_id is required )"})
		return
	}

	authHeader := c.GetHeader("Authorization")

	userMembershipRes, evaluateErr := p.membership.evaluateMembershipByUserID(c.Request.Context(), evalbody, authHeader)

	if evaluateErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while evaluating membership: %w", evaluateErr))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Evaluated!", "data": userMembershipRes})
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
	month := c.Query("month")
	year := c.Query("year")
	userID := c.Query("user_id")

	var (
		monthInt     int
		yearInt      int
		monthYearErr error
	)

	// month and year to int
	if month != "" {
		monthInt, monthYearErr = strconv.Atoi(month)
		if monthYearErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month"})
			return
		}
	}

	if year != "" {
		yearInt, monthYearErr = strconv.Atoi(year)
		if monthYearErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
			return
		}
	}

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

	res, err := p.membership.getMultipleMembership(c.Request.Context(), intSkip, intLimit, monthInt, yearInt, userID)
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
