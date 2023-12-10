package cmd

import (
	"gitlab.bbdev.team/vh/vh-srv-profile/membership"

	"github.com/spf13/cobra"
)

var membershipCmd = &cobra.Command{
	Use:   "membership",
	Short: "Eval all users",
	Run: func(cmd *cobra.Command, args []string) {
		membership.Migrate()
	},
}

func init() {
	rootCmd.AddCommand(membershipCmd)
}
