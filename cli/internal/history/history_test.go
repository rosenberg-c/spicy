package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// @req CORE-CLI-003
func TestSave(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Change to the temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Test saving a history entry
	data := map[string]interface{}{
		"question": "what is Go?",
		"result":   "Go is a programming language",
	}

	err = Save("ask", data, "what-is-go")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Check that the directory was created
	historyDir := filepath.Join(".spicy", "ask")
	if _, err := os.Stat(historyDir); os.IsNotExist(err) {
		t.Fatal("History directory was not created")
	}

	// Check that a file was created
	entries, err := os.ReadDir(historyDir)
	if err != nil {
		t.Fatalf("Failed to read history directory: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(entries))
	}

	// Read the file and verify its contents
	filePath := filepath.Join(historyDir, entries[0].Name())
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}

	var entry Entry
	if err := json.Unmarshal(fileData, &entry); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify the entry
	if entry.Command != "ask" {
		t.Errorf("Expected command 'ask', got '%s'", entry.Command)
	}

	if entry.Version != 1 {
		t.Errorf("Expected version 1, got %d", entry.Version)
	}

	if entry.Data["question"] != "what is Go?" {
		t.Errorf("Expected question 'what is Go?', got '%v'", entry.Data["question"])
	}

	if entry.Data["result"] != "Go is a programming language" {
		t.Errorf("Expected result 'Go is a programming language', got '%v'", entry.Data["result"])
	}

	if entry.Date == "" {
		t.Error("Expected non-empty date")
	}

	if entry.Timestamp == 0 {
		t.Error("Expected non-zero timestamp")
	}
}

func TestLoad(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Change to the temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Save a history entry first
	data := map[string]interface{}{
		"question": "test question",
		"result":   "test result",
	}

	err = Save("ask", data, "test-question")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Find the created file
	historyDir := filepath.Join(".spicy", "ask")
	entries, err := os.ReadDir(historyDir)
	if err != nil {
		t.Fatalf("Failed to read history directory: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("No history files found")
	}

	// Load the entry
	filePath := filepath.Join(historyDir, entries[0].Name())
	entry, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify the loaded entry
	if entry.Command != "ask" {
		t.Errorf("Expected command 'ask', got '%s'", entry.Command)
	}

	if entry.Data["question"] != "test question" {
		t.Errorf("Expected question 'test question', got '%v'",
			entry.Data["question"])
	}

	if entry.FilePath != filePath {
		t.Errorf("Expected FilePath '%s', got '%s'", filePath, entry.FilePath)
	}
}

func TestList(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Change to the temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Test listing when no history exists
	entries, err := List("ask")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if entries != nil && len(entries) != 0 {
		t.Errorf("Expected empty list, got %d entries", len(entries))
	}

	// Save history entry with new format
	data1 := map[string]interface{}{
		"question": "question 1",
		"result":   "result 1",
	}
	err = Save("ask", data1, "question-1")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// List entries
	entries, err = List("ask")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	// Verify entry data
	if entries[0].Command != "ask" {
		t.Errorf("Expected command 'ask', got '%s'", entries[0].Command)
	}
}

// @req CORE-HIST-001
func TestListReturnsNewestFirst(t *testing.T) {
	tmpDir := t.TempDir()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	historyDir := filepath.Join(".spicy", "ask")
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatalf("create history dir: %v", err)
	}

	older := Entry{
		Data:      map[string]interface{}{"question": "older"},
		Date:      "2026-01-01 10:00:00",
		Version:   1,
		Command:   "ask",
		Timestamp: 100,
	}
	newer := Entry{
		Data:      map[string]interface{}{"question": "newer"},
		Date:      "2026-01-01 10:00:01",
		Version:   1,
		Command:   "ask",
		Timestamp: 200,
	}

	olderJSON, err := json.Marshal(older)
	if err != nil {
		t.Fatalf("marshal older entry: %v", err)
	}
	newerJSON, err := json.Marshal(newer)
	if err != nil {
		t.Fatalf("marshal newer entry: %v", err)
	}

	if err := os.WriteFile(filepath.Join(historyDir, "older.json"), olderJSON, 0o644); err != nil {
		t.Fatalf("write older entry: %v", err)
	}
	if err := os.WriteFile(filepath.Join(historyDir, "newer.json"), newerJSON, 0o644); err != nil {
		t.Fatalf("write newer entry: %v", err)
	}

	entries, err := List("ask")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	if entries[0].Timestamp != 200 {
		t.Fatalf("Expected newest entry first, got timestamp %d", entries[0].Timestamp)
	}
	if entries[1].Timestamp != 100 {
		t.Fatalf("Expected older entry second, got timestamp %d", entries[1].Timestamp)
	}
}

func TestListAll(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Change to the temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Test listing when no history exists
	allEntries, err := ListAll()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}

	if allEntries != nil && len(allEntries) != 0 {
		t.Errorf("Expected empty map, got %d commands", len(allEntries))
	}

	// Save history for multiple commands
	commands := []string{"ask", "explain", "tutor"}
	for _, cmd := range commands {
		data := map[string]interface{}{
			"test": fmt.Sprintf("data for %s", cmd),
		}

		err = Save(cmd, data, fmt.Sprintf("test-%s", cmd))
		if err != nil {
			t.Fatalf("Save failed for %s: %v", cmd, err)
		}
	}

	// List all entries
	allEntries, err = ListAll()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}

	if len(allEntries) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(allEntries))
	}

	// Verify all commands are present
	for _, cmd := range commands {
		if _, ok := allEntries[cmd]; !ok {
			t.Errorf("Expected command '%s' in results", cmd)
		}
	}
}
