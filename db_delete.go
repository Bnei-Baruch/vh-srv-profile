package main

import (
	"context"
	"fmt"

	uuid "github.com/satori/go.uuid"
)

func (db *pgProfileDB) deleteProfile(ctx context.Context, keycloakID uuid.UUID) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `DELETE FROM phone_numbers WHERE user_id=(SELECT user_id FROM users WHERE keycloak_id=$1)`, keycloakID)
	if err != nil {
		return fmt.Errorf("problem deleting phone numbers for keycloak id %q: %w", keycloakID, err)
	}

	tag, err := tx.Exec(ctx, `DELETE FROM users WHERE keycloak_id=$1`, keycloakID)
	if err != nil {
		return fmt.Errorf("problem deleting users for keycloak id %q: %w", keycloakID, err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("no profile found for keycloak id %q", keycloakID)
	}

	return tx.Commit(ctx)
}
