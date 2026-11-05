# Hammerspoon Ask Wrapper Requirements

### `HAM-ASK-001`

Ask wrapper must expose a setup entrypoint for hotkey binding.

### `HAM-ASK-002`

Ask wrapper should keep a lightweight status indicator while a request is
running.

### `HAM-ASK-003`

Sublime ask chooser must include history-backed list behavior with newest
entries first.

### `HAM-ASK-004`

History-backed list items must show question text and a preview of the stored
answer.

### `HAM-ASK-005`

History-backed list interaction must support keyboard navigation and opening a
selected history result.

### `HAM-ASK-006`

History-backed list interaction must support deleting a selected history item.

### `HAM-ASK-007`

Ask wrapper must expose two user entry modes with dedicated hotkeys: iTerm ask
launcher and Sublime ask workflow.

### `HAM-ASK-008`

History data must persist to local storage and load on chooser open with
newest-first insertion semantics.

### `HAM-ASK-009`

History answer preview text must normalize whitespace and truncate long output
to keep chooser rows readable.

### `HAM-ASK-010`

Ask task lifecycle must always stop spinner UI on both success and failure
paths.

### `HAM-ASK-011`

User-visible error feedback must be surfaced for missing dependencies and ask
execution failures.

### `HAM-ASK-012`

Ask output processing must remove internal marker preamble and provide fallback
text when no response content is produced.
