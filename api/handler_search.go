package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

type userResponseWithNotification struct {
	userResponse
	NotifyCreatedFromOrders          bool `json:"notify_created_from_orders,omitempty"`
	NotifyRetrievedByEmailFromOrders bool `json:"notify_retrieved_by_email_from_orders,omitempty"`
	NotifyEmailUpdatedFromOrders     bool `json:"notify_email_updated_from_orders,omitempty"`
}

func (p *ProfileManager) searchProfiles(c *gin.Context) {

	if !p.HasAnyRole(c, common.RoleAnyAdmin...) {
		return
	}

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

	if email == "" || phoneNumber != "" {
		p.getProfiles(c)
		return
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

	profiles, err := p.repo.GetMultipleProfiles(c.Request.Context(), intSkip, intLimit, country, email, name, tenGroupName, language, firstLanguage, otherLanguageOne, otherLanguageTwo, otherLanguageThree, otherLanguageFour, updatedAt, createdAt, membership, membershipType, convention, ticket, galaxy, gender, true)
	if err != nil {
		if errors.Is(err, common.ErrUserNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.GetMultipleProfiles: %w", err))
		return
	}

	// If no profiles found by email, try to find the user in Orders.
	var createdFromOrders bool
	var retrievedByEmailFromOrders bool
	var emailUpdatedFromOrders bool
	if len(profiles) == 0 {
		account, err := p.ordersService.GetAccountByEmailIfExist(c.Request.Context(), email)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(fmt.Errorf("ordersService.GetAccountByEmail: %w", err))
			return
		}
		if account != nil {
			// Search profile by account keycloak id from orders.
			keycloakID, err := uuid.FromString(account.UserKey)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				_ = c.Error(fmt.Errorf("parse UserKey from orders result: %w", err))
				return
			}
			profile, err := p.repo.GetProfile(c, keycloakID)
			if err != nil {
				if errors.Is(err, common.ErrProfileNotFound) {
					// Profile not found. Create profile from orders account.
					// The types of firstNameVernacular and lastNameVernacular in DB are TEXT NOT NULL, so we pass empty strings if the values from orders are nil.
					empty := ""
					firstNameVernacular := account.FirstName
					if firstNameVernacular == nil {
						firstNameVernacular = &empty
					}

					lastNameVernacular := account.LastName
					if lastNameVernacular == nil {
						lastNameVernacular = &empty
					}
					userInput := repo.UserInput{
						KeycloakID: &keycloakID,
						Emails: repo.Emails{
							Primary: &account.Email,
						},
						FirstNameLatin:      account.FirstName,
						LastNameLatin:       account.LastName,
						FirstNameVernacular: firstNameVernacular,
						LastNameVernacular:  lastNameVernacular,
						Address: repo.Address{
							Country:       account.Country,
							StreetAddress: account.Street,
							City:          account.City,
							PostalCode:    account.Postcode,
							StateOrRegion: account.State,
						},
						Phones: repo.Phones{
							MobileNumber: account.Phone,
						},
					}
					err = p.repo.CreateProfileWithCountryCheck(c.Request.Context(), userInput)
					if err != nil {
						c.Status(http.StatusInternalServerError)
						_ = c.Error(fmt.Errorf("repo.CreateProfile: %w", err))
						return
					}
					profile, err = p.repo.GetProfile(c, keycloakID)
					if err != nil {
						c.Status(http.StatusInternalServerError)
						_ = c.Error(fmt.Errorf("repo.GetProfile: %w", err))
						return
					}
					createdFromOrders = true
				} else {
					c.Status(http.StatusInternalServerError)
					_ = c.Error(fmt.Errorf("repo.GetProfile: %w", err))
					return
				}
			} else {
				retrievedByEmailFromOrders = true
				if len(account.Email) > 0 {
					// Profile found by keycloak of orders account. Append email from orders account to the profile.
					emailUpdated := false
					if profile.UserInput.Emails.Primary == nil {
						profile.UserInput.Emails.Primary = &account.Email
						emailUpdated = true
					} else if profile.UserInput.Emails.Alternate1 == nil {
						profile.UserInput.Emails.Alternate1 = &account.Email
						emailUpdated = true
					} else if profile.UserInput.Emails.Alternate2 == nil {
						profile.UserInput.Emails.Alternate2 = &account.Email
						emailUpdated = true
					} else if profile.UserInput.Emails.Primary == profile.UserInput.Emails.Alternate1 {
						profile.UserInput.Emails.Alternate1 = &account.Email
						emailUpdated = true
					} else if profile.UserInput.Emails.Primary == profile.UserInput.Emails.Alternate2 {
						profile.UserInput.Emails.Alternate2 = &account.Email
						emailUpdated = true
					} else if profile.UserInput.Emails.Alternate1 == profile.UserInput.Emails.Alternate2 {
						profile.UserInput.Emails.Alternate2 = &account.Email
						emailUpdated = true
					}
					if emailUpdated {
						err := p.repo.UpdateProfile(c, keycloakID, profile.UserInput)
						if err != nil {
							c.Status(http.StatusInternalServerError)
							_ = c.Error(fmt.Errorf("repo.UpdateProfile: %w", err))
							return
						}
						emailUpdatedFromOrders = true
					}
				}
			}
			profiles = append(profiles, profile)
		}
	}

	var arrUserRes []userResponseWithNotification

	for _, profile := range profiles {
		result := userResponseWithNotification{
			userResponse: userResponse{
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
			},
		}
		var birthDate *string
		if profile.UserInput.DateOfBirth != nil {
			birthDate = utils.PointerString(profile.UserInput.DateOfBirth.Format("2006-01-02"))
		}

		result.DateOfBirth = birthDate

		arrUserRes = append(arrUserRes, result)
	}

	if createdFromOrders {
		arrUserRes[0].NotifyCreatedFromOrders = true
	}
	if retrievedByEmailFromOrders {
		arrUserRes[0].NotifyRetrievedByEmailFromOrders = true
	}
	if emailUpdatedFromOrders {
		arrUserRes[0].NotifyEmailUpdatedFromOrders = true
	}

	c.JSON(http.StatusOK, arrUserRes)
}
