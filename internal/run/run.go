package run

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var defaultPATH = "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

func RunContainer(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: run <overlay-dir> <command> [args...]\n")
		os.Exit(1)
	}

	overlayDir := args[0]   // First argument is the overlay directory
	commandArgs := args[1:] // Remaining arguments are the command and its arguments

	if os.Args[0] == "/proc/self/exe" {
		handleChildProcess(overlayDir, commandArgs)
		return
	}

	spawnChildProcess(overlayDir, commandArgs)
}

func handleChildProcess(overlayDir string, commandArgs []string) {
	fmt.Println("In child process:")

	if err := setupNamespaces(overlayDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up namespaces: %v\n", err)
		os.Exit(1)
	}

	if err := runCommand(commandArgs); err != nil {
		fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
		os.Exit(1)
	}
}

func setupNamespaces(overlayDir string) error {
	if err := syscall.Sethostname([]byte("container")); err != nil {
		return fmt.Errorf("setting hostname: %w", err)
	}

	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("making mount private: %w", err)
	}

	if err := performPivotRoot(overlayDir); err != nil {
		return fmt.Errorf("performing pivot_root: %w", err)
	}

	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("remounting /proc: %w", err)
	}

	return os.Setenv("PATH", defaultPATH)
}

func performPivotRoot(overlayDir string) error {
	oldRoot := overlayDir + "/oldroot"
	if err := os.MkdirAll(oldRoot, 0755); err != nil {
		return fmt.Errorf("creating oldroot directory: %w", err)
	}

	if err := syscall.PivotRoot(overlayDir, oldRoot); err != nil {
		return fmt.Errorf("pivot_root: %w", err)
	}

	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("changing directory: %w", err)
	}

	if err := syscall.Unmount("oldroot", syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmounting old root: %w", err)
	}

	return os.Remove("oldroot")
}

func runCommand(commandArgs []string) error {
	fmt.Println("Running command:", commandArgs)
	commandStr := strings.Join(commandArgs, " ")
	cmd := exec.Command("/bin/sh", "-c", commandStr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func spawnChildProcess(overlayDir string, commandArgs []string) {
	fmt.Println("Spawning child with new namespaces...")

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
		fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Child process finished.")
}
