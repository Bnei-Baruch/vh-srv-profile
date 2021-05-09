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
			keycloakID:     "11000000-0000-0000-0000-000000000000",
			firstNameLatin: "First",
			lastNameLatin:  "Name",
			gender:         "female",
			maritalStatus:  "Married",
			emails:         emails{primary: "something@fakemail.com"},
			phones:         phones{mobileNumber: "+330601010101"},
			languages: languages{
				first:     "English",
				listening: "French",
				reading:   "Spanish",
				email:     "Hebrew",
			},
			studyStartYear: 2005,
			studyFramework: "some framework",
			ten:            ten{hasGroup: true},
		}).Return(nil)
	profile := profile{db: &sm}
	g := gin.New()
	g.POST("/", profile.create)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{
		"keycloak_id":"11000000-0000-0000-0000-000000000000",
		"first_name_latin":"First",
		"last_name_latin":"Name",
		"gender":"female",
		"marital_status":"Married",
		"primary_email":"something@fakemail.com",
		"mobile_number":"+330601010101",
		"first_language":"English",
		"listening_language":"French",
		"reading_language":"Spanish",
		"email_language":"Hebrew",
		"study_start_year":2005,
		"study_framework":"some framework",
		"has_ten_group":true
	}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func Test_ProfileCreate_returns_500_when_storage_returns_error(t *testing.T) {
	sm := storageMock{}
	sm.On("createUser", mock.Anything, mock.Anything).Return(fmt.Errorf("some error"))
	profile := profile{db: &sm}
	g := gin.New()
	g.POST("/", profile.create)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{
		"keycloak_id":"11000000-0000-0000-0000-000000000000",
		"first_name_latin":"First",
		"last_name_latin":"Name",
		"gender":"female",
		"marital_status":"Married",
		"primary_email":"something@fakemail.com",
		"mobile_number":"+330601010101",
		"first_language":"English",
		"listening_language":"French",
		"reading_language":"Spanish",
		"email_language":"Hebrew",
		"study_start_year":2005,
		"study_framework":"some framework",
		"has_ten_group":true
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
