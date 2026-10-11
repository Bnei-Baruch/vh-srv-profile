package repo

import (
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
)

const (
	specialUserKC    = "61000000-0000-0000-0000-000000000000"
	specialUserEmail = "kconly@test.test"
)

// newSpecialLookupUser creates a profile whose specials orders returns only by
// keycloak_id: the shape of a special created from a user page in the back
// office, which has no email.
func newSpecialLookupUser(t *testing.T, db *ProfileDB, byKeycloakID []orders.Special) {
	t.Helper()
	osMock := orderServiceMock{}
	db.SetOrdersServiceFactory(func() orders.OrdersService { return &osMock })
	osMock.On("GetOrders", mock.Anything, mock.Anything, "globalmembership", true, "desc",
		mock.Anything, mock.Anything).Return([]orders.Order{}, nil)
	osMock.On("GetSpecials", mock.Anything, mock.Anything).Return([]orders.Special{}, nil)
	osMock.On("GetSpecialsByKeycloakID", mock.Anything, specialUserKC).Return(byKeycloakID, nil)
	osMock.On("GetSpecialsByKeycloakID", mock.Anything, mock.Anything).Return([]orders.Special{}, nil)

	require.NoError(t, db.CreateProfile(testContext(), UserInput{
		KeycloakID:          utils.PointerUUID(uuid.FromStringOrNil(specialUserKC)),
		FirstNameVernacular: utils.PointerString("first"),
		LastNameVernacular:  utils.PointerString("last"),
		Emails:              Emails{Primary: utils.PointerString(specialUserEmail)},
	}))
}

// Evaluation used to look specials up by email only, so a special without
// one never counted.
func TestEvaluate_FindsSpecialByKeycloakID(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	newSpecialLookupUser(t, db, []orders.Special{{
		Id:         1,
		KeycloakId: specialUserKC,
		StartDate:  time.Now().Add(-time.Hour),
		EndDate:    time.Now().Add(24 * time.Hour),
		Category:   "hh",
	}})

	res, err := db.EvaluateMembershipByUserID(testContext(), EmailKeycloakAndUserIDBody{KeycloakID: utils.PointerString(specialUserKC)})
	require.NoError(t, err)
	require.NotNil(t, res.Type)
	assert.Equal(t, "special", *res.Type)
	require.NotNil(t, res.Active)
	assert.True(t, *res.Active)
}

// An event's keycloak_id may belong to no user (a merged-away account); the
// email it carries still finds the user.
func TestEvaluate_UnknownKeycloakIDFallsBackToEmail(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	newSpecialLookupUser(t, db, []orders.Special{})

	_, err := db.EvaluateMembershipByUserID(testContext(), EmailKeycloakAndUserIDBody{
		KeycloakID: utils.PointerString("62000000-0000-0000-0000-000000000000"),
		Email:      utils.PointerString(specialUserEmail),
	})
	require.NoError(t, err)

	_, err = db.EvaluateMembershipByUserID(testContext(), EmailKeycloakAndUserIDBody{
		KeycloakID: utils.PointerString("62000000-0000-0000-0000-000000000000"),
	})
	assert.Error(t, err, "without an email there is nothing to fall back to")
}
