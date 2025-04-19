package instructions

import (
	"fmt"
	"strings"

	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/overlay"
	"github.com/lariskovski/containy/internal/container"
	"github.com/lariskovski/containy/internal/utils"
)


func runCmd(arg string, state *BuildState) error {
	config.Log.Infof("Processing RUN instruction with argument: %s", arg)

	inst := "RUN " + arg
	id := utils.GenerateHexID(inst)

	// Use the current layer's merged directory as the base for the next layer's lower directory
	var previousLayer string
	if state.Instruction == "FROM" {
		previousLayer = state.CurrentLayer.LowerDir
	} else {
		previousLayer = state.CurrentLayer.UpperDir
	}
	newLowerDir := previousLayer
	if state.CurrentLayer.LowerDir != "" && state.Instruction != "FROM" {
		newLowerDir = state.CurrentLayer.LowerDir + ":" + previousLayer
		config.Log.Debugf("New lower directory: %s", newLowerDir)
	}

	config.Log.Debugf("Creating new overlay filesystem with lowerDir: %s", newLowerDir)
	overlayFS, err := overlay.NewOverlayFS(newLowerDir, id, false)
	if err != nil {
		config.Log.Errorf("Failed to setup overlay filesystem: %v", err)
		return fmt.Errorf("failed to setup overlay filesystem: %w", err)
	}
	config.Log.Infof("Executing command in layer: %s", overlayFS.MergedDir)
	err = overlayFS.Mount()
	if err != nil {
		config.Log.Errorf("Failed to mount overlay filesystem: %v", err)
		return fmt.Errorf("failed to mount overlay filesystem: %w", err)
	}

	// Split the arg string into a slice of strings
	args := strings.Fields(arg)

	// Prepend the overlayFS.LowerDir to the args slice
	command := append([]string{overlayFS.MergedDir}, args...)

	container.Create(command)

	// Update the state with the current layer
	state.CurrentLayer = *overlayFS
	state.Instruction = "RUN"

	return nil
}