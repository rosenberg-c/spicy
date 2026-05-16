# Engineering Rules

These rules are language-agnostic defaults for this repository.

Project-specific rules live in `docs/PROJECT_RULES.md`.
When a generic rule and a project rule conflict, the project rule takes precedence for this project.

## 1. Follow existing conventions first

- match established project style and structure
- reuse existing patterns before introducing new ones

---

## 2. Keep changes small and scoped

- implement the smallest effective change
- avoid unrelated refactors in feature work
- preserve existing behavior unless change is intentional

---

## 3. Keep boundaries explicit

Separate and map at boundaries:

- transport/API models
- domain models
- persistence models
- UI/view models

Avoid leaking framework or generated types across layers.

---

## 4. Prefer standard patterns over custom abstractions

- use official docs and library-native patterns
- avoid introducing abstraction before it is needed

---

## 5. Design for testability

- inject dependencies at boundaries
- avoid hidden global state in business logic
- keep logic deterministic where practical

---

## 6. Validate at boundaries

- validate input before mutating state
- enforce invariants close to entry points
- return explicit, stable domain errors

---

## 7. Treat storage and network as failure-prone

- assume I/O can fail or be slow
- wrap errors with clear context
- avoid unnecessary full read-modify-write cycles

---

## 8. Be explicit about concurrency

- define ownership for background work
- document start/stop/error behavior
- protect shared mutable state

---

## 9. Logging does not replace error handling

- return errors to callers when possible
- log at boundaries (service entrypoints)
- never log secrets or credentials

---

## 10. Generated code is read-only

- do not edit generated files manually
- update source schema/config and regenerate

---

## 11. Test behavior, not implementation details

- test observable behavior and contracts
- keep tests fast, deterministic, and focused
- add integration coverage for boundary-critical flows

---

## 12. Project-specific rules apply

- use `docs/PROJECT_RULES.md` for project-specific requirements
- keep this file generic; put implementation-specific policy in project rules

---

## 13. Allow small adjacent maintenance, discuss broad refactors

- include minor, low-risk adjacent cleanup that improves touched areas
- keep adjacent maintenance directly related and behavior-preserving by default
- discuss medium/large scope refactors before implementation

---

## 14. Group repeated UI state passed through multiple layers

- when the same 3+ related fields are repeatedly passed through layers, prefer a small grouped value type
- keep grouped types focused and feature-scoped
- avoid wrapper types for one-off cases where direct parameters are clearer

---

## 15. Accessibility labels and selectors for symbolic controls

- prefer stable selectors for automated UI tests on interactive controls and key containers
- when visible text is symbolic or abbreviated (icons/arrows), provide explicit accessibility labels
- keep display text and accessibility text separate when needed

---

## 16. Prefer selector-first UI tests with documented fallback

- write UI tests using stable selectors first, then role/text fallback only when selectors are unavailable or unreliable
- if fallback selectors are required, document the reason in the test for maintainability
- keep fallback usage localized and avoid broad text-only matching across repeated controls

For TypeScript/React-specific defaults, see `docs/RULES_TYPESCRIPT.md`.
