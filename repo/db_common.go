package repo

import (
	"context"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
)

type ProfileRepository interface {
	createStorage
	updateStorage
	deleteStorage
	readStorage
	readMultipleProfileStorage
	hardDeleteStorage
	createRequestStorage
	updateRequestStorage
	deleteRequestStorage
	readMultipleRequestStorage
	membershipInterface
	grantInterface
	notificationInterface
	userNotificationInterface
	operationInterface
}

type ProfileDB struct {
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

func MakeDBURL() string {
	db_user := getDBUser()
	db_pass := getDBPassword()
	db_host := getDBHost()
	db_port := getDBPort()
	db_name := getDBName()
	return "postgres://" + db_user + ":" + db_pass + "@" + db_host + ":" + db_port + "/" + db_name
}

func NewProfileDB(ctx context.Context, databaseURL string) (*ProfileDB, error) {
	pool, err := pgxpool.Connect(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &ProfileDB{pool}, nil
}

func SyncDBStructInsertionAndMigrations() error {
	fmt.Println("Syncing starting DB Struct Insertion and Migrations")
	m, err := migrate.New(
		"file://./db/migrations", MakeDBURL()+"?sslmode=disable")
	if err != nil {
		fmt.Println("Error while creating migrate instance :: ", err)
		return err
	}
	// Syncing Table struct (UP Mig), Insertion ( Up Mig ) & UP Migrations
	if err := m.Up(); err != nil {
		m.Close()
		if err == migrate.ErrNoChange {
			fmt.Println("No changes in UP migration")
			return nil
		}
		return err
	}
	m.Close()
	fmt.Println("UP Migration Done!")
	return nil
}
