# Tool Ideas

This document contains ideas for future CLI tools to build using the existing agent infrastructure.

## Current Tools

- ✅ **tutor** - Generates technical tutorials from questions
- ✅ **gitmessage** - Generates git commit messages from staged diffs
- ✅ **explain** - Explains code and saves as markdown files

---

## Next Tool: Pull Request Description Generator

### `pr` or `prdesc`

**Why it's the natural next step:** Completes the git workflow after `gitmessage`.

#### Workflow:
```bash
# Stage changes
git add .

# Generate commit message
gitmessage feat -c

# Commit
git commit -m "$(pbpaste)"

# When ready for PR...
pr --base main
# → Analyzes all commits + full diff since branching
# → Generates title + description with bullet points
# → Optional: creates the PR directly via `gh pr create`
```

#### Features:
- Compare current branch vs base (main/develop)
- Analyze all commit messages in the branch
- Look at the full diff
- Generate:
  - **Title** (conventional commit style)
  - **Summary** (what changed and why)
  - **Bullet points** (per-commit or per-area)
  - **Test plan** suggestions (optional)
  - **Breaking changes** section (if applicable)

#### Usage Examples:
```bash
pr                          # Compare current branch to main
pr --base develop           # Compare to develop
pr --base main -c           # Copy to clipboard
pr --create                 # Auto-create PR via gh CLI
pr --template conventional  # Use conventional commit format
```

---

## Other Strong Candidates

### 1. `explain` - Code Explainer

Explains what code does in plain English.

#### Use Cases:
```bash
pbpaste | explain              # From clipboard
explain ./internal/agent/      # From file/directory
explain main.go:45-67          # Specific line range
cat complex.go | explain -v    # Verbose explanation
```

#### Features:
- Explains code logic and patterns
- Identifies design patterns
- Highlights potential issues
- Suggests improvements (optional flag)
- Supports multiple languages

#### Output:
- High-level summary
- Step-by-step breakdown
- Key concepts used
- Dependencies and interactions

---

### 2. `review` - Code Review Assistant

AI-powered code review before committing or for PRs.

#### Use Cases:
```bash
review --staged              # Review staged changes
review                       # Review all uncommitted changes
review PR-123                # Review a GitHub PR
review --security            # Focus on security issues
review --performance         # Focus on performance
```

#### Features:
- Checks for common mistakes
- Suggests improvements
- Identifies security issues
- Checks against style guide
- Performance optimization suggestions

#### Output Format:
- **Issues**: Critical problems
- **Suggestions**: Improvements
- **Nitpicks**: Style/minor issues
- **Good**: Positive feedback

---

### 3. `changelog` - Changelog Generator

Generates CHANGELOG.md entries from git history.

#### Use Cases:
```bash
changelog v1.2.0..HEAD       # Since last tag
changelog --unreleased       # All unreleased changes
changelog --format keep      # Keep-a-Changelog format
changelog --output append    # Append to CHANGELOG.md
```

#### Features:
- Groups by conventional commit type (feat, fix, etc.)
- Extracts breaking changes
- Links to commits/PRs
- Filters merge commits
- Multiple format support (Keep-a-Changelog, semantic-release)

---

### 4. `readme` - README Generator

Analyzes project and generates/updates README.md.

#### Use Cases:
```bash
readme --init               # Generate new README
readme --update             # Update existing README
readme --section install    # Update installation section only
```

#### Features:
- Detects project type (Go, Python, Node, etc.)
- Extracts description from code/comments
- Generates installation instructions
- Creates usage examples from CLI help
- API documentation from code
- Badges for CI, coverage, etc.

---

### 5. `docgen` - Documentation Generator

Generates documentation from code comments and structure.

#### Use Cases:
```bash
docgen ./internal/          # Generate docs for package
docgen --format markdown    # Output as markdown
docgen --format html        # Output as HTML
```

#### Features:
- Extracts function/type documentation
- Generates API reference
- Creates architecture diagrams (ASCII/Mermaid)
- Usage examples
- Cross-references

---

### 6. `test` - Test Generator

Generates test cases from code.

#### Use Cases:
```bash
test generate agent.go           # Generate tests
test generate --table-driven     # Use table-driven style
test generate --coverage-gaps    # Only missing coverage
```

#### Features:
- Analyzes function signatures
- Generates table-driven tests
- Creates mock examples
- Edge case identification
- Test data generation

---

### 7. `refactor` - Refactoring Suggester

Analyzes code and suggests refactoring improvements.

#### Use Cases:
```bash
refactor ./cmd/tutor/main.go    # Analyze file
refactor --type extract-func    # Suggest function extractions
refactor --type simplify        # Suggest simplifications
```

#### Features:
- Detects code smells
- Suggests pattern improvements
- Identifies duplicate code
- Recommends abstractions
- Performance improvements

---

### 8. `deps` - Dependency Analyzer

Analyzes and explains dependencies.

#### Use Cases:
```bash
deps why github.com/pkg/errors   # Why is this imported?
deps tree                        # Show dependency tree
deps unused                      # Find unused dependencies
deps audit                       # Security audit
```

---

### 9. `migrate` - Code Migration Assistant

Helps migrate code between versions/languages.

#### Use Cases:
```bash
migrate --from python --to go script.py
migrate --go-version 1.20-to-1.23
migrate --lint-fix ./...
```

---

### 10. `snippet` - Code Snippet Generator

Generates common code patterns.

#### Use Cases:
```bash
snippet http-handler --lang go       # HTTP handler boilerplate
snippet crud-repo --db postgres      # CRUD repository
snippet cli-cobra                    # Cobra CLI app
```

---

## Implementation Priority

### High Priority (Complete the workflow):
1. **pr** - Pull request description generator
2. **explain** - Code explainer
3. **review** - Code review assistant

### Medium Priority (High value, standalone):
4. **changelog** - Changelog generator
5. **readme** - README generator
6. **test** - Test generator

### Lower Priority (Nice to have):
7. **refactor** - Refactoring suggester
8. **docgen** - Documentation generator
9. **deps** - Dependency analyzer
10. **migrate** - Migration assistant
11. **snippet** - Snippet generator

---

## Common Infrastructure Needs

All tools can leverage:
- ✅ `internal/agent` - OpenCode integration
- ✅ `internal/filewriter` - Atomic file writing
- ⚠️ Need: Git operations helper (`internal/git`)
- ⚠️ Need: GitHub API helper (`internal/github`)
- ⚠️ Need: Code parser/AST helper (`internal/codeparse`)

---

## Decision Criteria

When choosing the next tool, consider:
1. **Workflow integration** - Does it fit the existing flow?
2. **Reusable infrastructure** - Can we leverage existing code?
3. **User value** - How often would developers use it?
4. **Complexity** - How much new code is needed?
5. **Uniqueness** - Are there good existing tools?

**Recommendation:** Start with `pr` as it completes the commit → PR workflow and reuses git operations similar to `gitmessage`.
