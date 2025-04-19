package instructions

import (
	"fmt"

	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/overlay"
	"github.com/lariskovski/containy/internal/utils"
)

func from(arg string, state *BuildState) error {
	config.Log.Infof("Processing FROM instruction with argument: %s", arg)

	inst := "FROM " + arg
	id := utils.GenerateHexID(inst)

	// Create and setup overlay filesystem in one step
	overlayFS, err := overlay.NewOverlayFS("", id, true)
	if err != nil {
		config.Log.Errorf("Failed to create overlay filesystem: %v", err)
		return fmt.Errorf("failed to create overlay filesystem: %w", err)
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
	state.Instruction = "FROM"

	config.Log.Debugf("Overlay filesystem mounted successfully at %s", overlayFS.MergedDir)
	return nil
}
