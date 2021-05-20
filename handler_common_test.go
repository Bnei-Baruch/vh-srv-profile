package main

import (
	"context"

	uuid "github.com/satori/go.uuid"

	"github.com/stretchr/testify/mock"
)

type storageMock struct {
	mock.Mock
}

func (m *storageMock) createProfile(ctx context.Context, user userInput) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *storageMock) getProfile(ctx context.Context, keycloakID uuid.UUID) (user, error) {
	args := m.Called(ctx, keycloakID)
	return args.Get(0).(user), args.Error(1)
}

func (m *storageMock) updateProfile(ctx context.Context, keycloakID uuid.UUID, user userInput) error {
	args := m.Called(ctx, keycloakID, user)
	return args.Error(0)
}
