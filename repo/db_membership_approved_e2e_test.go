package repo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
)

// End-to-end guard for the stale "hh_request_approved" banner. A user who once
// had a help-haver grant (notice active) but now holds a plain paid membership
// must NOT keep the approved banner after re-evaluation.
//
// This drives the public EvaluateMembershipByUserID so it compiles on both the
// pre-fix and post-fix trees:
//   - RED before the fix  — eval never clears hh_request_approved for non-helphaver.
//   - GREEN after the fix — applyMembershipNotifications clears it unless helphaver.
func Test_EvaluateMembership_ClearsStaleApprovedForNonHelphaver(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	// user starts with an active "HH approved" notice (a legacy grant).
	userID := newUserWithActiveNotification(t, db, common.NotificationSlugHHRequestApproved)

	// now the user holds a plain paid (manual) membership — no help-haver grant.
	os := &orderServiceMock{}
	db.SetOrdersServiceFactory(func() orders.OrdersService { return os })
	now := time.Now()
	order := orders.Order{ID: 111, Type: "regular", Status: "paid", Quantity: 12,
		PaymentDate: now, StartingDate: now}
	os.On("GetOrders", mock.Anything, mock.Anything, "globalmembership", true, "desc",
		mock.Anything, mock.Anything).Return([]orders.Order{order}, nil)
	os.On("GetOrderPayments", mock.Anything, 111, "desc", 1, 0).
		Return([]orders.Payment{{ID: 222, Status: "success"}}, nil)
	os.On("GetPaymentByID", mock.Anything, 222).
		Return(&orders.Payment{ID: 222, Status: "success"}, nil)

	_, err := db.EvaluateMembershipByUserID(ctx, EmailKeycloakAndUserIDBody{UserID: &userID})
	require.NoError(t, err)

	assert.NotContains(t, activeSlugs(t, db, userID), common.NotificationSlugHHRequestApproved,
		"stale hh_request_approved must be cleared for a non-help-haver membership")
}
