package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"module/lib/internal/agent"
)

// Config holds command-line arguments.
type Config struct {
	Prefix  string
	Copy    bool
	Verbose bool
	Model   string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	config := parseArgs(os.Args[1:])

	// Get staged diff
	fmt.Fprintln(os.Stderr, "Running: git diff --staged")
	diff, err := getStagedDiff(ctx)
	if err != nil {
		return fmt.Errorf("get staged diff: %w", err)
	}

	// Check if there are staged changes
	if strings.TrimSpace(diff) == "" {
		fmt.Fprintln(os.Stderr, "Warning: No staged changes available. Please stage your changes first.")
		return nil
	}

	// Build prompt
	prompt := buildPrompt(diff)

	// Generate commit message
	fmt.Fprintf(os.Stderr, "Running: opencode run --agent build -m %s\n", config.Model)
	fmt.Fprintln(os.Stderr, "Generating commit message...")

	agentRunner := agent.New(config.Verbose)
	generatedMsg, err := agentRunner.Run(ctx, config.Model, prompt)
	if err != nil {
		return fmt.Errorf("generate commit message: %w", err)
	}

	generatedMsg = strings.TrimSpace(generatedMsg)

	// Conditionally prepend prefix
	var finalMsg string
	if config.Prefix != "" {
		finalMsg = fmt.Sprintf("%s: %s", config.Prefix, generatedMsg)
	} else {
		finalMsg = generatedMsg
	}

	// Print result
	fmt.Printf("==> %s\n", finalMsg)

	// Optionally copy to clipboard
	if config.Copy {
		if err := copyToClipboard(finalMsg); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to copy to clipboard: %v\n", err)
		}
	}

	return nil
}

func parseArgs(args []string) Config {
	config := Config{
		Model:   "openai/gpt-5.2-codex",
		Verbose: false,
		Copy:    false,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-v", "--verbose":
			config.Verbose = true
		case "-c", "--copy":
			config.Copy = true
		case "-m", "--model":
			if i+1 < len(args) {
				config.Model = args[i+1]
				i++
			}
		case "-h", "--help":
			printHelp()
			os.Exit(0)
		default:
			// First non-flag argument is the prefix
			if config.Prefix == "" {
				config.Prefix = arg
			}
		}
	}

	return config
}

func printHelp() {
	fmt.Println("Usage: gitmessage [options] [prefix]")
	fmt.Println()
	fmt.Println("Generate commit messages using AI based on staged git changes.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -c, --copy              Copy result to clipboard")
	fmt.Println("  -v, --verbose           Show verbose agent output")
	fmt.Println("  -m, --model STR         Model to use (default: openai/gpt-5.2-codex)")
	fmt.Println("  -h, --help              Show this help")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  prefix                  Optional prefix for commit message (e.g., 'feat', 'fix')")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gitmessage")
	fmt.Println("  gitmessage feat")
	fmt.Println("  gitmessage fix -c")
	fmt.Println("  gitmessage -m openai/gpt-4 refactor")
}

func buildPrompt(diff string) string {
	return fmt.Sprintf(`You are a senior coder: write a short commit message, one row only.
Do not include the actual diff, or any other thoughts, only the commit message.
Always use Capital character at the beginning of the commit message.
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
