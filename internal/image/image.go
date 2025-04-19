package image

import (
	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/instructions"
	"github.com/lariskovski/containy/internal/parser"
)

func Build(filepath string) {
	config.Log.Debugf("Building container from file: %s", filepath)
	lines, err := parser.ParseFile(filepath)
	if err != nil {
		config.Log.Fatalf("Failed to parse file: %v", err)
	}
	// Convert lines to the expected type
	instructionsList := make([]instructions.Instruction, len(lines))
	for i, line := range lines {
		instructionsList[i] = instructions.Instruction(line)
	}

	// Execute the parsed instructions
	if err := instructions.Execute(instructionsList); err != nil {
		config.Log.Fatalf("Failed to execute instruction: %v", err)
	}

	config.Log.Infof("Container build completed successfully.")
}
