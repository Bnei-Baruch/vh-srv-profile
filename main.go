package main

import (
	"context"
	"log"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db_url := makeDBURL()

	profileDB, err := newPgProfileDB(ctx, db_url)
	if err != nil {
		log.Fatalf("Unable to initialize profile db: %s \n***\n %s \n ***", err, db_url)
	}

	profile := &profileManager{creator: profileDB, getter: profileDB, updater: profileDB, deleter: profileDB, hardDeleter: profileDB, fetchProfiles: profileDB}

	app := initApp(appHandlers{
		create:      profile.create,
		get:         profile.get,
		update:      profile.update,
		delete:      profile.delete,
		hardDelete:  profile.hardDelete,
		getProfiles: profile.getProfiles,
	})

	if err := app.Run(":" + config.appPort); err != nil {
		log.Printf("server stopped: %s", err)
	}
}

type appHandlers struct {
	create      gin.HandlerFunc
	get         gin.HandlerFunc
	update      gin.HandlerFunc
	delete      gin.HandlerFunc
	hardDelete  gin.HandlerFunc
	getProfiles gin.HandlerFunc
}

func initApp(handlers appHandlers) *gin.Engine {
	app := gin.Default()
	//app.Use(cors.Default())

	app.POST("/v1/profile", handlers.create)
	app.GET("/v1/profiles", handlers.getProfiles)
	app.GET("/v1/profile/:keycloak_id", handlers.get)
	app.PATCH("/v1/profile/:keycloak_id", handlers.update)
	app.DELETE("/v1/profile/:keycloak_id", handlers.delete)
	app.DELETE("/admin/v1/profile/:keycloak_id", handlers.hardDelete)

	return app
}
