package main

import "fmt"

var (
	errProfileNotFound = fmt.Errorf("no profile found for keycloak id")
)

type profileManager struct {
	creator     createStorage
	getter      readStorage
	updater     updateStorage
	deleter     deleteStorage
	hardDeleter hardDeleteStorage
}

type profileRequest struct {
	KeycloakID          *string `json:"keycloak_id"`
	FirstNameLatin      *string `json:"first_name_latin"`
	FirstNameVernacular *string `json:"first_name_vernacular"`
	LastNameLatin       *string `json:"last_name_latin"`
	LastNameVernacular  *string `json:"last_name_vernacular"`
	StreetAddress       *string `json:"street_address"`
	Country             *string `json:"country"`
	StateOrRegion       *string `json:"state_region"`
	PostalCode          *string `json:"postal_code"`
	City                *string `json:"city"`
	Gender              *string `json:"gender"`
	MaritalStatus       *string `json:"marital_status"`
	DateOfBirth         *string `json:"date_of_birth"`
	PrimaryEmail        *string `json:"primary_email"`
	AlternateEmail1     *string `json:"alternate_email_1"`
	AlternateEmail2     *string `json:"alternate_email_2"`
	MobileNumber        *string `json:"mobile_number"`
	WhatsAppNumber      *string `json:"whats_app_number"`
	TelegramNumber      *string `json:"telegram_number"`
	FirstLanguage       *string `json:"first_language"`
	OtherLanguage1      *string `json:"other_language_1"`
	OtherLanguage2      *string `json:"other_language_2"`
	OtherLanguage3      *string `json:"other_language_3"`
	OtherLanguage4      *string `json:"other_language_4"`
	ListeningLanguage   *string `json:"listening_language"`
	ReadingLanguage     *string `json:"reading_language"`
	EmailLanguage       *string `json:"email_language"`
	StudyStartYear      *int    `json:"study_start_year"`
	StudyFramework      *string `json:"study_framework"`
	HasGroup            *bool   `json:"has_ten_group"`
	WantsGroup          *bool   `json:"wants_ten_group"`
	NameOfGroup         *string `json:"name_ten_group"`
}
