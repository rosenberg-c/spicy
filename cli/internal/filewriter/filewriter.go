package filewriter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteAtomic writes content to a file atomically using a temp file and rename.
// Expands ~ to home directory and creates parent directories if needed.
// Ensures content ends with newline. Returns the absolute path of the written file.
func WriteAtomic(path, content string) (string, error) {
	// Expand ~ to home directory
	expandedPath, err := expandPath(path)
	if err != nil {
		return "", fmt.Errorf("expand path: %w", err)
	}

	// Get absolute path
	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		return "", fmt.Errorf("get absolute path: %w", err)
	}

	// Create parent directories if needed
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create parent directories: %w", err)
	}

	// Ensure content ends with newline
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	// Create temporary file in same directory as target
	tmpFile, err := os.CreateTemp(dir, fmt.Sprintf(".%s.*.tmp", filepath.Base(absPath)))
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Write content to temp file
	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath) // Clean up
		return "", fmt.Errorf("write to temp file: %w", err)
	}

	// Close temp file
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath) // Clean up
		return "", fmt.Errorf("close temp file: %w", err)
	}

	// Atomically rename temp file to target
	if err := os.Rename(tmpPath, absPath); err != nil {
		os.Remove(tmpPath) // Clean up
		return "", fmt.Errorf("rename temp file: %w", err)
	}

	return absPath, nil
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	if path == "~" {
		return home, nil
	}

	// ~/something -> $HOME/something
	return filepath.Join(home, path[2:]), nil
}
