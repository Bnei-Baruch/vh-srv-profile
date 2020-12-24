package app

import (
	"crypto/rand"
	"fmt"
	"log"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

func GetSystemMessage(key string) string {
	return SystemMessages.Data[key]
}

func IsEmailValid(e string) bool {

	var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}

func HashPassword(password string) (string, error) {

	pwd := []byte(password)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)

	if err != nil {
		log.Println("Error: " + err.Error())
	}

	return string(hash), err

}

func VerifyPassword(hashedPwd string, plainPwd string) bool {

	byteHash := []byte(hashedPwd)
	bytePwd := []byte(plainPwd)

	err := bcrypt.CompareHashAndPassword(byteHash, bytePwd)

	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func GetJWTKey() []byte {
	return []byte("ONLY_FOR_TEST_KEY")
}

func TokenGenerator() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
