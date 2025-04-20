package main

import (
	"fmt"
	"os"

	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/container"
	"github.com/lariskovski/containy/internal/image"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "containy"}
	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "build [file]",
			Short: "Build a container",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				if err := image.Build(args[0]); err != nil {
					// It's appropriate to log and exit here as we're at the app boundary
					config.Log.Errorf("Build failed: %v", err)
					os.Exit(1)
				}
			},
		},
		&cobra.Command{
			Use:   "run [overlay-dir] [command]",
			Short: "Run a container",
			Args:  cobra.MinimumNArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				if err := container.Create(args); err != nil {
					config.Log.Errorf("Container execution failed: %v", err)
					os.Exit(1)
				}
			},
		},
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
