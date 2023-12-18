package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.bbdev.team/vh/vh-srv-profile/membership"
)

var membershipCmd = &cobra.Command{
	Use:   "membership",
	Short: "Membership commands, please use one of the sub-commands",
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Eval all users (migrate v2)",
	Run: func(cmd *cobra.Command, args []string) {
		membership.Migrate()
	},
}

var invalidateCmd = &cobra.Command{
	Use:   "invalidate",
	Short: "Invalidate expired memberships",
	Run: func(cmd *cobra.Command, args []string) {
		membership.Invalidate()
	},
}

func init() {
	membershipCmd.AddCommand(migrateCmd)
	membershipCmd.AddCommand(invalidateCmd)
	rootCmd.AddCommand(membershipCmd)
}
