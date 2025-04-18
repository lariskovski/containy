package instructions

import (
	"fmt"
	"strings"

	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/overlay"
	"github.com/lariskovski/containy/internal/run"
	"github.com/lariskovski/containy/internal/utils"
)


func runCmd(arg string, state *BuildState) error {
	config.Log.Infof("Processing RUN instruction with argument: %s", arg)

	inst := "RUN " + arg

	// Use the current layer's merged directory as the base for the next layer's lower directory
	var previousLayer string
	if state.CurrentLayer.Instruction == "FROM" {
		previousLayer = state.CurrentLayer.LowerDir
	} else {
		previousLayer = state.CurrentLayer.UpperDir
	}
	newLowerDir := previousLayer
	if state.CurrentLayer.LowerDir != "" && state.CurrentLayer.Instruction != "FROM" {
		newLowerDir = state.CurrentLayer.LowerDir + ":" + previousLayer
		config.Log.Debugf("New lower directory: %s", newLowerDir)
	}

	ofs := overlay.OverlayFS{
		Instruction: "RUN",
		ID:          utils.GenerateHexID(inst),
		LowerDir:    newLowerDir,
		UpperDir:    "upper",
		WorkDir:     "work",
		MergedDir:   "merged",
	}

	// Create a new overlay filesystem for the command
	overlayFS, err := ofs.Setup()
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

	run.RunContainer(command)

	// Update the state with the current layer
	state.CurrentLayer = *overlayFS

	return nil
}