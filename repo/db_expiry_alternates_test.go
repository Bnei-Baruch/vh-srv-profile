package repo

import (
	"database/sql"
	"fmt"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
)

// stubOrdersService makes every evaluation find no orders and no specials, so
// the evaluated type has no fixed expiry.
func stubOrdersService(db *ProfileDB) {
	osMock := orderServiceMock{}
	db.SetOrdersServiceFactory(func() orders.OrdersService { return &osMock })
	osMock.On("GetOrders", mock.Anything, mock.Anything, "globalmembership", true, "desc",
		mock.Anything, mock.Anything).Return([]orders.Order{}, nil)
	osMock.On("GetSpecial", mock.Anything, mock.Anything).Return(nil, nil)
	osMock.On("GetSpecials", mock.Anything, mock.Anything).Return([]orders.Special{}, nil)
}

func createTestProfile(t *testing.T, db *ProfileDB, n int, email string) (keycloakID, userID string) {
	t.Helper()
	keycloakID = fmt.Sprintf("%02d000000-0000-0000-0000-000000000000", n)
	require.NoError(t, db.CreateProfile(testContext(), UserInput{
		KeycloakID:          utils.PointerUUID(uuid.FromStringOrNil(keycloakID)),
		FirstNameVernacular: utils.PointerString("first"),
		LastNameVernacular:  utils.PointerString("last"),
		Emails:              Emails{Primary: utils.PointerString(email)},
	}))
	require.NoError(t, db.QueryRow(testContext(), `SELECT user_id::text FROM users WHERE keycloak_id = $1`, keycloakID).Scan(&userID))
	return keycloakID, userID
}

// leaveStaleExpiry gives the user a lapsed manual membership: active, with a
// past expiry, which the invalidator selects. Re-evaluated with no grants or
// orders, the user has a type without a fixed expiry.
func leaveStaleExpiry(t *testing.T, db *ProfileDB, userID string) {
	t.Helper()
	_, err := db.Exec(testContext(),
		`UPDATE membership SET type = 'manual', active = true, expiry = '2021-10-23' WHERE user_id = $1`, userID)
	require.NoError(t, err)
}

func membershipExpiry(t *testing.T, db *ProfileDB, userID string) sql.NullTime {
	t.Helper()
	var expiry sql.NullTime
	require.NoError(t, db.QueryRow(testContext(), `SELECT expiry FROM membership WHERE user_id = $1`, userID).Scan(&expiry))
	return expiry
}

// Re-evaluating into a type without a fixed expiry used to leave the previous
// type's date in place; the invalidator then selected the row on every run.
func TestEvaluate_TypeWithoutExpiryClearsTheOldDate(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	stubOrdersService(db)
	_, userID := createTestProfile(t, db, 11, "stale@test.test")
	leaveStaleExpiry(t, db, userID)

	res, err := db.EvaluateMembershipByUserID(testContext(), EmailKeycloakAndUserIDBody{UserID: &userID})
	require.NoError(t, err)
	assert.Nil(t, res.Expiry)
	assert.False(t, membershipExpiry(t, db, userID).Valid, "the stale expiry is cleared")

	expired, err := db.GetExpiredMemberships(testContext(), 0, 100)
	require.NoError(t, err)
	assert.Empty(t, expired, "the invalidator no longer selects the row")
}

// primary_email is shared by design. An email-only evaluation used to pick
// one arbitrary user with that email; every one of them is evaluated now.
func TestEvaluate_ByEmailEvaluatesEveryUserWithIt(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	stubOrdersService(db)
	_, user1 := createTestProfile(t, db, 12, "shared@test.test")
	_, user2 := createTestProfile(t, db, 13, "shared@test.test")
	leaveStaleExpiry(t, db, user1)
	leaveStaleExpiry(t, db, user2)

	_, err := db.EvaluateMembershipByUserID(testContext(), EmailKeycloakAndUserIDBody{Email: utils.PointerString("shared@test.test")})
	require.NoError(t, err)
	assert.False(t, membershipExpiry(t, db, user1).Valid, "first user evaluated")
	assert.False(t, membershipExpiry(t, db, user2).Valid, "second user evaluated")

	_, err = db.EvaluateMembershipByUserID(testContext(), EmailKeycloakAndUserIDBody{Email: utils.PointerString("nobody@test.test")})
	assert.Error(t, err)
}

// Merging keeps the source's email as the destination's first free alternate,
// and never overwrites one that is taken. The old condition was inverted: it
// dropped the email when the alternate was free and overwrote it when not.
func TestMergeAccounts_SourceEmailGoesToAFreeAlternate(t *testing.T) {
	cases := []struct {
		name               string
		alt1, alt2         any
		wantAlt1, wantAlt2 sql.NullString
	}{
		{"both free", nil, nil, sql.NullString{String: "src@test.test", Valid: true}, sql.NullString{}},
		{"first taken", "old@test.test", nil, sql.NullString{String: "old@test.test", Valid: true}, sql.NullString{String: "src@test.test", Valid: true}},
		{"both taken", "old1@test.test", "old2@test.test", sql.NullString{String: "old1@test.test", Valid: true}, sql.NullString{String: "old2@test.test", Valid: true}},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := newTestProfileDBIsolated(t)
			stubOrdersService(db)
			src, _ := createTestProfile(t, db, 20+2*i, "src@test.test")
			dst, _ := createTestProfile(t, db, 21+2*i, "dst@test.test")
			_, err := db.Exec(testContext(), `UPDATE users SET alternate_email_1 = $1, alternate_email_2 = $2 WHERE keycloak_id = $3`, tc.alt1, tc.alt2, dst)
			require.NoError(t, err)

			_, err = db.MergeAccounts(testContext(), AccountsMergeData{SourceId: src, DestinationId: dst})
			require.NoError(t, err)

			var alt1, alt2 sql.NullString
			require.NoError(t, db.QueryRow(testContext(), `SELECT alternate_email_1, alternate_email_2 FROM users WHERE keycloak_id = $1`, dst).Scan(&alt1, &alt2))
			assert.Equal(t, tc.wantAlt1, alt1)
			assert.Equal(t, tc.wantAlt2, alt2)
		})
	}
}

// An empty alternate in an update clears it to NULL instead of storing ”.
func TestUpdateProfile_EmptyAlternateIsNull(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	stubOrdersService(db)
	kc, _ := createTestProfile(t, db, 40, "alt@test.test")

	require.NoError(t, db.UpdateProfile(testContext(), uuid.FromStringOrNil(kc), UserInput{
		Emails: Emails{Alternate1: utils.PointerString("a1@test.test"), Alternate2: utils.PointerString("a2@test.test")},
	}))
	require.NoError(t, db.UpdateProfile(testContext(), uuid.FromStringOrNil(kc), UserInput{
		Emails: Emails{Alternate1: utils.PointerString("")},
	}))

	var alt1, alt2 sql.NullString
	require.NoError(t, db.QueryRow(testContext(), `SELECT alternate_email_1, alternate_email_2 FROM users WHERE keycloak_id = $1`, kc).Scan(&alt1, &alt2))
	assert.False(t, alt1.Valid, "cleared to NULL")
	assert.Equal(t, "a2@test.test", alt2.String, "untouched")
}

// CreateProfile stores an empty alternate as NULL as well.
func TestCreateProfile_EmptyAlternateIsNull(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	stubOrdersService(db)
	kc := "41000000-0000-0000-0000-000000000000"
	require.NoError(t, db.CreateProfile(testContext(), UserInput{
		KeycloakID:          utils.PointerUUID(uuid.FromStringOrNil(kc)),
		FirstNameVernacular: utils.PointerString("first"),
		LastNameVernacular:  utils.PointerString("last"),
		Emails:              Emails{Primary: utils.PointerString("c@test.test"), Alternate1: utils.PointerString("")},
	}))

	var alt1 sql.NullString
	require.NoError(t, db.QueryRow(testContext(), `SELECT alternate_email_1 FROM users WHERE keycloak_id = $1`, kc).Scan(&alt1))
	assert.False(t, alt1.Valid)
}
