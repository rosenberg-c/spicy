# Comprehensive Testing Plan for Spicy Project

## Executive Summary

This testing plan provides a systematic approach to achieving comprehensive test coverage for the Spicy project, a collection of AI-powered CLI tools written in Go. The plan prioritizes testability through dependency injection and real implementations over mocks (per Rule #7), follows the existing table-driven test pattern, and ensures critical business logic is thoroughly tested.

## 1. Test Priorities (Recommended Order)

### Phase 1: Foundation (High Priority)
1. **internal/filewriter** - Critical file operations, high risk
2. **internal/agent** - Core AI interaction, shared by all tools
3. **cmd/tutor/validator** - Expand existing tests

### Phase 2: CLI Tools (Medium Priority)
4. **cmd/ask** - Simplest CLI, good template for others
5. **cmd/gitmessage** - Git integration testing patterns
6. **cmd/explain** - Complex input handling (file/directory/stdin)
7. **cmd/tutor** - Most complex, interactive inputs

### Phase 3: Low Priority
8. **internal/constants** - Simple constants, low value

## 2. Test Strategy by Component

### 2.1 internal/filewriter

**Test file location**: `internal/filewriter/filewriter_test.go`

**Strategy**: Use real filesystem with temp directories (per Rule #7)

**Critical test scenarios**:
- Path expansion (~ to home directory)
- Absolute path resolution
- Parent directory creation
- Atomic write operation (temp file + rename)
- Newline appending behavior
- Error handling for invalid paths
- Concurrent writes to same file
- Permission errors

**Mock vs Real**: Use `t.TempDir()` for real filesystem operations

**Example test structure**:
```go
func TestWriteAtomic(t *testing.T) {
    tests := []struct {
        name        string
        path        string
        content     string
        wantContent string
        wantErr     bool
        errContains string
    }{
        {
            name:        "simple write",
            path:        "test.txt",
            content:     "hello",
            wantContent: "hello\n",
            wantErr:     false,
        },
        {
            name:        "content already has newline",
            path:        "test.txt",
            content:     "hello\n",
            wantContent: "hello\n",
            wantErr:     false,
        },
        {
            name:        "nested directory creation",
            path:        "a/b/c/test.txt",
            content:     "nested",
            wantContent: "nested\n",
            wantErr:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tempDir := t.TempDir()
            testPath := filepath.Join(tempDir, tt.path)

            gotPath, err := WriteAtomic(testPath, tt.content)

            if (err != nil) != tt.wantErr {
                t.Errorf("WriteAtomic() error = %v, wantErr %v",
                    err, tt.wantErr)
                return
            }

            if tt.wantErr {
                if tt.errContains != "" &&
                   !strings.Contains(err.Error(), tt.errContains) {
                    t.Errorf("error = %v, want error containing %q",
                        err, tt.errContains)
                }
                return
            }

            // Verify file exists and has correct content
            got, err := os.ReadFile(gotPath)
            if err != nil {
                t.Fatalf("failed to read result: %v", err)
            }

            if string(got) != tt.wantContent {
                t.Errorf("content = %q, want %q", got, tt.wantContent)
            }
        })
    }
}
```

### 2.2 internal/agent

**Test file location**: `internal/agent/agent_test.go`

**Strategy**: Mock the external `opencode` command using test doubles

**Critical test scenarios**:
- Successful command execution
- Command timeout via context
- Command failure (non-zero exit)
- Verbose vs non-verbose logging
- stdout capture
- stderr handling
- Model parameter passing
- Prompt parameter passing
- Context cancellation

**Mock vs Real**: Create test implementation that satisfies `agent.Runner`
interface

**Example test structure**:
```go
func TestAgent_Run(t *testing.T) {
    tests := []struct {
        name          string
        verbose       bool
        model         string
        prompt        string
        mockExecFunc  func(ctx context.Context, cmd string,
            args ...string) ([]byte, error)
        want          string
        wantErr       bool
        errContains   string
        contextCancel bool
    }{
        {
            name:    "successful execution",
            verbose: false,
            model:   "test-model",
            prompt:  "test prompt",
            mockExecFunc: func(ctx context.Context, cmd string,
                args ...string) ([]byte, error) {
                return []byte("response from AI\n"), nil
            },
            want:    "response from AI",
            wantErr: false,
        },
        {
            name:          "context cancellation",
            verbose:       false,
            model:         "test-model",
            prompt:        "test prompt",
            contextCancel: true,
            wantErr:       true,
            errContains:   "context",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            if tt.contextCancel {
                var cancel context.CancelFunc
                ctx, cancel = context.WithCancel(ctx)
                cancel()
            }

            mockRunner := &mockAgentRunner{
                execFunc: tt.mockExecFunc,
            }

            got, err := mockRunner.Run(ctx, tt.model, tt.prompt)

            if (err != nil) != tt.wantErr {
                t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if tt.wantErr && tt.errContains != "" {
                if !strings.Contains(err.Error(), tt.errContains) {
                    t.Errorf("error = %v, want error containing %q",
                        err, tt.errContains)
                }
                return
            }

            if got != tt.want {
                t.Errorf("Run() = %q, want %q", got, tt.want)
            }
        })
    }
}

// Mock implementation for testing
type mockAgentRunner struct {
    execFunc func(ctx context.Context, cmd string,
        args ...string) ([]byte, error)
}

func (m *mockAgentRunner) Run(ctx context.Context, model,
    prompt string) (string, error) {
    if m.execFunc != nil {
        output, err := m.execFunc(ctx, "opencode", "run",
            "--agent", "build", "-m", model, prompt)
        if err != nil {
            return "", fmt.Errorf("agent command failed: %w", err)
        }
        return strings.TrimSpace(string(output)), nil
    }
    return "", fmt.Errorf("no exec function configured")
}
```

### 2.3 cmd/tutor/validator

**Test file location**: `cmd/tutor/validator/validator_test.go` (expand existing)

**Additional test scenarios to add**:
- Invalid JSON response
- Missing required fields
- Invalid action value
- Context timeout
- Agent error propagation
- Nil suggestions vs empty array
- Filename validation (kebab-case format)

### 2.4 cmd/ask

**Test file location**: `cmd/ask/main_test.go`

**Test helper functions**:
- `getUserInput()` - args vs stdin
- `buildPrompt()` - prompt generation

**Example test structure**:
```go
func TestGetUserInput(t *testing.T) {
    tests := []struct {
        name    string
        args    []string
        want    string
        wantErr bool
    }{
        {
            name:    "multiple args joined",
            args:    []string{"how", "does", "git", "work"},
            want:    "how does git work",
            wantErr: false,
        },
        {
            name:    "single arg",
            args:    []string{"hello"},
            want:    "hello",
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := getUserInput(tt.args)
            if (err != nil) != tt.wantErr {
                t.Errorf("getUserInput() error = %v, wantErr %v",
                    err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("getUserInput() = %q, want %q", got, tt.want)
            }
        })
    }
}

func TestBuildPrompt(t *testing.T) {
    tests := []struct {
        name         string
        input        string
        wantContains []string
    }{
        {
            name:  "includes user input",
            input: "explain closures",
            wantContains: []string{
                "explain closures",
                "senior coder",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := buildPrompt(tt.input)
            for _, want := range tt.wantContains {
                if !strings.Contains(got, want) {
                    t.Errorf("buildPrompt() missing %q", want)
                }
            }
        })
    }
}
```

### 2.5 cmd/gitmessage

**Test file location**: `cmd/gitmessage/main_test.go`

**Strategy**: Test git command execution separately from AI generation

**Critical test scenarios**:
- Successful message generation with staged changes
- No staged changes warning
- Git diff error handling
- Hint flag integration
- Prefix handling
- Model selection

### 2.6 cmd/explain

**Test file location**: `cmd/explain/main_test.go`

**Strategy**: Test complex input handling

**Critical test scenarios**:
- File input
- Directory input (recursive code gathering)
- Stdin input
- Language detection from extension
- Language detection from content
- Manual language override
- Directory skipping (node_modules, .git, etc.)

**Key functions to test**:
- `getCodeInput()` - file/dir/stdin handling
- `readDirectory()` - recursive file gathering
- `detectLanguage()` - extension and content detection
- `suggestFilename()` - filename generation
- `addLineNumbers()` - code formatting

### 2.7 cmd/tutor

**Test file location**: `cmd/tutor/main_test.go`

**Critical test scenarios**:
- Model flag handling (base, validation-model, generation-model)
- User input from args
- Prompt building
- Integration with validator

## 3. Mock vs Real Implementation Decisions

Following Rule #7 ("Keep repositories testable"):

| Component | Strategy | Rationale |
|-----------|----------|-----------|
| **filewriter** | Real filesystem with `t.TempDir()` | Safe, isolated |
| **agent** | Mock via interface | External command |
| **validator** | Mock agent dependency | Existing pattern |
| **Helper functions** | Real implementations | Pure functions |
| **Git commands** | Real git in temp repos | Can isolate |
| **User input** | Mock/skip | Terminal interaction |

## 4. File Organization

```
spicy/
├── internal/
│   ├── agent/
│   │   ├── agent.go
│   │   ├── agent_test.go          # NEW
│   │   └── types.go
│   ├── filewriter/
│   │   ├── filewriter.go
│   │   └── filewriter_test.go     # NEW
│   └── constants/
│       ├── constants.go
│       └── constants_test.go      # NEW
├── cmd/
│   ├── tutor/
│   │   ├── main.go
│   │   ├── main_test.go           # NEW
│   │   └── validator/
│   │       ├── validator_test.go  # EXPAND
│   │       └── ...
│   ├── ask/
│   │   ├── main.go
│   │   └── main_test.go           # NEW
│   ├── explain/
│   │   ├── main.go
│   │   └── main_test.go           # NEW
│   └── gitmessage/
│       ├── main.go
│       └── main_test.go           # NEW
```

## 5. Test Coverage Goals

### Minimum Acceptable Coverage
- **Overall**: 70%+ code coverage
- **Critical paths**: 90%+ (filewriter, validator, agent interface)
- **Helper functions**: 80%+

### Coverage by Component
| Component | Target | Priority |
|-----------|--------|----------|
| internal/filewriter | 95% | High |
| internal/agent | 85% | High |
| cmd/tutor/validator | 90% | High |
| cmd/ask | 70% | Medium |
| cmd/gitmessage | 70% | Medium |
| cmd/explain | 75% | Medium |
| cmd/tutor | 65% | Medium |

## 6. Test Helper Template

```go
// Helper to create temp directory with files
func createTestFiles(t *testing.T, files map[string]string) string {
    t.Helper()
    tempDir := t.TempDir()

    for path, content := range files {
        fullPath := filepath.Join(tempDir, path)
        if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
            t.Fatalf("create dir: %v", err)
        }
        if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
            t.Fatalf("write file: %v", err)
        }
    }

    return tempDir
}

// Helper to initialize git repo for testing
func initTestGitRepo(t *testing.T, dir string) {
    t.Helper()

    cmd := exec.Command("git", "init")
    cmd.Dir = dir
    if err := cmd.Run(); err != nil {
        t.Fatalf("git init: %v", err)
    }

    cmd = exec.Command("git", "config", "user.email", "test@example.com")
    cmd.Dir = dir
    cmd.Run()

    cmd = exec.Command("git", "config", "user.name", "Test User")
    cmd.Dir = dir
    cmd.Run()
}
```

## 7. Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run only fast tests (skip integration)
go test -short ./...

# Run specific package
go test ./internal/filewriter/

# Verbose output
go test -v ./...

# Run specific test
go test -run TestWriteAtomic ./internal/filewriter/
```

## 8. Implementation Sequence

### Week 1: Foundation
1. `internal/filewriter/filewriter_test.go` - Critical path
2. `internal/agent/agent_test.go` - Core interface
3. Expand `cmd/tutor/validator/validator_test.go`

### Week 2: Simple CLI Tools
4. `cmd/ask/main_test.go` - Template for others
5. `cmd/tutor/validator/types_test.go`
6. `cmd/tutor/validator/errors_test.go`

### Week 3: Complex CLI Tools
7. `cmd/gitmessage/main_test.go` - Git integration
8. `cmd/explain/main_test.go` - Complex input

### Week 4: Integration and Coverage
9. `cmd/tutor/main_test.go` - Most complex
10. `internal/constants/constants_test.go`
11. Coverage analysis and gap filling

## 9. Testing Best Practices

1. **Follow existing pattern**: Use table-driven tests
2. **Use real implementations**: Per Rule #7, prefer `t.TempDir()`
3. **Test error paths**: All error returns should be exercised
4. **Context testing**: Test timeout behavior
5. **Isolation**: Each test should be independent
6. **Helper functions**: Extract common setup with `t.Helper()`
7. **Short flag**: Use `testing.Short()` for integration tests
8. **Clear names**: Describe the scenario
9. **Subtests**: Always use `t.Run()`
10. **Line length**: Keep under 80 characters per Rule #11

## 10. Success Metrics

- [ ] All Phase 1 components have >85% coverage
- [ ] All exported functions have at least one test
- [ ] All error paths are tested
- [ ] `make test` runs successfully
- [ ] Coverage report shows no critical gaps
- [ ] New code includes tests
