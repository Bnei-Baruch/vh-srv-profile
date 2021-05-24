package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type readStorage interface {
	getProfile(ctx context.Context, keycloakID uuid.UUID) (user, error)
}

type userResponse struct {
	UpdatedAt           time.Time `json:"updated_at"`
	CreatedAt           time.Time `json:"created_at"`
	Deleted             bool      `json:"deleted"`
	FirstNameLatin      *string   `json:"first_name_latin,omitempty"`
	FirstNameVernacular *string   `json:"first_name_vernacular" `
	LastNameLatin       *string   `json:"last_name_latin,omitempty"`
	LastNameVernacular  *string   `json:"last_name_vernacular"  `
	StreetAddress       *string   `json:"street_address,omitempty"`
	Country             *string   `json:"country,omitempty"`
	StateOrRegion       *string   `json:"state_region,omitempty"`
	PostalCode          *string   `json:"postal_code,omitempty"`
	City                *string   `json:"city,omitempty"`
	Gender              *string   `json:"gender,omitempty"`
	MaritalStatus       *string   `json:"marital_status,omitempty"`
	DateOfBirth         *string   `json:"date_of_birth,omitempty"`
	PrimaryEmail        *string   `json:"primary_email" `
	AlternateEmail1     *string   `json:"alternate_email_1,omitempty"`
	AlternateEmail2     *string   `json:"alternate_email_2,omitempty"`
	MobileNumber        *string   `json:"mobile_number,omitempty"`
	WhatsAppNumber      *string   `json:"whats_app_number,omitempty"`
	TelegramNumber      *string   `json:"telegram_number,omitempty"`
	FirstLanguage       *string   `json:"first_language,omitempty"`
	OtherLanguage1      *string   `json:"other_language_1,omitempty"`
	OtherLanguage2      *string   `json:"other_language_2,omitempty"`
	OtherLanguage3      *string   `json:"other_language_3,omitempty"`
	OtherLanguage4      *string   `json:"other_language_4,omitempty"`
	ListeningLanguage   *string   `json:"listening_language,omitempty"`
	ReadingLanguage     *string   `json:"reading_language,omitempty"`
	EmailLanguage       *string   `json:"email_language,omitempty"`
	StudyStartYear      *int      `json:"study_start_year,omitempty"`
	StudyFramework      *string   `json:"study_framework,omitempty"`
	HasGroup            *bool     `json:"has_ten_group,omitempty"`
	WantsGroup          *bool     `json:"wants_ten_group,omitempty"`
	NameOfGroup         *bool     `json:"name_ten_group,omitempty"`
}

func (p *profileManager) get(c *gin.Context) {
	keycloakIDString, ok := c.Params.Get("keycloak_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keycloak ID missing"})
		return
	}
	keycloakID, err := uuid.FromString(keycloakIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	profile, err := p.getter.getProfile(c.Request.Context(), keycloakID)
	if err != nil {
		if errors.Is(err, errProfileNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while getting user %q: %w", keycloakIDString, err))
		return
	}

	result := userResponse{
		UpdatedAt:           profile.updatedAt,
		CreatedAt:           profile.createdAt,
		Deleted:             profile.deleted,
		FirstNameLatin:      profile.userInput.firstNameLatin,
		FirstNameVernacular: profile.userInput.firstNameVernacular,
		LastNameLatin:       profile.userInput.lastNameLatin,
		LastNameVernacular:  profile.userInput.lastNameVernacular,
		StreetAddress:       profile.userInput.address.streetAddress,
		Country:             profile.userInput.address.country,
		StateOrRegion:       profile.userInput.address.stateOrRegion,
		PostalCode:          profile.userInput.address.postalCode,
		City:                profile.userInput.address.city,
		Gender:              profile.userInput.gender,
		MaritalStatus:       profile.userInput.maritalStatus,
		DateOfBirth:         profile.userInput.dateOfBirth,
		PrimaryEmail:        profile.userInput.emails.primary,
		AlternateEmail1:     profile.userInput.emails.alternate1,
		AlternateEmail2:     profile.userInput.emails.alternate2,
		MobileNumber:        profile.userInput.phones.mobileNumber,
		WhatsAppNumber:      profile.userInput.phones.whatsAppNumber,
		TelegramNumber:      profile.userInput.phones.telegramNumber,
		FirstLanguage:       profile.userInput.languages.first,
		OtherLanguage1:      profile.userInput.languages.other1,
		OtherLanguage2:      profile.userInput.languages.other2,
		OtherLanguage3:      profile.userInput.languages.other3,
		OtherLanguage4:      profile.userInput.languages.other4,
		ListeningLanguage:   profile.userInput.languages.listening,
		ReadingLanguage:     profile.userInput.languages.reading,
		EmailLanguage:       profile.userInput.languages.email,
		StudyStartYear:      profile.userInput.studyStartYear,
		StudyFramework:      profile.userInput.studyFramework,
		HasGroup:            profile.userInput.ten.hasGroup,
		WantsGroup:          profile.userInput.ten.wantsGroup,
		NameOfGroup:         profile.userInput.ten.nameOfGroup,
	}

	c.JSON(http.StatusOK, result)
}
