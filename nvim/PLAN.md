# Spicy.nvim - Implementation Plan

**Version**: 1.0
**Last Updated**: 2026-03-15
**Status**: Planning Phase

## Executive Summary

A comprehensive Neovim plugin providing seamless integration with the spicy CLI tools (ask, tutor, explain, gitmessage). The plugin will offer multiple interaction patterns, rich UI components, and extensive configurability while maintaining the Unix philosophy of composability.

---

## 1. Design Philosophy

### Core Principles

1. **Minimal by default, powerful when needed**: Work out of the box with zero config, but allow deep customization
2. **Async-first**: Never block the editor, all CLI interactions are asynchronous
3. **Context-aware**: Understand what the user is working on (language, file type, git state)
4. **Keyboard-driven**: Optimize for vim users who rarely use mouse
5. **Composable**: Work well with existing plugins (telescope, which-key, lualine)
6. **Fail gracefully**: Clear error messages, health checks, auth validation

### Non-Goals

- Don't reimplement the CLI logic in Lua
- Don't add AI features not in the CLI
- Don't force specific keybindings
- Don't require heavy dependencies (no treesitter, no LSP)

---

## 2. Architecture Overview

### Module Structure

```
spicy.nvim/
├── lua/
│   └── spicy/
│       ├── init.lua              # Main entry, setup()
│       ├── config.lua            # Configuration management
│       ├── health.lua            # :checkhealth spicy
│       ├── commands/
│       │   ├── init.lua          # Command registry
│       │   ├── ask.lua           # SpicyAsk implementation
│       │   ├── tutor.lua         # SpicyTutor implementation
│       │   ├── explain.lua       # SpicyExplain implementation
│       │   └── gitmessage.lua    # SpicyGitmessage implementation
│       ├── ui/
│       │   ├── float.lua         # Floating window utilities
│       │   ├── buffer.lua        # Buffer management
│       │   ├── input.lua         # Input prompts
│       │   └── spinner.lua       # Loading indicators
│       ├── utils/
│       │   ├── job.lua           # Async job wrapper (plenary)
│       │   ├── git.lua           # Git utilities
│       │   ├── fs.lua            # File system helpers
│       │   └── helpers.lua       # General utilities
│       └── telescope/
│           └── spicy.lua         # Telescope extensions (optional)
├── plugin/
│   └── spicy.vim                 # Vim commands definition
├── doc/
│   └── spicy.nvim.txt            # Vimdoc help
├── tests/
│   ├── spicy_spec.lua
│   ├── commands/
│   │   ├── ask_spec.lua
│   │   ├── tutor_spec.lua
│   │   ├── explain_spec.lua
│   │   └── gitmessage_spec.lua
│   └── minimal_init.lua          # Minimal config for testing
├── .github/
│   └── workflows/
│       ├── test.yml
│       └── lint.yml
├── README.md
├── CHANGELOG.md
├── LICENSE
└── Makefile
```

---

## 3. Feature Breakdown by Command

### 3.1 SpicyAsk - Q&A Assistant

**Purpose**: Quick answers to coding questions

**Input Methods**:
1. Command line args: `:SpicyAsk how does async/await work in Rust`
2. Prompt: `:SpicyAsk` → opens input prompt
3. Visual selection + prompt: Select code → `:SpicyAsk` → ask about selection
4. Operator pending: `<leader>sa` in normal mode → motion → ask about range

**Output Options**:
- Floating window (default)
- Split window
- New buffer
- Quickfix list (for multi-line answers with references)

**Key Features**:
- Syntax highlighting in response (detect code blocks)
- Copy response to clipboard
- Save to scratch buffer
- History of recent questions
- Context: Include current file/language in question automatically

**Example Usage**:
```vim
" Quick question
:SpicyAsk why is my variable undefined

" Visual selection
:'<,'>SpicyAsk explain this code

" With context
:SpicyAsk -v how to optimize this function
```

**Lua API**:
```lua
require('spicy').ask("question here", {
  context = "optional code context",
  model = "anthropic/claude-3-opus",
  verbose = false,
  on_complete = function(response)
    -- custom handler
  end
})
```

---

### 3.2 SpicyTutor - Tutorial Generator

**Purpose**: Generate detailed tutorials on topics

**Input Methods**:
1. Command line args: `:SpicyTutor how to use docker compose`
2. Prompt: `:SpicyTutor` → opens input prompt
3. From word under cursor: `:SpicyTutorWord` → generates tutorial for current word

**Output Options**:
- New buffer with markdown (default)
- Auto-save to configured directory
- Open in markdown preview
- Send to floating window for quick review

**Key Features**:
- Auto-detect save location based on topic
- Markdown preview integration
- Tutorial history browser (via telescope)
- Tags/categories for organizing tutorials

**Example Usage**:
```vim
" Generate tutorial
:SpicyTutor git rebase interactive

" Save to specific location
:SpicyTutor -o ~/notes/docker.md docker networking

" Use different model
:SpicyTutor -m openai/gpt-4 explain closures
```

**Lua API**:
```lua
require('spicy.commands.tutor').generate({
  topic = "tmux basics",
  output_path = "~/tutorials/tmux.md",
  auto_open = true,
  validation_model = "anthropic/claude-3-opus",
  generation_model = "anthropic/claude-3-opus",
})
```

---

### 3.3 SpicyExplain - Code Explainer

**Purpose**: Explain code and save explanations

**Input Methods**:
1. Current buffer: `:SpicyExplain` → explain whole file
2. Visual selection: `'<,'>SpicyExplain` → explain selection
3. Range: `:10,50SpicyExplain` → explain lines 10-50
4. File/directory: `:SpicyExplain path/to/file.go`
5. Multiple files: `:SpicyExplain src/**/*.rs`

**Output Options**:
- New buffer with explanation (default)
- Side-by-side split (code | explanation)
- Floating window
- Save to file automatically

**Key Features**:
- Language detection from filetype
- Manual language override
- Line number references in explanation
- Link back to original code
- Diff-style view (code on left, explanation on right)

**Example Usage**:
```vim
" Explain current file
:SpicyExplain

" Explain selection
:'<,'>SpicyExplain

" Explain with custom language
:SpicyExplain -l rust %

" Don't save, just preview
:SpicyExplain
```

**Lua API**:
```lua
require('spicy.commands.explain').explain_buffer({
  bufnr = 0, -- current buffer
  range = { start_line = 10, end_line = 50 },
  language = "go",
  output = "split",
})
```

---

### 3.4 SpicyGitmessage - Commit Message Generator

**Purpose**: Generate git commit messages from staged changes

**Input Methods**:
1. Auto: `:SpicyGitmessage` → uses git staged changes
2. With hint: `:SpicyGitmessage refactor auth system`
3. With prefix: `:SpicyGitmessage feat user login`
4. From visual selection (hint from code comments)

**Output Options**:
- Floating window with preview (default)
- Auto-copy to clipboard
- Insert into commit message buffer
- Show in quickfix for review

**Key Features**:
- Conventional commits format detection
- Prefix support (feat, fix, chore, etc.)
- Integration with git commit workflow
- Preview diff before generating
- Multi-commit support (generate for each logical change)

**Git Integration**:
- Detect if in git repository
- Check for staged changes
- Auto-insert into `COMMIT_EDITMSG` if in git commit
- Keybinding in commit buffer to regenerate

**Example Usage**:
```vim
" Generate from staged changes
:SpicyGitmessage

" With hint
:SpicyGitmessage fix auth bug

" With prefix and copy
:SpicyGitmessage -c feat add login

" Preview diff first
:SpicyGitmessage --preview
```

**Lua API**:
```lua
require('spicy.commands.gitmessage').generate({
  hint = "optional hint",
  prefix = "feat",
  copy = true,
  model = "anthropic/claude-3-sonnet",
})
```

**Git Commit Integration**:
```vim
" In git commit buffer, add mapping
autocmd FileType gitcommit nnoremap <buffer> <C-g> <cmd>SpicyGitmessageInsert<CR>
```

---

## 4. Configuration System

### Default Configuration

```lua
{
  -- CLI binary configuration
  bin = {
    path = "spicy",  -- Assumes in PATH, or provide full path
    ask = nil,       -- Override individual commands if needed
    tutor = nil,
    explain = nil,
    gitmessage = nil,
  },

  -- Default models for each command
  models = {
    ask = "anthropic/claude-3-opus",
    tutor = "anthropic/claude-3-opus",
    explain = "anthropic/claude-3-opus",
    gitmessage = "anthropic/claude-3-sonnet",
  },

  -- UI configuration per command
  ui = {
    ask = {
      output = "float",  -- "float", "split", "vsplit", "buffer"
      float_opts = {
        relative = "editor",
        width = 0.8,
        height = 0.6,
        border = "rounded",
        title = " Spicy Ask ",
        title_pos = "center",
      },
      syntax = "markdown",
      show_spinner = true,
      auto_close = false,
    },

    tutor = {
      output = "buffer",
      auto_save = true,
      save_dir = vim.fn.expand("~/tutorials"),
      filename_format = function(topic)
        return topic:gsub(" ", "-"):lower() .. ".md"
      end,
      auto_open = true,
      markdown_preview = false,  -- Integrate with preview plugins
    },

    explain = {
      output = "buffer",
      auto_save = true,
      save_dir = vim.fn.expand("~/explanations"),
      filename_format = function(source)
        return source .. "-explanation.md"
      end,
      side_by_side = false,  -- Show code and explanation split
      show_line_numbers = true,
    },

    gitmessage = {
      output = "float",
      auto_copy = true,
      float_opts = {
        relative = "editor",
        width = 0.6,
        height = 0.4,
        border = "rounded",
        title = " Git Commit Message ",
      },
      conventional_commits = true,
      show_diff = true,
      auto_insert = false,  -- Auto-insert in commit buffer
    },
  },

  -- Keybindings (set to false to disable, or provide custom)
  keymaps = {
    enabled = true,
    ask = "<leader>sa",
    tutor = "<leader>st",
    explain = "<leader>se",
    gitmessage = "<leader>sg",

    -- Additional mappings
    ask_visual = "<leader>sa",
    explain_visual = "<leader>se",

    -- Operator pending (not implemented in phase 1)
    -- ask_operator = "<leader>sa",
  },

  -- Behavior
  verbose = false,
  timeout = 300000,  -- 5 minutes in milliseconds

  -- History
  history = {
    enabled = true,
    max_entries = 100,
    save_to_file = true,
    file_path = vim.fn.stdpath("data") .. "/spicy_history.json",
  },

  -- Telescope integration (if available)
  telescope = {
    enabled = true,
    theme = "dropdown",
  },

  -- Statusline integration (optional)
  statusline = {
    enabled = false,
    show_running = true,
    show_last_result = false,
  },
}
```

### Configuration API

```lua
-- Setup with custom config
require('spicy').setup({
  models = {
    ask = "openai/gpt-4",
  },
  ui = {
    ask = {
      output = "split",
    },
  },
})

-- Runtime config changes
require('spicy.config').set('ui.ask.output', 'float')
local model = require('spicy.config').get('models.ask')
```

---

## 5. User Interface Components

### 5.1 Floating Windows

**Features**:
- Responsive sizing (percentage-based)
- Border styles (rounded, single, double, shadow)
- Title bars with icons
- Close on `q`, `<Esc>`, or custom keybinding
- Scroll support for long content
- Syntax highlighting
- Copy to clipboard button

**Implementation**:
```lua
-- lua/spicy/ui/float.lua
local M = {}

function M.create(opts)
  -- opts: width, height, title, content, syntax, border
  -- Returns: bufnr, winid
end

function M.update_content(bufnr, content)
  -- Update content without recreating window
end

function M.close(winid)
  -- Close and cleanup
end

return M
```

### 5.2 Input Prompts

**Features**:
- Floating prompt for user input
- History navigation (up/down arrows)
- Tab completion (for file paths, commands)
- Multi-line support for complex queries
- Cancel on `<C-c>`

**Implementation**:
```lua
-- lua/spicy/ui/input.lua
local M = {}

function M.prompt(opts, callback)
  -- opts: title, default, completion, history
  -- callback: function(input) called with result
end

return M
```

### 5.3 Loading Indicators

**Features**:
- Spinner animation in floating window
- Progress percentage (if available)
- Cancel button/keybinding
- Status line integration

**Types**:
1. Floating spinner overlay
2. Virtual text in status line
3. Popup notification (nvim-notify integration)

**Implementation**:
```lua
-- lua/spicy/ui/spinner.lua
local M = {}

function M.start(message)
  -- Returns: spinner_id
end

function M.stop(spinner_id)
  -- Stop and cleanup
end

function M.update_message(spinner_id, message)
  -- Update spinner message
end

return M
```

---

## 6. Async Job Management

### Requirements

- Non-blocking command execution
- Stream stdout/stderr as it arrives
- Cancel running jobs
- Timeout support
- Error handling

### Implementation Strategy

Use `plenary.job` for async execution:

```lua
-- lua/spicy/utils/job.lua
local Job = require('plenary.job')
local M = {}

function M.run(cmd, args, opts)
  -- opts: on_stdout, on_stderr, on_exit, timeout, cwd
  local job = Job:new({
    command = cmd,
    args = args,
    cwd = opts.cwd or vim.fn.getcwd(),
    on_stdout = opts.on_stdout,
    on_stderr = opts.on_stderr,
    on_exit = function(j, return_code)
      if opts.on_exit then
        opts.on_exit(j:result(), j:stderr_result(), return_code)
      end
    end,
  })

  job:start()

  -- Handle timeout
  if opts.timeout then
    vim.defer_fn(function()
      if job.is_shutdown == false then
        job:shutdown()
        if opts.on_timeout then
          opts.on_timeout()
        end
      end
    end, opts.timeout)
  end

  return job
end

-- Active jobs registry for cancellation
M.active_jobs = {}

function M.cancel_all()
  for _, job in pairs(M.active_jobs) do
    if not job.is_shutdown then
      job:shutdown()
    end
  end
  M.active_jobs = {}
end

return M
```

---

## 7. Health Check System

### Requirements

- Verify spicy CLI is installed
- Check version compatibility
- Validate auth status
- Check for optional dependencies
- Provide actionable error messages

### Implementation

```lua
-- lua/spicy/health.lua
local health = vim.health or require("health")

local M = {}

function M.check()
  health.report_start("spicy.nvim")

  -- Check spicy CLI
  local spicy_path = vim.fn.exepath("spicy")
  if spicy_path == "" then
    health.report_error(
      "spicy CLI not found in PATH",
      {
        "Install spicy: https://github.com/user/spicy",
        "Or configure bin.path in setup()",
      }
    )
  else
    health.report_ok("spicy CLI found: " .. spicy_path)

    -- Check version
    local version = vim.fn.system("spicy --version")
    health.report_info("Version: " .. version)
  end

  -- Check individual commands
  for _, cmd in ipairs({"ask", "tutor", "explain", "gitmessage"}) do
    local cmd_path = vim.fn.exepath(cmd)
    if cmd_path ~= "" then
      health.report_ok(cmd .. " command found")
    else
      health.report_warn(cmd .. " command not found")
    end
  end

  -- Check auth
  -- TODO: Call spicy to validate auth

  -- Check optional dependencies
  local has_plenary = pcall(require, "plenary")
  if has_plenary then
    health.report_ok("plenary.nvim found")
  else
    health.report_error(
      "plenary.nvim not found (required)",
      "Install: https://github.com/nvim-lua/plenary.nvim"
    )
  end

  local has_telescope = pcall(require, "telescope")
  if has_telescope then
    health.report_ok("telescope.nvim found (optional)")
  else
    health.report_info("telescope.nvim not found (optional)")
  end
end

return M
```

---

## 8. Implementation Phases

### Phase 1: MVP (Week 1-2)

**Goal**: Basic functionality, can execute all 4 commands

**Tasks**:
1. ✅ Project setup
   - ✅ Directory structure created
   - ✅ Basic files (init.lua, config.lua)
   - ✅ Git ignore
   - ✅ README skeleton

2. ✅ Core infrastructure
   - ✅ Configuration system (lua/spicy/config.lua)
   - ✅ Job execution wrapper (lua/spicy/utils/job.lua)
   - ✅ Filesystem helpers (lua/spicy/utils/fs.lua)
   - ✅ General helpers (lua/spicy/utils/helpers.lua)
   - ✅ Health check (lua/spicy/health.lua)

3. ✅ UI basics
   - ✅ Floating window utility (lua/spicy/ui/float.lua)
   - ✅ Input prompt (lua/spicy/ui/input.lua)
   - ✅ Simple spinner (lua/spicy/ui/spinner.lua)

4. 🔄 Command implementations
   - ✅ SpicyAsk (basic) - lua/spicy/commands/ask.lua
   - 🚧 SpicyTutor (stub)
   - 🚧 SpicyExplain (stub)
   - 🚧 SpicyGitmessage (stub)

5. ✅ Vim commands
   - ✅ `:SpicyAsk` (plugin/spicy.lua)
   - ✅ `:SpicyTutor` (stub)
   - ✅ `:SpicyExplain` (stub)
   - ✅ `:SpicyGitmessage` (stub)

6. ✅ Documentation
   - ✅ README with installation
   - ✅ Basic usage examples
   - ✅ Configuration docs

**Success Criteria**:
- ✅ Can execute SpicyAsk command
- ✅ Output shown in floating window
- ✅ Basic error handling works
- ✅ Health check validates installation
- 🚧 Other commands (stubs ready for implementation)

**Progress**: 85% complete
**Next**: Testing and implementing remaining commands

---

### Phase 2: Enhanced UX (Week 3-4)

**Goal**: Better user experience, visual mode support, history

**Tasks**:
1. ⬜ Visual selection support
   - Ask about selected code
   - Explain selected code
   - Context-aware prompts

2. ⬜ Range support
   - Explain lines 10-50
   - Ask about specific range

3. ⬜ Output options
   - Buffer output mode
   - Split output mode
   - Auto-save for tutor/explain

4. ⬜ Loading UX
   - Better spinner animations
   - Progress indicators
   - Cancellation support

5. ⬜ History system
   - Save query history
   - Retrieve previous results
   - Clear history

6. ⬜ Error handling
   - Auth error detection
   - Network error handling
   - Timeout handling
   - User-friendly error messages

7. ⬜ Testing
   - Unit tests for utilities
   - Integration tests
   - Mock CLI for testing

**Success Criteria**:
- Visual selection works smoothly
- Can cancel long-running operations
- History persists across sessions
- Comprehensive error messages

---

### Phase 3: Advanced Features (Week 5-6)

**Goal**: Polish, integrations, advanced features

**Tasks**:
1. ⬜ Telescope integration
   - Browse history with telescope
   - Search previous questions/answers
   - Preview results

2. ⬜ Git commit integration
   - Auto-detect commit buffer
   - Insert message at cursor
   - Keybinding in commit buffer
   - Preview diff before generating

3. ⬜ Statusline integration
   - Show running status
   - Display last result summary
   - Integration with lualine/galaxyline

4. ⬜ Advanced UI
   - Side-by-side splits (code | explanation)
   - Diff view with syntax highlighting
   - Custom highlights/themes
   - Icons (if nerd fonts available)

5. ⬜ Code actions integration
   - LSP-style code actions
   - "Explain this" code action
   - "Ask about this" code action

6. ⬜ Operator pending mode
   - `<leader>sa` + motion to ask about text object
   - `<leader>se` + motion to explain

7. ⬜ Additional commands
   - `:SpicyHistory` - browse history
   - `:SpicyConfig` - edit configuration
   - `:SpicyModels` - list available models

8. ⬜ Documentation
   - Complete vimdoc
   - GIF demos
   - Advanced examples
   - Troubleshooting guide

**Success Criteria**:
- Telescope integration seamless
- Git workflow integration natural
- Code actions work with LSP
- Documentation comprehensive

---

### Phase 4: Polish & Release (Week 7)

**Goal**: Production-ready, stable release

**Tasks**:
1. ⬜ Bug fixes
   - Address all known issues
   - Edge case handling
   - Performance optimization

2. ⬜ Code quality
   - Linting (luacheck, stylua)
   - Code review
   - Refactoring

3. ⬜ CI/CD
   - GitHub Actions for tests
   - Automated linting
   - Release automation

4. ⬜ Documentation polish
   - Final README review
   - Changelog
   - Migration guide
   - FAQ

5. ⬜ Community prep
   - Contributing guide
   - Issue templates
   - PR templates
   - Code of conduct

6. ⬜ Release
   - Tag v1.0.0
   - Announce on reddit/discord
   - Submit to awesome-neovim
   - Post on HN/lobsters if significant interest

**Success Criteria**:
- Zero known critical bugs
- All tests passing
- Documentation complete
- Ready for public use

---

## 9. Testing Strategy

### Unit Tests

**Framework**: plenary.nvim test suite

**Coverage**:
- Configuration parsing and validation
- Utility functions (string manipulation, path handling)
- Job execution wrapper
- UI components (floating window creation, input prompts)

**Example**:
```lua
-- tests/config_spec.lua
local config = require('spicy.config')

describe('config', function()
  it('should load default config', function()
    local cfg = config.get_default()
    assert.is_not_nil(cfg.models)
    assert.equals('anthropic/claude-3-opus', cfg.models.ask)
  end)

  it('should merge user config', function()
    config.setup({ models = { ask = 'openai/gpt-4' } })
    assert.equals('openai/gpt-4', config.get('models.ask'))
  end)
end)
```

### Integration Tests

**Coverage**:
- End-to-end command execution
- Mock CLI responses for predictable testing
- Visual mode integration
- Git workflow integration

**Mock CLI**:
```bash
#!/bin/bash
# tests/mocks/spicy
case "$1" in
  ask)
    echo "This is a mock answer"
    ;;
  tutor)
    echo "# Mock Tutorial\n\nContent here"
    ;;
  *)
    echo "Unknown command"
    exit 1
    ;;
esac
```

### Manual Testing Checklist

- [ ] Install in fresh Neovim config
- [ ] All commands work with default config
- [ ] Visual selection works
- [ ] Range commands work
- [ ] Error handling graceful
- [ ] Health check accurate
- [ ] Keybindings work
- [ ] Telescope integration works
- [ ] Git commit workflow smooth

---

## 10. Dependencies

### Required

- **plenary.nvim**: Async jobs, functional utilities
  - Used for: Job execution, async operations, test framework
  - Why: De facto standard for Neovim Lua plugins

### Optional

- **telescope.nvim**: History browsing, fuzzy finding
  - Used for: History picker, model selector
  - Fallback: Simple vim.ui.select

- **nui.nvim**: Advanced UI components
  - Used for: Rich floating windows, input components
  - Fallback: Built-in floating windows

- **nvim-notify**: Better notifications
  - Used for: Error messages, progress notifications
  - Fallback: vim.notify

- **markdown-preview.nvim**: Preview tutorials
  - Used for: Auto-preview generated tutorials
  - Fallback: None, just open in buffer

### Dependency Management

- Keep required deps minimal (only plenary)
- Gracefully degrade when optional deps missing
- Document optional integrations clearly
- Never hard-depend on UI preferences

---

## 11. Documentation Plan

### README.md

Sections:
1. Introduction (what is spicy.nvim)
   - Monorepo note: this plugin lives in `nvim/` when using the combined repo
2. Features (with GIFs/screenshots)
3. Installation (lazy.nvim, packer, vim-plug)
4. Prerequisites (spicy CLI)
5. Quick start (minimal config)
6. Configuration (full example)
7. Usage (each command with examples)
8. Keybindings (default + custom)
9. Integration (telescope, statusline, git)
10. Troubleshooting
11. Contributing
12. License

### Vimdoc (doc/spicy.nvim.txt)

Sections:
```
*spicy.nvim.txt*  AI-powered coding assistance for Neovim

==============================================================================
CONTENTS                                                    *spicy-contents*

1. Introduction ...................... |spicy-introduction|
2. Requirements ...................... |spicy-requirements|
3. Installation ...................... |spicy-installation|
4. Commands .......................... |spicy-commands|
5. Configuration ..................... |spicy-configuration|
6. Lua API ........................... |spicy-api|
7. Telescope Integration ............. |spicy-telescope|
8. FAQ ............................... |spicy-faq|
9. Changelog ......................... |spicy-changelog|

==============================================================================
1. INTRODUCTION                                         *spicy-introduction*

Spicy.nvim provides seamless Neovim integration for the spicy CLI tools...

==============================================================================
2. REQUIREMENTS                                         *spicy-requirements*

- Neovim 0.8+
- plenary.nvim
- spicy CLI tools v1.0+

Optional:
- telescope.nvim (for history browsing)
- nui.nvim (for enhanced UI)

...
```

### Code Documentation

- Every public function documented with LuaCATS annotations
- Module-level documentation
- Example usage in comments
- Type hints where applicable

Example:
```lua
---Ask a question and display answer
---@param question string The question to ask
---@param opts table|nil Optional configuration
---  - context: string Additional context
---  - model: string Model to use
---  - on_complete: function Callback on completion
---@return number|nil job_id The job ID or nil on error
function M.ask(question, opts)
  -- implementation
end
```

---

## 12. Performance Considerations

### Lazy Loading

- Don't load commands until first use
- Defer UI components until needed
- Load telescope integration only if telescope installed

### Async Everything

- Never block on network calls
- Stream output as it arrives
- Cancel-able operations

### Memory Management

- Clean up floating windows
- Limit history size
- Don't cache large responses indefinitely

### Startup Time

- Minimal impact on Neovim startup
- Use autoload directories properly
- Lazy load optional integrations

**Measurement**:
```lua
-- Benchmark startup impact
vim.cmd('profile start profile.log')
vim.cmd('profile func *')
require('spicy')
vim.cmd('profile stop')
```

---

## 13. Security Considerations

### Input Sanitization

- Sanitize file paths
- Escape shell arguments
- Validate user input

### Auth Handling

- Never log auth tokens
- Respect CLI's auth mechanism
- Clear error on auth failure

### File Operations

- Validate save paths
- Check permissions before writing
- Don't overwrite without confirmation

### Code Execution

- Use plenary.job, not vim.fn.system
- Avoid shell injection
- Validate command paths

---

## 14. Accessibility

### Colorblind Support

- Don't rely solely on color for status
- Use symbols/text for indicators
- Configurable highlights

### Keyboard Navigation

- All features keyboard-accessible
- No mouse-only features
- Standard vim motions work

### Screen Readers

- Meaningful window titles
- Status updates via messages
- Clear error text

---

## 15. Versioning & Release Strategy

### Semantic Versioning

- MAJOR: Breaking changes
- MINOR: New features, backwards compatible
- PATCH: Bug fixes

### Release Process

1. Update CHANGELOG.md
2. Bump version in plugin files
3. Create git tag
4. GitHub release with notes
5. Announce in community

### Changelog Format

```markdown
# Changelog

## [1.0.0] - 2026-03-30

### Added
- Initial release
- SpicyAsk command
- SpicyTutor command
- SpicyExplain command
- SpicyGitmessage command
- Floating window UI
- History system

### Changed
- N/A (initial release)

### Fixed
- N/A (initial release)

### Breaking Changes
- N/A (initial release)
```

---

## 16. Success Metrics

### Adoption

- GitHub stars > 100 in first month
- 10+ community contributions
- Featured on awesome-neovim

### Quality

- Zero critical bugs in production
- 90%+ test coverage
- < 5ms startup overhead

### User Satisfaction

- Positive feedback on reddit/discord
- Low issue-to-PR ratio
- Active users contributing ideas

---

## 17. Future Enhancements (Post v1.0)

### Potential Features

1. **Streaming responses**: Show AI response as it's generated
2. **Multi-file context**: Include multiple files in question context
3. **Conversation mode**: Follow-up questions in same context
4. **Custom prompts**: User-defined prompt templates
5. **Workspace integration**: Project-aware context
6. **Diff integration**: Explain git diffs
7. **LSP integration**: "Explain this error" code action
8. **Collaborative**: Share questions/answers with team
9. **Offline mode**: Cache common questions/answers
10. **Voice input**: Whisper integration for voice questions

### Community Requests

- Track feature requests in GitHub issues
- Community voting on features
- Regular roadmap updates

---

## 18. Risk Assessment & Mitigation

### Technical Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Plenary API changes | High | Low | Pin version, monitor updates |
| Neovim API changes | High | Medium | Support multiple versions |
| CLI breaking changes | High | Medium | Version compatibility check |
| Performance issues | Medium | Low | Benchmark, optimize early |

### Product Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Low adoption | Medium | Medium | Good docs, demos, marketing |
| Competition | Low | High | Focus on unique features |
| Maintenance burden | High | Medium | Clear contribution guide |

---

## 19. Open Questions

1. **Model selection UI**: Should we provide a picker for models?
2. **Context limits**: How much code context to include in asks?
3. **Caching**: Should we cache any responses?
4. **Conflicts**: How to handle multiple simultaneous requests?
5. **Pricing awareness**: Should plugin warn about API costs?
6. **Offline fallback**: What happens without internet?
7. **Multi-language**: Support for non-English queries?

---

## 20. Development Environment Setup

### Prerequisites

```bash
# Install Neovim
brew install neovim  # macOS
# or
sudo apt install neovim  # Linux

# Install spicy CLI
cd /Users/christian/.local/bin/spicy
make install-all

# Clone plugin repo
cd /Users/christian/personal/dev
git clone spicy
cd spicy/nvim
```

### Testing Setup

```lua
-- tests/minimal_init.lua
-- Minimal config for running tests
local plenary_dir = vim.fn.stdpath("data") .. "/plenary.nvim"

if vim.fn.isdirectory(plenary_dir) == 0 then
  vim.fn.system({
    "git",
    "clone",
    "https://github.com/nvim-lua/plenary.nvim",
    plenary_dir,
  })
end

vim.opt.rtp:append(".")
vim.opt.rtp:append(plenary_dir)

vim.cmd("runtime plugin/plenary.vim")
```

### Run Tests

```bash
# Run all tests
make test

# Run specific test
nvim --headless -c "PlenaryBustedFile tests/config_spec.lua"

# Watch mode (requires entr)
ls lua/**/*.lua tests/**/*.lua | entr make test
```

### Linting

```bash
# Install tools
luarocks install luacheck
luarocks install stylua

# Run checks
make lint
make format
```

---

## 21. Timeline Summary

| Week | Phase | Focus | Deliverables |
|------|-------|-------|--------------|
| 1-2 | MVP | Core functionality | Working commands, basic UI |
| 3-4 | Enhancement | UX improvements | Visual mode, history, tests |
| 5-6 | Advanced | Integrations | Telescope, git, statusline |
| 7 | Polish | Release prep | Docs, CI/CD, v1.0.0 |

**Total**: ~7 weeks to v1.0.0

---

## 22. Appendix

### Naming Conventions

- **Commands**: PascalCase with Spicy prefix (`:SpicyAsk`)
- **Lua modules**: lowercase snake_case (`spicy/utils/job.lua`)
- **Functions**: snake_case (`function M.create_float()`)
- **Constants**: UPPER_SNAKE_CASE (`local MAX_HISTORY = 100`)
- **Private**: Leading underscore (`local function _internal()`)

### Code Style

- 2 spaces indentation
- 80 character line limit
- Double quotes for strings
- Trailing commas in tables
- Early returns for error handling

### Git Workflow

- `main` branch: stable releases
- `develop` branch: integration
- Feature branches: `feature/ask-visual-mode`
- Bug fixes: `fix/floating-window-resize`

### Commit Messages

```
feat(ask): add visual selection support

- Support visual mode for code context
- Add range support for line selections
- Update documentation

Closes #42
```

---

## End of Plan

This plan serves as a living document. Update as the project evolves.

**Next Steps**:
1. Review and refine this plan
2. Set up project structure (Phase 1, Task 1)
3. Begin MVP implementation
4. Track progress in GitHub project board

---

**Document Version**: 1.0
**Author**: AI Assistant + Christian
**Date**: 2026-03-15
**Status**: Draft - Pending Review
