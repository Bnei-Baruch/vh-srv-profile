package main

import (
	"context"

	uuid "github.com/satori/go.uuid"
)

func (db *pgProfileDB) getProfile(ctx context.Context, keycloakID uuid.UUID) (user, error) {
	var profile user
	var userID uuid.UUID
	if err := db.QueryRow(ctx, `
	SELECT user_id,
       updated_at,
       created_at,
       deleted,
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
       name_of_ten_group
	FROM users
	WHERE keycloak_id = $1`, keycloakID).Scan(
		&userID,
		&profile.updatedAt,
		&profile.createdAt,
		&profile.deleted,
		&profile.userInput.firstNameLatin,
		&profile.userInput.firstNameVernacular,
		&profile.userInput.lastNameLatin,
		&profile.userInput.lastNameVernacular,
		&profile.userInput.address.streetAddress,
		&profile.userInput.address.country,
		&profile.userInput.address.stateOrRegion,
		&profile.userInput.address.postalCode,
		&profile.userInput.address.city,
		&profile.userInput.gender,
		&profile.userInput.maritalStatus,
		&profile.userInput.dateOfBirth,
		&profile.userInput.emails.primary,
		&profile.userInput.emails.alternate1,
		&profile.userInput.emails.alternate2,
		&profile.userInput.languages.first,
		&profile.userInput.languages.other1,
		&profile.userInput.languages.other2,
		&profile.userInput.languages.other3,
		&profile.userInput.languages.other4,
		&profile.userInput.languages.listening,
		&profile.userInput.languages.reading,
		&profile.userInput.languages.email,
		&profile.userInput.studyStartYear,
		&profile.userInput.studyFramework,
		&profile.userInput.ten.hasGroup,
		&profile.userInput.ten.wantsGroup,
		&profile.userInput.ten.nameOfGroup,
	); err != nil {
		return user{}, err
	}

	type phone struct {
		number    *string
		phoneType string
	}
	var phoneNumbers []phone
	rows, err := db.Query(ctx, `
	SELECT phone_number, 
		type 
	FROM phone_numbers
	WHERE user_id = $1`, userID)
	if err != nil {
		return user{}, err
	}
	for rows.Next() {
		var temp phone
		if err := rows.Scan(&temp.number, &temp.phoneType); err != nil {
			return user{}, err
		}
		phoneNumbers = append(phoneNumbers, temp)
	}

	var mobileNumber, whatsAppNumber, telegramNumber *string
	for _, num := range phoneNumbers {
		switch num.phoneType {
		case mobile:
			mobileNumber = num.number
		case whatsApp:
			whatsAppNumber = num.number
		case telegram:
			telegramNumber = num.number
		}
	}

	profile.userInput.phones = phones{
		mobileNumber:   mobileNumber,
		whatsAppNumber: whatsAppNumber,
		telegramNumber: telegramNumber,
	}

	return profile, nil
}
