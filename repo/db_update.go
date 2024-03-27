package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

type updateStorage interface {
	UpdateProfile(ctx context.Context, keycloakID uuid.UUID, toUpdate UserInput) error
}

func (db *ProfileDB) UpdateProfile(ctx context.Context, keycloakID uuid.UUID, user UserInput) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("db.Begin: %w", err)
	}
	defer func() error {
		if err := tx.Rollback(ctx); err != nil {
			return fmt.Errorf("tx.Rollback: %w", err)
		}
		return nil
	}()

	var userID uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT user_id FROM users WHERE keycloak_id=$1 AND deleted=false`, keycloakID).Scan(&userID); err != nil {
		if err == pgx.ErrNoRows {
			return common.ErrProfileNotFound
		}
		return fmt.Errorf("tx.QueryRow: %w", err)
	}

	usersToUpdate, usersArgs := prepareUserUpdate(user)
	if len(usersArgs) != 0 {
		if _, err = tx.Exec(ctx, fmt.Sprintf(`UPDATE users SET %s WHERE user_id='%s'`, usersToUpdate, userID),
			usersArgs...); err != nil {
			return fmt.Errorf("tx.Exec [user]: %w", err)
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
			return fmt.Errorf("tx.Exec [phone]: %w", err)
		}
	}

	/* Update status of the user in the status table if provided in the request */
	userStatusToUpdate, userStatusArgs := prepareUserStatusUpdateQuery(user)

	if len(userStatusArgs) != 0 {
		updateRes, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE status SET %s WHERE user_id='%s'`, userStatusToUpdate, userID),
			userStatusArgs...)
		if err != nil {
			return fmt.Errorf("tx.Exec [status]: %w", err)
		}

		/* Add new status row for the user if 0 rows are affected i.e. user status is not present in status table */
		if updateRes.RowsAffected() == 0 {
			if err := insertUserMembershipStatus(ctx, tx, userID, user.Status.Membership, user.Status.MembershipType, user.Status.Ticket, user.Status.Convention, user.Status.Galaxy); err != nil {
				return fmt.Errorf("insertUserMembershipStatus: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func prepareUserUpdate(user UserInput) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if user.FirstNameLatin != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("first_name_latin=$%d", len(updateStrings)+1))
		args = append(args, user.FirstNameLatin)
	}
	if user.FirstNameVernacular != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("first_name_vernacular=$%d", len(updateStrings)+1))
		args = append(args, user.FirstNameVernacular)
	}
	if user.LastNameLatin != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("last_name_latin=$%d", len(updateStrings)+1))
		args = append(args, user.LastNameLatin)
	}
	if user.LastNameVernacular != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("last_name_vernacular=$%d", len(updateStrings)+1))
		args = append(args, user.LastNameVernacular)
	}
	if user.Address.StreetAddress != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("street_address=$%d", len(updateStrings)+1))
		args = append(args, user.Address.StreetAddress)
	}
	if user.Address.Country != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("country=$%d", len(updateStrings)+1))
		args = append(args, user.Address.Country)
	}
	if user.Address.StateOrRegion != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("state_region=$%d", len(updateStrings)+1))
		args = append(args, user.Address.StateOrRegion)
	}
	if user.Address.PostalCode != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("postal_code=$%d", len(updateStrings)+1))
		args = append(args, user.Address.PostalCode)
	}
	if user.Address.City != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("city=$%d", len(updateStrings)+1))
		args = append(args, user.Address.City)
	}
	if user.Gender != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("gender=$%d", len(updateStrings)+1))
		args = append(args, user.Gender)
	}
	if user.MaritalStatus != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("marital_status=$%d", len(updateStrings)+1))
		args = append(args, user.MaritalStatus)
	}
	if user.DateOfBirth != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("date_of_birth=$%d", len(updateStrings)+1))
		args = append(args, user.DateOfBirth)
	}
	if user.Emails.Primary != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("primary_email=$%d", len(updateStrings)+1))
		args = append(args, user.Emails.Primary)
	}
	if user.Emails.Alternate1 != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("alternate_email_1=$%d", len(updateStrings)+1))
		args = append(args, user.Emails.Alternate1)
	}
	if user.Emails.Alternate2 != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("alternate_email_2=$%d", len(updateStrings)+1))
		args = append(args, user.Emails.Alternate2)
	}
	if user.Languages.First != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("first_language=$%d", len(updateStrings)+1))
		args = append(args, user.Languages.First)
	}
	if user.Languages.Other1 != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("other_language_1=$%d", len(updateStrings)+1))
		args = append(args, user.Languages.Other1)
	}
	if user.Languages.Other2 != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("other_language_2=$%d", len(updateStrings)+1))
		args = append(args, user.Languages.Other2)
	}
	if user.Languages.Other3 != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("other_language_3=$%d", len(updateStrings)+1))
		args = append(args, user.Languages.Other3)
	}
	if user.Languages.Other4 != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("other_language_4=$%d", len(updateStrings)+1))
		args = append(args, user.Languages.Other4)
	}
	if user.Languages.Listening != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("listening_language=$%d", len(updateStrings)+1))
		args = append(args, user.Languages.Listening)
	}
	if user.Languages.Reading != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("reading_language=$%d", len(updateStrings)+1))
		args = append(args, user.Languages.Reading)
	}
	if user.Languages.Email != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("email_language=$%d", len(updateStrings)+1))
		args = append(args, user.Languages.Email)
	}
	if user.StudyStartYear != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("study_start_year=$%d", len(updateStrings)+1))
		args = append(args, user.StudyStartYear)
	}
	if user.StudyFramework != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("study_framework=$%d", len(updateStrings)+1))
		args = append(args, user.StudyFramework)
	}
	if user.Ten.HasGroup != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("has_ten_group=$%d", len(updateStrings)+1))
		args = append(args, user.Ten.HasGroup)
	}
	if user.Ten.WantsGroup != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("wants_ten_group=$%d", len(updateStrings)+1))
		args = append(args, user.Ten.WantsGroup)
	}
	if user.Ten.NameOfGroup != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("name_of_ten_group=$%d", len(updateStrings)+1))
		args = append(args, user.Ten.NameOfGroup)
	}
	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func preparePhoneUpdate(userID uuid.UUID, user UserInput) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if user.Phones.MobileNumber != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("('%s', $%d, '%s')", userID, len(updateStrings)+1, common.Mobile))
		args = append(args, user.Phones.MobileNumber)
	}
	if user.Phones.WhatsAppNumber != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("('%s', $%d, '%s')", userID, len(updateStrings)+1, common.WhatsApp))
		args = append(args, user.Phones.WhatsAppNumber)
	}
	if user.Phones.TelegramNumber != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("('%s', $%d, '%s')", userID, len(updateStrings)+1, common.Telegram))
		args = append(args, user.Phones.TelegramNumber)
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

/* status update query builder */
func prepareUserStatusUpdateQuery(user UserInput) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if user.Status.Membership != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("membership=$%d", len(updateStrings)+1))
		args = append(args, user.Status.Membership)
	}
	if user.Status.MembershipType != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("membership_type=$%d", len(updateStrings)+1))
		args = append(args, user.Status.MembershipType)
	}
	if user.Status.Convention != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("convention=$%d", len(updateStrings)+1))
		args = append(args, user.Status.Convention)
	}
	if user.Status.Galaxy != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("galaxy=$%d", len(updateStrings)+1))
		args = append(args, user.Status.Galaxy)
	}
	if user.Status.Ticket != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("ticket=$%d", len(updateStrings)+1))
		args = append(args, user.Status.Ticket)
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}
