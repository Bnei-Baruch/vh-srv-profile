package repo

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
)

// An APPROVED help-haver request with no matching grant row (legacy requests
// predating the grant table, or a hand-removed grant) must not panic the
// evaluator — the grant-less request is treated as "no help-haver grant".
// Without the guard this panics on the nil Grant.* dereference.
func Test_EvaluateMembership_ApprovedRequestWithoutGrant_NoPanic(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	kc := "55000000-0000-0000-0000-000000000001"
	osMock := &orderServiceMock{}
	db.SetOrdersServiceFactory(func() orders.OrdersService { return osMock })
	osMock.On("GetOrders", mock.Anything, mock.Anything, "globalmembership", true, "desc",
		mock.Anything, mock.Anything).Return([]orders.Order{}, nil)
	osMock.On("GetSpecials", mock.Anything, mock.Anything).Return([]orders.Special{}, nil)
	require.NoError(t, db.CreateProfile(ctx, UserInput{
		KeycloakID:          utils.PointerUUID(uuid.FromStringOrNil(kc)),
		FirstNameVernacular: utils.PointerString("f"),
		LastNameVernacular:  utils.PointerString("l"),
		Emails:              Emails{Primary: utils.PointerString("nilgrant@test.test")},
	}))
	var userID string
	require.NoError(t, db.QueryRow(ctx, `SELECT user_id::text FROM users WHERE keycloak_id=$1`, kc).Scan(&userID))

	// APPROVED help-haver request with NO grant row.
	_, err := db.Exec(ctx,
		`INSERT INTO request (keycloak_id, name, status, type, months, created_at, updated_at)
		 VALUES ($1, 'test', 'APPROVED', 'hhmembership', 6, now(), now())`, kc)
	require.NoError(t, err)

	res, err := db.EvaluateMembershipByUserID(ctx, EmailKeycloakAndUserIDBody{UserID: &userID})
	require.NoError(t, err) // guard prevents the nil-grant panic
	// A grant-less approved request is dropped, so with no orders the user falls
	// to the "new" classification (inactive) rather than yielding a help-haver
	// membership from a nil grant.
	require.NotNil(t, res.Type)
	assert.Equal(t, "new", *res.Type)
	require.NotNil(t, res.Active)
	assert.False(t, *res.Active)
}
