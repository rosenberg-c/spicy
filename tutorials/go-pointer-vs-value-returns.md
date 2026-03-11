# Go: Why Return Pointers vs Values from Constructors

When writing constructor functions in Go, you need to decide whether to return a pointer (`*Type`) or a value (`Type`). This guide explains the reasoning with practical examples.

## The Question

```go
type Agent struct {
    verbose bool
    logger  *slog.Logger
}

// Why this?
func New(verbose bool) *Agent {
    return &Agent{
        verbose: verbose,
        logger:  slog.New(handler),
    }
}

// Instead of this?
func New(verbose bool) Agent {
    return Agent{
        verbose: verbose,
        logger:  slog.New(handler),
    }
}
```

---

## Why Return Pointers

### 1. Identity Semantics (Most Important)

An `Agent` represents a **service/object** with identity, not a **value** like an integer or coordinate.

```go
// With pointer - identity semantics
agent1 := agent.New(true)
agent2 := agent.New(true)
// agent1 != agent2 (different instances, different memory addresses)

// With value - value semantics
agent1 := agent.New(true)
agent2 := agent1
// agent2 is a complete copy
// Changes to one don't affect the other (if there were any mutating methods)
```

**Key principle:** Services, controllers, and managers should have identity. You want "this specific agent" not "an agent with these values."

### 2. Method Receivers Match

The `Run` method uses a pointer receiver:

```go
func (a *Agent) Run(ctx context.Context, model, prompt string) (string, error) {
    // ...
}
```

**Why this matters:**
- Returning `*Agent` from constructor matches the pointer receiver
- Clear and explicit: "this type is meant to be used by pointer"
- Avoids implicit address-taking operations

If you returned `Agent` by value:
```go
agent := agent.New(true)  // returns Agent (value)
agent.Run(...)            // Go auto-converts to &agent
```

Go handles this automatically, but returning `*Agent` makes the intent clearer.

### 3. Go Standard Library Convention

Constructor functions for "object-like" types conventionally return pointers:

```go
// Standard library examples
http.NewServeMux() *ServeMux
sql.Open() *DB
slog.New() *Logger
bufio.NewReader() *Reader
```

Types representing **pure values** return by value:
```go
time.Now() Time
errors.New() error
context.Background() Context
```

### 4. Consistency with Internal Fields

The `Agent` struct contains `*slog.Logger` (already a pointer). The `Logger` type is meant to be used by pointer, so it's natural for `Agent` to follow the same pattern.

### 5. Prevents Accidental Copies

Returning a pointer signals: "don't copy this."

```go
// With pointer - clear that you're passing the same instance
func processWithAgent(a *Agent) {
    a.Run(...)
}

// With value - might accidentally copy
func processWithAgent(a Agent) {
    // 'a' is a copy - any state changes won't persist
}
```

---

## When to Return by Value

Not everything should return pointers! Return **values** when:

### 1. Pure Data Structures
```go
type Point struct {
    X, Y int
}

// Value return - it's just coordinates
func NewPoint(x, y int) Point {
    return Point{X: x, Y: y}
}
```

### 2. Small, Immutable Configs
```go
type Color struct {
    R, G, B uint8
}

func RGB(r, g, b uint8) Color {
    return Color{R: r, G: g, B: b}
}
```

### 3. Types Meant to be Compared
```go
type Status string

const (
    StatusPending  Status = "pending"
    StatusComplete Status = "complete"
)

// Values can be compared with ==
if status == StatusComplete { ... }
```

### 4. No Mutating Methods
```go
type Dimensions struct {
    Width, Height int
}

// All methods are read-only, value receivers
func (d Dimensions) Area() int {
    return d.Width * d.Height
}
```

---

## Decision Matrix

| Return Pointer When... | Return Value When... |
|------------------------|----------------------|
| Type represents a service/object | Type represents data |
| Has pointer receivers | Has only value receivers |
| Contains mutex/channels | Small and immutable |
| Shouldn't be copied | Safe to copy |
| Has identity semantics | Has value semantics |
| Follows stdlib pattern | Pure data structure |

---

## Common Patterns

### Services/Controllers → Pointer
```go
type Server struct { ... }
func NewServer() *Server { ... }

type Database struct { ... }
func NewDatabase() *Database { ... }

type Cache struct { ... }
func NewCache() *Cache { ... }
```

### Data/Values → Value
```go
type Config struct { ... }
func LoadConfig() Config { ... }

type Result struct { ... }
func Compute() Result { ... }

type Point struct { ... }
func NewPoint(x, y int) Point { ... }
```

### When in Doubt

Ask: "Is this a **thing** (identity) or a **value** (data)?"

- Thing → pointer (`*Agent`, `*Server`, `*Logger`)
- Value → value (`Config`, `Point`, `Color`)

---

## Performance Considerations

### Small Structs
For small structs (< 64 bytes), the performance difference is negligible:
- Passing by value: copies the struct
- Passing by pointer: copies an 8-byte pointer (on 64-bit systems)

**Don't optimize prematurely** - choose based on semantics, not performance.

### Large Structs
For large structs (> 64 bytes), pointers avoid copying overhead:
```go
type LargeConfig struct {
    // Many fields...
}

// Pointer avoids copying
func NewLargeConfig() *LargeConfig { ... }
```

---

## Example: Agent Type

```go
type Agent struct {
    verbose bool
    logger  *slog.Logger
}
```

**Why return `*Agent`:**
1. ✅ Represents a service (executes commands, wraps logger)
2. ✅ Has pointer receiver methods
3. ✅ Follows stdlib convention (like `*slog.Logger`, `*http.Client`)
4. ✅ Identity semantics - each agent is a unique instance
5. ✅ Contains a pointer field (`*slog.Logger`)

**Would `Agent` value return work?**
Yes, technically. But it would be unconventional and might confuse readers who expect service types to be pointers.

---

## Quick Reference

```go
// Service/object pattern - return pointer
type Service struct { ... }
func NewService() *Service {
    return &Service{ ... }
}
func (s *Service) DoWork() { ... }

// Data/value pattern - return value
type Config struct { ... }
func NewConfig() Config {
    return Config{ ... }
}
func (c Config) Validate() bool { ... }
```

---

## Key Takeaway

The choice isn't about performance or memory - it's about **intent and semantics**:

- **Pointers** = "This is a unique thing with behavior and identity"
- **Values** = "This is data that can be copied and compared"

For types like `Agent`, `Server`, `Database`, `Logger` → use pointers.
For types like `Point`, `Color`, `Config`, `Time` → use values.
