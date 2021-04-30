package models

import (
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
