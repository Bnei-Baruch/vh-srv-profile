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

func Test_profileHandler_get_full_succeeds(t *testing.T) {
	sm := storageMock{}
	parisTZ, err := time.LoadLocation("Europe/Paris")
	require.NoError(t, err)
	sm.On("getProfile", mock.Anything, uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")).
		Return(user{
			createdAt: time.Date(2021, 12, 12, 0, 0, 0, 0, parisTZ),
			updatedAt: time.Date(2021, 12, 12, 0, 0, 0, 0, parisTZ),
			userInput: userInput{
				firstNameLatin:      pointerString("yasha"),
				firstNameVernacular: pointerString("yasha"),
				lastNameLatin:       pointerString("sol"),
				lastNameVernacular:  pointerString("sol"),
				address: address{
					streetAddress: pointerString("some street somewhere"),
					country:       pointerString("France"),
					stateOrRegion: pointerString("Normandie"),
					postalCode:    pointerString("76600"),
					city:          pointerString("Le Havre"),
				},
				gender:        pointerString("male"),
				maritalStatus: pointerString("Married"),
				dateOfBirth:   pointerTime(time.Date(1082, 9, 30, 0, 0, 0, 0, time.UTC)),
				emails: emails{
					primary:    pointerString("yaakov.sabal@gmail.com"),
					alternate1: pointerString("johann.savalle@gmail.com"),
					alternate2: pointerString("contact@yasha.solution"),
				},
				phones: phones{
					mobileNumber:   pointerString("+33783691190"),
					whatsAppNumber: pointerString("+33783691191"),
					telegramNumber: pointerString("+33783691192"),
				},
				languages: languages{
					first:     pointerString("French"),
					other1:    pointerString("English"),
					other2:    pointerString("Chinese"),
					other3:    pointerString("Spanish"),
					other4:    pointerString("Italian"),
					listening: pointerString("English"),
					reading:   pointerString("French"),
					email:     pointerString("French"),
				},
				studyStartYear: pointerInt(1985),
				studyFramework: pointerString("some framework"),
				ten: ten{
					hasGroup:    pointerBool(true),
					wantsGroup:  pointerBool(true),
					nameOfGroup: pointerString("some name"),
				},
			}}, nil)
	profile := profileManager{getter: &sm}
	g := gin.New()
	g.GET("/:keycloak_id", profile.get)

	r := httptest.NewRequest(http.MethodGet, "/11000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `
	{
	   "updated_at":"2021-12-12T00:00:00+01:00",
	   "created_at":"2021-12-12T00:00:00+01:00",
	   "deleted":false,
	   "first_name_latin":"yasha",
	   "first_name_vernacular":"yasha",
	   "last_name_latin":"sol",
	   "last_name_vernacular":"sol",
	   "street_address":"some street somewhere",
	   "country":"France",
	   "state_region":"Normandie",
	   "postal_code":"76600",
	   "city":"Le Havre",
	   "gender":"male",
	   "marital_status":"Married",
	   "date_of_birth":"1082-09-30",
	   "primary_email":"yaakov.sabal@gmail.com",
	   "alternate_email_1":"johann.savalle@gmail.com",
	   "alternate_email_2":"contact@yasha.solution",
	   "mobile_number":"+33783691190",
	   "whats_app_number":"+33783691191",
	   "telegram_number":"+33783691192",
	   "first_language":"French",
	   "other_language_1":"English",
	   "other_language_2":"Chinese",
	   "other_language_3":"Spanish",
	   "other_language_4":"Italian",
	   "listening_language":"English",
	   "reading_language":"French",
	   "email_language":"French",
	   "study_start_year":1985,
	   "study_framework":"some framework",
	   "has_ten_group":true,
	   "wants_ten_group":true,
	   "name_ten_group":"some name"
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
