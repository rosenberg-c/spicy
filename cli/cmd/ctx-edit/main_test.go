package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildPrompt(t *testing.T) {
	prompt := "rename foo to bar"
	contextInput := "const foo = 1"

	got := buildPrompt(prompt, contextInput)
	checks := []string{
		prompt,
		contextInput,
		"Only output the updated code",
		"Selected context:",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("buildPrompt missing %q", check)
		}
	}
}

func TestReadStdin(t *testing.T) {
	withStdin(t, "line 1\nline 2", func() {
		got, err := readStdin()
		if err != nil {
			t.Fatalf("readStdin() error = %v", err)
		}
		if got != "line 1\nline 2" {
			t.Fatalf("readStdin() = %q", got)
		}
	})
}

func TestReadStdinEmpty(t *testing.T) {
	withStdin(t, "   \n\n", func() {
		_, err := readStdin()
		if err == nil {
			t.Fatalf("expected error for empty stdin")
		}
	})
}

// @req CLI-CTX-001
func TestApplyUpdate_ReplacesSelection(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "sample.txt")
	input := "line 1\nline 2\nline 3\n"

	if err := os.WriteFile(filePath, []byte(input), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	sel := selection{filePath: filePath, start: 2, end: 2}
	updated, err := applyUpdate(sel, "updated")
	if err != nil {
		t.Fatalf("applyUpdate() error = %v", err)
	}
	if updated != filePath {
		t.Fatalf("applyUpdate() returned path %q", updated)
	}

	gotBytes, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}
	got := string(gotBytes)
	if got != "line 1\nupdated\nline 3\n" {
		t.Fatalf("updated file content = %q", got)
	}
}

// @req CLI-CTX-002
func TestApplyUpdate_InvalidRange(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "sample.txt")
	input := "line 1\nline 2\n"

	if err := os.WriteFile(filePath, []byte(input), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	sel := selection{filePath: filePath, start: 3, end: 4}
	_, err := applyUpdate(sel, "updated")
	if err == nil {
		t.Fatalf("expected error for invalid selection range")
	}
}

func withStdin(t *testing.T, input string, fn func()) {
	t.Helper()
	old := os.Stdin
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	if _, err := writer.WriteString(input); err != nil {
		t.Fatalf("failed to write to stdin pipe: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	os.Stdin = reader
	defer func() {
		os.Stdin = old
		reader.Close()
	}()

	fn()
}
