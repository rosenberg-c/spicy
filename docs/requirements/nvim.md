# Neovim Plugin Requirements

### `NVIM-CONFIG-001`

Owner: `nvim/lua/spicy/config.lua`, `nvim/tests/config_spec.lua`

Plugin setup must merge configuration with precedence:
defaults < `~/.config/spicy/config.json` < `setup(opts)`.

### `NVIM-CONFIG-002`

Owner: `nvim/lua/spicy/config.lua`, `nvim/tests/config_spec.lua`

Dot-notation getters must return nested values correctly.

### `NVIM-CONFIG-003`

Owner: `nvim/lua/spicy/config.lua`, `nvim/tests/config_spec.lua`

Invalid config types must produce validation errors.

### `NVIM-CONFIG-004`

Owner: `nvim/lua/spicy/config.lua`, `nvim/tests/config_spec.lua`

Lazy config access must also load `~/.config/spicy/config.json`, ignore a
missing file, and warn without crashing when the file is invalid JSON or
fails validation.

### `NVIM-CONFIG-005`

Owner: `nvim/lua/spicy/config.lua`, `nvim/lua/spicy/utils/cli.lua`,
`nvim/tests/config_spec.lua`, `nvim/tests/cli_spec.lua`

When `default_model` is configured, spicy.nvim must use it for commands that
do not have an explicit `models.<command>` override from config, while keeping
model precedence as: command option (`opts.model`) > explicit
`models.<command>` > `default_model`.

### `NVIM-CONFIG-006`

Owner: `nvim/lua/spicy/config.lua`, `nvim/lua/spicy/utils/fs.lua`,
`nvim/tests/config_spec.lua`, `nvim/tests/fs_spec.lua`

spicy.nvim must support a per-repo local config file named
`.spicy.config.json` with the same JSON shape as the global config.
Discovery starts at Neovim's current working directory and checks that
directory and each parent directory up to the filesystem root; the first
existing regular file named `.spicy.config.json` wins, and a missing file
means no local config. Precedence is `setup(opts)` over local
`.spicy.config.json` over global `~/.config/spicy/config.json` over the
built-in defaults, merged per key: a local `default_model` overrides the
global one, each local `models.<command>` entry overrides the matching
global entry, and global `models` keys absent from the local file are
preserved. A missing, unreadable, or invalid local config file (invalid JSON
or failed validation) must be ignored with a one-time warning so the global
config and defaults still apply.

### `NVIM-ASK-001`

Owner: `nvim/plugin/spicy.lua`, `nvim/lua/spicy/commands/ask.lua`,
`nvim/lua/spicy/utils/runner.lua`, `nvim/lua/spicy/utils/job.lua`,
`nvim/tests/ask_spec.lua`

When a visual selection is active, `SpicyAsk` (both the `v`-mapping form and a
`:SpicyAsk` issued after a visual selection) must send the selected lines to
the ask CLI as context. That context must be transported via `--context-file`
(a temp file that is removed on completion, timeout, or failure) rather than
inlined into the prompt text. A synchronous CLI spawn failure must surface the
error, stop the spinner, and clean up instead of leaving the UI hung.
