package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/events"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

func init() {
	rootCmd.AddCommand(syncUsersCmd)
}

var syncUsersCmd = &cobra.Command{
	Use:   "sync-users",
	Short: "emit fake update events for all users",
	Run:   syncUsersFn,
}

func syncUsersFn(cmd *cobra.Command, args []string) {
	slog.Info("running users sync")

	// Setup sentry
	sentryTransport := sentry.NewHTTPSyncTransport()
	sentryTransport.Timeout = 3 * time.Second
	err := sentry.Init(sentry.ClientOptions{
		Release:     common.GitSHA,
		Environment: common.Config.Env,
		Transport:   sentryTransport,
		Tags: map[string]string{
			"command": "sync-users",
		},
	})
	if err != nil {
		utils.LogFatal("sentry.Init", slog.Any("err", err))
	}
	defer sentry.Flush(2 * time.Second)

	// sync users
	syncer := NewUserSyncer()
	if err := syncer.init(); err != nil {
		sentry.CaptureException(err)
		utils.LogFatal("syncer.init", slog.Any("err", err))
	}
	defer syncer.close()

	if err := syncer.run(); err != nil {
		sentry.CaptureException(err)
		utils.LogFatal("syncer.run", slog.Any("err", err))
	}

	slog.Info("done syncing users. closing...")
}

type UserSyncer struct {
	repo         repo.ProfileRepository
	eventEmitter events.EventEmitter
}

func NewUserSyncer() *UserSyncer {
	return new(UserSyncer)
}

func (s *UserSyncer) init() error {
	var err error

	s.eventEmitter, err = events.CreateEmitter()
	if err != nil {
		return fmt.Errorf("events.CreateEmitter: %w", err)
	}

	dbUrl := repo.MakeDBURL()
	db, err := repo.NewProfileDB(context.TODO(), dbUrl, s.eventEmitter)
	if err != nil {
		return fmt.Errorf("repo.NewProfileDB %s: %w", dbUrl, err)
	}
	s.repo = db

	return nil
}

func (s *UserSyncer) close() {
	s.repo.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	s.eventEmitter.Close(ctx)
	cancel()
	sentry.Flush(2 * time.Second)
}

func (s *UserSyncer) run() error {
	pageSize := 1000
	page := 0
	ctx := context.TODO()

	for {
		slog.Info("fetching users", slog.Int("page", page))
		users, err := s.repo.GetMultipleProfiles(ctx, page*pageSize, pageSize,
			"", "", "", "", "", "", "",
			"", "", "", "",
			"", "", "", "", "", "", "", "", "", false, repo.AND_CLAUSE)
		if err != nil {
			return fmt.Errorf("repo.GetMultipleProfiles: %w", err)
		}
		slog.Info("fetched users", slog.Int("count", len(users)))

		for _, user := range users {
			event := events.MakeEvent(events.TypeUpdateProfile, map[string]interface{}{
				"keycloak_id": user.UserInput.KeycloakID.String(),
				"user_id":     user.UserID.String(),
			})
			event.Component = events.ComponentUserSyncer
			event.Actor = events.ActorSystem
			s.eventEmitter.Emit(ctx, event)
		}

		page++
		if len(users) < pageSize {
			break
		}
	}
	return nil
}
