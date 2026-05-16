# Requirements

This document is the entry point for product requirements in this repo.
Detailed requirements are split by domain under `docs/requirements/`.

## Conventions

- Requirement IDs are globally unique and stable.
- Requirement IDs must not be recycled.
- Test tags should reference requirement IDs using `@req` markers in comments.
  Example: `// @req CLI-ASK-001, CORE-HIST-001`.
- Every observable behavior change should map to one or more requirement IDs
  before implementation.

## Domains

- [Shared CLI Foundations](requirements/shared-cli.md)
- [Ask Command](requirements/ask.md)
- [Explain Command](requirements/explain.md)
- [Tutor Command](requirements/tutor.md)
- [Gitmessage Command](requirements/gitmessage.md)
- [Context Edit Command](requirements/ctx-edit.md)
- [History Command](requirements/history.md)
- [Neovim Plugin](requirements/nvim.md)
- [Hammerspoon Ask Wrapper](requirements/hammerspoon.md)
