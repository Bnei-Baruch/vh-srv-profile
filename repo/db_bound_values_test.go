package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"

	uuid "github.com/satori/go.uuid"
)

// UpdateProfile binds user_id after the SET args in all three statements
// (users, phone_numbers, status); apostrophes in values must round-trip.
func Test_ProfileDB_UpdateProfile_binds_user_id_and_apostrophes(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	const userID = "22000000-0000-0000-0000-000000000010"
	const kc = "11000000-0000-0000-0000-000000000010"
	insertOperationUser(t, db, userID, kc, "o'user@test.test")
	_, err := db.Exec(ctx, `INSERT INTO status (user_id, membership_type) VALUES ($1, 'inactive')`, userID)
	require.NoError(t, err)

	err = db.UpdateProfile(ctx, uuid.FromStringOrNil(kc), UserInput{
		FirstNameVernacular: utils.PointerString("O'Neil"),
		LastNameLatin:       utils.PointerString("D'Arcy"),
		Phones: Phones{
			MobileNumber:   utils.PointerString("05'1"),
			TelegramNumber: utils.PointerString("@o'tg"),
		},
		Status: UserStatus{MembershipType: utils.PointerString("o'type")},
	})
	require.NoError(t, err)

	var first, last, mType string
	require.NoError(t, db.QueryRow(ctx, `SELECT u.first_name_vernacular, u.last_name_latin, s.membership_type
		FROM users u JOIN status s ON s.user_id = u.user_id WHERE u.user_id=$1`, userID).Scan(&first, &last, &mType))
	assert.Equal(t, "O'Neil", first)
	assert.Equal(t, "D'Arcy", last)
	assert.Equal(t, "o'type", mType)

	rows, err := db.Query(ctx, `SELECT type, phone_number FROM phone_numbers WHERE user_id=$1 ORDER BY type`, userID)
	require.NoError(t, err)
	defer rows.Close()
	phones := map[string]string{}
	for rows.Next() {
		var typ, num string
		require.NoError(t, rows.Scan(&typ, &num))
		phones[typ] = num
	}
	require.NoError(t, rows.Err())
	assert.Equal(t, map[string]string{common.Mobile: "05'1", common.Telegram: "@o'tg"}, phones)
}

// PatchNotification binds the id after the SET args.
func Test_ProfileDB_PatchNotification_binds_id(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	var id int
	require.NoError(t, db.QueryRow(ctx, `INSERT INTO notification (slug) VALUES ('a') RETURNING id`).Scan(&id))
	var other int
	require.NoError(t, db.QueryRow(ctx, `INSERT INTO notification (slug) VALUES ('b') RETURNING id`).Scan(&other))

	got, err := db.PatchNotification(ctx, Notification{Slug: utils.PointerString("it's")}, id)
	require.NoError(t, err)
	assert.Equal(t, id, got)

	var slug string
	require.NoError(t, db.QueryRow(ctx, `SELECT slug FROM notification WHERE id=$1`, id).Scan(&slug))
	assert.Equal(t, "it's", slug)
	require.NoError(t, db.QueryRow(ctx, `SELECT slug FROM notification WHERE id=$1`, other).Scan(&slug))
	assert.Equal(t, "b", slug)
}

// PatchMembershipByID binds the id after the SET args; a missing id is ErrNotFound.
func Test_ProfileDB_PatchMembershipByID_binds_id(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	userID := newProfileForQuery(t, db, "33000000-0000-0000-0000-000000000020", "o'memb@test.test")
	var id int
	require.NoError(t, db.QueryRow(ctx, `SELECT id FROM membership WHERE user_id=$1 LIMIT 1`, userID).Scan(&id))

	got, err := db.PatchMembershipByID(ctx, Membership{Type: utils.PointerString("o'type")}, id)
	require.NoError(t, err)
	assert.Equal(t, id, got)

	var typ string
	require.NoError(t, db.QueryRow(ctx, `SELECT type FROM membership WHERE id=$1`, id).Scan(&typ))
	assert.Equal(t, "o'type", typ)

	_, err = db.PatchMembershipByID(ctx, Membership{Type: utils.PointerString("x")}, 999999)
	assert.ErrorIs(t, err, common.ErrNotFound)
}

// updateRequest binds the id after the SET args.
func Test_ProfileDB_updateRequest_binds_id(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	kc := "33000000-0000-0000-0000-000000000021"
	newProfileForQuery(t, db, kc, "o'req@test.test")
	var reqID int
	require.NoError(t, db.QueryRow(ctx,
		`INSERT INTO request (keycloak_id, name, status, type, months, created_at, updated_at)
		 VALUES ($1, 'test', 'PENDING', 'hhmembership', 6, now(), now()) RETURNING id`, kc).Scan(&reqID))

	require.NoError(t, db.updateRequest(ctx, NewRequest{
		ID:            &reqID,
		Status:        utils.PointerString(common.RequestStatusDenied),
		RejectionNote: utils.PointerString("doesn't qualify"),
	}))

	var status, note string
	require.NoError(t, db.QueryRow(ctx, `SELECT status, rejection_note FROM request WHERE id=$1`, reqID).Scan(&status, &note))
	assert.Equal(t, common.RequestStatusDenied, status)
	assert.Equal(t, "doesn't qualify", note)
}

// Apostrophes in request filters are matched as values.
func Test_GetMultipleRequest_filters_with_apostrophe(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	kc := "33000000-0000-0000-0000-000000000022"
	newProfileForQuery(t, db, kc, "o'brien@test.test")
	_, err := db.Exec(ctx,
		`INSERT INTO request (keycloak_id, name, status, type, months, created_at, updated_at)
		 VALUES ($1, 'O''Neil', 'PENDING', 'hhmembership', 6, now(), now())`, kc)
	require.NoError(t, err)

	res, err := db.GetMultipleRequest(ctx, 0, 10, "", "", "o'neil", "o'brien", "", "desc")
	require.NoError(t, err)
	require.Len(t, res, 1)
	assert.Equal(t, "O'Neil", *res[0].Request.RequestName)

	count, err := db.GetMultipleRequestCount(ctx, "", "", "o'neil", "o'brien", "", "")
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
