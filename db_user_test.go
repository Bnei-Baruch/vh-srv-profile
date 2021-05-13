package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func checkIntegrationTest(t testing.TB) {
	t.Helper()
	if os.Getenv("GO_INTEGRATION_TESTS") != "1" {
		t.SkipNow()
	}
}

func newTestPgProfileDb(t *testing.T) *pgProfileDB {
	t.Helper()
	db, err := newPgProfileDB(context.Background(), os.Getenv("DATABASE_URL"))
	require.NoError(t, err)

	_, err = db.Exec(context.Background(), `TRUNCATE users CASCADE`)
	require.NoError(t, err)

	return db
}

func Test_newPgProfileDb(t *testing.T) {
	checkIntegrationTest(t)
	_, err := newPgProfileDB(context.Background(), os.Getenv("DATABASE_URL"))
	assert.NoError(t, err)
}

func Test_createUser_with_minimum_info_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)

	err := db.createUser(context.Background(), user{
		keycloakID:          "some keycloak id",
		firstNameVernacular: "first name",
		lastNameVernacular:  "last name ",
		emails:              emails{primary: "someemail@email.email"},
	})

	assert.NoError(t, err)
}
