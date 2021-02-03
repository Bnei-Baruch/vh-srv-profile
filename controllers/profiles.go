package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.bbdev.team/vh/vh-srv-profile/models"
)

type ProfilesRequest struct {
	Limit  int `json:"limit" form:"limit"  `
	Offset int `json:"offset" form:"offset" `
}

//Profiles list of all profiles
func Profiles(c *gin.Context) {

	var request ProfilesRequest

	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if request.Limit == 0 {
		request.Limit = 10
	}

	count, users, err := models.FindUsers(models.UserFilter{}, request.Limit, request.Offset)

	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   users,
		"count":  count,
		"limit":  request.Limit,
		"offset": request.Offset,
	})

}
