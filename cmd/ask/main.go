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
		Name:  "explain",
		Usage: "Explain code and save the explanation as a markdown file",
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
		ArgsUsage: "[question]",
		UsageText: `ask [options] [question]
`,
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
	fmt.Println("== Generate Tutorial ==")

	// Get flag values from cmd
	verbose := cmd.Bool("verbose")
	baseModel := cmd.String("model")

	generationModel := cmd.String("generation-model")
	if generationModel == "" {
		generationModel = baseModel
	}

	// Get question from args
	question := cmd.Args().Slice()

	// Get user input
	userInput, err := getUserInput(question)
	if err != nil {
		return fmt.Errorf("get user input: %w", err)
	}

	// Create separate agents for validation and generation
	generationAgent := agent.New(verbose)

	// Generate tutorial
	fmt.Fprintln(os.Stderr, "Generating tutorial...")
	prompt := buildPrompt(userInput)
	content, err := generationAgent.Run(ctx, generationModel, prompt)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	if content == "" {
		return fmt.Errorf("agent returned empty content")
	}

	return nil
}

func getUserInput(question []string) (string, error) {
	if len(question) > 0 {
		input := strings.Join(question, " ")
		fmt.Printf("Question: %s\n", input)
		return input, nil
	}

	fmt.Print("-- input: ")
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
