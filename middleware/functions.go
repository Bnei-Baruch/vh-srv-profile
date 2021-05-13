package middleware

import (
	"fmt"
	"net/http"
	"strings"

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
	// This is the kind of headers to expect, when the request is been proxied
	//User-Agent   =>  curl/7.64.1
	//Accept   =>  application/json
	//X-Forwarded-Port   =>  8000
	//X-Forwarded-Prefix   =>  /profile
	//X-Forwarded-For   =>  92.232.183.222, 172.19.0.1
	//X-Forwarded-Proto   =>  http
	//X-Forwarded-Host   =>  api.althafm.com
	//X-Forwarded-Path   =>  /profile/v1/profiles
	//X-Real-Ip   =>  172.19.0.1
	//Authorization   =>  Bearer <access_token>
	//Connection   =>  keep-alive
	//X-Userinfo   =>  <id_token>
	// id_token will have all info regarding the user.



	for k,v := range c.Request.Header{
		fmt.Println(k,"  => ", strings.Join(v, " , "))
	}

	tokenString := c.Request.Header.Get("Authoriation")

	tokenParts := strings.Split(tokenString," ")

	validToken := tokenParts[1]

	if validToken == "" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": app.GetSystemMessage("access_denied"),
		})
		return
	}

	claims := &app.TokenClaims{}

	token, err := jwt.ParseWithClaims(validToken, claims, func(token *jwt.Token) (interface{}, error) {
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
