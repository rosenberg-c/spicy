# Spicy - AI-Powered Developer Tools

A collection of AI-powered CLI tools to boost developer productivity.

## Tools

### ✅ tutor - Tutorial Generator
Generate detailed technical tutorials from questions.

```sh
tutor how to use docker compose
tutor -v how does grep work
```

### ✅ gitmessage - Commit Message Generator
Generate git commit messages from staged changes.

```sh
git add .
gitmessage feat -c    # Generate and copy to clipboard
git commit -m "$(pbpaste)"
```

### ✅ explain - Code Explainer
Explain code and save explanations as markdown files.

```sh
explain main.go
explain ./internal/agent/
pbpaste | explain --no-save
cat complex.go | explain -o explanation.md
```

### ✅ ctx-edit - Context Editor
Update a selected code context based on a prompt.

```sh
ctx-edit -p "rename foo to bar" -c "const foo = 1"
ctx-edit -p "add error handling" -f main.go --start 12 --end 24
pbpaste | ctx-edit -p "make this more concise" -c -
ctx-edit -p "convert to for-range" -f main.go --start 10 --end 18 --write
```

### ✅ history - History Manager
Browse and export command history to markdown files.

All commands support `--history` flag to save execution history to `.spicy/`
directory with format: `YYYYMMDD-HHMMSS_cmd_description.json`

```sh
# Enable history saving
ask --history "What is Docker?"
tutor --history "How to use git rebase"
gitmessage --history

# Browse and export history
shistory list                          # List all history entries
shistory list --command ask            # List history for specific command
shistory export                        # Interactive export to markdown
shistory export --file .spicy/ask/20260317-134703_ask_what-is-docker.json
```

## Installation

### Build all tools
```sh
make build-all
```

### Install to ~/.local/bin
```sh
make install-all
```

### Build individual tools
```sh
make build-tutor
make build-gitmessage
make build-explain
```

## Requirements

- Go 1.23+
- [opencode](https://github.com/anthropics/anthropic-opencode) (for AI agent integration)

## Project Structure

```
spicy/
├── cmd/
│   ├── ask/         # AI question answerer
│   ├── tutor/       # Tutorial generator
│   ├── gitmessage/  # Commit message generator
│   ├── explain/     # Code explainer
│   ├── ctx-edit/    # Context editor
│   └── history/     # History manager
├── internal/
│   ├── agent/       # OpenCode agent integration
│   ├── validator/   # Input validation
│   ├── generator/   # Content generation
│   └── filewriter/  # Atomic file writing
├── docs/
│   └── 01_RULES.md  # Go coding standards
├── tutorials/       # Example generated tutorials
├── IDEAS.md         # Future tool ideas
└── Makefile         # Build system
```

## Development

### Format code
```sh
make fmt
```

### Run tests
```sh
make test
```

### Check (vet + test)
```sh
make check
```

## See Also

- [IDEAS.md](IDEAS.md) - Future tool ideas
- [docs/01_RULES.md](docs/01_RULES.md) - Go coding standards
- [tutorials/](tutorials/) - Example generated content
