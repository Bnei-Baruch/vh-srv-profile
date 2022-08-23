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

	migErr := SyncDBStructInsertionAndMigrations()
	if migErr != nil {
		log.Fatalf("Unable to migrate profile db: %s \n***\n %s \n ***", migErr, db_url)
	}

	profile := &profileManager{creator: profileDB, requestCreator: profileDB, getter: profileDB, updater: profileDB, requestUpdater: profileDB, deleter: profileDB, requestDeleter: profileDB, hardDeleter: profileDB, fetchProfiles: profileDB, fetchRequests: profileDB}

	app := initApp(appHandlers{
		create:        profile.create,
		createRequest: profile.createRequest,
		get:           profile.get,
		update:        profile.update,
		updateRequest: profile.updateRequest,
		delete:        profile.delete,
		deleteRequest: profile.deleteRequest,
		hardDelete:    profile.hardDelete,
		getProfiles:   profile.getProfiles,
		getRequests:   profile.getRequest,
	})

	if err := app.Run(":" + config.appPort); err != nil {
		log.Printf("server stopped: %s", err)
	}
}

type appHandlers struct {
	create        gin.HandlerFunc
	createRequest gin.HandlerFunc
	get           gin.HandlerFunc
	update        gin.HandlerFunc
	updateRequest gin.HandlerFunc
	delete        gin.HandlerFunc
	deleteRequest gin.HandlerFunc
	hardDelete    gin.HandlerFunc
	getProfiles   gin.HandlerFunc
	getRequests   gin.HandlerFunc
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

	app.GET("/v1/requests", handlers.getRequests)
	app.POST("/v1/request", handlers.createRequest)
	app.PATCH("/v1/request/:keycloak_id", handlers.updateRequest)
	app.DELETE("/v1/request/:keycloak_id", handlers.deleteRequest)

	return app
}
