package container

import (
	"os"
	"syscall"

	"github.com/lariskovski/containy/internal/config"
)

// setupPivotRoot sets up the pivot root for the container.
// It changes the root filesystem of the calling process to the specified overlay directory.
// This is a crucial step in containerization, as it isolates the container's filesystem from the host's filesystem.
func setupPivotRoot(overlayDir string) error {
	config.Log.Debugf("Performing pivot_root with overlayDir: %s", overlayDir)

	// Create the oldroot directory where the old root will be moved
	oldRoot := overlayDir + "/oldroot"
	if err := os.MkdirAll(oldRoot, 0755); err != nil {
		return logError("creating oldroot directory", err)
	}

	// Move the current root filesystem to the oldroot directory
	// This is necessary because pivot_root requires the new root to be a mount point
	// and the old root to be a directory
	if err := syscall.PivotRoot(overlayDir, oldRoot); err != nil {
		return logError("pivot_root", err)
	}

	// Change the working directory to the new root
	// This is necessary because the current working directory is still in the old root
	// After pivot_root, the old root is no longer accessible
	// and we need to change the working directory to the new root
	// to avoid any issues with file access
	if err := os.Chdir("/"); err != nil {
		return logError("changing directory", err)
	}

	// Unmount the old root filesystem
	if err := syscall.Unmount("oldroot", syscall.MNT_DETACH); err != nil {
		return logError("unmounting old root", err)
	}

	return os.Remove("oldroot")
}
