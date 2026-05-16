package askwrapper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadHistory_EmptyWhenMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	entries, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("LoadHistory() len = %d, want 0", len(entries))
	}
}

func TestAppendHistory_PrependsAndPersists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := AppendHistory("first question", "first answer"); err != nil {
		t.Fatalf("AppendHistory(first) error = %v", err)
	}
	if err := AppendHistory("second question", "second answer"); err != nil {
		t.Fatalf("AppendHistory(second) error = %v", err)
	}

	entries, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory() error = %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("LoadHistory() len = %d, want 2", len(entries))
	}

	if entries[0].Question != "second question" {
		t.Fatalf("entries[0].Question = %q, want %q", entries[0].Question, "second question")
	}
	if entries[0].Answer != "second answer" {
		t.Fatalf("entries[0].Answer = %q, want %q", entries[0].Answer, "second answer")
	}
	if entries[0].At == 0 {
		t.Fatalf("entries[0].At = 0, want non-zero")
	}

	if entries[1].Question != "first question" {
		t.Fatalf("entries[1].Question = %q, want %q", entries[1].Question, "first question")
	}

	path, err := HistoryPath()
	if err != nil {
		t.Fatalf("HistoryPath() error = %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("history file stat error = %v", err)
	}

	if filepath.Base(path) != historyFileName {
		t.Fatalf("history filename = %q, want %q", filepath.Base(path), historyFileName)
	}
}

func TestAppendHistory_EmptyQuestionNoWrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := AppendHistory("   ", "answer"); err != nil {
		t.Fatalf("AppendHistory() error = %v", err)
	}

	entries, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("LoadHistory() len = %d, want 0", len(entries))
	}
}
