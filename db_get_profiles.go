package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	uuid "github.com/satori/go.uuid"
)

// Fetch multiple profiles based in parameters provided ( If more than one parameters then AND operation will execute on them )
func (db *pgProfileDB) getMultipleProfiles(ctx context.Context, intSkip int, intLimit int, country string, email string, firstLastName string, tenGroupName string, language string, firstLanguage string, otherLanguageOne string, otherLanguageTwo string, otherLanguageThree string, otherLanguageFour string, updatedAt string, createdAt string, membership string, membershipType string, convention string, ticket string, galaxy string) ([]user, error) {
	var userID uuid.UUID
	var keycloakId string
	users := []user{}

	userDbWhereQuery, orderByQuery := buildAndGetWhereUserQuery(country, email, firstLastName, tenGroupName, language, firstLanguage, otherLanguageOne, otherLanguageTwo, otherLanguageThree, otherLanguageFour, updatedAt, createdAt, membership, membershipType, convention, ticket, galaxy)

	rows, err := db.Query(ctx, `
		SELECT users.user_id,
		keycloak_id,
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
		name_of_ten_group
	FROM users
	LEFT JOIN status ON users.user_id = status.user_id`+userDbWhereQuery+
		orderByQuery+
		" LIMIT $1 OFFSET $2", intLimit, intSkip)
	if err != nil {
		fmt.Println("--error-while-executing-query", err)
		return []user{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var profile user
		if err := rows.Scan(
			&userID,
			&keycloakId,
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
		); err != nil {
			return []user{}, err
		}

		// Add keycloakID and userID to user struct
		keyCloakUUID, err := uuid.FromString(keycloakId)
		profile.userInput.keycloakID = &keyCloakUUID
		profile.userID = &userID

		if err != nil {
			return []user{}, err
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
			fmt.Println("--error-while-executing-phone-num-query", err)
			return []user{}, err
		}

		for rows.Next() {
			var temp phone
			if err := rows.Scan(&temp.number, &temp.phoneType); err != nil {
				return []user{}, err
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

		users = append(users, profile)
	}

	// Manage if no user found
	if len(users) == 0 {
		return []user{}, fmt.Errorf("%w", errUserNotFound)
	}

	return users, nil
}

// Fetch user based on the mobile number provided
func (db *pgProfileDB) fetchProfileBasedOnPhoneNumber(ctx context.Context, phoneNumber string) (user, error) {

	var profile user
	var userID uuid.UUID
	var keycloakId string

	type phone struct {
		userID    uuid.UUID
		number    *string
		phoneType string
	}

	var phoneNum phone
	if err := db.QueryRow(ctx, `
	SELECT user_id,
	phone_number, 
		type 
	FROM phone_numbers
	WHERE phone_number=$1`, phoneNumber).Scan(
		&phoneNum.userID,
		&phoneNum.number,
		&phoneNum.phoneType,
	); err != nil {
		if err == pgx.ErrNoRows {
			return user{}, fmt.Errorf("%w", errUserNotFound)
		}
		return user{}, err
	}

	var mobileNumber, whatsAppNumber, telegramNumber *string

	switch phoneNum.phoneType {
	case mobile:
		mobileNumber = phoneNum.number
	case whatsApp:
		whatsAppNumber = phoneNum.number
	case telegram:
		telegramNumber = phoneNum.number
	}

	profile.userInput.phones = phones{
		mobileNumber:   mobileNumber,
		whatsAppNumber: whatsAppNumber,
		telegramNumber: telegramNumber,
	}

	if err := db.QueryRow(ctx, `
		SELECT users.user_id,
		keycloak_id,
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
		name_of_ten_group
		FROM users
		LEFT JOIN status ON users.user_id = status.user_id
		WHERE users.user_id = $1
		AND deleted = false`, phoneNum.userID).Scan(
		&profile.userID,
		&keycloakId,
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
	); err != nil {
		if err == pgx.ErrNoRows {
			return user{}, fmt.Errorf("%w: %q", errUserNotFound, userID)
		}
		return user{}, err
	}

	keyCloakUUID, err := uuid.FromString(keycloakId)

	if err != nil {
		return user{}, err
	}

	// Add keycloakID to the user
	profile.userInput.keycloakID = &keyCloakUUID

	return profile, nil
}

func buildAndGetWhereUserQuery(country string, email string, firstLastName string, tenGroupName string, language string, firstLanguage string, otherLanguageOne string, otherLanguageTwo string, otherLanguageThree string, otherLanguageFour string, updatedAt string, createdAt string, membership string, membershipType string, convention string, ticket string, galaxy string) (string, string) {

	var whereString strings.Builder
	var orderBy strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

	// WHERE query generation based on parameters
	if country != "" {
		whereCondition.WriteString(fmt.Sprintf(" country='%s'", country))
	}

	if email != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND primary_email LIKE '%%%s%%'", email))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" primary_email LIKE '%%%s%%'", email))
		}
	}

	if firstLastName != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND (first_name_latin LIKE '%%%s%%' OR last_name_latin LIKE '%%%s%%')", firstLastName, firstLastName))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" (first_name_latin LIKE '%%%s%%' OR last_name_latin LIKE '%%%s%%')", firstLastName, firstLastName))
		}
	}

	if tenGroupName != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND name_of_ten_group='%s'", tenGroupName))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" name_of_ten_group='%s'", tenGroupName))
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
			whereCondition.WriteString(fmt.Sprintf(" AND status.membership_type='%s'", membershipType))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" status.membership_type='%s'", membershipType))
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
	} else {
		if strings.ToLower(createdAt) == "" || (strings.ToLower(createdAt) != "desc" && strings.ToLower(createdAt) != "asc") {
			createdAt = "asc"
		}
		orderBy.WriteString(fmt.Sprintf(" ORDER BY created_at %s", createdAt))
	}

	if whereCondition.String() != "" {
		whereString.WriteString(whereCondition.String())
	} else {
		whereString.Reset()
	}
	return whereString.String(), orderBy.String()
}
