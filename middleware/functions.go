package middleware

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type TokenClaims struct {
	Token string `json:"token"`
	jwt.StandardClaims
}

//CheckEndpointAccess
func CheckEndpointAccess(c *gin.Context) {
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

	// Either we can use a debug logger or print the below with a flag.

	//for k,v := range c.Request.Header{
	//	fmt.Println(k,"  => ", strings.Join(v, " , "))
	//
	//}

	tokenString := c.Request.Header.Get("Authorization")

	tokenParts := strings.Split(tokenString, " ")

	validToken := tokenParts[1]

	if validToken == "" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "access_denied",
		})
		return
	}

	claims := &TokenClaims{}

	token, err := jwt.ParseWithClaims(validToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("ONLY_FOR_TEST_KEY"), nil
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

	c.Next()

}
