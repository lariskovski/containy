package instructions

import (
	"fmt"
	"strings"

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
	state := &BuildState{}
	for _, instr := range instructions {
		handler, ok := handlers[instr.Command]
		if !ok {
			return fmt.Errorf("unknown instruction: %s", instr.Command)
		}
		if err := handler(instr.Args, state); err != nil {
			return fmt.Errorf("%s failed: %w", instr.Command, err)
		}
	}
	return nil
}

func from(arg string, state *BuildState) error {
	fmt.Println("Using base image:", arg)

	inst := "FROM " + arg

	// Create overlay filesystem for the base image
	temp := overlay.OverlayFS{
		Instruction: "FROM",
		ID:        utils.GenerateHexID(inst, overlay.IDLength),
		LowerDir:  "lower",
		UpperDir:  "upper",
		WorkDir:   "work",
		MergedDir: "merged",
	}
	overlayFS, err := temp.Setup(inst)
	if err != nil {
		return fmt.Errorf("failed to setup overlay filesystem: %w", err)
	}

	// Check if layer already exists
	if utils.CheckIfLayerExists(overlayFS.LowerDir) {
		fmt.Println("Layer already exists, skipping download.")
	} else {
		err = utils.DownloadRootFS(arg, overlayFS.LowerDir)
		if err != nil {
			return fmt.Errorf("failed to download root filesystem: %w", err)
		}
		err = overlayFS.Mount()
		if err != nil {
			return fmt.Errorf("failed to mount overlay filesystem: %w", err)
		}
	}

	// Update the state with the current layer
	state.CurrentLayer = *overlayFS
	fmt.Println("Overlay filesystem mounted successfully.")
	return nil
}

func runCmd(arg string, state *BuildState) error {
	fmt.Println("Running shell command:", arg)

	inst := "RUN " + arg

	// Use the current layer's merged directory as the working directory
	previousLayer := state.CurrentLayer.MergedDir
	fmt.Printf("Executing command in layer: %s\n", previousLayer)

	ofs := overlay.OverlayFS{
		Instruction: "RUN",
		ID:        utils.GenerateHexID(inst, overlay.IDLength),
		LowerDir:  previousLayer,
		UpperDir:  "upper",
		WorkDir:   "work",
		MergedDir: "merged",
	}
	overlayFS, err := ofs.Setup(arg)
	if err != nil {
		return fmt.Errorf("failed to setup overlay filesystem: %w", err)
	}
	err = overlayFS.Mount()
	if err != nil {
		return fmt.Errorf("failed to mount overlay filesystem: %w", err)
	}

    // Split the arg string into a slice of strings
    args := strings.Fields(arg)

    // Prepend the overlayFS.LowerDir to the args slice
    command := append([]string{overlayFS.MergedDir}, args...)

	run.RunContainer(command)

	return nil
}

func copyCmd(arg string, state *BuildState) error {
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
	fmt.Println("Final command (not running it yet):", arg)
	// Optional: actually run it, or simulate it.
	return nil
}
