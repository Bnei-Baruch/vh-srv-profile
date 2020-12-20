package models

import "time"

//Language
type Language struct {
	tableName struct{}  `sql:"languages"`
	ID        uint64    `json:"id" pg:",pk"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	Active    bool      `json:"active" `
	Code      string    `json:"code" `
	Name      string    `json:"name"`
}
