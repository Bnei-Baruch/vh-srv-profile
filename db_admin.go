package main

import (
	"context"
	"fmt"

	uuid "github.com/satori/go.uuid"
)

func (db *pgProfileDB) hardDeleteProfile(ctx context.Context, keycloakID uuid.UUID) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `DELETE FROM phone_numbers WHERE user_id=(SELECT user_id FROM users WHERE keycloak_id=$1)`, keycloakID)
	if err != nil {
		return fmt.Errorf("problem deleting phone numbers for keycloak id %q: %w", keycloakID, err)
	}

	/* Delete status related to the user */
	_, err = tx.Exec(ctx, `DELETE FROM status WHERE user_id=(SELECT user_id FROM users WHERE keycloak_id=$1)`, keycloakID)
	if err != nil {
		return fmt.Errorf("problem deleting phone numbers for keycloak id %q: %w", keycloakID, err)
	}

	// delete membership
	_, err = tx.Exec(ctx, `DELETE FROM membership WHERE user_id=(SELECT user_id FROM users WHERE keycloak_id=$1)`, keycloakID)
	if err != nil {
		return fmt.Errorf("problem deleting membership for keycloak id %q: %w", keycloakID, err)
	}

	// delete membership_helphaver
	_, err = tx.Exec(ctx, `DELETE FROM membership_helphaver WHERE membership_id IN (SELECT membership_id FROM membership WHERE user_id=(SELECT user_id FROM users WHERE keycloak_id=$1))`, keycloakID)
	if err != nil {
		return fmt.Errorf("problem deleting membership_helphaver for keycloak id %q: %w", keycloakID, err)
	}

	// delete membership_automatic
	_, err = tx.Exec(ctx, `DELETE FROM membership_automatic WHERE membership_id IN (SELECT membership_id FROM membership WHERE user_id=(SELECT user_id FROM users WHERE keycloak_id=$1))`, keycloakID)
	if err != nil {
		return fmt.Errorf("problem deleting membership_automatic for keycloak id %q: %w", keycloakID, err)
	}

	// delete membership_manual
	_, err = tx.Exec(ctx, `DELETE FROM membership_manual WHERE membership_id IN (SELECT membership_id FROM membership WHERE user_id=(SELECT user_id FROM users WHERE keycloak_id=$1))`, keycloakID)
	if err != nil {
		return fmt.Errorf("problem deleting membership_manual for keycloak id %q: %w", keycloakID, err)
	}

	// delete membership_special
	_, err = tx.Exec(ctx, `DELETE FROM membership_special WHERE membership_id IN (SELECT membership_id FROM membership WHERE user_id=(SELECT user_id FROM users WHERE keycloak_id=$1))`, keycloakID)
	if err != nil {
		return fmt.Errorf("problem deleting membership_special for keycloak id %q: %w", keycloakID, err)
	}

	// delete membership_helphaver
	_, err = tx.Exec(ctx, `DELETE FROM membership_helphaver WHERE membership_id IN (SELECT membership_id FROM membership WHERE user_id=(SELECT user_id FROM users WHERE keycloak_id=$1))`, keycloakID)
	if err != nil {
		return fmt.Errorf("problem deleting membership_helphaver for keycloak id %q: %w", keycloakID, err)
	}

	// delete grant
	_, err = tx.Exec(ctx, `DELETE FROM grant WHERE user_id=(SELECT user_id FROM users WHERE keycloak_id=$1)`, keycloakID)
	if err != nil {
		return fmt.Errorf("problem deleting grant for keycloak id %q: %w", keycloakID, err)
	}

	// delete user_notification
	_, err = tx.Exec(ctx, `DELETE FROM user_notification WHERE user_id=(SELECT user_id FROM users WHERE keycloak_id=$1)`, keycloakID)
	if err != nil {
		return fmt.Errorf("problem deleting user_notification for keycloak id %q: %w", keycloakID, err)
	}

	tag, err := tx.Exec(ctx, `DELETE FROM users WHERE keycloak_id=$1`, keycloakID)
	if err != nil {
		return fmt.Errorf("problem deleting users for keycloak id %q: %w", keycloakID, err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("%w: %q", errProfileNotFound, keycloakID)
	}

	return tx.Commit(ctx)
}
