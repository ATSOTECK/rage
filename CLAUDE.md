# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

RAGE (Really Adequate Go-python Engine) is a pure Go Python 3 interpreter. No CGO or external Python runtime required.

## Build Commands

The project uses `tasks.grav` task runner:

```bash
go run cmd/rage/main.go <script.py>  # Run a Python script
go build -o bin/rage cmd/rage/main.go  # Build binary
go test ./...  # Run all tests
go run test/integration/integration_test_runner.go  # Run integration tests
```

## Architecture

**Compilation Pipeline:**
```
Python Source → Lexer → Parser → AST → Compiler → Bytecode → VM
```

**Core Packages (in `internal/`):**

- **compiler/** - Source to bytecode conversion
  - `lexer.go` - Tokenization with Python-specific indentation handling, f-string support
  - `parser.go` - AST generation using Pratt parsing for expressions
  - `compiler.go` - AST to bytecode compilation
  - `optimizer.go` - Bytecode optimization passes

- **runtime/** - Bytecode execution
  - `vm.go` - Stack-based virtual machine
  - `opcode.go` - Bytecode instruction definitions
  - `module.go` - Module system implementation

- **model/** - Data structures
  - `ast.go` - AST node definitions for Python constructs
  - `token.go` - Token type definitions

- **stdlib/** - Standard library modules (math, random, re, string, sys, time, collections)

**Public API (`pkg/rage/rage.go`):**
- `NewState()` - Create execution state with all modules
- `NewBareState()` - Minimal state with no modules
- `state.Run(source)` - Execute Python code
- `state.SetGlobal()/GetGlobal()` - Python/Go value exchange
- `state.Register(name, fn)` - Register Go functions callable from Python
- `state.Compile()/Execute()` - Compile once, run many times

## Testing

Tests use Go's `testing` package with testify assertions:

```bash
go test ./...  # All tests
go test ./test -run TestName  # Single test
```

Integration tests in `test/integration/scripts/` are Python files with expected outputs in `test/integration/expected/`.

## Current Limitations

Not yet implemented:
- File I/O
