package repo

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type readMultipleRequestStorage interface {
	GetMultipleRequest(ctx context.Context,
		intSkip int,
		intLimit int,
		kcid string,
		status string,
		name string,
		typeFilter string,
		orderByCreatedAt string) ([]RequestResponse, error)
}

type NewRequest struct {
	Grant
	GrantMembership
	RequestName   *string `json:"name"`
	KeycloakId    *string `json:"keycloak_id"`
	Status        *string `json:"status"`
	EventSlug     *string `json:"event_slug"`
	Type          *string `json:"type"`
	Months        *int    `json:"nb_month"`
	RequestNote   *string `json:"request_note,omitempty"`
	RejectionNote *string `json:"rejection_note,omitempty"`
}

type RequestResponse struct {
	ID            *int       `json:"id"`
	RequestName   *string    `json:"name" db:"name"`
	KeycloakID    *string    `json:"keycloak_id" db:"keycloak_id"`
	Status        *string    `json:"status" db:"status"`
	EventSlug     *string    `json:"event_slug" db:"event_slug"`
	Type          *string    `json:"type" db:"type"`
	RequestNote   *string    `json:"request_note" db:"request_note"`
	RejectionNote *string    `json:"rejection_note" db:"rejection_note"`
	CreatedAt     *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at" db:"updated_at"`
}

func (db *ProfileDB) GetMultipleRequest(ctx context.Context, intSkip int, intLimit int, kcid string, status string, name string, typeFilter string, orderByCreatedAt string) ([]RequestResponse, error) {
	requests := []RequestResponse{}

	userDbWhereQuery, orderByQuery := buildAndGetWhereRequestQuery(kcid, status, name, typeFilter, orderByCreatedAt)

	rows, err := db.Query(ctx, `
		SELECT 
		id,
		name,
		keycloak_id,
		status,
		event_slug,
		type,
		request_note,
		rejection_note,
		created_at,
		updated_at 
		FROM request `+userDbWhereQuery+
		orderByQuery+
		" LIMIT $1 OFFSET $2", intLimit, intSkip)
	if err != nil {
		fmt.Println("--error-while-executing-query", err)
		return []RequestResponse{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var r RequestResponse
		if err := rows.Scan(&r.ID, &r.RequestName, &r.KeycloakID, &r.Status, &r.EventSlug, &r.Type, &r.RequestNote, &r.RejectionNote, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return []RequestResponse{}, err
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
		whereCondition.WriteString(fmt.Sprintf(" keycloak_id='%s'", kcid))
	}

	if status != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND status='%s'", status))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" status='%s'", status))
		}
	}

	if name != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND name='%s'", name))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" name='%s'", name))
		}
	}

	if typeFilter != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND type='%s'", typeFilter))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" type='%s'", typeFilter))
		}
	}

	if orderByCreatedAt != "" {
		if strings.ToLower(orderByCreatedAt) != "desc" && strings.ToLower(orderByCreatedAt) != "asc" {
			orderByCreatedAt = "asc"
		}
		orderBy.WriteString(fmt.Sprintf(" ORDER BY created_at %s", orderByCreatedAt))
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
