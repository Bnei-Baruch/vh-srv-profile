package main

import (
	"context"
	"fmt"
	"strings"

	uuid "github.com/satori/go.uuid"
)

func (db *pgProfileDB) updateProfile(ctx context.Context, keycloakID uuid.UUID, user userInput) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT user_id FROM users WHERE keycloak_id=$1`, keycloakID).Scan(&userID); err != nil {
		return fmt.Errorf("problem finding profile for keycloak id %q: %w", keycloakID, err)
	}

	usersToUpdate, usersArgs := prepareUserUpdate(user)
	if len(usersArgs) != 0 {
		if _, err = tx.Exec(ctx, fmt.Sprintf(`UPDATE users SET %s WHERE user_id='%s'`, usersToUpdate, userID),
			usersArgs...); err != nil {
			return fmt.Errorf("problem updating users: %w", err)
		}
	}

	phonesToUpdate, phoneArgs := preparePhoneUpdate(userID, user)
	if len(phoneArgs) != 0 {
		if _, err = tx.Exec(ctx, fmt.Sprintf(`
		INSERT INTO phone_numbers (user_id, phone_number, type)
		VALUES %s
		ON CONFLICT (user_id, type)
		DO 
			UPDATE SET phone_number=EXCLUDED.phone_number`, phonesToUpdate),
			phoneArgs...); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func prepareUserUpdate(user userInput) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if user.firstNameLatin != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("first_name_latin=$%d", len(updateStrings)+1))
		args = append(args, user.firstNameLatin)
	}
	if user.firstNameVernacular != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("first_name_vernacular=$%d", len(updateStrings)+1))
		args = append(args, user.firstNameVernacular)
	}
	if user.lastNameLatin != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("last_name_latin=$%d", len(updateStrings)+1))
		args = append(args, user.lastNameLatin)
	}
	if user.lastNameVernacular != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("last_name_vernacular=$%d", len(updateStrings)+1))
		args = append(args, user.lastNameVernacular)
	}
	if user.address.streetAddress != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("street_address=$%d", len(updateStrings)+1))
		args = append(args, user.address.streetAddress)
	}
	if user.address.country != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("country=$%d", len(updateStrings)+1))
		args = append(args, user.address.country)
	}
	if user.address.stateOrRegion != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("state_region=$%d", len(updateStrings)+1))
		args = append(args, user.address.stateOrRegion)
	}
	if user.address.postalCode != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("postal_code=$%d", len(updateStrings)+1))
		args = append(args, user.address.postalCode)
	}
	if user.address.city != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("city=$%d", len(updateStrings)+1))
		args = append(args, user.address.city)
	}
	if user.gender != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("gender=$%d", len(updateStrings)+1))
		args = append(args, user.gender)
	}
	if user.maritalStatus != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("marital_status=$%d", len(updateStrings)+1))
		args = append(args, user.maritalStatus)
	}
	if user.dateOfBirth != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("date_of_birth=$%d", len(updateStrings)+1))
		args = append(args, user.dateOfBirth)
	}
	if user.emails.primary != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("primary_email=$%d", len(updateStrings)+1))
		args = append(args, user.emails.primary)
	}
	if user.emails.alternate1 != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("alternate_email_1=$%d", len(updateStrings)+1))
		args = append(args, user.emails.alternate1)
	}
	if user.emails.alternate2 != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("alternate_email_2=$%d", len(updateStrings)+1))
		args = append(args, user.emails.alternate2)
	}
	if user.languages.first != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("first_language=$%d", len(updateStrings)+1))
		args = append(args, user.languages.first)
	}
	if user.languages.other1 != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("other_language_1=$%d", len(updateStrings)+1))
		args = append(args, user.languages.other1)
	}
	if user.languages.other2 != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("other_language_2=$%d", len(updateStrings)+1))
		args = append(args, user.languages.other2)
	}
	if user.languages.other3 != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("other_language_3=$%d", len(updateStrings)+1))
		args = append(args, user.languages.other3)
	}
	if user.languages.other4 != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("other_language_4=$%d", len(updateStrings)+1))
		args = append(args, user.languages.other4)
	}
	if user.languages.listening != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("listening_language=$%d", len(updateStrings)+1))
		args = append(args, user.languages.listening)
	}
	if user.languages.reading != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("reading_language=$%d", len(updateStrings)+1))
		args = append(args, user.languages.reading)
	}
	if user.languages.email != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("email_language=$%d", len(updateStrings)+1))
		args = append(args, user.languages.email)
	}
	if user.studyStartYear != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("study_start_year=$%d", len(updateStrings)+1))
		args = append(args, user.studyStartYear)
	}
	if user.studyFramework != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("study_framework=$%d", len(updateStrings)+1))
		args = append(args, user.studyFramework)
	}
	if user.ten.hasGroup != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("has_ten_group=$%d", len(updateStrings)+1))
		args = append(args, user.ten.hasGroup)
	}
	if user.ten.wantsGroup != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("wants_ten_group=$%d", len(updateStrings)+1))
		args = append(args, user.ten.wantsGroup)
	}
	if user.ten.nameOfGroup != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("name_of_ten_group=$%d", len(updateStrings)+1))
		args = append(args, user.ten.nameOfGroup)
	}
	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func preparePhoneUpdate(userID uuid.UUID, user userInput) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if user.phones.mobileNumber != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("('%s', $%d, '%s')", userID, len(updateStrings)+1, mobile))
		args = append(args, user.phones.mobileNumber)
	}
	if user.phones.whatsAppNumber != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("('%s', $%d, '%s')", userID, len(updateStrings)+1, whatsApp))
		args = append(args, user.phones.whatsAppNumber)
	}
	if user.phones.telegramNumber != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("('%s', $%d, '%s')", userID, len(updateStrings)+1, telegram))
		args = append(args, user.phones.telegramNumber)
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}
