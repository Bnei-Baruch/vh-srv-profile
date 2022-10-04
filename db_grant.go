package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

func (db *pgProfileDB) getGrantByID(ctx context.Context, id int) (grant, error) {
	var grantRes grant

	if err := db.QueryRow(ctx, `
	SELECT id,
		amount,
		currency,
		type,
		loaned,
		granted,
		repayed,
		created_at,
		updated_at,
		deleted_at 
	FROM "grant" 
	WHERE id = $1`, id).Scan(
		&grantRes.ID,
		&grantRes.Amount,
		&grantRes.Currency,
		&grantRes.Type,
		&grantRes.Loaned,
		&grantRes.Granted,
		&grantRes.Repayed,
		&grantRes.CreatedAt,
		&grantRes.UpdatedAt,
		&grantRes.DeletedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return grant{}, errNotFound
		}
		return grant{}, fmt.Errorf("error while getting grant: %w", err)
	}

	return grantRes, nil
}
