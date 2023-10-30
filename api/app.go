package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/keycloak"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

type App struct {
	ProfileManager *ProfileManager
	ProfileDB      *repo.ProfileDB
	gEngine        *gin.Engine
	config         common.AppConfig
	kcClient       keycloak.KeycloakClient
}

func NewApp() *App {
	return new(App)
}

func (a *App) Initialize() {
	a.config = common.LoadConfig()
	if a.config.Mode == "dev" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbUrl := repo.MakeDBURL()
	fmt.Println("trying to connect to db:", dbUrl)

	profileDB, err := repo.NewProfileDB(ctx, dbUrl)
	if err != nil {
		log.Fatalf("Unable to initialize profile db: %s \n***\n %s \n ***", err, dbUrl)
	}
	a.ProfileDB = profileDB

	fmt.Println("Connected to profile db")

	migErr := repo.SyncDBStructInsertionAndMigrations()
	if migErr != nil {
		log.Fatalf("Unable to migrate profile db: %s \n***\n %s \n ***", migErr, dbUrl)
	}
	fmt.Println("Migrated profile db")

	a.ProfileManager = NewProfileManager(a.ProfileDB, a.kcClient)
	a.initGinEngine()
}

func (a *App) initGinEngine() {
	a.gEngine = gin.Default()
	//a.gEngine.Use(cors.Default())

	// Creating a group of routes that will be prefixed with `/v1`
	baseV1Path := a.gEngine.Group("/v1")

	a.gEngine.POST("/v1/profile", a.ProfileManager.create)
	a.gEngine.GET("/v1/profiles", a.ProfileManager.getProfiles)
	a.gEngine.GET("/v1/profile/:keycloak_id", a.ProfileManager.get)
	a.gEngine.PATCH("/v1/profile/:keycloak_id", a.ProfileManager.update)
	a.gEngine.DELETE("/v1/profile/:keycloak_id", a.ProfileManager.delete)
	a.gEngine.DELETE("/admin/v1/profile/:keycloak_id", a.ProfileManager.hardDelete)

	a.gEngine.GET("/v1/requests", a.ProfileManager.getRequests)
	a.gEngine.POST("/v1/request", a.ProfileManager.createRequest)
	a.gEngine.PATCH("/v1/request/:id", a.ProfileManager.updateRequest)
	a.gEngine.DELETE("/v1/request/:id", a.ProfileManager.deleteRequest)

	grant := baseV1Path.Group("/grant")
	{
		grant.GET("/:id", a.ProfileManager.handleGrantFetchByID)
		grant.POST("", a.ProfileManager.handleGrantCreate)
		grant.PATCH("/:id", a.ProfileManager.handleGrantPatchByID)
		grant.DELETE("/:id", a.ProfileManager.handleGrantSoftDeleteByID)
	}
	baseV1Path.GET("/grants", a.ProfileManager.handleGrantFetchAll)

	membership := baseV1Path.Group("/membership")
	{
		membership.GET("/user/:user_id", a.ProfileManager.handleMembershipFetchByUserID)
		membership.GET("/kcid/:kcid", a.ProfileManager.handleMembershipFetchByKCID)
		membership.GET("/id/:id", a.ProfileManager.handleMembershipFetchByID)
		membership.POST("/evaluation", a.ProfileManager.handleMembershipEvaluationByUserID)
		membership.PATCH("/:id", a.ProfileManager.handleMembershipPatchByID)
		membership.DELETE("/:id", a.ProfileManager.handleMembershipSoftDeleteByID)
		membership.POST("/cancellation", a.ProfileManager.handleMembershipCancellation)
	}
	baseV1Path.GET("/memberships", a.ProfileManager.handleMembershipFetchAll)

	// notification crud
	notification := baseV1Path.Group("/notification")
	{
		notification.POST("", a.ProfileManager.handleNotificationCreate)
		notification.GET("/:id", a.ProfileManager.handleNotificationFetchByID)
		notification.PATCH("/:id", a.ProfileManager.handleNotificationPatchByID)
		notification.DELETE("/:id", a.ProfileManager.handleNotificationSoftDeleteByID)
	}
	baseV1Path.GET("/notifications", a.ProfileManager.handleNotificationFetchAll)

	userNotification := baseV1Path.Group("/user/notification")
	{
		userNotification.POST("", a.ProfileManager.handleUserNotificationCreate)
		userNotification.GET("/:id", a.ProfileManager.handleUserNotificationFetchByID)
		userNotification.PATCH("/:id", a.ProfileManager.handleUserNotificationPatchByID)
		userNotification.DELETE("/:id", a.ProfileManager.handleUserNotificationSoftDeleteByID)
	}
	baseV1Path.GET("/user/notifications", a.ProfileManager.handleUserNotificationFetchAll)

	operation := baseV1Path.Group("/operation")
	{
		operation.POST("/", a.ProfileManager.handleOperationCreate)
		operation.POST("/revert", a.ProfileManager.handleOperationRevert)
	}
}

func (a *App) initKeycloak() {
	serverURL := os.Getenv("KEYCLOAK_SERVER_URL")
	realm := os.Getenv("KEYCLOAK_REALM")

	if serverURL == "" || realm == "" {
		log.Fatal("missing keycloak server url or realm")
	}

	a.kcClient = keycloak.NewClient(serverURL, realm)
}

func (a *App) Run() {
	if err := a.gEngine.Run(":" + a.config.Port); err != nil {
		log.Fatalf("server stopped: %s", err)
	}
}
