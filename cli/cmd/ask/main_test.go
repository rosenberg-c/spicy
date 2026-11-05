package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// @req CLI-ASK-001, CORE-CLI-001
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

// @req CLI-ASK-002
func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		context      string
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
		{
			name:    "includes context when provided",
			input:   "explain this",
			context: "func main() { fmt.Println(\"hi\") }",
			wantContains: []string{
				"explain this",
				"func main() { fmt.Println(\"hi\") }",
				"Additional context for reference:",
				"User input:",
			},
		},
		{
			name:    "blank-only context omitted",
			input:   "explain this",
			context: "   \n\t  ",
			wantContains: []string{
				"explain this",
				"User input:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPrompt(tt.input, tt.context)
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

// @req CLI-ASK-002
func TestBuildPrompt_EmptyContextMatchesLegacy(t *testing.T) {
	input := "test question"
	got := buildPrompt(input, "")
	want := "You are a senior coder.\n" +
		"Answer the user question in a short concise manner.\n" +
		"\n" +
		"The response must be valid markdown.\n" +
		"\n" +
		"User input:\n" +
		input
	if got != want {
		t.Errorf("buildPrompt() with empty context = %q, want legacy %q", got, want)
	}
}

// @req CLI-ASK-002
func TestBuildPrompt_ContextPlacement(t *testing.T) {
	got := buildPrompt("what is this?", "line1\nline2")
	// Context must appear before the user input section.
	idxCtx := strings.Index(got, "line1\nline2")
	idxInput := strings.Index(got, "User input:")
	if idxCtx == -1 || idxInput == -1 {
		t.Fatalf("buildPrompt() missing context or input section")
	}
	if idxCtx > idxInput {
		t.Error("buildPrompt() context appears after user input")
	}
}

func TestBuildPrompt_Format(t *testing.T) {
	input := "test question"
	got := buildPrompt(input, "")

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

// @req CLI-ASK-004
func TestBuildContextInput(t *testing.T) {
	tests := []struct {
		name         string
		contextInput string
		contextFile  string
		want         string
		wantErr      bool
	}{
		{
			name:         "both flags error",
			contextInput: "inline",
			contextFile:  "file.txt",
			wantErr:      true,
		},
		{
			name:         "inline context returned",
			contextInput: "inline context",
			contextFile:  "",
			want:         "inline context",
		},
		{
			name: "no context returns empty",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getContextInput(tt.contextInput, tt.contextFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("getContextInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("getContextInput() = %q, want %q", got, tt.want)
			}
		})
	}
}

// @req CLI-ASK-004
func TestBuildContextInput_ReadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ctx.txt")
	content := "package main\n\nfunc main() {}\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write context file: %v", err)
	}

	got, err := getContextInput("", path)
	if err != nil {
		t.Fatalf("getContextInput() error = %v", err)
	}
	if got != content {
		t.Errorf("getContextInput() = %q, want %q", got, content)
	}
}

// @req CLI-ASK-004
func TestBuildContextInput_MissingFile(t *testing.T) {
	_, err := getContextInput("", filepath.Join(t.TempDir(), "does-not-exist.txt"))
	if err == nil {
		t.Error("getContextInput() expected error for missing file")
	}
}

func TestGetUserInput_EmptyArgs(t *testing.T) {
	// Test with empty args would require stdin input
	// Skip in automated tests since it reads from stdin
	t.Skip("Skipping stdin test in automated testing")
}
