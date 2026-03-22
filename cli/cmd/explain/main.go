package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
	"module/lib/internal/agent"
	"module/lib/internal/constants"
	"module/lib/internal/filewriter"
	"module/lib/internal/history"
)

func main() {
	cmd := &cli.Command{
		Name:  "explain",
		Usage: "Explain code and optionally save to a markdown file",
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
			&cli.StringFlag{
				Name:    "language",
				Aliases: []string{"l", "lang"},
				Usage:   "Programming language (auto-detected if omitted)",
			},
			&cli.BoolFlag{
				Name:  "snippet",
				Usage: "Explain a snippet without requesting more context",
			},
			&cli.StringFlag{
				Name:  "context",
				Usage: "Additional context to include with the explanation",
			},
			&cli.StringFlag{
				Name:  "context-file",
				Usage: "File path to include as context",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file path (prompts if omitted)",
			},
			&cli.BoolFlag{
				Name:  "no-save",
				Usage: "Print to stdout instead of saving",
			},
			&cli.BoolFlag{
				Name:  "history",
				Usage: "Save command history to .spicy/explain/",
			},
			&cli.BoolFlag{
				Name:  "save",
				Usage: "Save explanation to file (prompts if --output omitted)",
			},
		},
		ArgsUsage: "[source]",
		UsageText: `explain [options] [source]

   source can be:
   - File path (e.g., main.go)
   - Directory path (e.g., ./internal/agent/)
   - '-' or omitted for stdin

EXAMPLES:
   explain main.go
   explain ./internal/agent/
   pbpaste | explain
   explain main.go --save
   explain main.go -o explanation.md
   cat complex.go | explain --lang go`,
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
	fmt.Println("== Code Explainer ==")

	// Get flag values from cmd
	verbose := cmd.Bool("verbose")
	model := cmd.String("model")
	language := cmd.String("language")
	snippet := cmd.Bool("snippet")
	contextInput := cmd.String("context")
	contextFile := cmd.String("context-file")
	output := cmd.String("output")
	noSave := cmd.Bool("no-save")
	saveHistory := cmd.Bool("history")
	saveFlag := cmd.Bool("save")
	saveRequested := shouldSave(output, saveFlag, noSave)

	// Get source from args (first positional argument)
	source := cmd.Args().First()

	// Get code input
	code, sourceName, err := getCodeInput(source)
	if err != nil {
		return fmt.Errorf("get code input: %w", err)
	}

	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("no code provided")
	}

	// Auto-detect language if not specified
	if language == "" {
		language = detectLanguage(source, code)
	}

	// Validate auth before running
	if err := agent.ValidateAuth(model); err != nil {
		return err
	}

	contextContent, err := getContextInput(contextInput, contextFile)
	if err != nil {
		return fmt.Errorf("get context input: %w", err)
	}

	// Build prompt
	prompt := buildExplanationPrompt(code, language, snippet, contextContent)

	// Generate explanation
	fmt.Fprintln(os.Stderr, "Generating explanation...")
	agentRunner := agent.New(verbose)
	explanation, err := agentRunner.Run(ctx, model, prompt)
	if err != nil {
		return fmt.Errorf("generate explanation: %w", err)
	}

	explanation = strings.TrimSpace(explanation)
	explanation = sanitizeExplanation(explanation, snippet)

	// Print to stdout if not saving
	if !saveRequested {
		fmt.Println(explanation)

		// Save to history if enabled
		if saveHistory {
			historyData := map[string]interface{}{
				"source":   sourceName,
				"code":     code,
				"language": language,
				"result":   explanation,
				"context":  contextContent,
				"params": map[string]interface{}{
					"model":        model,
					"verbose":      verbose,
					"output":       output,
					"no_save":      true,
					"save":         false,
					"history":      saveHistory,
					"source":       source,
					"language":     cmd.String("language"),
					"snippet":      snippet,
					"context":      contextInput,
					"context_file": contextFile,
				},
			}
			// Use source name as filename suggestion
			if err := history.Save("explain", historyData, sourceName); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save history: %v\n", err)
			}
		}

		return nil
	}

	// Determine output path
	outputPath := output
	if outputPath == "" {
		suggested := suggestFilename(sourceName, language)

		// If --save is used, prepend timestamp and command name to the suggestion
		if saveFlag {
			suggested = strings.TrimSuffix(suggested, ".md")
			timestamp := time.Now().Format("2006-01-02_15-04")
			suggested = fmt.Sprintf("%s_explain_%s.md", timestamp, suggested)
		}

		// Prompt for filename with the suggestion
		outputPath, err = getOutputPath(suggested)
		if err != nil {
			return fmt.Errorf("get output path: %w", err)
		}
	}

	// Write to file
	finalPath, err := filewriter.WriteAtomic(outputPath, explanation)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	fmt.Printf("Saved to: %s\n", finalPath)

	// Save to history if enabled
	if saveHistory {
		historyData := map[string]interface{}{
			"source":   sourceName,
			"code":     code,
			"language": language,
			"output":   finalPath,
			"result":   explanation,
			"context":  contextContent,
			"params": map[string]interface{}{
				"model":        model,
				"verbose":      verbose,
				"output":       output,
				"no_save":      false,
				"save":         true,
				"history":      saveHistory,
				"source":       source,
				"language":     cmd.String("language"),
				"snippet":      snippet,
				"context":      contextInput,
				"context_file": contextFile,
			},
		}
		// Use source name as filename suggestion
		if err := history.Save("explain", historyData, sourceName); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save history: %v\n", err)
		}
	}

	return nil
}

func shouldSave(output string, saveFlag bool, noSave bool) bool {
	if noSave {
		return false
	}

	return saveFlag || output != ""
}

func getCodeInput(source string) (code, sourceName string, err error) {
	// Read from stdin
	if source == "" || source == "-" {
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", "", fmt.Errorf("read stdin: %w", err)
		}
		return string(content), "stdin", nil
	}

	// Check if it's a file or directory
	info, err := os.Stat(source)
	if err != nil {
		return "", "", fmt.Errorf("stat %s: %w", source, err)
	}

	if info.IsDir() {
		// Read all code files in directory
		content, err := readDirectory(source)
		if err != nil {
			return "", "", fmt.Errorf("read directory: %w", err)
		}
		return content, filepath.Base(source), nil
	}

	// Read single file
	content, err := os.ReadFile(source)
	if err != nil {
		return "", "", fmt.Errorf("read file: %w", err)
	}

	return string(content), filepath.Base(source), nil
}

func readDirectory(dir string) (string, error) {
	var builder strings.Builder
	codeExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true,
		".java": true, ".c": true, ".cpp": true, ".rs": true,
		".rb": true, ".php": true, ".sh": true, ".md": true,
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip common non-code directories
		if info.IsDir() {
			skip := []string{"node_modules", "vendor", "bin", ".git"}

			if slices.Contains(skip, info.Name()) {
				return filepath.SkipDir
			}

			return nil
		}
		// Check if it's a code file
		ext := filepath.Ext(path)
		if !codeExts[ext] {
			return nil
		}

		// Read and append
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		relPath, _ := filepath.Rel(dir, path)
		fmt.Fprintf(&builder, "// File: %s\n", relPath)
		builder.WriteString(string(content))
		builder.WriteString("\n\n")

		return nil
	})
	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

func detectLanguage(source, code string) string {
	if source != "" && source != "-" && source != "stdin" {
		ext := filepath.Ext(source)
		langMap := map[string]string{
			".go":   "Go",
			".py":   "Python",
			".js":   "JavaScript",
			".ts":   "TypeScript",
			".java": "Java",
			".c":    "C",
			".cpp":  "C++",
			".rs":   "Rust",
			".rb":   "Ruby",
			".php":  "PHP",
			".sh":   "Shell",
		}
		if lang, ok := langMap[ext]; ok {
			return lang
		}
	}

	// Try to detect from code content
	if strings.Contains(code, "package main") || strings.Contains(code, "func ") {
		return "Go"
	}
	if strings.Contains(code, "def ") || strings.Contains(code, "import ") {
		return "Python"
	}
	if strings.Contains(code, "function ") || strings.Contains(code, "const ") {
		return "JavaScript"
	}

	return "code"
}

func buildExplanationPrompt(code, language string, snippet bool, context string) string {
	contextNote := ""
	if snippet {
		contextNote = "The input is a short snippet. Explain only what is present without asking for more context or requesting more code."
	}
	contextBlock := ""
	if strings.TrimSpace(context) != "" {
		contextBlock = fmt.Sprintf("\nAdditional context for reference (focus on the snippet):\n\n```\n%s\n```\n", context)
	}
	return fmt.Sprintf(`You are a senior software engineer and technical educator.
Explain the following %s code in detail.
Write the explanation as a clear, well-structured markdown document.

%s
%s

Rules:
- Do not ask for more context or additional code.
- Do not include system reminders or meta commentary.
- Focus on what can be inferred from the snippet and provided context.

Include:
- High-level summary of what the code does
- Step-by-step breakdown of key components
- Explanation of important patterns or techniques used
- Any notable design decisions or best practices
- Potential improvements or considerations (if applicable)

Be thorough but clear.
Assume the reader is a developer who wants to understand the code.

Code to explain:

%s`, language, contextNote, contextBlock, addLineNumbers(code))
}

func sanitizeExplanation(input string, snippet bool) string {
	if !snippet {
		return input
	}

	lines := strings.Split(input, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "system-reminder") {
			continue
		}
		if strings.Contains(lower, "please paste") ||
			strings.Contains(lower, "share the full") ||
			strings.Contains(lower, "i only have") ||
			strings.Contains(lower, "i only see") ||
			strings.Contains(lower, "can't explain") ||
			strings.Contains(lower, "cannot explain") ||
			strings.Contains(lower, "need more context") {
			continue
		}
		filtered = append(filtered, line)
	}

	result := strings.TrimSpace(strings.Join(filtered, "\n"))
	if result == "" {
		return strings.TrimSpace(input)
	}
	return result
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

func addLineNumbers(code string) string {
	lines := strings.Split(code, "\n")
	var builder strings.Builder
	builder.WriteString("```\n")
	for i, line := range lines {
		builder.WriteString(fmt.Sprintf("%4d  %s\n", i+1, line))
	}
	builder.WriteString("```\n")
	return builder.String()
}

func suggestFilename(sourceName, language string) string {
	if sourceName == "stdin" || sourceName == "" {
		return fmt.Sprintf("%s-explanation.md", strings.ToLower(language))
	}

	// Remove extension and add -explanation
	name := strings.TrimSuffix(sourceName, filepath.Ext(sourceName))
	return fmt.Sprintf("%s-explanation.md", name)
}

func getOutputPath(suggestedFilename string) (string, error) {
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
