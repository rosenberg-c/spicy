# Project Rules

These rules are specific to this project and take precedence over generic guidance in `docs/RULES.md` and `AGENT.md` when applicable.

## 1. Requirements-first for behavior changes

- every new feature or observable behavior change must map to requirement IDs before implementation
- update `docs/requirements/*.md` first, then implement code and tests
- keep automated tests tagged with requirement references using `@req` comments
- keep `docs/TEST_MATRIX.md` synchronized via `make sync-test-matrix`

---

## 2. Preserve requirement tag conventions

- keep test requirement annotations using `@req` comment markers
- keep requirement IDs globally unique and stable; do not recycle IDs

---

## 3. Keep generated artifacts read-only

- do not hand-edit generated outputs
- update source definitions/config and regenerate
- document the generator command in change notes when behavior depends on generated updates

---

## 4. Go code safety defaults

- never take the address of range loop variables; index into slices when pointer identity is required
- use `filepath.Join` for filesystem paths
- pass `context.Context` through I/O or request-scoped boundaries; do not store context on structs
- return errors (with wrapping context) instead of panicking for expected failures

---

## 5. Logging and secret handling constraints

- log at command/service boundaries, not as a replacement for returned errors
- never log credentials, tokens, API keys, or provider secrets
- keep user-facing error messages free of sensitive internals

---

## 6. Keep CLI, Neovim, and Hammerspoon behavior aligned with requirements

- when command behavior changes, update matching requirement docs and tests in the same change
- keep plugin integrations aligned with documented command flags, output formats, and history behavior
- call out intentional cross-surface differences explicitly in requirement docs

---

## 7. Consolidate UI copy in code

- for user-facing UI strings, prefer centralized constants/helpers per feature instead of scattered literals
- avoid duplicate text literals across handlers/render paths when they represent the same message
- when UI copy changes behavior expectations, update matching requirement docs in the same change

---

## 8. UI event-loop mutation safety

- do not mutate slice-backed UI state while iterating related event/click handler slices in the same pass
- when a UI event implies deletion/reload of list data, capture intent first and apply mutation after iteration, or exit iteration immediately after mutation
- prefer defensive bounds checks when list length can change due to actions in the same frame

---

## 9. Prevent duplicate mutation actions in UI flows

- disable mutation-triggering controls immediately after a mutation starts
- keep related controls disabled until mutation resolution (success or failure)
- avoid enqueuing duplicate requests from repeated clicks/keypresses
- add/adjust tests for disabled -> enabled transitions when mutation state is user-visible

---

## 10. Use typed error categories across command boundaries

- map domain/command failures using stable typed categories or sentinels, not raw error string matching
- keep user-facing copy decoupled from internal error detail
- when adding a new error category, add tests for both category emission and mapped behavior

---

## 11. Model ordered/batch updates with collection semantics

- when changing ordered collections (history, lists, grouped entries), model the target list state first
- validate list invariants at boundaries (membership, duplicates, stable ordering behavior)
- avoid parallel one-off mutation paths that diverge from collection-level behavior
