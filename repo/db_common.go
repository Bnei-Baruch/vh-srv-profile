package repo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	readMultipleRequestStorage
	membershipInterface
	grantInterface
	notificationInterface
	userNotificationInterface
	operationInterface
	mergeAccounts
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

	db := &ProfileDB{
		Pool:                 pool,
		ordersServiceFactory: orders.OrdersAPIFactory,
	}

	if err := InitNotificationRegistry(db); err != nil {
		return nil, fmt.Errorf("InitNotificationRegistry: %w", err)
	}

	return db, nil
}

func (db *ProfileDB) SetOrdersServiceFactory(factory orders.OrdersServiceFactory) {
	db.ordersServiceFactory = factory
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
	slog.Info("running db migrations")
	m, err := migrate.New("file://./db/migrations", MakeDBURL()+"?sslmode=disable")
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("no changes in migrations")
			return nil
		}
		return fmt.Errorf("migrate.Up: %w", err)
	}

	return nil
}
