package agent

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type Agent struct {
	verbose bool
	logger  *slog.Logger
}

func New(verbose bool) *Agent {
	// Set up logger with appropriate log level
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	if verbose {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	return &Agent{
		verbose: verbose,
		logger:  slog.New(handler),
	}
}

// scratchCwdDir returns the pinned working directory for agent subprocesses,
// creating it under base if needed.
func scratchCwdDir(base string) (string, error) {
	dir := filepath.Join(base, "spicy", "v-edit-cwd")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create scratch dir %s: %w", dir, err)
	}
	return dir, nil
}

// resolveWorkingDir returns the working directory for agent subprocesses.
// When the CLI runs from the user's home directory (or the home directory
// cannot be resolved), the subprocess is pinned to the scratch directory so
// model-generated files land there instead of the home directory. Otherwise
// the CLI's actual working directory is kept so files land next to the
// caller's project.
func resolveWorkingDir(cwd, base string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil || cwd == home {
		return scratchCwdDir(base)
	}
	return cwd, nil
}

// killProcessGroup best-effort kills the entire process group of cmd,
// including grandchildren spawned by the subprocess.
func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

func (a *Agent) Run(ctx context.Context, model, prompt string) (string, error) {
	// Build command: opencode run -m <model> <prompt>
	cmd := exec.CommandContext(ctx, "opencode", "run", "-m", model, prompt)

	// Run in its own process group so the whole tree can be killed on
	// timeout or failure.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Pin the working directory to a small scratch location only when the
	// CLI runs from the user's home directory, so model-generated files
	// land there instead of the home directory. Otherwise keep the CLI's
	// actual working directory.
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working dir: %w", err)
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache dir: %w", err)
	}
	workingDir, err := resolveWorkingDir(wd, cacheDir)
	if err != nil {
		return "", fmt.Errorf("prepare working dir: %w", err)
	}
	cmd.Dir = workingDir

	// Capture stdout
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// Handle stderr based on verbose flag
	if a.verbose {
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = nil // discard
	}

	a.logger.Debug("running agent command",
		"model", model,
		"prompt_length", len(prompt),
		"dir", workingDir)

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("agent command failed: %w", err)
	}

	// Watch for context cancellation and kill the whole process group;
	// CommandContext alone only kills the direct child.
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			killProcessGroup(cmd)
		case <-done:
		}
	}()

	waitErr := cmd.Wait()
	close(done)

	if waitErr != nil {
		killProcessGroup(cmd) // best effort for leftover grandchildren
		return "", fmt.Errorf("agent command failed: %w", waitErr)
	}

	// Get output
	output := stdout.String()

	a.logger.Debug("agent response received",
		"output_length", len(output))

	return strings.TrimSpace(output), nil
}
