package build

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Parse reads a file line by line, parses each line into a Instruction struct, and returns a slice of parsed lines.
// It skips empty lines and lines starting with a '#' (comments).
func parse(path string) ([]Instruction, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	var lines []Instruction
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse the command and arguments
		parsedLine, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line: %w", err)
		}

		lines = append(lines, parsedLine)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", path, err)
	}

	return lines, nil
}

// parseLine splits a line into its type (command) and arguments.
// It returns a Instruction struct containing the parsed type and arguments.
func parseLine(line string) (Instruction, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return Instruction{}, fmt.Errorf("empty line")
	}

	lineType := strings.ToUpper(fields[0])
	args := strings.Join(fields[1:], " ")

	return Instruction{Type: lineType, Args: args}, nil
}

func (i Instruction) GetType() string {
	return i.Type
}
func (i Instruction) GetArgs() string {
	return i.Args
}
