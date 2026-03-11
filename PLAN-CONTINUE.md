# Tutorial Generator - Implementation Continuation Plan

## Current State Analysis

### ✅ Completed
- [x] Project structure created (cmd/, internal/ directories)
- [x] go.mod initialized (`module module/tutor`)
- [x] Makefile with common targets
- [x] Type definitions (Action, ValidationResponse, Config, Runner interface)
- [x] Error types (ValidationError, sentinel errors)
- [x] Test structure (table-driven tests with mocks)
- [x] Test utilities (CreateTempDir)

### ⚠️ Issues to Fix
1. Missing package declarations in test files
2. Typo in `types.go` JSON tag: `suggested_file_name:omitempty` → `suggested_filename,omitempty`
3. Error in `errors.go`: method should be `Unwrap() error`, not `error()`
4. Missing `errors` import in `errors.go`

### ❌ Not Implemented (Stubs Only)
1. **agent package**: `New()`, `Run()` - subprocess execution
2. **validator package**: `New()`, `Validate()`, `BuildValidationPrompt()`
3. **generator package**: `New()`, `Generate()`
4. **filewriter package**: `WriteAtomic()`
5. **CLI**: `cmd/tutor/main.go` `run()` function

---

## Implementation Plan

### Phase 1: Fix Compilation Errors (15 minutes)

#### Step 1.1: Fix validator/types.go JSON tag
**File**: `internal/validator/types.go`
**Line**: 14
**Change**:
```go
// FROM:
SuggestedFilename string   `json:"suggested_file_name:omitempty"`

// TO:
SuggestedFilename string   `json:"suggested_filename,omitempty"`
```
**Why**: JSON key should match what the Python version uses and have proper comma separator.

#### Step 1.2: Fix validator/errors.go Unwrap method
**File**: `internal/validator/errors.go`
**Line**: 15-17
**Change**:
```go
// FROM:
func (e *ValidationError) error {
	return e.Err
}

// TO:
func (e *ValidationError) Unwrap() error {
	return e.Err
}
```
**Why**: `Unwrap()` is the standard Go interface for error unwrapping. `error` is not a valid method name.

#### Step 1.3: Add missing import to validator/errors.go
**File**: `internal/validator/errors.go`
**Add after line 3**:
```go
import (
	"errors"
	"fmt"
)
```
**Why**: The sentinel errors use `errors.New()` which requires the `errors` package.

#### Step 1.4: Add package declaration to validator_test.go
**File**: `internal/validator/validator_test.go`
**Add at line 1**:
```go
package validator

import (
	"context"
	"reflect"
	"testing"
)
```

#### Step 1.5: Add package declaration to testutil.go
**File**: `internal/testutil.go`
**Add at line 1**:
```go
package testutil

import (
	"os"
	"testing"
)
```

**Verify**: Run `make vet` - should now only show "missing function body" errors.

---

### Phase 2: Implement Agent Package (45 minutes)

The agent package is the foundation - everything else depends on it.

#### Step 2.1: Implement agent.New()
**File**: `internal/agent/agent.go`

**Complete implementation**:
```go
package agent

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"subprocess"
)

type Agent struct {
	verbose bool
	logger  *slog.Logger
}

func New(verbose bool) *Agent {
	// Set up logger with appropriate log level
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	if verbose {
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	return &Agent{
		verbose: verbose,
		logger:  slog.New(handler),
	}
}

func (a *Agent) Run(ctx context.Context, model, prompt string) (string, error) {
	// Build command: opencode run --agent build -m <model> <prompt>
	cmd := exec.CommandContext(ctx, "opencode", "run", "--agent", "build", "-m", model, prompt)

	// Capture stdout
	cmd.Stdout = &bytes.Buffer{}

	// Handle stderr based on verbose flag
	if a.verbose {
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = nil // discard
	}

	a.logger.Debug("running agent command",
		"model", model,
		"prompt_length", len(prompt))

	// Execute command
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("agent command failed: %w", err)
	}

	// Get output
	output := cmd.Stdout.(*bytes.Buffer).String()

	a.logger.Debug("agent response received",
		"output_length", len(output))

	return strings.TrimSpace(output), nil
}
```

**Add missing imports**:
```go
import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)
```

**Key Points**:
- Uses `exec.CommandContext` to respect context cancellation
- Redirects stderr to `os.Stderr` if verbose, otherwise discards it
- Returns trimmed output (removes leading/trailing whitespace)
- Wraps errors with context using `fmt.Errorf` with `%w`

**Test manually**:
```bash
# Create a simple test program
cat > test_agent.go <<'EOF'
package main

import (
	"context"
	"fmt"
	"log"

	"module/tutor/internal/agent"
)

func main() {
	a := agent.New(true)
	response, err := a.Run(context.Background(), "openai/gpt-4", "Say hello in one word")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Response:", response)
}
EOF

go run test_agent.go
rm test_agent.go
```

---

### Phase 3: Implement Validator Package (60 minutes)

#### Step 3.1: Implement BuildValidationPrompt()
**File**: `internal/validator/prompt.go`

**Complete implementation**:
```go
package validator

import "fmt"

func BuildValidationPrompt(input string) string {
	return fmt.Sprintf(`You are a senior technical writer and educator. Analyze the following user request for a tutorial:

"%s"

Determine if this request is specific and clear enough to create a useful tutorial, or if it's too ambiguous and needs clarification.

Respond ONLY with a valid JSON object (no markdown, no extra text) in this exact format:

If the request is specific enough:
{
  "action": "continue",
  "reason": "brief explanation of your decision",
  "suggested_filename": "short-descriptive-name.md"
}

If the request is too ambiguous:
{
  "action": "exit",
  "reason": "brief explanation of your decision",
  "suggestions": ["clarifying question 1", "clarifying question 2"]
}

Guidelines for suggested_filename:
- Use lowercase with hyphens (kebab-case)
- Keep it short but descriptive (2-5 words max)
- Must end with .md
- Examples: "ffmpeg-video-conversion.md", "pandas-csv-guide.md", "echo-command-basics.md"

Examples of decisions:
- "how does ffmpeg work" -> too broad, suggest: "how to convert video formats", "how to extract audio", etc.
- "how to convert mp4 to webm using ffmpeg" -> specific enough, continue, suggest: "ffmpeg-mp4-to-webm.md"
- "explain python" -> too vague, suggest: "which aspect of Python", "what's your experience level", etc.
- "how to read a CSV file in Python using pandas" -> specific enough, continue, suggest: "pandas-csv-reading.md"
- "docker" -> too vague, suggest: "docker basics", "docker compose", "dockerfile best practices", etc.
- "echo command" -> specific enough, continue, suggest: "echo-command-guide.md"

Think carefully about whether the request has enough context and specificity to create a useful, focused tutorial.`, input)
}
```

**Key Points**:
- Uses `fmt.Sprintf` for string interpolation
- Matches the Python version's prompt exactly
- Includes examples to guide the AI model

#### Step 3.2: Implement validator.New() and Validate()
**File**: `internal/validator/validator.go`

**Complete implementation**:
```go
package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"module/tutor/internal/agent"
)

type Validator struct {
	agent  agent.Runner
	logger *slog.Logger
}

func New(agent agent.Runner) *Validator {
	return &Validator{
		agent: agent,
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}
}

func (v *Validator) Validate(ctx context.Context, input string) (*ValidationResponse, error) {
	// Build validation prompt
	prompt := BuildValidationPrompt(input)

	v.logger.Debug("validating input", "input", input)

	// Call agent
	response, err := v.agent.Run(ctx, "openai/gpt-5.2", prompt)
	if err != nil {
		return nil, fmt.Errorf("agent call failed: %w", err)
	}

	v.logger.Debug("received validation response", "response", response)

	// Parse JSON response
	var result ValidationResponse
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, &ValidationError{
			Input:    input,
			Response: response,
			Err:      fmt.Errorf("%w: %v", ErrInvalidJSON, err),
		}
	}

	// Validate required fields
	if result.Action == "" || result.Reason == "" {
		return nil, &ValidationError{
			Input:    input,
			Response: response,
			Err:      ErrMissingField,
		}
	}

	// Validate action value
	if !result.Action.IsValid() {
		return nil, &ValidationError{
			Input:    input,
			Response: response,
			Err:      ErrInvalidAction,
		}
	}

	// Validate action-specific fields
	if result.Action == ActionContinue && result.SuggestedFilename == "" {
		return nil, &ValidationError{
			Input:    input,
			Response: response,
			Err:      fmt.Errorf("%w: suggested_filename required for continue action", ErrMissingField),
		}
	}

	v.logger.Info("validation complete",
		"action", result.Action,
		"filename", result.SuggestedFilename)

	return &result, nil
}
```

**Key Points**:
- Uses `encoding/json` for JSON parsing
- Returns custom `ValidationError` with context
- Uses sentinel errors for common failure cases
- Validates all required fields based on action type
- Logs at appropriate levels (Debug for details, Info for results)

**Run tests**:
```bash
make test
# Should see validator tests pass with the mock agent
```

---

### Phase 4: Implement Generator Package (30 minutes)

#### Step 4.1: Create generator/prompt.go
**File**: `internal/generator/prompt.go` (create new file)

```go
package generator

import "fmt"

func BuildTutorialPrompt(input string) string {
	return fmt.Sprintf(`You are a senior coder. Write a tutorial to answer the user question, as detailed as you can. The response must be valid markdown.

User input:
%s`, input)
}
```

#### Step 4.2: Implement generator.New() and Generate()
**File**: `internal/generator/generator.go`

**Complete implementation**:
```go
package generator

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"module/tutor/internal/agent"
)

type Generator struct {
	agent  agent.Runner
	logger *slog.Logger
}

func New(agent agent.Runner) *Generator {
	return &Generator{
		agent: agent,
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}
}

func (g *Generator) Generate(ctx context.Context, input string) (string, error) {
	// Build tutorial generation prompt
	prompt := BuildTutorialPrompt(input)

	g.logger.Debug("generating tutorial", "input", input)

	// Call agent
	content, err := g.agent.Run(ctx, "openai/gpt-5.2", prompt)
	if err != nil {
		return "", fmt.Errorf("agent call failed: %w", err)
	}

	// Validate content is not empty
	if content == "" {
		return "", fmt.Errorf("agent returned empty content")
	}

	g.logger.Info("tutorial generated",
		"content_length", len(content))

	return content, nil
}
```

**Key Points**:
- Simple implementation - just calls agent with tutorial prompt
- Validates output is non-empty
- Returns raw markdown content

---

### Phase 5: Implement FileWriter Package (45 minutes)

#### Step 5.1: Implement WriteAtomic()
**File**: `internal/filewriter/filewriter.go`

**Complete implementation**:
```go
package filewriter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteAtomic writes content to a file atomically using a temp file and rename.
// Returns the absolute path of the written file.
func WriteAtomic(path, content string) (string, error) {
	// Expand ~ to home directory
	expandedPath, err := expandPath(path)
	if err != nil {
		return "", fmt.Errorf("expand path: %w", err)
	}

	// Get absolute path
	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		return "", fmt.Errorf("get absolute path: %w", err)
	}

	// Create parent directories if needed
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create parent directories: %w", err)
	}

	// Ensure content ends with newline
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	// Create temporary file in same directory as target
	tmpFile, err := os.CreateTemp(dir, fmt.Sprintf(".%s.*.tmp", filepath.Base(absPath)))
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Write content to temp file
	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath) // Clean up
		return "", fmt.Errorf("write to temp file: %w", err)
	}

	// Close temp file
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath) // Clean up
		return "", fmt.Errorf("close temp file: %w", err)
	}

	// Atomically rename temp file to target
	if err := os.Rename(tmpPath, absPath); err != nil {
		os.Remove(tmpPath) // Clean up
		return "", fmt.Errorf("rename temp file: %w", err)
	}

	return absPath, nil
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	if path == "~" {
		return home, nil
	}

	// ~/something -> $HOME/something
	return filepath.Join(home, path[2:]), nil
}
```

**Key Points**:
- Uses `os.CreateTemp` to create temp file in same directory
- Atomicity via `os.Rename` (atomic on POSIX systems)
- Cleans up temp file on any error
- Expands `~` to home directory
- Ensures content ends with newline
- Returns absolute path

**Test**:
```bash
# Create simple test
cat > test_filewriter.go <<'EOF'
package main

import (
	"fmt"
	"log"
	"os"

	"module/tutor/internal/filewriter"
)

func main() {
	path, err := filewriter.WriteAtomic("test.md", "# Hello\n\nWorld")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Written to:", path)

	content, _ := os.ReadFile(path)
	fmt.Println("Content:", string(content))

	os.Remove(path)
}
EOF

go run test_filewriter.go
rm test_filewriter.go
```

---

### Phase 6: Implement CLI (60 minutes)

#### Step 6.1: Implement run() function
**File**: `cmd/tutor/main.go`

**Complete implementation**:
```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"module/tutor/internal/agent"
	"module/tutor/internal/filewriter"
	"module/tutor/internal/generator"
	"module/tutor/internal/validator"
)

// Config holds command-line arguments
type Config struct {
	Question []string
	Verbose  bool
	Model    string
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

	// Create agent
	agentRunner := agent.New(config.Verbose)

	// Validate input
	fmt.Fprintln(os.Stderr, "Validating input...")
	v := validator.New(agentRunner)
	validationResult, err := v.Validate(ctx, userInput)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if we should exit
	if validationResult.Action == validator.ActionExit {
		fmt.Fprintf(os.Stderr, "\n%s\n", validationResult.Reason)
		if len(validationResult.Suggestions) > 0 {
			fmt.Fprintln(os.Stderr, "\nSuggestions:")
			for _, suggestion := range validationResult.Suggestions {
				fmt.Fprintf(os.Stderr, "  - %s\n", suggestion)
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
	gen := generator.New(agentRunner)
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
		Model:   "openai/gpt-5.2",
		Verbose: false,
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "-v" || arg == "--verbose" {
			config.Verbose = true
		} else if arg == "-m" || arg == "--model" {
			if i+1 < len(args) {
				config.Model = args[i+1]
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
	fmt.Println("  -v, --verbose    Show verbose agent output")
	fmt.Println("  -m, --model STR  Model to use (default: openai/gpt-5.2)")
	fmt.Println("  -h, --help       Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  tutor how to use docker compose")
	fmt.Println("  tutor -v how does grep work")
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
```

**Key Points**:
- Simple flag parsing (no external dependencies)
- Matches Python version's behavior exactly
- Proper error handling with context
- User-friendly prompts and messages
- Supports verbose mode
- Graceful exit on validation failure

**Verify compilation**:
```bash
go build -o bin/tutor ./cmd/tutor
```

---

### Phase 7: Testing and Verification (30 minutes)

#### Step 7.1: Run unit tests
```bash
make test
```

**Expected output**:
- validator tests should pass (with mock agent)
- All other packages should compile without errors

#### Step 7.2: Manual end-to-end test
```bash
# Build
make build

# Test with a good question
./bin/tutor how to use grep command

# Test with a vague question
./bin/tutor -v grep

# Test with CLI args
./bin/tutor -v how to convert mp4 to webm using ffmpeg
```

#### Step 7.3: Compare with Python version
```bash
# Run Python version
python3 cmd/py/tutor.py how to use grep command

# Run Go version
./bin/tutor how to use grep command

# Output should be similar (both validations, both generations)
```

---

## Common Issues and Solutions

### Issue 1: "opencode: command not found"
**Solution**: Ensure `opencode` is installed and in PATH:
```bash
which opencode
# If not found, install it or update PATH
```

### Issue 2: Agent timeout
**Solution**: Increase timeout in main.go:
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
```

### Issue 3: JSON parse errors
**Cause**: Agent returns markdown-wrapped JSON like:
```
```json
{"action": "continue"}
```
```

**Solution**: Add JSON extraction in validator.Validate() before parsing:
```go
// Strip markdown code blocks if present
response = strings.TrimPrefix(response, "```json\n")
response = strings.TrimSuffix(response, "\n```")
response = strings.TrimSpace(response)
```

### Issue 4: Permission denied writing file
**Solution**: Check parent directory permissions or use different output path

---

## Next Steps After Implementation

### 1. Add More Tests
```bash
# Create test for agent package
touch internal/agent/agent_test.go

# Create test for generator package
touch internal/generator/generator_test.go

# Create test for filewriter package
touch internal/filewriter/filewriter_test.go
```

### 2. Improve CLI with Cobra
```bash
go get github.com/spf13/cobra
# Rewrite cmd/tutor/main.go to use cobra
```

### 3. Add Configuration File Support
```go
// Support ~/.tutor/config.yaml
type Config struct {
    Model   string `yaml:"model"`
    Verbose bool   `yaml:"verbose"`
}
```

### 4. Add Progress Indicators
```bash
go get github.com/briandowns/spinner
# Add spinner during generation
```

### 5. Setup CI/CD
Create `.github/workflows/test.yml`:
```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - run: make test
      - run: make build
```

---

## Success Criteria

- [ ] All files compile without errors
- [ ] `make test` passes all tests
- [ ] `make build` produces working binary
- [ ] Manual test with good question generates tutorial
- [ ] Manual test with vague question shows suggestions and exits
- [ ] Verbose mode shows agent output
- [ ] Files are written atomically
- [ ] Behavior matches Python version

---

## Estimated Timeline

| Phase | Task | Time |
|-------|------|------|
| 1 | Fix compilation errors | 15 min |
| 2 | Implement agent package | 45 min |
| 3 | Implement validator package | 60 min |
| 4 | Implement generator package | 30 min |
| 5 | Implement filewriter package | 45 min |
| 6 | Implement CLI | 60 min |
| 7 | Testing and verification | 30 min |
| **Total** | | **4.5 hours** |

---

## Implementation Order Summary

```
1. Fix errors (validator/types.go, errors.go, test files)
   ↓
2. Implement agent.Agent (foundation)
   ↓
3. Implement validator.BuildValidationPrompt()
   ↓
4. Implement validator.Validate()
   ↓
5. Implement generator.Generate()
   ↓
6. Implement filewriter.WriteAtomic()
   ↓
7. Implement cmd/tutor/main.go run()
   ↓
8. Test and verify
```

Each step builds on the previous, allowing you to test incrementally as you go.
