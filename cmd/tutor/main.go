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
	"module/lib/internal/filewriter"
	"module/lib/internal/generator"
	"module/lib/internal/validator"
)

// Config holds command-line arguments.
type Config struct {
	Question        []string
	Verbose         bool
	ValidationModel string
	GenerationModel string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	// Parse command-line arguments
	config := parseArgs()

	fmt.Println("== Generate Tutorial ==")

	// Get user input
	userInput, err := getUserInput(config.Question)
	if err != nil {
		return fmt.Errorf("get user input: %w", err)
	}

	// Create separate agents for validation and generation
	validationAgent := agent.New(config.Verbose)
	generationAgent := agent.New(config.Verbose)

	// Validate input
	fmt.Fprintln(os.Stderr, "Validating input...")
	v := validator.New(validationAgent, config.ValidationModel)
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
	gen := generator.New(generationAgent, config.GenerationModel)
	content, err := gen.Generate(ctx, userInput)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// Write to file
	finalPath, err := filewriter.WriteAtomic(outputPath, content)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	fmt.Printf("Saved to: %s\n", finalPath)
	return nil
}

func parseArgs() Config {
	config := Config{
		ValidationModel: "openai/gpt-5.2",
		GenerationModel: "openai/gpt-5.2",
		Verbose:         false,
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "-v" || arg == "--verbose" {
			config.Verbose = true
		} else if arg == "--validation-model" {
			if i+1 < len(args) {
				config.ValidationModel = args[i+1]
				i++
			}
		} else if arg == "--generation-model" {
			if i+1 < len(args) {
				config.GenerationModel = args[i+1]
				i++
			}
		} else if arg == "-m" || arg == "--model" {
			// Set both models to the same value for convenience
			if i+1 < len(args) {
				config.ValidationModel = args[i+1]
				config.GenerationModel = args[i+1]
				i++
			}
		} else if arg == "-h" || arg == "--help" {
			printHelp()
			os.Exit(0)
		} else {
			// Treat as question
			config.Question = append(config.Question, arg)
		}
	}

	return config
}

func printHelp() {
	fmt.Println("Usage: tutor [options] [question...]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -v, --verbose              Show verbose agent output")
	fmt.Println("  -m, --model STR            Model to use for both validation and generation")
	fmt.Println("                             (default: openai/gpt-5.2)")
	fmt.Println("  --validation-model STR     Model to use for validation only")
	fmt.Println("  --generation-model STR     Model to use for generation only")
	fmt.Println("  -h, --help                 Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  tutor how to use docker compose")
	fmt.Println("  tutor -v how does grep work")
	fmt.Println("  tutor -m openai/gpt-4 how to use ffmpeg")
	fmt.Println("  tutor --validation-model openai/gpt-4o --generation-model openai/o1 question")
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
	reader := bufio.NewReader(os.Stdin)
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
