package models

import "time"

//Directory
type Directory struct {
	tableName     struct{}        `sql:"directories"`
	ID            uint64          `json:"id" pg:",pk"`
	Created       time.Time       `json:"created"`
	Updated       time.Time       `json:"updated"`
	Active        bool            `json:"active" `
	DirectoryType string          `json:"directory_type" `
	Name          []NameTranslate `json:"name"`
	Weight        int             `json:"weight"`
}
