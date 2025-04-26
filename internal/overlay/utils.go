package overlay

import (
	"fmt"
	"os"

	"github.com/lariskovski/containy/internal/config"
)

func createDirectory(paths ...string) error {
	for _, path := range paths {
		config.Log.Debugf("Creating directory: %s", path)
		// Check if the directory already exists
		if _, err := os.Stat(path); err == nil {
			// Directory exists, no need to create it
			continue
		} else if !os.IsNotExist(err) {
			// An error occurred while checking the directory
			return fmt.Errorf("failed to check directory %s: %w", path, err)
		}
		// Directory does not exist, create it
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}
