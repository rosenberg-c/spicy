# Spicy

This repo contains the core tooling and editor/automation integrations. The
CLI tools are built on the `opencode` API.

Note: This is an alpha release and is prone to change. Since this repo is
focused on tools around AI, expect changes as I explore how I work with AI.

I believe in learning and understanding the code we write; that drive is what pushed
this tooling suite forward, especially `tutor` and `explain`.

- `cli/` - The Spicy CLI tools (Go)
- `nvim/` - The spicy.nvim Neovim plugin
- `hammerspoon/` - The Hammerspoon module that wraps `spicy ask`
- `imgwalker/` - Image browser desktop app (Gio)

Quick setup:

- CLI: see [CLI](#cli)
- Neovim: see [Neovim plugin](#neovim-plugin)
- Hammerspoon: see [Hammerspoon module](#hammerspoon-module)

## Inspiration

This project was inspired by https://github.com/ThePrimeagen/99, but I chose to build my own solution tailored to how I work.

## Requirement-driven workflow

Spicy now uses a requirement-first workflow for behavior changes:

- Requirements live in `docs/REQUIREMENTS.md` and `docs/requirements/*.md`
- Automated coverage mapping lives in `docs/TEST_MATRIX.md`
- Test-to-requirement links use `@req` comment tags in test files

Regenerate the test matrix with:

```sh
make sync-test-matrix
```

## Feature highlights

### CLI (`cli/`)

- `ask` - quick questions and brainstorming
- `explain` - code walk-throughs from files, dirs, or stdin
- `tutor` - step-by-step learning and tutorials
- `v-edit` - targeted edits on a scoped snippet
- `gitmessage` - commit message drafts from staged changes
- `record` - browse and export past runs

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

### ImgWalker (`imgwalker/`)

- Keyboard-driven image browser UI built with Gio.
- The `p` shortcut copies the selected image path to system clipboard.
- The `m` shortcut opens an OS directory picker and moves the selected image.
- Linux clipboard dependency: install `wl-clipboard` (`wl-copy`) for Wayland or `xclip` for X11.
- Linux move-picker dependency: install `zenity` (preferred) or `kdialog`.
- XFCE shortcut fallback command: `/bin/bash -lc 'source "$HOME/.bashprofile"; imgwalker'`.

## CLI

See `cli/README.md` for full docs.

Common commands (run from repo root):

```sh
make -C cli build-all
make -C cli install-all
```

You can run installs from any directory with:

```sh
make -C /path/to/spicy install
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

The Hammerspoon integration lives in `hammerspoon/modules/spicy.lua` and wraps the `spicy` CLI launchers.

Default hotkeys:

- `alt+shift+A` -> `askwrapper ui ask`
- `alt+shift+S` -> `askwrapper ui followup`
- `alt+shift+D` -> `imgwalker`

Setup (symlink into your Hammerspoon config):

```sh
ln -s "$(pwd)/hammerspoon/modules/spicy.lua" "$HOME/.hammerspoon/modules/spicy.lua"
```

Then enable it from your `~/.hammerspoon/init.lua`:

```lua
require("modules.spicy").setup()
```

## XFCE shortcut (Debian)

If you use XFCE instead of Hammerspoon, you can bind AskWrapper in:

- `Settings Manager -> Keyboard -> Application Shortcuts`

Use this command if `askwrapper` is not found from XFCE shortcuts:

```sh
/bin/bash -lc 'source "$HOME/.bashprofile"; askwrapper ui ask'
```

`~/.bashprofile` should include your `PATH` exports for both `~/.local/bin`
and `~/.opencode/bin`.

If `opencode` is still not found, use an explicit `PATH` fallback:

```sh
/bin/bash -lc 'export PATH="$HOME/.local/bin:$HOME/.opencode/bin:/usr/local/bin:/usr/bin:/bin"; askwrapper ui ask'
```
