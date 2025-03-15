package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

type readMultipleProfileStorage interface {
	GetMultipleProfiles(ctx context.Context,
		intSkip int,
		intLimit int,
		country string,
		email string,
		name string,
		tenGroupName string,
		language string,
		firstLanguage string,
		otherLanguageOne string,
		otherLanguageTwo string,
		otherLanguageThree string,
		otherLanguageFour string,
		updatedAt string,
		createdAt string,
		membership string,
		membershipType string,
		convention string,
		ticket string,
		galaxy string,
		gender string,
		checkAlternativeEmails bool) ([]User, error)
	FetchProfileBasedOnPhoneNumber(ctx context.Context, phoneNumber string) (User, error)
}

func (db *ProfileDB) GetMultipleProfiles(ctx context.Context, intSkip int, intLimit int, country string, email string,
	name string, tenGroupName string, language string, firstLanguage string, otherLanguageOne string,
	otherLanguageTwo string, otherLanguageThree string, otherLanguageFour string, updatedAt string, createdAt string,
	membership string, membershipType string, convention string, ticket string, galaxy string, gender string, checkAlternativeEmails bool) ([]User, error) {
	var keycloakId string
	users := []User{}

	userDbWhereQuery, orderByQuery := buildAndGetWhereUserQuery(country, email, name, tenGroupName, language, firstLanguage,
		otherLanguageOne, otherLanguageTwo, otherLanguageThree, otherLanguageFour, updatedAt, createdAt, membership,
		membershipType, convention, ticket, galaxy, gender, checkAlternativeEmails)

	rows, err := db.Query(ctx, `
		SELECT users.user_id,
		keycloak_id,
		users.updated_at,
		users.created_at,
		deleted,
		membership.active,
		membership.type,
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
	LEFT JOIN membership ON users.user_id = membership.user_id`+userDbWhereQuery+
		orderByQuery+
		" LIMIT $1 OFFSET $2", intLimit, intSkip)
	if err != nil {
		return []User{}, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var profile User
		if err := rows.Scan(
			&profile.UserID,
			&keycloakId,
			&profile.UpdatedAt,
			&profile.CreatedAt,
			&profile.Deleted,
			&profile.UserInput.MembershipActive,
			&profile.UserInput.MembershipType,
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
			return []User{}, fmt.Errorf("rows.Scan: %w", err)
		}

		// Add keycloakID and UserID to user struct
		keyCloakUUID, err := uuid.FromString(keycloakId)
		profile.UserInput.KeycloakID = &keyCloakUUID

		if err != nil {
			return []User{}, fmt.Errorf("uuid.FromString: %w", err)
		}

		users = append(users, profile)
	}
	if err := rows.Err(); err != nil {
		return []User{}, fmt.Errorf("rows.Err: %w", err)
	}

	return users, nil
}

func (db *ProfileDB) FetchProfileBasedOnPhoneNumber(ctx context.Context, phoneNumber string) (User, error) {

	var profile User
	var keycloakId string

	type phone struct {
		userID    uuid.UUID
		number    *string
		phoneType string
	}

	var phoneNum phone
	if err := db.QueryRow(ctx, `SELECT user_id, phone_number, type FROM phone_numbers WHERE phone_number=$1`, phoneNumber).
		Scan(&phoneNum.userID, &phoneNum.number, &phoneNum.phoneType); err != nil {
		if err == pgx.ErrNoRows {
			return User{}, common.ErrUserNotFound
		}
		return User{}, fmt.Errorf("db.QueryRow [phone]: %w", err)
	}

	var mobileNumber, whatsAppNumber, telegramNumber *string

	switch phoneNum.phoneType {
	case common.Mobile:
		mobileNumber = phoneNum.number
	case common.WhatsApp:
		whatsAppNumber = phoneNum.number
	case common.Telegram:
		telegramNumber = phoneNum.number
	}

	profile.UserInput.Phones = Phones{
		MobileNumber:   mobileNumber,
		WhatsAppNumber: whatsAppNumber,
		TelegramNumber: telegramNumber,
	}

	if err := db.QueryRow(ctx, `
		SELECT users.user_id,
		keycloak_id,
		users.updated_at,
		users.created_at,
		deleted,
		membership.active,
		membership.type,
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
		LEFT JOIN membership ON users.user_id = membership.user_id
		WHERE users.user_id = $1
		AND deleted = false`, phoneNum.userID).Scan(
		&profile.UserID,
		&keycloakId,
		&profile.UpdatedAt,
		&profile.CreatedAt,
		&profile.Deleted,
		&profile.UserInput.MembershipActive,
		&profile.UserInput.MembershipType,
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
	); err != nil {
		if err == pgx.ErrNoRows {
			return User{}, common.ErrUserNotFound
		}
		return User{}, fmt.Errorf("db.QueryRow [user]: %w", err)
	}

	keyCloakUUID, err := uuid.FromString(keycloakId)
	if err != nil {
		return User{}, fmt.Errorf("uuid.FromString: %w", err)
	}

	// Add keycloakID to the user
	profile.UserInput.KeycloakID = &keyCloakUUID

	return profile, nil
}

func buildAndGetWhereUserQuery(country string, email string, name string, tenGroupName string, language string,
	firstLanguage string, otherLanguageOne string, otherLanguageTwo string, otherLanguageThree string,
	otherLanguageFour string, updatedAt string, createdAt string, membership string, membershipType string,
	convention string, ticket string, galaxy string, gender string, checkAlternativeEmails bool) (string, string) {

	var whereString strings.Builder
	var orderBy strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

	// WHERE query generation based on parameters
	if country != "" {
		whereCondition.WriteString(fmt.Sprintf(" LOWER(country)=LOWER('%s')", country))
	}

	if email != "" {
		if checkAlternativeEmails {
			if whereCondition.String() != "" {
				whereCondition.WriteString(fmt.Sprintf(" AND (LOWER(primary_email) LIKE LOWER('%%%s%%') OR LOWER(alternate_email_1) LIKE LOWER('%%%s%%') OR LOWER(alternate_email_2) LIKE LOWER('%%%s%%'))", email))
			} else {
				whereCondition.WriteString(fmt.Sprintf(" (LOWER(primary_email) LIKE LOWER('%%%s%%') OR LOWER(alternate_email_1) LIKE LOWER('%%%s%%') OR LOWER(alternate_email_2) LIKE LOWER('%%%s%%'))", email))
			}
		} else {
			if whereCondition.String() != "" {
				whereCondition.WriteString(fmt.Sprintf(" AND LOWER(primary_email) LIKE LOWER('%%%s%%')", email))
			} else {
				whereCondition.WriteString(fmt.Sprintf(" LOWER(primary_email) LIKE LOWER('%%%s%%')", email))
			}
		}
	}

	if name != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND (LOWER(first_name_vernacular) LIKE LOWER('%%%s%%') OR LOWER(last_name_vernacular) LIKE LOWER('%%%s%%'))", name, name))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" (LOWER(first_name_vernacular) LIKE LOWER('%%%s%%') OR LOWER(last_name_vernacular) LIKE LOWER('%%%s%%'))", name, name))
		}
	}

	if tenGroupName != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND LOWER(name_of_ten_group)=LOWER('%s')", tenGroupName))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" LOWER(name_of_ten_group)=LOWER('%s')", tenGroupName))
		}
	}

	if gender != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND gender='%s'", gender))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" gender='%s'", gender))
		}
	}

	if language != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND (first_language='%s' OR other_language_1='%s' OR other_language_2='%s' OR other_language_3='%s' OR other_language_4='%s')", language, language, language, language, language))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" (first_language='%s' OR other_language_1='%s' OR other_language_2='%s' OR other_language_3='%s' OR other_language_4='%s')", language, language, language, language, language))
		}
	}

	if firstLanguage != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND first_language='%s'", firstLanguage))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" first_language='%s'", firstLanguage))
		}
	}

	if otherLanguageOne != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND other_language_1='%s'", otherLanguageOne))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" other_language_1='%s'", otherLanguageOne))
		}
	}

	if otherLanguageTwo != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND other_language_2='%s'", otherLanguageTwo))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" other_language_2='%s'", otherLanguageTwo))
		}
	}

	if otherLanguageThree != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND other_language_3='%s'", otherLanguageThree))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" other_language_3='%s'", otherLanguageThree))
		}
	}

	if otherLanguageFour != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND other_language_4='%s'", otherLanguageFour))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" other_language_4='%s'", otherLanguageFour))
		}
	}
	if membership != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND status.membership=%s", membership))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" status.membership=%s", membership))
		}
	}
	if membershipType != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND LOWER(status.membership_type)=LOWER('%s')", membershipType))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" LOWER(status.membership_type)=LOWER('%s')", membershipType))
		}
	}
	if convention != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND status.convention=%s", convention))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" status.convention=%s", convention))
		}
	}
	if ticket != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND status.ticket=%s", ticket))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" status.ticket=%s", ticket))
		}
	}
	if galaxy != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND status.galaxy=%s", galaxy))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" status.galaxy=%s", galaxy))
		}
	}
	if updatedAt != "" {
		if strings.ToLower(updatedAt) != "desc" && strings.ToLower(updatedAt) != "asc" {
			updatedAt = "asc"
		}
		orderBy.WriteString(fmt.Sprintf(" ORDER BY updated_at %s", updatedAt))
	} else if createdAt != "" {
		if strings.ToLower(createdAt) == "" || (strings.ToLower(createdAt) != "desc" && strings.ToLower(createdAt) != "asc") {
			createdAt = "asc"
		}
		orderBy.WriteString(fmt.Sprintf(" ORDER BY created_at %s", createdAt))
	} else {
		orderBy.WriteString(" ORDER BY user_id")
	}

	if whereCondition.String() != "" {
		whereString.WriteString(whereCondition.String())
	} else {
		whereString.Reset()
	}
	return whereString.String(), orderBy.String()
}
