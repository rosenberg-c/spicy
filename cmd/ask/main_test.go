package main

import (
	"strings"
	"testing"
)

func TestGetUserInput(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name:    "multiple args joined",
			args:    []string{"how", "does", "git", "work"},
			want:    "how does git work",
			wantErr: false,
		},
		{
			name:    "single arg",
			args:    []string{"hello"},
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "args with special characters",
			args:    []string{"what", "is", "Go's", "purpose?"},
			want:    "what is Go's purpose?",
			wantErr: false,
		},
		{
			name:    "args with spaces get joined",
			args:    []string{"explain", "context.Context", "in", "Go"},
			want:    "explain context.Context in Go",
			wantErr: false,
		},
		{
			name:    "single word question",
			args:    []string{"kubernetes"},
			want:    "kubernetes",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getUserInput(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserInput() error = %v, wantErr %v",
					err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getUserInput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantContains []string
	}{
		{
			name:  "includes user input",
			input: "explain closures",
			wantContains: []string{
				"explain closures",
				"senior coder",
				"User input:",
			},
		},
		{
			name:  "includes system prompt",
			input: "what is Go",
			wantContains: []string{
				"what is Go",
				"senior coder",
				"concise",
			},
		},
		{
			name:  "handles special characters",
			input: "what is <context>?",
			wantContains: []string{
				"what is <context>?",
			},
		},
		{
			name:  "handles long input",
			input: strings.Repeat("word ", 100),
			wantContains: []string{
				strings.Repeat("word ", 100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPrompt(tt.input)
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("buildPrompt() missing %q in result", want)
				}
			}

			// Verify prompt structure
			if got == "" {
				t.Error("buildPrompt() returned empty string")
			}

			// Verify input is at the end
			if !strings.HasSuffix(got, tt.input) {
				t.Error("buildPrompt() input not at end of prompt")
			}
		})
	}
}

func TestBuildPrompt_Format(t *testing.T) {
	input := "test question"
	got := buildPrompt(input)

	// Verify it's not just the input
	if got == input {
		t.Error("buildPrompt() returned only the input")
	}

	// Verify it contains the input
	if !strings.Contains(got, input) {
		t.Error("buildPrompt() does not contain input")
	}

	// Verify multiline format
	lines := strings.Split(got, "\n")
	if len(lines) < 3 {
		t.Errorf("buildPrompt() has %d lines, expected at least 3", len(lines))
	}
}

func TestGetUserInput_EmptyArgs(t *testing.T) {
	// Test with empty args would require stdin input
	// Skip in automated tests since it reads from stdin
	t.Skip("Skipping stdin test in automated testing")
}
