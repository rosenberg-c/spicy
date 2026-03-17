package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Entry represents a history entry for a command
type Entry struct {
	Data      map[string]interface{} `json:"data"`
	Date      string                 `json:"date"`
	Version   int                    `json:"version"`
	Command   string                 `json:"command"`
	Timestamp int64                  `json:"timestamp"`
	FilePath  string                 `json:"-"` // Not saved to JSON
}

// Save writes a history entry to the appropriate directory
func Save(command string, data map[string]interface{}) error {
	now := time.Now()

	// Create the directory path: .spicy/<command>/
	historyDir := filepath.Join(".spicy", command)
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return fmt.Errorf("create history directory: %w", err)
	}

	// Create filename: [hour][minute][second].json
	filename := fmt.Sprintf("%02d%02d%02d.json", now.Hour(), now.Minute(), now.Second())
	filePath := filepath.Join(historyDir, filename)

	// Create the entry
	entry := Entry{
		Data:      data,
		Date:      now.Format("2006-01-02 15:04:05"),
		Version:   1,
		Command:   command,
		Timestamp: now.Unix(),
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("write history file: %w", err)
	}

	return nil
}

// Load reads a single history entry from a file
func Load(filePath string) (*Entry, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal JSON: %w", err)
	}

	entry.FilePath = filePath
	return &entry, nil
}

// List returns all history entries for a specific command,
// sorted by timestamp (newest first)
func List(command string) ([]Entry, error) {
	historyDir := filepath.Join(".spicy", command)

	// Check if directory exists
	if _, err := os.Stat(historyDir); os.IsNotExist(err) {
		return nil, nil // Return empty slice if no history exists
	}

	entries, err := os.ReadDir(historyDir)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	var results []Entry
	for i := range entries {
		if entries[i].IsDir() {
			continue
		}

		if filepath.Ext(entries[i].Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(historyDir, entries[i].Name())
		entry, err := Load(filePath)
		if err != nil {
			// Skip invalid files but continue
			continue
		}

		results = append(results, *entry)
	}

	// Sort by timestamp, newest first
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp > results[j].Timestamp
	})

	return results, nil
}

// ListAll returns all history entries across all commands,
// grouped by command name
func ListAll() (map[string][]Entry, error) {
	spicyDir := ".spicy"

	// Check if .spicy directory exists
	if _, err := os.Stat(spicyDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(spicyDir)
	if err != nil {
		return nil, fmt.Errorf("read .spicy directory: %w", err)
	}

	results := make(map[string][]Entry)
	for i := range entries {
		if !entries[i].IsDir() {
			continue
		}

		command := entries[i].Name()
		commandEntries, err := List(command)
		if err != nil {
			// Skip commands with errors
			continue
		}

		if len(commandEntries) > 0 {
			results[command] = commandEntries
		}
	}

	return results, nil
}
