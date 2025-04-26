package overlay

import (
	"fmt"
	"os"

	"github.com/lariskovski/containy/internal/config"
)

// CheckIfLayerExists determines if a layer with the given ID already exists on disk.
// This is used for layer caching during builds.
//
// Parameters:
//   - id: The unique layer identifier to check
//
// Returns:
//   - bool: true if the layer exists, false otherwise
func CheckIfLayerExists(id string) bool {
	basePath := config.BaseOverlayDir + id + "/"
	config.Log.Debugf("Checking if layer exists at path: %s", basePath)
	// Check if the directory exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return false // Directory does not exist
	}
	return true // Directory exists
}

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
