package main

type user struct {
	keycloakID          string
	firstNameLatin      string
	firstNameVernacular *string
	lastNameLatin       string
	lastNameVernacular  *string
	address             address
	gender              string
	maritalStatus       string
	dateOfBirth         *string
	emails              emails
	phones              phones
	languages           languages
	studyStartYear      int
	studyFramework      string
	ten                 ten
}

type address struct {
	streetAddress *string
	country       *string
	stateOrRegion *string
	postalCode    *string
	city          *string
}

type emails struct {
	primary    string
	alternate1 *string
	alternate2 *string
}

type phones struct {
	mobileNumber   string
	whatsAppNumber *string
	telegramNumber *string
}

type languages struct {
	first     string
	other1    *string
	other2    *string
	other3    *string
	other4    *string
	listening string
	reading   string
	email     string
}

type ten struct {
	hasGroup    bool
	wantsGroup  *bool
	nameOfGroup *bool
}

const (
	mobile   = "mobile"
	whatsApp = "WhatsApp"
	telegram = "Telegram"
)
