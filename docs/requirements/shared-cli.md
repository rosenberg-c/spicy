# Shared CLI Foundations Requirements

### `CORE-CLI-001`

Every CLI command must accept direct arguments for primary input without
forcing an interactive prompt.

### `CORE-CLI-002`

Commands that persist outputs must write files atomically.

### `CORE-CLI-003`

History persistence must never fail silently.

### `CORE-CLI-004`

User-facing validation errors must be actionable.
