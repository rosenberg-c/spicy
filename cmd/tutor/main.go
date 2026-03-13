package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
	"module/lib/internal/agent"
	"module/lib/internal/filewriter"
	"module/lib/internal/validator"
)

func main() {
	cmd := &cli.Command{
		Name:  "tutor",
		Usage: "Generate technical tutorials using AI",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show verbose agent output",
			},
			&cli.StringFlag{
				Name:    "model",
				Aliases: []string{"m"},
				Value:   "openai/gpt-5.2",
				Usage:   "Model to use for both validation and generation",
			},
			&cli.StringFlag{
				Name:  "validation-model",
				Usage: "Model to use for validation only",
			},
			&cli.StringFlag{
				Name:  "generation-model",
				Usage: "Model to use for generation only",
			},
		},
		ArgsUsage: "[question...]",
		UsageText: `tutor [options] [question...]

EXAMPLES:
   tutor how to use docker compose
   tutor -v how does grep work
   tutor -m openai/gpt-4 how to use ffmpeg
   tutor --validation-model openai/gpt-4o --generation-model openai/o1 question`,
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

	// Determine validation and generation models
	validationModel := cmd.String("validation-model")
	if validationModel == "" {
		validationModel = baseModel
	}

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
	validationAgent := agent.New(verbose)
	generationAgent := agent.New(verbose)

	// Validate input
	fmt.Fprintln(os.Stderr, "Validating input...")
	v := validator.New(validationAgent, validationModel)
	validationResult, err := v.Validate(ctx, userInput)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if we should exit
	if validationResult.Action == validator.ActionExit {
		fmt.Fprintf(os.Stderr, "\n%s\n", validationResult.Reason)
		if len(validationResult.Suggestions) > 0 {
			fmt.Fprintln(os.Stderr, "\nSuggestions:")
			for i := range validationResult.Suggestions {
				fmt.Fprintf(os.Stderr, "  - %s\n", validationResult.Suggestions[i])
			}
		}
		return nil // Exit gracefully
	}

	// Ask for output path
	outputPath, err := getOutputPath(validationResult.SuggestedFilename)
	if err != nil {
		return fmt.Errorf("get output path: %w", err)
	}

	// Generate tutorial
	fmt.Fprintln(os.Stderr, "Generating tutorial...")
	prompt := buildTutorialPrompt(userInput)
	content, err := generationAgent.Run(ctx, generationModel, prompt)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	if content == "" {
		return fmt.Errorf("agent returned empty content")
	}

	// Write to file
	finalPath, err := filewriter.WriteAtomic(outputPath, content)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	fmt.Printf("Saved to: %s\n", finalPath)
	return nil
}

func getUserInput(question []string) (string, error) {
	if len(question) > 0 {
		input := strings.Join(question, " ")
		fmt.Printf("Question: %s\n", input)
		return input, nil
	}

	// Prompt for input
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

func getOutputPath(suggestedFilename string) (string, error) {
	if suggestedFilename == "" {
		suggestedFilename = "tutorial.md"
	}

	fmt.Printf("Save to file (default: %s) => ", suggestedFilename)

	// Open /dev/tty to read from terminal instead of stdin
	// This allows reading user input even when stdin is piped
	tty, err := os.Open("/dev/tty")
	if err != nil {
		// If /dev/tty is not available (non-interactive), use default
		fmt.Fprintln(os.Stderr, "\nNo terminal available, using default filename")
		return suggestedFilename, nil
	}
	defer tty.Close()

	reader := bufio.NewReader(tty)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return suggestedFilename, nil
	}

	return input, nil
}
