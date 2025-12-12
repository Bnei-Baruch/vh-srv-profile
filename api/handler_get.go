package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/volatiletech/null/v9"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

type status struct {
	UserID         *string `json:"user_id,omitempty"`
	Membership     *bool   `json:"membership,omitempty"`
	MembershipType *string `json:"membership_type,omitempty"`
	Ticket         *bool   `json:"ticket,omitempty"`
	Convention     *bool   `json:"convention,omitempty"`
	Galaxy         *bool   `json:"galaxy,omitempty"`
}
type userResponse struct {
	UserID              *uuid.UUID  `json:"user_id"`
	KeycloakID          *uuid.UUID  `json:"keycloak_id"`
	UpdatedAt           time.Time   `json:"updated_at"`
	CreatedAt           time.Time   `json:"created_at"`
	Deleted             bool        `json:"deleted"`
	Status              status      `json:"status"`
	MembershipActive    *bool       `json:"membership_active"`
	MembershipType      *string     `json:"membership_type"`
	FirstNameLatin      *string     `json:"first_name_latin,omitempty"`
	FirstNameVernacular *string     `json:"first_name_vernacular" `
	LastNameLatin       *string     `json:"last_name_latin,omitempty"`
	LastNameVernacular  *string     `json:"last_name_vernacular"  `
	StreetAddress       *string     `json:"street_address,omitempty"`
	Country             *string     `json:"country,omitempty"`
	StateOrRegion       *string     `json:"state_region,omitempty"`
	PostalCode          *string     `json:"postal_code,omitempty"`
	City                *string     `json:"city,omitempty"`
	Gender              *string     `json:"gender,omitempty"`
	MaritalStatus       null.String `json:"marital_status"`
	SpouseKeycloakID    *string     `json:"spouse_keycloak_id,omitempty"`
	DateOfBirth         *string     `json:"date_of_birth,omitempty"`
	PrimaryEmail        *string     `json:"primary_email" `
	AlternateEmail1     *string     `json:"alternate_email_1,omitempty"`
	AlternateEmail2     *string     `json:"alternate_email_2,omitempty"`
	MobileNumber        *string     `json:"mobile_number,omitempty"`
	WhatsAppNumber      *string     `json:"whats_app_number,omitempty"`
	TelegramNumber      *string     `json:"telegram_number,omitempty"`
	FirstLanguage       *string     `json:"first_language,omitempty"`
	OtherLanguage1      *string     `json:"other_language_1,omitempty"`
	OtherLanguage2      *string     `json:"other_language_2,omitempty"`
	OtherLanguage3      *string     `json:"other_language_3,omitempty"`
	OtherLanguage4      *string     `json:"other_language_4,omitempty"`
	ListeningLanguage   *string     `json:"listening_language,omitempty"`
	ReadingLanguage     *string     `json:"reading_language,omitempty"`
	EmailLanguage       *string     `json:"email_language,omitempty"`
	StudyStartYear      *int        `json:"study_start_year,omitempty"`
	StudyFramework      *string     `json:"study_framework,omitempty"`
	HasGroup            *bool       `json:"has_ten_group,omitempty"`
	WantsGroup          *bool       `json:"wants_ten_group,omitempty"`
	NameOfGroup         *string     `json:"name_ten_group,omitempty"`
}

type ShortProfile struct {
	UserID    uuid.UUID `json:"id"`
	FirstName *string   `json:"first_name"`
	LastName  *string   `json:"last_name"`
	Email     string    `json:"email"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

func (p *ProfileManager) getProfiles(c *gin.Context) {

	if !p.HasAnyRole(c, common.RoleAnyAdmin...) {
		return
	}

	// Fetching all the query strings if present in url
	skip := c.Query("skip")
	limit := c.Query("limit")
	country := c.Query("country")
	email := c.Query("email")
	name := c.Query("name")
	keycloakID := c.Query("keycloak_id")
	tenGroupName := c.Query("ten-group-name")
	language := c.Query("language")
	firstLanguage := c.Query("first-language")
	otherLanguageOne := c.Query("other-language-1")
	otherLanguageTwo := c.Query("other-language-2")
	otherLanguageThree := c.Query("other-language-3")
	otherLanguageFour := c.Query("other-language-4")
	phoneNumber := c.Query("phone-number")
	updatedAt := c.Query("updated")
	gender := c.Query("gender")
	userID := c.Query("user_id") // Add user_id
	clauseParam := c.Query("clause")

	// Validate clause parameter
	clause := repo.AND_CLAUSE // Default clause
	if clauseParam != "" {
		upperClauseParam := strings.ToUpper(clauseParam)
		if upperClauseParam == repo.AND_CLAUSE || upperClauseParam == repo.OR_CLAUSE {
			clause = upperClauseParam
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clause value! Accepted values are 'and' or 'or'"})
			return
		}
	}
	if updatedAt != "" && updatedAt != "desc" && updatedAt != "asc" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid updatedAt value! Accepted values are desc for descending & asc for ascending"})
		return
	}

	createdAt := c.Query("created")
	if createdAt != "" && createdAt != "desc" && createdAt != "asc" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid createdAt value! Accepted values are desc for descending & asc for ascending"})
		return
	}

	membership := c.Query("membership")
	if membership != "" && membership != "false" && membership != "true" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid membership value! Accepted value is either true or false"})
		return
	}

	membershipType := c.Query("membership-type")
	convention := c.Query("convention")
	if convention != "" && convention != "false" && convention != "true" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid convention value! Accepted value is either true or false"})
		return
	}

	ticket := c.Query("ticket")
	if ticket != "" && ticket != "false" && ticket != "true" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket value! Accepted value is either true or false"})
		return
	}

	galaxy := c.Query("galaxy")
	if galaxy != "" && galaxy != "false" && galaxy != "true" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid galaxy value! Accepted value is either true or false"})
		return
	}

	// To fetch single user based on mobile number
	if phoneNumber != "" {
		profile, err := p.repo.FetchProfileBasedOnPhoneNumber(c.Request.Context(), phoneNumber)
		if err != nil {
			if errors.Is(err, common.ErrUserNotFound) {
				c.Status(http.StatusNotFound)
				return
			}
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.FetchProfileBasedOnPhoneNumber: %w", err))
			return
		}

		result := userResponse{
			UserID:     profile.UserID,
			KeycloakID: profile.UserInput.KeycloakID,
			UpdatedAt:  profile.UpdatedAt,
			CreatedAt:  profile.CreatedAt,
			Deleted:    profile.Deleted,
			Status: status{
				Membership:     profile.UserInput.Status.Membership,
				MembershipType: profile.UserInput.Status.MembershipType,
				Ticket:         profile.UserInput.Status.Ticket,
				Convention:     profile.UserInput.Status.Convention,
				Galaxy:         profile.UserInput.Status.Galaxy,
			},
			MembershipActive:    profile.UserInput.MembershipActive,
			MembershipType:      profile.UserInput.MembershipType,
			FirstNameLatin:      profile.UserInput.FirstNameLatin,
			FirstNameVernacular: profile.UserInput.FirstNameVernacular,
			LastNameLatin:       profile.UserInput.LastNameLatin,
			LastNameVernacular:  profile.UserInput.LastNameVernacular,
			StreetAddress:       profile.UserInput.Address.StreetAddress,
			Country:             profile.UserInput.Address.Country,
			StateOrRegion:       profile.UserInput.Address.StateOrRegion,
			PostalCode:          profile.UserInput.Address.PostalCode,
			City:                profile.UserInput.Address.City,
			Gender:              profile.UserInput.Gender,
			MaritalStatus:       profile.UserInput.MaritalStatus,
			PrimaryEmail:        profile.UserInput.Emails.Primary,
			AlternateEmail1:     profile.UserInput.Emails.Alternate1,
			AlternateEmail2:     profile.UserInput.Emails.Alternate2,
			MobileNumber:        profile.UserInput.Phones.MobileNumber,
			WhatsAppNumber:      profile.UserInput.Phones.WhatsAppNumber,
			TelegramNumber:      profile.UserInput.Phones.TelegramNumber,
			FirstLanguage:       profile.UserInput.Languages.First,
			OtherLanguage1:      profile.UserInput.Languages.Other1,
			OtherLanguage2:      profile.UserInput.Languages.Other2,
			OtherLanguage3:      profile.UserInput.Languages.Other3,
			OtherLanguage4:      profile.UserInput.Languages.Other4,
			ListeningLanguage:   profile.UserInput.Languages.Listening,
			ReadingLanguage:     profile.UserInput.Languages.Reading,
			EmailLanguage:       profile.UserInput.Languages.Email,
			StudyStartYear:      profile.UserInput.StudyStartYear,
			StudyFramework:      profile.UserInput.StudyFramework,
			HasGroup:            profile.UserInput.Ten.HasGroup,
			WantsGroup:          profile.UserInput.Ten.WantsGroup,
			NameOfGroup:         profile.UserInput.Ten.NameOfGroup,
		}

		var birthDate *string
		if profile.UserInput.DateOfBirth != nil {
			birthDate = utils.PointerString(profile.UserInput.DateOfBirth.Format("2006-01-02"))
		}

		result.DateOfBirth = birthDate

		c.JSON(http.StatusOK, result)
	} else {
		// fetch all the users based on parameters provided

		if skip == "" {
			skip = "0"
		}
		if limit == "" {
			limit = "10"
		}

		// String conversion to int
		intSkip, err := strconv.Atoi(skip)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid skip value! Accepted value is INTEGER"})
			return
		}

		// String conversion to int
		intLimit, err := strconv.Atoi(limit)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value! Accepted value is INTEGER"})
			return
		}

		profiles, err := p.repo.GetMultipleProfiles(c.Request.Context(), intSkip, intLimit, country, email, name, keycloakID, tenGroupName, language, firstLanguage, otherLanguageOne, otherLanguageTwo, otherLanguageThree, otherLanguageFour, updatedAt, createdAt, membership, membershipType, convention, ticket, galaxy, gender, userID, false, clause)
		if err != nil {
			if errors.Is(err, common.ErrUserNotFound) {
				c.Status(http.StatusNotFound)
				return
			}
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("repo.GetMultipleProfiles: %w", err))
			return
		}

		var arrUserRes []userResponse

		for _, profile := range profiles {
			result := userResponse{
				UserID:     profile.UserID,
				KeycloakID: profile.UserInput.KeycloakID,
				UpdatedAt:  profile.UpdatedAt,
				CreatedAt:  profile.CreatedAt,
				Deleted:    profile.Deleted,
				Status: status{
					Membership:     profile.UserInput.Status.Membership,
					MembershipType: profile.UserInput.Status.MembershipType,
					Ticket:         profile.UserInput.Status.Ticket,
					Convention:     profile.UserInput.Status.Convention,
					Galaxy:         profile.UserInput.Status.Galaxy,
				},
				MembershipActive:    profile.UserInput.MembershipActive,
				MembershipType:      profile.UserInput.MembershipType,
				FirstNameLatin:      profile.UserInput.FirstNameLatin,
				FirstNameVernacular: profile.UserInput.FirstNameVernacular,
				LastNameLatin:       profile.UserInput.LastNameLatin,
				LastNameVernacular:  profile.UserInput.LastNameVernacular,
				StreetAddress:       profile.UserInput.Address.StreetAddress,
				Country:             profile.UserInput.Address.Country,
				StateOrRegion:       profile.UserInput.Address.StateOrRegion,
				PostalCode:          profile.UserInput.Address.PostalCode,
				City:                profile.UserInput.Address.City,
				Gender:              profile.UserInput.Gender,
				MaritalStatus:       profile.UserInput.MaritalStatus,
				PrimaryEmail:        profile.UserInput.Emails.Primary,
				AlternateEmail1:     profile.UserInput.Emails.Alternate1,
				AlternateEmail2:     profile.UserInput.Emails.Alternate2,
				MobileNumber:        profile.UserInput.Phones.MobileNumber,
				WhatsAppNumber:      profile.UserInput.Phones.WhatsAppNumber,
				TelegramNumber:      profile.UserInput.Phones.TelegramNumber,
				FirstLanguage:       profile.UserInput.Languages.First,
				OtherLanguage1:      profile.UserInput.Languages.Other1,
				OtherLanguage2:      profile.UserInput.Languages.Other2,
				OtherLanguage3:      profile.UserInput.Languages.Other3,
				OtherLanguage4:      profile.UserInput.Languages.Other4,
				ListeningLanguage:   profile.UserInput.Languages.Listening,
				ReadingLanguage:     profile.UserInput.Languages.Reading,
				EmailLanguage:       profile.UserInput.Languages.Email,
				StudyStartYear:      profile.UserInput.StudyStartYear,
				StudyFramework:      profile.UserInput.StudyFramework,
				HasGroup:            profile.UserInput.Ten.HasGroup,
				WantsGroup:          profile.UserInput.Ten.WantsGroup,
				NameOfGroup:         profile.UserInput.Ten.NameOfGroup,
			}
			var birthDate *string
			if profile.UserInput.DateOfBirth != nil {
				birthDate = utils.PointerString(profile.UserInput.DateOfBirth.Format("2006-01-02"))
			}

			result.DateOfBirth = birthDate

			arrUserRes = append(arrUserRes, result)
		}

		c.JSON(http.StatusOK, arrUserRes)
	}
}

func (p *ProfileManager) get(c *gin.Context) {
	keycloakIDString, ok := c.Params.Get("keycloak_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keycloak ID missing"})
		return
	}
	keycloakID, err := uuid.FromString(keycloakIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("malformed keycloak_id: %v", err)})
		return
	}

	if !p.isSubjectOrHasAnyRole(c, keycloakIDString, common.RoleAnyAdmin...) {
		return
	}

	profile, err := p.repo.GetProfile(c.Request.Context(), keycloakID)
	if err != nil {
		if errors.Is(err, common.ErrProfileNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetProfile: %w", err))
		return
	}

	result := userResponse{
		UserID:              profile.UserID,
		KeycloakID:          profile.UserInput.KeycloakID,
		UpdatedAt:           profile.UpdatedAt,
		CreatedAt:           profile.CreatedAt,
		Deleted:             profile.Deleted,
		MembershipActive:    profile.UserInput.MembershipActive,
		MembershipType:      profile.UserInput.MembershipType,
		FirstNameLatin:      profile.UserInput.FirstNameLatin,
		FirstNameVernacular: profile.UserInput.FirstNameVernacular,
		LastNameLatin:       profile.UserInput.LastNameLatin,
		LastNameVernacular:  profile.UserInput.LastNameVernacular,
		StreetAddress:       profile.UserInput.Address.StreetAddress,
		Country:             profile.UserInput.Address.Country,
		StateOrRegion:       profile.UserInput.Address.StateOrRegion,
		PostalCode:          profile.UserInput.Address.PostalCode,
		City:                profile.UserInput.Address.City,
		Gender:              profile.UserInput.Gender,
		MaritalStatus:       profile.UserInput.MaritalStatus,
		SpouseKeycloakID:    profile.UserInput.SpouseKeycloakID,
		PrimaryEmail:        profile.UserInput.Emails.Primary,
		AlternateEmail1:     profile.UserInput.Emails.Alternate1,
		AlternateEmail2:     profile.UserInput.Emails.Alternate2,
		MobileNumber:        profile.UserInput.Phones.MobileNumber,
		WhatsAppNumber:      profile.UserInput.Phones.WhatsAppNumber,
		TelegramNumber:      profile.UserInput.Phones.TelegramNumber,
		FirstLanguage:       profile.UserInput.Languages.First,
		OtherLanguage1:      profile.UserInput.Languages.Other1,
		OtherLanguage2:      profile.UserInput.Languages.Other2,
		OtherLanguage3:      profile.UserInput.Languages.Other3,
		OtherLanguage4:      profile.UserInput.Languages.Other4,
		ListeningLanguage:   profile.UserInput.Languages.Listening,
		ReadingLanguage:     profile.UserInput.Languages.Reading,
		EmailLanguage:       profile.UserInput.Languages.Email,
		StudyStartYear:      profile.UserInput.StudyStartYear,
		StudyFramework:      profile.UserInput.StudyFramework,
		HasGroup:            profile.UserInput.Ten.HasGroup,
		WantsGroup:          profile.UserInput.Ten.WantsGroup,
		NameOfGroup:         profile.UserInput.Ten.NameOfGroup,
	}

	var birthDate *string
	if profile.UserInput.DateOfBirth != nil {
		birthDate = utils.PointerString(profile.UserInput.DateOfBirth.Format("2006-01-02"))
	}

	result.DateOfBirth = birthDate

	c.JSON(http.StatusOK, result)
}

func (p *ProfileManager) getProfileShort(c *gin.Context) {
	keycloakIDString, ok := c.Params.Get("keycloak_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keycloak ID missing"})
		return
	}
	keycloakID, err := uuid.FromString(keycloakIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("malformed keycloak_id: %v", err)})
		return
	}

	if !p.isSubjectOrHasAnyRole(c, keycloakIDString, common.RoleAnyAdmin...) {
		return
	}

	profile, err := p.repo.GetProfile(c.Request.Context(), keycloakID)
	if err != nil {
		if errors.Is(err, common.ErrProfileNotFound) {
			c.JSON(http.StatusOK, gin.H{"error": "profile not found"})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetProfile: %w", err))
		return
	}

	result := ShortProfile{
		UserID:    *profile.UserID,
		FirstName: profile.UserInput.FirstNameVernacular,
		LastName:  profile.UserInput.LastNameVernacular,
		Email:     *profile.UserInput.Emails.Primary,
		Active:    profile.UserInput.MembershipActive != nil && *profile.UserInput.MembershipActive,
		CreatedAt: profile.CreatedAt,
	}

	c.JSON(http.StatusOK, result)
}
