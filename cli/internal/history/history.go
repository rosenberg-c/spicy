package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"module/lib/internal/filename"
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

// Save writes a history entry to the appropriate directory.
// Format: [YearMonthDay]-[hourMinuteSec]_[cmd][_suggestion].json
// suggestedFilename is optional and will be sanitized if provided.
func Save(command string, data map[string]interface{}, suggestedFilename string) error {
	now := time.Now()

	// Create the directory path: .spicy/<command>/
	historyDir := filepath.Join(".spicy", command)
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return fmt.Errorf("create history directory: %w", err)
	}

	// Create filename: YYYYMMDD-HHMMSS_cmd[_suggestion].json
	dateTime := now.Format("20060102-150405")
	var fname string
	if suggestedFilename != "" {
		sanitized := filename.Sanitize(suggestedFilename)
		// Truncate if too long
		if len(sanitized) > 40 {
			sanitized = sanitized[:40]
		}
		fname = fmt.Sprintf("%s_%s_%s.json", dateTime, command, sanitized)
	} else {
		fname = fmt.Sprintf("%s_%s.json", dateTime, command)
	}
	filePath := filepath.Join(historyDir, fname)

	// Create the entry
	entry := Entry{
		Data:      normalizeData(data),
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

func normalizeData(data map[string]interface{}) map[string]interface{} {
	wrapped := make(map[string]interface{}, len(data))
	for k, v := range data {
		wrapped[k] = normalizeValue(v)
	}
	return wrapped
}

func normalizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return wrapText(v, 80)
	case []interface{}:
		out := make([]interface{}, len(v))
		for i, item := range v {
			out[i] = normalizeValue(item)
		}
		return out
	case map[string]interface{}:
		return normalizeData(v)
	default:
		return value
	}
}

func wrapText(input string, width int) string {
	if width <= 0 || input == "" {
		return input
	}

	lines := strings.Split(input, "\n")
	for i, line := range lines {
		lines[i] = wrapLine(line, width)
	}

	return strings.Join(lines, "\n")
}

func wrapLine(line string, width int) string {
	if len(line) <= width {
		return line
	}

	var wrapped []string
	remaining := line
	for len(remaining) > width {
		breakAt := strings.LastIndex(remaining[:width+1], " ")
		if breakAt <= 0 {
			breakAt = width
		}
		segment := strings.TrimRight(remaining[:breakAt], " ")
		wrapped = append(wrapped, segment)
		remaining = strings.TrimLeft(remaining[breakAt:], " ")
	}
	if remaining != "" {
		wrapped = append(wrapped, remaining)
	}

	return strings.Join(wrapped, "\n")
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
