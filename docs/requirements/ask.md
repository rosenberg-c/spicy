# Ask Command Requirements

### `CLI-ASK-001`

`ask` joins positional arguments into one prompt in order.

### `CLI-ASK-002`

`ask` prompt construction must include both system guidance and raw user input.

### `CLI-ASK-003`

Empty argument mode may read from stdin interactively.

### `CLI-ASK-004`

`ask` must accept optional context via `--context` (inline) or
`--context-file` (path). The two flags are mutually exclusive and specifying
both is an error. When context is present it is included in the prompt before
the user input; a missing, empty, or blank-only context must produce the
legacy prompt unchanged.
