package build

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/lariskovski/containy/internal/config"
)

// BuildState maintains context during a container image build.
// It tracks the current layer and instruction being processed,
// allowing instructions to build upon previous ones.
type BuildState struct {
	// CurrentLayer represents the most recently created filesystem layer
	CurrentLayer Layer

	// CurrentInstructionType stores the type of the most recently executed instruction
	CurrentInstructionType string
}

// Build parses a container build file and executes its instructions to build an image.
// The file at 'filepath' should contain container build instructions (e.g., FROM, RUN).
// Each instruction is parsed, converted to the instructions.Instruction interface, and executed in order.
// If any instruction fails, the build process is aborted and an error is logged.
func Build(filepath, alias string) error {
	config.Log.Infof("Building container from file: %s", filepath)

	// Parse the build file into a slice of parser.Line instructions
	instructions, err := parse(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	buildState := &BuildState{}

	for step, instruction := range instructions {
		instructionType := instruction.GetType()
		instructionArgs := instruction.GetArgs()

		// Check if the instruction type is valid
		if !isValidCommand(instructionType) {
			return fmt.Errorf("unknown command: %s", instructionType)
		}

		config.Log.Infof("STEP %d: %s %s", step+1, instructionType, instructionArgs)

		id := GenerateHexID(strings.Join([]string{instructionType, instructionArgs}, " "))
		if checkIfLayerExists(id) {
			config.Log.Infof("Layer is cached: %s", id)
			// Load the cached layer and update build state
			cachedLayer, err := loadCachedLayer(id)
			if err != nil {
				return fmt.Errorf("failed to load cached layer %s: %w", id, err)
			}
			updateBuildState(buildState, cachedLayer, instructionType)
			continue
		}

		// Execute the instruction using the appropriate handler
		// Create a new layer for the instruction and returns it
		// in order to centralize build state updating
		layer, err := instruction.execute(buildState)
		if err != nil {
			return fmt.Errorf("failed to execute instruction %s: %w", instructionType, err)
		}
		config.Log.Debugf("Instruction executed successfully: %s", instructionType)

		// Update the build state with the new layer and instruction
		updateBuildState(buildState, layer, instructionType)
	}

	// Create an alias for the last layer if specified
	// This is useful for tagging the final image with a name
	// and version (e.g., "myimage:latest")
	// If no alias is provided, use the layer ID as the alias
	// This allows users to refer to the final image by a friendly name
	// instead of a hash
	if buildState.CurrentLayer != nil {
		finalAlias := alias
		if finalAlias == "" {
			finalAlias = buildState.CurrentLayer.GetID()
		}
		if err := buildState.CurrentLayer.CreateAlias(finalAlias); err != nil {
			return fmt.Errorf("failed to create alias for layer %s: %w", buildState.CurrentLayer.GetID(), err)
		}
		config.Log.Infof("Created alias %s -> %s", finalAlias, buildState.CurrentLayer.GetMergedDir())
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

func updateBuildState(state *BuildState, layer Layer, instructionType string) {
	state.CurrentLayer = layer
	state.CurrentInstructionType = instructionType
	config.Log.Debugf("Updated build state to current layer: %s", layer.GetID())
}
