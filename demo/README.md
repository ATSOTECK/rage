# Python-as-Configuration Demo

This demo shows RAGE's two main value propositions:

1. **Python as an embedded configuration language** — static config formats (TOML, JSON, YAML) can't express computed values, conditionals, or validation logic. Python can.
2. **Go-defined classes via the ClassBuilder API** — expose rich, Pythonic types to scripts without writing any Python class definitions.

## Running

From the repo root:

```bash
go run demo/main.go                       # development (default)
RAGE_ENV=production go run demo/main.go   # production settings
RAGE_ENV=staging go run demo/main.go      # staging settings
```

## What the Demo Shows

A Go "game server" loads its configuration from Python scripts. Shared balance constants live in `common.py` and are imported by the other scripts — just like a real project.

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

This config script uses **four classes defined entirely in Go** using the ClassBuilder API. No Python class definitions needed — the classes are registered from Go and used naturally in Python.

#### `Vec2(x, y)` — 2D position/vector

Operators and builtins: `+`, `-`, `*` (scalar), unary `-`, `abs()`, `round()`, `==`, `hash()`

| Kind | Name | Description |
|---|---|---|
| Property | `length` | Euclidean length (`sqrt(x*x + y*y)`) |
| Method | `distance_to(other)` | Distance between two points |
| Method | `normalized()` | Unit vector (length 1, same direction) |
| Method | `dot(other)` | Dot product |
| Method | `lerp(other, t)` | Linear interpolation toward `other` by factor `t` |
| StaticMethod | `zero()` | Returns `Vec2(0, 0)` |
| ClassMethod | `from_angle(deg, length=1)` | Construct from angle in degrees |
| Attr | `ORIGIN` | Class-level constant `Vec2(0, 0)` |

```python
center = Vec2(400, 300)
tower_positions = [center + offset for offset in offsets]
direction = Vec2(3, 4).normalized()          # Vec2(0.6, 0.8)
midpoint = spawn_points[0].lerp(spawn_points[1], 0.5)
angled = Vec2.from_angle(45, 100)            # Vec2(70.71, 70.71)
```

#### `Color(r, g, b)` — RGB color

Operators and builtins: `+` (additive blend), `*` (brightness), `==`, `format(c, "hex")`, `int()`, `float()`

| Kind | Name | Description |
|---|---|---|
| Method | `inverted()` | Complementary color (`255 - r`, etc.) |
| Method | `grayscale()` | Luminance-weighted grayscale |
| MethodKw | `mix(other, t=0.5)` | Blend two colors by factor `t` |
| ClassMethod | `from_hex("#rrggbb")` | Construct from hex string |
| ClassGetItem | `Color[int]` | Generic subscript syntax |

```python
sunset = Color(255, 100, 50).mix(Color(100, 0, 150), t=0.4)
gold = Color.from_hex("#ffd700")
hex_str = format(Color(255, 0, 0), "hex")   # "#ff0000"
packed = int(Color(255, 128, 0))             # 0xff8000
```

#### `Inventory(capacity)` — game container

Operators and builtins: `[]`, `del`, `in`, `len()`, `bool()`, `+=`, `reversed()`, iteration

| Kind | Name | Description |
|---|---|---|
| Method | `keys()` | Item names in insertion order |
| Method | `total_weight()` | Sum of `"weight"` fields across all items |
| Method | `drop(name)` | Remove and return an item (or `None`) |

```python
starter = Inventory(10)
starter["sword"] = {"damage": 5, "weight": 3}
carry_weight = starter.total_weight()        # 3
dropped = starter.drop("sword")              # returns the item dict
"sword" in starter                           # False after drop
```

#### `GameSession(name)` — context manager + event logger

Protocols: `with` statement (context manager), callable instances with kwargs, dynamic attribute access

| Kind | Name | Description |
|---|---|---|
| Method | `filter(event_name)` | Events matching a given name |
| Method | `event_count()` | Total number of logged events |
| GetAttr | `.name`, `.events` | Dynamic attribute access |

```python
with GameSession("config_load") as session:
    session(event="spawn_loaded", count=4)
    session(event="biomes_loaded", count=3)

spawn_events = session.filter("spawn_loaded")  # 1 matching event
total = session.event_count()                   # 2
```

#### ClassBuilder API coverage

| Category | API | Class |
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
| Class methods | `ClassMethod` | Vec2, Color |
| Instance methods | `Method` | Vec2, Color, Inventory, GameSession |
| Keyword methods | `MethodKw` | Color |
| Class subscript | `ClassGetItem` | Color |
| Rounding | `Round` | Vec2 |

## How It Works (Go Side)

### Loading Python config

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

### Defining classes with ClassBuilder

```go
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
    // Instance method — regular named method callable from Python
    Method("distance_to", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
        o, _ := rage.AsObject(args[0])
        x1, _ := rage.AsFloat(self.Get("x"))
        y1, _ := rage.AsFloat(self.Get("y"))
        x2, _ := rage.AsFloat(o.Get("x"))
        y2, _ := rage.AsFloat(o.Get("y"))
        dx, dy := x2-x1, y2-y1
        return rage.Float(math.Sqrt(dx*dx + dy*dy)), nil
    }).
    // ClassMethod — receives the class, can create instances
    ClassMethod("from_angle", func(s *rage.State, cls rage.ClassValue, args ...rage.Value) (rage.Value, error) {
        deg, _ := rage.AsFloat(args[0])
        rad := deg * math.Pi / 180
        result := cls.NewInstance()
        result.Set("x", rage.Float(math.Cos(rad)))
        result.Set("y", rage.Float(math.Sin(rad)))
        return result, nil
    }).
    Str(func(s *rage.State, self rage.Object) (string, error) {
        return fmt.Sprintf("Vec2(%g, %g)",
            self.Get("x"), self.Get("y")), nil
    }).
    Build(state)

state.SetGlobal("Vec2", vec2)
```

Then in Python:

```python
a = Vec2(3, 4)
b = Vec2(6, 8)
print(a + b)                # Vec2(9, 12)
print(a.distance_to(b))     # 5.0
v = Vec2.from_angle(90)     # Vec2(0, 1)
```
