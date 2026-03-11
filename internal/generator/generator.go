package generator

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"module/lib/internal/agent"
)

type Generator struct {
	agent  agent.Runner
	model  string
	logger *slog.Logger
}

// New creates a Generator that uses the given agent and model to generate tutorials.
func New(agent agent.Runner, model string) *Generator {
	return &Generator{
		agent:  agent,
		model:  model,
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}
}

// Generate creates tutorial content for the given input.
// Returns the markdown content or an error if generation fails.
func (g *Generator) Generate(ctx context.Context, input string) (string, error) {
	prompt := BuildTutorialPrompt(input)

	g.logger.Debug("generating tutorial", "input", input, "model", g.model)

	// Call agent with configured model
	content, err := g.agent.Run(ctx, g.model, prompt)
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
