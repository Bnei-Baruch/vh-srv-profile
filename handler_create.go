package main

import (
	"context"
	"fmt"
	"net/http"

	uuid "github.com/satori/go.uuid"

	"github.com/gin-gonic/gin"
)

type createStorage interface {
	createProfile(ctx context.Context, user userInput) error
}

func (p *profileManager) create(c *gin.Context) {
	var request profileRequest

	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	if request.KeycloakID == nil || request.FirstNameVernacular == nil || request.LastNameVernacular == nil ||
		request.PrimaryEmail == nil {
		err := fmt.Errorf("missing a required field for provided request: %#v", request)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	keycloakID, err := uuid.FromString(*request.KeycloakID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	if err := p.creator.createProfile(c.Request.Context(), userInput{
		keycloakID:          &keycloakID,
		firstNameVernacular: request.FirstNameVernacular,
		firstNameLatin:      request.FirstNameLatin,
		lastNameVernacular:  request.LastNameVernacular,
		lastNameLatin:       request.LastNameLatin,
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
		_ = c.Error(fmt.Errorf("error while creating user %q: %w", *request.KeycloakID, err))
		return
	}

	c.Status(http.StatusCreated)
}
