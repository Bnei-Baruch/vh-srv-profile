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
		email string,
		typeFilter string,
		orderByCreatedAt string) ([]RequestAndGrant, error)
	GetMultipleRequestCount(ctx context.Context,
		kcid string,
		status string,
		name string,
		email string,
		typeFilter string,
		orderByCreatedAt string) (int, error)
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
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	request.ID = &id

	return &request, nil
}

func (db *ProfileDB) GetMultipleRequest(ctx context.Context, intSkip int, intLimit int, kcid string, status string, name string, email string, typeFilter string, orderByCreatedAt string) ([]RequestAndGrant, error) {
	requests := []RequestAndGrant{}

	whereQuery, orderByQuery, args := buildAndGetWhereRequestQuery(kcid, status, name, typeFilter, orderByCreatedAt, email)
	queryString := `SELECT r.id, r.name, r.keycloak_id, r.status, r.type, r.request_note, r.rejection_note, r.created_at, r.updated_at, r.months, g.id, g.user_id, g.request_id, g.type, g.created_at, g.updated_at, g.cancelled_at, g.properties
		FROM request r LEFT JOIN "grant" g ON g.request_id=r.id`
	if email != "" {
		queryString += ` LEFT JOIN "users" u ON u.keycloak_id=r.keycloak_id `
	}
	args = append(args, intLimit, intSkip)
	finalQueryString := fmt.Sprintf("%s %s %s LIMIT $%d OFFSET $%d", queryString, whereQuery, orderByQuery, len(args)-1, len(args))
	rows, err := db.Query(ctx, finalQueryString, args...)
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
			&r.Request.RequestNote,
			&r.Request.RejectionNote,
			&r.Request.CreatedAt,
			&r.Request.UpdatedAt,
			&r.Request.Months,
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

func (db *ProfileDB) GetMultipleRequestCount(ctx context.Context, kcid string, status string, name string, email string, typeFilter string, orderByCreatedAt string) (int, error) {

	whereQuery, _, args := buildAndGetWhereRequestQuery(kcid, status, name, typeFilter, orderByCreatedAt, email)
	queryString := `SELECT COUNT(*) FROM request r`
	queryStringJoin := ""
	if email != "" {
		queryStringJoin = ` LEFT JOIN "users" u ON u.keycloak_id=r.keycloak_id `
	}
	queryStringFinal := fmt.Sprintf("%s %s %s", queryString, queryStringJoin, whereQuery)
	var count int
	err := db.QueryRow(ctx, queryStringFinal, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("db.QueryRow get count: %w", err)
	}

	return count, nil
}

// buildAndGetWhereRequestQuery builds a parameterized WHERE clause and its bound
// args (never interpolating caller input), plus a safe ORDER BY. It returns the
// WHERE string, the ORDER BY string, and the positional args in placeholder order.
func buildAndGetWhereRequestQuery(kcid string, status string, name string, typeFilter string, orderByCreatedAt string, email string) (string, string, []interface{}) {
	var conditions []string
	var args []interface{}

	// eq appends "col=$N" bound to the next positional placeholder.
	eq := func(col, val string) {
		args = append(args, val)
		conditions = append(conditions, fmt.Sprintf("%s=$%d", col, len(args)))
	}

	if kcid != "" {
		eq("r.keycloak_id", kcid)
	}
	if status != "" {
		eq("r.status", status)
	}
	if name != "" {
		args = append(args, "%"+name+"%")
		conditions = append(conditions, fmt.Sprintf("lower(r.name) like lower($%d)", len(args)))
	}
	if typeFilter != "" {
		eq("r.type", typeFilter)
	}
	if email != "" {
		args = append(args, "%"+email+"%")
		n := len(args)
		conditions = append(conditions, fmt.Sprintf(
			"(lower(u.primary_email) like lower($%[1]d) OR lower(u.alternate_email_1) like lower($%[1]d) OR lower(u.alternate_email_2) like lower($%[1]d))", n))
	}

	var whereString string
	if len(conditions) > 0 {
		whereString = " WHERE " + strings.Join(conditions, " AND ")
	}

	// orderByCreatedAt is reduced to a fixed asc/desc literal — never interpolated user data.
	orderColumn, orderDir := "r.updated_at", "desc"
	if orderByCreatedAt != "" {
		orderColumn = "r.created_at"
		if strings.ToLower(orderByCreatedAt) == "desc" {
			orderDir = "desc"
		} else {
			orderDir = "asc"
		}
	}
	orderBy := fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDir)

	return whereString, orderBy, args
}
