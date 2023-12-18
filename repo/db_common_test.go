package repo

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
)

type orderServiceMock struct {
	mock.Mock
}

func (m *orderServiceMock) GetAccountByID(ctx context.Context, accountID int) (*orders.Account, error) {
	args := m.Called(ctx, accountID)
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

func (m *orderServiceMock) DeleteSpecial(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *orderServiceMock) StatusByEmail(ctx context.Context, email string) (*orders.Status, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*orders.Status), args.Error(1)
}
