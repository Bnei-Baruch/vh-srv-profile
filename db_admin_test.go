package main

import (
	"context"
	"errors"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_pgProfileDb_hardDelete_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)
	_, err := db.Exec(context.Background(), `
	INSERT into users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email) 
	VALUES ('22000000-0000-0000-0000-000000000000', '11000000-0000-0000-0000-000000000000', 'first name', 'last name', 
	'someemail@email.email')`)
	require.NoError(t, err)

	assert.NoError(t, db.hardDeleteProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")))

	var usersCount int
	require.NoError(t, db.QueryRow(context.Background(), `SELECT COUNT(1) FROM users`).Scan(&usersCount))
	assert.Zero(t, usersCount)
}

func Test_pgProfileDb_hardDelete_with_phone_numbers_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)
	_, err := db.Exec(context.Background(), `
	INSERT into users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email) 
	VALUES ('22000000-0000-0000-0000-000000000000', '11000000-0000-0000-0000-000000000000', 'first name', 'last name', 
	'someemail@email.email')`)
	require.NoError(t, err)
	_, err = db.Exec(context.Background(), `
	INSERT into phone_numbers (user_id, phone_number, type) VALUES 
	('22000000-0000-0000-0000-000000000000', '0100000000', 'mobile'),
	('22000000-0000-0000-0000-000000000000', '0200000000', 'WhatsApp')`)
	require.NoError(t, err)

	assert.NoError(t, db.hardDeleteProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")))

	var phoneCount int
	require.NoError(t, db.QueryRow(context.Background(), `SELECT COUNT(1) FROM phone_numbers`).Scan(&phoneCount))
	assert.Zero(t, phoneCount)

	var usersCount int
	require.NoError(t, db.QueryRow(context.Background(), `SELECT COUNT(1) FROM users`).Scan(&usersCount))
	assert.Zero(t, usersCount)
}

func Test_pgProfileDb_hardDelete_inexistant_user_returns_error(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)

	err := db.hardDeleteProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errProfileNotFound))
}
