# ImgWalker

Gio desktop app for browsing and managing images. See [requirements](../docs/requirements/imgwalker.md) for supported behavior.

Requires Go 1.23.8+ and [Gio's platform dependencies](../cli/README.md#install). Set the image directory in `~/.config/spicy/imgwalker.json`:

```json
{"imageDir": "~/Pictures"}
```

From this directory:

```sh
make build             # builds ./imgwalker
./imgwalker            # launches the app
make test
make install           # installs to ~/.local/bin
```

Use `INSTALL_DIR=/path/to/bin` to change the install destination, or run `make install-imgwalker` from the repository root.
