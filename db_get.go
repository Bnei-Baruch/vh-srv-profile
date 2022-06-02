package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"

	uuid "github.com/satori/go.uuid"
)

func (db *pgProfileDB) getProfile(ctx context.Context, keycloakID uuid.UUID) (user, error) {
	var profile user
	var userID uuid.UUID
	if err := db.QueryRow(ctx, `
	SELECT users.user_id,
		updated_at,
		created_at,
		deleted,
		status.membership,
		status.membership_type,
		status.ticket,
		status.convention,
		status.galaxy,
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
		name_of_ten_group,
		(SELECT phone_number FROM phone_numbers as p WHERE p.user_id = users.user_id and type='WhatsApp' ) as whats_app,
		(SELECT phone_number FROM phone_numbers as p WHERE p.user_id = users.user_id and type='mobile' ) as mobile,
		(SELECT phone_number FROM phone_numbers as p WHERE p.user_id = users.user_id and type='Telegram' ) as telegram 
	FROM users
	LEFT JOIN status ON users.user_id = status.user_id
	WHERE keycloak_id = $1
	AND deleted = false`, keycloakID).Scan(
		&userID,
		&profile.updatedAt,
		&profile.createdAt,
		&profile.deleted,
		&profile.userInput.status.membership,
		&profile.userInput.status.membershipType,
		&profile.userInput.status.ticket,
		&profile.userInput.status.convention,
		&profile.userInput.status.galaxy,
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
		&profile.userInput.phones.whatsAppNumber,
		&profile.userInput.phones.mobileNumber,
		&profile.userInput.phones.telegramNumber,
	); err != nil {
		if err == pgx.ErrNoRows {
			return user{}, fmt.Errorf("%w: %q", errProfileNotFound, keycloakID)
		}
		return user{}, err
	}

	// Attach keycloak & user Id to profile struct
	profile.userInput.keycloakID = &keycloakID
	profile.userID = &userID

	return profile, nil
}
