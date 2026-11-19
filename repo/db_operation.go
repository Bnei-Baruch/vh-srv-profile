package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

type operationInterface interface {
	PerformOperation(ctx context.Context, opr OperationReq) (int, error)
	RevertOperation(ctx context.Context, newEmail string, oldEmail string) error
}

type OperationReq struct {
	NewEmail      *string `json:"new_email" form:"new_email" binding:"required"`
	OldEmail      *string `json:"old_email" form:"old_email" binding:"required"`
	NewKeycloakID *string `json:"new_keycloak_id" form:"new_keycloak_id"`
	OldKeycloakID *string `json:"old_keycloak_id" form:"old_keycloak_id"`
	Type          *string `json:"type" form:"type"`
}

// sqlQuery is one parameterized statement, kept in operation_trace so a
// revert can run it later with the same arguments.
type sqlQuery struct {
	SQL  string   `json:"sql"`
	Args []string `json:"args"`
}

func (q sqlQuery) args() []any {
	args := make([]any, len(q.Args))
	for i, a := range q.Args {
		args[i] = a
	}
	return args
}

type QueryLog struct {
	Queries []sqlQuery `json:"queries"`
	Logs    []string   `json:"logs"`
}

type emailInput struct {
	NewEmail      *string `json:"new_email"`
	NewKeycloakID *string `json:"new_keycloak_id"`
	OldKeycloakID *string `json:"old_keycloak_id"`
	OldEmail      *string `json:"old_email"`
}

func (db *ProfileDB) PerformOperation(ctx context.Context, req OperationReq) (int, error) {

	newKcId := req.NewKeycloakID
	oldKcId := req.OldKeycloakID
	newEmail := req.NewEmail
	oldEmail := req.OldEmail

	var output QueryLog
	var input emailInput
	var revert QueryLog

	var query sqlQuery

	tx, err := db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// migration query and the statement that undoes it
	if oldKcId == nil {
		const sql = `UPDATE users SET primary_email = $1 WHERE primary_email = $2 AND keycloak_id = $3`
		query = sqlQuery{SQL: sql, Args: []string{*newEmail, *oldEmail, *newKcId}}
		revert.Queries = append(revert.Queries, sqlQuery{SQL: sql, Args: []string{*oldEmail, *newEmail, *newKcId}})
	} else {
		const sql = `UPDATE users SET keycloak_id = $1, primary_email = $2 WHERE keycloak_id = $3`
		query = sqlQuery{SQL: sql, Args: []string{*newKcId, *newEmail, *oldKcId}}
		revert.Queries = append(revert.Queries, sqlQuery{SQL: sql, Args: []string{*oldKcId, *oldEmail, *newKcId}})
	}

	updatedRes, err := tx.Exec(ctx, query.SQL, query.args()...)
	if err != nil {
		return 0, fmt.Errorf("problem updating users: %w", err)
	}

	output.Queries = append(output.Queries, query)
	output.Logs = append(output.Logs, updatedRes.String())

	input.NewEmail = newEmail
	input.NewKeycloakID = newKcId
	input.OldKeycloakID = oldKcId
	input.OldEmail = oldEmail

	revert.Logs = []string{}

	// pgx encodes the structs as JSON for the json columns.
	var ID int
	if err := tx.QueryRow(ctx, `INSERT INTO operation_trace (input, output, revert, status, type)
		VALUES ($1, $2, $3, 'success', 'email_update') RETURNING id`,
		input, output, revert).Scan(&ID); err != nil {
		return 0, fmt.Errorf("problem creating operation_trace: %w", err)
	}

	return ID, tx.Commit(ctx)
}

// RevertOperation undoes the latest email_update for this email pair. The
// row is locked, so two reverts at once cannot both run.
func (db *ProfileDB) RevertOperation(ctx context.Context, newEmail string, oldEmail string) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("db.Begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		id     int
		status string
		revert QueryLog
	)
	if err := tx.QueryRow(ctx, `SELECT id, status, revert FROM operation_trace
		WHERE input->>'new_email'=$1 AND input->>'old_email'=$2
		ORDER BY id DESC LIMIT 1 FOR UPDATE`, newEmail, oldEmail).Scan(&id, &status, &revert); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return common.ErrOperationNotFound
		}
		return fmt.Errorf("problem getting operation_trace: %w", err)
	}

	if status == "reverted" {
		return fmt.Errorf("operation already reverted")
	}

	// first query in the Queries array is the query to revert
	if len(revert.Queries) == 0 {
		return fmt.Errorf("operation %d has no revert query", id)
	}
	query := revert.Queries[0]

	revertRes, err := tx.Exec(ctx, query.SQL, query.args()...)
	if err != nil {
		return fmt.Errorf("problem reverting operation: %w", err)
	}
	revert.Logs = append(revert.Logs, revertRes.String())

	if _, err := tx.Exec(ctx, `UPDATE operation_trace SET revert = $1, status = 'reverted' WHERE id = $2`,
		revert, id); err != nil {
		return fmt.Errorf("problem updating operation_trace: %w", err)
	}

	return tx.Commit(ctx)
}
