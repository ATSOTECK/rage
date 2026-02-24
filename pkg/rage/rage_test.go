package rage

import (
	"context"
	"strings"
	"testing"
	"time"
)

// =====================================
// State Creation Tests
// =====================================

func TestNewState(t *testing.T) {
	state := NewState()
	defer state.Close()
	if state == nil {
		t.Fatal("NewState returned nil")
	}
}

func TestNewBareState(t *testing.T) {
	state := NewBareState()
	defer state.Close()
	if state == nil {
		t.Fatal("NewBareState returned nil")
	}
}

func TestNewStateWithModules(t *testing.T) {
	state := NewStateWithModules(
		WithModule(ModuleMath),
		WithModule(ModuleString),
	)
	defer state.Close()

	if !state.IsModuleEnabled(ModuleMath) {
		t.Error("expected ModuleMath enabled")
	}
	if !state.IsModuleEnabled(ModuleString) {
		t.Error("expected ModuleString enabled")
	}
	if state.IsModuleEnabled(ModuleJSON) {
		t.Error("expected ModuleJSON disabled")
	}
}

func TestNewStateWithAllModules(t *testing.T) {
	state := NewStateWithModules(WithAllModules())
	defer state.Close()

	for _, m := range AllModules {
		if !state.IsModuleEnabled(m) {
			t.Errorf("expected module %d to be enabled", m)
		}
	}
}

func TestNewStateWithBuiltins(t *testing.T) {
	state := NewStateWithModules(
		WithBuiltin(BuiltinRepr),
		WithBuiltin(BuiltinDir),
	)
	defer state.Close()

	if !state.IsBuiltinEnabled(BuiltinRepr) {
		t.Error("expected BuiltinRepr enabled")
	}
	if !state.IsBuiltinEnabled(BuiltinDir) {
		t.Error("expected BuiltinDir enabled")
	}
	if state.IsBuiltinEnabled(BuiltinExec) {
		t.Error("expected BuiltinExec disabled")
	}
}

func TestWithReflectionBuiltins(t *testing.T) {
	state := NewStateWithModules(WithReflectionBuiltins())
	defer state.Close()

	for _, b := range ReflectionBuiltins {
		if !state.IsBuiltinEnabled(b) {
			t.Errorf("expected reflection builtin %d enabled", b)
		}
	}
}

func TestWithExecutionBuiltins(t *testing.T) {
	state := NewStateWithModules(WithExecutionBuiltins())
	defer state.Close()

	for _, b := range ExecutionBuiltins {
		if !state.IsBuiltinEnabled(b) {
			t.Errorf("expected execution builtin %d enabled", b)
		}
	}
}

func TestWithAllBuiltins(t *testing.T) {
	state := NewStateWithModules(WithAllBuiltins())
	defer state.Close()

	for _, b := range AllBuiltins {
		if !state.IsBuiltinEnabled(b) {
			t.Errorf("expected builtin %d enabled", b)
		}
	}
}

func TestWithResourceLimits(t *testing.T) {
	state := NewStateWithModules(
		WithMaxRecursionDepth(50),
		WithMaxMemoryBytes(1024 * 1024),
		WithMaxCollectionSize(1000),
	)
	defer state.Close()
	// Just verify state creation succeeds - limits are checked at runtime
}

// =====================================
// State.Close Tests
// =====================================

func TestStateClose(t *testing.T) {
	state := NewState()
	state.Close()

	// Double close should be safe
	state.Close()

	// Operations on closed state should fail or be no-ops
	_, err := state.Run("x = 1")
	if err == nil {
		t.Error("expected error running on closed state")
	}
	if !strings.Contains(err.Error(), "closed") {
		t.Errorf("expected 'closed' in error, got %q", err.Error())
	}

	// SetGlobal on closed state should be a no-op
	state.SetGlobal("x", Int(1))

	// GetGlobal on closed state should return nil
	if state.GetGlobal("x") != nil {
		t.Error("expected nil from GetGlobal on closed state")
	}

	// GetGlobals on closed state should return nil
	if state.GetGlobals() != nil {
		t.Error("expected nil from GetGlobals on closed state")
	}
}

func TestStateClosedCompile(t *testing.T) {
	state := NewState()
	state.Close()

	// Compile doesn't need the VM, so it may still work
	// (the implementation notes say it doesn't need checkClosed)
	_, err := state.Compile("x = 1", "test")
	// This might succeed since Compile doesn't use VM
	_ = err
}

// =====================================
// State.Run Tests
// =====================================

func TestStateRun_Simple(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.Run("x = 1 + 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := state.GetGlobal("x")
	n, ok := AsInt(result)
	if !ok || n != 3 {
		t.Errorf("expected x=3, got %v", result)
	}
}

func TestStateRun_CompileError(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.Run("def")
	if err == nil {
		t.Fatal("expected compile error")
	}

	var ce *CompileErrors
	if !strings.Contains(err.Error(), "") {
		// Just verify we get an error
		_ = ce
	}
}

func TestStateRun_RuntimeError(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.Run("x = 1 / 0")
	if err == nil {
		t.Fatal("expected runtime error for division by zero")
	}
}

func TestStateRun_MultipleRuns(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.Run("x = 10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = state.Run("y = x + 20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := state.GetGlobal("y")
	n, ok := AsInt(result)
	if !ok || n != 30 {
		t.Errorf("expected y=30, got %v", result)
	}
}

// =====================================
// State.RunWithTimeout Tests
// =====================================

func TestStateRunWithTimeout_Success(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.RunWithTimeout("x = 42", 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("x")
	n, ok := AsInt(result)
	if !ok || n != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestStateRunWithTimeout_Timeout(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.RunWithTimeout("while True: pass", 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

// =====================================
// State.RunWithContext Tests
// =====================================

func TestStateRunWithContext_Success(t *testing.T) {
	state := NewState()
	defer state.Close()

	ctx := context.Background()
	_, err := state.RunWithContext(ctx, "x = 100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("x")
	n, ok := AsInt(result)
	if !ok || n != 100 {
		t.Errorf("expected 100, got %v", result)
	}
}

func TestStateRunWithContext_Cancel(t *testing.T) {
	state := NewState()
	defer state.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := state.RunWithContext(ctx, "while True: pass")
	if err == nil {
		t.Fatal("expected cancellation error")
	}
}

// =====================================
// State.Compile / Execute Tests
// =====================================

func TestStateCompileAndExecute(t *testing.T) {
	state := NewState()
	defer state.Close()

	code, err := state.Compile("x = 42", "test.py")
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	if code.Name() == "" {
		t.Error("expected non-empty code name")
	}

	_, err = state.Execute(code)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	result := state.GetGlobal("x")
	n, ok := AsInt(result)
	if !ok || n != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestStateCompile_Error(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.Compile("def", "test.py")
	if err == nil {
		t.Fatal("expected compile error")
	}
}

func TestStateExecuteWithTimeout(t *testing.T) {
	state := NewState()
	defer state.Close()

	code, err := state.Compile("x = 1 + 1", "test.py")
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	_, err = state.ExecuteWithTimeout(code, 5*time.Second)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	result := state.GetGlobal("x")
	n, ok := AsInt(result)
	if !ok || n != 2 {
		t.Errorf("expected 2, got %v", result)
	}
}

// =====================================
// Globals Tests
// =====================================

func TestSetGetGlobal(t *testing.T) {
	state := NewState()
	defer state.Close()

	state.SetGlobal("myvar", Int(99))
	result := state.GetGlobal("myvar")
	n, ok := AsInt(result)
	if !ok || n != 99 {
		t.Errorf("expected 99, got %v", result)
	}
}

func TestGetGlobal_Nonexistent(t *testing.T) {
	state := NewState()
	defer state.Close()

	result := state.GetGlobal("nonexistent")
	if result != nil && !IsNone(result) {
		t.Errorf("expected nil or None for nonexistent global, got %v", result)
	}
}

func TestGetGlobals(t *testing.T) {
	state := NewState()
	defer state.Close()

	state.SetGlobal("a", Int(1))
	state.SetGlobal("b", String("two"))

	globals := state.GetGlobals()
	if globals == nil {
		t.Fatal("GetGlobals returned nil")
	}

	if n, ok := AsInt(globals["a"]); !ok || n != 1 {
		t.Errorf("expected a=1, got %v", globals["a"])
	}
	if s, ok := AsString(globals["b"]); !ok || s != "two" {
		t.Errorf("expected b='two', got %v", globals["b"])
	}
}

func TestSetGlobal_AllTypes(t *testing.T) {
	state := NewState()
	defer state.Close()

	state.SetGlobal("i", Int(42))
	state.SetGlobal("f", Float(3.14))
	state.SetGlobal("s", String("hello"))
	state.SetGlobal("b", Bool(true))
	state.SetGlobal("n", None)
	state.SetGlobal("l", List(Int(1), Int(2)))
	state.SetGlobal("t", Tuple(Int(3)))

	_, err := state.Run(`
result = []
result.append(i == 42)
result.append(f == 3.14)
result.append(s == "hello")
result.append(b == True)
result.append(n is None)
result.append(len(l) == 2)
result.append(len(t) == 1)
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := state.GetGlobal("result")
	items, ok := AsList(result)
	if !ok {
		t.Fatalf("expected list result, got %v", result)
	}
	for i, item := range items {
		b, ok := AsBool(item)
		if !ok || !b {
			t.Errorf("check %d failed: got %v", i, item)
		}
	}
}

// =====================================
// Register Tests
// =====================================

func TestRegister(t *testing.T) {
	state := NewState()
	defer state.Close()

	state.Register("add_nums", func(s *State, args ...Value) Value {
		a, _ := AsInt(args[0])
		b, _ := AsInt(args[1])
		return Int(a + b)
	})

	_, err := state.Run("result = add_nums(3, 4)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := state.GetGlobal("result")
	n, ok := AsInt(result)
	if !ok || n != 7 {
		t.Errorf("expected 7, got %v", result)
	}
}

func TestRegister_ReturnNone(t *testing.T) {
	state := NewState()
	defer state.Close()

	called := false
	state.Register("side_effect", func(s *State, args ...Value) Value {
		called = true
		return nil
	})

	_, err := state.Run("side_effect()")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected function to be called")
	}
}

func TestRegisterBuiltin(t *testing.T) {
	state := NewState()
	defer state.Close()

	state.RegisterBuiltin("my_builtin", func(s *State, args ...Value) Value {
		return String("builtin_result")
	})

	_, err := state.Run("result = my_builtin()")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := state.GetGlobal("result")
	s, ok := AsString(result)
	if !ok || s != "builtin_result" {
		t.Errorf("expected 'builtin_result', got %v", result)
	}
}

// =====================================
// Module Tests
// =====================================

func TestEnableModule(t *testing.T) {
	state := NewBareState()
	defer state.Close()

	if state.IsModuleEnabled(ModuleMath) {
		t.Error("expected ModuleMath not enabled initially")
	}

	state.EnableModule(ModuleMath)
	if !state.IsModuleEnabled(ModuleMath) {
		t.Error("expected ModuleMath enabled after EnableModule")
	}

	// Should be able to use math after enabling
	_, err := state.Run(`
import math
result = math.sqrt(16)
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("result")
	f, ok := AsFloat(result)
	if !ok || f != 4.0 {
		t.Errorf("expected 4.0, got %v", result)
	}
}

func TestEnableModules(t *testing.T) {
	state := NewBareState()
	defer state.Close()

	state.EnableModules(ModuleMath, ModuleString)
	if !state.IsModuleEnabled(ModuleMath) {
		t.Error("expected ModuleMath enabled")
	}
	if !state.IsModuleEnabled(ModuleString) {
		t.Error("expected ModuleString enabled")
	}
}

func TestEnableAllModules(t *testing.T) {
	state := NewBareState()
	defer state.Close()

	state.EnableAllModules()
	for _, m := range AllModules {
		if !state.IsModuleEnabled(m) {
			t.Errorf("expected module %d enabled", m)
		}
	}
}

func TestEnabledModules(t *testing.T) {
	state := NewStateWithModules(
		WithModule(ModuleMath),
		WithModule(ModuleJSON),
	)
	defer state.Close()

	enabled := state.EnabledModules()
	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled modules, got %d", len(enabled))
	}
}

func TestEnableBuiltin(t *testing.T) {
	state := NewBareState()
	defer state.Close()

	state.EnableBuiltin(BuiltinRepr)
	if !state.IsBuiltinEnabled(BuiltinRepr) {
		t.Error("expected BuiltinRepr enabled")
	}

	_, err := state.Run("result = repr(42)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("result")
	s, ok := AsString(result)
	if !ok || s != "42" {
		t.Errorf("expected '42', got %v", result)
	}
}

func TestEnableBuiltins(t *testing.T) {
	state := NewBareState()
	defer state.Close()

	state.EnableBuiltins(BuiltinRepr, BuiltinDir)
	if !state.IsBuiltinEnabled(BuiltinRepr) {
		t.Error("expected BuiltinRepr enabled")
	}
	if !state.IsBuiltinEnabled(BuiltinDir) {
		t.Error("expected BuiltinDir enabled")
	}
}

func TestEnableReflectionBuiltins(t *testing.T) {
	state := NewBareState()
	defer state.Close()

	state.EnableReflectionBuiltins()
	for _, b := range ReflectionBuiltins {
		if !state.IsBuiltinEnabled(b) {
			t.Errorf("expected reflection builtin %d enabled", b)
		}
	}
}

func TestEnableExecutionBuiltins(t *testing.T) {
	state := NewBareState()
	defer state.Close()

	state.EnableExecutionBuiltins()
	for _, b := range ExecutionBuiltins {
		if !state.IsBuiltinEnabled(b) {
			t.Errorf("expected execution builtin %d enabled", b)
		}
	}
}

func TestEnableAllBuiltins(t *testing.T) {
	state := NewBareState()
	defer state.Close()

	state.EnableAllBuiltins()
	for _, b := range AllBuiltins {
		if !state.IsBuiltinEnabled(b) {
			t.Errorf("expected builtin %d enabled", b)
		}
	}
}

func TestEnabledBuiltins(t *testing.T) {
	state := NewStateWithModules(
		WithBuiltin(BuiltinRepr),
		WithBuiltin(BuiltinDir),
	)
	defer state.Close()

	enabled := state.EnabledBuiltins()
	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled builtins, got %d", len(enabled))
	}
}

func TestIsBuiltinEnabled_NilMap(t *testing.T) {
	// Test with a state that has nil builtins map
	state := NewBareState()
	defer state.Close()
	if state.IsBuiltinEnabled(BuiltinRepr) {
		t.Error("expected false for non-enabled builtin")
	}
}

// =====================================
// RegisterPythonModule Tests
// =====================================

func TestRegisterPythonModule(t *testing.T) {
	state := NewState()
	defer state.Close()

	err := state.RegisterPythonModule("mymod", `
MY_CONST = 42
def my_func():
    return "hello from mymod"
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = state.Run(`
from mymod import MY_CONST, my_func
result_const = MY_CONST
result_func = my_func()
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := state.GetGlobal("result_const")
	n, ok := AsInt(result)
	if !ok || n != 42 {
		t.Errorf("expected 42, got %v", result)
	}

	result = state.GetGlobal("result_func")
	s, ok := AsString(result)
	if !ok || s != "hello from mymod" {
		t.Errorf("expected 'hello from mymod', got %v", result)
	}
}

func TestRegisterPythonModule_DottedName(t *testing.T) {
	state := NewState()
	defer state.Close()

	err := state.RegisterPythonModule("pkg.sub.mod", `
VALUE = "nested"
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = state.Run(`
from pkg.sub.mod import VALUE
result = VALUE
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := state.GetGlobal("result")
	s, ok := AsString(result)
	if !ok || s != "nested" {
		t.Errorf("expected 'nested', got %v", result)
	}
}

func TestRegisterPythonModule_CompileError(t *testing.T) {
	state := NewState()
	defer state.Close()

	err := state.RegisterPythonModule("badmod", "def")
	if err == nil {
		t.Fatal("expected compile error")
	}
}

func TestRegisterPythonModule_ClosedState(t *testing.T) {
	state := NewState()
	state.Close()

	err := state.RegisterPythonModule("mymod", "x = 1")
	if err == nil {
		t.Fatal("expected error on closed state")
	}
}

// =====================================
// GetModuleAttr Tests
// =====================================

func TestGetModuleAttr(t *testing.T) {
	state := NewState()
	defer state.Close()

	err := state.RegisterPythonModule("testmod", `ANSWER = 42`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val := state.GetModuleAttr("testmod", "ANSWER")
	n, ok := AsInt(val)
	if !ok || n != 42 {
		t.Errorf("expected 42, got %v", val)
	}
}

func TestGetModuleAttr_NonexistentModule(t *testing.T) {
	state := NewState()
	defer state.Close()

	val := state.GetModuleAttr("nonexistent", "attr")
	if val != nil {
		t.Errorf("expected nil, got %v", val)
	}
}

func TestGetModuleAttr_NonexistentAttr(t *testing.T) {
	state := NewState()
	defer state.Close()

	err := state.RegisterPythonModule("testmod2", `X = 1`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val := state.GetModuleAttr("testmod2", "nonexistent")
	if val != nil {
		t.Errorf("expected nil, got %v", val)
	}
}

func TestGetModuleAttr_ClosedState(t *testing.T) {
	state := NewState()
	state.Close()

	val := state.GetModuleAttr("mod", "attr")
	if val != nil {
		t.Errorf("expected nil on closed state, got %v", val)
	}
}

// =====================================
// Resource Limits Tests
// =====================================

func TestSetMaxRecursionDepth(t *testing.T) {
	state := NewState()
	defer state.Close()

	state.SetMaxRecursionDepth(10)

	_, err := state.Run(`
def recurse(n):
    return recurse(n + 1)
try:
    recurse(0)
except RecursionError:
    result = "caught"
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := state.GetGlobal("result")
	s, ok := AsString(result)
	if !ok || s != "caught" {
		t.Errorf("expected 'caught', got %v", result)
	}
}

func TestSetMaxRecursionDepth_ClosedState(t *testing.T) {
	state := NewState()
	state.Close()
	// Should not panic
	state.SetMaxRecursionDepth(10)
}

func TestSetMaxMemoryBytes(t *testing.T) {
	state := NewState()
	defer state.Close()

	state.SetMaxMemoryBytes(1024 * 1024) // 1MB
	// Just verify it doesn't panic
}

func TestSetMaxMemoryBytes_ClosedState(t *testing.T) {
	state := NewState()
	state.Close()
	state.SetMaxMemoryBytes(1024)
}

func TestSetMaxCollectionSize(t *testing.T) {
	state := NewState()
	defer state.Close()

	state.SetMaxCollectionSize(100)
	// Just verify it doesn't panic
}

func TestSetMaxCollectionSize_ClosedState(t *testing.T) {
	state := NewState()
	state.Close()
	state.SetMaxCollectionSize(100)
}

func TestAllocatedBytes(t *testing.T) {
	state := NewState()
	defer state.Close()

	// After running some code, allocated bytes should be >= 0
	_, _ = state.Run("x = [1, 2, 3, 4, 5]")
	bytes := state.AllocatedBytes()
	if bytes < 0 {
		t.Errorf("expected non-negative allocated bytes, got %d", bytes)
	}
}

func TestAllocatedBytes_ClosedState(t *testing.T) {
	state := NewState()
	state.Close()

	bytes := state.AllocatedBytes()
	if bytes != 0 {
		t.Errorf("expected 0 for closed state, got %d", bytes)
	}
}

// =====================================
// Package-level Convenience Functions
// =====================================

func TestPackageRun(t *testing.T) {
	_, err := Run("x = 1 + 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPackageRunWithTimeout(t *testing.T) {
	_, err := RunWithTimeout("x = 42", 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPackageEval(t *testing.T) {
	result, err := Eval("1 + 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n, ok := AsInt(result)
	if !ok || n != 3 {
		t.Errorf("expected 3, got %v", result)
	}
}

func TestPackageEval_String(t *testing.T) {
	result, err := Eval(`"hello" + " world"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s, ok := AsString(result)
	if !ok || s != "hello world" {
		t.Errorf("expected 'hello world', got %v", result)
	}
}

func TestPackageEval_Error(t *testing.T) {
	_, err := Eval("1 / 0")
	if err == nil {
		t.Fatal("expected error")
	}
}

// =====================================
// CompileErrors Tests
// =====================================

func TestCompileErrors_Single(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.Run("def")
	if err == nil {
		t.Fatal("expected error")
	}
	// Error message should be present
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestCompileErrors_Unwrap(t *testing.T) {
	ce := &CompileErrors{Errors: []error{
		TypeError("test error"),
	}}
	unwrapped := ce.Unwrap()
	if unwrapped == nil {
		t.Error("Unwrap should return first error")
	}
}

func TestCompileErrors_UnwrapEmpty(t *testing.T) {
	ce := &CompileErrors{Errors: []error{}}
	unwrapped := ce.Unwrap()
	if unwrapped != nil {
		t.Error("Unwrap on empty errors should return nil")
	}
}

// =====================================
// Error Helper Tests
// =====================================

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"TypeError", TypeError("bad type"), "TypeError: bad type"},
		{"ValueError", ValueError("bad value"), "ValueError: bad value"},
		{"KeyError", KeyError("missing key"), "KeyError: missing key"},
		{"IndexError", IndexError("out of range"), "IndexError: out of range"},
		{"AttributeError", AttributeError("no attr"), "AttributeError: no attr"},
		{"RuntimeError", RuntimeError("runtime fail"), "RuntimeError: runtime fail"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("expected %q, got %q", tt.want, tt.err.Error())
			}
		})
	}
}

// =====================================
// State.Call Tests
// =====================================

func TestStateCall_PythonFunction(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.Run(`
def multiply(a, b):
    return a * b
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fn := state.GetGlobal("multiply")
	result, err := state.Call(fn, Int(6), Int(7))
	if err != nil {
		t.Fatalf("call error: %v", err)
	}

	n, ok := AsInt(result)
	if !ok || n != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestStateCall_ClosedState(t *testing.T) {
	state := NewState()
	state.Close()

	_, err := state.Call(None)
	if err == nil {
		t.Fatal("expected error on closed state")
	}
}

// =====================================
// Helper Function Tests
// =====================================

func TestSplitModuleName(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"foo", []string{"foo"}},
		{"foo.bar", []string{"foo", "bar"}},
		{"a.b.c", []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		result := splitModuleName(tt.input)
		if len(result) != len(tt.want) {
			t.Errorf("splitModuleName(%q): expected %v, got %v", tt.input, tt.want, result)
			continue
		}
		for i := range result {
			if result[i] != tt.want[i] {
				t.Errorf("splitModuleName(%q)[%d]: expected %q, got %q", tt.input, i, tt.want[i], result[i])
			}
		}
	}
}

func TestJoinModuleName(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"foo"}, "foo"},
		{[]string{"a", "b", "c"}, "a.b.c"},
	}
	for _, tt := range tests {
		result := joinModuleName(tt.input)
		if result != tt.want {
			t.Errorf("joinModuleName(%v): expected %q, got %q", tt.input, tt.want, result)
		}
	}
}

func TestLastIndexByte(t *testing.T) {
	tests := []struct {
		s    string
		c    byte
		want int
	}{
		{"", '.', -1},
		{"foo", '.', -1},
		{"foo.bar", '.', 3},
		{"a.b.c", '.', 3},
	}
	for _, tt := range tests {
		result := lastIndexByte(tt.s, tt.c)
		if result != tt.want {
			t.Errorf("lastIndexByte(%q, %c): expected %d, got %d", tt.s, tt.c, tt.want, result)
		}
	}
}

// =====================================
// Integration-style Tests
// =====================================

func TestStateRunPythonFeatures(t *testing.T) {
	state := NewState()
	defer state.Close()

	// Test list comprehension
	_, err := state.Run("squares = [x**2 for x in range(5)]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("squares")
	items, ok := AsList(result)
	if !ok || len(items) != 5 {
		t.Fatalf("expected list of 5, got %v", result)
	}
	expected := []int64{0, 1, 4, 9, 16}
	for i, want := range expected {
		n, _ := AsInt(items[i])
		if n != want {
			t.Errorf("squares[%d]: expected %d, got %d", i, want, n)
		}
	}
}

func TestStateRunClasses(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.Run(`
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y
    def magnitude(self):
        return (self.x ** 2 + self.y ** 2) ** 0.5

p = Point(3, 4)
mag = p.magnitude()
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := state.GetGlobal("mag")
	f, ok := AsFloat(result)
	if !ok || f != 5.0 {
		t.Errorf("expected 5.0, got %v", result)
	}
}

func TestStateRunExceptions(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.Run(`
try:
    x = 1 / 0
except ZeroDivisionError as e:
    caught = str(e)
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := state.GetGlobal("caught")
	s, ok := AsString(result)
	if !ok || s == "" {
		t.Errorf("expected non-empty caught message, got %v", result)
	}
}

func TestStateRunWithMathModule(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.Run(`
import math
result = math.pi
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := state.GetGlobal("result")
	f, ok := AsFloat(result)
	if !ok || f < 3.14 || f > 3.15 {
		t.Errorf("expected ~3.14159, got %v", result)
	}
}

func TestGoFuncAccessState(t *testing.T) {
	state := NewState()
	defer state.Close()

	state.Register("set_via_go", func(s *State, args ...Value) Value {
		s.SetGlobal("go_set_value", Int(123))
		return None
	})

	_, err := state.Run("set_via_go()")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := state.GetGlobal("go_set_value")
	n, ok := AsInt(result)
	if !ok || n != 123 {
		t.Errorf("expected 123, got %v", result)
	}
}
