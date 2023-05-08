package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type QueryLog struct {
	Queries []interface{} `json:"queries"`
	Logs    []interface{} `json:"logs"`
}

type emailInput struct {
	NewEmail      *string `json:"new_email"`
	NewKeycloakID *string `json:"new_keycloak_id"`
	OldKeycloakID *string `json:"old_keycloak_id"`
	OldEmail      *string `json:"old_email"`
}

func convertStructToJSONString(input interface{}) string {
	jsonString, err := json.Marshal(input)
	if err != nil {
		return ""
	}
	return string(jsonString)
}

func (db *pgProfileDB) performOperation(ctx context.Context, req operationReq) (int, error) {

	newKcId := req.NewKeycloakID
	oldKcId := req.OldKeycloakID
	newEmail := req.NewEmail
	oldEmail := req.OldEmail

	var output QueryLog
	var input emailInput
	var revert QueryLog

	tx, err := db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query := `UPDATE users SET keycloak_id = '` + *newKcId + `', primary_email = '` + *newEmail + `' WHERE keycloak_id = '` + *oldKcId + `';`

	updatedRes, err := tx.Exec(ctx, query)

	if err != nil {
		return 0, fmt.Errorf("problem updating users: %w", err)
	}

	output.Queries = append(output.Queries, query)
	output.Logs = append(output.Logs, updatedRes.String())

	input.NewEmail = newEmail
	input.NewKeycloakID = newKcId
	input.OldKeycloakID = oldKcId
	input.OldEmail = req.OldEmail

	revertQuery := `UPDATE users SET keycloak_id = '` + *oldKcId + `', primary_email = '` + *oldEmail + `' WHERE keycloak_id = '` + *newKcId + `';`

	revert.Queries = append(revert.Queries, revertQuery)

	revert.Logs = []interface{}{}

	var ID int

	inputJson := convertStructToJSONString(input)
	req.Input = &inputJson

	outputJson := convertStructToJSONString(output)
	req.Output = &outputJson

	revertJson := convertStructToJSONString(revert)
	req.Revert = &revertJson

	success := "success"
	req.Status = &success

	emailUpdate := "email_update"
	req.Type = &emailUpdate

	createString, numString, createQueryArgs := prepareOperationCreateQuery(req)

	if len(createQueryArgs) != 0 {
		if err := tx.QueryRow(ctx, fmt.Sprintf(`INSERT INTO operation_trace (%s) VALUES (%s) RETURNING id`, createString, numString),
			createQueryArgs...).Scan(&ID); err != nil {
			return 0, fmt.Errorf("problem creating operation_trace: %w", err)
		}

		return ID, tx.Commit(ctx)
	} else {
		return 0, fmt.Errorf("invalid values")
	}
}

// BEGIN;

// UPDATE users
// SET keycloak_id = 'new_kc_id', primary_email = 'new_email'
// WHERE keycloak_id = 'old_kc_id';

// COMMIT;

func prepareOperationCreateQuery(req operationReq) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.Input != nil {
		createStrings = append(createStrings, "input")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Input)
	}

	if req.Output != nil {
		createStrings = append(createStrings, "output")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Output)
	}

	if req.Revert != nil {
		createStrings = append(createStrings, "revert")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Revert)
	}

	if req.Status != nil {
		createStrings = append(createStrings, "status")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Status)
	}

	if req.Type != nil {
		createStrings = append(createStrings, "type")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Type)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}
