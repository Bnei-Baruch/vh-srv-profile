package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	uuid "github.com/satori/go.uuid"
	"github.com/volatiletech/null/v9"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

type grantInterface interface {
	GetGrantByID(ctx context.Context, id int) (Grant, error)
	GetMultipleGrant(ctx context.Context, intSkip int, intLimit int, boolCancelled *bool, userID string, grantType string,
		createdAt string) ([]Grant, error)
}

type Grant struct {
	ID          *int       `json:"id"`
	UserID      *uuid.UUID `json:"user_id"`
	RequestID   *int       `json:"request_id"`
	Type        *string    `json:"type"`
	Properties  null.JSON  `json:"properties"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	CancelledAt *time.Time `json:"cancelled_at"`
}

func (db *ProfileDB) GetGrantByID(ctx context.Context, id int) (Grant, error) {
	var grant Grant

	if err := db.QueryRow(ctx, `SELECT id, user_id, request_id, type, properties, created_at, updated_at, cancelled_at 
	FROM "grant" WHERE id=$1`, id).Scan(
		&grant.ID,
		&grant.UserID,
		&grant.RequestID,
		&grant.Type,
		&grant.Properties,
		&grant.CreatedAt,
		&grant.UpdatedAt,
		&grant.CancelledAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return Grant{}, common.ErrNotFound
		}
		return Grant{}, err
	}

	return grant, nil
}

func (db *ProfileDB) GetMultipleGrant(ctx context.Context, intSkip int, intLimit int, cancelled *bool, userID string, grantType string,
	createdAt string) ([]Grant, error) {
	grants := []Grant{}

	whereQuery, orderByQuery, args := buildAndGetWhereGrantQuery(cancelled, userID, grantType, createdAt)
	args = append(args, intLimit, intSkip)

	rows, err := db.Query(ctx, `
		SELECT
		id,
		user_id,
		request_id,
		type,
		properties ,
		created_at,
		updated_at,
		cancelled_at
		FROM "grant" `+whereQuery+orderByQuery+fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args...)
	if err != nil {
		return []Grant{}, fmt.Errorf("db.Query: %w", err)
	}

	defer rows.Close()
	for rows.Next() {
		var r Grant
		if err := rows.Scan(
			&r.ID,
			&r.UserID,
			&r.RequestID,
			&r.Type,
			&r.Properties,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.CancelledAt,
		); err != nil {
			return []Grant{}, fmt.Errorf("rows.Scan: %w", err)
		}

		grants = append(grants, r)
	}

	if err := rows.Err(); err != nil {
		return grants, fmt.Errorf("rows.Err: %w", err)
	}

	return grants, nil
}

func (db *ProfileDB) createGrant(ctx context.Context, req Grant) (int, error) {
	createString, numString, createQueryArgs := prepareGrantCreateQuery(req)
	if len(createQueryArgs) == 0 {
		return 0, fmt.Errorf("invalid values")
	}

	var ID int
	if err := db.QueryRow(ctx, fmt.Sprintf(`INSERT INTO "grant" (%s) VALUES (%s) RETURNING id`, createString, numString),
		createQueryArgs...).Scan(&ID); err != nil {
		return 0, err
	}

	return ID, nil
}

// buildAndGetWhereGrantQuery builds a parameterized WHERE clause and its bound
// args (never interpolating caller input), plus a safe ORDER BY. Returns the
// WHERE string, the ORDER BY string, and the positional args in placeholder order.
func buildAndGetWhereGrantQuery(cancelled *bool, userID string, grantType string, createdAt string) (string, string, []interface{}) {
	var conditions []string
	var args []interface{}

	// eq appends "col=$N" bound to the next positional placeholder.
	eq := func(col, val string) {
		args = append(args, val)
		conditions = append(conditions, fmt.Sprintf("%s=$%d", col, len(args)))
	}

	// cancelled is a typed bool, not caller text — its SQL is a fixed literal.
	if cancelled != nil {
		if *cancelled {
			conditions = append(conditions, "cancelled_at IS NOT NULL")
		} else {
			conditions = append(conditions, "cancelled_at IS NULL")
		}
	}
	if grantType != "" {
		eq("type", grantType)
	}
	if userID != "" {
		eq("user_id", userID)
	}

	var whereString string
	if len(conditions) > 0 {
		whereString = " WHERE " + strings.Join(conditions, " AND ")
	}

	// createdAt is reduced to a fixed asc/desc literal — never interpolated user data.
	orderColumn, orderDir := "updated_at", "desc"
	if createdAt != "" {
		orderColumn = "created_at"
		if strings.ToLower(createdAt) != "desc" {
			orderDir = "asc"
		}
	}
	orderBy := fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDir)

	return whereString, orderBy, args
}

func prepareGrantCreateQuery(req Grant) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.UserID != nil {
		createStrings = append(createStrings, "user_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.UserID)
	}
	if req.Type != nil {
		createStrings = append(createStrings, "type")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Type)
	}
	if req.RequestID != nil {
		createStrings = append(createStrings, "request_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RequestID)
	}
	if req.Properties.Valid {
		createStrings = append(createStrings, "properties")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, req.Properties)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}
