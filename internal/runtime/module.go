package runtime

import (
	"fmt"
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

// RegisterModule registers a module loader with the given name
// The loader will be called when the module is first imported
func RegisterModule(name string, loader ModuleLoader) {
	moduleRegistry[name] = loader
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
	typeMetatables[typeName] = mt

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
	// Check if already loaded
	if mod, ok := loadedModules[name]; ok {
		return mod, nil
	}

	// Check if there's a registered loader
	if loader, ok := moduleRegistry[name]; ok {
		mod := loader(vm)
		loadedModules[name] = mod
		return mod, nil
	}

	return nil, fmt.Errorf("ModuleNotFoundError: No module named '%s'", name)
}

// GetModule retrieves a loaded module by name
func (vm *VM) GetModule(name string) (*PyModule, bool) {
	mod, ok := loadedModules[name]
	return mod, ok
}

// RegisterModule registers a module loader on the VM
func (vm *VM) RegisterModule(name string, loader ModuleLoader) {
	moduleRegistry[name] = loader
}

// RegisterModuleInstance directly registers a pre-built module
func (vm *VM) RegisterModuleInstance(name string, module *PyModule) {
	loadedModules[name] = module
	moduleRegistry[name] = func(vm *VM) *PyModule {
		return module
	}
}

// ResetModules clears the loaded modules cache and related registries.
// This should be called before initializing a new State to ensure
// a clean slate (useful for testing and creating isolated states).
func ResetModules() {
	loadedModules = make(map[string]*PyModule)
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

	loadedModules["builtins"] = builtins
}
