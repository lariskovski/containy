package instructions

import (
	"fmt"
	"strings"

	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/utils"
)

// Instruction represents a single directive in a container build file.
// Each instruction describes an action to take when building the container image,
// such as setting up the base filesystem (FROM), running commands (RUN)
//
// The Instruction interface allows for a pluggable architecture where
// different types of instructions can be implemented and processed uniformly.
type Instruction interface {
	// GetType returns the instruction type (e.g., "FROM", "RUN", "COPY")
	GetType() string

	// GetArgs returns the instruction arguments as a string
	// (e.g., for "RUN apt-get update", it returns "apt-get update")
	GetArgs() string
}

// BuildState maintains context during a container image build.
// It tracks the current layer and instruction being processed,
// allowing instructions to build upon previous ones.
type BuildState struct {
	// CurrentLayer represents the most recently created filesystem layer
	CurrentLayer Layer

	// Instruction stores the type of the most recently executed instruction
	Instruction string
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
func Execute(instructions []Instruction) error {
	config.Log.Info("Executing instructions")

	buildState := &BuildState{}

	for _, instruction := range instructions {
		instructionType := instruction.GetType()
		if !isValidCommand(instructionType) {
			return fmt.Errorf("unknown command: %s", instructionType)
		}
		instructionArgs := instruction.GetArgs()

		// check if layer exists
		id := utils.GenerateHexID(strings.Join([]string{instructionType, instructionArgs}, " "))
		if LayerExists(id) {
			config.Log.Infof("Layer already exists for instruction: %s", instructionType)
			continue
		}

		config.Log.Infof("Executing instruction: %s with args: %s", instructionType, instructionArgs)

		// Execute the instruction using the appropriate handler
		handler := handlers[instructionType]
		if err := handler(instructionArgs, buildState); err != nil {
			return fmt.Errorf("%s failed: %w", instructionType, err)
		}
	}

	return nil
}

// isValidCommand checks if an instruction type is supported by the system.
// It verifies the instruction against the handlers map to determine if
// there's an implementation available for the instruction.
//
// Parameters:
//   - cmd: The instruction type to check (e.g., "FROM", "RUN")
//
// Returns:
//   - bool: true if the instruction is supported, false otherwise
func isValidCommand(cmd string) bool {
	_, ok := handlers[cmd]
	return ok
}
