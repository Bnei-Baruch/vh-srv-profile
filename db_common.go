package main

import (
	"context"
	"fmt"
	"os"

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

func getDBUser() string {
	if value, ok := os.LookupEnv("DB_USER"); ok {
		return value
	}
	return "DEFAULT_USER"
}
func getDBPassword() string {
	if value, ok := os.LookupEnv("DB_PASSWORD"); ok {
		return value
	}
	return "DEFAULT_PASS"
}
func getDBHost() string {
	if value, ok := os.LookupEnv("DB_HOST"); ok {
		return value
	}
	return "db"
}

func getDBPort() string {
	if value, ok := os.LookupEnv("DB_PORT"); ok {
		return value
	}
	return "5432"
}

func getDBName() string {
	if value, ok := os.LookupEnv("DB_NAME"); ok {
		return value
	}
	return "default"
}

func makeDBURL() string {
	db_url := getEnvOrFatal("DATABASE_URL")
	//  "postgres://${PROD_PROFILE_DB_USER}:${PROD_PROFILE_DB_PASS}@${PROD_PG_HOST}:${PROD_PG_PORT}/${PROD_PROFILE_DB_NAME}"
	if "" == db_url {
		db_user := getDBUser()
		db_pass := getDBPassword()
		db_host := getDBHost()
		db_port := getDBPort()
		db_name := getDBName()
		db_url = "postgres://" + db_user + ":" + db_pass + "@" + db_host + ":" + db_port + "/" + db_name
	}
	return db_url
}

func newPgProfileDB(ctx context.Context, databaseURL string) (*pgProfileDB, error) {
	pool, err := pgxpool.Connect(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &pgProfileDB{pool}, nil
}
