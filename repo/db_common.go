package repo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
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
	ordersServiceFactory orders.OrdersServiceFactory
}

func NewProfileDB(ctx context.Context, databaseURL string) (*ProfileDB, error) {
	pool, err := pgxpool.Connect(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &ProfileDB{
		Pool:                 pool,
		ordersServiceFactory: orders.OrdersAPIFactory,
	}, nil
}

func MakeDBURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		url.QueryEscape(common.Config.PgUser),
		url.QueryEscape(common.Config.PgPass),
		common.Config.PgHost,
		common.Config.PgPort,
		url.QueryEscape(common.Config.PgDbName))
}

func SyncDBStructInsertionAndMigrations() error {
	log.Println("Syncing starting DB Struct Insertion and Migrations")
	m, err := migrate.New("file://./db/migrations", MakeDBURL()+"?sslmode=disable")
	if err != nil {
		log.Printf("Error while creating migrate instance :: %v\n", err)
		return err
	}
	defer m.Close()

	// Syncing Table struct (UP Mig), Insertion ( Up Mig ) & UP Migrations
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("No changes in UP migration")
			return nil
		}
		return err
	}

	return nil
}
