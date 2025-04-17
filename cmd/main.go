package main

import (
	"os"

	"github.com/lariskovski/containy/internal/build"
	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/run"
)

func main() {
	args := os.Args[1:] // Get command-line arguments (excluding the program name)
	if len(args) == 0 {
		config.Log.Error("Usage: main.go <command> [<args>]")
		os.Exit(1)
	}

	command := args[0]
	switch command {
	case "run":
		config.Log.Debugf("Executing run command")
		if len(args) > 1 {
			run.RunContainer(args[1:]) // Pass the command after "run" to RunContainer
		} else {
			run.RunContainer([]string{"/bin/sh"}) // Default to "/bin/sh" if no command is provided
		}
	case "build":
		config.Log.Debugf("Executing build command")
		build.Build(args[1]) // Pass the file path to the build function
	default:
		config.Log.Errorf("Unknown command: %s", command)
		config.Log.Info("Available commands: run, build")
		os.Exit(1)
	}
}
