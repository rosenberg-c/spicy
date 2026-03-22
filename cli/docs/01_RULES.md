# Go Repository & Storage Rules

## 1. Never take addresses of range loop variables

- Bad: for \_, v := range slice { return &v }
- Good: Use index-based access: for i := range slice { return &slice[i] }

## 2. Build file paths with filepath.Join

- Bad: String concatenation for paths
- Good: filepath.Join(basePath, "file.json")

## 3. Prefer values over pointers for loaded data

- Return structs unless you truly need shared mutable state.
- If returning pointers, ensure they point to stable memory (slice element, not
  loop variable).

## 4. Guard against duplicate data on create

- Validate uniqueness (IDs, usernames) before appending.
- Return explicit domain errors (ErrAlreadyExists).

## 5. Assume persistence is slow and fragile

- Minimize full read–modify–write cycles.
- Expect failures on load and save; wrap errors with context.

## 6. Protect read–modify–write with locking

- If concurrent access is possible, use sync.Mutex (repo or store level).
- JSON-file storage is not concurrency-safe by default.

## 7. Keep repositories testable

Prefer dependency injection:

- Bad: NewRepo(path string)
- Good: NewRepo(store Store)
- Avoid hardcoding file paths inside logic.

Anything that interacts with the environment is behind an interface, so no
mocking is needed—use real implementations in tests.

## 8. Document exported symbols

- Every exported type or function should have a GoDoc comment only when it adds
  information not obvious from the name or signature.

- Comments must explain non-obvious behavior, edge cases, guarantees, or reasons
  (“why”), not restate the function name (“what”).

Guidelines:

- Avoid comments that merely paraphrase the identifier.
- Prefer documenting:
  - error conditions and guarantees
  - side effects
  - performance or concurrency implications
  - invariants and assumptions

Examples:

- Bad: // GetByID gets by ID
- Good: // GetByID returns ErrNotFound if no user with the given ID exists.

## 9. Nil slices are valid

- Returning nil slices is idiomatic.
- Only normalize to empty slices if required for JSON output or API contracts.

## 10. Design for future growth

- Consider context.Context in method signatures.
- Treat file-backed repos as prototypes, not scalable storage.

## 11. Keep lines under 80 characters

- Wrap long strings, especially in prompts, error messages, and help text.
- Use multiline strings or concatenation for readability.

Examples:

- Bad: `fmt.Sprintf("You are a senior coder: write a short commit message, one row only. Do not include the actual diff, or any other thoughts, only the commit message. Always use Capital character at the beginning of the commit message.")`

- Good:
  ```go
  fmt.Sprintf(`You are a senior coder: write a short commit message, one row only.
  Do not include the actual diff, or any other thoughts, only the commit message.
  Always use Capital character at the beginning of the commit message.`)
  ```

- For format strings, break at natural boundaries (sentences, clauses)
- For function calls with many args, use one arg per line when exceeding 100 chars
