package instructions

import (
	"fmt"

	"github.com/lariskovski/containy/internal/config"
)

// Instruction represents a single instruction in the Dockerfile.
// It is an interface that defines the methods to get the type and arguments of the instruction.
// The type of instruction can be "FROM", "RUN", "COPY", etc.
type Instruction interface {
	GetType() string
	GetArgs() string
}

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

func Execute(instructions []Instruction) error {
	config.Log.Info("Executing instructions")

	buildState := &BuildState{}

	for _, instruction := range instructions {
		instructionType := instruction.GetType()
		if !isValidCommand(instructionType) {
			return fmt.Errorf("unknown command: %s", instructionType)
		}
		instructionArgs := instruction.GetArgs()

		handler := handlers[instructionType]
		if err := handler(instructionArgs, buildState); err != nil {
			config.Log.Errorf("%s failed: %v", instructionType, err)
			return fmt.Errorf("%s failed: %w", instructionType, err)
		}
	}

	return nil
}

func isValidCommand(cmd string) bool {
	_, ok := handlers[cmd]
	return ok
}
