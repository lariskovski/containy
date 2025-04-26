package cmd

import (
	"os"

	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/container"
	"github.com/spf13/cobra"
)

// init initializes the run command and adds it to the root command
func init() {
	// Add the run command to the root command
	rootCmd.AddCommand(runCmd)
}

// NewRunCmd creates the run command
var runCmd = &cobra.Command{
	Use:   "run [overlay-dir] [command]",
	Short: "Run a container",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := container.Create(args); err != nil {
			config.Log.Errorf("Container execution failed: %v", err)
			os.Exit(1)
		}
	},
}
