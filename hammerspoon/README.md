# Spicy Hammerspoon Module

Hotkeys for launching the Go-native `askwrapper` UI flows from Hammerspoon.

## Requirements

- Hammerspoon
- Spicy CLI in your `PATH` (`askwrapper` and `ask` commands)

## Setup

1. Make the module available to Hammerspoon:
   - Option A: copy `hammerspoon/modules/askwrapper.lua` into `~/.hammerspoon/modules/`
   - Option B: symlink the repo `hammerspoon/modules` into `~/.hammerspoon/modules/`
2. Add this to `~/.hammerspoon/init.lua`:

```lua
local askwrapper = require("modules.askwrapper")
askwrapper.setup()
```

Reload Hammerspoon after saving.

## Hotkeys

- `alt+shift+A` - launch `askwrapper ui ask` directly
- `alt+shift+S` - launch `askwrapper ui followup` directly

## History

History is stored at `~/.askwrapper/history.json` and managed by the Go UI.

## Notes

- Hammerspoon is only responsible for hotkeys + launching askwrapper.
- All ask/follow-up UI behavior now lives in the Go `askwrapper` command.
