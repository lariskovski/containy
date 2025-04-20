package instructions

import (
	"fmt"
	"strings"

	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/container"
	"github.com/lariskovski/containy/internal/utils"
)

// runCmd implements the RUN instruction from a container build file.
// It executes commands in a new container layer and captures any changes
// to the filesystem.
//
// The function:
// 1. Creates a unique layer ID based on the RUN command
// 2. Builds the proper lowerdir path based on previous layers
// 3. Creates and mounts a new overlay filesystem
// 4. Executes the specified command inside the container
// 5. Updates the build state with the new layer information
//
// Parameters:
//   - arg: The command to execute (e.g., "apt-get update")
//   - state: The current build state containing layer information
//
// Returns:
//   - error: Any error encountered during the process
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

// buildLowerDir constructs the lowerdir path for overlayfs mounting.
//
// The lowerdir for a RUN instruction depends on the previous instruction:
//   - After FROM: Uses the lower directory of the base image
//   - After other instructions: Chains the current layer's upper directory
//     with previous lower directories
//
// Parameters:
//   - state: The current build state containing layer information
//
// Returns:
//   - string: The formatted lowerdir path for overlayfs mount
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

// prepareCommandArgs constructs the argument slice for container execution.
//
// This function prepends the container's merged directory path to the
// command arguments, enabling the container runtime to execute the command
// in the correct filesystem context.
//
// Parameters:
//   - mergedDir: The path to the merged overlay filesystem
//   - arg: The raw command string to be executed
//
// Returns:
//   - []string: A slice containing the merged directory followed by command arguments
func prepareCommandArgs(mergedDir, arg string) []string {
	args := strings.Fields(arg)
	return append([]string{mergedDir}, args...)
}
