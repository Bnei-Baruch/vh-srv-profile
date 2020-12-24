package models

import "time"

//SocialAccount socials accounts
type SocialAccount struct {
	tableName struct{}  `sql:"socials_accounts"`
	ID        uint64    `json:"id" pg:",pk"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	Active    bool      `json:"active" `
	UserID    uint64    `json:"user_id"`
	SocialID  uint64    `json:"social_id"`
	URL       string    `json:"url"`
	Token     string    `json:"token"`
}
