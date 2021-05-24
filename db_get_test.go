package main

import (
	"context"
	"errors"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		firstNameVernacular: pointerString("first name"),
		lastNameVernacular:  pointerString("last name"),
		emails:              emails{primary: pointerString("someemail@email.email")},
		phones: phones{
			mobileNumber:   &expectedMobile,
			whatsAppNumber: &expectedWhatsApp,
		},
	}, actual.userInput)
}

func Test_pgProfileDb_getUser_deleted_true_returns_nothing(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)
	_, err := db.Exec(context.Background(), `
	INSERT into users (keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, deleted) 
	VALUES ('11000000-0000-0000-0000-000000000000', 'first name', 'last name', 'someemail@email.email', true)`)
	require.NoError(t, err)

	_, err = db.getProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"))

	assert.Error(t, err)
	assert.True(t, errors.Is(err, errProfileNotFound))
}
