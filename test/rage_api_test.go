package test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ATSOTECK/rage/pkg/rage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Package-level convenience functions
// =============================================================================

func TestRageRun(t *testing.T) {
	result, err := rage.Run(`x = 1 + 2`)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestRageRunReturnsNilOnEmpty(t *testing.T) {
	result, err := rage.Run(`pass`)
	require.NoError(t, err)
	_ = result
}

func TestRageRunCompileError(t *testing.T) {
	_, err := rage.Run(`def`)
	require.Error(t, err)
	var compErr *rage.CompileErrors
	require.ErrorAs(t, err, &compErr)
	assert.Greater(t, len(compErr.Errors), 0)
}

func TestRageRunRuntimeError(t *testing.T) {
	_, err := rage.Run(`x = 1 / 0`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ZeroDivisionError")
}

func TestRageEval(t *testing.T) {
	result, err := rage.Eval(`2 ** 10`)
	require.NoError(t, err)
	require.NotNil(t, result)
	n, ok := rage.AsInt(result)
	assert.True(t, ok)
	assert.Equal(t, int64(1024), n)
}

func TestRageRunWithTimeout(t *testing.T) {
	_, err := rage.RunWithTimeout(`
x = 0
while True:
    x += 1
`, 50*time.Millisecond)
	require.Error(t, err)
}

// =============================================================================
// State lifecycle
// =============================================================================

func TestStateCreateAndClose(t *testing.T) {
	state := rage.NewState()
	assert.NotNil(t, state)
	state.Close()

	// Operations on a closed state should return errors.
	_, err := state.Run(`x = 1`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

func TestStateCloseIdempotent(t *testing.T) {
	state := rage.NewState()
	state.Close()
	state.Close() // Should not panic.
}

func TestBareState(t *testing.T) {
	state := rage.NewBareState()
	defer state.Close()

	// Core builtins should still work without modules.
	_, err := state.Run(`x = len([1, 2, 3])`)
	require.NoError(t, err)
	v := state.GetGlobal("x")
	n, ok := rage.AsInt(v)
	assert.True(t, ok)
	assert.Equal(t, int64(3), n)
}

func TestBareStateModuleNotEnabled(t *testing.T) {
	state := rage.NewBareState()
	defer state.Close()

	assert.False(t, state.IsModuleEnabled(rage.ModuleMath))
	assert.False(t, state.IsModuleEnabled(rage.ModuleJSON))
	assert.Equal(t, 0, len(state.EnabledModules()))
}

// =============================================================================
// Module selection
// =============================================================================

func TestStateWithSpecificModules(t *testing.T) {
	state := rage.NewStateWithModules(rage.WithModule(rage.ModuleMath))
	defer state.Close()

	_, err := state.Run(`import math; result = math.sqrt(16)`)
	require.NoError(t, err)

	v := state.GetGlobal("result")
	f, ok := rage.AsFloat(v)
	assert.True(t, ok)
	assert.Equal(t, 4.0, f)
}

func TestStateWithMultipleModules(t *testing.T) {
	state := rage.NewStateWithModules(
		rage.WithModules(rage.ModuleMath, rage.ModuleString),
	)
	defer state.Close()

	_, err := state.Run(`
import math
import string
result = string.ascii_lowercase
pi = math.pi
`)
	require.NoError(t, err)

	s, ok := rage.AsString(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.Equal(t, "abcdefghijklmnopqrstuvwxyz", s)

	pi, ok := rage.AsFloat(state.GetGlobal("pi"))
	assert.True(t, ok)
	assert.InDelta(t, 3.14159, pi, 0.001)
}

func TestEnableModuleAfterCreation(t *testing.T) {
	state := rage.NewBareState()
	defer state.Close()

	assert.False(t, state.IsModuleEnabled(rage.ModuleJSON))

	state.EnableModule(rage.ModuleJSON)
	assert.True(t, state.IsModuleEnabled(rage.ModuleJSON))

	_, err := state.Run(`import json; result = json.dumps([1, 2])`)
	require.NoError(t, err)

	s, ok := rage.AsString(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.Equal(t, "[1,2]", s)
}

func TestEnableAllModulesAfterCreation(t *testing.T) {
	state := rage.NewBareState()
	defer state.Close()

	state.EnableAllModules()
	assert.True(t, state.IsModuleEnabled(rage.ModuleMath))
	assert.True(t, state.IsModuleEnabled(rage.ModuleJSON))
	assert.True(t, state.IsModuleEnabled(rage.ModuleRe))
}

func TestEnabledModulesListing(t *testing.T) {
	state := rage.NewStateWithModules(
		rage.WithModules(rage.ModuleMath, rage.ModuleString),
	)
	defer state.Close()

	modules := state.EnabledModules()
	assert.Equal(t, 2, len(modules))
}

// =============================================================================
// Opt-in builtins
// =============================================================================

func TestBuiltinsDisabledByDefault(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	assert.False(t, state.IsBuiltinEnabled(rage.BuiltinExec))
	assert.False(t, state.IsBuiltinEnabled(rage.BuiltinEval))
}

func TestEnableBuiltins(t *testing.T) {
	state := rage.NewStateWithModules(
		rage.WithAllModules(),
		rage.WithBuiltin(rage.BuiltinRepr),
	)
	defer state.Close()

	assert.True(t, state.IsBuiltinEnabled(rage.BuiltinRepr))
	assert.False(t, state.IsBuiltinEnabled(rage.BuiltinExec))

	_, err := state.Run(`result = repr(42)`)
	require.NoError(t, err)

	s, ok := rage.AsString(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.Equal(t, "42", s)
}

func TestEnableReflectionBuiltins(t *testing.T) {
	state := rage.NewStateWithModules(rage.WithAllModules(), rage.WithReflectionBuiltins())
	defer state.Close()

	assert.True(t, state.IsBuiltinEnabled(rage.BuiltinRepr))
	assert.True(t, state.IsBuiltinEnabled(rage.BuiltinDir))
	assert.False(t, state.IsBuiltinEnabled(rage.BuiltinExec))
}

func TestEnableExecutionBuiltins(t *testing.T) {
	state := rage.NewStateWithModules(rage.WithAllModules(), rage.WithExecutionBuiltins())
	defer state.Close()

	assert.True(t, state.IsBuiltinEnabled(rage.BuiltinExec))
	assert.True(t, state.IsBuiltinEnabled(rage.BuiltinEval))
}

func TestEnableBuiltinAfterCreation(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.EnableBuiltin(rage.BuiltinRepr)
	assert.True(t, state.IsBuiltinEnabled(rage.BuiltinRepr))

	_, err := state.Run(`result = repr("hello")`)
	require.NoError(t, err)

	s, ok := rage.AsString(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.Equal(t, "'hello'", s)
}

// =============================================================================
// SetGlobal / GetGlobal
// =============================================================================

func TestSetAndGetGlobals(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.SetGlobal("x", rage.Int(42))
	state.SetGlobal("name", rage.String("Alice"))
	state.SetGlobal("pi", rage.Float(3.14))
	state.SetGlobal("active", rage.Bool(true))

	_, err := state.Run(`
result_int = x * 2
result_str = name + " Bob"
result_float = pi * 2
result_bool = active and True
`)
	require.NoError(t, err)

	n, ok := rage.AsInt(state.GetGlobal("result_int"))
	assert.True(t, ok)
	assert.Equal(t, int64(84), n)

	s, ok := rage.AsString(state.GetGlobal("result_str"))
	assert.True(t, ok)
	assert.Equal(t, "Alice Bob", s)

	f, ok := rage.AsFloat(state.GetGlobal("result_float"))
	assert.True(t, ok)
	assert.InDelta(t, 6.28, f, 0.001)

	b, ok := rage.AsBool(state.GetGlobal("result_bool"))
	assert.True(t, ok)
	assert.True(t, b)
}

func TestGetGlobalNonexistent(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	v := state.GetGlobal("nonexistent")
	assert.True(t, rage.IsNone(v))
}

func TestGetGlobalsMap(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	_, err := state.Run(`x = 10; y = 20`)
	require.NoError(t, err)

	globals := state.GetGlobals()
	assert.NotNil(t, globals)

	xVal, ok := rage.AsInt(globals["x"])
	assert.True(t, ok)
	assert.Equal(t, int64(10), xVal)

	yVal, ok := rage.AsInt(globals["y"])
	assert.True(t, ok)
	assert.Equal(t, int64(20), yVal)
}

func TestSetGlobalCollections(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.SetGlobal("items", rage.List(rage.Int(1), rage.Int(2), rage.Int(3)))
	state.SetGlobal("pair", rage.Tuple(rage.String("a"), rage.String("b")))
	state.SetGlobal("data", rage.Dict("key", rage.String("value"), "count", rage.Int(5)))

	_, err := state.Run(`
list_len = len(items)
tuple_len = len(pair)
dict_val = data["key"]
dict_count = data["count"]
`)
	require.NoError(t, err)

	n, ok := rage.AsInt(state.GetGlobal("list_len"))
	assert.True(t, ok)
	assert.Equal(t, int64(3), n)

	n, ok = rage.AsInt(state.GetGlobal("tuple_len"))
	assert.True(t, ok)
	assert.Equal(t, int64(2), n)

	s, ok := rage.AsString(state.GetGlobal("dict_val"))
	assert.True(t, ok)
	assert.Equal(t, "value", s)

	n, ok = rage.AsInt(state.GetGlobal("dict_count"))
	assert.True(t, ok)
	assert.Equal(t, int64(5), n)
}

func TestSetGlobalComplex(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.SetGlobal("c", rage.Complex(1.0, 2.0))

	_, err := state.Run(`result = c + (3+4j)`)
	require.NoError(t, err)

	v := state.GetGlobal("result")
	re, im, ok := rage.AsComplex(v)
	assert.True(t, ok)
	assert.Equal(t, 4.0, re)
	assert.Equal(t, 6.0, im)
}

// =============================================================================
// Register Go functions
// =============================================================================

func TestRegisterGoFunction(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.Register("double", func(s *rage.State, args ...rage.Value) rage.Value {
		n, _ := rage.AsInt(args[0])
		return rage.Int(n * 2)
	})

	_, err := state.Run(`result = double(21)`)
	require.NoError(t, err)

	n, ok := rage.AsInt(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.Equal(t, int64(42), n)
}

func TestRegisterGoFunctionNoReturn(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	var captured string
	state.Register("log_msg", func(s *rage.State, args ...rage.Value) rage.Value {
		captured, _ = rage.AsString(args[0])
		return nil
	})

	_, err := state.Run(`log_msg("hello from python")`)
	require.NoError(t, err)
	assert.Equal(t, "hello from python", captured)
}

func TestRegisterGoFunctionMultipleArgs(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.Register("add3", func(s *rage.State, args ...rage.Value) rage.Value {
		a, _ := rage.AsInt(args[0])
		b, _ := rage.AsInt(args[1])
		c, _ := rage.AsInt(args[2])
		return rage.Int(a + b + c)
	})

	_, err := state.Run(`result = add3(10, 20, 12)`)
	require.NoError(t, err)

	n, ok := rage.AsInt(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.Equal(t, int64(42), n)
}

func TestRegisterGoFunctionReturningList(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.Register("make_list", func(s *rage.State, args ...rage.Value) rage.Value {
		n, _ := rage.AsInt(args[0])
		items := make([]rage.Value, n)
		for i := int64(0); i < n; i++ {
			items[i] = rage.Int(i * i)
		}
		return rage.List(items...)
	})

	_, err := state.Run(`
squares = make_list(5)
total = sum(squares)
`)
	require.NoError(t, err)

	n, ok := rage.AsInt(state.GetGlobal("total"))
	assert.True(t, ok)
	assert.Equal(t, int64(0+1+4+9+16), n)
}

func TestRegisterGoFunctionReturningDict(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.Register("make_person", func(s *rage.State, args ...rage.Value) rage.Value {
		name, _ := rage.AsString(args[0])
		age, _ := rage.AsInt(args[1])
		return rage.Dict("name", rage.String(name), "age", rage.Int(age))
	})

	_, err := state.Run(`
p = make_person("Alice", 30)
result_name = p["name"]
result_age = p["age"]
`)
	require.NoError(t, err)

	s, ok := rage.AsString(state.GetGlobal("result_name"))
	assert.True(t, ok)
	assert.Equal(t, "Alice", s)

	n, ok := rage.AsInt(state.GetGlobal("result_age"))
	assert.True(t, ok)
	assert.Equal(t, int64(30), n)
}

func TestRegisterGoFunctionAccessesState(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.SetGlobal("multiplier", rage.Int(10))

	state.Register("scaled", func(s *rage.State, args ...rage.Value) rage.Value {
		n, _ := rage.AsInt(args[0])
		m, _ := rage.AsInt(s.GetGlobal("multiplier"))
		return rage.Int(n * m)
	})

	_, err := state.Run(`result = scaled(5)`)
	require.NoError(t, err)

	n, ok := rage.AsInt(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.Equal(t, int64(50), n)
}

// =============================================================================
// Compile / Execute
// =============================================================================

func TestCompileAndExecute(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	code, err := state.Compile(`result = x * 2`, "multiply.py")
	require.NoError(t, err)
	assert.NotEmpty(t, code.Name())

	state.SetGlobal("x", rage.Int(21))
	_, err = state.Execute(code)
	require.NoError(t, err)

	n, ok := rage.AsInt(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.Equal(t, int64(42), n)
}

func TestCompileOnceRunMany(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	code, err := state.Compile(`result = n ** 2`, "square.py")
	require.NoError(t, err)

	expected := []int64{0, 1, 4, 9, 16}
	for i, want := range expected {
		state.SetGlobal("n", rage.Int(int64(i)))
		_, err = state.Execute(code)
		require.NoError(t, err)

		got, ok := rage.AsInt(state.GetGlobal("result"))
		assert.True(t, ok)
		assert.Equal(t, want, got)
	}
}

func TestCompileError(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	_, err := state.Compile(`def`, "bad.py")
	require.Error(t, err)

	var compErr *rage.CompileErrors
	require.ErrorAs(t, err, &compErr)
}

func TestExecuteWithTimeout(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	code, err := state.Compile(`
x = 0
while True:
    x += 1
`, "infinite.py")
	require.NoError(t, err)

	_, err = state.ExecuteWithTimeout(code, 50*time.Millisecond)
	require.Error(t, err)
}

// =============================================================================
// Timeouts and cancellation
// =============================================================================

func TestRunWithTimeoutSuccess(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	_, err := state.RunWithTimeout(`x = sum(range(100))`, 5*time.Second)
	require.NoError(t, err)

	n, ok := rage.AsInt(state.GetGlobal("x"))
	assert.True(t, ok)
	assert.Equal(t, int64(4950), n)
}

func TestRunWithTimeoutExceeded(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	_, err := state.RunWithTimeout(`
while True:
    pass
`, 50*time.Millisecond)
	require.Error(t, err)
}

func TestRunWithContextCancellation(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := state.RunWithContext(ctx, `
while True:
    pass
`)
	require.Error(t, err)
}

func TestRunWithContextSuccess(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	ctx := context.Background()
	_, err := state.RunWithContext(ctx, `result = 42`)
	require.NoError(t, err)

	n, ok := rage.AsInt(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.Equal(t, int64(42), n)
}

// =============================================================================
// Value constructors and type checks
// =============================================================================

func TestValueConstructors(t *testing.T) {
	assert.Equal(t, "NoneType", rage.None.Type())
	assert.Equal(t, "None", rage.None.String())
	assert.Nil(t, rage.None.GoValue())

	assert.Equal(t, "bool", rage.True.Type())
	assert.Equal(t, "bool", rage.False.Type())

	i := rage.Int(42)
	assert.Equal(t, "int", i.Type())
	assert.Equal(t, "42", i.String())
	assert.Equal(t, int64(42), i.GoValue())

	f := rage.Float(3.14)
	assert.Equal(t, "float", f.Type())
	assert.Equal(t, float64(3.14), f.GoValue())

	c := rage.Complex(1, 2)
	assert.Equal(t, "complex", c.Type())
	assert.Equal(t, complex(1.0, 2.0), c.GoValue())

	s := rage.String("hello")
	assert.Equal(t, "str", s.Type())
	assert.Equal(t, "hello", s.String())
	assert.Equal(t, "hello", s.GoValue())
}

func TestValueTypeChecks(t *testing.T) {
	assert.True(t, rage.IsNone(rage.None))
	assert.False(t, rage.IsNone(rage.Int(0)))

	assert.True(t, rage.IsBool(rage.True))
	assert.True(t, rage.IsBool(rage.False))
	assert.False(t, rage.IsBool(rage.Int(1)))

	assert.True(t, rage.IsInt(rage.Int(42)))
	assert.False(t, rage.IsInt(rage.Float(42.0)))

	assert.True(t, rage.IsFloat(rage.Float(1.0)))
	assert.False(t, rage.IsFloat(rage.Int(1)))

	assert.True(t, rage.IsComplex(rage.Complex(1, 2)))
	assert.False(t, rage.IsComplex(rage.Float(1.0)))

	assert.True(t, rage.IsString(rage.String("hi")))
	assert.False(t, rage.IsString(rage.Int(0)))

	assert.True(t, rage.IsList(rage.List(rage.Int(1))))
	assert.False(t, rage.IsList(rage.Tuple(rage.Int(1))))

	assert.True(t, rage.IsTuple(rage.Tuple(rage.Int(1))))
	assert.False(t, rage.IsTuple(rage.List(rage.Int(1))))

	assert.True(t, rage.IsDict(rage.Dict("k", rage.Int(1))))
	assert.False(t, rage.IsDict(rage.List()))

	assert.True(t, rage.IsUserData(rage.UserData(42)))
	assert.False(t, rage.IsUserData(rage.Int(42)))
}

func TestValueAssertionHelpers(t *testing.T) {
	// AsInt
	n, ok := rage.AsInt(rage.Int(42))
	assert.True(t, ok)
	assert.Equal(t, int64(42), n)
	_, ok = rage.AsInt(rage.String("nope"))
	assert.False(t, ok)

	// AsFloat (also accepts int)
	f, ok := rage.AsFloat(rage.Float(3.14))
	assert.True(t, ok)
	assert.Equal(t, 3.14, f)
	f, ok = rage.AsFloat(rage.Int(5))
	assert.True(t, ok)
	assert.Equal(t, 5.0, f)

	// AsString
	s, ok := rage.AsString(rage.String("hello"))
	assert.True(t, ok)
	assert.Equal(t, "hello", s)
	_, ok = rage.AsString(rage.Int(0))
	assert.False(t, ok)

	// AsBool
	b, ok := rage.AsBool(rage.True)
	assert.True(t, ok)
	assert.True(t, b)
	b, ok = rage.AsBool(rage.False)
	assert.True(t, ok)
	assert.False(t, b)

	// AsComplex
	re, im, ok := rage.AsComplex(rage.Complex(3, 4))
	assert.True(t, ok)
	assert.Equal(t, 3.0, re)
	assert.Equal(t, 4.0, im)
	_, _, ok = rage.AsComplex(rage.Int(0))
	assert.False(t, ok)

	// AsList
	items, ok := rage.AsList(rage.List(rage.Int(1), rage.Int(2)))
	assert.True(t, ok)
	assert.Equal(t, 2, len(items))
	_, ok = rage.AsList(rage.Tuple(rage.Int(1)))
	assert.False(t, ok)

	// AsTuple
	items, ok = rage.AsTuple(rage.Tuple(rage.String("a"), rage.String("b")))
	assert.True(t, ok)
	assert.Equal(t, 2, len(items))
	_, ok = rage.AsTuple(rage.List(rage.Int(1)))
	assert.False(t, ok)

	// AsDict
	d, ok := rage.AsDict(rage.Dict("x", rage.Int(1)))
	assert.True(t, ok)
	assert.Equal(t, 1, len(d))
	_, ok = rage.AsDict(rage.List())
	assert.False(t, ok)

	// AsUserData
	ud, ok := rage.AsUserData(rage.UserData("payload"))
	assert.True(t, ok)
	assert.Equal(t, "payload", ud)
	_, ok = rage.AsUserData(rage.Int(0))
	assert.False(t, ok)
}

func TestListValueMethods(t *testing.T) {
	l := rage.List(rage.Int(10), rage.Int(20), rage.Int(30))
	lv, ok := l.(rage.ListValue)
	require.True(t, ok)

	assert.Equal(t, 3, lv.Len())

	v0, ok := rage.AsInt(lv.Get(0))
	assert.True(t, ok)
	assert.Equal(t, int64(10), v0)

	v2, ok := rage.AsInt(lv.Get(2))
	assert.True(t, ok)
	assert.Equal(t, int64(30), v2)

	// Out of bounds returns None.
	assert.True(t, rage.IsNone(lv.Get(99)))
	assert.True(t, rage.IsNone(lv.Get(-1)))
}

func TestTupleValueMethods(t *testing.T) {
	tup := rage.Tuple(rage.String("a"), rage.String("b"))
	tv, ok := tup.(rage.TupleValue)
	require.True(t, ok)

	assert.Equal(t, 2, tv.Len())

	s, ok := rage.AsString(tv.Get(0))
	assert.True(t, ok)
	assert.Equal(t, "a", s)

	assert.True(t, rage.IsNone(tv.Get(5)))
}

func TestDictValueMethods(t *testing.T) {
	d := rage.Dict("name", rage.String("Bob"), "age", rage.Int(25))
	dv, ok := d.(rage.DictValue)
	require.True(t, ok)

	assert.Equal(t, 2, dv.Len())

	name, ok := rage.AsString(dv.Get("name"))
	assert.True(t, ok)
	assert.Equal(t, "Bob", name)

	assert.True(t, rage.IsNone(dv.Get("missing")))
}

// =============================================================================
// FromGo conversion
// =============================================================================

func TestFromGo(t *testing.T) {
	// nil -> None
	assert.True(t, rage.IsNone(rage.FromGo(nil)))

	// bool
	b, ok := rage.AsBool(rage.FromGo(true))
	assert.True(t, ok)
	assert.True(t, b)

	// int types
	n, ok := rage.AsInt(rage.FromGo(42))
	assert.True(t, ok)
	assert.Equal(t, int64(42), n)

	n, ok = rage.AsInt(rage.FromGo(int64(99)))
	assert.True(t, ok)
	assert.Equal(t, int64(99), n)

	n, ok = rage.AsInt(rage.FromGo(int32(7)))
	assert.True(t, ok)
	assert.Equal(t, int64(7), n)

	n, ok = rage.AsInt(rage.FromGo(uint16(100)))
	assert.True(t, ok)
	assert.Equal(t, int64(100), n)

	// float
	f, ok := rage.AsFloat(rage.FromGo(2.718))
	assert.True(t, ok)
	assert.InDelta(t, 2.718, f, 0.001)

	f, ok = rage.AsFloat(rage.FromGo(float32(1.5)))
	assert.True(t, ok)
	assert.InDelta(t, 1.5, f, 0.001)

	// complex
	re, im, ok := rage.AsComplex(rage.FromGo(complex(3, 4)))
	assert.True(t, ok)
	assert.Equal(t, 3.0, re)
	assert.Equal(t, 4.0, im)

	// string
	s, ok := rage.AsString(rage.FromGo("hello"))
	assert.True(t, ok)
	assert.Equal(t, "hello", s)

	// []any -> list
	items, ok := rage.AsList(rage.FromGo([]any{1, "two", 3.0}))
	assert.True(t, ok)
	assert.Equal(t, 3, len(items))

	// map[string]any -> dict
	d, ok := rage.AsDict(rage.FromGo(map[string]any{"key": "val"}))
	assert.True(t, ok)
	assert.Equal(t, 1, len(d))

	// unknown type -> userdata
	type custom struct{ X int }
	assert.True(t, rage.IsUserData(rage.FromGo(custom{X: 1})))

	// Value passthrough
	original := rage.Int(77)
	assert.Equal(t, original, rage.FromGo(original))
}

// =============================================================================
// Round-trip: Go -> Python -> Go
// =============================================================================

func TestRoundTripInt(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.SetGlobal("x", rage.Int(42))
	_, err := state.Run(`y = x + 8`)
	require.NoError(t, err)

	n, ok := rage.AsInt(state.GetGlobal("y"))
	assert.True(t, ok)
	assert.Equal(t, int64(50), n)
}

func TestRoundTripString(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.SetGlobal("greeting", rage.String("Hello"))
	_, err := state.Run(`result = greeting + ", World!"`)
	require.NoError(t, err)

	s, ok := rage.AsString(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.Equal(t, "Hello, World!", s)
}

func TestRoundTripList(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.SetGlobal("nums", rage.List(rage.Int(3), rage.Int(1), rage.Int(2)))
	_, err := state.Run(`sorted_nums = sorted(nums)`)
	require.NoError(t, err)

	items, ok := rage.AsList(state.GetGlobal("sorted_nums"))
	assert.True(t, ok)
	require.Equal(t, 3, len(items))
	n0, _ := rage.AsInt(items[0])
	n1, _ := rage.AsInt(items[1])
	n2, _ := rage.AsInt(items[2])
	assert.Equal(t, int64(1), n0)
	assert.Equal(t, int64(2), n1)
	assert.Equal(t, int64(3), n2)
}

func TestRoundTripDict(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.SetGlobal("config", rage.Dict(
		"host", rage.String("localhost"),
		"port", rage.Int(8080),
	))

	_, err := state.Run(`
url = config["host"] + ":" + str(config["port"])
`)
	require.NoError(t, err)

	s, ok := rage.AsString(state.GetGlobal("url"))
	assert.True(t, ok)
	assert.Equal(t, "localhost:8080", s)
}

func TestRoundTripBool(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.SetGlobal("flag", rage.Bool(false))
	_, err := state.Run(`result = not flag`)
	require.NoError(t, err)

	b, ok := rage.AsBool(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.True(t, b)
}

func TestRoundTripNone(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	state.SetGlobal("val", rage.None)
	_, err := state.Run(`result = val is None`)
	require.NoError(t, err)

	b, ok := rage.AsBool(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.True(t, b)
}

// =============================================================================
// RegisterPythonModule
// =============================================================================

func TestRegisterPythonModule(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	err := state.RegisterPythonModule("mymath", `
def add(a, b):
    return a + b

PI = 3
`)
	require.NoError(t, err)

	_, err = state.Run(`
from mymath import add, PI
result = add(10, PI)
`)
	require.NoError(t, err)

	n, ok := rage.AsInt(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.Equal(t, int64(13), n)
}

func TestRegisterPythonModuleImportStar(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	err := state.RegisterPythonModule("greetings", `
HELLO = "hello"
WORLD = "world"
`)
	require.NoError(t, err)

	_, err = state.Run(`
import greetings
result = greetings.HELLO + " " + greetings.WORLD
`)
	require.NoError(t, err)

	s, ok := rage.AsString(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.Equal(t, "hello world", s)
}

func TestRegisterPythonModuleDotted(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	err := state.RegisterPythonModule("mypkg.utils", `
def greet(name):
    return "Hi, " + name
`)
	require.NoError(t, err)

	_, err = state.Run(`
from mypkg.utils import greet
result = greet("Alice")
`)
	require.NoError(t, err)

	s, ok := rage.AsString(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.Equal(t, "Hi, Alice", s)
}

func TestRegisterPythonModuleCompileError(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	err := state.RegisterPythonModule("badmod", `def`)
	require.Error(t, err)
}

// =============================================================================
// GetModuleAttr
// =============================================================================

func TestGetModuleAttr(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	err := state.RegisterPythonModule("config", `
VERSION = "1.0.0"
MAX_RETRIES = 3
`)
	require.NoError(t, err)

	// Import first so the module is loaded.
	_, err = state.Run(`import config`)
	require.NoError(t, err)

	v := state.GetModuleAttr("config", "VERSION")
	s, ok := rage.AsString(v)
	assert.True(t, ok)
	assert.Equal(t, "1.0.0", s)

	v = state.GetModuleAttr("config", "MAX_RETRIES")
	n, ok := rage.AsInt(v)
	assert.True(t, ok)
	assert.Equal(t, int64(3), n)

	// Non-existent attribute.
	assert.Nil(t, state.GetModuleAttr("config", "NOPE"))

	// Non-existent module.
	assert.Nil(t, state.GetModuleAttr("nope", "x"))
}

// =============================================================================
// Multiple executions share state
// =============================================================================

func TestMultipleRunsShareState(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	_, err := state.Run(`x = 10`)
	require.NoError(t, err)

	_, err = state.Run(`y = x + 20`)
	require.NoError(t, err)

	_, err = state.Run(`z = x + y`)
	require.NoError(t, err)

	n, ok := rage.AsInt(state.GetGlobal("z"))
	assert.True(t, ok)
	assert.Equal(t, int64(40), n)
}

func TestFunctionDefinedInOneRunCalledInAnother(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	_, err := state.Run(`
def factorial(n):
    if n <= 1:
        return 1
    return n * factorial(n - 1)
`)
	require.NoError(t, err)

	_, err = state.Run(`result = factorial(10)`)
	require.NoError(t, err)

	n, ok := rage.AsInt(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.Equal(t, int64(3628800), n)
}

func TestClassDefinedInOneRunUsedInAnother(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	_, err := state.Run(`
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y
    def magnitude(self):
        return (self.x ** 2 + self.y ** 2) ** 0.5
`)
	require.NoError(t, err)

	_, err = state.Run(`
p = Point(3, 4)
result = p.magnitude()
`)
	require.NoError(t, err)

	f, ok := rage.AsFloat(state.GetGlobal("result"))
	assert.True(t, ok)
	assert.InDelta(t, 5.0, f, 0.0001)
}

// =============================================================================
// Error handling
// =============================================================================

func TestCompileErrorsInterface(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	_, err := state.Run(`if`)
	require.Error(t, err)

	var compErr *rage.CompileErrors
	require.ErrorAs(t, err, &compErr)
	assert.Greater(t, len(compErr.Errors), 0)

	// Error() message should be non-empty.
	assert.NotEmpty(t, compErr.Error())

	// Unwrap should return first error.
	assert.NotNil(t, compErr.Unwrap())
}

func TestRuntimeErrorTypes(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		errStr string
	}{
		{"ZeroDivisionError", `x = 1 / 0`, "ZeroDivisionError"},
		{"NameError", `x = undefined_var`, "not defined"},
		{"TypeError", `x = "a" + 1`, "unsupported operand"},
		{"IndexError", `x = [1, 2][5]`, "IndexError"},
		{"KeyError", `x = {}["missing"]`, "KeyError"},
		{"ValueError", `x = int("abc")`, "ValueError"},
		{"AttributeError", `x = (1).nonexistent`, "has no attribute"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := rage.NewState()
			defer state.Close()

			_, err := state.Run(tt.code)
			require.Error(t, err)
			assert.True(t, strings.Contains(err.Error(), tt.errStr),
				"expected error containing %q, got: %s", tt.errStr, err.Error())
		})
	}
}

func TestPythonExceptionCaught(t *testing.T) {
	state := rage.NewState()
	defer state.Close()

	_, err := state.Run(`
try:
    x = 1 / 0
except ZeroDivisionError:
    x = -1
`)
	require.NoError(t, err)

	n, ok := rage.AsInt(state.GetGlobal("x"))
	assert.True(t, ok)
	assert.Equal(t, int64(-1), n)
}

// =============================================================================
// Concurrency: separate states are independent
// =============================================================================

func TestConcurrentStates(t *testing.T) {
	const goroutines = 10

	// Create states sequentially (module init uses global state).
	states := make([]*rage.State, goroutines)
	for i := 0; i < goroutines; i++ {
		states[i] = rage.NewState()
		states[i].SetGlobal("n", rage.Int(int64(i)))
	}

	// Run code concurrently on separate states.
	results := make(chan int64, goroutines)
	for i := 0; i < goroutines; i++ {
		go func(s *rage.State) {
			defer s.Close()
			_, err := s.Run(`result = n * n`)
			if err != nil {
				results <- -1
				return
			}
			v, ok := rage.AsInt(s.GetGlobal("result"))
			if !ok {
				results <- -1
				return
			}
			results <- v
		}(states[i])
	}

	seen := make(map[int64]bool)
	for i := 0; i < goroutines; i++ {
		r := <-results
		assert.NotEqual(t, int64(-1), r)
		seen[r] = true
	}
	// Each goroutine computed a unique n*n.
	assert.Equal(t, goroutines, len(seen))
}
