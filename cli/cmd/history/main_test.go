package main

import (
	"strings"
	"testing"

	"module/lib/internal/history"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "needs truncation",
			input:    "hello world this is a long string",
			maxLen:   15,
			expected: "hello world ...",
		},
		{
			name:     "with newlines",
			input:    "hello\nworld\ntest",
			maxLen:   20,
			expected: "hello world test",
		},
		{
			name:     "with multiple spaces",
			input:    "hello    world     test",
			maxLen:   20,
			expected: "hello world test",
		},
		{
			name:     "newlines and truncation",
			input:    "hello\nworld\nthis is a very long string",
			maxLen:   15,
			expected: "hello world ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q",
					tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// @req CORE-HIST-002
func TestFormatEntryAsMarkdown(t *testing.T) {
	tests := []struct {
		name    string
		command string
		data    map[string]interface{}
		want    []string // strings that should appear in output
	}{
		{
			name:    "ask command",
			command: "ask",
			data: map[string]interface{}{
				"question": "What is Go?",
				"result":   "Go is a programming language",
			},
			want: []string{
				"# Ask History Entry",
				"## Question",
				"What is Go?",
				"## Answer",
				"Go is a programming language",
			},
		},
		{
			name:    "explain command",
			command: "explain",
			data: map[string]interface{}{
				"source":   "main.go",
				"language": "Go",
				"result":   "This is an explanation",
			},
			want: []string{
				"# Explain History Entry",
				"## Source",
				"`main.go`",
				"## Language",
				"Go",
				"## Explanation",
				"This is an explanation",
			},
		},
		{
			name:    "gitmessage command",
			command: "gitmessage",
			data: map[string]interface{}{
				"hint":   "fix bug",
				"prefix": "feat",
				"result": "Add new feature",
			},
			want: []string{
				"# Gitmessage History Entry",
				"## Hint",
				"fix bug",
				"## Prefix",
				"feat",
				"## Commit Message",
				"Add new feature",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &history.Entry{
				Command:   tt.command,
				Data:      tt.data,
				Date:      "2026-03-17 10:00:00",
				Timestamp: 1234567890,
				FilePath:  "/path/to/file.json",
			}

			result := formatEntryAsMarkdown(entry)

			for _, want := range tt.want {
				if !strings.Contains(result, want) {
					t.Errorf("formatEntryAsMarkdown() missing %q\nGot:\n%s",
						want, result)
				}
			}
		})
	}
}
