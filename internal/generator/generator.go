package generator

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"module/tutor/internal/agent"
)

type Generator struct {
	agent  agent.Runner
	logger *slog.Logger
}

// New creates a Generator that uses the given agent to generate tutorials.
func New(agent agent.Runner) *Generator {
	return &Generator{
		agent: agent,
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}
}

// Generate creates tutorial content for the given input.
// Returns the markdown content or an error if generation fails.
func (g *Generator) Generate(ctx context.Context, input string) (string, error) {
	prompt := BuildTutorialPrompt(input)

	g.logger.Debug("generating tutorial", "input", input)

	// Call agent with hardcoded model
	content, err := g.agent.Run(ctx, "openai/gpt-5.2", prompt)
	if err != nil {
		return "", fmt.Errorf("agent call failed: %w", err)
	}

	// Validate content is not empty
	if content == "" {
		return "", fmt.Errorf("agent returned empty content")
	}

	g.logger.Info("tutorial generated",
		"content_length", len(content))

	return content, nil
}
