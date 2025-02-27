package api

import (
	"time"

	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/keycloak"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

type ProfileManager struct {
	repo                   repo.ProfileRepository
	keycloakServiceFactory keycloak.KeycloakServiceFactory
	ordersService          orders.OrdersService
}

func NewProfileManager(db repo.ProfileRepository) *ProfileManager {
	return &ProfileManager{
		repo:                   db,
		keycloakServiceFactory: keycloak.KeycloakAPIFactory,
		ordersService:          orders.NewOrdersAPI(),
	}
}

func (p *ProfileManager) SetKeycloakServiceFactory(keycloakServiceFactory keycloak.KeycloakServiceFactory) {
	p.keycloakServiceFactory = keycloakServiceFactory
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
