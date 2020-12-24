package app

import "github.com/dgrijalva/jwt-go"

//SystemMessagesType
type SystemMessagesType struct {
	Data map[string]string `json:"data"`
}

//tokenClaims
type TokenClaims struct {
	Token string `json:"token"`
	jwt.StandardClaims
}
