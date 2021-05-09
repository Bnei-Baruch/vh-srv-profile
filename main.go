package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gitlab.bbdev.team/vh/vh-srv-profile/app"
	"gitlab.bbdev.team/vh/vh-srv-profile/config"
	"gitlab.bbdev.team/vh/vh-srv-profile/controllers"
	"gitlab.bbdev.team/vh/vh-srv-profile/middleware"
	"gitlab.bbdev.team/vh/vh-srv-profile/models"
)

func init() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found. Read variables from environment.")
	}

	//Init log output to file
	logOutput, err := os.OpenFile("./output.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Panicf("Error opening log file: %v", err)
	}
	log.SetOutput(logOutput)

	app.Config = config.New()

	models.OpenDBConnection()
}

func main() {

	if app.Config.IsDev() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	server := gin.Default()

	server.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    cors.DefaultConfig().AllowMethods,
		AllowHeaders:    cors.DefaultConfig().AllowHeaders,
		MaxAge:          cors.DefaultConfig().MaxAge,
		AllowWebSockets: true,
	}))

	server.Use(middleware.CheckEndpointAccess)

	api := server.Group("/v1")
	{
		api.GET("/profile", controllers.Profiles)
	}

	server.Run(":" + app.Config.AppPort)

}
