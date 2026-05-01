package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/volatiletech/null/v9"
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
		email string,
		typeFilters []string,
		orderByCreatedAt string) ([]RequestAndGrant, error)
	GetMultipleRequestCount(ctx context.Context,
		kcid string,
		status string,
		name string,
		email string,
		typeFilters []string,
		orderByCreatedAt string) (int, error)
}

type NewRequest struct {
	ID            *int       `json:"id"`
	RequestName   *string    `json:"name"`
	KeycloakId    *string    `json:"keycloak_id"`
	Status        *string    `json:"status"`
	Type          *string    `json:"type"`
	Properties    null.JSON  `json:"properties"`
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
	Approved      bool      `json:"approved"`
	RejectionNote *string   `json:"rejection_note"`
	Type          *string   `json:"type"`
	Properties    null.JSON `json:"properties"`
}

func (db *ProfileDB) GetRequestByID(ctx context.Context, id int) (*NewRequest, error) {
	var request NewRequest

	err := db.QueryRow(ctx, `SELECT name, keycloak_id, status, type, properties, request_note, rejection_note, created_at, updated_at
	FROM request WHERE id=$1`, id).
		Scan(&request.RequestName,
			&request.KeycloakId,
			&request.Status,
			&request.Type,
			&request.Properties,
			&request.RequestNote,
			&request.RejectionNote,
			&request.CreatedAt,
			&request.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	request.ID = &id

	return &request, nil
}

func (db *ProfileDB) GetMultipleRequest(ctx context.Context, intSkip int, intLimit int, kcid string, status string, name string, email string, typeFilters []string, orderByCreatedAt string) ([]RequestAndGrant, error) {
	requests := []RequestAndGrant{}

	userDbWhereQuery, orderByQuery := buildAndGetWhereRequestQuery(kcid, status, name, typeFilters, orderByCreatedAt, email)
	queryString := `SELECT r.id, r.name, r.keycloak_id, r.status, r.type, r.properties, r.request_note, r.rejection_note, r.created_at, r.updated_at, g.id, g.user_id, g.request_id, g.type, g.created_at, g.updated_at, g.cancelled_at, g.properties FROM request r LEFT JOIN "grant" g ON g.request_id=r.id`
	if email != "" {
		queryString += ` LEFT JOIN "users" u ON u.keycloak_id=r.keycloak_id `

	}
	finalQueryString := fmt.Sprintf("%s %s %s  LIMIT $1 OFFSET $2", queryString, userDbWhereQuery, orderByQuery)
	rows, err := db.Query(ctx, finalQueryString, intLimit, intSkip)
	if err != nil {
		return []RequestAndGrant{}, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var r RequestAndGrant
		if err := rows.Scan(&r.Request.ID,
			&r.Request.RequestName,
			&r.Request.KeycloakId,
			&r.Request.Status,
			&r.Request.Type,
			&r.Request.Properties,
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
			return []RequestAndGrant{}, fmt.Errorf("rows.Scan: %w", err)
		}

		requests = append(requests, r)
	}

	if err := rows.Err(); err != nil {
		return []RequestAndGrant{}, fmt.Errorf("rows.Err: %w", err)
	}

	return requests, nil
}

func (db *ProfileDB) GetMultipleRequestCount(ctx context.Context, kcid string, status string, name string, email string, typeFilters []string, orderByCreatedAt string) (int, error) {

	userDbWhereQuery, _ := buildAndGetWhereRequestQuery(kcid, status, name, typeFilters, orderByCreatedAt, email)
	queryString := `SELECT COUNT(*) FROM request r`
	queryStringJoin := ""
	if email != "" {
		queryStringJoin = ` LEFT JOIN "users" u ON u.keycloak_id=r.keycloak_id `
	}
	queryStringFinal := fmt.Sprintf("%s %s %s", queryString, queryStringJoin, userDbWhereQuery)
	var count int
	err := db.QueryRow(ctx, queryStringFinal).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("db.QueryRow get count: %w", err)
	}

	return count, nil
}

func buildAndGetWhereRequestQuery(kcid string, status string, name string, typeFilters []string, orderByCreatedAt string, email string) (string, string) {

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
			whereCondition.WriteString(fmt.Sprintf(" AND lower(r.name) like lower('%%%s%%')", name))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" lower(r.name) like lower('%%%s%%')", name))
		}
	}

	if len(typeFilters) > 0 {
		quoted := make([]string, len(typeFilters))
		for i, t := range typeFilters {
			quoted[i] = "'" + t + "'"
		}
		typeCondition := fmt.Sprintf("r.type IN (%s)", strings.Join(quoted, ","))
		if whereCondition.String() != "" {
			whereCondition.WriteString(" AND " + typeCondition)
		} else {
			whereCondition.WriteString(typeCondition)
		}
	}

	if email != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND (lower(u.primary_email) like lower('%%%[1]s%%') OR lower(u.alternate_email_1) like lower('%%%[1]s%%') OR lower(u.alternate_email_2) like lower('%%%[1]s%%')) ", email))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" (lower(u.primary_email) like lower('%%%[1]s%%') OR lower(u.alternate_email_1) like lower('%%%[1]s%%') OR lower(u.alternate_email_2) like lower('%%%[1]s%%')) ", email))
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
