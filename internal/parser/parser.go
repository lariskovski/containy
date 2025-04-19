package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Line represents a generic parsed line
type Line struct {
	Type string
	Args string
}

// ParseFile reads a file line by line, parses each line into a Line struct, and returns a slice of parsed lines.
// It skips empty lines and lines starting with a '#' (comments).
func ParseFile(path string) ([]Line, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	var lines []Line
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
// It returns a Line struct containing the parsed type and arguments.
func parseLine(line string) (Line, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return Line{}, fmt.Errorf("empty or invalid line")
	}

	lineType := strings.ToUpper(fields[0])
	args := strings.Join(fields[1:], " ")

	return Line{Type: lineType, Args: args}, nil
}

func (l Line) GetType() string {
	return l.Type
}
func (l Line) GetArgs() string {
	return l.Args
}