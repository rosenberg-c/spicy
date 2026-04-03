# Spicy Monorepo

This repo contains the core tooling and editor/automation integrations:

- `cli/` - The Spicy CLI tools (Go)
- `nvim/` - The spicy.nvim Neovim plugin
- `hammerspoon/` - The Hammerspoon module that wraps `spicy ask`

Quick setup:

- CLI: see [CLI](#cli)
- Neovim: see [Neovim plugin](#neovim-plugin)
- Hammerspoon: see [Hammerspoon module](#hammerspoon-module)

## Inspiration

This project was inspired by https://github.com/ThePrimeagen/99, but I chose to build my own solution tailored to how I work.

## Feature highlights

### CLI (`cli/`)

- `ask`: interactive or CLI arguments, model selection, save to markdown, optional history
- `explain`: file/dir/stdin inputs, language detection, optional context, save to markdown, history export
- `tutor`: input validation, separate validation/generation models, save to markdown, history
- `ctx-edit`: edit selected context from file/lines or stdin, optional in-place write, JSON output
- `gitmessage`: staged diff summary, optional hint/prefix, copy to clipboard, history
- `history`: list/filter entries and export to markdown

Examples:

```sh
ask "what is a closure"
ask --history -m openai/gpt-5.2-codex "explain rust lifetimes"

explain main.go
explain ./internal/agent --save
pbpaste | explain --lang go

tutor "how does git rebase work"
tutor --save --validation-model openai/gpt-4o --generation-model openai/o1 "how to use ffmpeg"

ctx-edit -p "rename foo to bar" -c "const foo = 1"
ctx-edit -p "add error handling" -f main.go --start 12 --end 24 --write

gitmessage feat -c
gitmessage -i "focus on perf" fix

shistory list --command ask
shistory export --file .spicy/ask/20260317-134703_ask_what-is-docker.json
```

### Neovim plugin (`nvim/`)

- Commands for `SpicyAsk` and `SpicyCtxEdit` with visual selection support
- Configurable models and UI output modes (float/buffer/split)
- CLI-backed execution with built-in history saving
- Health check integration (`:checkhealth spicy`)

Examples:

```vim
:SpicyAsk what is a closure in JavaScript
:'<,'>SpicyAsk explain this selection
:'<,'>SpicyCtxEdit
```

### Hammerspoon module (`hammerspoon/`)

- Hotkeys to run `ask` in iTerm or fetch output into Sublime
- History browser with inline previews
- Keyboard navigation + delete entries (backspace/delete or Ctrl+D)
- Lightweight spinner UI while running `ask`

Examples:

- `alt+shift+A` -> prompt, run `ask` in a new iTerm window
- `alt+shift+S` -> prompt, run `ask`, open response in Sublime
- Use arrow keys to select history, press `Ctrl+D` to delete

## CLI

See `cli/README.md` for full docs.

Common commands (run from repo root):

```sh
make -C cli build-all
make -C cli install-all
```

## Neovim plugin

See `nvim/README.md` for full docs.

Lazy.nvim example:

```lua
{
  dir = "/path/to/spicy/nvim",
  name = "spicy",
  dependencies = {
    "nvim-lua/plenary.nvim",
  },
  config = function()
    require("spicy").setup({
      verbose = true,
    })
  end,
}
```

## Hammerspoon module

The Hammerspoon integration lives in `hammerspoon/modules/askwrapper.lua` and wraps the `spicy ask` CLI.

Setup (symlink into your Hammerspoon config):

```sh
ln -s "$(pwd)/hammerspoon/modules/askwrapper.lua" "$HOME/.hammerspoon/modules/askwrapper.lua"
```

Then enable it from your `~/.hammerspoon/init.lua`:

```lua
require("modules.askwrapper").setup()
```
