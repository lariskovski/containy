package build

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/overlay"
)

// BuildState maintains context during a container image build.
// It tracks the current layer and instruction being processed,
// allowing instructions to build upon previous ones.
type BuildState struct {
	// CurrentLayer represents the most recently created filesystem layer
	CurrentLayer *overlay.OverlayFS

	// Instruction stores the type of the most recently executed instruction
	Instruction string
}

// Build parses a container build file and executes its instructions to build an image.
// The file at 'filepath' should contain container build instructions (e.g., FROM, RUN).
// Each instruction is parsed, converted to the instructions.Instruction interface, and executed in order.
// If any instruction fails, the build process is aborted and an error is logged.
func Build(filepath string) error {
	config.Log.Debugf("Building container from file: %s", filepath)

	// Parse the build file into a slice of parser.Line instructions
	instructions, err := parse(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	buildState := &BuildState{}

	// Iterate through the instructions and execute each one
	// The build state is updated after each instruction
	// to reflect the current layer and instruction type
	// The build state is passed to each instruction handler
	// to allow them to modify the state as needed
	for _, instruction := range instructions {
		instructionType := instruction.GetType()

		// Check if the instruction type is valid
		if !isValidCommand(instructionType) {
			return fmt.Errorf("unknown command: %s", instructionType)
		}
		instructionArgs := instruction.GetArgs()

		// check if layer exists
		id := GenerateHexID(strings.Join([]string{instructionType, instructionArgs}, " "))
		if overlay.CheckIfLayerExists(id){
			config.Log.Infof("Layer is already in cache: %s", id)
			continue
		}

		// Execute the instruction using the appropriate handler
		config.Log.Infof("Executing instruction: %s %s", instructionType, instructionArgs)
		err := execute(instruction, buildState)
		if err != nil {
			return fmt.Errorf("failed to execute instruction %s: %w", instructionType, err)
		}
		config.Log.Infof("Instruction executed successfully: %s", instructionType)

		// Currently the build state is updated by the commands FIX IT so the handler returns the new overlay
		// and it gets updated here
		// Update the current layer in the build state
		// buildState.CurrentLayer = buildState.CurrentLayer
		// buildState.Instruction = instructionType
	}

	config.Log.Infof("Container build completed successfully.")
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

func GenerateHexID(input string) string {
	length := config.IDLength
	config.Log.Debugf("Generating hex ID for input: %s", input)
	hash := sha256.Sum256([]byte(input))
	hexString := hex.EncodeToString(hash[:])
	if length > len(hexString) {
		length = len(hexString)
	}
	return hexString[:length]
}
