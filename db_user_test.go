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

func Test_pgProfileDb_createUser_with_minimum_info_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)

	err := db.createUser(context.Background(), user{
		keycloakID:          "some keycloak id",
		firstNameVernacular: "first name",
		lastNameVernacular:  "last name",
		emails:              emails{primary: "someemail@email.email"},
	})

	assert.NoError(t, err)
	actualRows, err := db.Query(context.Background(), `SELECT keycloak_id, first_name_vernacular, last_name_vernacular,
	primary_email FROM users`)
	require.NoError(t, err)
	defer actualRows.Close()
	var actual [][]interface{}
	for actualRows.Next() {
		actualRow, err := actualRows.Values()
		require.NoError(t, err)

		actual = append(actual, actualRow)
	}
	assert.Len(t, actual, 1)
	assert.Equal(t, [][]interface{}{{"some keycloak id", "first name", "last name", "someemail@email.email"}}, actual)

}

func Test_pgProfileDb_getUser_minimal_data_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)
	_, err := db.Exec(context.Background(), `
	INSERT into users (keycloak_id, first_name_vernacular, last_name_vernacular, primary_email) 
	VALUES ('some keycloak id', 'first name', 'last name', 'someemail@email.email')`)
	require.NoError(t, err)

	actual, err := db.getUser(context.Background(), "some keycloak id")

	assert.NoError(t, err)
	assert.Equal(t, user{
		keycloakID:          "some keycloak id",
		firstNameVernacular: "first name",
		lastNameVernacular:  "last name",
		emails:              emails{primary: "someemail@email.email"},
	}, actual)
}
