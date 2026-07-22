package repo

import (
	"fmt"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
)

// The keycloak_id filter accepts a comma-separated list — admin tables use it
// to fetch profile briefs in bulk. Single-id form must keep working.
func Test_GetMultipleProfiles_KeycloakIDList(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	osMock := orderServiceMock{}
	db.SetOrdersServiceFactory(func() orders.OrdersService { return &osMock })
	osMock.On("GetOrders", mock.Anything, mock.Anything, "globalmembership", true, "desc",
		mock.Anything, mock.Anything).Return([]orders.Order{}, nil)
	osMock.On("GetSpecial", mock.Anything, mock.Anything).Return(nil, nil)
	osMock.On("GetSpecials", mock.Anything, mock.Anything).Return([]orders.Special{}, nil)

	ids := []string{
		"11000000-0000-0000-0000-000000000001",
		"11000000-0000-0000-0000-000000000002",
	}
	for i, id := range ids {
		require.NoError(t, db.CreateProfile(ctx, UserInput{
			KeycloakID:          utils.PointerUUID(uuid.FromStringOrNil(id)),
			FirstNameVernacular: utils.PointerString("first"),
			LastNameVernacular:  utils.PointerString("last"),
			Emails:              Emails{Primary: utils.PointerString(fmt.Sprintf("u%d@test.test", i))},
		}))
	}

	// List form: both known profiles returned, the unknown id simply matches nothing.
	profiles, err := db.GetMultipleProfiles(ctx, 0, 10, "", "", "",
		ids[0]+","+ids[1]+",99000000-0000-0000-0000-000000000099",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", false, AND_CLAUSE)
	require.NoError(t, err)
	assert.Len(t, profiles, 2)

	// Single-id form unchanged.
	profiles, err = db.GetMultipleProfiles(ctx, 0, 10, "", "", "", ids[0],
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", false, AND_CLAUSE)
	require.NoError(t, err)
	assert.Len(t, profiles, 1)
}
