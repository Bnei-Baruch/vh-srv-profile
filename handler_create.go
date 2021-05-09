package main

type createUser struct {
	KeycloakID          string  `json:"keycloak_id" binding:"required"`
	FirstNameLatin      string  `json:"first_name_latin" binding:"required"`
	FirstNameVernacular *string `json:"first_name_vernacular"`
	LastNameLatin       string  `json:"last_name_latin" binding:"required"`
	LastNameVernacular  *string `json:"last_name_vernacular"`
	StreetAddress       *string `json:"street_address"`
	Country             *string `json:"country"`
	StateOrRegion       *string `json:"state_region"`
	PostalCode          *string `json:"postal_code"`
	City                *string `json:"city"`
	Gender              string  `json:"gender" binding:"required"`
	MaritalStatus       string  `json:"marital_status" binding:"required"`
	DateOfBirth         *string `json:"date_of_birth"`
	PrimaryEmail        string  `json:"primary_email" binding:"required"`
	AlternateEmail1     *string `json:"alternate_email_1"`
	AlternateEmail2     *string `json:"alternate_email_2"`
	MobileNumber        string  `json:"mobile_number" binding:"required"`
	WhatsAppNumber      *string `json:"whats_app_number"`
	TelegramNumber      *string `json:"telegram_number"`
	FirstLanguage       string  `json:"first_language" binding:"required"`
	OtherLanguage1      *string `json:"other_language_1"`
	OtherLanguage2      *string `json:"other_language_2"`
	OtherLanguage3      *string `json:"other_language_3"`
	OtherLanguage4      *string `json:"other_language_4"`
	ListeningLanguage   string  `json:"listening_language" binding:"required"`
	ReadingLanguage     string  `json:"reading_language" binding:"required"`
	EmailLanguage       string  `json:"email_language" binding:"required"`
	StudyStartYear      int     `json:"study_start_year" binding:"required"`
	StudyFramework      string  `json:"study_framework" binding:"required"`
	HasGroup            bool    `json:"has_ten_group" binding:"required"`
	WantsGroup          *bool   `json:"wants_ten_group"`
	NameOfGroup         *bool   `json:"name_ten_group"`
}
