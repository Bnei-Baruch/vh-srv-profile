package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

type readStorage interface {
	GetProfile(ctx context.Context, keycloakID uuid.UUID) (User, error)
	IsSubjectID(ctx context.Context, keycloakID, userID string) (bool, error)
}

func (db *ProfileDB) GetProfile(ctx context.Context, keycloakID uuid.UUID) (User, error) {
	var profile User
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
		&profile.UpdatedAt,
		&profile.CreatedAt,
		&profile.Deleted,
		&profile.UserInput.Status.Membership,
		&profile.UserInput.Status.MembershipType,
		&profile.UserInput.Status.Ticket,
		&profile.UserInput.Status.Convention,
		&profile.UserInput.Status.Galaxy,
		&profile.UserInput.FirstNameLatin,
		&profile.UserInput.FirstNameVernacular,
		&profile.UserInput.LastNameLatin,
		&profile.UserInput.LastNameVernacular,
		&profile.UserInput.Address.StreetAddress,
		&profile.UserInput.Address.Country,
		&profile.UserInput.Address.StateOrRegion,
		&profile.UserInput.Address.PostalCode,
		&profile.UserInput.Address.City,
		&profile.UserInput.Gender,
		&profile.UserInput.MaritalStatus,
		&profile.UserInput.DateOfBirth,
		&profile.UserInput.Emails.Primary,
		&profile.UserInput.Emails.Alternate1,
		&profile.UserInput.Emails.Alternate2,
		&profile.UserInput.Languages.First,
		&profile.UserInput.Languages.Other1,
		&profile.UserInput.Languages.Other2,
		&profile.UserInput.Languages.Other3,
		&profile.UserInput.Languages.Other4,
		&profile.UserInput.Languages.Listening,
		&profile.UserInput.Languages.Reading,
		&profile.UserInput.Languages.Email,
		&profile.UserInput.StudyStartYear,
		&profile.UserInput.StudyFramework,
		&profile.UserInput.Ten.HasGroup,
		&profile.UserInput.Ten.WantsGroup,
		&profile.UserInput.Ten.NameOfGroup,
		&profile.UserInput.Phones.WhatsAppNumber,
		&profile.UserInput.Phones.MobileNumber,
		&profile.UserInput.Phones.TelegramNumber,
	); err != nil {
		if err == pgx.ErrNoRows {
			return User{}, common.ErrProfileNotFound
		}
		return User{}, err
	}

	// Attach keycloak & User Id to profile struct
	profile.UserInput.KeycloakID = &keycloakID
	profile.UserID = &userID

	return profile, nil
}

func (db *ProfileDB) IsSubjectID(ctx context.Context, keycloakID, userID string) (bool, error) {
	row := db.QueryRow(ctx, "SELECT 1 FROM users WHERE keycloak_id = $1 AND user_id = $2", keycloakID, userID)
	var x int
	if err := row.Scan(&x); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
