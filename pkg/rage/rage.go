// Package rage provides a public API for embedding the RAGE Python runtime in Go applications.
//
// Basic usage:
//
//	// Run Python code and get a result
//	result, err := rage.Run(`x = 1 + 2`)
//
//	// Create a state for more control
//	state := rage.NewState()
//	defer state.Close()
//	state.SetGlobal("name", rage.String("World"))
//	result, err := state.Run(`greeting = "Hello, " + name`)
//	greeting := state.GetGlobal("greeting")
//
// The API is inspired by gopher-lua for familiarity.
package rage

import (
	"context"
	"fmt"
	"time"

	"github.com/ATSOTECK/RAGE/internal/compiler"
	"github.com/ATSOTECK/RAGE/internal/runtime"
	"github.com/ATSOTECK/RAGE/internal/stdlib"
)

// Module represents a standard library module that can be enabled.
type Module int

const (
	ModuleMath Module = iota
	ModuleRandom
	ModuleString
	ModuleSys
	ModuleTime
	ModuleRe
	ModuleCollections
)

// AllModules is a convenience slice containing all available modules.
var AllModules = []Module{
	ModuleMath,
	ModuleRandom,
	ModuleString,
	ModuleSys,
	ModuleTime,
	ModuleRe,
	ModuleCollections,
}

// StateOption is a functional option for configuring State creation.
type StateOption func(*stateConfig)

type stateConfig struct {
	modules map[Module]bool
}

// WithModule enables a specific stdlib module.
func WithModule(m Module) StateOption {
	return func(c *stateConfig) {
		c.modules[m] = true
	}
}

// WithModules enables multiple stdlib modules.
func WithModules(modules ...Module) StateOption {
	return func(c *stateConfig) {
		for _, m := range modules {
			c.modules[m] = true
		}
	}
}

// WithAllModules enables all stdlib modules.
func WithAllModules() StateOption {
	return func(c *stateConfig) {
		for _, m := range AllModules {
			c.modules[m] = true
		}
	}
}

// State represents a Python execution state.
// It wraps the VM and provides a clean API for running Python code.
type State struct {
	vm             *runtime.VM
	compiled       map[string]*runtime.CodeObject
	enabledModules map[Module]bool
}

// NewState creates a new Python execution state with all stdlib modules enabled.
// This is a convenience function equivalent to NewStateWithModules(WithAllModules()).
func NewState() *State {
	return NewStateWithModules(WithAllModules())
}

// NewBareState creates a new Python execution state with no stdlib modules enabled.
// Use EnableModule or EnableAllModules to enable modules after creation.
func NewBareState() *State {
	return NewStateWithModules()
}

// NewStateWithModules creates a new Python execution state with the specified options.
//
// Example:
//
//	// Create state with only math and string modules
//	state := rage.NewStateWithModules(
//	    rage.WithModule(rage.ModuleMath),
//	    rage.WithModule(rage.ModuleString),
//	)
//
//	// Or use WithModules for multiple at once
//	state := rage.NewStateWithModules(
//	    rage.WithModules(rage.ModuleMath, rage.ModuleString, rage.ModuleTime),
//	)
//
//	// Enable all modules
//	state := rage.NewStateWithModules(rage.WithAllModules())
func NewStateWithModules(opts ...StateOption) *State {
	cfg := &stateConfig{
		modules: make(map[Module]bool),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	runtime.ResetModules()

	// Initialize only the requested modules
	for m := range cfg.modules {
		initModule(m)
	}

	return &State{
		vm:             runtime.NewVM(),
		compiled:       make(map[string]*runtime.CodeObject),
		enabledModules: cfg.modules,
	}
}

// initModule initializes a single stdlib module.
func initModule(m Module) {
	switch m {
	case ModuleMath:
		stdlib.InitMathModule()
	case ModuleRandom:
		stdlib.InitRandomModule()
	case ModuleString:
		stdlib.InitStringModule()
	case ModuleSys:
		stdlib.InitSysModule()
	case ModuleTime:
		stdlib.InitTimeModule()
	case ModuleRe:
		stdlib.InitReModule()
	case ModuleCollections:
		stdlib.InitCollectionsModule()
	}
}

// EnableModule enables a specific stdlib module.
// This can be called after state creation to add modules.
func (s *State) EnableModule(m Module) {
	if !s.enabledModules[m] {
		initModule(m)
		s.enabledModules[m] = true
	}
}

// EnableModules enables multiple stdlib modules.
func (s *State) EnableModules(modules ...Module) {
	for _, m := range modules {
		s.EnableModule(m)
	}
}

// EnableAllModules enables all available stdlib modules.
func (s *State) EnableAllModules() {
	for _, m := range AllModules {
		s.EnableModule(m)
	}
}

// IsModuleEnabled returns true if the specified module is enabled.
func (s *State) IsModuleEnabled(m Module) bool {
	return s.enabledModules[m]
}

// EnabledModules returns a slice of all enabled modules.
func (s *State) EnabledModules() []Module {
	var result []Module
	for m := range s.enabledModules {
		result = append(result, m)
	}
	return result
}

// Close releases resources associated with the state.
// Always call this when done with the state.
func (s *State) Close() {
	s.vm = nil
	s.compiled = nil
}

// Run compiles and executes Python source code.
// Returns the result of the last expression or nil.
func (s *State) Run(source string) (Value, error) {
	return s.RunWithFilename(source, "<string>")
}

// RunWithFilename compiles and executes Python source code with a filename for error messages.
func (s *State) RunWithFilename(source, filename string) (Value, error) {
	code, errs := compiler.CompileSource(source, filename)
	if len(errs) > 0 {
		return nil, &CompileErrors{Errors: errs}
	}

	result, err := s.vm.Execute(code)
	if err != nil {
		return nil, err
	}
	return fromRuntime(result), nil
}

// RunWithTimeout executes Python code with a timeout.
func (s *State) RunWithTimeout(source string, timeout time.Duration) (Value, error) {
	code, errs := compiler.CompileSource(source, "<string>")
	if len(errs) > 0 {
		return nil, &CompileErrors{Errors: errs}
	}

	result, err := s.vm.ExecuteWithTimeout(timeout, code)
	if err != nil {
		return nil, err
	}
	return fromRuntime(result), nil
}

// RunWithContext executes Python code with a context for cancellation.
func (s *State) RunWithContext(ctx context.Context, source string) (Value, error) {
	code, errs := compiler.CompileSource(source, "<string>")
	if len(errs) > 0 {
		return nil, &CompileErrors{Errors: errs}
	}

	result, err := s.vm.ExecuteWithContext(ctx, code)
	if err != nil {
		return nil, err
	}
	return fromRuntime(result), nil
}

// Compile compiles Python source code without executing it.
// The compiled code can be executed later with Execute.
func (s *State) Compile(source, filename string) (*Code, error) {
	code, errs := compiler.CompileSource(source, filename)
	if len(errs) > 0 {
		return nil, &CompileErrors{Errors: errs}
	}
	return &Code{code: code}, nil
}

// Execute runs previously compiled code.
func (s *State) Execute(code *Code) (Value, error) {
	result, err := s.vm.Execute(code.code)
	if err != nil {
		return nil, err
	}
	return fromRuntime(result), nil
}

// ExecuteWithTimeout runs previously compiled code with a timeout.
func (s *State) ExecuteWithTimeout(code *Code, timeout time.Duration) (Value, error) {
	result, err := s.vm.ExecuteWithTimeout(timeout, code.code)
	if err != nil {
		return nil, err
	}
	return fromRuntime(result), nil
}

// SetGlobal sets a global variable accessible from Python code.
func (s *State) SetGlobal(name string, value Value) {
	s.vm.SetGlobal(name, toRuntime(value))
}

// GetGlobal retrieves a global variable set by Python code.
func (s *State) GetGlobal(name string) Value {
	return fromRuntime(s.vm.GetGlobal(name))
}

// GetGlobals returns all global variables as a map.
func (s *State) GetGlobals() map[string]Value {
	result := make(map[string]Value)
	for k, v := range s.vm.Globals {
		result[k] = fromRuntime(v)
	}
	return result
}

// Register registers a Go function that can be called from Python.
//
// Example:
//
//	state.Register("greet", func(s *rage.State, args ...rage.Value) rage.Value {
//	    name := args[0].String()
//	    return rage.String("Hello, " + name + "!")
//	})
//
// Then in Python: greet("World")
func (s *State) Register(name string, fn GoFunc) {
	wrapper := func(vm *runtime.VM) int {
		// Collect arguments from stack
		nargs := vm.GetTop()
		args := make([]Value, nargs)
		for i := 1; i <= nargs; i++ {
			args[i-1] = fromRuntime(vm.Get(i))
		}

		// Call the Go function
		result := fn(s, args...)

		// Push result if not nil
		if result != nil {
			vm.Push(toRuntime(result))
			return 1
		}
		return 0
	}
	s.vm.Register(name, wrapper)
}

// RegisterBuiltin registers a Go function as a builtin.
func (s *State) RegisterBuiltin(name string, fn GoFunc) {
	wrapper := func(vm *runtime.VM) int {
		nargs := vm.GetTop()
		args := make([]Value, nargs)
		for i := 1; i <= nargs; i++ {
			args[i-1] = fromRuntime(vm.Get(i))
		}
		result := fn(s, args...)
		if result != nil {
			vm.Push(toRuntime(result))
			return 1
		}
		return 0
	}
	s.vm.RegisterBuiltin(name, wrapper)
}

// Code represents compiled Python bytecode.
type Code struct {
	code *runtime.CodeObject
}

// Name returns the name of the compiled code (module/function name).
func (c *Code) Name() string {
	return c.code.Name
}

// GoFunc is the signature for Go functions callable from Python.
type GoFunc func(s *State, args ...Value) Value

// =====================================
// Package-level convenience functions
// =====================================

// Run is a convenience function that creates a temporary state,
// runs the code, and returns the result.
func Run(source string) (Value, error) {
	state := NewState()
	defer state.Close()
	return state.Run(source)
}

// RunWithTimeout runs Python code with a timeout.
func RunWithTimeout(source string, timeout time.Duration) (Value, error) {
	state := NewState()
	defer state.Close()
	return state.RunWithTimeout(source, timeout)
}

// Eval evaluates a Python expression and returns the result.
// Unlike Run, this expects an expression, not statements.
func Eval(expr string) (Value, error) {
	// Wrap expression to capture result
	source := fmt.Sprintf("__result__ = (%s)", expr)
	state := NewState()
	defer state.Close()
	_, err := state.Run(source)
	if err != nil {
		return nil, err
	}
	return state.GetGlobal("__result__"), nil
}

// CompileErrors wraps multiple compilation errors.
type CompileErrors struct {
	Errors []error
}

func (e *CompileErrors) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%d compilation errors (first: %s)", len(e.Errors), e.Errors[0].Error())
}

// Unwrap returns the first error for errors.Is/As compatibility.
func (e *CompileErrors) Unwrap() error {
	if len(e.Errors) > 0 {
		return e.Errors[0]
	}
	return nil
}
