package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
	"module/lib/internal/agent"
)

func main() {
	cmd := &cli.Command{
		Name:  "gitmessage",
		Usage: "Generate commit messages using AI",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show verbose agent output",
			},
			&cli.BoolFlag{
				Name:    "copy",
				Aliases: []string{"c"},
				Usage:   "Copy result to clipboard",
			},
			&cli.StringFlag{
				Name:    "model",
				Aliases: []string{"m"},
				Value:   "openai/gpt-5.2-codex",
				Usage:   "Model to use",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			defer cancel()

			return run(runCtx, cmd)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	// Get flag values from cmd instead of c
	verbose := cmd.Bool("verbose")
	model := cmd.String("model")
	copy := cmd.Bool("copy")
	prefix := cmd.Args().First()

	// Get staged diff
	fmt.Fprintln(os.Stderr, "Running: git diff --staged")
	diff, err := getStagedDiff(ctx)
	if err != nil {
		return fmt.Errorf("get staged diff: %w", err)
	}

	// Check if there are staged changes
	if strings.TrimSpace(diff) == "" {
		fmt.Fprintln(
			os.Stderr,
			"Warning: No staged changes available. Please stage your changes first.",
		)
		return nil
	}

	// Build prompt
	prompt := buildPrompt(diff)

	// Generate commit message
	fmt.Fprintf(os.Stderr, "Running: opencode run --agent build -m %s\n", model)
	fmt.Fprintln(os.Stderr, "Generating commit message...")

	agentRunner := agent.New(verbose)
	generatedMsg, err := agentRunner.Run(ctx, model, prompt)
	if err != nil {
		return fmt.Errorf("generate commit message: %w", err)
	}

	generatedMsg = strings.TrimSpace(generatedMsg)

	// Conditionally prepend prefix
	var finalMsg string
	if prefix != "" {
		finalMsg = fmt.Sprintf("%s: %s", prefix, generatedMsg)
	} else {
		finalMsg = generatedMsg
	}

	// Print result
	fmt.Printf("==> %s\n", finalMsg)

	// Optionally copy to clipboard
	if copy {
		if err := copyToClipboard(finalMsg); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Warning: failed to copy to clipboard: %v\n",
				err,
			)
		}
	}

	return nil
}

func buildPrompt(diff string) string {
	return fmt.Sprintf(`You are a senior coder.
Write a short commit message, one row only.
Do not include the actual diff, or any other thoughts.
Only output the commit message.
Always use Capital character at the beginning.
Do not add any quotes or special characters around the response.

Diff:
%s`, diff)
}

func getStagedDiff(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", "--staged")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	return string(output), nil
}

func copyToClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
