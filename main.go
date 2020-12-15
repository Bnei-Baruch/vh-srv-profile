package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gitlab.bbdev.team/vh/vh-srv-profile/app"
	"gitlab.bbdev.team/vh/vh-srv-profile/config"
)

func init() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found. Read variables from environment.")
	}

	app.Config = config.New()
}

func main() {

	if app.Config.AppMode == "dev" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	server := gin.Default()

	server.Run(":" + app.Config.AppPort)

}
