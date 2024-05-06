package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

func (p *ProfileManager) create(c *gin.Context) {
	var request profileRequest

	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.KeycloakID == nil || request.FirstNameVernacular == nil || request.LastNameVernacular == nil ||
		request.PrimaryEmail == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing a required field"})
		return
	}

	keycloakID, err := uuid.FromString(*request.KeycloakID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("malformed keycloak_id: %v", err)})
		return
	}

	if !p.isSubjectOrHasAnyRole(c, *request.KeycloakID, common.RoleRoot, common.RoleAdmin) {
		return
	}

	if err := p.repo.CreateProfile(c.Request.Context(), repo.UserInput{
		KeycloakID:          &keycloakID,
		FirstNameVernacular: request.FirstNameVernacular,
		FirstNameLatin:      request.FirstNameLatin,
		LastNameVernacular:  request.LastNameVernacular,
		LastNameLatin:       request.LastNameLatin,
		Address: repo.Address{
			StreetAddress: request.StreetAddress,
			Country:       request.Country,
			StateOrRegion: request.StateOrRegion,
			PostalCode:    request.PostalCode,
			City:          request.City,
		},
		Status: repo.UserStatus{
			UserID:         &keycloakID,
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
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.CreateProfile: %w", err))
		return
	}

	// check and update first name and last name if they are not same in keycloak
	keycloakService := p.keycloakServiceFactory()
	updateErr := keycloakService.UpdateUser(c.Request.Context(), *request.KeycloakID,
		request.FirstNameVernacular, request.LastNameVernacular)

	if updateErr != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("keycloakService.UpdateUser: %w", updateErr))
		return
	}

	c.Status(http.StatusCreated)
}
