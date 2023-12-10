package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vh-srv-profile",
	Short: "Virtual home profile management",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(common.LoadConfig)
}
