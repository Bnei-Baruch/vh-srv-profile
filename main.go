package main

import (
	"context"
	"fmt"
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

	fmt.Println("trying to connect to db:", db_url)

	profileDB, err := newPgProfileDB(ctx, db_url)
	if err != nil {
		log.Fatalf("Unable to initialize profile db: %s \n***\n %s \n ***", err, db_url)
	}

	fmt.Println("Connected to profile db")

	migErr := SyncDBStructInsertionAndMigrations()
	if migErr != nil {
		log.Fatalf("Unable to migrate profile db: %s \n***\n %s \n ***", migErr, db_url)
	}

	fmt.Println("Migrated profile db")

	profile := &profileManager{
		creator:          profileDB,
		requestCreator:   profileDB,
		getter:           profileDB,
		updater:          profileDB,
		requestUpdater:   profileDB,
		deleter:          profileDB,
		requestDeleter:   profileDB,
		hardDeleter:      profileDB,
		fetchProfiles:    profileDB,
		fetchRequests:    profileDB,
		grant:            profileDB,
		membership:       profileDB,
		notification:     profileDB,
		userNotification: profileDB,
		operation:        profileDB,
	}

	app := initApp(appHandlers{
		create:                               profile.create,
		createRequest:                        profile.createRequest,
		get:                                  profile.get,
		update:                               profile.update,
		updateRequest:                        profile.updateRequest,
		delete:                               profile.delete,
		deleteRequest:                        profile.deleteRequest,
		hardDelete:                           profile.hardDelete,
		getProfiles:                          profile.getProfiles,
		getRequests:                          profile.getRequest,
		handleGrantFetchByID:                 profile.handleGrantFetchByID,
		handleGrantCreate:                    profile.handleGrantCreate,
		handleGrantPatchByID:                 profile.handleGrantPatchByID,
		handleGrantSoftDeleteByID:            profile.handleGrantSoftDeleteByID,
		handleGrantFetchAll:                  profile.handleGrantFetchAll,
		handleMembershipFetchByID:            profile.handleMembershipFetchByID,
		handleMembershipFetchByUserID:        profile.handleMembershipFetchByUserID,
		handleMembershipPatchByID:            profile.handleMembershipPatchByID,
		handleMembershipSoftDeleteByID:       profile.handleMembershipSoftDeleteByID,
		handleMembershipFetchAll:             profile.handleMembershipFetchAll,
		handleMembershipCancellation:         profile.handleMembershipCancellation,
		handleMembershipEvaluationByUserID:   profile.handleMembershipEvaluationByUserID,
		handleNotificationFetchByID:          profile.handleNotificationFetchByID,
		handleNotificationCreate:             profile.handleNotificationCreate,
		handleNotificationPatchByID:          profile.handleNotificationPatchByID,
		handleNotificationSoftDeleteByID:     profile.handleNotificationSoftDelete,
		handleNotificationFetchAll:           profile.handleNotificationFetchAll,
		handleUserNotificationFetchByID:      profile.handleUserNotificationFetchByID,
		handleUserNotificationCreate:         profile.handleUserNotificationCreate,
		handleUserNotificationPatchByID:      profile.handleUserNotificationPatchByID,
		handleUserNotificationSoftDeleteByID: profile.handleUserNotificationSoftDelete,
		handleUserNotificationFetchAll:       profile.handleUserNotificationFetchAll,
		handleOperationCreate:                profile.handleOperationCreate,
		handleOperationRevert:                profile.handleOperationRevert,
	})

	if err := app.Run(":" + config.appPort); err != nil {
		log.Printf("server stopped: %s", err)
	}
}

type appHandlers struct {
	create                               gin.HandlerFunc
	createRequest                        gin.HandlerFunc
	get                                  gin.HandlerFunc
	update                               gin.HandlerFunc
	updateRequest                        gin.HandlerFunc
	delete                               gin.HandlerFunc
	deleteRequest                        gin.HandlerFunc
	hardDelete                           gin.HandlerFunc
	getProfiles                          gin.HandlerFunc
	getRequests                          gin.HandlerFunc
	handleGrantFetchByID                 gin.HandlerFunc
	handleGrantCreate                    gin.HandlerFunc
	handleGrantPatchByID                 gin.HandlerFunc
	handleGrantSoftDeleteByID            gin.HandlerFunc
	handleGrantFetchAll                  gin.HandlerFunc
	handleMembershipFetchByID            gin.HandlerFunc
	handleMembershipFetchByUserID        gin.HandlerFunc
	handleMembershipPatchByID            gin.HandlerFunc
	handleMembershipFetchAll             gin.HandlerFunc
	handleMembershipSoftDeleteByID       gin.HandlerFunc
	handleMembershipCancellation         gin.HandlerFunc
	handleMembershipEvaluationByUserID   gin.HandlerFunc
	handleNotificationFetchByID          gin.HandlerFunc
	handleNotificationCreate             gin.HandlerFunc
	handleNotificationPatchByID          gin.HandlerFunc
	handleNotificationSoftDeleteByID     gin.HandlerFunc
	handleNotificationFetchAll           gin.HandlerFunc
	handleUserNotificationFetchByID      gin.HandlerFunc
	handleUserNotificationCreate         gin.HandlerFunc
	handleUserNotificationPatchByID      gin.HandlerFunc
	handleUserNotificationSoftDeleteByID gin.HandlerFunc
	handleUserNotificationFetchAll       gin.HandlerFunc
	handleOperationCreate                gin.HandlerFunc
	handleOperationRevert                gin.HandlerFunc
}

func initApp(handlers appHandlers) *gin.Engine {
	app := gin.Default()
	//app.Use(cors.Default())

	// Creating a group of routes that will be prefixed with `/v1`
	baseV1Path := app.Group("/v1")

	app.POST("/v1/profile", handlers.create)
	app.GET("/v1/profiles", handlers.getProfiles)
	app.GET("/v1/profile/:keycloak_id", handlers.get)
	app.PATCH("/v1/profile/:keycloak_id", handlers.update)
	app.DELETE("/v1/profile/:keycloak_id", handlers.delete)
	app.DELETE("/admin/v1/profile/:keycloak_id", handlers.hardDelete)

	app.GET("/v1/requests", handlers.getRequests)
	app.POST("/v1/request", handlers.createRequest)
	app.PATCH("/v1/request/:id", handlers.updateRequest)
	app.DELETE("/v1/request/:id", handlers.deleteRequest)

	grant := baseV1Path.Group("/grant")
	{
		grant.GET("/:id", handlers.handleGrantFetchByID)
		grant.POST("", handlers.handleGrantCreate)
		grant.PATCH("/:id", handlers.handleGrantPatchByID)
		grant.DELETE("/:id", handlers.handleGrantSoftDeleteByID)
	}
	baseV1Path.GET("/grants", handlers.handleGrantFetchAll)

	membership := baseV1Path.Group("/membership")
	{
		membership.GET("/user/:user_id", handlers.handleMembershipFetchByUserID)
		membership.GET("/id/:id", handlers.handleMembershipFetchByID)
		membership.POST("/evaluation", handlers.handleMembershipEvaluationByUserID)
		membership.PATCH("/:id", handlers.handleMembershipPatchByID)
		membership.DELETE("/:id", handlers.handleMembershipSoftDeleteByID)
		membership.POST("/cancellation", handlers.handleMembershipCancellation)
	}
	baseV1Path.GET("/memberships", handlers.handleMembershipFetchAll)

	// notification crud
	notification := baseV1Path.Group("/notification")
	{
		notification.POST("", handlers.handleNotificationCreate)
		notification.GET("/:id", handlers.handleNotificationFetchByID)
		notification.PATCH("/:id", handlers.handleNotificationPatchByID)
		notification.DELETE("/:id", handlers.handleNotificationSoftDeleteByID)
	}
	baseV1Path.GET("/notifications", handlers.handleNotificationFetchAll)

	userNotification := baseV1Path.Group("/user/notification")
	{
		userNotification.POST("", handlers.handleUserNotificationCreate)
		userNotification.GET("/:id", handlers.handleUserNotificationFetchByID)
		userNotification.PATCH("/:id", handlers.handleUserNotificationPatchByID)
		userNotification.DELETE("/:id", handlers.handleUserNotificationSoftDeleteByID)
	}
	baseV1Path.GET("/user/notifications", handlers.handleUserNotificationFetchAll)
	operation := baseV1Path.Group("/operation")
	{
		operation.POST("/", handlers.handleOperationCreate)
		operation.POST("/revert", handlers.handleOperationRevert)
	}

	return app
}
