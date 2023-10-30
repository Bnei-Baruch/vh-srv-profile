package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
)

func Test_ProfileDB_getUser_minimal_data_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestProfileDB(t)
	defer newTestProfileDB(t)
	_, err := db.Exec(context.Background(), `
	INSERT into users (keycloak_id, first_name_vernacular, last_name_vernacular, primary_email) 
	VALUES ('11000000-0000-0000-0000-000000000000', 'first name', 'last name', 'someemail@email.email')`)
	require.NoError(t, err)

	actual, err := db.GetProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"))

	assert.NoError(t, err)
	assert.Equal(t, UserInput{
		KeycloakID:          utils.PointerUUID(uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")),
		FirstNameVernacular: utils.PointerString("first name"),
		LastNameVernacular:  utils.PointerString("last name"),
		Emails:              Emails{Primary: utils.PointerString("someemail@email.email")},
	}, actual.UserInput)
	assert.WithinDuration(t, time.Now(), actual.CreatedAt, 3*time.Second)
	assert.WithinDuration(t, time.Now(), actual.UpdatedAt, 3*time.Second)
	assert.False(t, actual.Deleted)
}

func Test_ProfileDB_getUser_full_data_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestProfileDB(t)
	//defer newTestProfileDB(t)
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
        'FR',
        'Normandie',
        '76600',
        'Le Havre',
        'male',
        'Married',
        '1082-09-30',
        'yaakov.sabal@gmail.com',
        'johann.savalle@gmail.com',
        'contact@yasha.solution',
        'fr',
        'en',
        'zh',
        'es',
        'it',
        'en',
        'he',
        'fr',
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

	actual, err := db.GetProfile(context.Background(), uuid.FromStringOrNil("441dc8be-7f58-40fb-951e-085797917110"))

	assert.NoError(t, err)
	assert.Equal(t, UserInput{
		KeycloakID:          utils.PointerUUID(uuid.FromStringOrNil("441dc8be-7f58-40fb-951e-085797917110")),
		FirstNameLatin:      utils.PointerString("yasha"),
		FirstNameVernacular: utils.PointerString("yasha"),
		LastNameLatin:       utils.PointerString("sol"),
		LastNameVernacular:  utils.PointerString("sol"),
		Address: Address{
			StreetAddress: utils.PointerString("some street somewhere"),
			Country:       utils.PointerString("FR"),
			StateOrRegion: utils.PointerString("Normandie"),
			PostalCode:    utils.PointerString("76600"),
			City:          utils.PointerString("Le Havre"),
		},
		Gender:        utils.PointerString("male"),
		MaritalStatus: utils.PointerString("Married"),
		DateOfBirth:   utils.PointerTime(time.Date(1082, 9, 30, 0, 0, 0, 0, time.UTC)),
		Emails: Emails{
			Primary:    utils.PointerString("yaakov.sabal@gmail.com"),
			Alternate1: utils.PointerString("johann.savalle@gmail.com"),
			Alternate2: utils.PointerString("contact@yasha.solution"),
		},
		Phones: Phones{
			MobileNumber:   utils.PointerString("+33783691190"),
			WhatsAppNumber: utils.PointerString("+33783691191"),
			TelegramNumber: utils.PointerString("+33783691192"),
		},
		Languages: Languages{
			First:     utils.PointerString("fr"),
			Other1:    utils.PointerString("en"),
			Other2:    utils.PointerString("zh"),
			Other3:    utils.PointerString("es"),
			Other4:    utils.PointerString("it"),
			Listening: utils.PointerString("en"),
			Reading:   utils.PointerString("he"),
			Email:     utils.PointerString("fr"),
		},
		StudyStartYear: utils.PointerInt(1985),
		StudyFramework: utils.PointerString("some framework"),
		Ten: Ten{
			HasGroup:    utils.PointerBool(true),
			WantsGroup:  utils.PointerBool(true),
			NameOfGroup: utils.PointerString("some name"),
		},
	}, actual.UserInput)
	assert.WithinDuration(t, time.Now(), actual.CreatedAt, 3*time.Second)
	assert.WithinDuration(t, time.Now(), actual.UpdatedAt, 3*time.Second)
	assert.False(t, actual.Deleted)
}

func Test_ProfileDB_getUser_with_phone_numbers_succeeds(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestProfileDB(t)
	defer newTestProfileDB(t)
	var userID uuid.UUID
	err := db.QueryRow(context.Background(), `
	INSERT into users (keycloak_id, first_name_vernacular, last_name_vernacular, primary_email) 
	VALUES ('11000000-0000-0000-0000-000000000000', 'first name', 'last name', 'someemail@email.email') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)
	_, err = db.Exec(context.Background(), `
	INSERT into phone_numbers (user_id, phone_number, type) VALUES ($1, '0100000000', 'mobile'), ($1, '0200000000', 'WhatsApp')`, userID)
	require.NoError(t, err)

	actual, err := db.GetProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"))

	expectedMobile := "0100000000"
	expectedWhatsApp := "0200000000"
	assert.NoError(t, err)
	assert.Equal(t, UserInput{
		KeycloakID:          utils.PointerUUID(uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")),
		FirstNameVernacular: utils.PointerString("first name"),
		LastNameVernacular:  utils.PointerString("last name"),
		Emails:              Emails{Primary: utils.PointerString("someemail@email.email")},
		Phones: Phones{
			MobileNumber:   &expectedMobile,
			WhatsAppNumber: &expectedWhatsApp,
		},
	}, actual.UserInput)
}

func Test_ProfileDB_getUser_deleted_true_returns_nothing(t *testing.T) {
	checkIntegrationTest(t)
	db := newTestProfileDB(t)
	defer newTestProfileDB(t)
	_, err := db.Exec(context.Background(), `
	INSERT into users (keycloak_id, first_name_vernacular, last_name_vernacular, primary_email, deleted) 
	VALUES ('11000000-0000-0000-0000-000000000000', 'first name', 'last name', 'someemail@email.email', true)`)
	require.NoError(t, err)

	_, err = db.GetProfile(context.Background(), uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000"))

	assert.Error(t, err)
	assert.True(t, errors.Is(err, common.ErrProfileNotFound))
}
