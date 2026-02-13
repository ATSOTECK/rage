# RAGE

**R**eally **A**dequate **G**o-python **E**ngine

RAGE is an embeddable Python 3.14 runtime written in Go. It allows you to run Python code directly from your Go
applications without any external dependencies or CGO.

RAGE comes with optional batteries included. If you wan't some of the python standard library there are many modules you can use.
If you don't want to use the python standard library at all you don't have to. If you only want to use a few modules from the
standard library you can pick and choose which modules are made available to scripts.

## Features

- Pure Go implementation - no CGO, no external Python installation required
- Embeddable - designed to be used as a library in Go applications
- Timeout support - prevent infinite loops with execution timeouts
- Context cancellation - integrate with Go's context for graceful shutdown
- Standard library modules - math, random, string, sys, time, re, collections, json, os, datetime, typing, asyncio, csv, itertools, functools, base64, abc, dataclasses
- Go interoperability - call Go functions from Python and vice versa

## Installation

```bash
go get github.com/ATSOTECK/rage
```

## Tests
- **Unit tests**: `go test ./...`
- **Integration tests**: 93 scripts with 1934 tests covering data types, operators, control flow, functions, classes, exceptions, generators, comprehensions, closures, decorators, imports, context managers, metaclasses, descriptors, string formatting, dataclasses, and more
  - Run with `go run test/integration/integration_test_runner.go`

## Quick Start

### Run Python Code

```go
package main

import (
    "fmt"
    "log"

    "github.com/ATSOTECK/rage/pkg/rage"
)

func main() {
    // Simple one-liner
    result, err := rage.Run(`print("Hello from Python!")`)
    if err != nil {
        log.Fatal(err)
    }

    // Evaluate an expression
    result, err = rage.Eval(`2 ** 10`)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result) // 1024
}
```

### Using State for More Control

```go
package main

import (
    "fmt"
    "log"

    "github.com/ATSOTECK/rage/pkg/rage"
)

func main() {
    // Create a new execution state
    state := rage.NewState()
    defer state.Close()

    // Set variables accessible from Python
    state.SetGlobal("name", rage.String("World"))
    state.SetGlobal("count", rage.Int(42))

    // Run Python code
    _, err := state.Run(`
greeting = "Hello, " + name + "!"
doubled = count * 2
numbers = [1, 2, 3, 4, 5]
total = sum(numbers)
    `)
    if err != nil {
        log.Fatal(err)
    }

    // Get variables set by Python
    fmt.Println(state.GetGlobal("greeting")) // Hello, World!
    fmt.Println(state.GetGlobal("doubled"))  // 84
    fmt.Println(state.GetGlobal("total"))    // 15
}
```

### Registering Go Functions

```go
package main

import (
    "fmt"
    "strings"

    "github.com/ATSOTECK/rage/pkg/rage"
)

func main() {
    state := rage.NewState()
    defer state.Close()

    // Register a Go function callable from Python
    state.Register("shout", func(s *rage.State, args ...rage.Value) rage.Value {
        if len(args) == 0 {
            return rage.None
        }
        text, _ := rage.AsString(args[0])
        return rage.String(strings.ToUpper(text) + "!")
    })

    // Call it from Python
    state.Run(`message = shout("hello world")`)
    fmt.Println(state.GetGlobal("message")) // HELLO WORLD!
}
```

### Timeouts and Context

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/ATSOTECK/rage/pkg/rage"
)

func main() {
    // Using timeout
    _, err := rage.RunWithTimeout(`
while True:
    pass  # infinite loop
    `, 2*time.Second)
    fmt.Println(err) // execution timeout

    // Using context
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    state := rage.NewState()
    defer state.Close()

    _, err = state.RunWithContext(ctx, `
x = 0
while True:
    x += 1
    `)
    fmt.Println(err) // context deadline exceeded
}
```

## Demo: Python as Configuration

The `demo/` directory contains a complete example of RAGE's core use case: **replacing static config files (TOML, JSON, YAML) with Python scripts**. A Go "game server" loads its configuration from Python, showcasing things static formats can't do:

```bash
go run demo/main.go                       # development settings
RAGE_ENV=production go run demo/main.go   # production settings
```

The Go side registers helpers and injects values, then runs pure Python config files:

```go
state := rage.NewStateWithModules(
    rage.WithModules(rage.ModuleMath, rage.ModuleDataclasses, rage.ModuleCollections),
)
defer state.Close()

// Make Go functions callable from Python
state.Register("env", func(_ *rage.State, args ...rage.Value) rage.Value {
    name, _ := rage.AsString(args[0])
    if v := os.Getenv(name); v != "" {
        return rage.String(v)
    }
    return rage.String("")
})
state.SetGlobal("cpu_count", rage.Int(int64(runtime.NumCPU())))

// Load Python config and extract results as Go types
src, _ := os.ReadFile("config/settings.py")
state.RunWithFilename(string(src), "config/settings.py")
settings, _ := rage.AsDict(state.GetGlobal("settings"))
```

The Python config scripts use conditionals, computed values, validation, and comprehensions:

```python
# config/settings.py — environment-aware, self-validating config
environment = env("RAGE_ENV", "development")

if environment == "production":
    db_pool_size = cpu_count * 4    # computed from Go-injected value
else:
    db_pool_size = 2

assert db_pool_size > 0, "db_pool_size must be positive"  # validates at load time
api_url = f"https://{host}:{port}/api/v1"                  # derived string
features = {"ssl": port == 8443, "metrics": environment != "development"}
```

```python
# config/items.py — one template function generates 15 weapons
def make_weapon(name, base_damage, material, tier=1):
    mat = materials[material]
    return {"name": f"{material.title()} {name}", "damage": int(base_damage * mat["damage"] * (1 + tier * 0.25)), ...}

swords = [make_weapon("Sword", 15, mat, tier) for mat in materials for tier in [1, 3, 5]]
```

```python
# config/levels.py — 50 levels of XP/stats from a few formulas
import math
xp_for_level = [math.floor(100 * math.pow(1.15, level)) for level in range(50)]
```

See [`demo/README.md`](demo/README.md) for the full walkthrough.

## Controlling Standard Library Modules

By default, `NewState()` enables all available stdlib modules. For more control:

```go
// Create state with only specific modules
state := rage.NewStateWithModules(
    rage.WithModule(rage.ModuleMath),
    rage.WithModule(rage.ModuleString),
)
defer state.Close()

// Or enable multiple at once
state := rage.NewStateWithModules(
    rage.WithModules(rage.ModuleMath, rage.ModuleString, rage.ModuleTime),
)

// Create a bare state with no modules
state := rage.NewBareState()
state.EnableModule(rage.ModuleMath)    // Enable one
state.EnableAllModules()                // Or enable all later

// Check what's enabled
if state.IsModuleEnabled(rage.ModuleMath) {
    fmt.Println("Math module is available")
}
```

### Available Modules

| Module | Constant | Description |
|--------|----------|-------------|
| math | `rage.ModuleMath` | Mathematical functions (sin, cos, sqrt, etc.) |
| random | `rage.ModuleRandom` | Random number generation |
| string | `rage.ModuleString` | String constants (ascii_letters, digits, etc.) |
| sys | `rage.ModuleSys` | System information (version, platform) |
| time | `rage.ModuleTime` | Time functions (time, sleep) |
| re | `rage.ModuleRe` | Regular expressions |
| collections | `rage.ModuleCollections` | Container datatypes (Counter, defaultdict) |
| json | `rage.ModuleJSON` | JSON encoding and decoding |
| os | `rage.ModuleOS` | OS interface (environ, path manipulation) |
| datetime | `rage.ModuleDatetime` | Date and time types |
| typing | `rage.ModuleTyping` | Type hint support |
| asyncio | `rage.ModuleAsyncio` | Basic async/await support |
| csv | `rage.ModuleCSV` | CSV file reading and writing |
| itertools | `rage.ModuleItertools` | Iterator building blocks (chain, combinations, permutations, etc.) |
| functools | `rage.ModuleFunctools` | Higher-order functions (partial, reduce, lru_cache, wraps) |
| io | `rage.ModuleIO` | File I/O operations |
| base64 | `rage.ModuleBase64` | Base16, Base32, Base64 data encodings |

## Working with Values

RAGE uses the `rage.Value` interface to represent Python values.

### Creating Values

```go
// Primitives
none := rage.None
b := rage.Bool(true)
i := rage.Int(42)
f := rage.Float(3.14)
s := rage.String("hello")

// Collections
list := rage.List(rage.Int(1), rage.Int(2), rage.Int(3))
tuple := rage.Tuple(rage.String("a"), rage.String("b"))
dict := rage.Dict("name", rage.String("Alice"), "age", rage.Int(30))

// From Go values (automatic conversion)
val := rage.FromGo(map[string]any{
    "numbers": []any{1, 2, 3},
    "active":  true,
})
```

### Type Checking

```go
if rage.IsInt(val) {
    n, _ := rage.AsInt(val)
    fmt.Println("Integer:", n)
}

if rage.IsList(val) {
    items, _ := rage.AsList(val)
    for _, item := range items {
        fmt.Println(item)
    }
}

if rage.IsDict(val) {
    dict, _ := rage.AsDict(val)
    for k, v := range dict {
        fmt.Printf("%s: %v\n", k, v)
    }
}
```

### Converting to Go Values

```go
val := state.GetGlobal("result")

// Get the underlying Go value
goVal := val.GoValue()

// Or use type-specific helpers
if n, ok := rage.AsInt(val); ok {
    fmt.Println("Got integer:", n)
}
```

## Compile Once, Run Many

For repeated execution, compile once and run multiple times:

```go
state := rage.NewState()
defer state.Close()

// Compile once
code, err := state.Compile(`result = x * 2`, "multiply.py")
if err != nil {
    log.Fatal(err)
}

// Execute multiple times with different inputs
for i := 0; i < 5; i++ {
    state.SetGlobal("x", rage.Int(int64(i)))
    state.Execute(code)
    fmt.Println(state.GetGlobal("result"))
}
// Output: 0, 2, 4, 6, 8
```

## Error Handling

```go
// Compilation errors
_, err := rage.Run(`def broken syntax`)
if compErr, ok := err.(*rage.CompileErrors); ok {
    for _, e := range compErr.Errors {
        fmt.Println("Compile error:", e)
    }
}

// Runtime errors
_, err = rage.Run(`x = 1 / 0`)
if err != nil {
    fmt.Println("Runtime error:", err)
}
```

## Current Status

RAGE is under active development. Currently supported:

### Implemented
- Data types: None, bool, int, float, str, bytes, bytearray, list, tuple, dict, set, frozenset, range, slice
- Operators: arithmetic, comparison, logical, bitwise, matrix multiplication (`@`), in-place operations
- Control flow: if/elif/else, for, while, break, continue, pass, match/case
- Functions: def, lambda, recursion, closures, *args, **kwargs, default arguments, nonlocal
- Classes: class definitions, `__init__`, `__new__`, instance attributes, methods, single and multiple inheritance (C3 linearization), properties, classmethods, staticmethods, metaclasses (`class Foo(metaclass=Meta)`), `__slots__`
- Exception handling: try/except/else/finally, raise, raise from, custom exception types, exception attributes (`.args`, `.__cause__`, `.__context__`)
- Generators: yield, yield from, generator expressions
- Decorators: function and class decorators
- Comprehensions: list `[x for x in items]`, dict `{k: v for k, v in items}`, set `{x for x in items}`
- Imports: import, from...import, relative imports
- Context managers: with statement support
- String formatting: f-strings, `%` printf-style (`%s`, `%d`, `%f`, `%e`, `%g`, `%x`, `%o`, `%c`, `%(key)s` dict formatting, flags, `*` width/precision)
- Walrus operator: assignment expressions (`:=`)
- Extended unpacking: `a, *rest, b = [1, 2, 3, 4]`
- Descriptor protocol: `__get__`, `__set__`, `__delete__` (data descriptors, non-data descriptors, class-level access)
- Dunder methods for custom classes: `__new__`, `__init__`, `__str__`, `__repr__`, `__call__`, `__hash__`, `__len__`, `__iter__`, `__next__`, `__contains__`, `__getattr__`, `__setattr__`, `__delattr__`, `__getitem__`, `__setitem__`, `__delitem__`, `__enter__`, `__exit__`, `__bool__`, `__int__`, `__index__`, `__abs__`, `__neg__`, `__pos__`, `__invert__`, operator overloading (`__add__`, `__sub__`, `__mul__`, `__matmul__`, `__eq__`, `__lt__`, etc. including reflected variants)
- Built-in functions: print (with `sep`, `end`, `flush`), len, range, str, int, float, bool, list, dict, tuple, set, bytes, bytearray, type, isinstance, issubclass, abs, min, max, sum, enumerate, zip, map, filter, any, all, reversed, sorted, repr, input, ord, chr, hasattr, getattr, setattr, delattr, dir, vars, id, pow, divmod, hex, oct, bin, round, callable, property, classmethod, staticmethod, super, iter, next

### Not Yet Implemented
- Full async/await - async generators, async context managers (basic support via asyncio module)

### Security Notes
Reflection builtins (`globals`, `locals`, `compile`, `exec`, `eval`) are opt-in and disabled by default. Enable them explicitly if needed:

```go
state := rage.NewStateWithModules(
    rage.WithAllModules(),
    rage.WithBuiltin(rage.BuiltinGlobals),
    rage.WithBuiltin(rage.BuiltinExec),
)
```

## Thread Safety

Each `State` is NOT safe for concurrent use. Create separate States for concurrent execution, or use appropriate synchronization.

```go
// Safe: separate states per goroutine
for i := 0; i < 10; i++ {
    go func(n int) {
        state := rage.NewState()
        defer state.Close()
        state.SetGlobal("n", rage.Int(int64(n)))
        state.Run(`result = n * n`)
    }(i)
}
```

## License

MIT

## Contributing

pls
