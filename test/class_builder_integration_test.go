package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ATSOTECK/rage/pkg/rage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClassBuilderPythonIntegration registers Go-defined classes via the ClassBuilder API
// and then runs a Python test script that exercises them.
func TestClassBuilderPythonIntegration(t *testing.T) {
	state := rage.NewState()
	defer state.Close()
	state.EnableBuiltin(rage.BuiltinRepr)

	// --- Register test framework ---
	frameworkPath := findTestFramework(t)
	frameworkSrc, err := os.ReadFile(frameworkPath)
	require.NoError(t, err)
	err = state.RegisterPythonModule("test_framework", string(frameworkSrc))
	require.NoError(t, err)

	// --- Build and register Go-defined classes ---
	registerGoClasses(t, state)

	// --- Run the Python test script ---
	scriptPath := findScript(t, "107_go_defined_classes.py")
	scriptSrc, err := os.ReadFile(scriptPath)
	require.NoError(t, err)

	_, err = state.RunWithTimeout(string(scriptSrc), 30*time.Second)
	require.NoError(t, err, "Python script execution failed")

	// --- Extract and verify test results ---
	passed := 0
	failed := 0
	failures := ""

	if v := state.GetModuleAttr("test_framework", "__test_passed__"); v != nil {
		if i, ok := rage.AsInt(v); ok {
			passed = int(i)
		}
	}
	if v := state.GetModuleAttr("test_framework", "__test_failed__"); v != nil {
		if i, ok := rage.AsInt(v); ok {
			failed = int(i)
		}
	}
	if v := state.GetModuleAttr("test_framework", "__test_failures__"); v != nil {
		if s, ok := rage.AsString(v); ok {
			failures = s
		}
	}

	t.Logf("Python tests: %d passed, %d failed", passed, failed)
	if failures != "" {
		for _, line := range strings.Split(strings.TrimSpace(failures), "\n") {
			if line != "" {
				t.Errorf("  FAIL: %s", line)
			}
		}
	}
	assert.Equal(t, 0, failed, "Some Python tests failed")
	assert.Greater(t, passed, 0, "No Python tests ran")
}

// registerGoClasses builds all Go-defined classes needed by 107_go_defined_classes.py.
func registerGoClasses(t *testing.T, state *rage.State) {
	t.Helper()

	// Person(name, age) — __init__, greet(), __str__
	person := rage.NewClass("Person").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			self.Set("name", args[0])
			self.Set("age", args[1])
			return nil
		}).
		Method("greet", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			name, _ := rage.AsString(self.Get("name"))
			return rage.String("Hello, I'm " + name), nil
		}).
		Str(func(s *rage.State, self rage.Object) (string, error) {
			name, _ := rage.AsString(self.Get("name"))
			age, _ := rage.AsInt(self.Get("age"))
			return fmt.Sprintf("Person(%s, %d)", name, age), nil
		}).
		Build(state)
	state.SetGlobal("Person", person)

	// Animal(name) — base class
	animal := rage.NewClass("Animal").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			self.Set("name", args[0])
			return nil
		}).
		Method("speak", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			return rage.String("..."), nil
		}).
		Build(state)
	state.SetGlobal("Animal", animal)

	// Dog(name) — inherits Animal
	dog := rage.NewClass("Dog").
		Base(animal).
		Method("speak", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			name, _ := rage.AsString(self.Get("name"))
			return rage.String(name + " says Woof!"), nil
		}).
		Build(state)
	state.SetGlobal("Dog", dog)

	// Cat(name) — inherits Animal
	cat := rage.NewClass("Cat").
		Base(animal).
		Method("speak", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			name, _ := rage.AsString(self.Get("name"))
			return rage.String(name + " says Meow!"), nil
		}).
		Build(state)
	state.SetGlobal("Cat", cat)

	// Container(items) — __len__, __getitem__, __contains__, __eq__, __bool__, __str__
	container := rage.NewClass("Container").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			self.Set("items", args[0])
			return nil
		}).
		Len(func(s *rage.State, self rage.Object) (int64, error) {
			items, _ := rage.AsList(self.Get("items"))
			return int64(len(items)), nil
		}).
		GetItem(func(s *rage.State, self rage.Object, key rage.Value) (rage.Value, error) {
			items, _ := rage.AsList(self.Get("items"))
			idx, _ := rage.AsInt(key)
			if int(idx) < len(items) {
				return items[idx], nil
			}
			return rage.None, nil
		}).
		Contains(func(s *rage.State, self rage.Object, item rage.Value) (bool, error) {
			items, _ := rage.AsList(self.Get("items"))
			itemInt, ok := rage.AsInt(item)
			if !ok {
				return false, nil
			}
			for _, v := range items {
				if n, ok := rage.AsInt(v); ok && n == itemInt {
					return true, nil
				}
			}
			return false, nil
		}).
		Eq(func(s *rage.State, self rage.Object, other rage.Value) (bool, error) {
			otherObj, ok := other.(rage.Object)
			if !ok {
				return false, nil
			}
			selfItems, _ := rage.AsList(self.Get("items"))
			otherItems, _ := rage.AsList(otherObj.Get("items"))
			if len(selfItems) != len(otherItems) {
				return false, nil
			}
			for i := range selfItems {
				a, _ := rage.AsInt(selfItems[i])
				b, _ := rage.AsInt(otherItems[i])
				if a != b {
					return false, nil
				}
			}
			return true, nil
		}).
		Bool(func(s *rage.State, self rage.Object) (bool, error) {
			items, _ := rage.AsList(self.Get("items"))
			return len(items) > 0, nil
		}).
		Str(func(s *rage.State, self rage.Object) (string, error) {
			items, _ := rage.AsList(self.Get("items"))
			return fmt.Sprintf("Container(%d items)", len(items)), nil
		}).
		Build(state)
	state.SetGlobal("Container", container)

	// Multiplier(factor) — __call__
	multiplier := rage.NewClass("Multiplier").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			self.Set("factor", args[0])
			return nil
		}).
		Call(func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			factor, _ := rage.AsInt(self.Get("factor"))
			n, _ := rage.AsInt(args[0])
			return rage.Int(factor * n), nil
		}).
		Build(state)
	state.SetGlobal("Multiplier", multiplier)

	// Rect(w, h) — properties
	rect := rage.NewClass("Rect").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			self.Set("_w", args[0])
			self.Set("_h", args[1])
			return nil
		}).
		Property("area", func(s *rage.State, self rage.Object) (rage.Value, error) {
			w, _ := rage.AsInt(self.Get("_w"))
			h, _ := rage.AsInt(self.Get("_h"))
			return rage.Int(w * h), nil
		}).
		PropertyWithSetter("width",
			func(s *rage.State, self rage.Object) (rage.Value, error) {
				return self.Get("_w"), nil
			},
			func(s *rage.State, self rage.Object, val rage.Value) error {
				self.Set("_w", val)
				return nil
			},
		).
		Build(state)
	state.SetGlobal("Rect", rect)

	// Counter(n) — static method, class method, increment
	counter := rage.NewClass("Counter").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			if len(args) > 0 {
				self.Set("count", args[0])
			} else {
				self.Set("count", rage.Int(0))
			}
			return nil
		}).
		Method("increment", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			n, _ := rage.AsInt(self.Get("count"))
			self.Set("count", rage.Int(n+1))
			return rage.None, nil
		}).
		StaticMethod("from_string", func(s *rage.State, args ...rage.Value) (rage.Value, error) {
			str, _ := rage.AsString(args[0])
			return rage.Int(int64(len(str))), nil
		}).
		ClassMethod("class_name", func(s *rage.State, cls rage.ClassValue, args ...rage.Value) (rage.Value, error) {
			return rage.String(cls.Name()), nil
		}).
		Build(state)
	state.SetGlobal("Counter", counter)

	// Vec2(x, y) — __add__, __str__, __repr__
	vec2 := rage.NewClass("Vec2").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			self.Set("x", args[0])
			self.Set("y", args[1])
			return nil
		}).
		Dunder("__add__", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			other, ok := args[0].(rage.Object)
			if !ok {
				return rage.None, nil
			}
			x1, _ := rage.AsInt(self.Get("x"))
			y1, _ := rage.AsInt(self.Get("y"))
			x2, _ := rage.AsInt(other.Get("x"))
			y2, _ := rage.AsInt(other.Get("y"))
			return rage.List(rage.Int(x1+x2), rage.Int(y1+y2)), nil
		}).
		Str(func(s *rage.State, self rage.Object) (string, error) {
			x, _ := rage.AsInt(self.Get("x"))
			y, _ := rage.AsInt(self.Get("y"))
			return fmt.Sprintf("Vec2(%d, %d)", x, y), nil
		}).
		Repr(func(s *rage.State, self rage.Object) (string, error) {
			x, _ := rage.AsInt(self.Get("x"))
			y, _ := rage.AsInt(self.Get("y"))
			return fmt.Sprintf("Vec2(%d, %d)", x, y), nil
		}).
		Build(state)
	state.SetGlobal("Vec2", vec2)

	// GoBase(value) — simple base class for Python to inherit from
	goBase := rage.NewClass("GoBase").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			self.Set("value", args[0])
			return nil
		}).
		Method("get_value", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			return self.Get("value"), nil
		}).
		Build(state)
	state.SetGlobal("GoBase", goBase)

	// Store() — __setitem__, __getitem__
	store := rage.NewClass("Store").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			return nil
		}).
		SetItem(func(s *rage.State, self rage.Object, key, val rage.Value) error {
			k, _ := rage.AsString(key)
			self.Set("_item_"+k, val)
			return nil
		}).
		GetItem(func(s *rage.State, self rage.Object, key rage.Value) (rage.Value, error) {
			k, _ := rage.AsString(key)
			return self.Get("_item_" + k), nil
		}).
		Build(state)
	state.SetGlobal("Store", store)

	// Config instance — created from Go via NewInstance() (no __init__)
	config := rage.NewClass("Config").
		Method("get", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			key, _ := rage.AsString(args[0])
			return self.Get(key), nil
		}).
		Build(state)
	state.SetGlobal("Config", config)

	configInst := config.NewInstance()
	configInst.Set("host", rage.String("localhost"))
	configInst.Set("port", rage.Int(8080))
	state.SetGlobal("config", configInst)

	// Range(start, end) — __iter__ / __next__ (iterator protocol)
	goRange := rage.NewClass("GoRange").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			self.Set("start", args[0])
			self.Set("end", args[1])
			return nil
		}).
		Iter(func(s *rage.State, self rage.Object) (rage.Value, error) {
			// Return a new iterator instance
			start, _ := rage.AsInt(self.Get("start"))
			end, _ := rage.AsInt(self.Get("end"))
			iter := goRangeIter.NewInstance()
			iter.Set("current", rage.Int(start))
			iter.Set("end", rage.Int(end))
			return iter, nil
		}).
		Build(state)
	state.SetGlobal("GoRange", goRange)

	// GoRangeIter — the iterator companion for GoRange
	goRangeIter = rage.NewClass("GoRangeIter").
		Iter(func(s *rage.State, self rage.Object) (rage.Value, error) {
			return self, nil // iterators return themselves
		}).
		Next(func(s *rage.State, self rage.Object) (rage.Value, error) {
			cur, _ := rage.AsInt(self.Get("current"))
			end, _ := rage.AsInt(self.Get("end"))
			if cur >= end {
				return nil, rage.ErrStopIteration
			}
			self.Set("current", rage.Int(cur+1))
			return rage.Int(cur), nil
		}).
		Build(state)
	state.SetGlobal("GoRangeIter", goRangeIter)

	// Temperature(value) — comparison operators
	temp := rage.NewClass("Temperature").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			self.Set("value", args[0])
			return nil
		}).
		Eq(func(s *rage.State, self rage.Object, other rage.Value) (bool, error) {
			otherObj, ok := other.(rage.Object)
			if !ok {
				return false, nil
			}
			a, _ := rage.AsInt(self.Get("value"))
			b, _ := rage.AsInt(otherObj.Get("value"))
			return a == b, nil
		}).
		Lt(func(s *rage.State, self rage.Object, other rage.Value) (bool, error) {
			otherObj, ok := other.(rage.Object)
			if !ok {
				return false, rage.TypeError("unsupported comparison")
			}
			a, _ := rage.AsInt(self.Get("value"))
			b, _ := rage.AsInt(otherObj.Get("value"))
			return a < b, nil
		}).
		Le(func(s *rage.State, self rage.Object, other rage.Value) (bool, error) {
			otherObj, ok := other.(rage.Object)
			if !ok {
				return false, rage.TypeError("unsupported comparison")
			}
			a, _ := rage.AsInt(self.Get("value"))
			b, _ := rage.AsInt(otherObj.Get("value"))
			return a <= b, nil
		}).
		Gt(func(s *rage.State, self rage.Object, other rage.Value) (bool, error) {
			otherObj, ok := other.(rage.Object)
			if !ok {
				return false, rage.TypeError("unsupported comparison")
			}
			a, _ := rage.AsInt(self.Get("value"))
			b, _ := rage.AsInt(otherObj.Get("value"))
			return a > b, nil
		}).
		Ge(func(s *rage.State, self rage.Object, other rage.Value) (bool, error) {
			otherObj, ok := other.(rage.Object)
			if !ok {
				return false, rage.TypeError("unsupported comparison")
			}
			a, _ := rage.AsInt(self.Get("value"))
			b, _ := rage.AsInt(otherObj.Get("value"))
			return a >= b, nil
		}).
		Hash(func(s *rage.State, self rage.Object) (int64, error) {
			v, _ := rage.AsInt(self.Get("value"))
			return v, nil
		}).
		Str(func(s *rage.State, self rage.Object) (string, error) {
			v, _ := rage.AsInt(self.Get("value"))
			return fmt.Sprintf("Temperature(%d)", v), nil
		}).
		Build(state)
	state.SetGlobal("Temperature", temp)

	// Ledger() — __delitem__
	ledger := rage.NewClass("Ledger").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			return nil
		}).
		SetItem(func(s *rage.State, self rage.Object, key, val rage.Value) error {
			k, _ := rage.AsString(key)
			self.Set("_entry_"+k, val)
			return nil
		}).
		GetItem(func(s *rage.State, self rage.Object, key rage.Value) (rage.Value, error) {
			k, _ := rage.AsString(key)
			v := self.Get("_entry_" + k)
			return v, nil
		}).
		DelItem(func(s *rage.State, self rage.Object, key rage.Value) error {
			k, _ := rage.AsString(key)
			if !self.Has("_entry_" + k) {
				return rage.KeyError(k)
			}
			self.Delete("_entry_" + k)
			return nil
		}).
		Build(state)
	state.SetGlobal("Ledger", ledger)

	// GoContextManager(name) — __enter__ / __exit__
	ctxMgr := rage.NewClass("GoContextManager").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			self.Set("name", args[0])
			self.Set("entered", rage.False)
			self.Set("exited", rage.False)
			self.Set("had_error", rage.False)
			return nil
		}).
		Enter(func(s *rage.State, self rage.Object) (rage.Value, error) {
			self.Set("entered", rage.True)
			return self, nil
		}).
		Exit(func(s *rage.State, self rage.Object, excType, excVal, excTb rage.Value) (bool, error) {
			self.Set("exited", rage.True)
			if !rage.IsNone(excType) {
				self.Set("had_error", rage.True)
			}
			return false, nil // don't suppress exceptions
		}).
		Method("status", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			boolStr := func(v rage.Value) string {
				b, _ := rage.AsBool(v)
				if b {
					return "True"
				}
				return "False"
			}
			return rage.String(fmt.Sprintf("entered=%s exited=%s error=%s",
				boolStr(self.Get("entered")),
				boolStr(self.Get("exited")),
				boolStr(self.Get("had_error")))), nil
		}).
		Build(state)
	state.SetGlobal("GoContextManager", ctxMgr)

	// ErrorRaiser() — methods that return Go errors becoming Python exceptions
	errRaiser := rage.NewClass("ErrorRaiser").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) error {
			return nil
		}).
		Method("raise_value_error", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			return nil, rage.ValueError("bad value from Go")
		}).
		Method("raise_type_error", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			return nil, rage.TypeError("wrong type from Go")
		}).
		Method("raise_key_error", func(s *rage.State, self rage.Object, args ...rage.Value) (rage.Value, error) {
			return nil, rage.KeyError("missing_key")
		}).
		Build(state)
	state.SetGlobal("ErrorRaiser", errRaiser)
}

// goRangeIter is set during registerGoClasses so GoRange's Iter can reference it.
var goRangeIter rage.ClassValue

// findTestFramework locates the test_framework.py file.
func findTestFramework(t *testing.T) string {
	t.Helper()
	candidates := []string{
		"integration/scripts/common/test_framework.py",
		"test/integration/scripts/common/test_framework.py",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	t.Fatal("could not find test_framework.py")
	return ""
}

// findScript locates a Python test script by name.
func findScript(t *testing.T, name string) string {
	t.Helper()
	candidates := []string{
		filepath.Join("integration", "scripts", "lang", name),
		filepath.Join("test", "integration", "scripts", "lang", name),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	t.Fatalf("could not find script %s", name)
	return ""
}
