package models

import (
	"log"
	"time"
)

type User struct {
	tableName struct{}  `sql:"users"`
	ID        uint64    `json:"id" pg:",pk"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	Active    bool      `json:"active" `
	Roles     []int     `json:"roles" pg:",array"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Country   int       `json:"country"`
	Language  int       `json:"language"`
	BirthDate time.Time `json:"birthdate"`
	Gender    int       `json:"gender"`
}

//InsertUser insert new user
func InsertUser(data *User) error {

	data.Created = time.Now()
	data.Updated = data.Created

	_, err := DB.Model(data).Insert()

	if err != nil {
		log.Println("Error: " + err.Error())
	}

	return err
}
