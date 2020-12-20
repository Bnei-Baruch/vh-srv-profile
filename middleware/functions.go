package middleware

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gitlab.bbdev.team/vh/vh-srv-profile/app"
	"gitlab.bbdev.team/vh/vh-srv-profile/models"
)

//GetOpenEndpoints
func GetOpenEndpoints() (url []string) {

	url = []string{
		"/v1/profile/login",
		"/v1/profile/create",
	}

	return
}

//CheckEndpointAccess
func CheckEndpointAccess(c *gin.Context) {

	path := c.FullPath()

	for _, endpoint := range GetOpenEndpoints() {
		if endpoint == path {
			return
		}
	}
	tokenString := c.Request.Header.Get("token")

	if tokenString == "" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": app.GetSystemMessage("access_denied"),
		})
		return
	}

	claims := &app.TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return app.GetJWTKey(), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if !token.Valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	user, err := models.FindUserByToken(claims.Token)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": app.GetSystemMessage("access_denied"),
		})
		return
	}

	c.Set("UserID", user.ID)

	c.Next()

}

//CheckAuth
func CheckAuth(token string) bool {

	return false
}
