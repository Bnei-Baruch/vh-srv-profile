package models

import (
	"log"
	"time"
)

type Role struct {
	tableName struct{}  `sql:"roles"`
	ID        uint64    `json:"id" pg:",pk"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
}

//GetRoles get all roles
func GetRoles() ([]Role, error) {

	var role []Role

	err := DB.Model(&role).Where("active = ?", true).Select()

	if err != nil {
		log.Println("Error: " + err.Error())
	}

	return role, err
}
