package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"module/lib/internal/agent"
	"module/lib/internal/constants"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "ask",
		Usage: "Ask a question and get a concise answer using AI",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show verbose agent output",
			},
			&cli.StringFlag{
				Name:    "model",
				Aliases: []string{"m"},
				Value:   constants.DefaultModel,
				Usage:   "Model to use",
			},
		},
		ArgsUsage: "[question...]",
		UsageText: `ask [options] [question...]

EXAMPLES:
   ask what is the difference between Go and Rust
   ask -v how does git rebase work
   ask -m openai/gpt-4 explain closures in JavaScript`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()

			return run(runCtx, cmd)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	// Get flag values from cmd
	verbose := cmd.Bool("verbose")
	model := cmd.String("model")

	// Get question from args
	question := cmd.Args().Slice()

	// Get user input
	userInput, err := getUserInput(question)
	if err != nil {
		return fmt.Errorf("get user input: %w", err)
	}

	// Validate auth before running
	if err := agent.ValidateAuth(model); err != nil {
		return fmt.Errorf("auth error: %w", err)
	}

	// Create agent
	agentRunner := agent.New(verbose)

	// Generate answer
	fmt.Fprintln(os.Stderr, "Generating answer...")
	prompt := buildPrompt(userInput)
	content, err := agentRunner.Run(ctx, model, prompt)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	if content == "" {
		return fmt.Errorf("agent returned empty content")
	}

	// Print the answer
	fmt.Println(content)

	return nil
}

func getUserInput(question []string) (string, error) {
	if len(question) > 0 {
		input := strings.Join(question, " ")
		return input, nil
	}

	// Prompt for input
	fmt.Print("Question: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("empty input provided")
	}

	return input, nil
}

func buildPrompt(input string) string {
	return fmt.Sprintf(`You are a senior coder.
Answer the user question in a short concise manner.

The response must be valid markdown.

User input:
%s`, input)
}
