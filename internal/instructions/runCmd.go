package instructions

import (
	"fmt"
	"strings"

	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/container"
	"github.com/lariskovski/containy/internal/utils"
)

// The RUN instruction executes a command in the container's filesystem.
// It creates a new layer on top of the current layer and mounts it.
// The function also updates the BuildState with the current layer and instruction.
func runCmd(arg string, state *BuildState) error {
	config.Log.Infof("Processing RUN instruction with argument: %s", arg)

	inst := "RUN " + arg
	id := utils.GenerateHexID(inst)

	// Use the current layer's merged directory as the base for the next layer's lower directory
	var previousLayer string
	if state.Instruction == "FROM" {
		previousLayer = state.CurrentLayer.GetLowerDir()
	} else {
		previousLayer = state.CurrentLayer.GetUpperDir()
	}
	newLowerDir := previousLayer
	if state.CurrentLayer.GetLowerDir() != "" && state.Instruction != "FROM" {
		newLowerDir = state.CurrentLayer.GetLowerDir() + ":" + previousLayer
		config.Log.Debugf("New lower directory: %s", newLowerDir)
	}

	config.Log.Debugf("Creating new layer with lowerDir: %s", newLowerDir)
	layer, err := NewLayer(newLowerDir, id, false)
	if err != nil {
		config.Log.Errorf("Failed to setup layer: %v", err)
		return fmt.Errorf("failed to setup layer: %w", err)
	}
	config.Log.Infof("Executing command in layer: %s", layer.GetMergedDir())
	err = layer.Mount()
	if err != nil {
		config.Log.Errorf("Failed to mount layer: %v", err)
		return fmt.Errorf("failed to mount layer: %w", err)
	}

	// Split the arg string into a slice of strings
	args := strings.Fields(arg)

	// Prepend the layer merged dir to the args slice
	command := append([]string{layer.GetMergedDir()}, args...)

	container.Create(command)

	// Update the state with the current layer
	state.CurrentLayer = layer
	state.Instruction = "RUN"

	return nil
}
