# Why Go packages “require” at least one package comment

In plain Go (the compiler and `go` tool), **a package does not strictly require a package comment** to build, test, or run.  
When people say “Go packages require at least one package comment,” they’re usually referring to **documentation and linting conventions**—most commonly a lint rule such as:

- `golint`: `should have a package comment, unless it's in another file for this package`
- `golangci-lint` (via `golint`, `revive`, etc.): similar warnings

So the “requirement” is typically: **your project’s linting / CI checks fail unless there is at least one package-level doc comment**.

---

## What is a “package comment” in Go?

A *package comment* is a doc comment that documents the package as a whole. The canonical form is:

```go
// Package widgets provides tools for building widget pipelines.
package widgets
```

Key properties:

- It is a Go doc comment (starts with `//` or `/* ... */`) that appears **immediately before** a `package <name>` clause (often in a dedicated `doc.go`).
- By convention, it starts with `Package <name> ...` so `go doc`, `godoc`, and pkg.go.dev can recognize it as package documentation.

---

## Why linters insist on at least one package comment

### 1) Go documentation tooling is package-centric
Go’s documentation tooling (`go doc`, historical `godoc`, and pkg.go.dev) is organized around packages. When users land on a package page, the very first thing they should see is:

- what the package is for
- how to use it (often a short overview)
- any important constraints (thread-safety, performance, OS-specific behavior, etc.)

A package comment provides that “front door” explanation.

Without it, generated docs tend to look like a list of exported identifiers with no narrative, which is much harder for users to understand.

### 2) It’s part of Go’s “exported things should be documented” philosophy
Go has a strong convention:

- **Exported identifiers** (types, funcs, vars, consts) should have doc comments.
- The **package itself** is effectively an exported unit of API, so it should have a doc comment too.

Linters enforce this because otherwise public packages drift into being “self-documenting” by name alone—which rarely holds up.

### 3) It improves discoverability and search
Pkg.go.dev and similar tools index package documentation. A good package comment:

- makes the package searchable by intent (not just identifier names)
- clarifies the meaning of ambiguous package names
- helps readers decide quickly if they’re in the right place

### 4) It’s a low-effort, high-value signal
A single short comment can prevent repeated confusion:

- “Is this package internal or stable?”
- “What’s the intended entry point?”
- “What does it *not* do?”
- “Any gotchas?”

Linters treat it as a cheap baseline quality bar.

---

## Common misconception: “Go requires it”
Again: **the Go compiler does not require package comments**. This will compile fine:

```go
package widgets

func New() {}
```

But your linter/CI might fail, or your documentation site might show an empty description. That’s why it *feels* required in many codebases.

---

## Where to put the package comment (best practices)

### Option A (recommended): create a `doc.go`
This is the most common pattern in larger packages.

`doc.go`:

```go
// Package widgets provides tools for building widget pipelines.
//
// The widgets package is safe for concurrent use by multiple goroutines.
package widgets
```

Why `doc.go` is popular:

- keeps package docs in one obvious place
- avoids burying the package overview in a random file
- works well when other files start with imports, build tags, or large blocks of code

### Option B: put it at the top of an existing file
Fine for small packages:

```go
// Package widgets provides tools for building widget pipelines.
package widgets

// ...
```

---

## How “at least one package comment” is interpreted

Many lint rules are satisfied if **any file in the package** contains a proper package comment for that package. In practice:

- It must be a doc comment immediately preceding a `package` clause.
- It usually should start with `Package <name>` to be recognized as package documentation.

This comment **does not count** as a package comment:

```go
// Utilities for widgets.   // (doesn’t start with "Package widgets" in strict linters)
package widgets
```

Some tools accept it, but the most widely accepted form is the standard “Package …” wording.

---

## Examples: good package comments

### Minimal, lint-satisfying
```go
// Package widgets provides widget utilities.
package widgets
```

### Better: one-paragraph overview + key guarantees
```go
// Package widgets implements a small toolkit for building and executing widget pipelines.
//
// It provides a registry, common pipeline stages, and helpers for composing stages.
// All exported types are safe for concurrent use unless otherwise noted.
package widgets
```

### Include a tiny usage example (often helpful)
```go
// Package widgets provides widget pipelines.
//
// Basic usage:
//
//	p := widgets.NewPipeline(widgets.StageA(), widgets.StageB())
//	out, err := p.Run(ctx, in)
package widgets
```

(Indented example blocks in comments are commonly used by `go doc` and documentation renderers.)

---

## Edge cases and special situations

### `main` packages
Linters vary. Many teams don’t require package comments for `package main`, because it’s an executable, not a reusable library. Others still want it.

If your linter complains, this is usually enough:

```go
// Command mytool processes widget definitions.
package main
```

Notice the convention for commands is often “Command …” rather than “Package …”, but some linters still prefer `Package main ...`. Check your lint config.

### `internal` packages
Even internal packages benefit from docs. Linters often still require package comments because internal packages can be widely used inside a repo.

### Packages with only tests
If a directory contains only `_test.go` files with `package foo_test`, you might run into confusing lint behavior depending on tooling. Typically, you document the actual package (`foo`) in its non-test files; `foo_test` is a separate package used for external black-box tests.

### Build tags at the top of files
If you use build constraints, remember the build tag must appear before the package clause, but you can still place the package comment right before `package`:

```go
//go:build darwin

// Package widgets provides macOS-specific widget implementations.
package widgets
```

---

## Practical “why” in one sentence

Go packages don’t *technically* require package comments, but **Go’s docs ecosystem and many lint configurations treat a package comment as required because it’s the primary entry point for understanding and using the package**.

---

## Quick checklist (to satisfy linters and help readers)

- Comment is immediately before `package <name>`
- Starts with `Package <name> ...` (most compatible)
- First sentence is a crisp description
- Add a second paragraph only if there are important semantics (thread-safety, performance, invariants)
- Prefer a dedicated `doc.go` for non-trivial packages

---

If you tell me which tool is complaining (e.g., `golint`, `revive`, `golangci-lint`) and the exact warning text, I can show the smallest comment that will satisfy it and match that tool’s expectations.
