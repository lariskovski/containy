package instructions

import (
	"fmt"
	"strings"

	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/overlay"
	"github.com/lariskovski/containy/internal/parser"
	"github.com/lariskovski/containy/internal/run"
	"github.com/lariskovski/containy/internal/utils"
)

type BuildState struct {
	CurrentLayer overlay.OverlayFS
}

var handlers = map[string]func(string, *BuildState) error{
	"FROM": from,
	"RUN":  runCmd,
	"COPY": copyCmd,
	"CMD":  cmd,
}

func ExecuteInstructions(instructions []parser.Instruction) error {
	config.Log.Info("Executing instructions")
	state := &BuildState{}
	for _, instr := range instructions {
		handler, ok := handlers[instr.Command]
		if !ok {
			config.Log.Errorf("Unknown instruction: %s", instr.Command)
			return fmt.Errorf("unknown instruction: %s", instr.Command)
		}
		if err := handler(instr.Args, state); err != nil {
			config.Log.Errorf("%s failed: %v", instr.Command, err)
			return fmt.Errorf("%s failed: %w", instr.Command, err)
		}
	}
	return nil
}

func from(arg string, state *BuildState) error {
	config.Log.Debugf("Processing FROM instruction with argument: %s", arg)

	inst := "FROM " + arg

	// Create overlay filesystem for the base image
	temp := overlay.OverlayFS{
		Instruction: "FROM",
		ID:          utils.GenerateHexID(inst, overlay.IDLength),
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
		config.Log.Debugf("New lower directory:", newLowerDir)
	}

	ofs := overlay.OverlayFS{
		Instruction: "RUN",
		ID:          utils.GenerateHexID(inst, overlay.IDLength),
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

func copyCmd(arg string, state *BuildState) error {
	config.Log.Debugf("Processing COPY instruction with argument: %s", arg)
	// parts := strings.Fields(arg)
	// if len(parts) != 2 {
	// 	return fmt.Errorf("invalid COPY args: %s", arg)
	// }

	// src, dest := parts[0], parts[1]
	// fmt.Printf("Copying %s to %s\n", src, dest)
	// input, err := os.ReadFile(src)
	// if err != nil {
	// 	return err
	// }
	// return os.WriteFile(dest, input, 0644)
	return nil
}

func cmd(arg string, state *BuildState) error {
	config.Log.Debugf("Processing CMD instruction with argument: %s", arg)
	fmt.Println("Final command (not running it yet):", arg)
	// Optional: actually run it, or simulate it.
	return nil
}
