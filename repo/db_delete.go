package repo

import (
	"context"
	"fmt"

	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

type deleteStorage interface {
	DeleteProfile(ctx context.Context, keycloakID uuid.UUID) error
}

func (db *ProfileDB) DeleteProfile(ctx context.Context, keycloakID uuid.UUID) error {
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
		return fmt.Errorf("%w: %q", common.ErrProfileNotFound, keycloakID)
	}

	return tx.Commit(ctx)
}
