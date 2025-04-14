package main

import (
	"fmt"
	"os"

	"github.com/lariskovski/containy/internal/run"
)

func main() {
	args := os.Args[1:] // Get command-line arguments (excluding the program name)
	if len(args) == 0 {
		fmt.Println("Usage: main.go <command> [<args>]")
		os.Exit(1)
	}

	command := args[0]
	switch command {
	case "run":
		if len(args) > 1 {
			run.RunContainer(args[1:]) // Pass the command after "run" to RunContainer
		} else {
			run.RunContainer([]string{"/bin/sh"}) // Default to "/bin/sh" if no command is provided
		}
	case "build":
		// Placeholder for the build command
		fmt.Println("Build command executed with args:", args[1:])
		// Add logic for the build command here
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: run, build")
		os.Exit(1)
	}
}
