package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gitlab.bbdev.team/vh/vh-srv-profile/app"
	"gitlab.bbdev.team/vh/vh-srv-profile/config"
	"gitlab.bbdev.team/vh/vh-srv-profile/models"
)

func init() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found. Read variables from environment.")
	}

	app.Config = config.New()

	models.OpenDBConnection()
}

func main() {

	if app.Config.AppMode == "dev" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	server := gin.Default()

	api := server.Group("/v1")
	{
		api.GET("/signin", Account.Activate)
		api.GET("/signup", Account.Activate)

		api.GET("/profile/personal", Account.Activate)
		api.GET("/profile/framework", Account.Activate)
		api.GET("/profile/ten", Account.Activate)
		api.GET("/profile/skills", Account.Activate)
		api.GET("/profile/notification", Account.Activate)
	}

	server.Run(":" + app.Config.AppPort)

}
