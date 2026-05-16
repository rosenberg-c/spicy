package filewriter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// @req CORE-CLI-002
func TestWriteAtomic(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		content     string
		wantContent string
		wantErr     bool
		errContains string
	}{
		{
			name:        "simple write",
			path:        "test.txt",
			content:     "hello",
			wantContent: "hello\n",
			wantErr:     false,
		},
		{
			name:        "content already has newline",
			path:        "test.txt",
			content:     "hello\n",
			wantContent: "hello\n",
			wantErr:     false,
		},
		{
			name:        "nested directory creation",
			path:        "a/b/c/test.txt",
			content:     "nested",
			wantContent: "nested\n",
			wantErr:     false,
		},
		{
			name:        "multiple lines",
			path:        "multi.txt",
			content:     "line1\nline2\nline3",
			wantContent: "line1\nline2\nline3\n",
			wantErr:     false,
		},
		{
			name:        "empty content",
			path:        "empty.txt",
			content:     "",
			wantContent: "\n",
			wantErr:     false,
		},
		{
			name:        "special characters in content",
			path:        "special.txt",
			content:     "hello\tworld\n\nfoo",
			wantContent: "hello\tworld\n\nfoo\n",
			wantErr:     false,
		},
		{
			name:        "unicode content",
			path:        "unicode.txt",
			content:     "Hello 世界 🌍",
			wantContent: "Hello 世界 🌍\n",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			testPath := filepath.Join(tempDir, tt.path)

			gotPath, err := WriteAtomic(testPath, tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("WriteAtomic() error = %v, wantErr %v",
					err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" &&
					!strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want containing %q",
						err, tt.errContains)
				}
				return
			}

			// Verify returned path is absolute
			if !filepath.IsAbs(gotPath) {
				t.Errorf("returned path %q is not absolute", gotPath)
			}

			// Verify file exists at returned path
			if _, err := os.Stat(gotPath); os.IsNotExist(err) {
				t.Errorf("file does not exist at %q", gotPath)
			}

			// Verify file content
			got, err := os.ReadFile(gotPath)
			if err != nil {
				t.Fatalf("failed to read result: %v", err)
			}

			if string(got) != tt.wantContent {
				t.Errorf("content = %q, want %q", got, tt.wantContent)
			}
		})
	}
}

func TestWriteAtomic_Overwrite(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "overwrite.txt")

	// Write initial content
	_, err := WriteAtomic(testPath, "initial")
	if err != nil {
		t.Fatalf("initial write failed: %v", err)
	}

	// Overwrite with new content
	gotPath, err := WriteAtomic(testPath, "updated")
	if err != nil {
		t.Fatalf("overwrite failed: %v", err)
	}

	// Verify new content
	got, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}

	want := "updated\n"
	if string(got) != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

func TestWriteAtomic_AbsolutePath(t *testing.T) {
	tempDir := t.TempDir()
	absPath := filepath.Join(tempDir, "absolute.txt")

	gotPath, err := WriteAtomic(absPath, "absolute path test")
	if err != nil {
		t.Fatalf("WriteAtomic() failed: %v", err)
	}

	if gotPath != absPath {
		t.Errorf("returned path = %q, want %q", gotPath, absPath)
	}
}

func TestWriteAtomic_RelativePath(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	gotPath, err := WriteAtomic("relative.txt", "relative path")
	if err != nil {
		t.Fatalf("WriteAtomic() failed: %v", err)
	}

	// Returned path should be absolute
	if !filepath.IsAbs(gotPath) {
		t.Errorf("returned path %q is not absolute", gotPath)
	}

	// Verify file exists
	if _, err := os.Stat(gotPath); os.IsNotExist(err) {
		t.Errorf("file does not exist at %q", gotPath)
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantPrefix  string
		wantNoTilde bool
	}{
		{
			name:        "tilde only",
			path:        "~",
			wantNoTilde: true,
		},
		{
			name:        "tilde with slash",
			path:        "~/test.txt",
			wantNoTilde: true,
		},
		{
			name:        "tilde with subdirectory",
			path:        "~/subdir/test.txt",
			wantNoTilde: true,
		},
		{
			name:        "no tilde",
			path:        "/tmp/test.txt",
			wantPrefix:  "/tmp/test.txt",
			wantNoTilde: true,
		},
		{
			name:        "relative path",
			path:        "relative/test.txt",
			wantPrefix:  "relative/test.txt",
			wantNoTilde: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandPath(tt.path)
			if err != nil {
				t.Fatalf("expandPath() error = %v", err)
			}

			if tt.wantNoTilde && strings.Contains(got, "~") {
				t.Errorf("result %q still contains tilde", got)
			}

			if tt.wantPrefix != "" && !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("result %q does not start with %q",
					got, tt.wantPrefix)
			}

			// If path started with ~, verify expansion happened
			if strings.HasPrefix(tt.path, "~") {
				home, err := os.UserHomeDir()
				if err != nil {
					t.Fatalf("failed to get home dir: %v", err)
				}

				if tt.path == "~" {
					if got != home {
						t.Errorf("got %q, want %q", got, home)
					}
				} else {
					// Should start with home directory
					if !strings.HasPrefix(got, home) {
						t.Errorf("result %q does not start with home %q",
							got, home)
					}
				}
			}
		})
	}
}

func TestWriteAtomic_TildeExpansion(t *testing.T) {
	// Create a nested temp path using tilde
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	// Use temp directory under home
	tempSubdir := filepath.Join(".tmp-test-spicy", "tilde-test")
	tildePath := filepath.Join("~", tempSubdir, "test.txt")

	defer func() {
		// Cleanup
		os.RemoveAll(filepath.Join(home, ".tmp-test-spicy"))
	}()

	gotPath, err := WriteAtomic(tildePath, "tilde test")
	if err != nil {
		t.Fatalf("WriteAtomic() with tilde failed: %v", err)
	}

	// Verify path was expanded
	if strings.Contains(gotPath, "~") {
		t.Errorf("returned path %q still contains tilde", gotPath)
	}

	// Verify path starts with home
	if !strings.HasPrefix(gotPath, home) {
		t.Errorf("path %q does not start with home %q", gotPath, home)
	}

	// Verify file exists and has correct content
	got, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}

	want := "tilde test\n"
	if string(got) != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

func TestWriteAtomic_AtomicBehavior(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "atomic.txt")

	// Write initial content
	_, err := WriteAtomic(testPath, "initial")
	if err != nil {
		t.Fatalf("initial write failed: %v", err)
	}

	// Verify no temp files left behind
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".tmp") {
			t.Errorf("temp file left behind: %s", name)
		}
	}

	// Verify only the target file exists
	if len(entries) != 1 {
		t.Errorf("expected 1 file, got %d", len(entries))
	}
}
