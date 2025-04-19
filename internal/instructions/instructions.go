package instructions

import (
	"fmt"

	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/parser"
)

// BuildState holds the current state of the build process.
type BuildState struct {
	CurrentLayer Layer
	Instruction  string
}

var handlers = map[string]func(string, *BuildState) error{
	"FROM": from,
	"RUN":  runCmd,
	// "COPY": copyCmd,
	// "CMD":  cmd,
}

func ExecuteInstructions(lines []parser.Line) error {
	config.Log.Info("Executing instructions")
	instructions, err := validateAndConvertLines(lines)
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

func validateAndConvertLines(lines []parser.Line) ([]parser.Line, error) {
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
