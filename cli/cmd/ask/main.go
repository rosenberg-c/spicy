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
	"module/lib/internal/filename"
	"module/lib/internal/filewriter"
	"module/lib/internal/history"
	"module/lib/internal/params"

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
				Value:   constants.ModelForCommand("ask"),
				Usage:   "Model to use",
			},
			&cli.StringFlag{
				Name:  "context",
				Usage: "Additional context to include with the question",
			},
			&cli.StringFlag{
				Name:  "context-file",
				Usage: "File path to include as context",
			},
			&cli.BoolFlag{
				Name:  "history",
				Usage: "Save command history to .spicy/ask/",
			},
			&cli.BoolFlag{
				Name:  "save",
				Usage: "Save output to timestamped markdown file",
			},
		},
		ArgsUsage: "[question...]",
		UsageText: `ask [options] [question...]

EXAMPLES:
    ask what is the difference between Go and Rust
    ask -v how does git rebase work
    ask -m openai/gpt-4 explain closures in JavaScript
    ask --history what is a goroutine
    ask --context-file selection.txt explain this code`,
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
	contextInput := cmd.String("context")
	contextFile := cmd.String("context-file")
	saveHistory := cmd.Bool("history")
	saveToFile := cmd.Bool("save")

	// Get question from args
	question := cmd.Args().Slice()

	// Get user input
	userInput, err := getUserInput(question)
	if err != nil {
		return fmt.Errorf("get user input: %w", err)
	}

	// Get context input
	contextContent, err := getContextInput(contextInput, contextFile)
	if err != nil {
		return fmt.Errorf("get context input: %w", err)
	}

	// Validate auth before running
	if err := agent.ValidateAuth(model); err != nil {
		return fmt.Errorf("auth error: %w", err)
	}

	// Create agent
	agentRunner := agent.New(verbose)

	// Generate answer
	fmt.Fprintln(os.Stderr, "Generating answer...")
	prompt := buildPrompt(userInput, contextContent)
	content, err := agentRunner.Run(ctx, model, prompt)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	if content == "" {
		return fmt.Errorf("agent returned empty content")
	}

	// Print the answer
	fmt.Println(content)

	// Save to file if enabled
	if saveToFile {
		outputFilename := filename.GenerateTimestamped("ask", userInput)
		finalPath, err := filewriter.WriteAtomic(outputFilename, content)
		if err != nil {
			return fmt.Errorf("save file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Saved to: %s\n", finalPath)
	}

	// Save to history if enabled
	if saveHistory {
		paramsMap := params.Base(model, verbose, saveHistory, saveToFile, "")
		paramsMap["context"] = contextInput
		paramsMap["context_file"] = contextFile
		historyData := map[string]interface{}{
			"question": userInput,
			"context":  contextContent,
			"result":   content,
			"params":   paramsMap,
		}
		// Use first part of question as filename suggestion
		if err := history.Save("ask", historyData, userInput); err != nil {
			// Log error but don't fail the command
			fmt.Fprintf(os.Stderr, "Warning: failed to save history: %v\n", err)
		}
	}

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

func getContextInput(contextInput, contextFile string) (string, error) {
	if contextInput != "" && contextFile != "" {
		return "", fmt.Errorf("use either --context or --context-file, not both")
	}
	if contextInput != "" {
		return contextInput, nil
	}
	if contextFile == "" {
		return "", nil
	}

	content, err := os.ReadFile(contextFile)
	if err != nil {
		return "", fmt.Errorf("read context file: %w", err)
	}

	return string(content), nil
}

func buildPrompt(input, context string) string {
	contextBlock := ""
	if strings.TrimSpace(context) != "" {
		contextBlock = "Additional context for reference:\n\n" + context + "\n\n"
	}

	return fmt.Sprintf(`You are a senior coder.
Answer the user question in a short concise manner.

The response must be valid markdown.

%sUser input:
%s`, contextBlock, input)
}
