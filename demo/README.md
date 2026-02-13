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

A Go "game server" loads its configuration from three Python scripts:

### `config/settings.py` — Environment-Aware Settings
- **Conditionals**: Different hosts, ports, pool sizes per environment
- **Computed values**: `max_connections = db_pool_size * 10`
- **String interpolation**: `api_url = f"https://{host}:{port}/api/v1"`
- **Validation**: `assert 1 <= log_level <= 5` catches bad config at load time
- **Feature flags**: Built with expressions like `"ssl": port == 8443`

### `config/items.py` — Templated Item Definitions
- **Factory functions**: `make_weapon("Sword", 15, 3.0, "iron", tier=3)` generates a complete item with computed damage, weight, value, and rarity
- **Comprehensions**: 15 weapons generated from `[make_weapon(...) for mat in materials for tier in tiers]`
- **Probability tables**: Loot drop rates computed with exponential decay
- **Aggregation**: Set bonus stats computed from the items in the set

### `config/levels.py` — Generated Progression Tables
- **XP curve**: `[int(100 * (1.15 ** level)) for level in range(50)]` — one line instead of 50 hand-written entries
- **Stat formulas**: HP, attack, defense, speed all computed from level
- **Milestones**: Declarative rewards at specific levels
- **Boss scaling**: Stats derived from the same formulas as player stats

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
