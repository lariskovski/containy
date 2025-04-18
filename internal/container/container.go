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

	if err := runCommand(commandArgs); err != nil {
		config.Log.Fatalf("Error running command: %v", err)
	}
}

func setupNamespaces(overlayDir string) error {
	config.Log.Debugf("Setting up namespaces")

	if err := syscall.Sethostname([]byte("container")); err != nil {
		return logError("setting hostname", err)
	}

	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return logError("making mount private", err)
	}

	if err := performPivotRoot(overlayDir); err != nil {
		return logError("performing pivot_root", err)
	}

	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return logError("remounting /proc", err)
	}

	return os.Setenv("PATH", config.DefaultPATH)
}

func performPivotRoot(overlayDir string) error {
	config.Log.Debugf("Performing pivot_root with overlayDir: %s", overlayDir)

	oldRoot := overlayDir + "/oldroot"
	if err := os.MkdirAll(oldRoot, 0755); err != nil {
		return logError("creating oldroot directory", err)
	}

	if err := syscall.PivotRoot(overlayDir, oldRoot); err != nil {
		return logError("pivot_root", err)
	}

	if err := os.Chdir("/"); err != nil {
		return logError("changing directory", err)
	}

	if err := syscall.Unmount("oldroot", syscall.MNT_DETACH); err != nil {
		return logError("unmounting old root", err)
	}

	return os.Remove("oldroot")
}

func runCommand(commandArgs []string) error {
	config.Log.Debugf("Running command: %v", commandArgs)
	commandStr := strings.Join(commandArgs, " ")
	cmd := exec.Command("/bin/sh", "-c", commandStr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func spawnChildProcess(overlayDir string, commandArgs []string) {
	config.Log.Debugf("Spawning child with new namespaces")

	cmd := exec.Command("/proc/self/exe", append([]string{"run", overlayDir}, commandArgs...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWPID,
		Unshareflags: syscall.CLONE_NEWNS,
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
