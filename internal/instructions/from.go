package instructions

import (
	"fmt"

	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/overlay"
	"github.com/lariskovski/containy/internal/utils"
)


func from(arg string, state *BuildState) error {
	config.Log.Debugf("Processing FROM instruction with argument: %s", arg)

	inst := "FROM " + arg

	// Create overlay filesystem for the base image
	temp := overlay.OverlayFS{
		Instruction: "FROM",
		ID:          utils.GenerateHexID(inst),
		LowerDir:    "lower",
		UpperDir:    "upper",
		WorkDir:     "work",
		MergedDir:   "merged",
	}
	overlayFS, err := temp.Setup()
	if err != nil {
		config.Log.Errorf("Failed to setup overlay filesystem: %v", err)
		return fmt.Errorf("failed to setup overlay filesystem: %w", err)
	}

	err = utils.DownloadRootFS(arg, overlayFS.LowerDir)
	if err != nil {
		config.Log.Errorf("Failed to download root filesystem: %v", err)
		return fmt.Errorf("failed to download root filesystem: %w", err)
	}
	err = overlayFS.Mount()
	if err != nil {
		config.Log.Errorf("Failed to mount overlay filesystem: %v", err)
		return fmt.Errorf("failed to mount overlay filesystem: %w", err)
	}

	// Update the state with the current layer
	state.CurrentLayer = *overlayFS
	config.Log.Debugf("Overlay filesystem mounted successfully at %s", overlayFS.MergedDir)
	return nil
}
