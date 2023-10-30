package repo

import (
	"context"
	"errors"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
)

func Test_ProfileDB_updateProfile_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestProfileDB(t)
	defer newTestProfileDB(t)
	_, err := db.Exec(context.Background(), `
	INSERT into users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email) 
	VALUES ('22000000-0000-0000-0000-000000000000', '11000000-0000-0000-0000-000000000000', 'first name', 'last name', 
	'someemail@email.email')`)
	require.NoError(t, err)

	err = db.UpdateProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"),
		UserInput{FirstNameVernacular: utils.PointerString("updated name")})

	actualRows, err := db.Query(context.Background(), `SELECT deleted, keycloak_id::text, first_name_vernacular, last_name_vernacular,
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
	assert.Equal(t, [][]interface{}{{false, "11000000-0000-0000-0000-000000000000", "updated name", "last name", "someemail@email.email"}}, actual)

	assert.NoError(t, err)
}

func Test_ProfileDB_updateProfile_with_phone_numbers_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestProfileDB(t)
	defer newTestProfileDB(t)
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

	err = db.UpdateProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"),
		UserInput{Phones: Phones{MobileNumber: utils.PointerString("updated number")}})

	actualRows, err := db.Query(context.Background(), `SELECT phone_number, type FROM phone_numbers`)
	require.NoError(t, err)
	defer actualRows.Close()
	var actual [][]interface{}
	for actualRows.Next() {
		actualRow, err := actualRows.Values()
		require.NoError(t, err)

		actual = append(actual, actualRow)
	}
	require.Len(t, actual, 2)
	assert.Equal(t, [][]interface{}{{"0200000000", "WhatsApp"}, {"updated number", "mobile"}}, actual)

	assert.NoError(t, err)
}

func Test_ProfileDB_updateProfile_add_phone_number_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestProfileDB(t)
	defer newTestProfileDB(t)
	_, err := db.Exec(context.Background(), `
	INSERT into users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email) 
	VALUES ('22000000-0000-0000-0000-000000000000', '11000000-0000-0000-0000-000000000000', 'first name', 'last name', 
	'someemail@email.email')`)
	require.NoError(t, err)

	err = db.UpdateProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"),
		UserInput{Phones: Phones{WhatsAppNumber: utils.PointerString("added number")}})

	actualRows, err := db.Query(context.Background(), `SELECT phone_number, type FROM phone_numbers`)
	require.NoError(t, err)
	defer actualRows.Close()
	var actual [][]interface{}
	for actualRows.Next() {
		actualRow, err := actualRows.Values()
		require.NoError(t, err)

		actual = append(actual, actualRow)
	}
	require.Len(t, actual, 1)
	assert.Equal(t, [][]interface{}{{"added number", "WhatsApp"}}, actual)
	assert.NoError(t, err)
}

func Test_ProfileDB_updateProfile_returns_error_on_deleted_profile(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestProfileDB(t)
	defer newTestProfileDB(t)
	_, err := db.Exec(context.Background(), `
	INSERT into users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, deleted) 
	VALUES ('22000000-0000-0000-0000-000000000000', '11000000-0000-0000-0000-000000000000', 'first name', 'last name', 
	'someemail@email.email', true)`)
	require.NoError(t, err)

	err = db.UpdateProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"),
		UserInput{FirstNameVernacular: utils.PointerString("updated name")})

	assert.Error(t, err)
	assert.True(t, errors.Is(err, common.ErrProfileNotFound))
}
