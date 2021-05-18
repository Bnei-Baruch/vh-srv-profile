package main

import "context"

type storage interface {
	createUser(ctx context.Context, user userInput) error
	getUser(ctx context.Context, keycloakID string) (user, error)
}

type profileManager struct {
	db storage
}
