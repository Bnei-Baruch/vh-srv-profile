package main

import (
	"context"
	"errors"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_pgProfileDb_getUser_minimal_data_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)
	_, err := db.Exec(context.Background(), `
	INSERT into users (keycloak_id, first_name_vernacular, last_name_vernacular, primary_email) 
	VALUES ('11000000-0000-0000-0000-000000000000', 'first name', 'last name', 'someemail@email.email')`)
	require.NoError(t, err)

	actual, err := db.getProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"))

	assert.NoError(t, err)
	assert.Equal(t, userInput{
		firstNameVernacular: pointerString("first name"),
		lastNameVernacular:  pointerString("last name"),
		emails:              emails{primary: pointerString("someemail@email.email")},
	}, actual.userInput)
	assert.WithinDuration(t, time.Now(), actual.createdAt, 3*time.Second)
	assert.WithinDuration(t, time.Now(), actual.updatedAt, 3*time.Second)
	assert.False(t, actual.deleted)
}

func Test_pgProfileDb_getUser_full_data_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	//defer newTestPgProfileDb(t)
	var userID uuid.UUID
	err := db.QueryRow(context.Background(), `
	INSERT INTO users (keycloak_id,
                   first_name_latin,
                   first_name_vernacular,
                   last_name_latin,
                   last_name_vernacular,
                   street_address,
                   country,
                   state_region,
                   postal_code,
                   city,
                   gender,
                   marital_status,
                   date_of_birth,
                   primary_email,
                   alternate_email_1,
                   alternate_email_2,
                   first_language,
                   other_language_1,
                   other_language_2,
                   other_language_3,
                   other_language_4,
                   listening_language,
                   reading_language,
                   email_language,
                   study_start_year,
                   study_framework,
                   has_ten_group,
                   wants_ten_group,
                   name_of_ten_group)
VALUES ('441dc8be-7f58-40fb-951e-085797917110',
        'yasha',
        'yasha',
        'sol',
        'sol',
        'some street somewhere',
        'France',
        'Normandie',
        '76600',
        'Le Havre',
        'male',
        'Married',
        '1082-09-30',
        'yaakov.sabal@gmail.com',
        'johann.savalle@gmail.com',
        'contact@yasha.solution',
        'French',
        'English',
        'Chinese',
        'Spanish',
        'Italian',
        'English',
        'French',
        'French',
        1985,
        'some framework',
        true,
        true,
        'some name')
	RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)
	_, err = db.Exec(context.Background(), `
	INSERT into phone_numbers (user_id, phone_number, type) VALUES ($1, '+33783691190', 'mobile'), ($1, '+33783691191', 'WhatsApp'),
	($1, '+33783691192', 'Telegram')`, userID)
	require.NoError(t, err)

	actual, err := db.getProfile(context.Background(), uuid.FromStringOrNil("441dc8be-7f58-40fb-951e-085797917110"))

	assert.NoError(t, err)
	assert.Equal(t, userInput{
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
	}, actual.userInput)
	assert.WithinDuration(t, time.Now(), actual.createdAt, 3*time.Second)
	assert.WithinDuration(t, time.Now(), actual.updatedAt, 3*time.Second)
	assert.False(t, actual.deleted)
}

func Test_pgProfileDb_getUser_with_phone_numbers_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)
	var userID uuid.UUID
	err := db.QueryRow(context.Background(), `
	INSERT into users (keycloak_id, first_name_vernacular, last_name_vernacular, primary_email) 
	VALUES ('11000000-0000-0000-0000-000000000000', 'first name', 'last name', 'someemail@email.email') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)
	_, err = db.Exec(context.Background(), `
	INSERT into phone_numbers (user_id, phone_number, type) VALUES ($1, '0100000000', 'mobile'), ($1, '0200000000', 'WhatsApp')`, userID)
	require.NoError(t, err)

	actual, err := db.getProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"))

	expectedMobile := "0100000000"
	expectedWhatsApp := "0200000000"
	assert.NoError(t, err)
	assert.Equal(t, userInput{
		firstNameVernacular: pointerString("first name"),
		lastNameVernacular:  pointerString("last name"),
		emails:              emails{primary: pointerString("someemail@email.email")},
		phones: phones{
			mobileNumber:   &expectedMobile,
			whatsAppNumber: &expectedWhatsApp,
		},
	}, actual.userInput)
}

func Test_pgProfileDb_getUser_deleted_true_returns_nothing(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestPgProfileDb(t)
	defer newTestPgProfileDb(t)
	_, err := db.Exec(context.Background(), `
	INSERT into users (keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, deleted) 
	VALUES ('11000000-0000-0000-0000-000000000000', 'first name', 'last name', 'someemail@email.email', true)`)
	require.NoError(t, err)

	_, err = db.getProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"))

	assert.Error(t, err)
	assert.True(t, errors.Is(err, errProfileNotFound))
}
