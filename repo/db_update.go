package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/events"
)

type updateStorage interface {
	UpdateProfile(ctx context.Context, keycloakID uuid.UUID, toUpdate UserInput) error
	SetSpouse(ctx context.Context, keycloakID1, keycloakID2 uuid.UUID, forceUpdate bool) error
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

	if c := userUpdateCols(user); !c.empty() {
		q := `UPDATE users SET ` + c.assignments() + ` WHERE user_id=` + c.bind(userID)
		if _, err = tx.Exec(ctx, q, c.args...); err != nil {
			return fmt.Errorf("tx.Exec [user]: %w", err)
		}
	}

	for _, p := range []struct {
		number    *string
		phoneType string
	}{
		{user.Phones.MobileNumber, common.Mobile},
		{user.Phones.WhatsAppNumber, common.WhatsApp},
		{user.Phones.TelegramNumber, common.Telegram},
	} {
		if p.number == nil {
			continue
		}
		if _, err = tx.Exec(ctx, `
		INSERT INTO phone_numbers (user_id, phone_number, type)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, type)
		DO UPDATE SET phone_number=EXCLUDED.phone_number`,
			userID, *p.number, p.phoneType); err != nil {
			return fmt.Errorf("tx.Exec [phone]: %w", err)
		}
	}

	/* Update status of the user in the status table if provided in the request */
	if c := userStatusUpdateCols(user); !c.empty() {
		q := `UPDATE status SET ` + c.assignments() + ` WHERE user_id=` + c.bind(userID)
		updateRes, err := tx.Exec(ctx, q, c.args...)
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

	db.emitEvent(ctx, events.TypeUpdateProfile, map[string]interface{}{
		"keycloak_id": keycloakID,
		"user_id":     userID,
	})

	return nil
}

// userUpdateCols lists the users columns set in user.
func userUpdateCols(user UserInput) sqlCols {
	var c sqlCols
	setIf(&c, "first_name_latin", user.FirstNameLatin)
	setIf(&c, "first_name_vernacular", user.FirstNameVernacular)
	setIf(&c, "last_name_latin", user.LastNameLatin)
	setIf(&c, "last_name_vernacular", user.LastNameVernacular)
	setIf(&c, "street_address", user.Address.StreetAddress)
	setIf(&c, "country", user.Address.Country)
	setIf(&c, "state_region", user.Address.StateOrRegion)
	setIf(&c, "postal_code", user.Address.PostalCode)
	setIf(&c, "city", user.Address.City)
	setIf(&c, "gender", user.Gender)
	if user.MaritalStatus.Set {
		if user.MaritalStatus.Valid {
			c.set("marital_status", user.MaritalStatus)
		} else {
			// Clear marital status, set it NULL.
			c.set("marital_status", nil)
		}
	}
	setIf(&c, "date_of_birth", user.DateOfBirth)
	setIf(&c, "primary_email", user.Emails.Primary)
	setIf(&c, "alternate_email_1", user.Emails.Alternate1)
	setIf(&c, "alternate_email_2", user.Emails.Alternate2)
	setIf(&c, "first_language", user.Languages.First)
	setIf(&c, "other_language_1", user.Languages.Other1)
	setIf(&c, "other_language_2", user.Languages.Other2)
	setIf(&c, "other_language_3", user.Languages.Other3)
	setIf(&c, "other_language_4", user.Languages.Other4)
	setIf(&c, "listening_language", user.Languages.Listening)
	setIf(&c, "reading_language", user.Languages.Reading)
	setIf(&c, "email_language", user.Languages.Email)
	setIf(&c, "study_start_year", user.StudyStartYear)
	setIf(&c, "study_framework", user.StudyFramework)
	setIf(&c, "has_ten_group", user.Ten.HasGroup)
	setIf(&c, "wants_ten_group", user.Ten.WantsGroup)
	setIf(&c, "name_of_ten_group", user.Ten.NameOfGroup)
	return c
}

// userStatusUpdateCols lists the status columns set in user.Status.
func userStatusUpdateCols(user UserInput) sqlCols {
	var c sqlCols
	setIf(&c, "membership", user.Status.Membership)
	setIf(&c, "membership_type", user.Status.MembershipType)
	setIf(&c, "convention", user.Status.Convention)
	setIf(&c, "galaxy", user.Status.Galaxy)
	setIf(&c, "ticket", user.Status.Ticket)
	return c
}

// SetSpouse manages symmetric spouse relationships and marital status.
// Purpose: Establishes/breaks spouse links (spouse_keycloak_id) and updates marital_status (Married/Single).
// - Prevents self-spousing; uses transactions and `FOR UPDATE` for data integrity.
// - On unset (keycloakID2=uuid.Nil), clears spouse links and sets marital_status to 'Single' for affected parties.
// - On set (keycloakID2!=uuid.Nil), links spouses, sets marital_status to 'Married' for both.
// - Conflicts (existing spouses) are checked; forceUpdate can bypass.
// - Unlinks any prior spouses of keycloakID1/keycloakID2, setting their marital_status to 'Single'.
// - Returns errors for non-existent users.
func (db *ProfileDB) SetSpouse(ctx context.Context, keycloakID1, keycloakID2 uuid.UUID, forceUpdate bool) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("db.Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	affectedKeycloakStringIDs, err := db.SetSpouseWithTx(ctx, tx, keycloakID1, keycloakID2, forceUpdate)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	db.emitUpdateProfileEvents(ctx, affectedKeycloakStringIDs)
	return nil
}

func (db *ProfileDB) SetSpouseWithTx(ctx context.Context, tx pgx.Tx, keycloakID1, keycloakID2 uuid.UUID, forceUpdate bool) ([]string, error) {
	if keycloakID1 == keycloakID2 {
		return nil, common.ErrSpouseSelf
	}

	var affectedKeycloakStringIDs []string

	// Get user 1's current spouse
	var spouse1_id *string
	err := tx.QueryRow(ctx, "SELECT spouse_keycloak_id FROM users WHERE keycloak_id = $1 FOR UPDATE", keycloakID1).Scan(&spouse1_id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user %s: %w", keycloakID1, common.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to get user 1: %w", err)
	}

	// Handle cancellation
	// TODO: Marital status handling after cancellation needs better logic.
	// Currently defaults to 'Divorced', but we should consider the previous state
	// and whether the user had a spouse to cancel or not.
	if keycloakID2 == uuid.Nil {
		if spouse1_id != nil {
			// Unlink user 1 and set marital status to Divorced
			_, err := tx.Exec(ctx, "UPDATE users SET spouse_keycloak_id = NULL, marital_status = $1 WHERE keycloak_id = $2", common.MaritalStatusDivorced, keycloakID1)
			if err != nil {
				return nil, err
			}
			affectedKeycloakStringIDs = append(affectedKeycloakStringIDs, keycloakID1.String())
			// Unlink old spouse and set marital status to Divorced
			_, err = tx.Exec(ctx, "UPDATE users SET spouse_keycloak_id = NULL, marital_status = $1 WHERE keycloak_id = $2", common.MaritalStatusDivorced, *spouse1_id)
			if err != nil {
				return nil, err
			}
			affectedKeycloakStringIDs = append(affectedKeycloakStringIDs, *spouse1_id)
		} else { // If there was no spouse before, just set current user to Divorced if they were not already.
			res, err := tx.Exec(ctx, "UPDATE users SET marital_status = $1 WHERE keycloak_id = $2 AND marital_status != $1", common.MaritalStatusDivorced, keycloakID1)
			if err != nil {
				return nil, err
			}
			if res.RowsAffected() > 0 {
				affectedKeycloakStringIDs = append(affectedKeycloakStringIDs, keycloakID1.String())
			}
		}
		return affectedKeycloakStringIDs, nil
	}

	// Get user 2's current spouse
	var spouse2_id *string
	err = tx.QueryRow(ctx, "SELECT spouse_keycloak_id FROM users WHERE keycloak_id = $1 FOR UPDATE", keycloakID2).Scan(&spouse2_id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user %s: %w", keycloakID2, common.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to get user 2: %w", err)
	}

	// Conflict check for non-forced updates
	if !forceUpdate {
		isUser1Ok := spouse1_id == nil || *spouse1_id == keycloakID2.String()
		isUser2Ok := spouse2_id == nil || *spouse2_id == keycloakID1.String()
		if !isUser1Ok || !isUser2Ok {
			return nil, common.ErrSpouseConflict
		}
	}

	// Unlink previous spouses and set their marital status to Divorced, if they were indeed spouses.
	if spouse1_id != nil && *spouse1_id != keycloakID2.String() {
		// Only unlink if current spouse is not the new spouse
		_, err := tx.Exec(ctx, "UPDATE users SET spouse_keycloak_id = NULL, marital_status = $1 WHERE keycloak_id = $2", common.MaritalStatusDivorced, *spouse1_id)
		if err != nil {
			return nil, err
		}
		affectedKeycloakStringIDs = append(affectedKeycloakStringIDs, *spouse1_id)
	}
	if spouse2_id != nil && *spouse2_id != keycloakID1.String() {
		// Only unlink if current spouse is not the new spouse
		_, err := tx.Exec(ctx, "UPDATE users SET spouse_keycloak_id = NULL, marital_status = $1 WHERE keycloak_id = $2", common.MaritalStatusDivorced, *spouse2_id)
		if err != nil {
			return nil, err
		}
		affectedKeycloakStringIDs = append(affectedKeycloakStringIDs, *spouse2_id)
	}

	// Set new spouse relationship and marital status to Married
	s1 := keycloakID1.String()
	s2 := keycloakID2.String()
	_, err = tx.Exec(ctx, "UPDATE users SET spouse_keycloak_id = $1, marital_status = $2 WHERE keycloak_id = $3", s2, common.MaritalStatusMarried, keycloakID1)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, "UPDATE users SET spouse_keycloak_id = $1, marital_status = $2 WHERE keycloak_id = $3", s1, common.MaritalStatusMarried, keycloakID2)
	if err != nil {
		return nil, err
	}
	affectedKeycloakStringIDs = append(affectedKeycloakStringIDs, s1, s2)

	return affectedKeycloakStringIDs, nil
}
