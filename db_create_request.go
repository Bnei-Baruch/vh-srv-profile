package main

import (
	"context"
	"fmt"
	"strings"
)

func (db *pgProfileDB) createRequest(ctx context.Context, req newRequest) error {

	createString, numString, createQueryArgs := prepareRequestCreateQuery(req)

	if len(createQueryArgs) != 0 {
		_, err := db.Exec(ctx, fmt.Sprintf(`INSERT INTO request (%s) VALUES (%s)`, createString, numString),
			createQueryArgs...)
		if err != nil {
			return fmt.Errorf("problem creating request: %w", err)
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func prepareRequestCreateQuery(req newRequest) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.KeycloakId != nil {
		createStrings = append(createStrings, "keycloak_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.KeycloakId)
	}
	if req.RejectionNote != nil {
		createStrings = append(createStrings, "rejection_note")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RejectionNote)
	}
	if req.RequestName != nil {
		createStrings = append(createStrings, "request_name")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RequestName)
	}
	if req.RequestNote != nil {
		createStrings = append(createStrings, "request_note")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RequestNote)
	}
	if req.Status != nil {
		createStrings = append(createStrings, "status")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Status)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}
