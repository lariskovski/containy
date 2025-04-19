package image

import (
	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/instructions"
	"github.com/lariskovski/containy/internal/parser"
)

// Build parses a container build file and executes its instructions to build an image.
// The file at 'filepath' should contain container build instructions (e.g., FROM, RUN).
// Each instruction is parsed, converted to the instructions.Instruction interface, and executed in order.
// If any instruction fails, the build process is aborted and an error is logged.
func Build(filepath string) {
	config.Log.Debugf("Building container from file: %s", filepath)
	// Parse the build file into a slice of parser.Line instructions
	lines, err := parser.Parse(filepath)
	if err != nil {
		config.Log.Fatalf("Failed to parse file: %v", err)
	}
	// Convert parser.Line to instructions.Instruction interface
	instructionsList := make([]instructions.Instruction, len(lines))
	for i, line := range lines {
		instructionsList[i] = instructions.Instruction(line)
	}

	// Execute all parsed instructions in order
	if err := instructions.Execute(instructionsList); err != nil {
		config.Log.Fatalf("Failed to execute instruction: %v", err)
	}

	config.Log.Infof("Container build completed successfully.")
}
