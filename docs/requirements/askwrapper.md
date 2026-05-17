# Askwrapper Command Requirements

### `CLI-ASKWRAPPER-001`

`askwrapper ui ask` must run an interactive ask flow.

### `CLI-ASKWRAPPER-002`

`askwrapper ui followup` must run an interactive follow-up flow that uses a selected history entry as context.

### `CLI-ASKWRAPPER-003`

Askwrapper history must be stored at `~/.askwrapper/history.json`.

### `CLI-ASKWRAPPER-004`

Askwrapper history writes must prepend the newest entry first.

### `CLI-ASKWRAPPER-005`

`askwrapper ui ask` and `askwrapper ui followup` must accept a `--timeout` option in seconds for ask execution.

### `CLI-ASKWRAPPER-006`

Follow-up prompt construction must include previous question, previous answer, and new follow-up question.

### `CLI-ASKWRAPPER-007`

The terminal askwrapper flow must support history preview by index using `:N`.

### `CLI-ASKWRAPPER-008`

The terminal askwrapper flow must support history deletion by index using `:dN`.

### `CLI-ASKWRAPPER-009`

The Gio askwrapper flow must support selecting history entries and previewing selected content.

### `CLI-ASKWRAPPER-010`

The Gio askwrapper flow must disable question input while ask execution is running.

### `CLI-ASKWRAPPER-011`

The Gio askwrapper flow must disable history interactions while ask execution is running.

### `CLI-ASKWRAPPER-012`

The Gio askwrapper flow must provide history deletion via explicit UI action and keyboard shortcut.

### `CLI-ASKWRAPPER-013`

The Gio askwrapper flow must provide an in-app mode switch between ask mode and follow-up mode.

### `CLI-ASKWRAPPER-032`

The Gio askwrapper mode switch must be rendered as mutually exclusive radio options for `Ask` and `Follow-up`.

### `CLI-ASKWRAPPER-014`

In follow-up mode, Gio askwrapper must keep submit disabled until a history context is selected.

### `CLI-ASKWRAPPER-015`

The Gio askwrapper flow must display inline helper text that communicates current mode actions and available delete shortcuts.

### `CLI-ASKWRAPPER-016`

After deleting a history entry in Gio askwrapper, deletion is immediate and no undo action is provided.

### `CLI-ASKWRAPPER-017`

The Gio askwrapper flow must provide a cancel action that stops an in-flight ask request.

### `CLI-ASKWRAPPER-018`

The Gio askwrapper preview panel must be scrollable when content exceeds available viewport height.

### `CLI-ASKWRAPPER-019`

The Gio askwrapper primary action must use a single button that switches label/behavior between `Ask` and `Cancel` based on run state.

### `CLI-ASKWRAPPER-020`

History deletion in Gio askwrapper must be exposed per history row and not as a global delete-selected action.

### `CLI-ASKWRAPPER-021`

Operational messages (for example running/cancelling/deleted/error status) must render in the lower status area and must not replace preview content.

### `CLI-ASKWRAPPER-022`

While ask execution is running, the status area must display an animated braille spinner indicator.

### `CLI-ASKWRAPPER-023`

In Gio askwrapper, the question input field and primary submit/cancel control must appear on the same row.

### `CLI-ASKWRAPPER-024`

When not running, the Gio askwrapper primary submit control must use a right-arrow label instead of text `Ask`.

### `CLI-ASKWRAPPER-025`

Hammerspoon integration must only launch askwrapper commands and must not embed ask execution or history logic.

### `CLI-ASKWRAPPER-026`

Hammerspoon hotkey `alt+shift+A` must launch `askwrapper ui ask`.

### `CLI-ASKWRAPPER-027`

Hammerspoon hotkey `alt+shift+S` must launch `askwrapper ui followup`.

### `CLI-ASKWRAPPER-028`

When the Gio askwrapper window opens, keyboard focus must be set to the question input field.

### `CLI-ASKWRAPPER-029`

When the Gio askwrapper window is closed while an ask request is running, the in-flight ask request must be canceled.

### `CLI-ASKWRAPPER-030`

If ask execution succeeds but history append fails, Gio askwrapper must keep the produced answer visible in preview and show a warning status message.

### `CLI-ASKWRAPPER-031`

After deleting a history row via keyboard activation of that row's delete control, keyboard tab focus must continue from the next available row delete control (or previous row if the last row was deleted), rather than resetting to the top-level controls.
