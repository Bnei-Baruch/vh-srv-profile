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

const e2eKcID = "11000000-0000-0000-0000-000000000000"

// seedApprovedV1Grant inserts an APPROVED help-haver request + its grant for the
// user, so EvaluateMembershipByUserID finds it as lastApprovedRequest. grantCreatedAt
// controls whether a later payment supersedes (cancels) the grant.
func seedApprovedV1Grant(t *testing.T, db *ProfileDB, userID string, grantCreatedAt time.Time) {
	t.Helper()
	ctx := testContext()
	var requestID int
	require.NoError(t, db.QueryRow(ctx, `
		INSERT INTO request (name, keycloak_id, status, type, months, created_at, updated_at)
		VALUES ('hh', $1, $2, $3, 3, $4, $4) RETURNING id`,
		e2eKcID, common.RequestStatusApproved, common.RequestTypeHelpHaver, grantCreatedAt).Scan(&requestID))
	_, err := db.Exec(ctx, `
		INSERT INTO "grant" (user_id, request_id, type, properties, created_at, updated_at)
		VALUES ($1, $2, $3, '{"months": 3}', $4, $4)`,
		userID, requestID, common.GrantTypeMembershipMonths, grantCreatedAt)
	require.NoError(t, err)
}

// mockPaidManualOrder wires a single paid regular order → manual membership.
func mockPaidManualOrder(os *orderServiceMock, paymentDate time.Time) {
	order := orders.Order{ID: 111, Type: "regular", Status: "paid", Quantity: 12,
		PaymentDate: paymentDate, StartingDate: paymentDate}
	os.On("GetOrders", mock.Anything, mock.Anything, "globalmembership", true, "desc",
		mock.Anything, mock.Anything).Return([]orders.Order{order}, nil)
	os.On("GetOrderPayments", mock.Anything, 111, "desc", 1, 0).
		Return([]orders.Payment{{ID: 222, Status: "success"}}, nil)
	os.On("GetPaymentByID", mock.Anything, 222).
		Return(&orders.Payment{ID: 222, Status: "success"}, nil)
}

// V1 stale case (the original bug): the member had an approved V1 grant, then paid a
// normal membership. The payment postdates the grant, so eval cancels the grant and
// the "approved" banner is now stale — it must be cleared.
func Test_EvaluateMembership_ClearsStaleApprovedForCancelledV1Grant(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	userID := newUserWithActiveNotification(t, db, common.NotificationSlugHHRequestApproved)
	seedApprovedV1Grant(t, db, userID, time.Now().AddDate(0, 0, -60)) // grant is old

	os := &orderServiceMock{}
	db.SetOrdersServiceFactory(func() orders.OrdersService { return os })
	mockPaidManualOrder(os, time.Now()) // payment postdates the grant → grant cancelled

	_, err := db.EvaluateMembershipByUserID(ctx, EmailKeycloakAndUserIDBody{UserID: &userID})
	require.NoError(t, err)

	assert.NotContains(t, activeSlugs(t, db, userID), common.NotificationSlugHHRequestApproved,
		"a superseded/cancelled V1 grant must clear the stale approved banner")
}

// V2 / no-profile-grant case (Grisha's high-severity regression): a help-haver whose
// approval lives only in orders has no profile grant here, so lastApprovedRequest is
// nil. That banner is a discount notice, not a membership signal — re-evaluation on a
// plain paid membership must NOT clear it.
func Test_EvaluateMembership_KeepsApprovedForV2HelpHaver(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	userID := newUserWithActiveNotification(t, db, common.NotificationSlugHHRequestApproved)
	// no seedApprovedV1Grant — a pure V2 help-haver has no profile request/grant.

	os := &orderServiceMock{}
	db.SetOrdersServiceFactory(func() orders.OrdersService { return os })
	mockPaidManualOrder(os, time.Now())

	_, err := db.EvaluateMembershipByUserID(ctx, EmailKeycloakAndUserIDBody{UserID: &userID})
	require.NoError(t, err)

	assert.Contains(t, activeSlugs(t, db, userID), common.NotificationSlugHHRequestApproved,
		"a V2 help-haver's approved banner must survive re-evaluation")
}
