# RAGE

**R**eally **A**dequate **G**o-python **E**ngine

RAGE is an embeddable Python 3.14 runtime written in Go. It allows you to run Python code directly from your Go
applications without any external dependencies or CGO.

## Features

- Pure Go implementation - no CGO, no external Python installation required
- Embeddable - designed to be used as a library in Go applications
- Timeout support - prevent infinite loops with execution timeouts
- Context cancellation - integrate with Go's context for graceful shutdown
- Standard library modules - math, random, string, sys, time, re, collections
- Go interoperability - call Go functions from Python and vice versa

## Installation

```bash
go get github.com/ATSOTECK/RAGE
```

## Quick Start

### Run Python Code

```go
package main

import (
    "fmt"
    "log"

    "github.com/ATSOTECK/RAGE/pkg/rage"
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

    "github.com/ATSOTECK/RAGE/pkg/rage"
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

    "github.com/ATSOTECK/RAGE/pkg/rage"
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

    "github.com/ATSOTECK/RAGE/pkg/rage"
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
val := rage.FromGo(map[string]interface{}{
    "numbers": []interface{}{1, 2, 3},
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
- Basic data types: None, bool, int, float, str, list, tuple, dict, set, range
- Operators: arithmetic, comparison, logical, bitwise
- Control flow: if/elif/else, for, while, break, continue
- Functions: def, lambda, recursion, closures, *args, **kwargs
- Classes: class definitions, `__init__`, instance attributes, methods, single inheritance
- Comprehensions: list `[x for x in items]`, dict `{k: v for k, v in items}`, set `{x for x in items}`
- Imports: import, from...import (for stdlib modules)
- Built-in functions: print, len, range, str, int, float, bool, list, dict, tuple, set, type, isinstance, abs, min, max, sum, enumerate, zip, map, filter, any, all, reversed, repr

### Not Yet Implemented
- Exception handling (try/except/finally)
- Generator expressions and yield
- Decorators
- Multiple inheritance (MRO is simplified)
- Async/await
- File I/O
- Most of the standard library

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
