package build

import (
	"fmt"

	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/container"
	"github.com/lariskovski/containy/internal/overlay"
)

// Instruction represents a single directive in a container build file.
// Each instruction describes an action to take when building the container image,
// such as setting up the base filesystem (FROM), running commands (RUN)
//
// The Instruction interface allows for a pluggable architecture where
// different types of instructions can be implemented and processed uniformly.
type Instruction struct {
	// GetType returns the instruction type (e.g., "FROM", "RUN", "COPY")
	Type string

	// GetArgs returns the instruction arguments as a string
	// (e.g., for "RUN apt-get update", it returns "apt-get update")
	Args string
}

// handlers maps instruction types to their implementation functions.
// To add support for a new instruction type, add an entry to this map
// with a handler function that implements the instruction's behavior.
var handlers = map[string]func(string, *BuildState) error{
	"FROM": from,
	"RUN":  runCmd,
	// "COPY": copyCmd,
	// "CMD":  cmd,
}

// Execute processes a sequence of build instructions to create a container image.
// It iterates through each instruction, checks its validity, and invokes
// the appropriate handler function with the instruction's arguments.
//
// For each instruction, it also:
// 1. Generates a unique layer ID based on the instruction
// 2. Checks if this layer has already been built (for caching)
// 3. Updates the build state after each instruction
//
// Parameters:
//   - instructions: A slice of Instruction objects to execute in order
//
// Returns:
//   - error: Any error encountered during execution, or nil on success
func execute(instruction Instruction, state *BuildState) error {
	instructionType := instruction.GetType()
	instructionArgs := instruction.GetArgs()

	// Execute the instruction using the appropriate handler
	handler := handlers[instructionType]
	if err := handler(instructionArgs, state); err != nil {
		return fmt.Errorf("%s failed: %w", instructionType, err)
	}
	return nil
}

// The FROM instruction specifies the base image to use for the container.
// Sets up the base layer for the container image by downloading and mounting the specified root filesystem.
// It creates a new layer and mounts it to the specified directory.
// The function also updates the BuildState with the current layer and instruction.
func from(arg string, state *BuildState) error {
	config.Log.Infof("Processing FROM instruction with argument: %s", arg)

	inst := "FROM " + arg
	id := GenerateHexID(inst)

	// Create and setup overlay filesystem in one step using the Layer abstraction
	layer, err := overlay.NewOverlayFS("", id, true)
	if err != nil {
		return fmt.Errorf("failed to create overlay filesystem: %w", err)
	}

	err = DownloadRootFS(arg, layer.GetLowerDir())
	if err != nil {
		return fmt.Errorf("failed to download root filesystem: %w", err)
	}

	err = layer.Mount()
	if err != nil {
		return fmt.Errorf("failed to mount overlay filesystem: %w", err)
	}

	// Update the state with the current layer
	state.CurrentLayer = layer
	state.Instruction = "FROM"

	config.Log.Debugf("Overlay filesystem mounted successfully at %s", layer.GetMergedDir())
	return nil
}

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
	config.Log.Debugf("Processing RUN instruction with argument: %s", arg)

	inst := "RUN " + arg
	id := GenerateHexID(inst)

	newLowerDir := buildLowerDir(state)
	layer, err := overlay.NewOverlayFS(newLowerDir, id, false)
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
