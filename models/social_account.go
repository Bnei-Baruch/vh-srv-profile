package models

import (
	"log"
	"time"
)

//SocialAccount socials accounts
type SocialAccount struct {
	tableName struct{}  `pg:"social_accounts"`
	ID        uint64    `json:"id" pg:",pk"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	Active    bool      `json:"active" `
	UserID    uint64    `json:"user_id"`
	SocialID  uint64    `json:"social_id"`
	Username  string    `json:"username"`
	URL       string    `json:"url"`
	Token     string    `json:"token"`
}

type SocialAccountInfo struct {
	*SocialAccount `json:"socials_accounts" pg:",inherit,discard_unknown_columns"`
	tableName      struct{} `pg:"social_accounts"`
	Social         *Social  `json:"social" pg:"rel:has-one,fk:social_id"`
}

func FindSocialAccountsByUserID(userID uint64) (accounts []*SocialAccountInfo, err error) {

	err = DB.Model(&accounts).
		Relation("Social").
		Select()

	if err != nil {
		log.Println("Error: " + err.Error())
	}

	return
}
