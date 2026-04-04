  # spicy.nvim

> Neovim plugin for seamless integration with spicy CLI tools

**Status**: 🚧 Phase 1 MVP - In Development

## Features

- ✅ **SpicyAsk**: Ask questions and get AI-powered answers
- ✅ **SpicyCtxEdit**: Edit a visual selection with AI
- ✅ **SpicyTutor**: Generate detailed tutorials
- ✅ **SpicyExplain**: Explain code with AI
- ✅ **SpicyGitmessage**: Generate commit messages

## Prerequisites

- Neovim 0.8+
- [plenary.nvim](https://github.com/nvim-lua/plenary.nvim) (required)
- Spicy CLI installed and in PATH (see `../cli/README.md`)

## Installation

### lazy.nvim (monorepo local)

```lua
{
  dir = "/path/to/spicy/nvim",
  name = "spicy",
  dependencies = {
    "nvim-lua/plenary.nvim",
  },
  config = function()
    require("spicy").setup({
      -- Optional configuration
      models = {
        ask = "anthropic/claude-3-opus",
      },
    })
  end,
}
```

### packer.nvim (monorepo local)

```lua
use {
  "/path/to/spicy/nvim",
  requires = { "nvim-lua/plenary.nvim" },
  config = function()
    require("spicy").setup()
  end,
}
```

## Quick Start

```vim
" Ask a question
:SpicyAsk how does async/await work in Rust

" Ask about visual selection
" 1. Select code in visual mode
" 2. Run:
:'<,'>SpicyAsk explain this code
```

See `example-config.lua` for keymap setup examples.

## Configuration

See `example-config.lua` for a complete configuration example with keymaps.

Basic setup:

```lua
require("spicy").setup({
  -- Models
  models = {
    ask = "openai/gpt-5.2-codex",
    tutor = "openai/gpt-5.2-codex",
    explain = "openai/gpt-5.2-codex",
    gitmessage = "openai/gpt-5.2-codex",
    ctx_edit = "openai/gpt-5.2-codex",
  },

  -- UI settings
  ui = {
    ask = {
      output = "float",  -- "float", "buffer", "split"
      float_opts = {
        width = 0.8,
        height = 0.6,
        border = "rounded",
      },
    },
    explain = {
      context_max_chars = 3000,
      context_surround_lines = 80,
    },
  },

  -- Behavior
  verbose = false,
  timeout = 300000,  -- 5 minutes
})

-- Set up keymaps (see example-config.lua for more options)
vim.keymap.set("n", "<leader>sa", "<cmd>SpicyAsk<CR>", { desc = "Spicy: Ask" })
vim.keymap.set("v", "<leader>sa", ":'<,'>SpicyAsk<CR>", { desc = "Spicy: Ask about selection" })
vim.keymap.set("n", "<leader>se", "<cmd>SpicyExplain<CR>", { desc = "Spicy: Explain" })
vim.keymap.set("n", "<leader>st", "<cmd>SpicyTutor<CR>", { desc = "Spicy: Tutor" })
vim.keymap.set("n", "<leader>sg", "<cmd>SpicyGitmessage<CR>", { desc = "Spicy: Git message" })
vim.keymap.set("v", "<leader>sc", ":'<,'>SpicyCtxEdit<CR>", { desc = "Spicy: Edit selection" })
```

## Commands

### :SpicyAsk

Ask a question and get an AI-powered answer.

```vim
" Ask from command line
:SpicyAsk what is a closure in JavaScript

" Ask interactively (prompts for question)
:SpicyAsk

" Ask about visual selection
:'<,'>SpicyAsk explain this
```

### :SpicyCtxEdit

Edit a visual selection using an instruction prompt.

```vim
" Select code in visual mode, then:
:'<,'>SpicyCtxEdit
```

### :SpicyExplain

Explain the current file or a visual selection.

```vim
" Explain current buffer
:SpicyExplain

" Explain a visual selection
:'<,'>SpicyExplain
```

### :SpicyTutor

Generate a tutorial from a topic.

```vim
:SpicyTutor how to use git rebase
```

### :SpicyGitmessage

Generate a commit message from staged changes.

```vim
:SpicyGitmessage
:SpicyGitmessage feat add caching
```

## Health Check

Run health check to verify installation:

```vim
:checkhealth spicy
```

## Development

### Running Tests

```bash
# Run all tests
make test

# Run a specific test file
make test-file FILE=tests/config_spec.lua

# Run linter
make lint

# Format code
make format

# Run all checks (lint + test)
make check
```

### Requirements for Development

- Neovim 0.8+
- [plenary.nvim](https://github.com/nvim-lua/plenary.nvim) (auto-installed by Makefile)
- Optional: luacheck (for linting)
- Optional: stylua (for formatting)

## Development Status

### Phase 1: MVP (Complete)

- [x] Configuration system
- [x] Core infrastructure
- [x] SpicyAsk command (basic)
- [x] Health check
- [ ] Basic tests
- [ ] Documentation

### Phase 2: Enhanced UX

- [ ] SpicyTutor command
- [ ] SpicyExplain command
- [ ] SpicyGitmessage command
- [ ] History system
- [ ] Better error handling

### Phase 3: Advanced Features

- [ ] Telescope integration
- [ ] Git workflow integration
- [ ] Statusline components

## Contributing

This plugin is in early development. Contributions welcome!

## License

MIT

## History

The plugin runs CLI commands with `--history` enabled by default, so history is
saved to:
```
.spicy/<command>/YYYYMMDD-HHMMSS_<command>[_suggestion].json
```

Each entry is saved as a JSON file with:
- Question/topic and answer/content
- Timestamp and metadata
- Pretty-printed for readability

View your history:
```bash
ls .spicy/ask/
cat .spicy/ask/*.json
```

## Links

- [Spicy CLI](https://github.com/user/spicy)
- [Issues](https://github.com/user/spicy.nvim/issues)
