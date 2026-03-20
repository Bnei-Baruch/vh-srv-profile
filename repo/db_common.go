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
	"gitlab.bbdev.team/vh/vh-srv-profile/events"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
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
	pageNoteInterface
	Close()
}

type ProfileDB struct {
	*pgxpool.Pool
	eventEmitter         events.EventEmitter
	ordersServiceFactory orders.OrdersServiceFactory
}

func NewProfileDB(ctx context.Context, databaseURL string, eventEmitter events.EventEmitter) (*ProfileDB, error) {
	pool, err := pgxpool.Connect(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	db := &ProfileDB{
		Pool:                 pool,
		eventEmitter:         eventEmitter,
		ordersServiceFactory: orders.OrdersAPIFactory,
	}

	if err := InitNotificationRegistry(db); err != nil {
		return nil, fmt.Errorf("InitNotificationRegistry: %w", err)
	}

	return db, nil
}

func (db *ProfileDB) SetEventEmitter(emitter events.EventEmitter) {
	db.eventEmitter = emitter
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

func (db *ProfileDB) emitEvent(ctx context.Context, eventType string, payload map[string]interface{}) {
	builder := ctx.Value(common.CtxEventBuilder).(events.EventBuilder)
	event := builder.BuildEvent(eventType, payload)
	db.eventEmitter.Emit(ctx, event)
}

// emitUpdateProfileEvents emits update_profile events for the given keycloak IDs.
// It fetches user_ids in a single query and emits one event per user.
func (db *ProfileDB) emitUpdateProfileEvents(ctx context.Context, affectedKeycloakStringIDs []string) {
	// Uncomment when running manual imports to suppress event emission.
	// if db.eventEmitter == nil {
	// 	return
	// }
	if len(affectedKeycloakStringIDs) == 0 {
		return
	}

	rows, err := db.Query(ctx, `SELECT keycloak_id, user_id FROM users WHERE keycloak_id::text = ANY($1)`, affectedKeycloakStringIDs)
	if err != nil {
		utils.LogFor(ctx).Error("emitUpdateProfileEvents: Query", slog.Any("err", err))
		utils.SentryFor(ctx).CaptureException(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var keycloakID, userID string
		if err := rows.Scan(&keycloakID, &userID); err != nil {
			utils.LogFor(ctx).Error("emitUpdateProfileEvents: Scan", slog.Any("err", err))
			utils.SentryFor(ctx).CaptureException(err)
			continue
		}
		db.emitEvent(ctx, events.TypeUpdateProfile, map[string]interface{}{
			"keycloak_id": keycloakID,
			"user_id":     userID,
		})
	}
}
