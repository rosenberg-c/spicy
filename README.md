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

- `ask` - quick questions and brainstorming
- `explain` - code walk-throughs from files, dirs, or stdin
- `tutor` - step-by-step learning and tutorials
- `ctx-edit` - targeted edits on a scoped snippet
- `gitmessage` - commit message drafts from staged changes
- `shistory` - browse and export past runs

### Neovim plugin (`nvim/`)

- Commands for `SpicyAsk`, `SpicyExplain`, `SpicyTutor`, `SpicyGitmessage`, and `SpicyCtxEdit`
- Configurable models and UI output modes (float/buffer/split)
- CLI-backed execution with built-in history saving
- Health check integration (`:checkhealth spicy`)
- See `nvim/README.md` for setup and usage

### Hammerspoon module (`hammerspoon/`)

- Hotkeys to run `ask` in iTerm or open results in Sublime Text
- History browser with inline previews and delete shortcuts
- Lightweight spinner UI while running `ask`
- See `hammerspoon/README.md` for setup and usage

## CLI

See `cli/README.md` for full docs.

Common commands (run from repo root):

```sh
make -C cli build-all
make -C cli install-all
```

You can run installs from any directory with:

```sh
make install-all
```

Ensure your shell `PATH` includes `~/.local/bin` (default install target):

```sh
echo $PATH
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
