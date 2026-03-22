package main

import (
	"strings"
	"testing"

	"module/lib/internal/cliutil"
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
			args:    []string{"how", "to", "use", "docker"},
			want:    "how to use docker",
			wantErr: false,
		},
		{
			name:    "single word",
			args:    []string{"docker"},
			want:    "docker",
			wantErr: false,
		},
		{
			name:    "args with special characters",
			args:    []string{"how", "to", "use", "docker-compose"},
			want:    "how to use docker-compose",
			wantErr: false,
		},
		{
			name:    "long question",
			args:    []string{"how", "do", "I", "configure", "kubernetes", "ingress", "with", "nginx"},
			want:    "how do I configure kubernetes ingress with nginx",
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

func TestBuildTutorialPrompt(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantContains []string
	}{
		{
			name:  "includes user input",
			input: "how to use ffmpeg",
			wantContains: []string{
				"how to use ffmpeg",
				"senior coder",
				"tutorial",
				"markdown",
			},
		},
		{
			name:  "includes system instructions",
			input: "docker basics",
			wantContains: []string{
				"docker basics",
				"detailed",
				"tutorial",
			},
		},
		{
			name:  "handles special characters",
			input: "how to use <template> in Vue.js",
			wantContains: []string{
				"how to use <template> in Vue.js",
			},
		},
		{
			name:  "handles long input",
			input: strings.Repeat("word ", 50),
			wantContains: []string{
				strings.Repeat("word ", 50),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTutorialPrompt(tt.input)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("buildTutorialPrompt() missing %q", want)
				}
			}

			// Verify prompt structure
			if got == "" {
				t.Error("buildTutorialPrompt() returned empty string")
			}

			// Verify input is at the end
			if !strings.HasSuffix(got, tt.input) {
				t.Error("buildTutorialPrompt() input not at end")
			}
		})
	}
}

func TestBuildTutorialPrompt_Format(t *testing.T) {
	input := "test question"
	got := buildTutorialPrompt(input)

	// Verify it's not just the input
	if got == input {
		t.Error("buildTutorialPrompt() returned only the input")
	}

	// Verify it contains the input
	if !strings.Contains(got, input) {
		t.Error("buildTutorialPrompt() does not contain input")
	}

	// Verify multiline format
	lines := strings.Split(got, "\n")
	if len(lines) < 3 {
		t.Errorf("buildTutorialPrompt() has %d lines, want at least 3",
			len(lines))
	}

	// Verify requirements
	requirements := []string{
		"senior coder",
		"tutorial",
		"markdown",
	}

	for _, req := range requirements {
		if !strings.Contains(got, req) {
			t.Errorf("buildTutorialPrompt() missing %q", req)
		}
	}
}

func TestPromptOutputPath_Default(t *testing.T) {
	// This function reads from /dev/tty which is not available in tests
	// We can only test the fallback behavior
	t.Skip("Skipping /dev/tty test in automated testing")
	_, _ = cliutil.PromptOutputPath("tutorial.md")
}
