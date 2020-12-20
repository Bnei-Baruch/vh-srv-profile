package controllers

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gitlab.bbdev.team/vh/vh-srv-profile/app"
	"gitlab.bbdev.team/vh/vh-srv-profile/models"
)

//ProfileLoginRequest
type ProfileLoginRequest struct {
	Email    string `json:"email" form:"email" binding:"required" `
	Password string `json:"password" form:"password" binding:"required" `
}

//ProfileLogin
func ProfileLogin(c *gin.Context) {

	var request ProfileLoginRequest

	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := models.FindUserByEmail(request.Email)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": app.GetSystemMessage("access_denied"),
		})
		return
	}

	if !user.Active {
		c.JSON(http.StatusForbidden, gin.H{
			"error": app.GetSystemMessage("account_blocked"),
		})
		return
	}

	if !app.VerifyPassword(user.Password, request.Password) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": app.GetSystemMessage("access_denied"),
		})
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &app.TokenClaims{
		Token: user.Token,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(app.GetJWTKey())

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
	})

}
