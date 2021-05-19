package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_profileHandler_create_succeeds_with_minimal_required_fields(t *testing.T) {
	sm := storageMock{}
	sm.On("createUser", mock.Anything,
		userInput{
			keycloakID:          pointerString("11000000-0000-0000-0000-000000000000"),
			firstNameVernacular: pointerString("First"),
			lastNameVernacular:  pointerString("Name"),
			emails:              emails{primary: pointerString("something@fakemail.com")},
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

func Test_profileHandler_create_returns_500_when_storage_returns_error(t *testing.T) {
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
