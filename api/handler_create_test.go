package api

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

	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

func Test_profileHandler_create_succeeds_with_minimal_required_fields(t *testing.T) {
	sm := storageMock{}
	sm.On("CreateProfile", mock.Anything,
		repo.UserInput{
			KeycloakID:          utils.PointerUUID(uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")),
			FirstNameVernacular: utils.PointerString("First"),
			LastNameVernacular:  utils.PointerString("Name"),
			Emails:              repo.Emails{Primary: utils.PointerString("something@fakemail.com")},
			Status:              repo.UserStatus{UserID: utils.PointerUUID(uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"))},
		}).Return(nil)

	km := keycloakMock{}
	km.On("UpdateUser", "", "11000000-0000-0000-0000-000000000000", "First", "Name").Return(nil)

	profile := NewProfileManager(&sm, &km)
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

func Test_profileHandler_create_returns_bad_request_when_request_is_empty(t *testing.T) {
	sm := storageMock{}
	sm.On("CreateProfile", mock.Anything, mock.Anything).Return(nil)
	profile := NewProfileManager(&sm, nil)
	g := gin.New()
	g.POST("/", profile.create)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_profileHandler_create_returns_500_when_storage_returns_error(t *testing.T) {
	sm := storageMock{}
	sm.On("CreateProfile", mock.Anything, mock.Anything).Return(fmt.Errorf("some error"))
	profile := NewProfileManager(&sm, nil)
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
