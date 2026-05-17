# ImgWalker GUI Requirements

### `IMGWALKER-001`

The startup window configuration must define title `ImgWalker` and size `640x360` density-independent pixels.

### `IMGWALKER-002` (Deprecated)

Deprecated/removed. The previous startup greeting (`Hello, World!`) requirement no longer applies because ImgWalker now starts directly in the image browser UI.

### `IMGWALKER-003`

The startup theme configuration must set palette values exactly to: `Bg(16,21,28,255)`, `Fg(222,228,236,255)`, `ContrastBg(41,109,196,255)`, and `ContrastFg(248,251,255,255)`.

### `IMGWALKER-004`

On startup, ImgWalker must load JSON config from `$HOME/.config/spicy/imgwalker.json`.

### `IMGWALKER-005`

The loaded config must expose an `imageDir` string field used as the image directory setting.

### `IMGWALKER-006`

When `$HOME/.config/spicy/imgwalker.json` does not exist, startup config loading must return a typed `not found` error category.

### `IMGWALKER-007`

When `$HOME/.config/spicy/imgwalker.json` contains invalid JSON, startup config loading must return a typed `invalid config` error category.

### `IMGWALKER-008`

After config load, startup validation must treat an empty `imageDir` value as a typed `invalid config` error category.

### `IMGWALKER-009`

After config load, startup validation must treat a non-existent `imageDir` path as a typed `invalid image dir` error category.

### `IMGWALKER-010`

After config load, startup validation must treat an `imageDir` path that exists but is not a directory as a typed `invalid image dir` error category.

### `IMGWALKER-010-A`

After config load, startup validation must treat non-`not found` filesystem errors while checking `imageDir` as a typed `invalid image dir` error category.

### `IMGWALKER-011`

When `imageDir` starts with `~/`, startup validation must expand it to an absolute path under the current user home directory before filesystem checks.

### `IMGWALKER-012`

When `imageDir` is a relative path, startup validation must resolve it to an absolute path using the current working directory before filesystem checks.

### `IMGWALKER-013`

When startup config loading or validation fails, ImgWalker must continue startup using an empty image directory setting.

### `IMGWALKER-014`

When `imageDir` is configured, the main view must render a two-pane layout with an image list pane on the left and a preview pane on the right.

### `IMGWALKER-015`

The left pane must list image file names discovered from `imageDir`, including only files with extensions `.png`, `.jpg`, `.jpeg`, `.gif`, `.webp`, and `.bmp` (case-insensitive).

### `IMGWALKER-015-A`

The left pane image list must be sorted in ascending filename order.

### `IMGWALKER-016`

When no image files are discovered in `imageDir`, the left pane must show an explicit empty-state message.

### `IMGWALKER-017`

In the left image list, pressing `j` or `DownArrow` must move selection down by one item and pressing `k` or `UpArrow` must move selection up by one item, clamped to list bounds.

### `IMGWALKER-018`

The left image list must visually differentiate the selected item with a distinct row background and foreground color from non-selected items.

### `IMGWALKER-019`

When an image is selected in the left pane, the right preview pane must display that image's full filesystem path.

### `IMGWALKER-020`

When reading `$HOME/.config/spicy/imgwalker.json` fails for reasons other than missing file, startup config loading must return a typed `io_error` error category.

### `IMGWALKER-021`

When no image files are discovered in `imageDir`, the right preview pane must render empty text.

### `IMGWALKER-022`

Clicking an image item in the left pane must update selection to that item.

### `IMGWALKER-023`

The left list pane background must use the base app background palette and must not use the contrast-accent pane background.

### `IMGWALKER-024`

The main view must render a visible vertical delimiter between the left list pane and the right preview pane.

### `IMGWALKER-025`

When one or more images are discovered in `imageDir`, the right preview pane must default to showing the first image path when current selection is invalid.

### `IMGWALKER-026`

When reading image entries from `imageDir` fails at runtime, the left pane must render a lightweight inline error message and continue rendering the empty-state/list area.
