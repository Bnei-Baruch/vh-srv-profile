package main

import (
	"context"
	"fmt"
)

func (db *pgProfileDB) deleteRequest(ctx context.Context, id int) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `DELETE FROM request WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("problem deleting request for id %d: %w", id, err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("%w: %q", errProfileNotFound, id)
	}

	return tx.Commit(ctx)
}
