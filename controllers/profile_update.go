package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.bbdev.team/vh/vh-srv-profile/models"
)

//ProfileUpdateRequest
type ProfileUpdateRequest struct {
	FirstName                string   `json:"first_name" binding:"required"`
	LastName                 string   `json:"last_name" binding:"required"`
	Phone                    string   `json:"phone" binding:"required"`
	Country                  int      `json:"country" binding:"required"`
	Language                 int      `json:"language" binding:"required"`
	BirthDate                string   `json:"birthdate" binding:"required"`
	Gender                   int      `json:"gender" binding:"required"`
	Address1                 string   `json:"address_1"`
	Address2                 string   `json:"address_2"`
	AddressState             string   `json:"address_state"`
	AddressCity              string   `json:"address_city"`
	AddressCountry           string   `json:"address_country"`
	AddressPostcode          uint     `json:"address_postcode"`
	ProfileImage             string   `json:"profile_image"`
	FirstYearOfStudy         uint16   `json:"first_year_of_study"`
	LearningCenter           string   `json:"learning_center"`
	TenName                  string   `json:"ten_name"`
	TenID                    string   `json:"ten_id"`
	FavoriteLearningPlatform string   `json:"favorite_learning_platform"`
	NativeLanguage           uint64   `json:"native_language"`
	AdditionalLanguages      []uint64 `json:"additional_languages"`
	LanguageForText          uint64   `json:"language_for_text"`
	LanguageForNotification  uint64   `json:"language_for_notification"`
	LanguageForVideo         uint64   `json:"language_for_video"`
}

func ProfileUpdate(c *gin.Context) {

	var request ProfileUpdateRequest
	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	birthdate, err := time.Parse("2006-01-02", request.BirthDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	userID, ok := c.Get("UserID")

	if !ok {
		c.Status(http.StatusInsufficientStorage)
		return
	}

	userData, _ := models.FindUserByID(userID.(uint64))

	userData.FirstName = request.FirstName
	userData.LastName = request.LastName
	userData.Phone = request.Phone
	userData.Country = request.Country
	userData.Language = request.Language
	userData.BirthDate = birthdate
	userData.Gender = request.Gender

	passwordData := models.PasswordData{}

	errors, ok := models.ValidateUpdateUser(userData, passwordData)

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors,
		})

		return
	}

	err = models.UpdateUser(userData)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.Status(http.StatusOK)
}
