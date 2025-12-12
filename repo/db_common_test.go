package repo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/events"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/testutil"
)

// newTestProfileDBIsolated creates a new isolated test database for unit tests
func newTestProfileDBIsolated(t *testing.T) *ProfileDB {
	t.Helper()

	ctx := context.Background()
	dbURL, err := testutil.NewTestProfileDB(t, ctx)
	require.NoError(t, err)

	eventEmitter, err := events.CreateEmitter()
	require.NoError(t, err)

	db, err := NewProfileDB(ctx, dbURL, eventEmitter)
	require.NoError(t, err)

	// Ensure database connection is closed after test
	t.Cleanup(func() {
		db.Close()
	})

	return db
}

type orderServiceMock struct {
	mock.Mock
}

func (m *orderServiceMock) GetAccountByID(ctx context.Context, accountID int) (*orders.Account, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).(*orders.Account), args.Error(1)
}

func (m *orderServiceMock) GetAccountByEmailIfExist(ctx context.Context, email string) (*orders.Account, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*orders.Account), args.Error(1)
}

func (m *orderServiceMock) GetOrders(ctx context.Context,
	email string,
	productType string,
	evaluateMembership bool,
	paymentDateOrder string,
	limit int, offset int) ([]orders.Order, error) {
	args := m.Called(ctx, email, productType, evaluateMembership, paymentDateOrder, limit, offset)
	return args.Get(0).([]orders.Order), args.Error(1)
}

func (m *orderServiceMock) GetOrderByID(ctx context.Context, orderID int) (*orders.Order, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(*orders.Order), args.Error(1)
}

func (m *orderServiceMock) SetOrderStartingDate(ctx context.Context, orderID int, startingDate time.Time) error {
	args := m.Called(ctx, orderID, startingDate)
	return args.Error(0)
}

func (m *orderServiceMock) CancelOrder(ctx context.Context, orderID int) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func (m *orderServiceMock) GetOrderPayments(ctx context.Context,
	orderID int,
	createdAtOrder string,
	limit int, offset int) ([]orders.Payment, error) {
	args := m.Called(ctx, orderID, createdAtOrder, limit, offset)
	return args.Get(0).([]orders.Payment), args.Error(1)
}

func (m *orderServiceMock) GetPaymentByID(ctx context.Context, paymentID int) (*orders.Payment, error) {
	args := m.Called(ctx, paymentID)
	return args.Get(0).(*orders.Payment), args.Error(1)
}

func (m *orderServiceMock) GetSpecial(ctx context.Context, email string) (*orders.Special, error) {
	args := m.Called(ctx, email)
	if args.Get(0) != nil {
		return args.Get(0).(*orders.Special), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *orderServiceMock) GetSpecials(ctx context.Context, email string) ([]orders.Special, error) {
	args := m.Called(ctx, email)
	if args.Get(0) != nil {
		return args.Get(0).([]orders.Special), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *orderServiceMock) DeleteSpecial(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *orderServiceMock) StatusByEmail(ctx context.Context, email string) (*orders.Status, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*orders.Status), args.Error(1)
}

func (m *orderServiceMock) DeleteSpecialIfExist(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *orderServiceMock) UpdateSpecialSetKeycloakIdByEmail(ctx context.Context, payload map[string]interface{}) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}
