package cmd

import (
	"context"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/events"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

func init() {
	rootCmd.AddCommand(devMigrateCmd)
}

var devMigrateCmd = &cobra.Command{
	Use:   "dev-migrate",
	Short: "Run database migrations and create NATS stream",
	Long:  "Run database migrations and create NATS stream for the profile service",
	Run:   devMigrateFn,
}

func devMigrateFn(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	slog.Info("Starting migration process")

	// Run database migrations
	slog.Info("Running database migrations")
	if err := repo.SyncDBStructInsertionAndMigrations(); err != nil {
		slog.Error("Database migration failed", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Database migrations completed successfully")

	// Create NATS stream
	if common.Config.NatsUrl != "" {
		slog.Info("Creating NATS stream", slog.String("nats_url", common.Config.NatsUrl))
		natsHandler, err := events.NewNatsEventHandler()
		if err != nil {
			slog.Error("Failed to create NATS event handler", slog.Any("error", err))
			os.Exit(1)
		}

		// Close the handler to ensure connection is properly established
		if err := natsHandler.Close(ctx); err != nil {
			slog.Error("Failed to close NATS event handler", slog.Any("error", err))
			os.Exit(1)
		}

		slog.Info("NATS stream created successfully")
	} else {
		slog.Info("NATS_URL not configured, skipping NATS stream creation")
	}

	slog.Info("Migration process completed successfully")
}
