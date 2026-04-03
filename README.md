# Spicy Monorepo

This repo contains two separate projects:

- `cli/` - The Spicy CLI tools (Go)
- `nvim/` - The spicy.nvim Neovim plugin

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
ln -s "/path/to/spicy/hammerspoon/modules/askwrapper.lua" "$HOME/.hammerspoon/modules/askwrapper.lua"
```

Then enable it from your `~/.hammerspoon/init.lua`:

```lua
require("modules.askwrapper").setup()
```
