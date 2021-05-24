package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_profileHandler_update_succeeds(t *testing.T) {
	sm := storageMock{}
	sm.On("updateProfile", mock.Anything,
		uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"),
		userInput{
			firstNameVernacular: pointerString("First"),
			lastNameVernacular:  pointerString("Name"),
			emails:              emails{primary: pointerString("something@fakemail.com")},
		}).Return(nil)
	profile := profileManager{updater: &sm}
	g := gin.New()
	g.PATCH("/:keycloak_id", profile.update)

	r := httptest.NewRequest(http.MethodPatch, "/11000000-0000-0000-0000-000000000000", strings.NewReader(`{
		"first_name_vernacular":"First",
		"last_name_vernacular":"Name",
		"primary_email":"something@fakemail.com"
	}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_profileHandler_update_returns_400_when_storage_returns_errProfileNotFound(t *testing.T) {
	sm := storageMock{}
	sm.On("updateProfile", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("%w: %q",
		errProfileNotFound, "example string"))
	profile := profileManager{updater: &sm}
	g := gin.New()
	g.PATCH("/:keycloak_id", profile.update)

	r := httptest.NewRequest(http.MethodPatch, "/11000000-0000-0000-0000-000000000000", strings.NewReader(`{
		"first_name_vernacular":"First",
		"last_name_vernacular":"Name",
		"primary_email":"something@fakemail.com"
	}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"no profile found for keycloak id: \"example string\""}`, w.Body.String())
}

func Test_profileHandler_update_returns_500_when_storage_returns_error(t *testing.T) {
	sm := storageMock{}
	sm.On("updateProfile", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("some error"))
	profile := profileManager{updater: &sm}
	g := gin.New()
	g.PATCH("/:keycloak_id", profile.update)

	r := httptest.NewRequest(http.MethodPatch, "/11000000-0000-0000-0000-000000000000", strings.NewReader(`{
		"first_name_vernacular":"First",
		"last_name_vernacular":"Name",
		"primary_email":"something@fakemail.com"
	}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
