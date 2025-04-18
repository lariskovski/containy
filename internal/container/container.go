package container

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/lariskovski/containy/internal/config"
)

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

func handleChildProcess(overlayDir string, commandArgs []string) {
	config.Log.Debugf("In child process")

	if err := setupNamespaces(overlayDir); err != nil {
		config.Log.Fatalf("Error setting up namespaces: %v", err)
	}

	cmd, err := execCommand(overlayDir, commandArgs, false); 
	if err != nil {
		config.Log.Fatalf("Error creating command: %v", err)
	}

	if err := cmd.Run(); err != nil {
		config.Log.Fatalf("Error running command: %v", err)
	}
}

func execCommand(overlayDir string, commandArgs []string, spawnChild bool) (*exec.Cmd, error) {
	config.Log.Debugf("Running command: %v", commandArgs)
	
	var cmd *exec.Cmd

	if spawnChild {
		cmd = exec.Command("/proc/self/exe", append([]string{"run", overlayDir}, commandArgs...)...)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWUTS |
				syscall.CLONE_NEWNS |
				syscall.CLONE_NEWPID,
			Unshareflags: syscall.CLONE_NEWNS,
		}
	} else {
		commandStr := strings.Join(commandArgs, " ")
		cmd = exec.Command("/bin/sh", "-c", commandStr)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return  cmd, nil
}

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

func logError(context string, err error) error {
	config.Log.Errorf("Error %s: %v", context, err)
	return err
}
