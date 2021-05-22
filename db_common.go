package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	mobile   = "mobile"
	whatsApp = "WhatsApp"
	telegram = "Telegram"
)

type pgProfileDB struct {
	*pgxpool.Pool
}

func newPgProfileDB(ctx context.Context, databaseURL string) (*pgProfileDB, error) {
	pool, err := pgxpool.Connect(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &pgProfileDB{pool}, nil
}
