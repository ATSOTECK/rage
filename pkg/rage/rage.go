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
	"os"
	"path/filepath"
	"time"

	"github.com/ATSOTECK/rage/internal/compiler"
	"github.com/ATSOTECK/rage/internal/runtime"
	"github.com/ATSOTECK/rage/internal/stdlib"
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
	ModuleAsyncio
	ModuleIO // File I/O
	ModuleJSON
	ModuleOS
	ModuleDatetime
	ModuleTyping
	ModuleCSV
	ModuleItertools
	ModuleFunctools
	ModuleBase64
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
	ModuleAsyncio,
	ModuleIO,
	ModuleJSON,
	ModuleOS,
	ModuleDatetime,
	ModuleTyping,
	ModuleCSV,
	ModuleItertools,
	ModuleFunctools,
	ModuleBase64,
}

// Builtin represents an opt-in builtin function that can be enabled.
// These builtins provide reflection and code execution capabilities
// that are not enabled by default for security reasons.
type Builtin int

const (
	BuiltinRepr Builtin = iota
	BuiltinDir
	BuiltinGlobals
	BuiltinLocals
	BuiltinVars
	BuiltinCompile
	BuiltinExec
	BuiltinEval
)

// ReflectionBuiltins contains all reflection-related builtins (repr, dir, globals, locals, vars).
// These are relatively safe introspection functions.
var ReflectionBuiltins = []Builtin{
	BuiltinRepr,
	BuiltinDir,
	BuiltinGlobals,
	BuiltinLocals,
	BuiltinVars,
}

// ExecutionBuiltins contains all code execution builtins (compile, exec, eval).
// These allow arbitrary code execution and should be enabled with caution.
var ExecutionBuiltins = []Builtin{
	BuiltinCompile,
	BuiltinExec,
	BuiltinEval,
}

// AllBuiltins contains all opt-in builtins.
var AllBuiltins = []Builtin{
	BuiltinRepr,
	BuiltinDir,
	BuiltinGlobals,
	BuiltinLocals,
	BuiltinVars,
	BuiltinCompile,
	BuiltinExec,
	BuiltinEval,
}

// StateOption is a functional option for configuring State creation.
type StateOption func(*stateConfig)

type stateConfig struct {
	modules  map[Module]bool
	builtins map[Builtin]bool
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

// WithBuiltin enables a specific opt-in builtin function.
func WithBuiltin(b Builtin) StateOption {
	return func(c *stateConfig) {
		c.builtins[b] = true
	}
}

// WithBuiltins enables multiple opt-in builtin functions.
func WithBuiltins(builtins ...Builtin) StateOption {
	return func(c *stateConfig) {
		for _, b := range builtins {
			c.builtins[b] = true
		}
	}
}

// WithReflectionBuiltins enables all reflection builtins (repr, dir, globals, locals, vars).
func WithReflectionBuiltins() StateOption {
	return func(c *stateConfig) {
		for _, b := range ReflectionBuiltins {
			c.builtins[b] = true
		}
	}
}

// WithExecutionBuiltins enables all execution builtins (compile, exec, eval).
func WithExecutionBuiltins() StateOption {
	return func(c *stateConfig) {
		for _, b := range ExecutionBuiltins {
			c.builtins[b] = true
		}
	}
}

// WithAllBuiltins enables all opt-in builtins.
func WithAllBuiltins() StateOption {
	return func(c *stateConfig) {
		for _, b := range AllBuiltins {
			c.builtins[b] = true
		}
	}
}

// State represents a Python execution state.
// It wraps the VM and provides a clean API for running Python code.
type State struct {
	vm              *runtime.VM
	compiled        map[string]*runtime.CodeObject
	enabledModules  map[Module]bool
	enabledBuiltins map[Builtin]bool
	closed          bool
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
		modules:  make(map[Module]bool),
		builtins: make(map[Builtin]bool),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	runtime.ResetModules()

	// Set up the compile function bridge for exec/eval/compile builtins
	runtime.CompileFunc = compileForBuiltin

	// Initialize only the requested modules
	for m := range cfg.modules {
		initModule(m)
	}

	vm := runtime.NewVM()

	// Set up filesystem imports
	vm.FileImporter = func(path string) (*runtime.CodeObject, error) {
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		code, errs := compiler.CompileSource(string(src), path)
		if len(errs) > 0 {
			return nil, errs[0]
		}
		return code, nil
	}

	// Initialize opt-in builtins
	for b := range cfg.builtins {
		initBuiltin(vm, b)
	}

	return &State{
		vm:              vm,
		compiled:        make(map[string]*runtime.CodeObject),
		enabledModules:  cfg.modules,
		enabledBuiltins: cfg.builtins,
	}
}

// compileForBuiltin wraps compiler.CompileSource for use by exec/eval/compile builtins
func compileForBuiltin(source, filename, mode string) (*runtime.CodeObject, error) {
	// For "eval" mode, wrap the expression to capture its result
	// For "exec" mode, compile as normal statements
	// For "single" mode, compile as a single interactive statement
	if mode == "eval" {
		// Wrap expression to capture result: __eval_result__ = (expression)
		source = "__eval_result__ = (" + source + ")"
	}
	code, errs := compiler.CompileSource(source, filename)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	return code, nil
}

// initBuiltin initializes a single opt-in builtin
func initBuiltin(vm *runtime.VM, b Builtin) {
	switch b {
	case BuiltinRepr:
		vm.RegisterBuiltin("repr", runtime.BuiltinRepr)
	case BuiltinDir:
		vm.RegisterBuiltin("dir", runtime.BuiltinDir)
	case BuiltinGlobals:
		vm.RegisterBuiltin("globals", runtime.BuiltinGlobals)
	case BuiltinLocals:
		vm.RegisterBuiltin("locals", runtime.BuiltinLocals)
	case BuiltinVars:
		vm.RegisterBuiltin("vars", runtime.BuiltinVars)
	case BuiltinCompile:
		vm.RegisterBuiltin("compile", runtime.BuiltinCompile)
	case BuiltinExec:
		vm.RegisterBuiltin("exec", runtime.BuiltinExec)
	case BuiltinEval:
		vm.RegisterBuiltin("eval", runtime.BuiltinEval)
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
	case ModuleAsyncio:
		stdlib.InitAsyncioModule()
	case ModuleIO:
		stdlib.InitIOModule()
	case ModuleJSON:
		stdlib.InitJSONModule()
	case ModuleOS:
		stdlib.InitOSModule()
	case ModuleDatetime:
		stdlib.InitDatetimeModule()
	case ModuleTyping:
		stdlib.InitTypingModule()
	case ModuleCSV:
		stdlib.InitCSVModule()
	case ModuleItertools:
		stdlib.InitItertoolsModule()
	case ModuleFunctools:
		stdlib.InitFunctoolsModule()
	case ModuleBase64:
		stdlib.InitBase64Module()
	}
}

// EnableModule enables a specific stdlib module.
// This can be called after state creation to add modules.
func (s *State) EnableModule(m Module) {
	if !s.enabledModules[m] {
		initModule(m)
		s.enabledModules[m] = true
		// Apply any pending builtins that were registered by initModule
		s.vm.ApplyPendingBuiltins()
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

// EnableBuiltin enables a specific opt-in builtin function.
// This can be called after state creation to add builtins.
func (s *State) EnableBuiltin(b Builtin) {
	if s.enabledBuiltins == nil {
		s.enabledBuiltins = make(map[Builtin]bool)
	}
	if !s.enabledBuiltins[b] {
		initBuiltin(s.vm, b)
		s.enabledBuiltins[b] = true
	}
}

// EnableBuiltins enables multiple opt-in builtin functions.
func (s *State) EnableBuiltins(builtins ...Builtin) {
	for _, b := range builtins {
		s.EnableBuiltin(b)
	}
}

// EnableReflectionBuiltins enables all reflection builtins (repr, dir, globals, locals, vars).
func (s *State) EnableReflectionBuiltins() {
	for _, b := range ReflectionBuiltins {
		s.EnableBuiltin(b)
	}
}

// EnableExecutionBuiltins enables all execution builtins (compile, exec, eval).
func (s *State) EnableExecutionBuiltins() {
	for _, b := range ExecutionBuiltins {
		s.EnableBuiltin(b)
	}
}

// EnableAllBuiltins enables all opt-in builtins.
func (s *State) EnableAllBuiltins() {
	for _, b := range AllBuiltins {
		s.EnableBuiltin(b)
	}
}

// IsBuiltinEnabled returns true if the specified builtin is enabled.
func (s *State) IsBuiltinEnabled(b Builtin) bool {
	if s.enabledBuiltins == nil {
		return false
	}
	return s.enabledBuiltins[b]
}

// EnabledBuiltins returns a slice of all enabled builtins.
func (s *State) EnabledBuiltins() []Builtin {
	var result []Builtin
	for b := range s.enabledBuiltins {
		result = append(result, b)
	}
	return result
}

// Close releases resources associated with the state.
// Always call this when done with the state.
// After Close is called, the state should not be used.
func (s *State) Close() {
	if s.closed {
		return // Already closed, idempotent
	}
	s.closed = true
	// Clear references to allow GC to reclaim memory
	s.vm = nil
	s.compiled = nil
	s.enabledModules = nil
	s.enabledBuiltins = nil
}

// checkClosed returns an error if the state has been closed.
func (s *State) checkClosed() error {
	if s.closed {
		return fmt.Errorf("operation on closed State")
	}
	return nil
}

// Run compiles and executes Python source code.
// Returns the result of the last expression or nil.
func (s *State) Run(source string) (Value, error) {
	if err := s.checkClosed(); err != nil {
		return nil, err
	}
	return s.RunWithFilename(source, "<string>")
}

// RunWithFilename compiles and executes Python source code with a filename for error messages.
func (s *State) RunWithFilename(source, filename string) (Value, error) {
	if err := s.checkClosed(); err != nil {
		return nil, err
	}

	// Set search paths from the script's directory so local imports work
	if filename != "<string>" {
		absPath, err := filepath.Abs(filename)
		if err == nil {
			dir := filepath.Dir(absPath)
			// Add the script's directory if not already present
			found := false
			for _, p := range s.vm.SearchPaths {
				if p == dir {
					found = true
					break
				}
			}
			if !found {
				s.vm.SearchPaths = append(s.vm.SearchPaths, dir)
			}
		}
	}

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
	if err := s.checkClosed(); err != nil {
		return nil, err
	}
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
	if err := s.checkClosed(); err != nil {
		return nil, err
	}
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
	// Note: Compile doesn't need checkClosed as it doesn't use the VM
	code, errs := compiler.CompileSource(source, filename)
	if len(errs) > 0 {
		return nil, &CompileErrors{Errors: errs}
	}
	return &Code{code: code}, nil
}

// Execute runs previously compiled code.
func (s *State) Execute(code *Code) (Value, error) {
	if err := s.checkClosed(); err != nil {
		return nil, err
	}
	result, err := s.vm.Execute(code.code)
	if err != nil {
		return nil, err
	}
	return fromRuntime(result), nil
}

// ExecuteWithTimeout runs previously compiled code with a timeout.
func (s *State) ExecuteWithTimeout(code *Code, timeout time.Duration) (Value, error) {
	if err := s.checkClosed(); err != nil {
		return nil, err
	}
	result, err := s.vm.ExecuteWithTimeout(timeout, code.code)
	if err != nil {
		return nil, err
	}
	return fromRuntime(result), nil
}

// SetGlobal sets a global variable accessible from Python code.
func (s *State) SetGlobal(name string, value Value) {
	if s.closed {
		return // Silently ignore on closed state
	}
	s.vm.SetGlobal(name, toRuntime(value))
}

// GetGlobal retrieves a global variable set by Python code.
func (s *State) GetGlobal(name string) Value {
	if s.closed {
		return nil
	}
	return fromRuntime(s.vm.GetGlobal(name))
}

// GetGlobals returns all global variables as a map.
func (s *State) GetGlobals() map[string]Value {
	if s.closed {
		return nil
	}
	result := make(map[string]Value)
	for k, v := range s.vm.Globals {
		result[k] = fromRuntime(v)
	}
	return result
}

// GetModuleAttr retrieves an attribute from an imported module.
// Returns nil if the module doesn't exist or the attribute isn't found.
func (s *State) GetModuleAttr(moduleName, attrName string) Value {
	if s.closed {
		return nil
	}
	mod, ok := s.vm.GetModule(moduleName)
	if !ok {
		return nil
	}
	val, ok := mod.Get(attrName)
	if !ok {
		return nil
	}
	return fromRuntime(val)
}

// RegisterPythonModule compiles and registers Python source code as an importable module.
// The module can then be imported using "import moduleName" or "from moduleName import ...".
func (s *State) RegisterPythonModule(moduleName, source string) error {
	if err := s.checkClosed(); err != nil {
		return err
	}

	// Ensure parent packages exist before registering the module
	s.ensureParentPackages(moduleName)

	// Compile the source
	code, errs := compiler.CompileSource(source, moduleName+".py")
	if len(errs) > 0 {
		return &CompileErrors{Errors: errs}
	}

	// Create a new module
	mod := runtime.NewModule(moduleName)

	// Set __package__ for relative imports
	// For a module like "pkg.sub.module", the package is "pkg.sub" (parent)
	// For a package like "pkg.sub" (typically __init__.py), the package is "pkg.sub" itself
	packageName := moduleName
	if lastDot := lastIndexByte(moduleName, '.'); lastDot >= 0 {
		// This is a submodule, package is the parent
		packageName = moduleName[:lastDot]
	}
	mod.Package = packageName
	mod.Dict["__package__"] = runtime.NewString(packageName)

	// Execute the code to populate the module's namespace
	err := s.vm.ExecuteInModule(code, mod)
	if err != nil {
		return err
	}

	// Register the module so it can be imported
	s.vm.RegisterModuleInstance(moduleName, mod)

	// Link this module as an attribute of its parent package
	if lastDot := lastIndexByte(moduleName, '.'); lastDot >= 0 {
		parentName := moduleName[:lastDot]
		childName := moduleName[lastDot+1:]
		if parentMod, ok := s.vm.GetModule(parentName); ok {
			parentMod.Dict[childName] = mod
		}
	}

	return nil
}

// ensureParentPackages creates empty parent packages for a dotted module name.
// For "a.b.c", it ensures "a" and "a.b" exist as packages.
func (s *State) ensureParentPackages(moduleName string) {
	parts := splitModuleName(moduleName)
	if len(parts) <= 1 {
		return // No parent packages needed
	}

	// Create each parent package if it doesn't exist
	for i := 1; i < len(parts); i++ {
		parentName := joinModuleName(parts[:i])
		if _, ok := s.vm.GetModule(parentName); !ok {
			// Create an empty package
			pkg := runtime.NewModule(parentName)
			pkg.Package = parentName
			pkg.Dict["__package__"] = runtime.NewString(parentName)
			pkg.Dict["__path__"] = runtime.NewList([]runtime.Value{}) // Mark as package
			s.vm.RegisterModuleInstance(parentName, pkg)
		}
	}

	// Link parent packages together
	for i := 1; i < len(parts)-1; i++ {
		parentName := joinModuleName(parts[:i])
		childName := joinModuleName(parts[:i+1])
		if parentMod, ok := s.vm.GetModule(parentName); ok {
			if childMod, ok := s.vm.GetModule(childName); ok {
				parentMod.Dict[parts[i]] = childMod
			}
		}
	}
}

// splitModuleName splits a module name by dots
func splitModuleName(name string) []string {
	if name == "" {
		return nil
	}
	var parts []string
	start := 0
	for i := 0; i < len(name); i++ {
		if name[i] == '.' {
			parts = append(parts, name[start:i])
			start = i + 1
		}
	}
	parts = append(parts, name[start:])
	return parts
}

// joinModuleName joins module name parts with dots
func joinModuleName(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += "." + parts[i]
	}
	return result
}

// lastIndexByte returns the index of the last occurrence of c in s, or -1 if not present.
func lastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
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
	if s.closed {
		return // Silently ignore on closed state
	}
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
	if s.closed {
		return // Silently ignore on closed state
	}
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
