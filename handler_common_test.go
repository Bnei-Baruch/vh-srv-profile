package main

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type storageMock struct {
	mock.Mock
}

func (m *storageMock) createUser(ctx context.Context, user userInput) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *storageMock) getUser(ctx context.Context, keycloakID string) (user, error) {
	args := m.Called(ctx, keycloakID)
	return args.Get(0).(user), args.Error(1)
}
