package run

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func RunContainer(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: run <overlay-dir> <command> [args...]\n")
		os.Exit(1)
	}

	overlayDir := args[0]   // First argument is the overlay directory
	commandArgs := args[1:] // Remaining arguments are the command and its arguments

	if os.Args[0] == "/proc/self/exe" {
		fmt.Println("In child process:")

		err := syscall.Sethostname([]byte("container"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting hostname: %v\n", err)
			os.Exit(1)
		}

		// Perform pivot_root
		err = os.MkdirAll(overlayDir+"/oldroot", 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating oldroot directory: %v\n", err)
			os.Exit(1)
		}

		err = syscall.PivotRoot(overlayDir, overlayDir+"/oldroot")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error performing pivot_root: %v\n", err)
			os.Exit(1)
		}
		err = os.Chdir("/") // Change working directory to new root
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error changing directory: %v\n", err)
			os.Exit(1)
		}

		// Unmount old root
		err = syscall.Unmount("oldroot", syscall.MNT_DETACH)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error unmounting old root: %v\n", err)
			os.Exit(1)
		}
		err = os.Remove("oldroot")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error removing old root: %v\n", err)
			os.Exit(1)
		}

		// Remount /proc in the new root
		if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
			fmt.Fprintf(os.Stderr, "Error remounting /proc: %v\n", err)
			os.Exit(1)
		}

		// Set PATH environment variable
		err = os.Setenv("PATH", "/bin:/sbin:"+os.Getenv("PATH"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting PATH: %v\n", err)
			os.Exit(1)
		}

		// Start the specified command or a shell
		fmt.Println("Running command:", commandArgs)
		commandStr := strings.Join(commandArgs, " ")
		cmd := exec.Command("/bin/sh", "-c", commandStr)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Println("Spawning child with new namespaces...")

	// Pass "run", the overlay directory, and the arguments to the child process
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

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Child process finished.")
}
