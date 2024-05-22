package repo

import (
	"context"
	"database/sql"
	"fmt"
	uuid "github.com/satori/go.uuid"
)

type mergeAccounts interface {
	MergeAccounts(ctx context.Context, data AccountsMergeData) (*EmailKeycloakAndUserIDBody, error)
}

type AccountsMergeData struct {
	SourceId      string
	DestinationId string
}

func (db *ProfileDB) MergeAccounts(ctx context.Context, data AccountsMergeData) (*EmailKeycloakAndUserIDBody, error) {
	var (
		sourceUserID      string
		destinationUserID string
		sourceEmail       string
		destinationEmail  string
		err               error
	)
	if err = db.QueryRow(ctx, `SELECT user_id from users WHERE keycloak_id = $1`, data.SourceId).Scan(&sourceUserID); err != nil {
		return nil, fmt.Errorf("db.QueryRow [Source user_id from keycloak_id(%s)]: %w", data.SourceId, err)
	}
	if err = db.QueryRow(ctx, `SELECT user_id from users WHERE keycloak_id = $1`, data.DestinationId).Scan(&destinationUserID); err != nil {
		return nil, fmt.Errorf("db.QueryRow [Destination user_id from keycloak_id(%s)]: %w", data.DestinationId, err)
	}

	if err = db.QueryRow(ctx, `SELECT primary_email from users WHERE keycloak_id = $1`, data.SourceId).Scan(&sourceEmail); err != nil {
		return nil, fmt.Errorf("db.QueryRow [SourceId primary_email from keycloak_id(%s)]: %w", data.SourceId, err)
	}
	if err = db.QueryRow(ctx, `SELECT primary_email from users WHERE keycloak_id = $1`, data.DestinationId).Scan(&destinationEmail); err != nil {
		return nil, fmt.Errorf("db.QueryRow [Destination primary_email from keycloak_id(%s)]: %w", data.DestinationId, err)
	}

	needUpdateEmail := false
	var destinationAlternativeEmail sql.NullString
	if sourceEmail != destinationEmail {
		if err := db.QueryRow(ctx, `SELECT alternate_email_1 from users WHERE keycloak_id = $1`, data.DestinationId).Scan(&destinationAlternativeEmail); err != nil {
			return nil, fmt.Errorf("db.QueryRow [alternate_email_1 from keycloak_id(%s)]: %w", data.DestinationId, err)
		}
		needUpdateEmail = true
	}

	// start transaction
	tx, txErr := db.Begin(ctx)
	if txErr != nil {
		return nil, fmt.Errorf("db.Begin: %w", txErr)
	}
	defer func() error {
		if err = tx.Rollback(ctx); err != nil {
			return fmt.Errorf("tx.Rollback: %w", txErr)
		}
		return nil
	}()

	if needUpdateEmail && destinationAlternativeEmail.Valid {
		_, err = tx.Exec(ctx, `UPDATE users set alternate_email_1=$1  WHERE keycloak_id=$2`, sourceEmail, data.DestinationId)
		if err != nil {
			return nil, fmt.Errorf("tx.Exec [UPDATE users set alternate_email_1]: %w", err)
		}
	}
	//Remove entries from the legacy table to prevent constraint issues when deleting the source account
	_, err = tx.Exec(ctx, `UPDATE status set user_id =$1  WHERE user_id=$2`, destinationUserID, sourceUserID)
	if err != nil {
		return nil, fmt.Errorf("tx.Exec [UPDATE status set user_id]: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE "grant" set user_id = $1 where user_id = $2 `, destinationUserID, sourceUserID)
	if err != nil {
		return nil, fmt.Errorf("tx.Exec [UPDATE grant set user_id]: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE phone_numbers AS pn1 SET pn1.user_id = $1 WHERE pn1.user_id = $2 AND NOT EXISTS (
	SELECT 1 FROM phone_numbers AS pn2 WHERE pn2.user_id = $1 AND pn2.phone_number = pn1.phone_number AND pn2.type = pn1.type
)`, destinationUserID, sourceUserID)
	if err != nil {
		return nil, fmt.Errorf("tx.Exec [UPDATE phone_numbers set user_id]: %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE user_notification set user_id = $1 where user_id = $2 `, destinationUserID, sourceUserID)
	if err != nil {
		return nil, fmt.Errorf("tx.Exec [UPDATE user_notification set user_id]: %w", err)
	}

	// commit transaction
	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, fmt.Errorf("tx.Commit: %w", commitErr)
	}

	if err := db.HardDeleteProfile(ctx, uuid.FromStringOrNil(sourceUserID)); err != nil {
		return nil, fmt.Errorf("db.HardDeleteProfile: %w", err)
	}

	return &EmailKeycloakAndUserIDBody{Email: &destinationEmail, KeycloakID: &data.DestinationId, UserID: &destinationUserID}, nil
}
