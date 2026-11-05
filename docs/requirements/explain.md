# Explain Command Requirements

### `CLI-EXP-001`

`explain` must read source from files and preserve content.

### `CLI-EXP-002`

Directory mode must skip hidden/vendor/generated noise paths.

### `CLI-EXP-003`

Language detection should infer from known extensions first.

### `CLI-EXP-004`

Owner: `cli/internal/history/history.go`, `cli/internal/history/history_test.go`

When saving an `explain` history entry, its `data.result` companion Markdown file must contain the returned explanation rather than the input `data.code`, regardless of source language (see `CORE-HIST-003`).
