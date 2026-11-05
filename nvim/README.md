# spicy.nvim

Neovim integration for the [Spicy CLI](../cli/README.md). Requires Neovim 0.8+, [plenary.nvim](https://github.com/nvim-lua/plenary.nvim), and installed Spicy CLI commands in your `PATH`.

## Install

With lazy.nvim and a local checkout:

```lua
{
  dir = "/path/to/spicy/nvim",
  name = "spicy",
  dependencies = { "nvim-lua/plenary.nvim" },
  config = function()
    require("spicy").setup()
  end,
}
```

For keymaps and more options, see [example-config.lua](example-config.lua) and [example-config.json](example-config.json). Run `:checkhealth spicy` to check the installation.

## Commands

| Command | Action |
| --- | --- |
| `:SpicyAsk [question]` | Ask a question; prompts if no question is given; accepts a visual selection |
| `:SpicyCtxEdit` | Edit a visual selection using a prompt |
| `:SpicyExplain` | Explain the current buffer or a selection |
| `:SpicyTutor [topic]` | Generate a tutorial |
| `:SpicyGitmessage [context]` | Draft a message from staged changes |

For example, select code in visual mode and run `:'<,'>SpicyAsk explain this code`, or use `:'<,'>SpicyCtxEdit` to edit the selection. Output can appear in a float, buffer, or split. The plugin enables CLI `--history` by default, saving JSON entries and matching result Markdown files under `.spicy/<command>/` in the working directory.

## Configuration

```lua
require("spicy").setup({
  default_model = "openai/gpt-5.3-codex",
  ui = {
    ask = { output = "float" }, -- "float", "buffer", or "split"
    explain = { context_max_chars = 3000, context_surround_lines = 80 },
  },
  verbose = false,
  timeout = 300000, -- milliseconds
})
```

The plugin reads the shared `~/.config/spicy/config.json` and the nearest `.spicy.config.json` from Neovim's current directory upward, like the CLI. Model settings are merged per key, with `setup(opts)` taking priority over local config, then global config, then built-in defaults. An explicit `opts.model` on a command takes priority over configured models. Invalid config warns and falls back to the remaining layers. See [CLI model configuration](../cli/README.md#models-and-configuration) for file format and merging details.

## Development

From `nvim/`, run `make test` (or `make test-file FILE=tests/config_spec.lua`). `make check` runs lint and tests; it needs `luacheck`. `make format-check` needs `stylua`. Install test dependencies with `make install-deps`.

To check the effective model for `explain` in Neovim:

```vim
:lua print(require("spicy.config").get_model("explain"))
```
