package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/lariskovski/containy/internal/config"
)

// containerNamespaceFlags defines the Linux namespaces to isolate for containers.
// These flags determine what aspects of the system are isolated:
// - CLONE_NEWUTS: Hostname and domain name
// - CLONE_NEWNS: Mount points
// - CLONE_NEWPID: Process IDs
const containerNamespaceFlags = syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID

// Create initializes and runs a new container.
// This function is the main entry point for container creation and execution.
// It handles both the parent and child processes in a fork/exec pattern.
//
// The function detects whether it's running in the parent or child process:
// - In the parent: It spawns a child process with namespace isolation
// - In the child: It sets up the containerized environment and runs the command
//
// Parameters:
//   - args: A slice where args[0] is the overlay directory path and
//     the remaining elements are the command and its arguments
//
// The function will terminate the process if errors are encountered.
func Create(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("insufficient arguments: expected at least overlay directory and command")
	}

	overlayDir := args[0]
	commandArgs := args[1:]

	// Check if the overlay directory is an alias
	// If it is, resolve the alias to the actual directory
	alias, err := os.Readlink(overlayDir)
	if err == nil  {
		overlayDir = alias
	}

	// Check if the overlay directory exists
	if _, err := os.Stat(overlayDir); os.IsNotExist(err) {
		return fmt.Errorf("overlay directory does not exist: %s", overlayDir)
	}

	// /proc/self/exe is the current executable this is used to re-execute
	// the current binary in the child process This is a common pattern in 
	// container runtimes to re-execute the current binary with new namespaces
	if os.Args[0] == "/proc/self/exe" {
		return handleChildProcess(overlayDir, commandArgs)
	}

	return spawnChildProcess(overlayDir, commandArgs)
}

// spawnChildProcess creates a new isolated process for the container.
// It uses Linux namespace isolation to create a containerized environment,
// then re-executes the current binary to set up the container.
//
// Parameters:
//   - overlayDir: Path to the overlay filesystem's merged directory
//   - commandArgs: The command and arguments to run inside the container
func spawnChildProcess(overlayDir string, commandArgs []string) error {
	config.Log.Debugf("Spawning child with new namespaces")
	cmd, err := execCommand(overlayDir, commandArgs, true)
	if err != nil {
		return fmt.Errorf("error creating command: %w", err)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running command: %w", err)
	}
	config.Log.Debugf("Child process finished")
	return nil
}

// handleChildProcess sets up the containerized environment and executes
// the specified command within it. This function runs in the child process
// after namespace isolation.
//
// It performs the following container setup:
// 1. Sets up namespaces (hostname, mount, etc.)
// 2. Configures the filesystem view via pivot_root
// 3. Mounts /proc and sets up the PATH environment
// 4. Executes the specified command
//
// Parameters:
//   - overlayDir: Path to the overlay filesystem's merged directory
//   - commandArgs: The command and arguments to run inside the container
func handleChildProcess(overlayDir string, commandArgs []string) error {
	config.Log.Debugf("In child process")

	if err := setupNamespaces(overlayDir); err != nil {
		return fmt.Errorf("error setting up namespaces: %w", err)
	}

	cmd, err := execCommand(overlayDir, commandArgs, false)
	if err != nil {
		return fmt.Errorf("error creating command: %w", err)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running command: %w", err)
	}

	return nil
}

// execCommand creates an exec.Cmd instance for container execution.
// It handles two different scenarios based on the spawnChild parameter:
//
//  1. When spawnChild is true:
//     Creates a command that re-executes the current binary with namespace
//     isolation flags to start the container
//
//  2. When spawnChild is false:
//     Creates a command that executes the specified command inside the
//     already-prepared container environment
//
// Parameters:
//   - overlayDir: Path to the overlay filesystem's merged directory
//   - commandArgs: The command and arguments to run
//   - spawnChild: Whether to create a child process with namespace isolation
//
// Returns:
//   - *exec.Cmd: The prepared command ready for execution
//   - error: Any error encountered during command creation
func execCommand(overlayDir string, commandArgs []string, spawnChild bool) (*exec.Cmd, error) {
	config.Log.Debugf("Running command: %v", commandArgs)

	var cmd *exec.Cmd

	if spawnChild {
		cmd = exec.Command("/proc/self/exe", append([]string{"run", overlayDir}, commandArgs...)...)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags:   containerNamespaceFlags,
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

// logError formats and logs an error message.
// It provides context about where the error occurred and returns the original
// error for further handling.
//
// Parameters:
//   - context: A string describing where the error occurred
//   - err: The original error object
//
// Returns:
//   - error: The original error, allowing for further handling
func logError(context string, err error) error {
	config.Log.Errorf("Error %s: %v", context, err)
	return err
}
