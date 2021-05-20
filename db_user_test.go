package main

import (
	"context"
	"os"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
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

	err := db.createProfile(context.Background(), userInput{
		keycloakID:          pointerUUID(uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")),
		firstNameVernacular: pointerString("first name"),
		lastNameVernacular:  pointerString("last name"),
		emails:              emails{primary: pointerString("someemail@email.email")},
	})

	assert.NoError(t, err)
	actualRows, err := db.Query(context.Background(), `SELECT deleted, keycloak_id, first_name_vernacular, last_name_vernacular,
	primary_email FROM users`)
	require.NoError(t, err)
	defer actualRows.Close()
	var actual [][]interface{}
	for actualRows.Next() {
		actualRow, err := actualRows.Values()
		require.NoError(t, err)

		actual = append(actual, actualRow)
	}
	require.Len(t, actual, 1)
	assert.Equal(t, [][]interface{}{{false, "11000000-0000-0000-0000-000000000000", "first name", "last name", "someemail@email.email"}}, actual)

	var createdAt, updatedAt time.Time
	require.NoError(t, db.QueryRow(context.Background(), `SELECT created_at, updated_at FROM users`).Scan(&createdAt, &updatedAt))
	assert.WithinDuration(t, time.Now(), createdAt, 3*time.Second)
	assert.WithinDuration(t, time.Now(), updatedAt, 3*time.Second)
}

func Test_pgProfileDb_createUser_with_phone_number_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)

	mobile := "0100000000"
	whatsApp := "0200000000"
	err := db.createProfile(context.Background(), userInput{
		keycloakID:          pointerUUID(uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")),
		firstNameVernacular: pointerString("first name"),
		lastNameVernacular:  pointerString("last name"),
		emails:              emails{primary: pointerString("someemail@email.email")},
		phones: phones{
			mobileNumber:   &mobile,
			whatsAppNumber: &whatsApp,
		},
	})

	assert.NoError(t, err)

	actualRows, err := db.Query(context.Background(), `SELECT phone_number, type FROM phone_numbers`)
	require.NoError(t, err)
	defer actualRows.Close()
	var actualPhoneNumbers [][]interface{}
	for actualRows.Next() {
		actual, err := actualRows.Values()
		require.NoError(t, err)

		actualPhoneNumbers = append(actualPhoneNumbers, actual)
	}
	assert.Len(t, actualPhoneNumbers, 2)
	assert.Contains(t, actualPhoneNumbers, []interface{}{"0100000000", "mobile"})
	assert.Contains(t, actualPhoneNumbers, []interface{}{"0200000000", "WhatsApp"})
}

func Test_pgProfileDb_getUser_minimal_data_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)
	_, err := db.Exec(context.Background(), `
	INSERT into users (keycloak_id, first_name_vernacular, last_name_vernacular, primary_email) 
	VALUES ('11000000-0000-0000-0000-000000000000', 'first name', 'last name', 'someemail@email.email')`)
	require.NoError(t, err)

	actual, err := db.getProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"))

	assert.NoError(t, err)
	assert.Equal(t, userInput{
		keycloakID:          pointerUUID(uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")),
		firstNameVernacular: pointerString("first name"),
		lastNameVernacular:  pointerString("last name"),
		emails:              emails{primary: pointerString("someemail@email.email")},
	}, actual.userInput)
	assert.WithinDuration(t, time.Now(), actual.createdAt, 3*time.Second)
	assert.WithinDuration(t, time.Now(), actual.updatedAt, 3*time.Second)
	assert.False(t, actual.deleted)
}

func Test_pgProfileDb_getUser_with_phone_numbers_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)
	var userID uuid.UUID
	err := db.QueryRow(context.Background(), `
	INSERT into users (keycloak_id, first_name_vernacular, last_name_vernacular, primary_email) 
	VALUES ('11000000-0000-0000-0000-000000000000', 'first name', 'last name', 'someemail@email.email') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)
	_, err = db.Exec(context.Background(), `
	INSERT into phone_numbers (user_id, phone_number, type) VALUES ($1, '0100000000', 'mobile'), ($1, '0200000000', 'WhatsApp')`, userID)
	require.NoError(t, err)

	actual, err := db.getProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"))

	expectedMobile := "0100000000"
	expectedWhatsApp := "0200000000"
	assert.NoError(t, err)
	assert.Equal(t, userInput{
		keycloakID:          pointerUUID(uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")),
		firstNameVernacular: pointerString("first name"),
		lastNameVernacular:  pointerString("last name"),
		emails:              emails{primary: pointerString("someemail@email.email")},
		phones: phones{
			mobileNumber:   &expectedMobile,
			whatsAppNumber: &expectedWhatsApp,
		},
	}, actual.userInput)
}
