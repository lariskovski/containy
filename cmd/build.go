package cmd

import (
    "os"

    "github.com/lariskovski/containy/internal/build"
    "github.com/lariskovski/containy/internal/config"
    "github.com/spf13/cobra"
)

// NewBuildCmd creates the build command
func NewBuildCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "build [file]",
        Short: "Build a container",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            if err := build.Build(args[0]); err != nil {
                // It's appropriate to log and exit here as we're at the app boundary
                config.Log.Errorf("Build failed: %v", err)
                os.Exit(1)
            }
        },
    }
}