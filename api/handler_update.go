package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

func (p *ProfileManager) update(c *gin.Context) {
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

	if err := p.repo.UpdateProfile(c.Request.Context(), keycloakID, repo.UserInput{
		FirstNameLatin:      request.FirstNameLatin,
		FirstNameVernacular: request.FirstNameVernacular,
		LastNameLatin:       request.LastNameLatin,
		LastNameVernacular:  request.LastNameVernacular,
		Address: repo.Address{
			StreetAddress: request.StreetAddress,
			Country:       request.Country,
			StateOrRegion: request.StateOrRegion,
			PostalCode:    request.PostalCode,
			City:          request.City,
		},
		Status: repo.UserStatus{
			Membership:     request.Status.Membership,
			MembershipType: request.Status.MembershipType,
			Ticket:         request.Status.Ticket,
			Convention:     request.Status.Convention,
			Galaxy:         request.Status.Galaxy,
		},
		Gender:        request.Gender,
		MaritalStatus: request.MaritalStatus,
		DateOfBirth:   request.DateOfBirth,
		Emails: repo.Emails{
			Primary:    request.PrimaryEmail,
			Alternate1: request.AlternateEmail1,
			Alternate2: request.AlternateEmail2,
		},
		Phones: repo.Phones{
			MobileNumber:   request.MobileNumber,
			WhatsAppNumber: request.WhatsAppNumber,
			TelegramNumber: request.TelegramNumber,
		},
		Languages: repo.Languages{
			First:     request.FirstLanguage,
			Other1:    request.OtherLanguage1,
			Other2:    request.OtherLanguage2,
			Other3:    request.OtherLanguage3,
			Other4:    request.OtherLanguage4,
			Listening: request.ListeningLanguage,
			Reading:   request.ReadingLanguage,
			Email:     request.EmailLanguage,
		},
		StudyStartYear: request.StudyStartYear,
		StudyFramework: request.StudyFramework,
		Ten: repo.Ten{
			HasGroup:    request.HasGroup,
			WantsGroup:  request.WantsGroup,
			NameOfGroup: request.NameOfGroup,
		},
	}); err != nil {
		if errors.Is(err, common.ErrProfileNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while updating user %q: %w", keycloakID, err))
		return
	}

	updateErr := p.kcClient.UpdateUser(c.Request.Header.Get("Authorization"), keycloakIDString, *request.FirstNameVernacular, *request.LastNameVernacular)

	if updateErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while syncing user with keycloak %s: %w", keycloakID, updateErr))
		return
	}

	c.Status(http.StatusOK)
}
