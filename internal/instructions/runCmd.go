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

	newLowerDir := buildLowerDir(state)
	layer, err := NewLayer(newLowerDir, id, false)
	if err != nil {
		return fmt.Errorf("failed to setup layer: %w", err)
	}

	if err := layer.Mount(); err != nil {
		return fmt.Errorf("failed to mount layer: %w", err)
	}

	command := prepareCommandArgs(layer.GetMergedDir(), arg)
	// Consider: return an error if container.Create fails, instead of calling it directly
	container.Create(command)

	state.CurrentLayer = layer
	state.Instruction = "RUN"
	return nil
}

// buildLowerDir builds the lowerdir string for the new layer.
func buildLowerDir(state *BuildState) string {
	var previousLayer string
	if state.Instruction == "FROM" {
		previousLayer = state.CurrentLayer.GetLowerDir()
	} else {
		previousLayer = state.CurrentLayer.GetUpperDir()
	}
	newLowerDir := previousLayer
	if state.CurrentLayer.GetLowerDir() != "" && state.Instruction != "FROM" {
		newLowerDir = state.CurrentLayer.GetLowerDir() + ":" + previousLayer
	}
	return newLowerDir
}

// prepareCommandArgs prepares the command arguments for container execution.
func prepareCommandArgs(mergedDir, arg string) []string {
	args := strings.Fields(arg)
	return append([]string{mergedDir}, args...)
}
