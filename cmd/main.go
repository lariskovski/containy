package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/lariskovski/containy/internal/overlay"
	"github.com/lariskovski/containy/internal/utils"
)

var rootfsUrl = "https://dl-cdn.alpinelinux.org/alpine/v3.21/releases/x86_64/alpine-minirootfs-3.21.3-x86_64.tar.gz"

func main() {
	if os.Args[0] == "/proc/self/exe" {
		fmt.Println("In child process:")
		must(syscall.Sethostname([]byte("container")))

		// Perform pivot_root
		must(os.MkdirAll("merged/oldroot", 0755))
		must(syscall.PivotRoot("merged", "merged/oldroot"))
		must(os.Chdir("/")) // Change working directory to new root

		// Unmount old root
		must(syscall.Unmount("oldroot", syscall.MNT_DETACH))
		must(os.Remove("oldroot"))

		// Remount /proc in the new root
		if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
			fmt.Fprintf(os.Stderr, "Error remounting /proc: %v\n", err)
			os.Exit(1)
		}

		// Set PATH environment variable
		must(os.Setenv("PATH", "/bin:"+os.Getenv("PATH")))

		// Start a shell
		cmd := exec.Command("sh")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		must(cmd.Run())
		return
	}

	fsSetup()

	fmt.Println("Spawning child with new namespaces...")

	cmd := exec.Command("/proc/self/exe") // self-exec trick
	cmd.Args = []string{"/proc/self/exe"}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWPID,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	must(cmd.Run())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func fsSetup() {
	err := utils.CreateDirectory("lower", "upper", "work", "merged")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directories: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Directories created successfully.")

	// Download the root filesystem
	err = utils.DownloadRootFS(rootfsUrl, "lower")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading root filesystem: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Root filesystem downloaded successfully.")

	// Create the overlay filesystem
	overlayFS := &overlay.OverlayFS{
		LowerDir:  "lower",
		UpperDir:  "upper",
		WorkDir:   "work",
		MergedDir: "merged",
	}
	err = overlayFS.Mount()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating overlay filesystem: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Overlay filesystem created successfully.")
}
