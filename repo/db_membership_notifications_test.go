package repo

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
)

// newUserWithActiveNotification creates a profile and activates the given HH
// notification for it, returning the user_id.
func newUserWithActiveNotification(t *testing.T, db *ProfileDB, slug string) string {
	t.Helper()
	ctx := testContext()

	osMock := orderServiceMock{}
	db.SetOrdersServiceFactory(func() orders.OrdersService { return &osMock })
	osMock.On("GetOrders", mock.Anything, mock.Anything, "globalmembership", true, "desc",
		mock.Anything, mock.Anything).Return([]orders.Order{}, nil)
	osMock.On("GetSpecial", mock.Anything, mock.Anything).Return(nil, nil)
	osMock.On("GetSpecials", mock.Anything, mock.Anything).Return([]orders.Special{}, nil)

	require.NoError(t, db.CreateProfile(ctx, UserInput{
		KeycloakID:          utils.PointerUUID(uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")),
		FirstNameVernacular: utils.PointerString("first name"),
		LastNameVernacular:  utils.PointerString("last name"),
		Emails:              Emails{Primary: utils.PointerString("notif@test.test")},
	}))

	var userID string
	require.NoError(t, db.QueryRow(ctx, `SELECT user_id::text FROM users`).Scan(&userID))

	require.NoError(t, db.CreateUserNotification(ctx, UserNotification{
		UserID:         &userID,
		NotificationID: NotificationsRegistry.BySlug[slug].ID,
		Active:         utils.PointerBool(true),
	}))
	return userID
}

func activeSlugs(t *testing.T, db *ProfileDB, userID string) []string {
	t.Helper()
	notis, err := db.GetActiveUserNotificationByUserID(testContext(), userID)
	require.NoError(t, err)
	slugs := make([]string, 0)
	for _, n := range notis {
		slugs = append(slugs, *n.Slug)
	}
	return slugs
}

// A valid membership makes a past HH refusal irrelevant — the sticky
// hh_request_refused banner must be cleared.
func Test_applyMembershipNotifications_ActiveMembershipClearsRefused(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	userID := newUserWithActiveNotification(t, db, common.NotificationSlugHHRequestRefused)

	require.NoError(t, db.applyMembershipNotifications(testContext(), userID, nil, true))

	assert.Empty(t, activeSlugs(t, db, userID))
}

// Without a valid membership the refusal sticks (until a new HH request or a payment).
func Test_applyMembershipNotifications_InactiveMembershipKeepsRefused(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	userID := newUserWithActiveNotification(t, db, common.NotificationSlugHHRequestRefused)

	require.NoError(t, db.applyMembershipNotifications(testContext(), userID,
		[]string{common.NotificationSlugMBNew}, false))

	assert.ElementsMatch(t,
		[]string{common.NotificationSlugMBNew, common.NotificationSlugHHRequestRefused},
		activeSlugs(t, db, userID))
}

// hh_request_approved coexists with a valid membership by design — never cleared here.
func Test_applyMembershipNotifications_ActiveMembershipKeepsApproved(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	userID := newUserWithActiveNotification(t, db, common.NotificationSlugHHRequestApproved)

	require.NoError(t, db.applyMembershipNotifications(testContext(), userID, nil, true))

	assert.Equal(t, []string{common.NotificationSlugHHRequestApproved}, activeSlugs(t, db, userID))
}
