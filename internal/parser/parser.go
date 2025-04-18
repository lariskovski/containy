package parser

import (
	"bufio"
	"os"
	"strings"
)

// Line represents a generic parsed line
type Line struct {
	Type string
	Args string
}

func ParseFile(path string) ([]Line, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
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

		// Split command and the rest
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		lineType := strings.ToUpper(fields[0])
		args := strings.Join(fields[1:], " ")

		lines = append(lines, Line{Type: lineType, Args: args})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
