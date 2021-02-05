package models

import (
	"time"
)

//Country
type Country struct {
	tableName struct{}        `pg:"countries"`
	ID        uint64          `json:"id" pg:",pk"`
	Created   time.Time       `json:"created"`
	Updated   time.Time       `json:"updated"`
	Active    bool            `json:"active" `
	Name      []NameTranslate `json:"name"`
}

//CountryInfo
type CountryInfo struct {
	tableName struct{}    `pg:"countries"`
	ID        uint64      `json:"id" pg:",pk"`
	Name      interface{} `json:"name" pg:",jsonb"`
}
