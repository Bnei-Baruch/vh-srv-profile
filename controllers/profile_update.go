package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.bbdev.team/vh/vh-srv-profile/models"
)

//ProfileUpdateRequest
type ProfileUpdateRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone" binding:"required"`
	Country   int    `json:"country" binding:"required"`
	Language  int    `json:"language" binding:"required"`
	BirthDate string `json:"birthdate" binding:"required"`
	Gender    int    `json:"gender" binding:"required"`
}

//ProfileUpdateLoginRequest
type ProfileUpdateLoginRequest struct {
	Email                string `json:"email" form:"email" binding:"required"`
	CurrentPassword      string `json:"current_password" form:"current_password" `
	Password             string `json:"password" form:"password" `
	PasswordConfirmation string `json:"password_confirmation" form:"password_confirmation"  `
}

func ProfileUpdateLogin(c *gin.Context) {

	var request ProfileUpdateLoginRequest
	if err := c.Bind(&request); err != nil {
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
	userData.Email = request.Email

	passwordData := models.PasswordData{
		CurrentPassword:      request.CurrentPassword,
		Password:             request.Password,
		PasswordConfirmation: request.PasswordConfirmation,
	}

	errors, ok := models.ValidateUpdateUser(userData, passwordData)

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors,
		})

		return
	}

	err := models.UpdateUser(userData)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	if passwordData.Password != "" {

		err = models.SetPassword(passwordData.Password, userData.ID)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.Status(http.StatusOK)

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
