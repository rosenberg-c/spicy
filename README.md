# Spicy

Spicy is a set of developer tools built around OpenCode. It includes Go command-line tools, a Neovim plugin, Hammerspoon hotkeys, and an image browser. This is an alpha project; interfaces may change.

| Component | What it does | Guide |
| --- | --- | --- |
| `cli/` | Ask questions, explain code, generate tutorials, edit snippets, draft commit messages, and browse history | [CLI](cli/README.md) |
| `nvim/` | Run Spicy commands from Neovim | [Neovim plugin](nvim/README.md) |
| `hammerspoon/` | Launch ask flows and ImgWalker with macOS hotkeys | [Hammerspoon module](hammerspoon/README.md) |
| `imgwalker/` | Browse and manage images with a Gio desktop UI | [ImgWalker](imgwalker/README.md) |

## Get started

Install the CLI tools to `~/.local/bin` (requires Go 1.23.8+ and [OpenCode](https://opencode.ai/)):

```sh
make install
```

Add `~/.local/bin` to your `PATH`, then see the component guides for usage and optional integrations. To install the Gio desktop variant of `askwrapper`, use `make install-gio` instead. Install the image browser separately with `make install-imgwalker` (requires Gio's platform dependencies).

## Development

Run `make test` for CLI and Neovim tests, or `make test-cli` / `make test-nvim` individually. ImgWalker has its own tests: `make -C imgwalker test`.

Behavior changes follow the [requirements](docs/REQUIREMENTS.md) and [test matrix](docs/TEST_MATRIX.md). Regenerate the matrix with `make sync-test-matrix`.

Spicy was inspired by [ThePrimeagen/99](https://github.com/ThePrimeagen/99), with a focus on learning and understanding the code we write.
