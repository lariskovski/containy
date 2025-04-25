package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	var rootCmd = &cobra.Command{Use: "containy"}
	rootCmd.AddCommand(
		NewBuildCmd(),
		NewRunCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
