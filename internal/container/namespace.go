package container

import (
	"os"
	"syscall"

	"github.com/lariskovski/containy/internal/config"
)

func setupNamespaces(overlayDir string) error {
	config.Log.Debugf("Setting up namespaces")

	if err := syscall.Sethostname([]byte("container")); err != nil {
		return logError("setting hostname", err)
	}

	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return logError("making mount private", err)
	}

	if err := setupPivotRoot(overlayDir); err != nil {
		return logError("performing pivot_root", err)
	}

	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return logError("remounting /proc", err)
	}

	return os.Setenv("PATH", config.DefaultPATH)
}
