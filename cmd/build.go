package cmd

import (
	"os"

	"github.com/lariskovski/containy/internal/build"
	"github.com/lariskovski/containy/internal/config"
	"github.com/spf13/cobra"
)

var (
// 	filePath string
	alias	string
)

func init() {
	// Add the build command to the root command
	rootCmd.AddCommand(buildCmd)

	// Define flags for the build command
	// buildCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the Dockerfile")
	buildCmd.Flags().StringVarP(&alias, "alias", "a", "", "Alias for the image")
}

// buildCmd creates the build command
var buildCmd = &cobra.Command{
	Use:   "build [file]",
	Short: "Build a container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := build.Build(args[0], alias); err != nil {
			// It's appropriate to log and exit here as we're at the app boundary
			config.Log.Errorf("Build failed: %v", err)
			os.Exit(1)
		}
	},
}
