package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Nerzal/gocloak"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var (
	errProfileNotFound = fmt.Errorf("no profile found for keycloak id")
	errUserNotFound    = fmt.Errorf("no profile found")
	errNotFound        = fmt.Errorf("not found")
)

type profileManager struct {
	creator        createStorage
	requestCreator createRequestStorage
	getter         readStorage
	updater        updateStorage
	requestUpdater updateRequestStorage
	deleter        deleteStorage
	requestDeleter deleteRequestStorage
	hardDeleter    hardDeleteStorage
	fetchRequests  readMultipleRequestStorage
	fetchProfiles  readMultipleProfileStorage
}

type profileRequest struct {
	KeycloakID          *string `json:"keycloak_id"`
	FirstNameLatin      *string `json:"first_name_latin,omitempty"`
	FirstNameVernacular *string `json:"first_name_vernacular"`
	LastNameLatin       *string `json:"last_name_latin,omitempty"`
	LastNameVernacular  *string `json:"last_name_vernacular"`
	StreetAddress       *string `json:"street_address,omitempty"`
	Country             *string `json:"country,omitempty"`
	StateOrRegion       *string `json:"state_region,omitempty"`
	PostalCode          *string `json:"postal_code,omitempty"`
	City                *string `json:"city,omitempty"`
	Status              struct {
		UserID         *string `json:"user_id,omitempty"`
		Membership     *bool   `json:"membership,omitempty"`
		MembershipType *string `json:"membership_type,omitempty"`
		Ticket         *bool   `json:"ticket,omitempty"`
		Convention     *bool   `json:"convention,omitempty"`
		Galaxy         *bool   `json:"galaxy,omitempty"`
	} `json:"status,omitempty"`
	Gender            *string    `json:"gender,omitempty"`
	MaritalStatus     *string    `json:"marital_status,omitempty"`
	DateOfBirth       *time.Time `json:"date_of_birth,omitempty"`
	PrimaryEmail      *string    `json:"primary_email,omitempty"`
	AlternateEmail1   *string    `json:"alternate_email_1,omitempty"`
	AlternateEmail2   *string    `json:"alternate_email_2,omitempty"`
	MobileNumber      *string    `json:"mobile_number,omitempty"`
	WhatsAppNumber    *string    `json:"whats_app_number,omitempty"`
	TelegramNumber    *string    `json:"telegram_number,omitempty"`
	FirstLanguage     *string    `json:"first_language,omitempty"`
	OtherLanguage1    *string    `json:"other_language_1,omitempty"`
	OtherLanguage2    *string    `json:"other_language_2,omitempty"`
	OtherLanguage3    *string    `json:"other_language_3,omitempty"`
	OtherLanguage4    *string    `json:"other_language_4,omitempty"`
	ListeningLanguage *string    `json:"listening_language,omitempty"`
	ReadingLanguage   *string    `json:"reading_language,omitempty"`
	EmailLanguage     *string    `json:"email_language,omitempty"`
	StudyStartYear    *int       `json:"study_start_year,omitempty"`
	StudyFramework    *string    `json:"study_framework,omitempty"`
	HasGroup          *bool      `json:"has_ten_group,omitempty"`
	WantsGroup        *bool      `json:"wants_ten_group,omitempty"`
	NameOfGroup       *string    `json:"name_ten_group,omitempty"`
}

type newRequest struct {
	RequestName   *string `json:"name"`
	KeycloakId    *string `json:"keycloak_id"`
	Status        *string `json:"status"`
	EventSlug     *string `json:"event_slug"`
	Type          *string `json:"type"`
	RequestNote   *string `json:"request_note,omitempty"`
	RejectionNote *string `json:"rejection_note,omitempty"`
}

func SyncWithKeycloak(tokenString string, keycloakID string, firstName *string, lastName *string) error {

	var serverURL string
	var realm string
	if value, ok := os.LookupEnv("KEYCLOAK_SERVER_URL"); ok {
		serverURL = value
	}

	if realmValue, ok := os.LookupEnv("KEYCLOAK_REALM"); ok {
		realm = realmValue
	}

	if serverURL == "" || realm == "" {
		return fmt.Errorf("missing keycloak server url or realm")
	}

	tokenParts := strings.Split(tokenString, " ")

	validToken := tokenParts[1]

	if validToken == "" {
		return fmt.Errorf("no token found")
	}

	client := gocloak.NewClient(serverURL)

	keycloakUserInfo, infoErr := client.GetUserByID(validToken, realm, keycloakID)

	if infoErr != nil {
		return infoErr
	}

	if firstName == nil {
		firstName = &keycloakUserInfo.FirstName
	}

	if lastName == nil {
		lastName = &keycloakUserInfo.LastName
	}

	// only update the user if user details are not same
	if keycloakUserInfo.ID == keycloakID && keycloakUserInfo.FirstName == *firstName && keycloakUserInfo.LastName == *lastName {
		return nil
	}

	// Only update firtName & lastName
	updateObj := gocloak.User{
		ID:                         keycloakID,
		FirstName:                  *firstName,
		LastName:                   *lastName,
		CreatedTimestamp:           keycloakUserInfo.CreatedTimestamp,
		Username:                   keycloakUserInfo.Username,
		Enabled:                    keycloakUserInfo.Enabled,
		Totp:                       keycloakUserInfo.Totp,
		EmailVerified:              keycloakUserInfo.EmailVerified,
		Email:                      keycloakUserInfo.Email,
		FederationLink:             keycloakUserInfo.FederationLink,
		Attributes:                 keycloakUserInfo.Attributes,
		DisableableCredentialTypes: keycloakUserInfo.DisableableCredentialTypes,
		RequiredActions:            keycloakUserInfo.RequiredActions,
		Access:                     keycloakUserInfo.Access,
	}

	updaterErr := client.UpdateUser(validToken, realm, updateObj)

	if updaterErr != nil {
		return updaterErr
	}

	return nil
}

func SyncDBStructInsertionAndMigrations() error {
	m, err := migrate.New(
		"file://./db/migrations", makeDBURL()+"?sslmode=disable")
	if err != nil {
		if err != migrate.ErrNoChange {
			return nil
		}
	}
	// Syncing Table struct (UP Mig), Insertion ( Up Mig ) & UP Migrations
	if err := m.Up(); err != nil {
		m.Close()
		fmt.Println("UP Migration Done!")
		if err == migrate.ErrNoChange {
			return nil
		}
		return err
	}

	return nil
}
