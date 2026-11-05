# History Command Requirements

### `CORE-HIST-001`

History entries must be listed in reverse chronological order.

### `CORE-HIST-002`

History command should support stable filtering modes.

### `CORE-HIST-003`

Owner: `cli/internal/history/history.go`, `cli/internal/history/history_test.go`

For every saved `.spicy` history entry with a string `data.result`, save that result to a `.md` file alongside the JSON entry with the same base name. Entries without a string result must not create a companion file. Report a companion write failure as a history save error.
