package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createUser struct {
	KeycloakID          string  `json:"keycloak_id" binding:"required"`
	FirstNameLatin      *string `json:"first_name_latin"`
	FirstNameVernacular string  `json:"first_name_vernacular" binding:"required"`
	LastNameLatin       *string `json:"last_name_latin"`
	LastNameVernacular  string  `json:"last_name_vernacular"  binding:"required"`
	StreetAddress       *string `json:"street_address"`
	Country             *string `json:"country"`
	StateOrRegion       *string `json:"state_region"`
	PostalCode          *string `json:"postal_code"`
	City                *string `json:"city"`
	Gender              *string `json:"gender"`
	MaritalStatus       *string `json:"marital_status"`
	DateOfBirth         *string `json:"date_of_birth"`
	PrimaryEmail        string  `json:"primary_email" binding:"required"`
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
	NameOfGroup         *bool   `json:"name_ten_group"`
}

type storage interface {
	createUser(ctx context.Context, user user) error
}

type profile struct {
	db storage
}

func (p *profile) create(c *gin.Context) {
	var request createUser

	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}
	if err := p.db.createUser(c.Request.Context(), user{
		keycloakID:          request.KeycloakID,
		firstNameLatin:      request.FirstNameLatin,
		firstNameVernacular: request.FirstNameVernacular,
		lastNameLatin:       request.LastNameLatin,
		lastNameVernacular:  request.LastNameVernacular,
		address: address{
			streetAddress: request.StreetAddress,
			country:       request.Country,
			stateOrRegion: request.StateOrRegion,
			postalCode:    request.PostalCode,
			city:          request.City,
		},
		gender:        request.Gender,
		maritalStatus: request.MaritalStatus,
		dateOfBirth:   request.DateOfBirth,
		emails: emails{
			primary:    request.PrimaryEmail,
			alternate1: request.AlternateEmail1,
			alternate2: request.AlternateEmail2,
		},
		phones: phones{
			mobileNumber:   request.MobileNumber,
			whatsAppNumber: request.WhatsAppNumber,
			telegramNumber: request.TelegramNumber,
		},
		languages: languages{
			first:     request.FirstLanguage,
			other1:    request.OtherLanguage1,
			other2:    request.OtherLanguage2,
			other3:    request.OtherLanguage3,
			other4:    request.OtherLanguage4,
			listening: request.ListeningLanguage,
			reading:   request.ReadingLanguage,
			email:     request.EmailLanguage,
		},
		studyStartYear: request.StudyStartYear,
		studyFramework: request.StudyFramework,
		ten: ten{
			hasGroup:    request.HasGroup,
			wantsGroup:  request.WantsGroup,
			nameOfGroup: request.NameOfGroup,
		},
	}); err != nil {
		c.Status(http.StatusInternalServerError)
		log.Printf("error while creating user %q: %s", request.KeycloakID, err.Error())
		_ = c.Error(fmt.Errorf("error while creating user: %s", err.Error()))
		return
	}

	c.Status(http.StatusCreated)
}
