# Lua Repository & Engineering Rules

These rules define conventions for repository, persistence, and general
engineering patterns in Lua projects. They are intended to keep code
predictable, testable, and easy to evolve.

---

# 1. Never rely on loop variable identity

Loop variables in generic `for` loops are reassigned each iteration.
Closures that capture them may observe unexpected values.

Bad:

```lua
for _, user in ipairs(users) do
    callbacks[#callbacks + 1] = function()
        return user
    end
end
```

Good:

```lua
for i = 1, #users do
    local user = users[i]

    callbacks[#callbacks + 1] = function()
        return user
    end
end
```

---

# 2. Build file paths with a helper

Avoid manual path concatenation in core logic.

Bad:

```lua
local path = base_dir .. "/" .. name .. ".json"
```

Good:

```lua
local path = fs.join(base_dir, name .. ".json")
```

Centralizing path logic avoids platform issues and duplication.

---

# 3. Prefer values over shared references

Return new tables unless shared mutable state is intentional.

Bad:

```lua
function Repo:list()
    return self.items
end
```

Good:

```lua
function Repo:list()
    local out = {}

    for i = 1, #self.items do
        out[i] = self.items[i]
    end

    return out
end
```

If returning internal tables intentionally, document that callers receive
a mutable reference.

---

# 4. Guard against duplicate data

Validate uniqueness before inserting new records.

Example:

```lua
if self.by_id[user.id] ~= nil then
    return nil, ErrAlreadyExists
end
```

Domain errors should be explicit and machine-checkable.

---

# 5. Assume persistence is slow and fragile

Loading and saving may fail. Always propagate errors with context.

Bad:

```lua
local ok, err = store:save(data)
if not ok then
    return nil, err
end
```

Good:

```lua
local ok, err = store:save(data)

if not ok then
    return nil, ("save users: %s"):format(err)
end
```

---

# 6. Protect read–modify–write operations

File-backed storage is not concurrency-safe.

If concurrent access is possible, serialize writes using a lock,
queue, or single writer process.

Never assume single-threaded Lua automatically guarantees safety.

---

# 7. Keep repositories testable

Prefer dependency injection.

Bad:

```lua
Repo.new(path)
```

Good:

```lua
Repo.new(store)
```

Inject dependencies such as:

* stores
* clocks
* ID generators
* loggers

Avoid hardcoding environment access.

---

# 8. Document public APIs when needed

Comments should explain non-obvious behavior.

Document:

* error conditions
* guarantees
* side effects
* ordering guarantees
* concurrency implications

Bad:

```lua
-- get_by_id gets a user by id
```

Good:

```lua
-- Returns nil and ErrNotFound when the user does not exist.
```

Avoid comments that restate the function name.

---

# 9. Nil tables are valid

Returning `nil` instead of `{}` is idiomatic in Lua.

Normalize to empty tables only when required by:

* JSON encoding
* API contracts
* external integrations

Consistency matters more than which option is chosen.

---

# 10. Design for future growth

File-backed repositories are prototypes.

Structure code so storage can later be replaced with:

* SQLite
* Redis
* Postgres
* HTTP services

Keep interfaces small and stable.

---

# 11. Keep lines under 80 characters

Wrap long strings and argument lists.

Bad:

```lua
local msg = "Write a short commit message in one line only. Do not include the diff or reasoning."
```

Good:

```lua
local msg =
    "Write a short commit message in one line only. " ..
    "Do not include the diff or reasoning."
```

Break strings at natural sentence boundaries.

---

# 12. Separate domain logic from storage logic

Repositories manage persistence, not business rules.

Bad:

```lua
function UserRepo:create(user)
    if #user.password < 12 then
        return nil, ErrWeakPassword
    end
end
```

Good:

```lua
local ok, err = User.validate(user)

if not ok then
    return nil, err
end

return repo:create(user)
```

---

# 13. Use stable IDs instead of indexes

Array positions are not durable identities.

Prefer explicit identifiers:

* `user.id`
* `session.token`
* `project.slug`

---

# 14. Maintain ordered and indexed views

Use arrays for ordering and maps for lookup.

Example:

```lua
self.items
self.items_by_id
```

Avoid repeatedly scanning arrays for key lookups.

---

# 15. Keep module boundaries clear

Each module should have one responsibility.

Example structure:

```
user.lua
user_repo.lua
json_store.lua
path.lua
```

Avoid large generic modules such as `utils.lua`.

---

# 16. Prefer explicit constructors

Construct domain objects through named functions.

Bad:

```lua
return {
    id = id,
    name = name,
}
```

Better:

```lua
return User.new(id, name)
```

Constructors preserve invariants and simplify refactoring.

---

# 17. Keep table shapes consistent

Objects of the same type should expose the same fields.

Bad:

```
some users have `active`, others do not
```

Good:

```
all users include `active`
```

Consistency simplifies reasoning and serialization.

---

# 18. Distinguish nil and false

Lua treats both as falsy but they have different meanings.

Use:

* `nil` → absent / unknown
* `false` → explicit negative value

Avoid mixing them arbitrarily.

---

# 19. Standardize domain errors

Use stable machine-readable error values.

Example:

```lua
ErrNotFound = "not_found"
ErrAlreadyExists = "already_exists"
ErrInvalid = "invalid"
```

Avoid relying on fragile freeform strings.

---

# 20. Return results and errors consistently

Prefer the standard Lua pattern:

```
return result, nil
return nil, err
```

Do not mix multiple error conventions within the same layer.

---

# 21. Reserve thrown errors for programmer mistakes

Expected failures should be returned, not thrown.

Throw errors only for:

* impossible states
* violated invariants
* internal misuse

---

# 22. Keep serialization isolated

JSON encoding and decoding should live in store modules,
not repositories.

Example layering:

```
repository -> store
store -> JSON library
```

---

# 23. Validate decoded data

Data loaded from disk is untrusted.

After decoding:

* validate required fields
* validate types
* validate invariants

Never assume persisted data always matches expectations.

---

# 24. Prefer atomic write patterns

Avoid overwriting the only copy of important data.

Safer approach:

1. write temporary file
2. flush
3. rename into place

This reduces corruption risk.

---

# 25. Version persisted formats

Include a version field when formats may evolve.

Example:

```lua
{
    version = 1,
    users = { ... }
}
```

Explicit versioning simplifies migrations.

---

# 26. Avoid hidden global state

Global mutable state makes code fragile.

Bad:

```lua
USERS = USERS or {}
```

Good:

```lua
local repo = UserRepo.new(store)
```

---

# 27. Keep metatables simple

Use metatables mainly for method dispatch.

Avoid complex behavior hidden in:

* `__index`
* `__newindex`

Prefer explicit functions.

---

# 28. Avoid mutating inputs

Treat function arguments as immutable unless explicitly documented.

Bad:

```lua
function create_user(attrs)
    attrs.id = attrs.id or new_id()
end
```

Better:

```lua
function create_user(attrs)
    local user = {
        id = attrs.id or new_id(),
        name = attrs.name,
    }
end
```

---

# 29. Prefer local scope

Declare helpers and constants as `local`.

Bad:

```lua
function helper()
end
```

Good:

```lua
local function helper()
end
```

---

# 30. Prefer explicit control flow

Repository and persistence code should be clear and predictable.

Avoid clever tricks that obscure behavior.

---

# 31. Be explicit about ordering

`pairs()` iteration order is not guaranteed.

If order matters, use arrays and numeric loops.

Bad:

```lua
for id, user in pairs(users) do
```

Good:

```lua
for i = 1, #users do
```

---

# 32. Avoid sparse arrays

Arrays with holes create unpredictable `#` behavior.

Prefer dense arrays or use maps when deletion is frequent.

---

# 33. Maintain one-way dependencies

Dependencies should flow downward:

```
domain -> repository -> store
```

Avoid circular module imports.

---

# 34. Make side effects visible

Functions that write files or mutate state should clearly
signal that behavior.

Prefer explicit names such as:

```
repo:save()
store:write_all()
```

---

# 35. Plan for contextual parameters

Lua lacks a built-in `context.Context`, but an options table can
provide:

* cancellation
* deadlines
* tracing
* logging metadata

Keep APIs flexible so this can be added later.

---

# 36. Use real implementations in tests

If environment access is abstracted, tests can often use:

* temporary directories
* real JSON encoding
* simple in-memory stores

Prefer this over heavy mocking.

---

# 37. Test persistence thoroughly

Repository tests should cover:

* create
* read
* update
* delete
* duplicate rejection
* persistence round trips
* empty stores
* malformed files

---

# 38. Keep public APIs small

Repositories typically expose:

* `get_by_id`
* `list`
* `create`
* `update`
* `delete`

Avoid exposing internal helpers.

---

# 39. Prefer clear names

Clarity is more important than brevity.

Prefer:

```
existing_user
encoded_data
storage_path
```

over single-letter names.

---

# 40. Treat file-backed repositories as transitional

They are suitable for:

* prototypes
* CLIs
* local tools
* small services

They are not ideal for:

* high write volume
* multi-process access
* large datasets
* transactional systems

Design code so replacing the storage layer is straightforward.

---
