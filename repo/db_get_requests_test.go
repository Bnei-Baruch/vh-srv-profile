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

// Caller input must be bound as positional args, never interpolated into the SQL.
func Test_buildAndGetWhereRequestQuery_Parameterized(t *testing.T) {
	inj := "x'; DROP TABLE users;--"
	where, orderBy, args := buildAndGetWhereRequestQuery(inj, "APPROVED", "smith", "hhmembership", "asc", "a@b.com")

	assert.NotContains(t, where, inj, "raw input must not be interpolated into the query")
	assert.Contains(t, where, "r.keycloak_id=$1")
	assert.Contains(t, where, "r.status=$2")
	assert.Contains(t, where, "lower(r.name) like lower($3)")
	assert.Contains(t, where, "r.type=$4")
	assert.Contains(t, where, "lower($5)") // email uses one placeholder across the OR
	assert.Equal(t, []interface{}{inj, "APPROVED", "%smith%", "hhmembership", "%a@b.com%"}, args)
	assert.Equal(t, " ORDER BY r.created_at asc", orderBy)

	// Order direction can't be injected: anything but desc collapses to asc.
	_, ob, _ := buildAndGetWhereRequestQuery("", "", "", "", "created_at; DROP TABLE users", "")
	assert.Equal(t, " ORDER BY r.created_at asc", ob)

	// No filters → no WHERE, no args, default order.
	w, ob2, a := buildAndGetWhereRequestQuery("", "", "", "", "", "")
	assert.Empty(t, w)
	assert.Empty(t, a)
	assert.Equal(t, " ORDER BY r.updated_at desc", ob2)
}

// The fixed SELECT must run and return the request joined with its grant.
func Test_GetMultipleRequest_ReturnsRequestWithGrant(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	kc := "33000000-0000-0000-0000-000000000009"

	// A request FK-references an existing user; create the profile first.
	osMock := &orderServiceMock{}
	db.SetOrdersServiceFactory(func() orders.OrdersService { return osMock })
	osMock.On("GetOrders", mock.Anything, mock.Anything, "globalmembership", true, "desc",
		mock.Anything, mock.Anything).Return([]orders.Order{}, nil)
	osMock.On("GetSpecial", mock.Anything, mock.Anything).Return(nil, nil)
	osMock.On("GetSpecials", mock.Anything, mock.Anything).Return([]orders.Special{}, nil)
	require.NoError(t, db.CreateProfile(ctx, UserInput{
		KeycloakID:          utils.PointerUUID(uuid.FromStringOrNil(kc)),
		FirstNameVernacular: utils.PointerString("f"),
		LastNameVernacular:  utils.PointerString("l"),
		Emails:              Emails{Primary: utils.PointerString("req@test.test")},
	}))
	var userID string
	require.NoError(t, db.QueryRow(ctx, `SELECT user_id::text FROM users WHERE keycloak_id=$1`, kc).Scan(&userID))

	var reqID int
	require.NoError(t, db.QueryRow(ctx,
		`INSERT INTO request (keycloak_id, name, status, type, months, created_at, updated_at)
		 VALUES ($1, 'test', 'APPROVED', 'hhmembership', 6, now(), now()) RETURNING id`, kc).Scan(&reqID))
	_, err := db.Exec(ctx,
		`INSERT INTO "grant" (request_id, user_id, type, properties, created_at, updated_at)
		 VALUES ($1, $2, 'mb_months', '{"months":6}', now(), now())`, reqID, userID)
	require.NoError(t, err)

	res, err := db.GetMultipleRequest(ctx, 0, 10, kc, "APPROVED", "", "", "hhmembership", "desc")
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.NotNil(t, res[0].Request.KeycloakId)
	assert.Equal(t, kc, *res[0].Request.KeycloakId)
	require.NotNil(t, res[0].Grant.RequestID)
	assert.Equal(t, reqID, *res[0].Grant.RequestID, "grant must join to its request")

	// Injection attempt through a filter returns no rows (bound, not executed).
	res, err = db.GetMultipleRequest(ctx, 0, 10, "' OR '1'='1", "", "", "", "", "desc")
	require.NoError(t, err)
	assert.Empty(t, res)
}
