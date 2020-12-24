package models

import "time"

//Language
type Country struct {
	tableName struct{}        `sql:"countries"`
	ID        uint64          `json:"id" pg:",pk"`
	Created   time.Time       `json:"created"`
	Updated   time.Time       `json:"updated"`
	Active    bool            `json:"active" `
	Name      []NameTranslate `json:"name"`
}
