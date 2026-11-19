package repo

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v9"

	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"

	uuid "github.com/satori/go.uuid"
)

// The INSERT/UPDATE helpers write the fields that are set and refuse an
// empty record, as before.
func Test_ProfileDB_insert_and_patch_helpers(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	kc := "33000000-0000-0000-0000-000000000030"
	userIDStr := newProfileForQuery(t, db, kc, "o'insert@test.test")
	userID := uuid.FromStringOrNil(userIDStr)

	membershipID, err := db.createMembership(ctx, Membership{
		Active: utils.PointerBool(true),
		Type:   utils.PointerString("o'type"),
		Expiry: utils.PointerTime(time.Now().Add(time.Hour)),
		UserID: &userID,
	})
	require.NoError(t, err)

	var reqID int
	require.NoError(t, db.QueryRow(ctx,
		`INSERT INTO request (keycloak_id, name, status, type, months, created_at, updated_at)
		 VALUES ($1, 'test', 'APPROVED', 'hhmembership', 6, now(), now()) RETURNING id`, kc).Scan(&reqID))
	props, _ := json.Marshal(map[string]any{"months": 6})
	grantID, err := db.createGrant(ctx, Grant{
		UserID:     &userID,
		RequestID:  &reqID,
		Type:       utils.PointerString("mb_months"),
		Properties: null.JSONFrom(props),
	})
	require.NoError(t, err)
	var grantType, grantProps string
	require.NoError(t, db.QueryRow(ctx, `SELECT type, properties::text FROM "grant" WHERE id=$1`, grantID).
		Scan(&grantType, &grantProps))
	assert.Equal(t, "mb_months", grantType)
	assert.JSONEq(t, `{"months":6}`, grantProps)

	one := utils.PointerInt(1)
	_, err = db.createManualMembership(ctx, MembershipManual{OrderID: one, PaymentID: one, MembershipID: &membershipID, Quantity: one})
	require.NoError(t, err)
	_, err = db.createAutomaticMembership(ctx, MembershipAutomatic{OrderID: one, PaymentID: one, MembershipID: &membershipID})
	require.NoError(t, err)
	_, err = db.createSpecialMembership(ctx, MembershipSpecial{MembershipID: &membershipID})
	require.NoError(t, err)
	_, err = db.createHelpHaverMembership(ctx, MembershipHelpHaver{MembershipID: &membershipID, GrantID: &grantID})
	require.NoError(t, err)

	for _, table := range []string{"membership_manual", "membership_automatic", "membership_special", "membership_helphaver"} {
		var n int
		require.NoError(t, db.QueryRow(ctx, `SELECT count(*) FROM `+table+` WHERE membership_id=$1`, membershipID).Scan(&n))
		assert.Equal(t, 1, n, table)
	}

	// Nothing set: same errors as before.
	_, err = db.createMembership(ctx, Membership{})
	assert.EqualError(t, err, "invalid values")
	_, err = db.createGrant(ctx, Grant{})
	assert.EqualError(t, err, "invalid values")
	_, err = db.createSpecialMembership(ctx, MembershipSpecial{})
	assert.EqualError(t, err, "invalid values")
	_, err = db.PatchMembershipByID(ctx, Membership{}, membershipID)
	assert.EqualError(t, err, "no fields to update")
	_, err = db.PatchNotification(ctx, Notification{}, 1)
	assert.EqualError(t, err, "invalid values")
	assert.EqualError(t, db.updateRequest(ctx, NewRequest{ID: &reqID}), "invalid values")
	assert.EqualError(t, db.CreateUserNotification(ctx, UserNotification{}), "invalid values")
}

func Test_ProfileDB_PatchUserNotification(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	userID := newProfileForQuery(t, db, "33000000-0000-0000-0000-000000000031", "un@test.test")
	var notiID int
	require.NoError(t, db.QueryRow(ctx, `INSERT INTO notification (slug) VALUES ('x') RETURNING id`).Scan(&notiID))
	require.NoError(t, db.CreateUserNotification(ctx, UserNotification{
		UserID: &userID, NotificationID: &notiID, Active: utils.PointerBool(true),
	}))
	var id int
	require.NoError(t, db.QueryRow(ctx,
		`SELECT id FROM user_notification WHERE user_id=$1 AND notification_id=$2`, userID, notiID).Scan(&id))

	got, err := db.PatchUserNotification(ctx, UserNotification{Active: utils.PointerBool(false)}, id)
	require.NoError(t, err)
	assert.Equal(t, id, got)
	var active bool
	require.NoError(t, db.QueryRow(ctx, `SELECT active FROM user_notification WHERE id=$1`, id).Scan(&active))
	assert.False(t, active)

	// Nothing set is a no-op, not an error.
	got, err = db.PatchUserNotification(ctx, UserNotification{}, id)
	require.NoError(t, err)
	assert.Zero(t, got)
}
