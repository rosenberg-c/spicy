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
	"module/lib/internal/constants"
	"module/lib/internal/history"
	"module/lib/internal/params"
)

func main() {
	cmd := &cli.Command{
		Name:  "gitmessage",
		Usage: "Generate commit messages using AI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "hint",
				Aliases: []string{"i"},
				Usage:   "Add hint to the llm",
			},
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
				Value:   constants.DefaultModel,
				Usage:   "Model to use",
			},
			&cli.BoolFlag{
				Name:  "history",
				Usage: "Save command history to .spicy/gitmessage/",
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
	hint := cmd.String("hint")
	prefix := cmd.Args().First()
	saveHistory := cmd.Bool("history")

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

	// Validate auth before running
	if err := agent.ValidateAuth(model); err != nil {
		return err
	}

	// Build prompt
	prompt := buildPrompt(hint, diff)

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

	// Save to history if enabled
	if saveHistory {
		paramsMap := params.Base(model, verbose, saveHistory, false, "")
		paramsMap["copy"] = copy
		paramsMap["hint"] = hint
		paramsMap["prefix"] = prefix

		historyData := map[string]interface{}{
			"hint":   hint,
			"prefix": prefix,
			"result": finalMsg,
			"params": paramsMap,
		}
		// Use commit message as filename suggestion
		if err := history.Save("gitmessage", historyData, finalMsg); err != nil {
			// Log error but don't fail the command
			fmt.Fprintf(os.Stderr, "Warning: failed to save history: %v\n", err)
		}
	}

	return nil
}

func buildPrompt(diff string, hint string) string {
	return fmt.Sprintf(`You are a senior coder.
Write a short commit message, one row only.
Do not include the actual diff, or any other thoughts.
Only output the commit message.
Always use Capital character at the beginning.
Do not add any quotes or special characters around the response.

Hint:
%s

Diff:
%s`, hint, diff)
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
