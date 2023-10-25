package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type operationInterface interface {
	PerformOperation(ctx context.Context, opr OperationReq) (int, error)
	RevertOperation(ctx context.Context, newEmail string, oldEmail string) error
}

type OperationReq struct {
	ID            *int    `json:"id" form:"id"`
	NewEmail      *string `json:"new_email" form:"new_email" binding:"required"`
	OldEmail      *string `json:"old_email" form:"old_email" binding:"required"`
	NewKeycloakID *string `json:"new_keycloak_id" form:"new_keycloak_id"`
	OldKeycloakID *string `json:"old_keycloak_id" form:"old_keycloak_id"`
	Input         *string `json:"input"`
	Type          *string `json:"type" form:"type"`
	Output        *string `json:"output"`
	Status        *string `json:"status"`
	Revert        *string `json:"revert"`
}

type OperationTrace struct {
	ID     *int    `json:"id"`
	Input  *string `json:"input"`
	Output *string `json:"output"`
	Type   *string `json:"type"`
	Status *string `json:"status"`
	Revert *string `json:"revert"`
}

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

func (db *ProfileDB) PerformOperation(ctx context.Context, req OperationReq) (int, error) {

	newKcId := req.NewKeycloakID
	oldKcId := req.OldKeycloakID
	newEmail := req.NewEmail
	oldEmail := req.OldEmail

	var output QueryLog
	var input emailInput
	var revert QueryLog

	var query string
	var revertQuery string

	tx, err := db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// migration query
	if oldKcId == nil {
		query = `UPDATE users SET primary_email = '` + *newEmail + `' WHERE primary_email = '` + *oldEmail + ` AND keycloak_id = ` + *newKcId + `';`

		revertQuery = `UPDATE users SET primary_email = '` + *oldEmail + `' WHERE primary_email = '` + *newEmail + ` AND keycloak_id = ` + *newKcId + `';`
	} else {
		query = `UPDATE users SET keycloak_id = '` + *newKcId + `', primary_email = '` + *newEmail + `' WHERE keycloak_id = '` + *oldKcId + `';`

		revertQuery = `UPDATE users SET keycloak_id = '` + *oldKcId + `', primary_email = '` + *oldEmail + `' WHERE keycloak_id = '` + *newKcId + `';`
	}

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

func (db *ProfileDB) RevertOperation(ctx context.Context, newEmail string, oldEmail string) error {
	// get operation by id
	var operation OperationTrace

	// get operation by newEmail and oldEmail

	if err := db.QueryRow(ctx, `SELECT id, status, revert FROM operation_trace WHERE input->>'new_email'=$1 AND input->>'old_email'=$2 ORDER BY id DESC LIMIT 1`, newEmail, oldEmail).Scan(
		&operation.ID,
		&operation.Status,
		&operation.Revert); err != nil {
		return fmt.Errorf("problem getting operation_trace: %w", err)
	}

	if *operation.Status == "reverted" {
		return fmt.Errorf("operation already reverted")
	}

	// revert operation
	var revert QueryLog
	if err := json.Unmarshal([]byte(*operation.Revert), &revert); err != nil {
		return fmt.Errorf("problem unmarshalling operation_trace: %w", err)
	}

	tx, err := db.Begin(ctx)
	defer func() { _ = tx.Rollback(ctx) }()

	if err != nil {
		return err
	}

	var query string
	// first query in the Queries array is the query to revert
	if revert.Queries != nil {
		query = revert.Queries[0].(string)
	}

	revertRes, err := tx.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("problem reverting operation: %w", err)
	}

	// update operation_trace
	revert.Logs = append(revert.Logs, revertRes.String())
	revertJson := convertStructToJSONString(revert)
	operation.Revert = &revertJson

	var revertedStr = "reverted"
	operation.Status = &revertedStr

	updateString, updateQueryArgs := prepareOperationTraceUpdateQuery(operation)

	if len(updateQueryArgs) != 0 {
		if err := tx.QueryRow(ctx, fmt.Sprintf(`UPDATE operation_trace SET %s WHERE id='%d' RETURNING id`, updateString, *operation.ID),
			updateQueryArgs...).Scan(&operation.ID); err != nil {
			return fmt.Errorf("problem updating operation_trace: %w", err)
		}

		return tx.Commit(ctx)
	} else {
		return fmt.Errorf("invalid values")
	}

}

func prepareOperationCreateQuery(req OperationReq) (string, string, []interface{}) {
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

func prepareOperationTraceUpdateQuery(req OperationTrace) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.Input != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("input=$%d", len(updateStrings)+1))
		args = append(args, *req.Input)
	}

	if req.Output != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("output=$%d", len(updateStrings)+1))
		args = append(args, *req.Output)
	}

	if req.Revert != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("revert=$%d", len(updateStrings)+1))
		args = append(args, *req.Revert)
	}

	if req.Status != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("status=$%d", len(updateStrings)+1))
		args = append(args, *req.Status)
	}

	if req.Type != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("type=$%d", len(updateStrings)+1))
		args = append(args, *req.Type)
	}

	// if len(args) != 0 {
	// 	updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
	// 	args = append(args, time.Now())
	// }

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}
