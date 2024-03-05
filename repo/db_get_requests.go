package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

type readMultipleRequestStorage interface {
	GetRequestByID(ctx context.Context, id int) (*NewRequest, error)
	GetMultipleRequest(ctx context.Context,
		intSkip int,
		intLimit int,
		kcid string,
		status string,
		name string,
		typeFilter string,
		orderByCreatedAt string) ([]RequestAndGrant, error)
}

type NewRequest struct {
	ID            *int       `json:"id"`
	RequestName   *string    `json:"name"`
	KeycloakId    *string    `json:"keycloak_id"`
	Status        *string    `json:"status"`
	Type          *string    `json:"type"`
	Months        *int       `json:"nb_month"`
	RequestNote   *string    `json:"request_note,omitempty"`
	RejectionNote *string    `json:"rejection_note,omitempty"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
}

type RequestAndGrant struct {
	Request NewRequest
	Grant   Grant
}

type RequestConclusion struct {
	Approved      bool    `json:"approved"`
	RejectionNote *string `json:"rejection_note"`
	Months        *int    `json:"months"`
}

func (db *ProfileDB) GetRequestByID(ctx context.Context, id int) (*NewRequest, error) {
	var request NewRequest

	err := db.QueryRow(ctx, `SELECT name, keycloak_id, status, type, request_note, rejection_note, created_at, updated_at 
	FROM request WHERE id=$1`, id).
		Scan(&request.RequestName,
			&request.KeycloakId,
			&request.Status,
			&request.Type,
			&request.RequestNote,
			&request.RejectionNote,
			&request.CreatedAt,
			&request.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.ErrNotFound
		}
	}

	request.ID = &id

	return &request, nil
}

func (db *ProfileDB) GetMultipleRequest(ctx context.Context, intSkip int, intLimit int, kcid string, status string, name string, typeFilter string, orderByCreatedAt string) ([]RequestAndGrant, error) {
	requests := []RequestAndGrant{}

	userDbWhereQuery, orderByQuery := buildAndGetWhereRequestQuery(kcid, status, name, typeFilter, orderByCreatedAt)

	rows, err := db.Query(ctx, `
		SELECT 
		r.id, r.name, r.keycloak_id, r.status, r.type, r.request_note, r.rejection_note, r.created_at, r.updated_at,
		g.id, g.user_id, g.request_id, g.type, g.created_at, g.updated_at, g.cancelled_at, g.properties
		FROM request r LEFT JOIN "grant" g ON g.request_id=r.id`+userDbWhereQuery+
		orderByQuery+
		" LIMIT $1 OFFSET $2", intLimit, intSkip)
	if err != nil {
		fmt.Println("--error-while-executing-query", err)
		return []RequestAndGrant{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var r RequestAndGrant
		if err := rows.Scan(&r.Request.ID,
			&r.Request.RequestName,
			&r.Request.KeycloakId,
			&r.Request.Status,
			&r.Request.Type,
			&r.Request.RequestNote,
			&r.Request.RejectionNote,
			&r.Request.CreatedAt,
			&r.Request.UpdatedAt,
			&r.Grant.ID,
			&r.Grant.UserID,
			&r.Grant.RequestID,
			&r.Grant.Type,
			&r.Grant.CreatedAt,
			&r.Grant.UpdatedAt,
			&r.Grant.CancelledAt,
			&r.Grant.Properties,
		); err != nil {
			return []RequestAndGrant{}, err
		}

		requests = append(requests, r)
	}

	return requests, nil
}

func buildAndGetWhereRequestQuery(kcid string, status string, name string, typeFilter string, orderByCreatedAt string) (string, string) {

	var whereString strings.Builder
	var orderBy strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

	// WHERE query generation based on parameters
	if kcid != "" {
		whereCondition.WriteString(fmt.Sprintf(" r.keycloak_id='%s'", kcid))
	}

	if status != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND r.status='%s'", status))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" r.status='%s'", status))
		}
	}

	if name != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND r.name='%s'", name))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" r.name='%s'", name))
		}
	}

	if typeFilter != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND r.type='%s'", typeFilter))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" r.type='%s'", typeFilter))
		}
	}

	if orderByCreatedAt != "" {
		if strings.ToLower(orderByCreatedAt) != "desc" && strings.ToLower(orderByCreatedAt) != "asc" {
			orderByCreatedAt = "asc"
		}
		orderBy.WriteString(fmt.Sprintf(" ORDER BY r.created_at %s", orderByCreatedAt))
	} else {
		orderBy.WriteString(fmt.Sprintf(" ORDER BY r.updated_at %s", "desc"))
	}

	if whereCondition.String() != "" {
		whereString.WriteString(whereCondition.String())
	} else {
		whereString.Reset()
	}
	return whereString.String(), orderBy.String()
}
