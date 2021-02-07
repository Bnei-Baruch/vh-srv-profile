package models

import "time"

//Social
type Social struct {
	tableName struct{}  `sql:"socials"`
	ID        uint64    `json:"id" pg:",pk"`
	Created   time.Time `json:"-"`
	Updated   time.Time `json:"-"`
	Active    bool      `json:"active" `
	Name      string    `json:"name"`
}
