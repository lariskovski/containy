package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Instruction struct {
	Command string
	Args    string
}

func ParseFile(path string) ([]Instruction, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var instructions []Instruction
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split command and the rest
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		command := strings.ToUpper(fields[0])
		args := strings.Join(fields[1:], " ")

		// Optional: only accept known Dockerfile commands
		switch command {
		case "FROM", "RUN", "COPY", "CMD":
			instructions = append(instructions, Instruction{Command: command, Args: args})
		default:
			fmt.Printf("Unknown command: %s\n", command)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return instructions, nil
}