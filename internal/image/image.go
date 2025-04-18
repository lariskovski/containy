package image

import (
	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/instructions"
	"github.com/lariskovski/containy/internal/parser"
)

func Build(filepath string) {
	config.Log.Debugf("Building container from file: %s", filepath)
	fileInstructions, err := parser.ParseFile(filepath)
	if err != nil {
		config.Log.Fatalf("Failed to parse file: %v", err)
	}
	// Execute the parsed instructions
	err = instructions.ExecuteInstructions(fileInstructions)
	if err != nil {
		config.Log.Fatalf("Failed to execute instructions: %v", err)
	}
}
