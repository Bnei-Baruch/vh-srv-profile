package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	uuid "github.com/satori/go.uuid"
)

type user struct {
	updatedAt time.Time
	createdAt time.Time
	deleted   bool
	userInput userInput
}

type userInput struct {
	keycloakID          *uuid.UUID
	firstNameLatin      *string
	firstNameVernacular *string
	lastNameLatin       *string
	lastNameVernacular  *string
	address             address
	gender              *string
	maritalStatus       *string
	dateOfBirth         *string
	emails              emails
	phones              phones
	languages           languages
	studyStartYear      *int
	studyFramework      *string
	ten                 ten
}

type address struct {
	streetAddress *string
	country       *string
	stateOrRegion *string
	postalCode    *string
	city          *string
}

type emails struct {
	primary    *string
	alternate1 *string
	alternate2 *string
}

type phones struct {
	mobileNumber   *string
	whatsAppNumber *string
	telegramNumber *string
}

type languages struct {
	first     *string
	other1    *string
	other2    *string
	other3    *string
	other4    *string
	listening *string
	reading   *string
	email     *string
}

type ten struct {
	hasGroup    *bool
	wantsGroup  *bool
	nameOfGroup *bool
}

const (
	mobile   = "mobile"
	whatsApp = "WhatsApp"
	telegram = "Telegram"
)

type pgProfileDB struct {
	*pgxpool.Pool
}

func newPgProfileDB(ctx context.Context, databaseURL string) (*pgProfileDB, error) {
	pool, err := pgxpool.Connect(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &pgProfileDB{pool}, nil
}

func (db *pgProfileDB) createProfile(ctx context.Context, user userInput) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

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
		user.keycloakID,
		user.firstNameLatin,
		user.firstNameVernacular,
		user.lastNameLatin,
		user.lastNameVernacular,
		user.address.streetAddress,
		user.address.country,
		user.address.stateOrRegion,
		user.address.postalCode,
		user.address.city,
		user.gender,
		user.maritalStatus,
		user.dateOfBirth,
		user.emails.primary,
		user.emails.alternate1,
		user.emails.alternate2,
		user.languages.first,
		user.languages.other1,
		user.languages.other2,
		user.languages.other3,
		user.languages.other4,
		user.languages.listening,
		user.languages.reading,
		user.languages.email,
		user.studyStartYear,
		user.studyFramework,
		user.ten.hasGroup,
		user.ten.wantsGroup,
		user.ten.nameOfGroup).Scan(&userID); err != nil {
		return err
	}

	if user.phones.mobileNumber != nil {
		if err := insertPhone(tx, userID, *user.phones.mobileNumber, mobile); err != nil {
			return err
		}
	}

	if user.phones.whatsAppNumber != nil {
		if err := insertPhone(tx, userID, *user.phones.whatsAppNumber, whatsApp); err != nil {
			return err
		}
	}

	if user.phones.telegramNumber != nil {
		if err := insertPhone(tx, userID, *user.phones.telegramNumber, telegram); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func insertPhone(tx pgx.Tx, userID uuid.UUID, number string, phoneType string) error {
	_, err := tx.Exec(context.Background(), `INSERT INTO phone_numbers (user_id, phone_number, type) VALUES ($1, $2, $3)`,
		userID, number, phoneType)
	return err
}

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
