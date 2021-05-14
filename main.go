package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	config := loadConfig()
	if config.appMode == "dev" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	//Init log output to file
	logOutput, err := os.OpenFile("./output.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %s", err)
	}
	log.SetOutput(logOutput)
	gin.DefaultWriter = logOutput
	gin.DefaultErrorWriter = logOutput

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	profileDB, err := newPgProfileDB(ctx, getEnvOrFatal("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to initialize profile db: %s", err)
	}

	profile := &profileManager{db: profileDB}

	app := initApp(appHandlers{
		create: profile.create,
		get:    profile.get,
	})

	if err := app.Run(config.appPort); err != nil {
		log.Printf("server stopped: %s", err)
	}
}

type appHandlers struct {
	create gin.HandlerFunc
	get    gin.HandlerFunc
}

func initApp(handlers appHandlers) *gin.Engine {
	app := gin.Default()
	app.Use(cors.Default())

	app.POST("/v1/profile", handlers.create)
	app.GET("/v1/profile/:keycloakID", handlers.get)

	return app
}
