# ImgWalker GUI Requirements

### `IMGWALKER-001`

The startup window configuration must define title `ImgWalker` and size `640x360` density-independent pixels.

### `IMGWALKER-002`

The startup greeting UI copy constant must be exactly `Hello, World!`.

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

### `IMGWALKER-011`

When `imageDir` starts with `~/`, startup validation must expand it to an absolute path under the current user home directory before filesystem checks.

### `IMGWALKER-012`

When `imageDir` is a relative path, startup validation must resolve it to an absolute path using the current working directory before filesystem checks.

### `IMGWALKER-013`

When startup config loading or validation fails, ImgWalker must continue startup using an empty image directory setting.
