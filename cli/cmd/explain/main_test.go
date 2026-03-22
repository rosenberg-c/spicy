package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetCodeInput_File(t *testing.T) {
	tests := []struct {
		name           string
		fileContent    string
		fileName       string
		wantSourceName string
		wantErr        bool
	}{
		{
			name:           "read go file",
			fileContent:    "package main\n\nfunc main() {}",
			fileName:       "main.go",
			wantSourceName: "main.go",
			wantErr:        false,
		},
		{
			name:           "read python file",
			fileContent:    "def hello():\n    print('hello')",
			fileName:       "script.py",
			wantSourceName: "script.py",
			wantErr:        false,
		},
		{
			name:           "empty file",
			fileContent:    "",
			fileName:       "empty.go",
			wantSourceName: "empty.go",
			wantErr:        false,
		},
		{
			name:           "file with unicode",
			fileContent:    "// Hello 世界",
			fileName:       "unicode.go",
			wantSourceName: "unicode.go",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, tt.fileName)
			err := os.WriteFile(filePath, []byte(tt.fileContent), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			code, sourceName, err := getCodeInput(filePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("getCodeInput() error = %v, wantErr %v",
					err, tt.wantErr)
				return
			}

			if sourceName != tt.wantSourceName {
				t.Errorf("sourceName = %q, want %q",
					sourceName, tt.wantSourceName)
			}

			if code != tt.fileContent {
				t.Errorf("code = %q, want %q", code, tt.fileContent)
			}
		})
	}
}

func TestReadDirectory(t *testing.T) {
	tests := []struct {
		name       string
		setupFiles map[string]string
		wantFiles  []string
		skipFiles  []string
	}{
		{
			name: "includes go files",
			setupFiles: map[string]string{
				"main.go":          "package main",
				"helper.go":        "package main",
				"internal/util.go": "package internal",
			},
			wantFiles: []string{"main.go", "helper.go", "internal/util.go"},
		},
		{
			name: "skips hidden and vendor",
			setupFiles: map[string]string{
				"main.go":             "package main",
				".hidden.go":          "package main",
				"vendor/lib.go":       "package vendor",
				"node_modules/dep.js": "module.exports",
			},
			wantFiles: []string{"main.go"},
			skipFiles: []string{".hidden", "vendor", "node_modules"},
		},
		{
			name: "includes multiple file types",
			setupFiles: map[string]string{
				"main.go":   "package main",
				"script.py": "print('hello')",
				"app.js":    "console.log('hi')",
				"README.md": "# Readme",
			},
			wantFiles: []string{"main.go", "script.py", "app.js", "README.md"},
		},
		{
			name: "skips bin directory",
			setupFiles: map[string]string{
				"main.go":      "package main",
				"bin/app":      "binary",
				"bin/other.go": "package main",
			},
			wantFiles: []string{"main.go"},
			skipFiles: []string{"bin/"},
		},
		{
			name:       "empty directory",
			setupFiles: map[string]string{},
			wantFiles:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create test files
			for path, content := range tt.setupFiles {
				fullPath := filepath.Join(tempDir, path)
				err := os.MkdirAll(filepath.Dir(fullPath), 0755)
				if err != nil {
					t.Fatalf("failed to create dir: %v", err)
				}
				err = os.WriteFile(fullPath, []byte(content), 0644)
				if err != nil {
					t.Fatalf("failed to write file: %v", err)
				}
			}

			got, err := readDirectory(tempDir)
			if err != nil {
				t.Fatalf("readDirectory() error = %v", err)
			}

			// Verify included files
			for _, wantFile := range tt.wantFiles {
				if !strings.Contains(got, wantFile) {
					t.Errorf("output missing expected file %q", wantFile)
				}
			}

			// Verify skipped files
			for _, skipFile := range tt.skipFiles {
				if strings.Contains(got, skipFile) {
					t.Errorf("output should not contain %q", skipFile)
				}
			}

			// Verify file markers present
			if len(tt.wantFiles) > 0 {
				for _, wantFile := range tt.wantFiles {
					marker := "// File: " + wantFile
					if !strings.Contains(got, marker) {
						t.Errorf("output missing file marker %q", marker)
					}
				}
			}
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name   string
		source string
		code   string
		want   string
	}{
		{
			name:   "from go extension",
			source: "main.go",
			code:   "",
			want:   "Go",
		},
		{
			name:   "from python extension",
			source: "script.py",
			code:   "",
			want:   "Python",
		},
		{
			name:   "from javascript extension",
			source: "app.js",
			code:   "",
			want:   "JavaScript",
		},
		{
			name:   "from typescript extension",
			source: "component.ts",
			code:   "",
			want:   "TypeScript",
		},
		{
			name:   "from go content",
			source: "-",
			code:   "package main\n\nfunc test() {}",
			want:   "Go",
		},
		{
			name:   "from python content",
			source: "-",
			code:   "def hello():\n    pass",
			want:   "Python",
		},
		{
			name:   "from javascript content",
			source: "-",
			code:   "function test() { const x = 1; }",
			want:   "JavaScript",
		},
		{
			name:   "fallback to code",
			source: "-",
			code:   "some random text",
			want:   "code",
		},
		{
			name:   "stdin source",
			source: "stdin",
			code:   "package main",
			want:   "Go",
		},
		{
			name:   "rust extension",
			source: "main.rs",
			code:   "",
			want:   "Rust",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectLanguage(tt.source, tt.code)
			if got != tt.want {
				t.Errorf("detectLanguage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSuggestFilename(t *testing.T) {
	tests := []struct {
		sourceName string
		language   string
		want       string
	}{
		{
			sourceName: "main.go",
			language:   "Go",
			want:       "main-explanation.md",
		},
		{
			sourceName: "stdin",
			language:   "Python",
			want:       "python-explanation.md",
		},
		{
			sourceName: "complex.rs",
			language:   "Rust",
			want:       "complex-explanation.md",
		},
		{
			sourceName: "",
			language:   "JavaScript",
			want:       "javascript-explanation.md",
		},
		{
			sourceName: "file.with.dots.go",
			language:   "Go",
			want:       "file.with.dots-explanation.md",
		},
		{
			sourceName: "agent",
			language:   "Go",
			want:       "agent-explanation.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.sourceName, func(t *testing.T) {
			got := suggestFilename(tt.sourceName, tt.language)
			if got != tt.want {
				t.Errorf("suggestFilename() = %q, want %q", got, tt.want)
			}

			// Verify it ends with .md
			if !strings.HasSuffix(got, ".md") {
				t.Errorf("filename %q does not end with .md", got)
			}

			// Verify it contains "explanation"
			if !strings.Contains(got, "explanation") {
				t.Errorf("filename %q does not contain 'explanation'", got)
			}
		})
	}
}

func TestAddLineNumbers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single line",
			input: "line 1",
			want:  []string{"```", "   1  line 1", "```"},
		},
		{
			name:  "multiple lines",
			input: "line 1\nline 2\nline 3",
			want:  []string{"   1  line 1", "   2  line 2", "   3  line 3"},
		},
		{
			name:  "empty input",
			input: "",
			want:  []string{"```", "   1  ", "```"},
		},
		{
			name:  "line with tabs",
			input: "func\tmain()",
			want:  []string{"```", "   1  func\tmain()", "```"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addLineNumbers(tt.input)

			// Verify code fences
			if !strings.HasPrefix(got, "```") {
				t.Error("output missing opening code fence")
			}
			if !strings.HasSuffix(got, "```\n") {
				t.Error("output missing closing code fence")
			}

			// Verify expected content
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q", want)
				}
			}

			// Count lines (excluding fences)
			lines := strings.Split(got, "\n")
			inputLines := strings.Split(tt.input, "\n")
			// +2 for opening/closing fences, +1 for final newline
			expectedLines := len(inputLines) + 2 + 1
			if len(lines) != expectedLines {
				t.Errorf("got %d lines, want %d", len(lines), expectedLines)
			}
		})
	}
}

func TestBuildExplanationPrompt(t *testing.T) {
	tests := []struct {
		name         string
		code         string
		language     string
		wantContains []string
	}{
		{
			name:     "go code",
			code:     "package main",
			language: "Go",
			wantContains: []string{
				"Go",
				"package main",
				"senior software engineer",
				"markdown",
			},
		},
		{
			name:     "python code",
			code:     "def hello():\n    pass",
			language: "Python",
			wantContains: []string{
				"Python",
				"def hello():",
				"markdown",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildExplanationPrompt(tt.code, tt.language)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("prompt missing %q", want)
				}
			}

			// Verify code with line numbers is included
			if !strings.Contains(got, "```") {
				t.Error("prompt missing code fence")
			}
		})
	}
}

func TestGetCodeInput_NonExistent(t *testing.T) {
	_, _, err := getCodeInput("/nonexistent/file.go")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestGetCodeInput_Directory(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"main.go":   "package main",
		"helper.go": "package main\n\nfunc help() {}",
	}

	for name, content := range files {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	code, sourceName, err := getCodeInput(tempDir)
	if err != nil {
		t.Fatalf("getCodeInput() error = %v", err)
	}

	// Verify source name is directory name
	if sourceName != filepath.Base(tempDir) {
		t.Errorf("sourceName = %q, want %q",
			sourceName, filepath.Base(tempDir))
	}

	// Verify both files are included
	for name := range files {
		if !strings.Contains(code, name) {
			t.Errorf("code missing file %q", name)
		}
	}
}
