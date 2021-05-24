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

	tag, err := tx.Exec(ctx, `UPDATE users SET deleted=true WHERE keycloak_id=$1`, keycloakID)
	if err != nil {
		return fmt.Errorf("problem deleting user for keycloak id %q: %w", keycloakID, err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("%w: %q", errProfileNotFound, keycloakID)
	}

	return tx.Commit(ctx)
}
