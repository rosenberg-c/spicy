package filename

import (
	"regexp"
	"strings"
	"testing"
)

func TestSanitize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple string",
			input: "hello world",
			want:  "hello-world",
		},
		{
			name:  "with extension",
			input: "main.go",
			want:  "main",
		},
		{
			name:  "special characters",
			input: "hello@world!test",
			want:  "helloworldtest",
		},
		{
			name:  "multiple spaces",
			input: "hello   world   test",
			want:  "hello-world-test",
		},
		{
			name:  "uppercase",
			input: "HelloWorld",
			want:  "helloworld",
		},
		{
			name:  "leading and trailing hyphens",
			input: "-hello-world-",
			want:  "hello-world",
		},
		{
			name:  "consecutive hyphens",
			input: "hello---world",
			want:  "hello-world",
		},
		{
			name:  "empty string",
			input: "",
			want:  "output",
		},
		{
			name:  "only special characters",
			input: "@#$%^&*()",
			want:  "output",
		},
		{
			name:  "underscores preserved",
			input: "hello_world_test",
			want:  "hello_world_test",
		},
		{
			name:  "numbers preserved",
			input: "test123file",
			want:  "test123file",
		},
		{
			name:  "mixed case with spaces",
			input: "What is Rust",
			want:  "what-is-rust",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sanitize(tt.input)
			if got != tt.want {
				t.Errorf("Sanitize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateTimestamped(t *testing.T) {
	tests := []struct {
		name  string
		cmd   string
		input string
	}{
		{
			name:  "simple input",
			cmd:   "ask",
			input: "what is rust",
		},
		{
			name:  "long input",
			cmd:   "explain",
			input: "this is a very long filename that should be truncated to forty characters maximum",
		},
		{
			name:  "input with special chars",
			cmd:   "ask",
			input: "what is @rust and #golang?",
		},
		{
			name:  "filename input",
			cmd:   "explain",
			input: "main.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateTimestamped(tt.cmd, tt.input)

			// Check format: YYYY-MM-DD_HH-MM_cmd_suggestion.md
			pattern := `^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}_` + tt.cmd + `_[a-z0-9\-_]+\.md$`
			matched, err := regexp.MatchString(pattern, got)
			if err != nil {
				t.Fatalf("regex error: %v", err)
			}
			if !matched {
				t.Errorf("GenerateTimestamped(%q, %q) = %q, doesn't match pattern %q", tt.cmd, tt.input, got, pattern)
			}

			// Check it ends with .md
			if !strings.HasSuffix(got, ".md") {
				t.Errorf("GenerateTimestamped(%q, %q) = %q, doesn't end with .md", tt.cmd, tt.input, got)
			}

			// Check suggestion is not too long (max 40 chars for suggestion part)
			parts := strings.Split(got, "_")
			if len(parts) < 4 {
				t.Errorf("GenerateTimestamped(%q, %q) = %q, invalid format", tt.cmd, tt.input, got)
				return
			}

			// Last part is "suggestion.md", extract suggestion
			suggestion := strings.TrimSuffix(parts[len(parts)-1], ".md")
			if len(suggestion) > 40 {
				t.Errorf("GenerateTimestamped(%q, %q) suggestion %q is too long (%d chars)", tt.cmd, tt.input, suggestion, len(suggestion))
			}
		})
	}
}

func TestGenerateTimestampedTruncation(t *testing.T) {
	longInput := "this is a very long input that will definitely exceed forty characters and should be truncated"
	result := GenerateTimestamped("test", longInput)

	// Extract the suggestion part
	parts := strings.Split(result, "_")
	suggestion := strings.TrimSuffix(parts[len(parts)-1], ".md")

	if len(suggestion) > 40 {
		t.Errorf("Expected suggestion to be truncated to 40 chars, got %d: %q", len(suggestion), suggestion)
	}
}
