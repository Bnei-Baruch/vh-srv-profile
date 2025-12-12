package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/volatiletech/null/v9"

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

func TestProfileHandler_UpdateMaritalStatus(t *testing.T) {
	const testKeycloakID = "22000000-0000-0000-0000-000000000000"

	testCases := []struct {
		name               string
		payload            string
		expectedStatusCode int
		setupMock          func(sm *storageMock)
	}{
		{
			name:               "update with valid string",
			payload:            `{"marital_status": "Married"}`,
			expectedStatusCode: http.StatusOK,
			setupMock: func(sm *storageMock) {
				expectedInput := repo.UserInput{
					MaritalStatus: null.String{String: "Married", Valid: true, Set: true},
				}
				sm.On("UpdateProfile", mock.Anything, uuid.FromStringOrNil(testKeycloakID), expectedInput).Return(nil).Once()
			},
		},
		{
			name:               "update without marital status should not update marital status",
			payload:            `{}`,
			expectedStatusCode: http.StatusOK,
			setupMock: func(sm *storageMock) {
				// A JSON 'null' unmarshals into a non valid value.
				expectedInput := repo.UserInput{
					MaritalStatus: null.String{Set: false},
				}
				sm.On("UpdateProfile", mock.Anything, uuid.FromStringOrNil(testKeycloakID), expectedInput).Return(nil).Once()
			},
		},
		{
			name:               "update with null should be accepted and set null",
			payload:            `{"marital_status": null}`,
			expectedStatusCode: http.StatusOK,
			setupMock: func(sm *storageMock) {
				// A JSON 'null' unmarshals into a non valid value.
				expectedInput := repo.UserInput{
					MaritalStatus: null.String{Set: true, Valid: false},
				}
				sm.On("UpdateProfile", mock.Anything, uuid.FromStringOrNil(testKeycloakID), expectedInput).Return(nil).Once()
			},
		},
		{
			name:               "update with a number (non-valid type)",
			payload:            `{"marital_status": 123}`,
			expectedStatusCode: http.StatusBadRequest,
			setupMock: func(sm *storageMock) {
				// UpdateProfile should NOT be called because binding will fail.
			},
		},
		{
			name:               "update with empty string (invalid value)",
			payload:            `{"marital_status": ""}`,
			expectedStatusCode: http.StatusInternalServerError,
			setupMock: func(sm *storageMock) {
				expectedInput := repo.UserInput{
					MaritalStatus: null.String{String: "", Valid: true, Set: true},
				}
				err := errors.New("Mock 500, as empty string is not valid marital status")
				sm.On("UpdateProfile", mock.Anything, uuid.FromStringOrNil(testKeycloakID), expectedInput).Return(err).Once()
			},
		},
		{
			name:               "update with non-empty wrong string",
			payload:            `{"marital_status": "some_random_string"}`,
			expectedStatusCode: http.StatusInternalServerError,
			setupMock: func(sm *storageMock) {
				expectedInput := repo.UserInput{
					MaritalStatus: null.String{String: "some_random_string", Valid: true, Set: true},
				}
				err := errors.New("Mock 500, as some_radom_string is not valid marital status")
				sm.On("UpdateProfile", mock.Anything, uuid.FromStringOrNil(testKeycloakID), expectedInput).Return(err).Once()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sm := storageMock{}
			tc.setupMock(&sm)

			// We don't need a Keycloak mock for this test as it's not touched by the update handler
			profile := NewProfileManager(&sm)

			g := gin.New()
			g.PATCH("/:keycloak_id", profile.update)

			r := NewRequestAsRoot(http.MethodPatch, fmt.Sprintf("/%s", testKeycloakID), strings.NewReader(tc.payload))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			g.ServeHTTP(w, r)

			assert.Equal(t, tc.expectedStatusCode, w.Code, "Unexpected status code")
			sm.AssertExpectations(t)
		})
	}
}
