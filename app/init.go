package app

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"gitlab.bbdev.team/vh/vh-srv-profile/config"
)

//Config app config
var Config config.AppConfig

//SystemMessages system text message
var SystemMessages SystemMessagesType

func init() {
	dir, _ := os.Getwd()

	jsonFile, err := os.Open(dir + "/messages.json")

	if err != nil {
		log.Panicln(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &SystemMessages)

}
