package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// PyModule represents a Python module
type PyModule struct {
	Name    string
	Dict    map[string]Value // Module namespace (__dict__)
	Doc     string           // __doc__
	Package string           // __package__
	Loader  Value            // __loader__
	Spec    Value            // __spec__
}

func (m *PyModule) Type() string   { return "module" }
func (m *PyModule) String() string { return fmt.Sprintf("<module '%s'>", m.Name) }

// Get retrieves a value from the module by name
func (m *PyModule) Get(name string) (Value, bool) {
	v, ok := m.Dict[name]
	return v, ok
}

// Set sets a value in the module namespace
func (m *PyModule) Set(name string, value Value) {
	m.Dict[name] = value
}

// ModuleLoader is a function that creates/loads a module
type ModuleLoader func(vm *VM) *PyModule

// moduleRegistry holds registered modules
var moduleRegistry = make(map[string]ModuleLoader)

// moduleMu protects moduleRegistry, loadedModules, and moduleLoading from concurrent access
var moduleMu sync.RWMutex

// moduleLoadState tracks an in-progress filesystem module load
type moduleLoadState struct {
	vm   *VM          // VM performing the load
	done chan struct{} // closed when loading completes
	err  error        // non-nil if loading failed
}

// moduleLoading tracks modules currently being loaded from the filesystem
var moduleLoading = make(map[string]*moduleLoadState)

// RegisterModule registers a module loader with the given name
// The loader will be called when the module is first imported
func RegisterModule(name string, loader ModuleLoader) {
	moduleMu.Lock()
	moduleRegistry[name] = loader
	moduleMu.Unlock()
}

// NewModule creates a new module with the given name
func NewModule(name string) *PyModule {
	return &PyModule{
		Name: name,
		Dict: map[string]Value{
			"__name__": NewString(name),
			"__doc__":  None,
		},
	}
}

// ModuleBuilder provides a fluent API for building modules
type ModuleBuilder struct {
	module *PyModule
}

// NewModuleBuilder creates a new module builder
func NewModuleBuilder(name string) *ModuleBuilder {
	return &ModuleBuilder{
		module: NewModule(name),
	}
}

// Doc sets the module's docstring
func (b *ModuleBuilder) Doc(doc string) *ModuleBuilder {
	b.module.Doc = doc
	b.module.Dict["__doc__"] = NewString(doc)
	return b
}

// Const adds a constant value to the module
func (b *ModuleBuilder) Const(name string, value Value) *ModuleBuilder {
	b.module.Dict[name] = value
	return b
}

// Func adds a Go function to the module
func (b *ModuleBuilder) Func(name string, fn GoFunction) *ModuleBuilder {
	b.module.Dict[name] = NewGoFunction(name, fn)
	return b
}

// Method is an alias for Func
func (b *ModuleBuilder) Method(name string, fn GoFunction) *ModuleBuilder {
	return b.Func(name, fn)
}

// Type adds a type/class to the module with methods
func (b *ModuleBuilder) Type(typeName string, constructor GoFunction, methods map[string]GoFunction) *ModuleBuilder {
	// Register the type metatable
	mt := &TypeMetatable{
		Name:    typeName,
		Methods: methods,
	}
	typeMetaMu.Lock()
	typeMetatables[typeName] = mt
	typeMetaMu.Unlock()

	// Add constructor to module
	if constructor != nil {
		b.module.Dict[typeName] = NewGoFunction(typeName, constructor)
	}
	return b
}

// SubModule adds a submodule to the module
func (b *ModuleBuilder) SubModule(name string, submodule *PyModule) *ModuleBuilder {
	b.module.Dict[name] = submodule
	return b
}

// Build returns the constructed module
func (b *ModuleBuilder) Build() *PyModule {
	return b.module
}

// Register builds and registers the module with the global registry
func (b *ModuleBuilder) Register() *PyModule {
	module := b.Build()
	RegisterModule(module.Name, func(vm *VM) *PyModule {
		return module
	})
	return module
}

// =====================================
// VM Methods for Module Management
// =====================================

// modules holds loaded modules (cached)
var loadedModules = make(map[string]*PyModule)

// ImportModule imports a module by name
func (vm *VM) ImportModule(name string) (*PyModule, error) {
	moduleMu.Lock()
	defer moduleMu.Unlock()

	// Check if already loaded
	if mod, ok := loadedModules[name]; ok {
		// If another VM is still loading this module, wait for it to finish
		// so we don't return a half-initialized module. Same-VM loads are
		// circular imports and should return the partial module (like CPython).
		if ls, loading := moduleLoading[name]; loading && ls.vm != vm {
			moduleMu.Unlock()
			<-ls.done
			moduleMu.Lock()
			if ls.err != nil {
				return nil, fmt.Errorf("error executing '%s': %v", name, ls.err)
			}
			// Re-check cache after waiting (may have been deleted on failure)
			if mod, ok := loadedModules[name]; ok {
				return mod, nil
			}
			return nil, fmt.Errorf("ModuleNotFoundError: No module named '%s'", name)
		}
		return mod, nil
	}

	// If another VM is loading this module but hasn't cached it yet,
	// wait for it (handles the window between moduleLoading entry and cache)
	if ls, loading := moduleLoading[name]; loading && ls.vm != vm {
		moduleMu.Unlock()
		<-ls.done
		moduleMu.Lock()
		if ls.err != nil {
			return nil, fmt.Errorf("error executing '%s': %v", name, ls.err)
		}
		if mod, ok := loadedModules[name]; ok {
			return mod, nil
		}
		return nil, fmt.Errorf("ModuleNotFoundError: No module named '%s'", name)
	}

	// Check if there's a registered loader
	if loader, ok := moduleRegistry[name]; ok {
		mod := loader(vm)
		loadedModules[name] = mod
		return mod, nil
	}

	// Filesystem fallback: search SearchPaths for <name>.py
	if vm.FileImporter != nil {
		for _, dir := range vm.SearchPaths {
			pyFile := filepath.Join(dir, name+".py")
			if _, err := os.Stat(pyFile); err == nil {
				code, err := vm.FileImporter(pyFile)
				if err != nil {
					return nil, fmt.Errorf("error importing '%s': %v", name, err)
				}

				mod := NewModule(name)
				mod.Package = name
				mod.Dict["__package__"] = NewString(name)
				mod.Dict["__file__"] = NewString(pyFile)

				// Track loading state so concurrent importers can wait
				ls := &moduleLoadState{vm: vm, done: make(chan struct{})}
				moduleLoading[name] = ls

				// Cache before executing to handle circular imports
				loadedModules[name] = mod

				// Unlock while executing module code (may re-enter ImportModule)
				moduleMu.Unlock()
				err = vm.ExecuteInModule(code, mod)
				moduleMu.Lock()

				if err != nil {
					ls.err = err
					delete(loadedModules, name)
				}
				close(ls.done)
				delete(moduleLoading, name)

				if err != nil {
					return nil, fmt.Errorf("error executing '%s': %v", name, err)
				}

				return mod, nil
			}
		}
	}

	return nil, fmt.Errorf("ModuleNotFoundError: No module named '%s'", name)
}

// GetModule retrieves a loaded module by name
func (vm *VM) GetModule(name string) (*PyModule, bool) {
	moduleMu.RLock()
	mod, ok := loadedModules[name]
	moduleMu.RUnlock()
	return mod, ok
}

// ResolveRelativeImport resolves a relative import to an absolute module name.
// Parameters:
//   - name: the module name from the import statement (may be empty for "from . import x")
//   - level: number of dots (1 for ".", 2 for "..", etc.)
//   - packageName: the current package (__package__ or derived from __name__)
//
// Returns the resolved absolute module name or an error.
func ResolveRelativeImport(name string, level int, packageName string) (string, error) {
	if level == 0 {
		// Not a relative import
		return name, nil
	}

	if packageName == "" {
		return "", fmt.Errorf("ImportError: attempted relative import with no known parent package")
	}

	// Split the package into parts
	parts := splitModuleName(packageName)

	// For level 1, we stay in the current package
	// For level 2, we go up one level, etc.
	// The number of parts we need to keep is len(parts) - (level - 1)
	keep := len(parts) - (level - 1)

	if keep < 0 {
		return "", fmt.Errorf("ImportError: attempted relative import beyond top-level package")
	}

	// Build the base path
	var base string
	if keep > 0 {
		base = joinModuleName(parts[:keep])
	}

	// Append the module name if provided
	if name == "" {
		if base == "" {
			return "", fmt.Errorf("ImportError: attempted relative import with no known parent package")
		}
		return base, nil
	}

	if base == "" {
		return name, nil
	}
	return base + "." + name, nil
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

// RegisterModule registers a module loader on the VM
func (vm *VM) RegisterModule(name string, loader ModuleLoader) {
	moduleMu.Lock()
	moduleRegistry[name] = loader
	moduleMu.Unlock()
}

// RegisterModuleInstance directly registers a pre-built module
func (vm *VM) RegisterModuleInstance(name string, module *PyModule) {
	moduleMu.Lock()
	loadedModules[name] = module
	moduleRegistry[name] = func(vm *VM) *PyModule {
		return module
	}
	moduleMu.Unlock()
}

// ResetModules clears the loaded modules cache and related registries.
// This should be called before initializing a new State to ensure
// a clean slate (useful for testing and creating isolated states).
func ResetModules() {
	moduleMu.Lock()
	loadedModules = make(map[string]*PyModule)
	moduleLoading = make(map[string]*moduleLoadState)
	moduleMu.Unlock()
	ResetPendingBuiltins()
	ResetTypeMetatables()
}

// =====================================
// Built-in module: builtins
// =====================================

// InitBuiltinsModule creates and registers the builtins module
func (vm *VM) initBuiltinsModule() {
	builtins := NewModule("builtins")
	builtins.Doc = "Built-in functions, exceptions, and other objects."

	// Copy all builtins to the module
	for name, value := range vm.builtins {
		builtins.Dict[name] = value
	}

	moduleMu.Lock()
	loadedModules["builtins"] = builtins
	moduleMu.Unlock()
}
