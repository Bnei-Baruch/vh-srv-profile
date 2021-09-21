package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type readStorage interface {
	getProfile(ctx context.Context, keycloakID uuid.UUID) (user, error)
}
type readMultipleProfileStorage interface {
	//Fetch multiple profile based on paramteres
	getMultipleProfiles(ctx context.Context, intSkip int, intLimit int, country string, email string, firstLastName string, tenGroupName string, language string, firstLanguage string, otherLanguageOne string, otherLanguageTwo string, otherLanguageThree string, otherLanguageFour string, updatedAt string, createdAt string, membership string, membershipType string, convention string, ticket string, galaxy string, gender string) ([]user, error)
	// Fetch single profle based on phone number provided
	fetchProfileBasedOnPhoneNumber(ctx context.Context, phoneNumber string) (user, error)
}

type status struct {
	UserID         *string `json:"user_id,omitempty"`
	Membership     *bool   `json:"membership,omitempty"`
	MembershipType *string `json:"membership_type,omitempty"`
	Ticket         *bool   `json:"ticket,omitempty"`
	Convention     *bool   `json:"convention,omitempty"`
	Galaxy         *bool   `json:"galaxy,omitempty"`
}
type userResponse struct {
	UserID              *uuid.UUID `json:"user_id"`
	KeycloakID          *uuid.UUID `json:"keycloak_id"`
	UpdatedAt           time.Time  `json:"updated_at"`
	CreatedAt           time.Time  `json:"created_at"`
	Deleted             bool       `json:"deleted"`
	Status              status     `json:"status"`
	FirstNameLatin      *string    `json:"first_name_latin,omitempty"`
	FirstNameVernacular *string    `json:"first_name_vernacular" `
	LastNameLatin       *string    `json:"last_name_latin,omitempty"`
	LastNameVernacular  *string    `json:"last_name_vernacular"  `
	StreetAddress       *string    `json:"street_address,omitempty"`
	Country             *string    `json:"country,omitempty"`
	StateOrRegion       *string    `json:"state_region,omitempty"`
	PostalCode          *string    `json:"postal_code,omitempty"`
	City                *string    `json:"city,omitempty"`
	Gender              *string    `json:"gender,omitempty"`
	MaritalStatus       *string    `json:"marital_status,omitempty"`
	DateOfBirth         *string    `json:"date_of_birth,omitempty"`
	PrimaryEmail        *string    `json:"primary_email" `
	AlternateEmail1     *string    `json:"alternate_email_1,omitempty"`
	AlternateEmail2     *string    `json:"alternate_email_2,omitempty"`
	MobileNumber        *string    `json:"mobile_number,omitempty"`
	WhatsAppNumber      *string    `json:"whats_app_number,omitempty"`
	TelegramNumber      *string    `json:"telegram_number,omitempty"`
	FirstLanguage       *string    `json:"first_language,omitempty"`
	OtherLanguage1      *string    `json:"other_language_1,omitempty"`
	OtherLanguage2      *string    `json:"other_language_2,omitempty"`
	OtherLanguage3      *string    `json:"other_language_3,omitempty"`
	OtherLanguage4      *string    `json:"other_language_4,omitempty"`
	ListeningLanguage   *string    `json:"listening_language,omitempty"`
	ReadingLanguage     *string    `json:"reading_language,omitempty"`
	EmailLanguage       *string    `json:"email_language,omitempty"`
	StudyStartYear      *int       `json:"study_start_year,omitempty"`
	StudyFramework      *string    `json:"study_framework,omitempty"`
	HasGroup            *bool      `json:"has_ten_group,omitempty"`
	WantsGroup          *bool      `json:"wants_ten_group,omitempty"`
	NameOfGroup         *string    `json:"name_ten_group,omitempty"`
}

func (p *profileManager) getProfiles(c *gin.Context) {

	// Fetching all the query strings if present in url
	skip := c.Query("skip")
	limit := c.Query("limit")
	country := c.Query("country")
	email := c.Query("email")
	name := c.Query("name")
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
		profile, err := p.fetchProfiles.fetchProfileBasedOnPhoneNumber(c.Request.Context(), phoneNumber)
		if err != nil {
			if errors.Is(err, errUserNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("error while getting the user with phone number %q: %w", phoneNumber, err))
			return
		}

		result := userResponse{
			UserID:     profile.userID,
			KeycloakID: profile.userInput.keycloakID,
			UpdatedAt:  profile.updatedAt,
			CreatedAt:  profile.createdAt,
			Deleted:    profile.deleted,
			Status: status{
				Membership:     profile.userInput.status.membership,
				MembershipType: profile.userInput.status.membershipType,
				Ticket:         profile.userInput.status.ticket,
				Convention:     profile.userInput.status.convention,
				Galaxy:         profile.userInput.status.galaxy,
			},
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

		var birthDate *string
		if profile.userInput.dateOfBirth != nil {
			birthDate = pointerString(profile.userInput.dateOfBirth.Format("2006-01-02"))
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

		profiles, err := p.fetchProfiles.getMultipleProfiles(c.Request.Context(), intSkip, intLimit, country, email, name, tenGroupName, language, firstLanguage, otherLanguageOne, otherLanguageTwo, otherLanguageThree, otherLanguageFour, updatedAt, createdAt, membership, membershipType, convention, ticket, galaxy, gender)
		if err != nil {
			if errors.Is(err, errUserNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("error while getting users: %w", err))
			return
		}

		var arrUserRes []userResponse

		for _, profile := range profiles {
			result := userResponse{
				UserID:     profile.userID,
				KeycloakID: profile.userInput.keycloakID,
				UpdatedAt:  profile.updatedAt,
				CreatedAt:  profile.createdAt,
				Deleted:    profile.deleted,
				Status: status{
					Membership:     profile.userInput.status.membership,
					MembershipType: profile.userInput.status.membershipType,
					Ticket:         profile.userInput.status.ticket,
					Convention:     profile.userInput.status.convention,
					Galaxy:         profile.userInput.status.galaxy,
				},
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
			var birthDate *string
			if profile.userInput.dateOfBirth != nil {
				birthDate = pointerString(profile.userInput.dateOfBirth.Format("2006-01-02"))
			}

			result.DateOfBirth = birthDate

			arrUserRes = append(arrUserRes, result)
		}

		c.JSON(http.StatusOK, arrUserRes)
	}
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
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while getting user %q: %w", keycloakIDString, err))
		return
	}

	result := userResponse{
		UserID:     profile.userID,
		KeycloakID: profile.userInput.keycloakID,
		UpdatedAt:  profile.updatedAt,
		CreatedAt:  profile.createdAt,
		Deleted:    profile.deleted,
		Status: status{
			Membership:     profile.userInput.status.membership,
			MembershipType: profile.userInput.status.membershipType,
			Ticket:         profile.userInput.status.ticket,
			Convention:     profile.userInput.status.convention,
			Galaxy:         profile.userInput.status.galaxy,
		},
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

	var birthDate *string
	if profile.userInput.dateOfBirth != nil {
		birthDate = pointerString(profile.userInput.dateOfBirth.Format("2006-01-02"))
	}

	result.DateOfBirth = birthDate

	c.JSON(http.StatusOK, result)
}
