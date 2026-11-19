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

// sqlQuery is one parameterized statement. operation_trace keeps them as a
// record of what ran; a revert never runs SQL read back from the table.
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

// renameStatements moves fromEmail to toEmail. With fromKc set, the user
// also moves from fromKc to toKc; without it, only the email of toKc's user
// changes. A revert is the same call with the sides swapped.
func renameStatements(fromEmail, toEmail string, fromKc *string, toKc string) []sqlQuery {
	if fromKc == nil {
		return []sqlQuery{{`UPDATE users SET primary_email = $1 WHERE primary_email = $2 AND keycloak_id = $3`,
			[]string{toEmail, fromEmail, toKc}}}
	}
	return []sqlQuery{{`UPDATE users SET keycloak_id = $1, primary_email = $2 WHERE keycloak_id = $3`,
		[]string{toKc, toEmail, *fromKc}}}
}

// revertStatements undoes renameStatements for the same input.
func revertStatements(in emailInput) []sqlQuery {
	if in.OldKeycloakID == nil {
		return renameStatements(*in.NewEmail, *in.OldEmail, nil, *in.NewKeycloakID)
	}
	return renameStatements(*in.NewEmail, *in.OldEmail, in.NewKeycloakID, *in.OldKeycloakID)
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

	input := emailInput{
		NewEmail:      req.NewEmail,
		NewKeycloakID: req.NewKeycloakID,
		OldKeycloakID: req.OldKeycloakID,
		OldEmail:      req.OldEmail,
	}
	var output QueryLog
	revert := QueryLog{Queries: revertStatements(input), Logs: []string{}}

	tx, err := db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, query := range renameStatements(*input.OldEmail, *input.NewEmail, input.OldKeycloakID, *input.NewKeycloakID) {
		updatedRes, err := tx.Exec(ctx, query.SQL, query.args()...)
		if err != nil {
			return 0, fmt.Errorf("problem updating users: %w", err)
		}
		output.Queries = append(output.Queries, query)
		output.Logs = append(output.Logs, updatedRes.String())
	}

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
		input  emailInput
	)
	if err := tx.QueryRow(ctx, `SELECT id, status, input FROM operation_trace
		WHERE input->>'new_email'=$1 AND input->>'old_email'=$2
		ORDER BY id DESC LIMIT 1 FOR UPDATE`, newEmail, oldEmail).Scan(&id, &status, &input); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return common.ErrOperationNotFound
		}
		return fmt.Errorf("problem getting operation_trace: %w", err)
	}

	if status == "reverted" {
		return fmt.Errorf("operation already reverted")
	}

	// The statements come from the code, built from the stored input.
	revert := QueryLog{Queries: revertStatements(input), Logs: []string{}}
	for _, query := range revert.Queries {
		revertRes, err := tx.Exec(ctx, query.SQL, query.args()...)
		if err != nil {
			return fmt.Errorf("problem reverting operation: %w", err)
		}
		revert.Logs = append(revert.Logs, revertRes.String())
	}

	if _, err := tx.Exec(ctx, `UPDATE operation_trace SET revert = $1, status = 'reverted' WHERE id = $2`,
		revert, id); err != nil {
		return fmt.Errorf("problem updating operation_trace: %w", err)
	}

	return tx.Commit(ctx)
}
