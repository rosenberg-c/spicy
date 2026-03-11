# Tutorial Generator - Go Rewrite Plan

## Overview
Rewrite the Python tutorial generator in idiomatic Go, following standard Go conventions and project layout patterns.

## Project Structure

```
tutorial-generator/
├── cmd/
│   └── tutorial/
│       └── main.go              # CLI entry point and orchestration
├── internal/
│   ├── agent/
│   │   ├── agent.go             # Agent communication interface and implementation
│   │   └── types.go             # Agent-specific types
│   ├── validator/
│   │   ├── validator.go         # Input validation logic
│   │   ├── prompt.go            # Validation prompt builder
│   │   └── types.go             # ValidationResponse, Action constants
│   ├── generator/
│   │   ├── generator.go         # Tutorial generation
│   │   └── prompt.go            # Tutorial prompt builder
│   └── filewriter/
│       └── writer.go            # Atomic file writing operations
├── pkg/
│   └── tutorial/
│       └── types.go             # Public types (if needed for library use)
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Core Type Definitions

### internal/validator/types.go
```go
type Action string

const (
    ActionContinue Action = "continue"
    ActionExit     Action = "exit"
)

type ValidationResponse struct {
    Action            Action   `json:"action"`
    Reason            string   `json:"reason"`
    SuggestedFilename string   `json:"suggested_filename,omitempty"`
    Suggestions       []string `json:"suggestions,omitempty"`
}

func (a Action) String() string {
    return string(a)
}

func (a Action) IsValid() bool {
    return a == ActionContinue || a == ActionExit
}
```

### internal/agent/types.go
```go
type Config struct {
    Model   string
    Verbose bool
}

type Runner interface {
    Run(ctx context.Context, model, prompt string) (string, error)
}
```

## Package Responsibilities

### 1. cmd/tutorial/main.go
**Responsibility**: CLI interface, user interaction, orchestration

**Key functions**:
- `main()` - Entry point
- `run()` - Main execution flow
- CLI setup using `cobra` or `flag`
- Handle user input from args or stdin
- Coordinate validator → generator → file writer flow
- Display results and errors to user

**Dependencies**:
- `internal/agent`
- `internal/validator`
- `internal/generator`
- `internal/filewriter`
- `github.com/spf13/cobra` (recommended)

### 2. internal/agent/agent.go
**Responsibility**: Execute external opencode commands

**Key types**:
```go
type Agent struct {
    verbose bool
    logger  *slog.Logger
}

func New(verbose bool) *Agent
func (a *Agent) Run(ctx context.Context, model, prompt string) (string, error)
```

**Implementation details**:
- Use `os/exec.CommandContext` for subprocess management
- Respect context cancellation
- Redirect stderr based on verbose flag
- Return wrapped errors with context
- Handle command failures gracefully

### 3. internal/validator/validator.go
**Responsibility**: Validate user input specificity

**Key types**:
```go
type Validator struct {
    agent  agent.Runner
    logger *slog.Logger
}

func New(agent agent.Runner) *Validator
func (v *Validator) Validate(ctx context.Context, input string) (*ValidationResponse, error)
```

**Implementation details**:
- Build validation prompt with proper escaping
- Parse JSON response using `encoding/json`
- Validate response structure and fields
- Return typed errors for different failure modes
- Log validation steps when verbose

### 4. internal/validator/prompt.go
**Responsibility**: Build validation prompts

**Key functions**:
```go
func BuildValidationPrompt(input string) string
```

**Implementation details**:
- Use `text/template` or string builder
- Consider embedding prompt template with `embed` package
- Include examples in prompt
- Format JSON schema clearly

### 5. internal/generator/generator.go
**Responsibility**: Generate tutorial content

**Key types**:
```go
type Generator struct {
    agent  agent.Runner
    logger *slog.Logger
}

func New(agent agent.Runner) *Generator
func (g *Generator) Generate(ctx context.Context, input string) (string, error)
```

**Implementation details**:
- Build tutorial prompt
- Call agent with prompt
- Return markdown content
- Handle generation errors

### 6. internal/filewriter/writer.go
**Responsibility**: Atomic file writing

**Key functions**:
```go
func WriteAtomic(path, content string) (string, error)
```

**Implementation details**:
- Create temp file in same directory as target
- Write content to temp file
- Use `os.Rename` for atomic operation
- Create parent directories if needed
- Expand `~` in paths
- Ensure content ends with newline
- Clean up temp file on errors

## Error Handling Strategy

### 1. Custom Error Types
```go
// internal/validator/errors.go
type ValidationError struct {
    Input    string
    Response string
    Err      error
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %q: %v", e.Input, e.Err)
}

func (e *ValidationError) Unwrap() error {
    return e.Err
}

// Sentinel errors
var (
    ErrInvalidAction   = errors.New("invalid action in response")
    ErrMissingField    = errors.New("missing required field")
    ErrInvalidJSON     = errors.New("invalid JSON response")
)
```

### 2. Error Wrapping
Use `fmt.Errorf` with `%w` verb to wrap errors:
```go
if err != nil {
    return nil, fmt.Errorf("spawn agent: %w", err)
}
```

### 3. Error Checking
Use `errors.Is()` and `errors.As()` for error type checking:
```go
if errors.Is(err, validator.ErrInvalidAction) {
    // Handle specific error
}
```

## CLI Design

### Using cobra (recommended)
```go
var rootCmd = &cobra.Command{
    Use:   "tutorial [question...]",
    Short: "Generate technical tutorials using AI",
    Long: `Tutorial generator validates your question for specificity,
suggests improvements if needed, and generates detailed markdown tutorials.`,
    Args: cobra.ArbitraryArgs,
    RunE: run,
}

func init() {
    rootCmd.Flags().BoolP("verbose", "v", false, "Show verbose agent output")
    rootCmd.Flags().StringP("model", "m", "openai/gpt-5.2", "Model to use")
    rootCmd.Flags().StringP("output", "o", "", "Output file path (overrides suggested filename)")
}
```

### Flow
1. Parse CLI flags and args
2. Get question from args or prompt user
3. Create agent with verbose setting
4. Validate input
5. If exit action: display reason and suggestions, exit 0
6. If continue action: prompt for output path (with suggested default)
7. Generate tutorial
8. Write to file atomically
9. Display success message

## Configuration Management

### Option 1: Simple struct
```go
type Config struct {
    Model   string
    Verbose bool
    Output  string
}

func LoadConfig() Config {
    return Config{
        Model:   getEnvOrDefault("TUTORIAL_MODEL", "openai/gpt-5.2"),
        Verbose: false,
    }
}
```

### Option 2: Functional options pattern
```go
type Option func(*Config)

func WithModel(model string) Option {
    return func(c *Config) { c.Model = model }
}

func WithVerbose(verbose bool) Option {
    return func(c *Config) { c.Verbose = verbose }
}

func NewGenerator(opts ...Option) *Generator {
    cfg := &Config{Model: "openai/gpt-5.2"}
    for _, opt := range opts {
        opt(cfg)
    }
    return &Generator{config: cfg}
}
```

## Logging Strategy

Use `log/slog` (Go 1.21+) for structured logging:

```go
func New(verbose bool) *Agent {
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

// Usage
a.logger.Info("validating input", "input", userInput, "model", model)
a.logger.Debug("agent response", "response", response)
```

## Testing Strategy

### 1. Unit Tests
Each package should have comprehensive unit tests:

```go
// internal/validator/validator_test.go
func TestValidator_Validate(t *testing.T) {
    tests := []struct {
        name           string
        input          string
        agentResponse  string
        want           *ValidationResponse
        wantErr        bool
    }{
        {
            name:  "continue action with filename",
            input: "how to use ffmpeg",
            agentResponse: `{
                "action": "continue",
                "reason": "Specific enough",
                "suggested_filename": "ffmpeg-guide.md"
            }`,
            want: &ValidationResponse{
                Action:            ActionContinue,
                Reason:            "Specific enough",
                SuggestedFilename: "ffmpeg-guide.md",
            },
            wantErr: false,
        },
        {
            name:  "exit action with suggestions",
            input: "ffmpeg",
            agentResponse: `{
                "action": "exit",
                "reason": "Too vague",
                "suggestions": ["be more specific"]
            }`,
            want: &ValidationResponse{
                Action:      ActionExit,
                Reason:      "Too vague",
                Suggestions: []string{"be more specific"},
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockAgent := &mockAgentRunner{response: tt.agentResponse}
            v := New(mockAgent)

            got, err := v.Validate(context.Background(), tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Validate() = %v, want %v", got, tt.want)
            }
        })
    }
}

// Mock implementation
type mockAgentRunner struct {
    response string
    err      error
}

func (m *mockAgentRunner) Run(ctx context.Context, model, prompt string) (string, error) {
    if m.err != nil {
        return "", m.err
    }
    return m.response, nil
}
```

### 2. Integration Tests
Test the full flow in cmd package:

```go
// cmd/tutorial/main_test.go
func TestE2E(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Test with mock agent or actual opencode if available
    // Test file creation
    // Test error handling
}
```

### 3. Test Helpers
```go
// internal/testutil/testutil.go
func CreateTempDir(t *testing.T) string {
    t.Helper()
    dir, err := os.MkdirTemp("", "tutorial-test-*")
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { os.RemoveAll(dir) })
    return dir
}
```

## Concurrency Considerations

### Context Usage
Pass `context.Context` to all long-running operations:

```go
func (v *Validator) Validate(ctx context.Context, input string) (*ValidationResponse, error) {
    // Check context before expensive operations
    if err := ctx.Err(); err != nil {
        return nil, err
    }

    response, err := v.agent.Run(ctx, v.config.Model, prompt)
    // ...
}
```

### Timeout Example
```go
func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    if err := run(ctx); err != nil {
        log.Fatal(err)
    }
}
```

### Potential Parallelism
Consider running validation and asking for filename in parallel (if we pre-prompt for filename):
```go
type validationResult struct {
    resp *ValidationResponse
    err  error
}

func run(ctx context.Context) error {
    // Could parallelize independent operations
    errg, ctx := errgroup.WithContext(ctx)

    var validationResp *ValidationResponse
    errg.Go(func() error {
        var err error
        validationResp, err = validator.Validate(ctx, input)
        return err
    })

    // Wait for validation
    if err := errg.Wait(); err != nil {
        return err
    }

    // Continue with generation...
}
```

## Build and Distribution

### Makefile
```makefile
.PHONY: build test clean install

BINARY_NAME=tutorial
INSTALL_PATH=/usr/local/bin

build:
	go build -o bin/$(BINARY_NAME) cmd/tutorial/main.go

test:
	go test -v -race -coverprofile=coverage.out ./...

test-short:
	go test -v -short ./...

coverage:
	go tool cover -html=coverage.out

clean:
	rm -rf bin/ coverage.out

install: build
	cp bin/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)

lint:
	golangci-lint run

fmt:
	go fmt ./...
	gofmt -s -w .

vet:
	go vet ./...
```

### Go Module
```go
// go.mod
module github.com/yourusername/tutorial-generator

go 1.21

require (
    github.com/spf13/cobra v1.8.0
)
```

### goreleaser (optional)
For multi-platform releases:
```yaml
# .goreleaser.yml
before:
  hooks:
    - go mod tidy

builds:
  - main: ./cmd/tutorial
    binary: tutorial
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{ .Version }}_
      {{ .Os }}_
      {{ .Arch }}
```

## Advanced Features to Consider

### 1. Embed Prompt Templates
```go
import _ "embed"

//go:embed prompts/validation.tmpl
var validationPromptTemplate string

func BuildValidationPrompt(input string) (string, error) {
    tmpl, err := template.New("validation").Parse(validationPromptTemplate)
    if err != nil {
        return "", err
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, map[string]string{"Input": input}); err != nil {
        return "", err
    }

    return buf.String(), nil
}
```

### 2. Configuration File Support
```go
// Support ~/.tutorial/config.yaml
type Config struct {
    Model   string `yaml:"model"`
    Verbose bool   `yaml:"verbose"`
}

func LoadConfig() (*Config, error) {
    home, _ := os.UserHomeDir()
    data, err := os.ReadFile(filepath.Join(home, ".tutorial", "config.yaml"))
    if err != nil {
        return defaultConfig(), nil // Use defaults if no config
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

### 3. Progress Indicators
```go
import "github.com/briandowns/spinner"

func (g *Generator) Generate(ctx context.Context, input string) (string, error) {
    s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
    s.Suffix = " Generating tutorial..."
    s.Start()
    defer s.Stop()

    // ... generation logic
}
```

### 4. Retry Logic
```go
func (a *Agent) RunWithRetry(ctx context.Context, model, prompt string, maxRetries int) (string, error) {
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        result, err := a.Run(ctx, model, prompt)
        if err == nil {
            return result, nil
        }

        lastErr = err

        // Exponential backoff
        backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
        select {
        case <-time.After(backoff):
            continue
        case <-ctx.Done():
            return "", ctx.Err()
        }
    }

    return "", fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

## Migration Checklist

- [ ] Set up Go module and project structure
- [ ] Implement core types (Action, ValidationResponse, etc.)
- [ ] Implement agent package with subprocess execution
- [ ] Implement validator package with prompt building and parsing
- [ ] Implement generator package
- [ ] Implement filewriter package with atomic operations
- [ ] Implement CLI with cobra
- [ ] Add comprehensive unit tests (aim for >80% coverage)
- [ ] Add integration tests
- [ ] Set up Makefile for common tasks
- [ ] Add README with installation and usage instructions
- [ ] Add examples and documentation
- [ ] Set up CI/CD (GitHub Actions)
- [ ] Optional: Set up goreleaser for releases
- [ ] Optional: Add shell completion generation

## Key Differences from Python Version

1. **Strong typing**: Structs instead of dicts, typed errors
2. **Explicit error handling**: No exceptions, return errors explicitly
3. **Interfaces**: For testability and dependency injection
4. **Context**: For cancellation and timeouts
5. **Packages**: Better code organization
6. **Built-in concurrency**: Goroutines and channels (if needed)
7. **Compilation**: Single binary, no runtime dependencies
8. **Performance**: Faster startup and execution
9. **Standard library**: More comprehensive, less external dependencies

## Estimated Implementation Time

- Project setup and structure: 1-2 hours
- Core packages implementation: 4-6 hours
- CLI implementation: 2-3 hours
- Testing: 3-4 hours
- Documentation: 1-2 hours
- **Total**: ~15-20 hours for complete implementation

## Next Steps

1. Initialize Go module: `go mod init github.com/username/tutorial-generator`
2. Create directory structure
3. Implement in order: types → agent → validator → generator → filewriter → CLI
4. Test each package as you go
5. Add integration tests
6. Document and polish
