package container

import (
	"os"
	"syscall"

	"github.com/lariskovski/containy/internal/config"
)

// setupNamespaces sets up the necessary namespaces for the container environment
func setupNamespaces(overlayDir string) error {
	config.Log.Debugf("Setting up namespaces in overlayDir: %s", overlayDir)

	if err := syscall.Sethostname([]byte("container")); err != nil {
		return logError("setting hostname", err)
	}

	// makes the mount namespace private
	// this is a workaround for the fact that the mount namespace
	// is not private by default in some Linux distributions
	// and to prevent the host from seeing mounts made by the container
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return logError("making mount private", err)
	}

	if err := setupPivotRoot(overlayDir); err != nil {
		return logError("performing pivot_root", err)
	}

	// proc filesystem is used for process information
	// and is required for the container to function properly
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return logError("remounting /proc", err)
	}

	// path is used to find executables
	// and is required for the container to function properly
	return os.Setenv("PATH", config.DefaultPATH)
}
