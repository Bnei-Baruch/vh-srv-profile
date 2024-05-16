package repo

import (
	"context"
	"fmt"
	uuid "github.com/satori/go.uuid"
)

type mergeAccounts interface {
	MergeAccounts(ctx context.Context, data AccountsMergeData) (*EmailKeycloakAndUserIDBody, error)
}

type AccountsMergeData struct {
	SourceId         string
	DestinationId    string
	SourceEmail      string
	DestinationEmail string
}

func (db *ProfileDB) MergeAccounts(ctx context.Context, data AccountsMergeData) (*EmailKeycloakAndUserIDBody, error) {
	var (
		sourceUserID      string
		destinationUserID string
	)
	if err := db.QueryRow(ctx, `SELECT user_id from users WHERE keycloak_id = $1`, data.SourceId).Scan(&sourceUserID); err != nil {
		return nil, fmt.Errorf("db.QueryRow [user_id from keycloak_id(%s)]: %w", data.SourceId, err)
	}
	if err := db.QueryRow(ctx, `SELECT user_id from users WHERE keycloak_id = $1`, data.DestinationId).Scan(&destinationUserID); err != nil {
		return nil, fmt.Errorf("db.QueryRow [user_id from keycloak_id(%s)]: %w", data.DestinationId, err)
	}

	// start transaction
	tx, txErr := db.Begin(ctx)
	if txErr != nil {
		return nil, fmt.Errorf("db.Begin: %w", txErr)
	}
	defer func() error {
		if err := tx.Rollback(ctx); err != nil {
			return fmt.Errorf("tx.Rollback: %w", txErr)
		}
		return nil
	}()

	_, err := tx.Exec(ctx, `UPDATE users set primary_email =$1  WHERE keycloak_id=$2`, data.DestinationEmail, data.SourceId)
	if err != nil {
		return nil, fmt.Errorf("tx.Exec [manual]: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE status set user_id =$1  WHERE user_id=$2`, destinationUserID, sourceUserID)
	if err != nil {
		return nil, fmt.Errorf("tx.Exec [manual]: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE "grant" set user_id = $1 where user_id = $2 `, destinationUserID, sourceUserID)
	if err != nil {
		return nil, fmt.Errorf("tx.Exec [special]: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE phone_numbers set user_id = $1 where user_id = $2 `, destinationUserID, sourceUserID)
	if err != nil {
		return nil, fmt.Errorf("tx.Exec [automatic]: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE user_notification set user_id = $1 where user_id = $2 `, destinationUserID, sourceUserID)
	if err != nil {
		return nil, fmt.Errorf("tx.Exec [automatic]: %w", err)
	}

	// commit transaction
	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, fmt.Errorf("tx.Commit: %w", commitErr)
	}

	if err := db.HardDeleteProfile(ctx, uuid.FromStringOrNil(sourceUserID)); err != nil {
		return nil, fmt.Errorf("db.HardDeleteProfile: %w", err)
	}

	return &EmailKeycloakAndUserIDBody{Email: &data.DestinationEmail, KeycloakID: &data.DestinationId, UserID: &destinationUserID}, nil
}
