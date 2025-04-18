package main

import (
	"github.com/spf13/cobra"
	"github.com/lariskovski/containy/internal/image"
	"github.com/lariskovski/containy/internal/run"
)

func main() {
	var rootCmd = &cobra.Command{Use: "containy"}
	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "build [file]",
			Short: "Build a container",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				image.Build(args[0])
			},
		},
		&cobra.Command{
			Use:   "run [overlay-dir] [command]",
			Short: "Run a container",
			Args:  cobra.MinimumNArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				run.RunContainer(args)
			},
		},
	)
	rootCmd.Execute()
}
