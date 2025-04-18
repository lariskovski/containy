package container

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/lariskovski/containy/internal/config"
)

// Container namespace flags for isolating the container environment.
const containerNamespaceFlags = syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID

// Create initializes and runs a new container process.
// It takes a slice of arguments where args[0] is the overlay directory path
// and the remaining args are the command and its arguments to be executed
// inside the container. If the current process is the child process
// (identified by /proc/self/exe), it handles the child process setup.
// Otherwise, it spawns a new child process.
func Create(args []string) {
	if len(args) < 2 {
		config.Log.Errorf("Usage: run <overlay-dir> <command> [args...]")
		os.Exit(1)
	}

	overlayDir := args[0]
	commandArgs := args[1:]

	if os.Args[0] == "/proc/self/exe" {
		handleChildProcess(overlayDir, commandArgs)
		return
	}

	spawnChildProcess(overlayDir, commandArgs)
}

// spawnChildProcess creates a new child process with isolated namespaces.
// It calls itself with the overlay directory and command arguments.
// The child process will run the specified command in a new environment.
func spawnChildProcess(overlayDir string, commandArgs []string) {
	config.Log.Debugf("Spawning child with new namespaces")
	cmd, err := execCommand(overlayDir, commandArgs, true)
	if err != nil {
		config.Log.Fatalf("Error creating command: %v", err)
	}
	if err := cmd.Run(); err != nil {
		config.Log.Fatalf("Error running command: %v", err)
	}
	config.Log.Debugf("Child process finished")
}

// handleChildProcess sets up the container environment in the child process.
// It sets up the necessary namespaces, the overlay filesystem,
// and executes the specified command within the containerized environment.
func handleChildProcess(overlayDir string, commandArgs []string) {
	config.Log.Debugf("In child process")

	if err := setupNamespaces(overlayDir); err != nil {
		config.Log.Fatalf("Error setting up namespaces: %v", err)
	}

	cmd, err := execCommand(overlayDir, commandArgs, false)
	if err != nil {
		config.Log.Fatalf("Error creating command: %v", err)
	}

	if err := cmd.Run(); err != nil {
		config.Log.Fatalf("Error running command: %v", err)
	}
}

// execCommand creates an exec.Cmd instance for either spawning a child process
// or executing a command within the container. When spawnChild is true, it
// creates a command that will re-execute the current binary with namespace
// isolation. When false, it creates a command to run within the container.
func execCommand(overlayDir string, commandArgs []string, spawnChild bool) (*exec.Cmd, error) {
	config.Log.Debugf("Running command: %v", commandArgs)

	var cmd *exec.Cmd

	if spawnChild {
		cmd = exec.Command("/proc/self/exe", append([]string{"run", overlayDir}, commandArgs...)...)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: containerNamespaceFlags,
			Unshareflags: syscall.CLONE_NEWNS,
		}
	} else {
		cmd = exec.Command("/bin/sh", "-c", strings.Join(commandArgs, " "))
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}

// logError creates a formatted error message, logs it using the configured
// logger, and returns the original error for further error handling.
func logError(context string, err error) error {
	config.Log.Errorf("Error %s: %v", context, err)
	return err
}
