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
		return fmt.Errorf("db.Begin: %w", err)
	}

	defer func() error {
		err := tx.Rollback(ctx)
		if err != nil {
			return fmt.Errorf("tx.Rollback: %w", err)
		}
		return nil
	}()

	tag, err := tx.Exec(ctx, `UPDATE users SET deleted=true WHERE keycloak_id=$1`, keycloakID)
	if err != nil {
		return fmt.Errorf("tx.Exec: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return common.ErrProfileNotFound
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}
	return nil
}
