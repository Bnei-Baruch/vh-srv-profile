package repo

import (
	"context"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/events"
)

// testEventBuilder is a simple test implementation of EventBuilder
type testEventBuilder struct{}

func (b *testEventBuilder) BuildEvent(eventType string, payload map[string]interface{}) events.Event {
	return events.MakeEvent(eventType, payload)
}

// testContext creates a context with EventBuilder for tests
func testContext() context.Context {
	return context.WithValue(context.Background(), common.CtxEventBuilder, &testEventBuilder{})
}

func Test_ProfileDB_HardDeleteProfile_succeeds(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	_, err := db.Exec(ctx, `
		INSERT into users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email)
		VALUES ('22000000-0000-0000-0000-000000000000', '11000000-0000-0000-0000-000000000000', 'first name', 'last name',
		'someemail@email.email')`)
	require.NoError(t, err)

	assert.NoError(t, db.HardDeleteProfile(ctx, uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")))

	var usersCount int
	require.NoError(t, db.QueryRow(ctx, `SELECT COUNT(1) FROM users`).Scan(&usersCount))
	assert.Zero(t, usersCount)
}

func Test_ProfileDB_HardDeleteProfile_with_phone_numbers_succeeds(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	_, err := db.Exec(ctx, `
		INSERT into users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email)
		VALUES ('22000000-0000-0000-0000-000000000000', '11000000-0000-0000-0000-000000000000', 'first name', 'last name',
		'someemail@email.email')`)
	require.NoError(t, err)
	_, err = db.Exec(ctx, `
		INSERT into phone_numbers (user_id, phone_number, type) VALUES
		('22000000-0000-0000-0000-000000000000', '0100000000', 'mobile'),
		('22000000-0000-0000-0000-000000000000', '0200000000', 'WhatsApp')`)
	require.NoError(t, err)

	assert.NoError(t, db.HardDeleteProfile(ctx, uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")))

	var phoneCount int
	require.NoError(t, db.QueryRow(ctx, `SELECT COUNT(1) FROM phone_numbers`).Scan(&phoneCount))
	assert.Zero(t, phoneCount)

	var usersCount int
	require.NoError(t, db.QueryRow(ctx, `SELECT COUNT(1) FROM users`).Scan(&usersCount))
	assert.Zero(t, usersCount)
}

func Test_ProfileDB_HardDeleteProfile_unlinks_spouse(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID1 := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	keycloakID2 := uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")

	// Create two married users
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, spouse_keycloak_id, marital_status)
		VALUES
			('33000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com', $2, 'Married'),
			('44000000-0000-0000-0000-000000000000', $2, 'User', 'Two', 'user2@email.com', $1, 'Married')`,
		keycloakID1, keycloakID2)
	require.NoError(t, err)

	// Hard delete user 1
	err = db.HardDeleteProfile(ctx, keycloakID1)
	require.NoError(t, err)

	// Verify user 1 is deleted
	var count int
	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE keycloak_id = $1`, keycloakID1).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Verify user 2 is unlinked and set to Divorced
	var spouseID *string
	var maritalStatus string
	err = db.QueryRow(ctx, `SELECT spouse_keycloak_id, marital_status FROM users WHERE keycloak_id = $1`, keycloakID2).Scan(&spouseID, &maritalStatus)
	require.NoError(t, err)
	assert.Nil(t, spouseID)
	assert.Equal(t, "Divorced", maritalStatus)
}

func Test_ProfileDB_HardDeleteProfile_without_spouse_succeeds(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")

	// Create a user without spouse
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email)
		VALUES ('33000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com')`,
		keycloakID)
	require.NoError(t, err)

	// Hard delete user
	err = db.HardDeleteProfile(ctx, keycloakID)
	require.NoError(t, err)

	// Verify user is deleted
	var count int
	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE keycloak_id = $1`, keycloakID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func Test_ProfileDB_HardDeleteProfile_with_spouse_and_phone_numbers(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID1 := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	keycloakID2 := uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")

	// Create two married users with phone numbers
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, spouse_keycloak_id, marital_status)
		VALUES
			('33000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com', $2, 'Married'),
			('44000000-0000-0000-0000-000000000000', $2, 'User', 'Two', 'user2@email.com', $1, 'Married')`,
		keycloakID1, keycloakID2)
	require.NoError(t, err)

	// Add phone numbers for user 1
	_, err = db.Exec(ctx, `
		INSERT into phone_numbers (user_id, phone_number, type) VALUES
		('33000000-0000-0000-0000-000000000000', '0100000000', 'mobile'),
		('33000000-0000-0000-0000-000000000000', '0200000000', 'WhatsApp')`)
	require.NoError(t, err)

	// Hard delete user 1
	err = db.HardDeleteProfile(ctx, keycloakID1)
	require.NoError(t, err)

	// Verify user 1 is deleted
	var count int
	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE keycloak_id = $1`, keycloakID1).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Verify phone numbers are deleted
	var phoneCount int
	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM phone_numbers WHERE user_id = '33000000-0000-0000-0000-000000000000'`).Scan(&phoneCount)
	require.NoError(t, err)
	assert.Equal(t, 0, phoneCount)

	// Verify user 2 is unlinked and set to Divorced
	var spouseID *string
	var maritalStatus string
	err = db.QueryRow(ctx, `SELECT spouse_keycloak_id, marital_status FROM users WHERE keycloak_id = $1`, keycloakID2).Scan(&spouseID, &maritalStatus)
	require.NoError(t, err)
	assert.Nil(t, spouseID)
	assert.Equal(t, "Divorced", maritalStatus)
}
