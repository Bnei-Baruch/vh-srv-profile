package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.bbdev.team/vh/vh-srv-profile/app"
	"gitlab.bbdev.team/vh/vh-srv-profile/models"
)

//ProfileCreateRequest
type ProfileCreateRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Country   int    `json:"country" binding:"required"`
	Language  int    `json:"language" binding:"required"`
	BirthDate string `json:"birthdate" binding:"required"`
	Gender    int    `json:"gender" binding:"required"`
}

//ProfileCreate create profile
func ProfileCreate(c *gin.Context) {

	var request ProfileCreateRequest

	var roles []int

	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	roles = append(roles, 1)

	user := &models.User{
		Roles:     roles,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Phone:     request.Phone,
		Email:     request.Email,
		Active:    true,
		Gender:    request.Gender,
		Country:   request.Country,
		Language:  request.Country,
		Token:     app.TokenGenerator(),
	}

	birthdate, err := time.Parse("2006-01-02", request.BirthDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	user.BirthDate = birthdate

	errors, ok := models.ValidateUser(user, true)

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors,
		})

		return
	}

	err = models.InsertUser(user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)

}
