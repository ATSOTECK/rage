# Python-as-Configuration Demo

This demo shows RAGE's main value proposition: **using Python as an embedded configuration language** in a Go application. Static config formats (TOML, JSON, YAML) can't express computed values, conditionals, or validation logic. Python can.

## Running

From the repo root:

```bash
go run demo/main.go                       # development (default)
RAGE_ENV=production go run demo/main.go   # production settings
RAGE_ENV=staging go run demo/main.go      # staging settings
```

## What the Demo Shows

A Go "game server" loads its configuration from Python scripts. Shared balance constants live in `common.py` and are imported by the other scripts — just like a real project:

### `config/common.py` — Shared Game Balance Data
- **Single source of truth**: Materials, rarities, and zones defined once
- **Imported by siblings**: `from common import materials, rarities` in items.py, `from common import zones` in levels.py
- **Change once, update everywhere**: Add a material here and every config file picks it up

### `config/settings.py` — Environment-Aware Settings
- **Conditionals**: Different hosts, ports, pool sizes per environment
- **Computed values**: `max_connections = db_pool_size * 10`
- **String interpolation**: `api_url = f"https://{host}:{port}/api/v1"`
- **Validation**: `assert 1 <= log_level <= 5` catches bad config at load time
- **Feature flags**: Built with expressions like `"ssl": port == 8443`

### `config/items.py` — Templated Item Definitions
- **Cross-script imports**: `from common import materials, rarities`
- **Factory functions**: `make_weapon("Sword", 15, 3.0, "iron", tier=3)` generates a complete item with computed damage, weight, value, and rarity
- **Comprehensions**: 15 weapons generated from `[make_weapon(...) for mat in materials for tier in tiers]`
- **Probability tables**: Loot drop rates computed with exponential decay
- **Aggregation**: Set bonus stats computed from the items in the set

### `config/levels.py` — Generated Progression Tables
- **Stdlib + local imports**: `import math` and `from common import zones`
- **XP curve**: `[math.floor(100 * math.pow(1.15, level)) for level in range(50)]` — one line instead of 50 hand-written entries
- **Stat formulas**: HP, attack, defense, speed all computed from level
- **Milestones**: Declarative rewards at specific levels
- **Boss scaling**: Stats derived from the same formulas as player stats

### `config/entities.py` — Go-Defined Game Classes

This config script uses **four classes defined entirely in Go** using the ClassBuilder API. No Python class definitions — the classes are registered from Go and used naturally in Python:

- **`Vec2(x, y)`** — 2D vector with arithmetic (`+`, `-`, `*`), `abs()`, `round()`, a `.length` property, and `Vec2.zero()`/`Vec2.from_angle()` factories. Methods: `distance_to()`, `normalized()`, `dot()`, `lerp()`.
- **`Color(r, g, b)`** — RGB color with additive blending (`+`), brightness scaling (`*`), hex formatting (`format(c, "hex")` → `#rrggbb`), and numeric conversions (`int()` → packed RGB, `float()` → luminance). Methods: `inverted()`, `grayscale()`, `mix(other, t=0.5)`. Factory: `Color.from_hex("#rrggbb")`.
- **`Inventory(capacity)`** — Ordered container with `[]` access, `in` membership, `len()`, `del`, iteration, and `reversed()`. Methods: `keys()`, `total_weight()`, `drop(name)`.
- **`GameSession(name)`** — Context manager (`with` statement) that logs config events via `session(event=..., count=...)`. Dynamic attribute access exposes `.name` and `.events`. Methods: `filter(event_name)`, `event_count()`.

**ClassBuilder features demonstrated:**

| Category | Methods | Class |
|---|---|---|
| Binary operators | `Add`, `Sub`, `Mul` | Vec2, Color |
| Unary operators | `Neg`, `Abs` | Vec2 |
| Comparison + hashing | `Eq`, `Hash` | Vec2, Color |
| String representations | `Str`, `Repr` | Vec2, Color, Inventory |
| Numeric conversions | `IntConv`, `FloatConv`, `Format` | Color |
| Container protocol | `Len`, `GetItem`, `SetItem`, `DelItem`, `Contains`, `Bool` | Inventory |
| Iteration | `Iter`, `Next`, `Reversed` | Inventory |
| In-place operators | `IAdd` | Inventory |
| Context manager | `Enter`, `Exit` | GameSession |
| Callable instances | `CallKw` | GameSession |
| Dynamic attributes | `GetAttr` | GameSession |
| Properties | `Property` | Vec2 |
| Static methods | `StaticMethod` | Vec2 |
| Class subscript | `ClassGetItem` | Color |
| Rounding | `Round` | Vec2 |
| Instance methods | `Method` | Vec2, Color, Inventory, GameSession |
| Keyword methods | `MethodKw` | Color (`mix`) |
| Class methods | `ClassMethod` | Vec2 (`from_angle`), Color (`from_hex`) |

## How It Works (Go Side)

```go
// 1. Create a state with only the modules config scripts need
state := rage.NewStateWithModules(
    rage.WithModules(rage.ModuleMath, rage.ModuleDataclasses, rage.ModuleCollections),
)

// 2. Register Go functions callable from Python
state.Register("env", func(_ *rage.State, args ...rage.Value) rage.Value {
    // ... reads os.Getenv()
})

// 3. Inject Go values into Python's global scope
state.SetGlobal("cpu_count", rage.Int(int64(runtime.NumCPU())))

// 4. Load and execute Python config
src, _ := os.ReadFile("demo/config/settings.py")
state.RunWithFilename(string(src), "demo/config/settings.py")

// 5. Extract results as Go types
settings, _ := rage.AsDict(state.GetGlobal("settings"))
host, _ := rage.AsString(settings["host"])
port, _ := rage.AsInt(settings["port"])
```

The Python scripts are **pure Python** — no RAGE-specific imports needed. The Go side provides context (`env()`, `cpu_count`) and reads results (`GetGlobal`).

## Defining Classes in Go (ClassBuilder)

```go
// Build a Vec2 class with operators, properties, and static methods
vec2 := rage.NewClass("Vec2").
    Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
        self.Set("x", args[0])
        self.Set("y", args[1])
        return nil
    }).
    Add(func(s *rage.State, self rage.Object, other rage.Value) (rage.Value, error) {
        o, _ := rage.AsObject(other)
        x1, _ := rage.AsFloat(self.Get("x"))
        y1, _ := rage.AsFloat(self.Get("y"))
        x2, _ := rage.AsFloat(o.Get("x"))
        y2, _ := rage.AsFloat(o.Get("y"))
        result := self.Class().NewInstance()
        result.Set("x", rage.Float(x1+x2))
        result.Set("y", rage.Float(y1+y2))
        return result, nil
    }).
    Str(func(s *rage.State, self rage.Object) (string, error) {
        x, _ := rage.AsFloat(self.Get("x"))
        y, _ := rage.AsFloat(self.Get("y"))
        return fmt.Sprintf("Vec2(%g, %g)", x, y), nil
    }).
    Property("length", func(s *rage.State, self rage.Object) (rage.Value, error) {
        x, _ := rage.AsFloat(self.Get("x"))
        y, _ := rage.AsFloat(self.Get("y"))
        return rage.Float(math.Sqrt(x*x + y*y)), nil
    }).
    Build(state)

state.SetGlobal("Vec2", vec2)
```

Then in Python:

```python
center = Vec2(400, 300)
offset = Vec2(-50, -50)
pos = center + offset           # Vec2(350, 250)
print(pos.length)               # 430.116...
```
