package repo

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

func Test_ProfileDB_SetSpouse_prevents_self_spousing(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email)
		VALUES ('22000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com')`, keycloakID)
	require.NoError(t, err)

	err = db.SetSpouse(ctx, keycloakID, keycloakID, false)

	assert.Error(t, err)
	assert.ErrorIs(t, err, common.ErrSpouseSelf)
}

func Test_ProfileDB_SetSpouse_returns_error_when_user1_not_found(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID1 := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	keycloakID2 := uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email)
		VALUES ('33000000-0000-0000-0000-000000000000', $1, 'User', 'Two', 'user2@email.com')`, keycloakID2)
	require.NoError(t, err)

	err = db.SetSpouse(ctx, keycloakID1, keycloakID2, false)

	assert.Error(t, err)
	assert.ErrorIs(t, err, common.ErrUserNotFound)
}

func Test_ProfileDB_SetSpouse_returns_error_when_user2_not_found(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID1 := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	keycloakID2 := uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email)
		VALUES ('33000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com')`, keycloakID1)
	require.NoError(t, err)

	err = db.SetSpouse(ctx, keycloakID1, keycloakID2, false)

	assert.Error(t, err)
	assert.ErrorIs(t, err, common.ErrUserNotFound)
}

func Test_ProfileDB_SetSpouse_cancels_existing_spouse_relationship(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID1 := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	keycloakID2 := uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, spouse_keycloak_id, marital_status)
		VALUES
			('33000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com', $2, 'Married'),
			('44000000-0000-0000-0000-000000000000', $2, 'User', 'Two', 'user2@email.com', $1, 'Married')`,
		keycloakID1, keycloakID2)
	require.NoError(t, err)

	err = db.SetSpouse(ctx, keycloakID1, uuid.Nil, false)
	require.NoError(t, err)

	// Verify both users are unlinked and set to Single
	actualRows, err := db.Query(ctx, `
		SELECT keycloak_id::text, spouse_keycloak_id, marital_status
		FROM users
		ORDER BY keycloak_id`)
	require.NoError(t, err)
	defer actualRows.Close()

	var actual [][]interface{}
	for actualRows.Next() {
		actualRow, err := actualRows.Values()
		require.NoError(t, err)
		actual = append(actual, actualRow)
	}
	require.Len(t, actual, 2)
	assert.Equal(t, keycloakID1.String(), actual[0][0])
	assert.Nil(t, actual[0][1])
	assert.Equal(t, "Divorced", actual[0][2])
	assert.Equal(t, keycloakID2.String(), actual[1][0])
	assert.Nil(t, actual[1][1])
	assert.Equal(t, "Divorced", actual[1][2])
}

func Test_ProfileDB_SetSpouse_cancels_when_no_existing_spouse(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID1 := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, marital_status)
		VALUES ('33000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com', 'Divorced')`,
		keycloakID1)
	require.NoError(t, err)

	err = db.SetSpouse(ctx, keycloakID1, uuid.Nil, false)
	require.NoError(t, err)

	// Verify user is set to Divorced
	var maritalStatus string
	err = db.QueryRow(ctx, `SELECT marital_status FROM users WHERE keycloak_id = $1`, keycloakID1).Scan(&maritalStatus)
	require.NoError(t, err)
	assert.Equal(t, "Divorced", maritalStatus)
}

func Test_ProfileDB_SetSpouse_sets_spouse_for_two_single_users(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID1 := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	keycloakID2 := uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email)
		VALUES
			('33000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com'),
			('44000000-0000-0000-0000-000000000000', $2, 'User', 'Two', 'user2@email.com')`,
		keycloakID1, keycloakID2)
	require.NoError(t, err)

	err = db.SetSpouse(ctx, keycloakID1, keycloakID2, false)
	require.NoError(t, err)

	// Verify both users are linked and set to Married
	actualRows, err := db.Query(ctx, `
		SELECT keycloak_id::text, spouse_keycloak_id, marital_status
		FROM users
		ORDER BY keycloak_id`)
	require.NoError(t, err)
	defer actualRows.Close()

	var actual [][]interface{}
	for actualRows.Next() {
		actualRow, err := actualRows.Values()
		require.NoError(t, err)
		actual = append(actual, actualRow)
	}
	require.Len(t, actual, 2)
	assert.Equal(t, keycloakID1.String(), actual[0][0])
	assert.Equal(t, keycloakID2.String(), actual[0][1])
	assert.Equal(t, "Married", actual[0][2])
	assert.Equal(t, keycloakID2.String(), actual[1][0])
	assert.Equal(t, keycloakID1.String(), actual[1][1])
	assert.Equal(t, "Married", actual[1][2])
}

func Test_ProfileDB_SetSpouse_is_idempotent_when_already_married(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID1 := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	keycloakID2 := uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, spouse_keycloak_id, marital_status)
		VALUES
			('33000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com', $2, 'Married'),
			('44000000-0000-0000-0000-000000000000', $2, 'User', 'Two', 'user2@email.com', $1, 'Married')`,
		keycloakID1, keycloakID2)
	require.NoError(t, err)

	// Call SetSpouse again with same pair
	err = db.SetSpouse(ctx, keycloakID1, keycloakID2, false)
	require.NoError(t, err)

	// Verify relationship is still intact
	actualRows, err := db.Query(ctx, `
		SELECT keycloak_id::text, spouse_keycloak_id, marital_status
		FROM users
		ORDER BY keycloak_id`)
	require.NoError(t, err)
	defer actualRows.Close()

	var actual [][]interface{}
	for actualRows.Next() {
		actualRow, err := actualRows.Values()
		require.NoError(t, err)
		actual = append(actual, actualRow)
	}
	require.Len(t, actual, 2)
	assert.Equal(t, keycloakID2.String(), actual[0][1])
	assert.Equal(t, "Married", actual[0][2])
	assert.Equal(t, keycloakID1.String(), actual[1][1])
	assert.Equal(t, "Married", actual[1][2])
}

func Test_ProfileDB_SetSpouse_returns_conflict_error_when_user1_has_different_spouse(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID1 := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	keycloakID2 := uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")
	keycloakID3 := uuid.FromStringOrNil("33000000-0000-0000-0000-000000000000")
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, spouse_keycloak_id, marital_status)
		VALUES
			('44000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com', $3, 'Married'),
			('55000000-0000-0000-0000-000000000000', $2, 'User', 'Two', 'user2@email.com', NULL, NULL),
			('66000000-0000-0000-0000-000000000000', $3, 'User', 'Three', 'user3@email.com', $1, 'Married')`,
		keycloakID1, keycloakID2, keycloakID3)
	require.NoError(t, err)

	err = db.SetSpouse(ctx, keycloakID1, keycloakID2, false)

	assert.Error(t, err)
	assert.ErrorIs(t, err, common.ErrSpouseConflict)
}

func Test_ProfileDB_SetSpouse_returns_conflict_error_when_user2_has_different_spouse(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID1 := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	keycloakID2 := uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")
	keycloakID3 := uuid.FromStringOrNil("33000000-0000-0000-0000-000000000000")
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, spouse_keycloak_id, marital_status)
		VALUES
			('44000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com', NULL, NULL),
			('55000000-0000-0000-0000-000000000000', $2, 'User', 'Two', 'user2@email.com', $3, 'Married'),
			('66000000-0000-0000-0000-000000000000', $3, 'User', 'Three', 'user3@email.com', $2, 'Married')`,
		keycloakID1, keycloakID2, keycloakID3)
	require.NoError(t, err)

	err = db.SetSpouse(ctx, keycloakID1, keycloakID2, false)

	assert.Error(t, err)
	assert.ErrorIs(t, err, common.ErrSpouseConflict)
}

func Test_ProfileDB_SetSpouse_returns_conflict_error_when_both_users_have_different_spouses(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID1 := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	keycloakID2 := uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")
	keycloakID3 := uuid.FromStringOrNil("33000000-0000-0000-0000-000000000000")
	keycloakID4 := uuid.FromStringOrNil("44000000-0000-0000-0000-000000000000")
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, spouse_keycloak_id, marital_status)
		VALUES
			('55000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com', $3, 'Married'),
			('66000000-0000-0000-0000-000000000000', $2, 'User', 'Two', 'user2@email.com', $4, 'Married'),
			('77000000-0000-0000-0000-000000000000', $3, 'User', 'Three', 'user3@email.com', $1, 'Married'),
			('88000000-0000-0000-0000-000000000000', $4, 'User', 'Four', 'user4@email.com', $2, 'Married')`,
		keycloakID1, keycloakID2, keycloakID3, keycloakID4)
	require.NoError(t, err)

	err = db.SetSpouse(ctx, keycloakID1, keycloakID2, false)

	assert.Error(t, err)
	assert.ErrorIs(t, err, common.ErrSpouseConflict)
}

func Test_ProfileDB_SetSpouse_with_force_update_unlinks_user1_previous_spouse(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID1 := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	keycloakID2 := uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")
	keycloakID3 := uuid.FromStringOrNil("33000000-0000-0000-0000-000000000000")
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, spouse_keycloak_id, marital_status)
		VALUES
			('44000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com', $3, 'Married'),
			('55000000-0000-0000-0000-000000000000', $2, 'User', 'Two', 'user2@email.com', NULL, NULL),
			('66000000-0000-0000-0000-000000000000', $3, 'User', 'Three', 'user3@email.com', $1, 'Married')`,
		keycloakID1, keycloakID2, keycloakID3)
	require.NoError(t, err)

	err = db.SetSpouse(ctx, keycloakID1, keycloakID2, true)
	require.NoError(t, err)

	// Verify new relationship is established and old spouse is unlinked
	actualRows, err := db.Query(ctx, `
		SELECT keycloak_id::text, spouse_keycloak_id, marital_status
		FROM users
		ORDER BY keycloak_id`)
	require.NoError(t, err)
	defer actualRows.Close()

	var actual [][]interface{}
	for actualRows.Next() {
		actualRow, err := actualRows.Values()
		require.NoError(t, err)
		actual = append(actual, actualRow)
	}
	require.Len(t, actual, 3)
	// User 1 should be married to User 2
	assert.Equal(t, keycloakID1.String(), actual[0][0])
	assert.Equal(t, keycloakID2.String(), actual[0][1])
	assert.Equal(t, "Married", actual[0][2])
	// User 2 should be married to User 1
	assert.Equal(t, keycloakID2.String(), actual[1][0])
	assert.Equal(t, keycloakID1.String(), actual[1][1])
	assert.Equal(t, "Married", actual[1][2])
	// User 3 (old spouse) should be unlinked and Divorced
	assert.Equal(t, keycloakID3.String(), actual[2][0])
	assert.Nil(t, actual[2][1])
	assert.Equal(t, "Divorced", actual[2][2])
}

func Test_ProfileDB_SetSpouse_with_force_update_unlinks_user2_previous_spouse(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID1 := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	keycloakID2 := uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")
	keycloakID3 := uuid.FromStringOrNil("33000000-0000-0000-0000-000000000000")
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, spouse_keycloak_id, marital_status)
		VALUES
			('44000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com', NULL, NULL),
			('55000000-0000-0000-0000-000000000000', $2, 'User', 'Two', 'user2@email.com', $3, 'Married'),
			('66000000-0000-0000-0000-000000000000', $3, 'User', 'Three', 'user3@email.com', $2, 'Married')`,
		keycloakID1, keycloakID2, keycloakID3)
	require.NoError(t, err)

	err = db.SetSpouse(ctx, keycloakID1, keycloakID2, true)
	require.NoError(t, err)

	// Verify new relationship is established and old spouse is unlinked
	actualRows, err := db.Query(ctx, `
		SELECT keycloak_id::text, spouse_keycloak_id, marital_status
		FROM users
		ORDER BY keycloak_id`)
	require.NoError(t, err)
	defer actualRows.Close()

	var actual [][]interface{}
	for actualRows.Next() {
		actualRow, err := actualRows.Values()
		require.NoError(t, err)
		actual = append(actual, actualRow)
	}
	require.Len(t, actual, 3)
	// User 1 should be married to User 2
	assert.Equal(t, keycloakID1.String(), actual[0][0])
	assert.Equal(t, keycloakID2.String(), actual[0][1])
	assert.Equal(t, "Married", actual[0][2])
	// User 2 should be married to User 1
	assert.Equal(t, keycloakID2.String(), actual[1][0])
	assert.Equal(t, keycloakID1.String(), actual[1][1])
	assert.Equal(t, "Married", actual[1][2])
	// User 3 (old spouse) should be unlinked and Divorced
	assert.Equal(t, keycloakID3.String(), actual[2][0])
	assert.Nil(t, actual[2][1])
	assert.Equal(t, "Divorced", actual[2][2])
}

func Test_ProfileDB_SetSpouse_with_force_update_unlinks_both_previous_spouses(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	keycloakID1 := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	keycloakID2 := uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")
	keycloakID3 := uuid.FromStringOrNil("33000000-0000-0000-0000-000000000000")
	keycloakID4 := uuid.FromStringOrNil("44000000-0000-0000-0000-000000000000")
	_, err := db.Exec(ctx, `
		INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, spouse_keycloak_id, marital_status)
		VALUES
			('55000000-0000-0000-0000-000000000000', $1, 'User', 'One', 'user1@email.com', $3, 'Married'),
			('66000000-0000-0000-0000-000000000000', $2, 'User', 'Two', 'user2@email.com', $4, 'Married'),
			('77000000-0000-0000-0000-000000000000', $3, 'User', 'Three', 'user3@email.com', $1, 'Married'),
			('88000000-0000-0000-0000-000000000000', $4, 'User', 'Four', 'user4@email.com', $2, 'Married')`,
		keycloakID1, keycloakID2, keycloakID3, keycloakID4)
	require.NoError(t, err)

	err = db.SetSpouse(ctx, keycloakID1, keycloakID2, true)
	require.NoError(t, err)

	// Verify new relationship is established and both old spouses are unlinked
	actualRows, err := db.Query(ctx, `
		SELECT keycloak_id::text, spouse_keycloak_id, marital_status
		FROM users
		ORDER BY keycloak_id`)
	require.NoError(t, err)
	defer actualRows.Close()

	var actual [][]interface{}
	for actualRows.Next() {
		actualRow, err := actualRows.Values()
		require.NoError(t, err)
		actual = append(actual, actualRow)
	}
	require.Len(t, actual, 4)
	// User 1 should be married to User 2
	assert.Equal(t, keycloakID1.String(), actual[0][0])
	assert.Equal(t, keycloakID2.String(), actual[0][1])
	assert.Equal(t, "Married", actual[0][2])
	// User 2 should be married to User 1
	assert.Equal(t, keycloakID2.String(), actual[1][0])
	assert.Equal(t, keycloakID1.String(), actual[1][1])
	assert.Equal(t, "Married", actual[1][2])
	// User 3 (old spouse of User 1) should be unlinked and Divorced
	assert.Equal(t, keycloakID3.String(), actual[2][0])
	assert.Nil(t, actual[2][1])
	assert.Equal(t, "Divorced", actual[2][2])
	// User 4 (old spouse of User 2) should be unlinked and Divorced
	assert.Equal(t, keycloakID4.String(), actual[3][0])
	assert.Nil(t, actual[3][1])
	assert.Equal(t, "Divorced", actual[3][2])
}
