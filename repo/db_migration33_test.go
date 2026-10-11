package repo

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/events"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/testutil"
)

func requireCheckViolation(t *testing.T, err error, constraint string) {
	t.Helper()
	var pgErr *pgconn.PgError
	require.True(t, errors.As(err, &pgErr), "want a postgres error, got %v", err)
	assert.Equal(t, "23514", pgErr.Code)
	assert.Equal(t, constraint, pgErr.ConstraintName)
}

// Migration 33 clears the expiry left on types without one and the ”
// alternates, then adds the CHECKs that keep them out.
func TestMigration33_ClearsLeftoversAndAddsChecks(t *testing.T) {
	ctx := context.Background()
	dbURL, err := testutil.NewTestProfileDB(t, ctx)
	require.NoError(t, err)
	emitter, err := events.CreateEmitter()
	require.NoError(t, err)
	db, err := NewProfileDB(ctx, dbURL, emitter)
	require.NoError(t, err)
	t.Cleanup(db.Close)
	stubOrdersService(db)

	m, err := migrate.New("file://../db/migrations", dbURL)
	require.NoError(t, err)
	t.Cleanup(func() { m.Close() })
	require.NoError(t, m.Migrate(32))

	_, userID := createTestProfile(t, db, 50, "m33@test.test")
	_, err = db.Exec(ctx, `UPDATE membership SET type = 'cancelled', expiry = '2021-10-23' WHERE user_id = $1`, userID)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `UPDATE users SET alternate_email_1 = '', alternate_email_2 = '' WHERE user_id = $1`, userID)
	require.NoError(t, err)

	require.NoError(t, m.Migrate(33))

	assert.False(t, membershipExpiry(t, db, userID).Valid, "leftover expiry cleared")
	var alt1, alt2 sql.NullString
	require.NoError(t, db.QueryRow(ctx, `SELECT alternate_email_1, alternate_email_2 FROM users WHERE user_id = $1`, userID).Scan(&alt1, &alt2))
	assert.False(t, alt1.Valid)
	assert.False(t, alt2.Valid)

	for _, tc := range []struct{ name, sql, constraint string }{
		{"automatic with expiry", `UPDATE membership SET type = 'automatic', expiry = now() WHERE user_id = $1`, "membership_expiry_matches_type"},
		{"manual without expiry", `UPDATE membership SET type = 'manual', expiry = NULL WHERE user_id = $1`, "membership_expiry_matches_type"},
		{"empty alternate 1", `UPDATE users SET alternate_email_1 = '' WHERE user_id = $1`, "users_alternate_email_1_not_empty"},
		{"empty alternate 2", `UPDATE users SET alternate_email_2 = '' WHERE user_id = $1`, "users_alternate_email_2_not_empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.Exec(ctx, tc.sql, userID)
			requireCheckViolation(t, err, tc.constraint)
		})
	}

	// Evaluation writes rows the constraint accepts.
	_, err = db.EvaluateMembershipByUserID(testContext(), EmailKeycloakAndUserIDBody{UserID: &userID})
	require.NoError(t, err)
}
