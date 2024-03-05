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
	GetMultipleGrant(ctx context.Context, intSkip int, intLimit int, boolCancelled *bool, userID string, grantType string, createdAt string) ([]Grant, error)
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
		return Grant{}, fmt.Errorf("error while getting grant: %w", err)
	}

	return grant, nil
}

func (db *ProfileDB) GetMultipleGrant(ctx context.Context, intSkip int, intLimit int, cancelled *bool, userID string, grantType string, createdAt string) ([]Grant, error) {
	grants := []Grant{}

	userDbWhereQuery, orderByQuery := buildAndGetWhereGrantQuery(cancelled, userID, grantType, createdAt)

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
		FROM "grant" `+userDbWhereQuery+orderByQuery+" LIMIT $1 OFFSET $2", intLimit, intSkip)
	if err != nil {
		fmt.Println("--error-while-executing-query", err)
		return []Grant{}, err
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
			return []Grant{}, err
		}

		grants = append(grants, r)
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
		return 0, fmt.Errorf("problem creating grant: %w", err)
	}

	return ID, nil
}

func buildAndGetWhereGrantQuery(cancelled *bool, userID string, grantType string, createdAt string) (string, string) {

	var whereString strings.Builder
	var orderBy strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

	// Add where conditions when required

	if cancelled != nil {
		if whereCondition.String() != "" {
			if *cancelled {
				whereCondition.WriteString(" AND cancelled_at IS NOT NULL")
			} else {
				whereCondition.WriteString(" AND cancelled_at IS NULL")
			}
		} else {
			if *cancelled {
				whereCondition.WriteString(" cancelled_at IS NOT NULL")
			} else {
				whereCondition.WriteString(" cancelled_at IS NULL")
			}
		}
	}

	if grantType != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND type='%s'", grantType))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" type='%s'", grantType))
		}
	}

	if userID != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND user_id='%s'", userID))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" user_id='%s'", userID))
		}
	}

	if createdAt != "" {
		if strings.ToLower(createdAt) != "desc" && strings.ToLower(createdAt) != "asc" {
			createdAt = "asc"
		}
		orderBy.WriteString(fmt.Sprintf(" ORDER BY created_at %s", createdAt))
	} else {
		orderBy.WriteString(fmt.Sprintf(" ORDER BY updated_at %s", "desc"))
	}

	if whereCondition.String() != "" {
		whereString.WriteString(whereCondition.String())
	} else {
		whereString.Reset()
	}
	return whereString.String(), orderBy.String()
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
