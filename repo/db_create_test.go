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

func Test_ProfileDB_createUser_with_minimum_info_succeeds(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	osMock := orderServiceMock{}
	db.SetOrdersServiceFactory(func() orders.OrdersService { return &osMock })
	osMock.On("GetOrders", mock.Anything, "someemail@email.email", "globalmembership", true, "desc",
		mock.Anything, mock.Anything).Return([]orders.Order{}, nil)
	osMock.On("GetSpecial", mock.Anything, "someemail@email.email").Return(nil, nil)
	osMock.On("GetSpecials", mock.Anything, "someemail@email.email").Return([]orders.Special{}, nil)

	err := db.CreateProfile(ctx, UserInput{
		KeycloakID:          utils.PointerUUID(uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")),
		FirstNameVernacular: utils.PointerString("first name"),
		LastNameVernacular:  utils.PointerString("last name"),
		Emails:              Emails{Primary: utils.PointerString("someemail@email.email")},
	})

	assert.NoError(t, err)
	actualRows, err := db.Query(ctx, `SELECT deleted, keycloak_id::text, first_name_vernacular, last_name_vernacular,
	primary_email FROM users`)
	require.NoError(t, err)
	defer actualRows.Close()
	var actual [][]interface{}
	for actualRows.Next() {
		actualRow, err := actualRows.Values()
		require.NoError(t, err)

		actual = append(actual, actualRow)
	}
	require.Len(t, actual, 1)
	assert.Equal(t, [][]interface{}{{false, "11000000-0000-0000-0000-000000000000", "first name", "last name", "someemail@email.email"}}, actual)

	var createdAt, updatedAt time.Time
	require.NoError(t, db.QueryRow(ctx, `SELECT created_at, updated_at FROM users`).Scan(&createdAt, &updatedAt))
	assert.WithinDuration(t, time.Now(), createdAt, 3*time.Second)
	assert.WithinDuration(t, time.Now(), updatedAt, 3*time.Second)
}

func Test_ProfileDB_createUser_with_language_info_succeeds(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	osMock := orderServiceMock{}
	db.SetOrdersServiceFactory(func() orders.OrdersService { return &osMock })
	osMock.On("GetOrders", mock.Anything, "someemail@email.email", "globalmembership", true, "desc",
		mock.Anything, mock.Anything).Return([]orders.Order{}, nil)
	osMock.On("GetSpecial", mock.Anything, "someemail@email.email").Return(nil, nil)
	osMock.On("GetSpecials", mock.Anything, "someemail@email.email").Return([]orders.Special{}, nil)

	err := db.CreateProfile(ctx, UserInput{
		KeycloakID:          utils.PointerUUID(uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")),
		FirstNameVernacular: utils.PointerString("first name"),
		LastNameVernacular:  utils.PointerString("last name"),
		Emails:              Emails{Primary: utils.PointerString("someemail@email.email")},
		Languages: Languages{
			First:     utils.PointerString("he"),
			Listening: utils.PointerString("ch"),
			Reading:   utils.PointerString("en"),
			Email:     utils.PointerString("fr"),
		},
	})

	assert.NoError(t, err)
	actualRows, err := db.Query(ctx, `
	SELECT deleted,
       keycloak_id::text,
       first_name_vernacular,
       last_name_vernacular,
       primary_email,
       first_language,
       listening_language,
       reading_language,
       email_language
	FROM users`)
	require.NoError(t, err)
	defer actualRows.Close()
	var actual [][]interface{}
	for actualRows.Next() {
		actualRow, err := actualRows.Values()
		require.NoError(t, err)

		actual = append(actual, actualRow)
	}
	require.Len(t, actual, 1)
	assert.Equal(t, [][]interface{}{{
		false,
		"11000000-0000-0000-0000-000000000000",
		"first name",
		"last name",
		"someemail@email.email",
		"he",
		"ch",
		"en",
		"fr",
	}}, actual)

	var createdAt, updatedAt time.Time
	require.NoError(t, db.QueryRow(ctx, `SELECT created_at, updated_at FROM users`).Scan(&createdAt, &updatedAt))
	assert.WithinDuration(t, time.Now(), createdAt, 3*time.Second)
	assert.WithinDuration(t, time.Now(), updatedAt, 3*time.Second)
}

func Test_ProfileDB_createUser_with_phone_number_succeeds(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	osMock := orderServiceMock{}
	db.SetOrdersServiceFactory(func() orders.OrdersService { return &osMock })
	osMock.On("GetOrders", mock.Anything, "someemail@email.email", "globalmembership", true, "desc",
		mock.Anything, mock.Anything).Return([]orders.Order{}, nil)
	osMock.On("GetSpecial", mock.Anything, "someemail@email.email").Return(nil, nil)
	osMock.On("GetSpecials", mock.Anything, "someemail@email.email").Return([]orders.Special{}, nil)

	mobile := "0100000000"
	whatsApp := "0200000000"
	err := db.CreateProfile(ctx, UserInput{
		KeycloakID:          utils.PointerUUID(uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")),
		FirstNameVernacular: utils.PointerString("first name"),
		LastNameVernacular:  utils.PointerString("last name"),
		Emails:              Emails{Primary: utils.PointerString("someemail@email.email")},
		Phones: Phones{
			MobileNumber:   &mobile,
			WhatsAppNumber: &whatsApp,
		},
	})

	assert.NoError(t, err)

	actualRows, err := db.Query(ctx, `SELECT phone_number, type FROM phone_numbers`)
	require.NoError(t, err)
	defer actualRows.Close()
	var actualPhoneNumbers [][]interface{}
	for actualRows.Next() {
		actual, err := actualRows.Values()
		require.NoError(t, err)

		actualPhoneNumbers = append(actualPhoneNumbers, actual)
	}
	assert.Len(t, actualPhoneNumbers, 2)
	assert.Contains(t, actualPhoneNumbers, []interface{}{"0100000000", "mobile"})
	assert.Contains(t, actualPhoneNumbers, []interface{}{"0200000000", "WhatsApp"})
}
