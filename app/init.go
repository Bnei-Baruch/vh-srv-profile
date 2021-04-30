package app

import (
	"encoding/json"
	"log"

	"gitlab.bbdev.team/vh/vh-srv-profile/config"
)

//Config app config
var Config config.AppConfig

//SystemMessages system text message
var SystemMessages SystemMessagesType

func init() {
	if err := json.Unmarshal([]byte(`{
    "data": {
        "email_not_valid": "E-mail is not valid.",
        "email_exists": "E-mail already exists.",
        "phone_exists":"Phone already exists.",
        "password_min_length_error": "Minimum password length 6 characters.",
        "password_max_length_error": "Maximum password length 32 characters.",
        "wrong_current_password": "wrong  current password",
        "wrong_password_confirmation": "The password confirmation does not match.",
        "access_denied": "Access denied",
        "account_blocked": "Account blocked"
    }
}`), &SystemMessages); err != nil {
		log.Fatal(err)
	}
}
