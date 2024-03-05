package api

import (
	"context"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"gitlab.bbdev.team/vh/vh-srv-profile/api/middleware"
	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/membership"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

type App struct {
	profileManager          *ProfileManager
	profileDB               *repo.ProfileDB
	eventListener           *orders.EventListener
	membershipEventsHandler *membership.EventsHandler
	gEngine                 *gin.Engine
}

func NewApp() *App {
	return new(App)
}

func (a *App) Initialize() {
	a.initDB()
	a.initEventListener()
	a.profileManager = NewProfileManager(a.profileDB)
	a.initGinEngine()
}

func (a *App) initDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbUrl := repo.MakeDBURL()
	log.Printf("Trying to connect to db: %s\n", dbUrl)

	var err error
	a.profileDB, err = repo.NewProfileDB(ctx, dbUrl)
	if err != nil {
		log.Fatalf("Unable to initialize profile db: %v\n", err)
	}
	log.Println("Connected to profile db")

	if err = repo.SyncDBStructInsertionAndMigrations(); err != nil {
		log.Fatalf("Unable to migrate profile db: %v\n", err)
	}
	log.Println("Migrated profile db")
}

func (a *App) initEventListener() {
	if common.Config.NatsUrl != "" {
		log.Println("Initializing events listener")

		var err error
		a.eventListener, err = orders.NewEventListener()
		if err != nil {
			log.Fatalf("Error creating event listener: %s \n", err)
		}

		a.membershipEventsHandler = membership.NewEventsHandler(a.profileDB)
		a.eventListener.RegisterHandler(a.membershipEventsHandler.HandleOrdersEvent)

		if err = a.eventListener.Run(); err != nil {
			log.Fatalf("Error running event listener: %s \n", err)
		}
	}
}

func (a *App) initGinEngine() {
	gin.SetMode(common.Config.Mode)
	a.gEngine = gin.Default()
	a.gEngine.Use(middleware.TokenSource())
	if gin.IsDebugging() {
		a.gEngine.Use(cors.Default())
	}

	// Creating a group of routes that will be prefixed with `/v1`
	baseV1Path := a.gEngine.Group("/v1")

	a.gEngine.POST("/v1/profile", a.profileManager.create)
	a.gEngine.GET("/v1/profiles", a.profileManager.getProfiles)
	a.gEngine.GET("/v1/profile/:keycloak_id", a.profileManager.get)
	a.gEngine.PATCH("/v1/profile/:keycloak_id", a.profileManager.update)
	a.gEngine.DELETE("/v1/profile/:keycloak_id", a.profileManager.delete)
	a.gEngine.DELETE("/admin/v1/profile/:keycloak_id", a.profileManager.hardDelete)

	a.gEngine.GET("/v1/requests", a.profileManager.getRequests)
	a.gEngine.POST("/v1/request", a.profileManager.createRequest)
	a.gEngine.POST("/v1/request/:id/conclude", a.profileManager.concludeRequest)

	grant := baseV1Path.Group("/grant")
	{
		grant.GET("/:id", a.profileManager.handleGrantFetchByID)
	}
	baseV1Path.GET("/grants", a.profileManager.handleGrantFetchAll)

	membershipRoutes := baseV1Path.Group("/membership")
	{
		membershipRoutes.GET("/user/:user_id", a.profileManager.handleMembershipFetchByUserID)
		membershipRoutes.GET("/kcid/:kcid", a.profileManager.handleMembershipFetchByKCID)
		membershipRoutes.GET("/id/:id", a.profileManager.handleMembershipFetchByID)
		membershipRoutes.POST("/evaluation", a.profileManager.handleMembershipEvaluationByUserID)
		membershipRoutes.PATCH("/:id", a.profileManager.handleMembershipPatchByID)
		membershipRoutes.DELETE("/:id", a.profileManager.handleMembershipSoftDeleteByID)
		membershipRoutes.POST("/cancellation", a.profileManager.handleMembershipCancellation)
	}
	baseV1Path.GET("/memberships", a.profileManager.handleMembershipFetchAll)

	// notification crud
	notification := baseV1Path.Group("/notification")
	{
		notification.POST("", a.profileManager.handleNotificationCreate)
		notification.GET("/:id", a.profileManager.handleNotificationFetchByID)
		notification.PATCH("/:id", a.profileManager.handleNotificationPatchByID)
		notification.DELETE("/:id", a.profileManager.handleNotificationSoftDeleteByID)
	}
	baseV1Path.GET("/notifications", a.profileManager.handleNotificationFetchAll)

	userNotification := baseV1Path.Group("/user/notification")
	{
		userNotification.POST("", a.profileManager.handleUserNotificationCreate)
		userNotification.GET("/:id", a.profileManager.handleUserNotificationFetchByID)
		userNotification.PATCH("/:id", a.profileManager.handleUserNotificationPatchByID)
		userNotification.DELETE("/:id", a.profileManager.handleUserNotificationSoftDeleteByID)
	}
	baseV1Path.GET("/user/notifications", a.profileManager.handleUserNotificationFetchAll)

	operation := baseV1Path.Group("/operation")
	{
		operation.POST("/", a.profileManager.handleOperationCreate)
		operation.POST("/revert", a.profileManager.handleOperationRevert)
	}
}

func (a *App) Run() {
	if err := a.gEngine.Run(":" + common.Config.Port); err != nil {
		log.Fatalf("server stopped: %s", err)
	}
}

func (a *App) Shutdown() {
	a.eventListener.Close()
	a.profileDB.Close()
}
