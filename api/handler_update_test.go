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

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

func Test_profileHandler_update_succeeds(t *testing.T) {
	sm := storageMock{}
	sm.On("UpdateProfile", mock.Anything,
		uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"),
		repo.UserInput{
			FirstNameVernacular: utils.PointerString("First"),
			LastNameVernacular:  utils.PointerString("Name"),
			Emails:              repo.Emails{Primary: utils.PointerString("something@fakemail.com")},
		}).Return(nil)

	profile := NewProfileManager(&sm)
	profile.SetKeycloakServiceFactory(MakeKeycloakAPIMockFactory(func(km *keycloakAPIMock) {
		km.On("UpdateUser", mock.Anything, "11000000-0000-0000-0000-000000000000",
			utils.PointerString("First"), utils.PointerString("Name")).Return(nil)
	}))
	g := gin.New()
	g.PATCH("/:keycloak_id", profile.update)

	r := NewRequestAsRoot(http.MethodPatch, "/11000000-0000-0000-0000-000000000000", strings.NewReader(`{
		"first_name_vernacular":"First",
		"last_name_vernacular":"Name",
		"primary_email":"something@fakemail.com"
	}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_profileHandler_update_returns_404_when_storage_returns_errProfileNotFound(t *testing.T) {
	sm := storageMock{}
	sm.On("UpdateProfile", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("%w: %q",
		common.ErrProfileNotFound, "example string"))
	profile := NewProfileManager(&sm)
	g := gin.New()
	g.PATCH("/:keycloak_id", profile.update)

	r := NewRequestAsRoot(http.MethodPatch, "/11000000-0000-0000-0000-000000000000", strings.NewReader(`{
		"first_name_vernacular":"First",
		"last_name_vernacular":"Name",
		"primary_email":"something@fakemail.com"
	}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_profileHandler_update_returns_500_when_storage_returns_error(t *testing.T) {
	sm := storageMock{}
	sm.On("UpdateProfile", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("some error"))
	profile := NewProfileManager(&sm)
	g := gin.New()
	g.PATCH("/:keycloak_id", profile.update)

	r := NewRequestAsRoot(http.MethodPatch, "/11000000-0000-0000-0000-000000000000", strings.NewReader(`{
		"first_name_vernacular":"First",
		"last_name_vernacular":"Name",
		"primary_email":"something@fakemail.com"
	}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
