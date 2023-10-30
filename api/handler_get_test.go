package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

func Test_profileHandler_get_succeeds(t *testing.T) {
	sm := storageMock{}
	parisTZ, err := time.LoadLocation("Europe/Paris")
	require.NoError(t, err)
	userID := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	sm.On("GetProfile", mock.Anything, userID).
		Return(repo.User{
			UserID:    utils.PointerUUID(userID),
			UpdatedAt: time.Date(2020, 1, 1, 1, 0, 0, 0, parisTZ),
			CreatedAt: time.Date(2019, 12, 12, 12, 0, 0, 0, parisTZ),
			Deleted:   false,
			UserInput: repo.UserInput{
				KeycloakID:          utils.PointerUUID(userID),
				FirstNameVernacular: utils.PointerString("first name"),
				LastNameVernacular:  utils.PointerString("last name"),
				Emails:              repo.Emails{Primary: utils.PointerString("someemail@email.com")},
			},
		}, nil)
	profile := NewProfileManager(&sm, nil)
	g := gin.New()
	g.GET("/:keycloak_id", profile.get)

	r := httptest.NewRequest(http.MethodGet, "/11000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `
	{
		"user_id":"11000000-0000-0000-0000-000000000000",
		"keycloak_id":"11000000-0000-0000-0000-000000000000",
	   	"updated_at":"2020-01-01T01:00:00+01:00",
	   	"created_at":"2019-12-12T12:00:00+01:00",
	   	"deleted":false,
	   	"first_name_vernacular":"first name",
	   	"last_name_vernacular":"last name",
	   	"primary_email":"someemail@email.com",
		"status":{}
	}`, w.Body.String())
}

func Test_profileHandler_get_full_succeeds(t *testing.T) {
	sm := storageMock{}
	parisTZ, err := time.LoadLocation("Europe/Paris")
	require.NoError(t, err)
	userID := uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")
	sm.On("GetProfile", mock.Anything, userID).
		Return(repo.User{
			UserID:    utils.PointerUUID(userID),
			CreatedAt: time.Date(2021, 12, 12, 0, 0, 0, 0, parisTZ),
			UpdatedAt: time.Date(2021, 12, 12, 0, 0, 0, 0, parisTZ),
			UserInput: repo.UserInput{
				KeycloakID:          utils.PointerUUID(userID),
				FirstNameLatin:      utils.PointerString("yasha"),
				FirstNameVernacular: utils.PointerString("yasha"),
				LastNameLatin:       utils.PointerString("sol"),
				LastNameVernacular:  utils.PointerString("sol"),
				Address: repo.Address{
					StreetAddress: utils.PointerString("some street somewhere"),
					Country:       utils.PointerString("France"),
					StateOrRegion: utils.PointerString("Normandie"),
					PostalCode:    utils.PointerString("76600"),
					City:          utils.PointerString("Le Havre"),
				},
				Gender:        utils.PointerString("male"),
				MaritalStatus: utils.PointerString("Married"),
				DateOfBirth:   utils.PointerTime(time.Date(1082, 9, 30, 0, 0, 0, 0, time.UTC)),
				Emails: repo.Emails{
					Primary:    utils.PointerString("yaakov.sabal@gmail.com"),
					Alternate1: utils.PointerString("johann.savalle@gmail.com"),
					Alternate2: utils.PointerString("contact@yasha.solution"),
				},
				Phones: repo.Phones{
					MobileNumber:   utils.PointerString("+33783691190"),
					WhatsAppNumber: utils.PointerString("+33783691191"),
					TelegramNumber: utils.PointerString("+33783691192"),
				},
				Languages: repo.Languages{
					First:     utils.PointerString("fr"),
					Other1:    utils.PointerString("en"),
					Other2:    utils.PointerString("zh"),
					Other3:    utils.PointerString("es"),
					Other4:    utils.PointerString("it"),
					Listening: utils.PointerString("en"),
					Reading:   utils.PointerString("fr"),
					Email:     utils.PointerString("fr"),
				},
				StudyStartYear: utils.PointerInt(1985),
				StudyFramework: utils.PointerString("some framework"),
				Ten: repo.Ten{
					HasGroup:    utils.PointerBool(true),
					WantsGroup:  utils.PointerBool(true),
					NameOfGroup: utils.PointerString("some name"),
				},
			}}, nil)
	profile := NewProfileManager(&sm, nil)
	g := gin.New()
	g.GET("/:keycloak_id", profile.get)

	r := httptest.NewRequest(http.MethodGet, "/11000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `
	{
		"user_id":"11000000-0000-0000-0000-000000000000",
		"keycloak_id":"11000000-0000-0000-0000-000000000000",
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
	   	"first_language":"fr",
	   	"other_language_1":"en",
	   	"other_language_2":"zh",
	   	"other_language_3":"es",
	   	"other_language_4":"it",
	   	"listening_language":"en",
	   	"reading_language":"fr",
	   	"email_language":"fr",
	   	"study_start_year":1985,
	   	"study_framework":"some framework",
	   	"has_ten_group":true,
	   	"wants_ten_group":true,
	   	"name_ten_group":"some name",
		"status":{}
	}`, w.Body.String())
}

func Test_profileHandler_get_returns_404_when_storage_returns_errProfileNotFound(t *testing.T) {
	sm := storageMock{}
	sm.On("GetProfile", mock.Anything, mock.Anything).Return(repo.User{}, fmt.Errorf("%w: %q",
		common.ErrProfileNotFound, "example string"))
	profile := NewProfileManager(&sm, nil)
	g := gin.New()
	g.GET("/:keycloak_id", profile.get)

	r := httptest.NewRequest(http.MethodGet, "/11000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":"no profile found for keycloak id: \"example string\""}`, w.Body.String())
}

func Test_profileHandler_get_returns_500_when_storage_returns_error(t *testing.T) {
	sm := storageMock{}
	sm.On("GetProfile", mock.Anything, mock.Anything).Return(repo.User{}, fmt.Errorf("some error"))
	profile := NewProfileManager(&sm, nil)
	g := gin.New()
	g.GET("/:keycloak_id", profile.get)

	r := httptest.NewRequest(http.MethodGet, "/11000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
