package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	uuid "github.com/satori/go.uuid"
	"github.com/volatiletech/null/v9"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/events"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
)

type createStorage interface {
	CreateProfile(ctx context.Context, user UserInput) error
	CreateProfileWithCountryCheck(ctx context.Context, user UserInput) error
}

type User struct {
	UserID    *uuid.UUID
	UpdatedAt time.Time
	CreatedAt time.Time
	Deleted   bool
	UserInput UserInput
}

type UserInput struct {
	KeycloakID          *uuid.UUID
	FirstNameLatin      *string
	FirstNameVernacular *string
	LastNameLatin       *string
	LastNameVernacular  *string
	Address             Address
	Status              UserStatus // Deprecated: V1 membership status info.
	MembershipActive    *bool
	MembershipType      *string
	Gender              *string
	MaritalStatus       null.String
	SpouseKeycloakID    *string
	DateOfBirth         *time.Time
	Emails              Emails
	Phones              Phones
	Languages           Languages
	StudyStartYear      *int
	StudyFramework      *string
	Ten                 Ten
}

type UserStatus struct {
	UserID         *uuid.UUID
	Membership     *bool
	MembershipType *string
	Ticket         *bool
	Convention     *bool
	Galaxy         *bool
}

type Address struct {
	StreetAddress *string
	Country       *string
	StateOrRegion *string
	PostalCode    *string
	City          *string
}

type Emails struct {
	Primary    *string
	Alternate1 *string
	Alternate2 *string
}

type Phones struct {
	MobileNumber   *string
	WhatsAppNumber *string
	TelegramNumber *string
}

type Languages struct {
	First     *string
	Other1    *string
	Other2    *string
	Other3    *string
	Other4    *string
	Listening *string
	Reading   *string
	Email     *string
}

type Ten struct {
	HasGroup    *bool
	WantsGroup  *bool
	NameOfGroup *string
}

func (db *ProfileDB) CreateProfileWithCountryCheck(ctx context.Context, user UserInput) error {
	var countryCode *string = nil
	if user.Address.Country != nil {
		// The country value received from the 'orders' service may be either a country name or a country code.
		// Normalize it by looking up the corresponding code in the country_list table.
		err := db.QueryRow(ctx, `select code from country_list where name = $1 or code = $1`, *user.Address.Country).Scan(&countryCode)
		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("db.QueryRow: %w", err)
		}
	}
	user.Address.Country = countryCode
	return db.CreateProfile(ctx, user)
}

func (db *ProfileDB) CreateProfile(ctx context.Context, user UserInput) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("db.Begin: %w", err)
	}

	defer func() error {
		err := tx.Rollback(ctx)
		if err != nil {
			return fmt.Errorf("tx.Rollback: %w", err)
		}
		return nil
	}()

	var userID uuid.UUID
	if err := db.QueryRow(ctx, `
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
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24,
        $25, $26, $27, $28, $29)
	RETURNING user_id`,
		user.KeycloakID,
		user.FirstNameLatin,
		user.FirstNameVernacular,
		user.LastNameLatin,
		user.LastNameVernacular,
		user.Address.StreetAddress,
		user.Address.Country,
		user.Address.StateOrRegion,
		user.Address.PostalCode,
		user.Address.City,
		user.Gender,
		user.MaritalStatus,
		user.DateOfBirth,
		user.Emails.Primary,
		user.Emails.Alternate1,
		user.Emails.Alternate2,
		user.Languages.First,
		user.Languages.Other1,
		user.Languages.Other2,
		user.Languages.Other3,
		user.Languages.Other4,
		user.Languages.Listening,
		user.Languages.Reading,
		user.Languages.Email,
		user.StudyStartYear,
		user.StudyFramework,
		user.Ten.HasGroup,
		user.Ten.WantsGroup,
		user.Ten.NameOfGroup).Scan(&userID); err != nil {
		return fmt.Errorf("db.QueryRow: %w", err)
	}

	if user.Phones.MobileNumber != nil {
		if err := insertPhone(ctx, tx, userID, *user.Phones.MobileNumber, common.Mobile); err != nil {
			return fmt.Errorf("insertPhone [mobile]: %w", err)
		}
	}

	if user.Phones.WhatsAppNumber != nil {
		if err := insertPhone(ctx, tx, userID, *user.Phones.WhatsAppNumber, common.WhatsApp); err != nil {
			return fmt.Errorf("insertPhone [whatsapp]: %w", err)
		}
	}

	if user.Phones.TelegramNumber != nil {
		if err := insertPhone(ctx, tx, userID, *user.Phones.TelegramNumber, common.Telegram); err != nil {
			return fmt.Errorf("insertPhone [telegram]: %w", err)
		}
	}

	// TODO (edo): this is old. probably should remove
	if err := insertUserMembershipStatus(ctx, tx, userID, user.Status.Membership, user.Status.MembershipType,
		user.Status.Ticket, user.Status.Convention, user.Status.Galaxy); err != nil {
		return fmt.Errorf("insertUserMembershipStatus: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.commit: %w", err)
	}

	_, err = db.EvaluateMembershipByUserID(ctx,
		EmailKeycloakAndUserIDBody{UserID: utils.PointerString(userID.String())})
	if err != nil {
		return fmt.Errorf("db.EvaluateMembershipByUserID: %w", err)
	}

	db.emitEvent(ctx, events.TypeCreateProfile, map[string]interface{}{
		"user_id":     userID.String(),
		"keycloak_id": user.KeycloakID,
	})

	return nil
}

func insertPhone(ctx context.Context, tx pgx.Tx, userID uuid.UUID, number string, phoneType string) error {
	_, err := tx.Exec(ctx, `INSERT INTO phone_numbers (user_id, phone_number, type) VALUES ($1, $2, $3)`,
		userID, number, phoneType)
	return err
}

/* Function to insert status of user if provided else will insert default values */
func insertUserMembershipStatus(ctx context.Context, tx pgx.Tx, userID uuid.UUID, membership *bool,
	membershipType *string, ticket *bool, convention *bool, galaxy *bool) error {

	/* Setting default values */
	boolMembership := false
	boolTicket := false
	boolConvention := false
	boolGalaxy := false

	strMembershipType := "inactive"

	if membership != nil {
		boolMembership = *membership
	}

	if membershipType != nil {
		strMembershipType = *membershipType
	}

	if ticket != nil {
		boolTicket = *ticket
	}

	if convention != nil {
		boolConvention = *convention
	}

	if galaxy != nil {
		boolGalaxy = *galaxy
	}

	_, err := tx.Exec(ctx, `INSERT INTO status (user_id, membership, membership_type, ticket, convention, galaxy) VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, boolMembership, strMembershipType, boolTicket, boolConvention, boolGalaxy)
	return err
}
