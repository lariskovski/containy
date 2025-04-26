package build

import (
	"fmt"

	"github.com/lariskovski/containy/internal/overlay"
)

type Layer interface {
	// GetID returns the ID of the layer
	GetID() string
	// GetLowerDir returns the lower directory of the layer
	GetLowerDir() string
	// GetUpperDir returns the upper directory of the layer
	GetUpperDir() string
	// GetWorkDir returns the work directory of the layer
	GetWorkDir() string
	// GetMergedDir returns the merged directory of the layer
	GetMergedDir() string
	// Mount mounts the overlay filesystem
	Mount() error
	// CreateAlias creates an alias for the layer
	CreateAlias(alias string) error
}

func AddNewLayer(lowerDir, id string) (Layer, error) {
	// Create the base directory for the layers
	layer, err := overlay.NewOverlayFS(lowerDir, id)
	if err != nil {
		return nil, fmt.Errorf("failed to setup layer: %w", err)
	}

	if err := layer.Mount(); err != nil {
		return nil, fmt.Errorf("failed to mount layer: %w", err)
	}

	return layer, nil
}

func AddBaseLayer(id, fsURL string) (Layer, error) {
	layer, err := overlay.NewOverlayFS("", id)
	if err != nil {
		return nil, fmt.Errorf("failed to setup base layer: %w", err)
	}

	err = DownloadRootFS(fsURL, layer.GetLowerDir())
	if err != nil {
		return nil, fmt.Errorf("failed to download root filesystem: %w", err)
	}

	if err := layer.Mount(); err != nil {
		return nil, fmt.Errorf("failed to mount base layer: %w", err)
	}

	return layer, nil
}
