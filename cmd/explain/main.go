package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"module/lib/internal/agent"
	"module/lib/internal/filewriter"
)

// Config holds command-line arguments.
type Config struct {
	Source   string
	Language string
	Output   string
	Verbose  bool
	Model    string
	NoSave   bool
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	config := parseArgs(os.Args[1:])

	fmt.Println("== Code Explainer ==")

	// Get code input
	code, sourceName, err := getCodeInput(config.Source)
	if err != nil {
		return fmt.Errorf("get code input: %w", err)
	}

	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("no code provided")
	}

	// Auto-detect language if not specified
	if config.Language == "" {
		config.Language = detectLanguage(config.Source, code)
	}

	// Build prompt
	prompt := buildExplanationPrompt(code, config.Language)

	// Generate explanation
	fmt.Fprintln(os.Stderr, "Generating explanation...")
	agentRunner := agent.New(config.Verbose)
	explanation, err := agentRunner.Run(ctx, config.Model, prompt)
	if err != nil {
		return fmt.Errorf("generate explanation: %w", err)
	}

	explanation = strings.TrimSpace(explanation)

	// Print to stdout if --no-save
	if config.NoSave {
		fmt.Println(explanation)
		return nil
	}

	// Determine output path
	outputPath := config.Output
	if outputPath == "" {
		outputPath, err = getOutputPath(suggestFilename(sourceName, config.Language))
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
	return nil
}

func parseArgs(args []string) Config {
	config := Config{
		Model:   "openai/gpt-5.2",
		Verbose: false,
		NoSave:  false,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-v", "--verbose":
			config.Verbose = true
		case "-m", "--model":
			if i+1 < len(args) {
				config.Model = args[i+1]
				i++
			}
		case "-l", "--language", "--lang":
			if i+1 < len(args) {
				config.Language = args[i+1]
				i++
			}
		case "-o", "--output":
			if i+1 < len(args) {
				config.Output = args[i+1]
				i++
			}
		case "--no-save":
			config.NoSave = true
		case "-h", "--help":
			printHelp()
			os.Exit(0)
		default:
			// First non-flag argument is the source
			if config.Source == "" {
				config.Source = arg
			}
		}
	}

	return config
}

func printHelp() {
	fmt.Println("Usage: explain [options] [source]")
	fmt.Println()
	fmt.Println("Explain code and save the explanation as a markdown file.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -v, --verbose           Show verbose agent output")
	fmt.Println("  -m, --model STR         Model to use (default: openai/gpt-5.2)")
	fmt.Println("  -l, --lang STR          Programming language (auto-detected if omitted)")
	fmt.Println("  -o, --output PATH       Output file path (prompts if omitted)")
	fmt.Println("  --no-save               Print to stdout instead of saving")
	fmt.Println("  -h, --help              Show this help")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  source                  File, directory, or '-' for stdin")
	fmt.Println("                          If omitted, reads from stdin")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  explain main.go")
	fmt.Println("  explain ./internal/agent/")
	fmt.Println("  pbpaste | explain")
	fmt.Println("  explain main.go -o explanation.md")
	fmt.Println("  cat complex.go | explain --lang go --no-save")
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
			for _, s := range skip {
				if info.Name() == s {
					return filepath.SkipDir
				}
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
		builder.WriteString(fmt.Sprintf("// File: %s\n", relPath))
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

func buildExplanationPrompt(code, language string) string {
	return fmt.Sprintf(`You are a senior software engineer and technical educator.
Explain the following %s code in detail.
Write the explanation as a clear, well-structured markdown document.

Include:
- High-level summary of what the code does
- Step-by-step breakdown of key components
- Explanation of important patterns or techniques used
- Any notable design decisions or best practices
- Potential improvements or considerations (if applicable)

Be thorough but clear. Assume the reader is a developer who wants to understand the code.

Code to explain:

%s`, language, addLineNumbers(code))
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
