package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name         string
		hint         string
		diff         string
		wantContains []string
	}{
		{
			name: "with hint",
			hint: "bugfix",
			diff: "some diff content",
			wantContains: []string{
				"bugfix",
				"some diff content",
				"commit message",
				"senior coder",
			},
		},
		{
			name: "without hint",
			hint: "",
			diff: "some diff content",
			wantContains: []string{
				"some diff content",
				"commit message",
			},
		},
		{
			name: "multiline diff",
			hint: "feature",
			diff: "diff --git a/file.go\n+added line\n-removed line",
			wantContains: []string{
				"feature",
				"diff --git",
				"added line",
			},
		},
		{
			name: "empty hint with diff",
			hint: "",
			diff: "diff content",
			wantContains: []string{
				"diff content",
				"commit message",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPrompt(tt.hint, tt.diff)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("buildPrompt() missing %q", want)
				}
			}

			// Verify structure
			if !strings.Contains(got, "Hint:") {
				t.Error("buildPrompt() missing 'Hint:' section")
			}
			if !strings.Contains(got, "Diff:") {
				t.Error("buildPrompt() missing 'Diff:' section")
			}
		})
	}
}

func TestBuildPrompt_Format(t *testing.T) {
	hint := "test hint"
	diff := "test diff"
	got := buildPrompt(hint, diff)

	// Verify multiline format
	lines := strings.Split(got, "\n")
	if len(lines) < 5 {
		t.Errorf("buildPrompt() has %d lines, expected at least 5",
			len(lines))
	}

	// Verify requirements are present
	requirements := []string{
		"short commit message",
		"one row only",
		"Capital character",
	}

	for _, req := range requirements {
		if !strings.Contains(got, req) {
			t.Errorf("buildPrompt() missing requirement %q", req)
		}
	}
}

func TestGetStagedDiff_NotGitRepo(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	ctx := context.Background()
	_, err = getStagedDiff(ctx)
	if err == nil {
		t.Error("expected error for non-git directory, got nil")
	}
}

func TestGetStagedDiff_EmptyRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping git integration test in short mode")
	}

	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Initialize git repo
	initGitRepo(t, tempDir)

	ctx := context.Background()
	diff, err := getStagedDiff(ctx)
	if err != nil {
		t.Fatalf("getStagedDiff() error = %v", err)
	}

	// Empty repo should have no staged changes
	if strings.TrimSpace(diff) != "" {
		t.Errorf("expected empty diff, got %q", diff)
	}
}

func TestGetStagedDiff_WithStagedChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping git integration test in short mode")
	}

	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Initialize git repo
	initGitRepo(t, tempDir)

	// Create and stage a file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cmd := exec.Command("git", "add", "test.txt")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add failed: %v", err)
	}

	ctx := context.Background()
	diff, err := getStagedDiff(ctx)
	if err != nil {
		t.Fatalf("getStagedDiff() error = %v", err)
	}

	// Should have diff content
	if strings.TrimSpace(diff) == "" {
		t.Error("expected non-empty diff")
	}

	// Should contain file name
	if !strings.Contains(diff, "test.txt") {
		t.Errorf("diff missing file name: %s", diff)
	}
}

// Helper function to initialize a git repo for testing
func initGitRepo(t *testing.T, dir string) {
	t.Helper()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Logf("git config user.email failed: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Logf("git config user.name failed: %v", err)
	}
}
