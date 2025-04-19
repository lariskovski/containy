package instructions

import (
	"github.com/lariskovski/containy/internal/overlay"
)

// Layer abstracts a container filesystem layer.
type Layer interface {
	GetID() string
	GetLowerDir() string
	GetUpperDir() string
	GetWorkDir() string
	GetMergedDir() string
	Mount() error
}

// NewLayer abstracts the creation of a new Layer.
// Arguments are passed through to the underlying implementation.
func NewLayer(lowerDir, id string, isBaseLayer bool) (Layer, error) {
	return overlay.NewOverlayFS(lowerDir, id, isBaseLayer)
}

func LayerExists(id string) bool {
	return overlay.CheckIfLayerExists(id)
}