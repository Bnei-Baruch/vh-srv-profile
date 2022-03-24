package main

import (
	"context"
	"fmt"
	"strings"
)

func (db *pgProfileDB) getMultipleRequest(ctx context.Context, intSkip int, intLimit int, kcid string, status string, name string) ([]requestResponse, error) {
	requests := []requestResponse{}

	userDbWhereQuery, orderByQuery := buildAndGetWhereRequestQuery(kcid, status, name)

	rows, err := db.Query(ctx, `
		SELECT 
		request_name,
		keycloak_id,
		status,
		request_note,
		rejection_note,
		created_at,
		updated_at 
		FROM request `+userDbWhereQuery+
		orderByQuery+
		" LIMIT $1 OFFSET $2", intLimit, intSkip)
	if err != nil {
		fmt.Println("--error-while-executing-query", err)
		return []requestResponse{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var r requestResponse
		if err := rows.Scan(&r.RequestName, r.KeycloakID, r.Status, r.RequestNote, r.RejectionNote, r.CreatedAt, r.UpdatedAt); err != nil {
			return []requestResponse{}, err
		}

		requests = append(requests, r)
	}

	// Manage if no user found
	if len(requests) == 0 {
		return []requestResponse{}, fmt.Errorf("%w", errUserNotFound)
	}

	return requests, nil
}

func buildAndGetWhereRequestQuery(kcid string, status string, name string) (string, string) {

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
			whereCondition.WriteString(fmt.Sprintf(" AND request_name='%s'", name))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" request_name='%s'", name))
		}
	}

	orderBy.WriteString(fmt.Sprintf(" ORDER BY updated_at %s", "desc"))

	if whereCondition.String() != "" {
		whereString.WriteString(whereCondition.String())
	} else {
		whereString.Reset()
	}
	return whereString.String(), orderBy.String()
}
