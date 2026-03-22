package cliutil

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ShouldSave returns true if output should be written to a file.
func ShouldSave(output string, saveFlag bool) bool {
	return saveFlag || output != ""
}

// PromptOutputPath prompts the user for an output path, using a suggested default.
func PromptOutputPath(suggestedFilename string) (string, error) {
	fmt.Printf("Save to file (default: %s) => ", suggestedFilename)

	// Open /dev/tty to read from terminal instead of stdin
	// This allows reading user input even when stdin is piped
	tty, err := os.Open("/dev/tty")
	if err != nil {
		// If /dev/tty is not available (non-interactive), use default
		fmt.Fprintln(os.Stderr, "\nNo terminal available, using default filename")
		return suggestedFilename, nil
	}
	defer tty.Close()

	reader := bufio.NewReader(tty)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return suggestedFilename, nil
	}

	return input, nil
}
