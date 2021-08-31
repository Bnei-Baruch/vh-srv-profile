package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type updateStorage interface {
	updateProfile(ctx context.Context, keycloakID uuid.UUID, toUpdate userInput) error
}

func (p *profileManager) update(c *gin.Context) {
	keycloakIDString, ok := c.Params.Get("keycloak_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keycloak ID missing"})
		return
	}
	keycloakID, err := uuid.FromString(keycloakIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var request profileRequest
	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	if err := p.updater.updateProfile(c.Request.Context(), keycloakID, userInput{
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
		status: userStatus{
			membership:     request.Status.Membership,
			membershipType: request.Status.MembershipType,
			ticket:         request.Status.Ticket,
			convention:     request.Status.Convention,
			galaxy:         request.Status.Galaxy,
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
		if errors.Is(err, errProfileNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while updating user %q: %w", keycloakID, err))
		return
	}

	c.Status(http.StatusOK)
}
