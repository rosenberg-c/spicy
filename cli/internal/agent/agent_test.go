package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// mockAgentRunner implements Runner for testing
type mockAgentRunner struct {
	response   string
	err        error
	callCount  int
	lastModel  string
	lastPrompt string
}

func (m *mockAgentRunner) Run(ctx context.Context, model, prompt string) (string, error) {
	m.callCount++
	m.lastModel = model
	m.lastPrompt = prompt

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if m.err != nil {
		return "", m.err
	}

	return m.response, nil
}

func TestAgent_New(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{
			name:    "non-verbose agent",
			verbose: false,
		},
		{
			name:    "verbose agent",
			verbose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := New(tt.verbose)
			if agent == nil {
				t.Error("New() returned nil")
			}
			if agent.verbose != tt.verbose {
				t.Errorf("verbose = %v, want %v", agent.verbose, tt.verbose)
			}
			if agent.logger == nil {
				t.Error("logger is nil")
			}
		})
	}
}

func TestMockAgentRunner_Run(t *testing.T) {
	tests := []struct {
		name          string
		model         string
		prompt        string
		mockResponse  string
		mockErr       error
		want          string
		wantErr       bool
		errContains   string
		contextCancel bool
	}{
		{
			name:         "successful execution",
			model:        "test-model",
			prompt:       "test prompt",
			mockResponse: "response from AI",
			want:         "response from AI",
			wantErr:      false,
		},
		{
			name:         "successful execution with whitespace",
			model:        "test-model",
			prompt:       "test prompt",
			mockResponse: "  response with spaces  ",
			want:         "  response with spaces  ",
			wantErr:      false,
		},
		{
			name:         "empty response",
			model:        "test-model",
			prompt:       "test prompt",
			mockResponse: "",
			want:         "",
			wantErr:      false,
		},
		{
			name:        "agent error",
			model:       "test-model",
			prompt:      "test prompt",
			mockErr:     fmt.Errorf("connection failed"),
			wantErr:     true,
			errContains: "connection failed",
		},
		{
			name:          "context cancellation",
			model:         "test-model",
			prompt:        "test prompt",
			mockResponse:  "should not return",
			contextCancel: true,
			wantErr:       true,
		},
		{
			name:         "multiline response",
			model:        "test-model",
			prompt:       "test prompt",
			mockResponse: "line1\nline2\nline3",
			want:         "line1\nline2\nline3",
			wantErr:      false,
		},
		{
			name:         "unicode response",
			model:        "test-model",
			prompt:       "test prompt",
			mockResponse: "Hello 世界 🌍",
			want:         "Hello 世界 🌍",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.contextCancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			mockRunner := &mockAgentRunner{
				response: tt.mockResponse,
				err:      tt.mockErr,
			}

			got, err := mockRunner.Run(ctx, tt.model, tt.prompt)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
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

			if got != tt.want {
				t.Errorf("Run() = %q, want %q", got, tt.want)
			}

			// Verify model and prompt were passed correctly
			if mockRunner.lastModel != tt.model {
				t.Errorf("lastModel = %q, want %q",
					mockRunner.lastModel, tt.model)
			}
			if mockRunner.lastPrompt != tt.prompt {
				t.Errorf("lastPrompt = %q, want %q",
					mockRunner.lastPrompt, tt.prompt)
			}
		})
	}
}

func TestMockAgentRunner_CallCount(t *testing.T) {
	mockRunner := &mockAgentRunner{
		response: "test response",
	}

	ctx := context.Background()

	// Call multiple times
	for i := 0; i < 3; i++ {
		_, err := mockRunner.Run(ctx, "model", "prompt")
		if err != nil {
			t.Fatalf("Run() failed: %v", err)
		}
	}

	if mockRunner.callCount != 3 {
		t.Errorf("callCount = %d, want 3", mockRunner.callCount)
	}
}

func TestMockAgentRunner_ContextTimeout(t *testing.T) {
	mockRunner := &mockAgentRunner{
		response: "response",
	}

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(1 * time.Millisecond)

	_, err := mockRunner.Run(ctx, "model", "prompt")
	if err == nil {
		t.Error("expected error from expired context, got nil")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("error = %v, want context.DeadlineExceeded", err)
	}
}

func TestAgent_RunnerInterface(t *testing.T) {
	// Verify Agent implements Runner interface
	var _ Runner = (*Agent)(nil)

	// Verify mockAgentRunner implements Runner interface
	var _ Runner = (*mockAgentRunner)(nil)
}

func TestMockUsage(t *testing.T) {
	// Create mock with specific response
	mock := &mockAgentRunner{
		response: "mocked AI response",
	}

	// Use mock as Runner
	var runner Runner = mock

	result, err := runner.Run(context.Background(), "gpt-4", "test prompt")
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if result != "mocked AI response" {
		t.Errorf("got %q, want %q", result, "mocked AI response")
	}

	// Verify the mock was called with correct parameters
	if mock.lastModel != "gpt-4" {
		t.Errorf("model = %q, want gpt-4", mock.lastModel)
	}

	if mock.callCount != 1 {
		t.Errorf("callCount = %d, want 1", mock.callCount)
	}
}

// @req CORE-CLI-008
func TestScratchCwdDir_CreatesPinnedDir(t *testing.T) {
	base := t.TempDir()

	dir, err := scratchCwdDir(base)
	if err != nil {
		t.Fatalf("scratchCwdDir() error = %v", err)
	}

	want := filepath.Join(base, "spicy", "v-edit-cwd")
	if dir != want {
		t.Fatalf("scratchCwdDir() = %q, want %q", dir, want)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("scratch dir was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("scratch dir is not a directory: %q", dir)
	}
}

// @req CORE-CLI-008
func TestResolveWorkingDir_HomeCwdUsesScratchDir(t *testing.T) {
	base := t.TempDir()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot resolve user home dir: %v", err)
	}

	got, err := resolveWorkingDir(home, base)
	if err != nil {
		t.Fatalf("resolveWorkingDir() error = %v", err)
	}

	want := filepath.Join(base, "spicy", "v-edit-cwd")
	if got != want {
		t.Fatalf("resolveWorkingDir() = %q, want %q", got, want)
	}
}

// @req CORE-CLI-008
func TestResolveWorkingDir_NonHomeCwdPreserved(t *testing.T) {
	base := t.TempDir()
	cwd := t.TempDir()

	got, err := resolveWorkingDir(cwd, base)
	if err != nil {
		t.Fatalf("resolveWorkingDir() error = %v", err)
	}

	if got != cwd {
		t.Fatalf("resolveWorkingDir() = %q, want %q", got, cwd)
	}

	// Must not create the scratch dir when preserving the real cwd.
	if _, err := os.Stat(filepath.Join(base, "spicy", "v-edit-cwd")); err == nil {
		t.Fatal("scratch dir created, want none when cwd is not home")
	}
}

// @req CORE-CLI-007
func TestKillProcessGroup_KillsGrandchildren(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "survived")

	// Direct child: a shell that spawns a grandchild and then blocks.
	// The grandchild writes the marker if it survives the group kill.
	cmd := exec.Command("sh", "-c", fmt.Sprintf(
		"(sleep 1 && touch %q) & exec sleep 30",
		marker,
	))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Give the grandchild a moment to fork.
	time.Sleep(100 * time.Millisecond)

	killProcessGroup(cmd)
	if err := cmd.Wait(); err == nil {
		t.Fatal("Wait() error = nil, want error after SIGKILL")
	}

	// Give a surviving grandchild the window to write the marker.
	time.Sleep(1200 * time.Millisecond)

	if _, err := os.Stat(marker); err == nil {
		t.Fatal("marker file exists: grandchild survived the process group kill")
	}
}

// @req CORE-CLI-007
func TestKillProcessGroup_NilProcess(t *testing.T) {
	cmd := exec.Command("true")

	// Must not panic on an unstarted command.
	killProcessGroup(cmd)
}
