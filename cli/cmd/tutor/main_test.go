package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
	"module/lib/internal/cliutil"
)

// @req CORE-CLI-010
func TestTutorValidationAlwaysArchived(t *testing.T) {
	for _, session := range []bool{false, true} {
		t.Run(map[bool]string{false: "default", true: "session"}[session], func(t *testing.T) {
			dir := t.TempDir()
			logPath := filepath.Join(dir, "calls")
			script := `#!/bin/sh
if [ "$2" = "--archived" ]; then printf 'archived\n' >> "$SPICY_CALLS"; else printf 'visible\n' >> "$SPICY_CALLS"; fi
case "$*" in
  *"Respond ONLY with a valid JSON"*) printf '%s\n' '{"action":"continue","reason":"specific","suggested_filename":"guide.md"}' ;;
  *) printf '%s\n' 'Tutorial content' ;;
esac
`
			if err := os.WriteFile(filepath.Join(dir, "opencode"), []byte(script), 0o755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("SPICY_CALLS", logPath)
			command := &cli.Command{
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "model", Value: "custom/test"},
					&cli.BoolFlag{Name: "session"},
				},
				Action: run,
			}
			args := []string{"tutor"}
			if session {
				args = append(args, "--session")
			}
			args = append(args, "how to read CSV files")
			if err := command.Run(context.Background(), args); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			want := "archived\narchived\n"
			if session {
				want = "archived\nvisible\n"
			}
			if string(got) != want {
				t.Errorf("OpenCode calls = %q, want %q", got, want)
			}
		})
	}
}

// @req CLI-TUTOR-001
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

// @req CLI-TUTOR-001
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
