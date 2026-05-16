package agent

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
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

func (a *Agent) Run(ctx context.Context, model, prompt string) (string, error) {
	// Build command: opencode run -m <model> <prompt>
	cmd := exec.CommandContext(ctx, "opencode", "run", "-m", model, prompt)

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
		"prompt_length", len(prompt))

	// Execute command
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("agent command failed: %w", err)
	}

	// Get output
	output := stdout.String()

	a.logger.Debug("agent response received",
		"output_length", len(output))

	return strings.TrimSpace(output), nil
}
