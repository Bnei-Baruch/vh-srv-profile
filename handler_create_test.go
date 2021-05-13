package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_ProfileCreate_succeeds_with_minimal_required_fields(t *testing.T) {
	sm := storageMock{}
	sm.On("createUser", mock.Anything,
		user{
			keycloakID:          "11000000-0000-0000-0000-000000000000",
			firstNameVernacular: "First",
			lastNameVernacular:  "Name",
			emails:              emails{primary: "something@fakemail.com"},
		}).Return(nil)
	profile := profileManager{db: &sm}
	g := gin.New()
	g.POST("/", profile.create)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{
		"keycloak_id":"11000000-0000-0000-0000-000000000000",
		"first_name_vernacular":"First",
		"last_name_vernacular":"Name",
		"primary_email":"something@fakemail.com"
	}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func Test_ProfileCreate_returns_500_when_storage_returns_error(t *testing.T) {
	sm := storageMock{}
	sm.On("createUser", mock.Anything, mock.Anything).Return(fmt.Errorf("some error"))
	profile := profileManager{db: &sm}
	g := gin.New()
	g.POST("/", profile.create)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{
		"keycloak_id":"11000000-0000-0000-0000-000000000000",
		"first_name_vernacular":"First",
		"last_name_vernacular":"Name",
		"primary_email":"something@fakemail.com"
	}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

type storageMock struct {
	mock.Mock
}

func (m *storageMock) createUser(ctx context.Context, user user) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
