# Spicy Hammerspoon Module

Hotkeys for running `ask` from Hammerspoon, with a lightweight history browser
and optional Sublime Text output.

## Requirements

- Hammerspoon
- Spicy CLI in your `PATH` (`ask` command)
- iTerm (for the terminal hotkey)
- Sublime Text + `subl` on your `PATH` (for the Sublime hotkey)

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

- `alt+shift+A` - prompt, run `ask` in a new iTerm window
- `alt+shift+S` - prompt, run `ask`, open the response in Sublime Text

## History

History is stored at `~/.askwrapper/history.json`.

In the Sublime picker:

- arrow keys navigate history
- `backspace`, `delete`, or `Ctrl+D` removes the selected entry

## Notes

- The Sublime integration looks for `subl` in common locations and will warn
  if it cannot be found.
- The module runs `ask` via your shell, so your normal CLI auth applies.
