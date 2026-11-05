# Context Edit Command Requirements

### `CLI-CTX-001`

Context edit must preserve non-target file content.

### `CLI-CTX-002`

File write behavior should surface clear write failures.

### `CLI-CTX-003`

The `--file` path must be resolved to an absolute path before the file is read
or written, and the `file_path` reported in JSON output must be absolute.
