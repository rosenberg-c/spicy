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
