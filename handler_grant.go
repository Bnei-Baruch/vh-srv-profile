package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type grantInterface interface {
	getGrantByID(ctx context.Context, id int) (grant, error)
}

type grant struct {
	ID        *int       `json:"id" db:"id"`
	Amount    *int       `json:"amount" db:"amount"`
	Currency  *string    `json:"currency" db:"currency"`
	Type      *string    `json:"type" db:"type"`
	Loaned    *int       `json:"loaned" db:"loaned"`
	Granted   *int       `json:"granted" db:"granted"`
	Repayed   *int       `json:"repayed" db:"repayed"`
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

	res, dbErr := p.fetchGrantByID.getGrantByID(c.Request.Context(), grantID)

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
