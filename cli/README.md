# Spicy CLI

Go command-line tools built around OpenCode. Requires Go 1.23.8+ to build and [OpenCode](https://opencode.ai/) to run AI commands. Run these commands from `cli/`, or use `make -C cli <target>` from the repository root.

## Install

```sh
make install-all                 # installs to ~/.local/bin
make install-askwrapper-gio     # optional Gio desktop version of askwrapper
```

Add `~/.local/bin` to your `PATH`, or set `INSTALL_DIR=/path/to/bin`. From the repository root, use `make install` or `make install-gio`. The Gio build needs platform dependencies; on Debian/Ubuntu:

```sh
sudo apt install pkg-config libx11-dev libxkbcommon-dev libxkbcommon-x11-dev \
  libx11-xcb-dev libxcursor-dev libxfixes-dev libwayland-dev libvulkan-dev \
  libegl1-mesa-dev libgl1-mesa-dev
```

## Commands

| Command | Purpose | Example |
| --- | --- | --- |
| `ask` | Ask a question | `ask "What is Docker?"` |
| `explain` | Explain a file, directory, or stdin | `explain main.go` |
| `tutor` | Generate a tutorial | `tutor how to use docker compose` |
| `v-edit` | Edit a snippet or file range | `v-edit -p "rename foo" -f main.go --start 12 --end 24` |
| `gitmessage` | Draft a commit message from staged changes | `gitmessage feat -c` |
| `record` | List, inspect, or export saved runs | `record list --command ask` |
| `askwrapper` | Ask with an interactive history browser | `askwrapper ui ask` |

`ask`, `explain`, and `tutor` print to stdout by default; use `--save` to save their output. `v-edit` prints updated text unless `--write` is given. `gitmessage` reads staged changes. `record cat 1` prints a saved entry; `record export` exports to Markdown (interactive unless `--file` is specified).

Use `--history` on AI commands to save JSON runs under `.spicy/<command>/`, with a matching `.md` file containing each returned result. `askwrapper` maintains its own history at `~/.askwrapper/history.json`. In its prompt, enter `:N` to preview an entry, `:dN` to delete one, or an empty response to cancel. `askwrapper ui followup` asks a question using an earlier entry as context.

Run `<command> --help` for flags and more examples.

## Models and configuration

The CLI and Neovim plugin share `~/.config/spicy/config.json`. For example:

```json
{
  "default_model": "openai/gpt-5.3-codex",
  "models": {
    "ask": "openai/gpt-5.3-codex"
  }
}
```

A `.spicy.config.json` in the current directory or nearest parent directory overrides the global config, merged per key. Non-empty local `default_model` and `models.<command>` values override matching global values; other global command models remain available. Within the merged config, a command model takes priority over the default. A command-line `-m` / `--model` takes priority over both. Missing, unreadable, or invalid local config is ignored; without configuration, the CLI falls back to `openai/gpt-5.3-codex`.

See [Neovim configuration](../nvim/README.md#configuration) for plugin-specific precedence.

## Development

```sh
make build-all      # outputs to cli/bin/
make test           # go test ./...
make check          # go vet ./... and go test ./...
make fmt            # go fmt ./...
```

Use `make build-ask`, `make build-v-edit`, or `make install-history` for an individual tool. The `v-edit` and `record` executables are built from `cmd/ctx-edit` and `cmd/history`, respectively.
