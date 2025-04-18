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
	// "COPY": copyCmd,
	// "CMD":  cmd,
}

func ValidateAndConvertLines(lines []parser.Line) ([]parser.Line, error) {
	var instructions []parser.Line
	for _, line := range lines {
		if !isValidCommand(line.Type) {
			return nil, fmt.Errorf("unknown command: %s", line.Type)
		}
		instructions = append(instructions, parser.Line{Type: line.Type, Args: line.Args})
	}
	return instructions, nil
}

func isValidCommand(cmd string) bool {
	_, ok := handlers[cmd]
	return ok
}

func ExecuteInstructions(lines []parser.Line) error {
	config.Log.Info("Executing instructions")
	instructions, err := ValidateAndConvertLines(lines)
	if err != nil {
		return err
	}

	state := &BuildState{}
	for _, instr := range instructions {
		handler := handlers[instr.Type]
		if err := handler(instr.Args, state); err != nil {
			config.Log.Errorf("%s failed: %v", instr.Type, err)
			return fmt.Errorf("%s failed: %w", instr.Type, err)
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
