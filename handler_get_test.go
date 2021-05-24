package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_profileHandler_get_succeeds(t *testing.T) {
	sm := storageMock{}
	parisTZ, err := time.LoadLocation("Europe/Paris")
	require.NoError(t, err)
	sm.On("getProfile", mock.Anything, uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")).
		Return(user{
			updatedAt: time.Date(2020, 1, 1, 1, 0, 0, 0, parisTZ),
			createdAt: time.Date(2019, 12, 12, 12, 0, 0, 0, parisTZ),
			deleted:   false,
			userInput: userInput{
				keycloakID:          pointerUUID(uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")),
				firstNameVernacular: pointerString("first name"),
				lastNameVernacular:  pointerString("last name"),
				emails:              emails{primary: pointerString("someemail@email.com")},
			},
		}, nil)
	profile := profileManager{getter: &sm}
	g := gin.New()
	g.GET("/:keycloak_id", profile.get)

	r := httptest.NewRequest(http.MethodGet, "/11000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `
	{
	   "updated_at":"2020-01-01T01:00:00+01:00",
	   "created_at":"2019-12-12T12:00:00+01:00",
	   "deleted":false,
	   "first_name_vernacular":"first name",
	   "last_name_vernacular":"last name",
	   "primary_email":"someemail@email.com"
	}`, w.Body.String())
}

func Test_profileHandler_get_returns_400_when_storage_returns_errProfileNotFound(t *testing.T) {
	sm := storageMock{}
	sm.On("getProfile", mock.Anything, mock.Anything).Return(user{}, fmt.Errorf("%w: %q",
		errProfileNotFound, "example string"))
	profile := profileManager{getter: &sm}
	g := gin.New()
	g.GET("/:keycloak_id", profile.get)

	r := httptest.NewRequest(http.MethodGet, "/11000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"no profile found for keycloak id: \"example string\""}`, w.Body.String())
}

func Test_profileHandler_get_returns_500_when_storage_returns_error(t *testing.T) {
	sm := storageMock{}
	sm.On("getProfile", mock.Anything, mock.Anything).Return(user{}, fmt.Errorf("some error"))
	profile := profileManager{getter: &sm}
	g := gin.New()
	g.GET("/:keycloak_id", profile.get)

	r := httptest.NewRequest(http.MethodGet, "/11000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
