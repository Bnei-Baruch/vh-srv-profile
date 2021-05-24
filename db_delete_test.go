package main

import (
	"context"
	"errors"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_pgProfileDb_delete_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)
	_, err := db.Exec(context.Background(), `
	INSERT into users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email) 
	VALUES ('22000000-0000-0000-0000-000000000000', '11000000-0000-0000-0000-000000000000', 'first name', 'last name', 
	'someemail@email.email')`)
	require.NoError(t, err)

	assert.NoError(t, db.deleteProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")))

	var deleted bool
	require.NoError(t, db.QueryRow(context.Background(), `SELECT deleted FROM users WHERE keycloak_id=$1`,
		"11000000-0000-0000-0000-000000000000").Scan(&deleted))
	assert.True(t, deleted)
}

func Test_pgProfileDb_delete_inexistant_user_returns_error(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)

	err := db.deleteProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errProfileNotFound))
}
