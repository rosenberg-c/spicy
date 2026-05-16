package askwrapper

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRunAsk_Success(t *testing.T) {
	binDir := t.TempDir()
	writeAskScript(t, binDir, "#!/usr/bin/env bash\n", "printf 'hello from fake ask\\n'\n")

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+originalPath)

	out, err := RunAsk(context.Background(), "why", 2*time.Second)
	if err != nil {
		t.Fatalf("RunAsk() error = %v", err)
	}
	if out != "hello from fake ask" {
		t.Fatalf("RunAsk() output = %q, want %q", out, "hello from fake ask")
	}
}

func TestRunAsk_FailureIncludesStderr(t *testing.T) {
	binDir := t.TempDir()
	writeAskScript(t, binDir, "#!/usr/bin/env bash\n", "echo 'boom' 1>&2\n", "exit 12\n")

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+originalPath)

	_, err := RunAsk(context.Background(), "why", 2*time.Second)
	if err == nil {
		t.Fatal("RunAsk() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "ask failed: boom") {
		t.Fatalf("RunAsk() error = %q, want to contain %q", err.Error(), "ask failed: boom")
	}
}

func TestRunAsk_Timeout(t *testing.T) {
	binDir := t.TempDir()
	writeAskScript(t, binDir, "#!/usr/bin/env bash\n", "sleep 2\n", "printf 'late output\\n'\n")

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+originalPath)

	_, err := RunAsk(context.Background(), "why", 50*time.Millisecond)
	if err == nil {
		t.Fatal("RunAsk() error = nil, want timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("RunAsk() error = %q, want timeout message", err.Error())
	}
}

func TestRunAsk_EmptyQuestion(t *testing.T) {
	_, err := RunAsk(context.Background(), "   ", time.Second)
	if err == nil {
		t.Fatal("RunAsk() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "empty question") {
		t.Fatalf("RunAsk() error = %q, want empty question error", err.Error())
	}
}

func writeAskScript(t *testing.T, dir string, lines ...string) {
	t.Helper()
	name := "ask"
	if runtime.GOOS == "windows" {
		name = "ask.bat"
	}
	path := filepath.Join(dir, name)
	content := strings.Join(lines, "")
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write fake ask script: %v", err)
	}
}
