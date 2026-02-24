package runtime

import (
	"testing"
)

// =====================================
// NewModule
// =====================================

func TestNewModule(t *testing.T) {
	mod := NewModule("mymodule")
	if mod == nil {
		t.Fatal("NewModule returned nil")
	}
	if mod.Name != "mymodule" {
		t.Errorf("Name = %q, want %q", mod.Name, "mymodule")
	}
	if mod.Dict == nil {
		t.Fatal("Dict should be initialized")
	}

	// Check __name__ is set
	nameVal, ok := mod.Dict["__name__"]
	if !ok {
		t.Error("__name__ not set in Dict")
	} else if s, ok := nameVal.(*PyString); !ok || s.Value != "mymodule" {
		t.Errorf("__name__ = %v, want 'mymodule'", nameVal)
	}

	// Check __doc__ is None by default
	docVal, ok := mod.Dict["__doc__"]
	if !ok {
		t.Error("__doc__ not set in Dict")
	} else if _, ok := docVal.(*PyNone); !ok {
		t.Errorf("__doc__ = %v, want None", docVal)
	}
}

func TestNewModuleTypeAndString(t *testing.T) {
	mod := NewModule("test_mod")
	if mod.Type() != "module" {
		t.Errorf("Type() = %q, want %q", mod.Type(), "module")
	}
	if mod.String() != "<module 'test_mod'>" {
		t.Errorf("String() = %q, want %q", mod.String(), "<module 'test_mod'>")
	}
}

// =====================================
// Module Get / Set
// =====================================

func TestModuleGetSet(t *testing.T) {
	mod := NewModule("testmod")

	mod.Set("x", MakeInt(42))
	val, ok := mod.Get("x")
	if !ok {
		t.Error("expected to find 'x'")
	}
	if v, ok := val.(*PyInt); !ok || v.Value != 42 {
		t.Errorf("Get('x') = %v, want 42", val)
	}

	// Get missing key
	_, ok = mod.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for missing key")
	}
}

func TestModuleSetOverwrite(t *testing.T) {
	mod := NewModule("testmod")

	mod.Set("x", MakeInt(1))
	mod.Set("x", MakeInt(2))
	val, ok := mod.Get("x")
	if !ok {
		t.Fatal("expected to find 'x'")
	}
	if v := val.(*PyInt).Value; v != 2 {
		t.Errorf("Get('x') = %d, want 2 after overwrite", v)
	}
}

// =====================================
// ModuleBuilder
// =====================================

func TestModuleBuilderDoc(t *testing.T) {
	mod := NewModuleBuilder("mymod").
		Doc("This is a test module.").
		Build()

	if mod.Doc != "This is a test module." {
		t.Errorf("Doc = %q", mod.Doc)
	}
	docVal, ok := mod.Dict["__doc__"]
	if !ok {
		t.Error("__doc__ not in Dict")
	}
	if s, ok := docVal.(*PyString); !ok || s.Value != "This is a test module." {
		t.Errorf("Dict[__doc__] = %v", docVal)
	}
}

func TestModuleBuilderConst(t *testing.T) {
	mod := NewModuleBuilder("mymod").
		Const("PI", &PyFloat{Value: 3.14159}).
		Const("VERSION", &PyString{Value: "1.0"}).
		Build()

	piVal, ok := mod.Dict["PI"]
	if !ok {
		t.Error("PI not found")
	}
	if f, ok := piVal.(*PyFloat); !ok || f.Value != 3.14159 {
		t.Errorf("PI = %v", piVal)
	}

	verVal, ok := mod.Dict["VERSION"]
	if !ok {
		t.Error("VERSION not found")
	}
	if s, ok := verVal.(*PyString); !ok || s.Value != "1.0" {
		t.Errorf("VERSION = %v", verVal)
	}
}

func TestModuleBuilderFunc(t *testing.T) {
	called := false
	mod := NewModuleBuilder("mymod").
		Func("do_stuff", func(vm *VM) int {
			called = true
			return 0
		}).
		Build()

	fnVal, ok := mod.Dict["do_stuff"]
	if !ok {
		t.Fatal("do_stuff not found")
	}
	fn, ok := fnVal.(*PyGoFunc)
	if !ok {
		t.Fatalf("expected *PyGoFunc, got %T", fnVal)
	}
	if fn.Name != "do_stuff" {
		t.Errorf("Name = %q, want %q", fn.Name, "do_stuff")
	}

	// Actually call it
	vm := NewVM()
	frame := &Frame{
		Code:  &CodeObject{Name: "<test>", Filename: "<test>", StackSize: 16},
		Stack: make([]Value, 16),
		SP:    0,
	}
	vm.frames = []*Frame{frame}
	vm.frame = frame
	fn.Fn(vm)
	if !called {
		t.Error("function was not called")
	}
}

func TestModuleBuilderMethod(t *testing.T) {
	// Method is an alias for Func
	mod := NewModuleBuilder("mymod").
		Method("helper", func(vm *VM) int { return 0 }).
		Build()

	_, ok := mod.Dict["helper"]
	if !ok {
		t.Error("helper not found (Method is alias for Func)")
	}
}

func TestModuleBuilderChaining(t *testing.T) {
	mod := NewModuleBuilder("mymod").
		Doc("test module").
		Const("A", MakeInt(1)).
		Const("B", MakeInt(2)).
		Func("f", func(vm *VM) int { return 0 }).
		Build()

	if mod.Name != "mymod" {
		t.Errorf("Name = %q", mod.Name)
	}
	if mod.Doc != "test module" {
		t.Errorf("Doc = %q", mod.Doc)
	}
	if _, ok := mod.Dict["A"]; !ok {
		t.Error("A not found")
	}
	if _, ok := mod.Dict["B"]; !ok {
		t.Error("B not found")
	}
	if _, ok := mod.Dict["f"]; !ok {
		t.Error("f not found")
	}
}

func TestModuleBuilderSubModule(t *testing.T) {
	sub := NewModule("parent.child")
	sub.Set("value", MakeInt(99))

	mod := NewModuleBuilder("parent").
		SubModule("child", sub).
		Build()

	childVal, ok := mod.Dict["child"]
	if !ok {
		t.Fatal("child submodule not found")
	}
	childMod, ok := childVal.(*PyModule)
	if !ok {
		t.Fatalf("expected *PyModule, got %T", childVal)
	}
	if childMod.Name != "parent.child" {
		t.Errorf("submodule Name = %q", childMod.Name)
	}
}

// =====================================
// RegisterModule and retrieval
// =====================================

func TestRegisterModuleAndImport(t *testing.T) {
	// Reset to clean state
	ResetModules()
	defer ResetModules()

	RegisterModule("test_reg_mod", func(vm *VM) *PyModule {
		mod := NewModule("test_reg_mod")
		mod.Set("answer", MakeInt(42))
		return mod
	})

	vm := NewVM()
	mod, err := vm.ImportModule("test_reg_mod")
	if err != nil {
		t.Fatalf("ImportModule failed: %v", err)
	}
	if mod == nil {
		t.Fatal("module is nil")
	}

	val, ok := mod.Get("answer")
	if !ok {
		t.Error("'answer' not found in module")
	}
	if v, ok := val.(*PyInt); !ok || v.Value != 42 {
		t.Errorf("answer = %v, want 42", val)
	}
}

func TestImportModuleCachesResult(t *testing.T) {
	ResetModules()
	defer ResetModules()

	callCount := 0
	RegisterModule("cached_mod", func(vm *VM) *PyModule {
		callCount++
		return NewModule("cached_mod")
	})

	vm := NewVM()
	mod1, err := vm.ImportModule("cached_mod")
	if err != nil {
		t.Fatalf("first import failed: %v", err)
	}

	mod2, err := vm.ImportModule("cached_mod")
	if err != nil {
		t.Fatalf("second import failed: %v", err)
	}

	if mod1 != mod2 {
		t.Error("expected same module on second import (cached)")
	}

	if callCount != 1 {
		t.Errorf("loader called %d times, want 1", callCount)
	}
}

func TestImportModuleNotFound(t *testing.T) {
	ResetModules()
	defer ResetModules()

	vm := NewVM()
	_, err := vm.ImportModule("nonexistent_module")
	if err == nil {
		t.Error("expected error for nonexistent module")
	}
}

// =====================================
// ResetModules
// =====================================

func TestResetModulesClearsState(t *testing.T) {
	ResetModules()

	RegisterModule("temp_mod", func(vm *VM) *PyModule {
		return NewModule("temp_mod")
	})

	vm := NewVM()
	_, err := vm.ImportModule("temp_mod")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	ResetModules()

	// After reset, module should not be loadable (registry cleared too, but only loadedModules is cleared)
	// Actually, ResetModules only clears loadedModules and moduleLoading, not moduleRegistry.
	// Let's verify the loaded cache is cleared by checking GetModule.
	_, ok := vm.GetModule("temp_mod")
	if ok {
		t.Error("expected module to be cleared after ResetModules")
	}
}

// =====================================
// VM.RegisterModule (instance method)
// =====================================

func TestVMRegisterModule(t *testing.T) {
	ResetModules()
	defer ResetModules()

	vm := NewVM()
	vm.RegisterModule("vm_mod", func(vm *VM) *PyModule {
		mod := NewModule("vm_mod")
		mod.Set("value", MakeInt(100))
		return mod
	})

	mod, err := vm.ImportModule("vm_mod")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	val, ok := mod.Get("value")
	if !ok || val.(*PyInt).Value != 100 {
		t.Errorf("value = %v, want 100", val)
	}
}

// =====================================
// VM.RegisterModuleInstance
// =====================================

func TestVMRegisterModuleInstance(t *testing.T) {
	ResetModules()
	defer ResetModules()

	vm := NewVM()
	mod := NewModule("direct_mod")
	mod.Set("data", &PyString{Value: "hello"})
	vm.RegisterModuleInstance("direct_mod", mod)

	retrieved, err := vm.ImportModule("direct_mod")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if retrieved != mod {
		t.Error("expected same module instance")
	}
}

// =====================================
// VM.GetModule
// =====================================

func TestGetModuleLoaded(t *testing.T) {
	ResetModules()
	defer ResetModules()

	vm := NewVM()
	mod := NewModule("get_test_mod")
	vm.RegisterModuleInstance("get_test_mod", mod)

	got, ok := vm.GetModule("get_test_mod")
	if !ok {
		t.Error("expected module to be found")
	}
	if got != mod {
		t.Error("expected same module")
	}
}

func TestGetModuleNotLoaded(t *testing.T) {
	ResetModules()
	defer ResetModules()

	vm := NewVM()
	_, ok := vm.GetModule("never_loaded")
	if ok {
		t.Error("expected module not found")
	}
}

// =====================================
// ResolveRelativeImport
// =====================================

func TestResolveRelativeImportLevel0(t *testing.T) {
	result, err := ResolveRelativeImport("os", 0, "mypackage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "os" {
		t.Errorf("got %q, want %q", result, "os")
	}
}

func TestResolveRelativeImportLevel1(t *testing.T) {
	// from . import foo (level=1, name="foo", package="mypackage")
	result, err := ResolveRelativeImport("foo", 1, "mypackage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "mypackage.foo" {
		t.Errorf("got %q, want %q", result, "mypackage.foo")
	}
}

func TestResolveRelativeImportLevel1NoName(t *testing.T) {
	// from . import (level=1, name="", package="mypackage")
	result, err := ResolveRelativeImport("", 1, "mypackage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "mypackage" {
		t.Errorf("got %q, want %q", result, "mypackage")
	}
}

func TestResolveRelativeImportLevel2(t *testing.T) {
	// from .. import bar (level=2, name="bar", package="mypackage.sub")
	result, err := ResolveRelativeImport("bar", 2, "mypackage.sub")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "mypackage.bar" {
		t.Errorf("got %q, want %q", result, "mypackage.bar")
	}
}

func TestResolveRelativeImportLevel1DeepPackage(t *testing.T) {
	// from . import baz (level=1, name="baz", package="a.b.c")
	result, err := ResolveRelativeImport("baz", 1, "a.b.c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "a.b.c.baz" {
		t.Errorf("got %q, want %q", result, "a.b.c.baz")
	}
}

func TestResolveRelativeImportLevel3DeepPackage(t *testing.T) {
	// from ... import x (level=3, name="x", package="a.b.c")
	result, err := ResolveRelativeImport("x", 3, "a.b.c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "a.x" {
		t.Errorf("got %q, want %q", result, "a.x")
	}
}

func TestResolveRelativeImportNoPackage(t *testing.T) {
	_, err := ResolveRelativeImport("foo", 1, "")
	if err == nil {
		t.Error("expected error for relative import with no package")
	}
}

func TestResolveRelativeImportBeyondTopLevel(t *testing.T) {
	// from ... import x (level=3, but package only has 1 part)
	_, err := ResolveRelativeImport("x", 3, "mypackage")
	if err == nil {
		t.Error("expected error for relative import beyond top-level")
	}
}

// =====================================
// ModuleBuilder Register
// =====================================

func TestModuleBuilderRegister(t *testing.T) {
	ResetModules()
	defer ResetModules()

	NewModuleBuilder("builder_mod").
		Doc("built by builder").
		Const("X", MakeInt(10)).
		Register()

	vm := NewVM()
	mod, err := vm.ImportModule("builder_mod")
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if mod.Name != "builder_mod" {
		t.Errorf("Name = %q", mod.Name)
	}
	val, ok := mod.Dict["X"]
	if !ok {
		t.Error("X not found")
	}
	if v, ok := val.(*PyInt); !ok || v.Value != 10 {
		t.Errorf("X = %v, want 10", val)
	}
}

// =====================================
// splitModuleName / joinModuleName
// =====================================

func TestSplitModuleName(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"os", []string{"os"}},
		{"a.b", []string{"a", "b"}},
		{"a.b.c.d", []string{"a", "b", "c", "d"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitModuleName(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("splitModuleName(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitModuleName(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestJoinModuleName(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"os"}, "os"},
		{[]string{"a", "b"}, "a.b"},
		{[]string{"a", "b", "c"}, "a.b.c"},
	}

	for _, tt := range tests {
		got := joinModuleName(tt.input)
		if got != tt.want {
			t.Errorf("joinModuleName(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// =====================================
// ModuleBuilder Type
// =====================================

func TestModuleBuilderType(t *testing.T) {
	ResetModules()
	defer ResetModules()

	mod := NewModuleBuilder("typed_mod").
		Type("Widget", func(vm *VM) int { return 0 }, map[string]GoFunction{
			"name": func(vm *VM) int { return 0 },
		}).
		Build()

	// Constructor should be in dict
	ctorVal, ok := mod.Dict["Widget"]
	if !ok {
		t.Fatal("Widget constructor not found in module dict")
	}
	fn, ok := ctorVal.(*PyGoFunc)
	if !ok {
		t.Fatalf("expected *PyGoFunc, got %T", ctorVal)
	}
	if fn.Name != "Widget" {
		t.Errorf("Name = %q, want %q", fn.Name, "Widget")
	}

	// Type metatable should be registered
	mt := GetRegisteredTypeMetatable("Widget")
	if mt == nil {
		t.Fatal("Widget type metatable not registered")
	}
	if _, ok := mt.Methods["name"]; !ok {
		t.Error("Widget.name method not registered")
	}
}
