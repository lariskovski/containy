package container

import (
	"os"
	"syscall"

	"github.com/lariskovski/containy/internal/config"
)

func setupPivotRoot(overlayDir string) error {
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
