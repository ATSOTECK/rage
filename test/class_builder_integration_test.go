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
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("name", args[0])
			self.Set("age", args[1])
		}).
		Method("greet", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			name, _ := rage.AsString(self.Get("name"))
			return rage.String("Hello, I'm " + name)
		}).
		Str(func(s *rage.State, self rage.Object) string {
			name, _ := rage.AsString(self.Get("name"))
			age, _ := rage.AsInt(self.Get("age"))
			return fmt.Sprintf("Person(%s, %d)", name, age)
		}).
		Build(state)
	state.SetGlobal("Person", person)

	// Animal(name) — base class
	animal := rage.NewClass("Animal").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("name", args[0])
		}).
		Method("speak", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			return rage.String("...")
		}).
		Build(state)
	state.SetGlobal("Animal", animal)

	// Dog(name) — inherits Animal
	dog := rage.NewClass("Dog").
		Base(animal).
		Method("speak", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			name, _ := rage.AsString(self.Get("name"))
			return rage.String(name + " says Woof!")
		}).
		Build(state)
	state.SetGlobal("Dog", dog)

	// Cat(name) — inherits Animal
	cat := rage.NewClass("Cat").
		Base(animal).
		Method("speak", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			name, _ := rage.AsString(self.Get("name"))
			return rage.String(name + " says Meow!")
		}).
		Build(state)
	state.SetGlobal("Cat", cat)

	// Container(items) — __len__, __getitem__, __contains__, __eq__, __bool__, __str__
	container := rage.NewClass("Container").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("items", args[0])
		}).
		Len(func(s *rage.State, self rage.Object) int64 {
			items, _ := rage.AsList(self.Get("items"))
			return int64(len(items))
		}).
		GetItem(func(s *rage.State, self rage.Object, key rage.Value) rage.Value {
			items, _ := rage.AsList(self.Get("items"))
			idx, _ := rage.AsInt(key)
			if int(idx) < len(items) {
				return items[idx]
			}
			return rage.None
		}).
		Contains(func(s *rage.State, self rage.Object, item rage.Value) bool {
			items, _ := rage.AsList(self.Get("items"))
			itemInt, ok := rage.AsInt(item)
			if !ok {
				return false
			}
			for _, v := range items {
				if n, ok := rage.AsInt(v); ok && n == itemInt {
					return true
				}
			}
			return false
		}).
		Eq(func(s *rage.State, self rage.Object, other rage.Value) bool {
			otherObj, ok := other.(rage.Object)
			if !ok {
				return false
			}
			selfItems, _ := rage.AsList(self.Get("items"))
			otherItems, _ := rage.AsList(otherObj.Get("items"))
			if len(selfItems) != len(otherItems) {
				return false
			}
			for i := range selfItems {
				a, _ := rage.AsInt(selfItems[i])
				b, _ := rage.AsInt(otherItems[i])
				if a != b {
					return false
				}
			}
			return true
		}).
		Bool(func(s *rage.State, self rage.Object) bool {
			items, _ := rage.AsList(self.Get("items"))
			return len(items) > 0
		}).
		Str(func(s *rage.State, self rage.Object) string {
			items, _ := rage.AsList(self.Get("items"))
			return fmt.Sprintf("Container(%d items)", len(items))
		}).
		Build(state)
	state.SetGlobal("Container", container)

	// Multiplier(factor) — __call__
	multiplier := rage.NewClass("Multiplier").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("factor", args[0])
		}).
		Call(func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			factor, _ := rage.AsInt(self.Get("factor"))
			n, _ := rage.AsInt(args[0])
			return rage.Int(factor * n)
		}).
		Build(state)
	state.SetGlobal("Multiplier", multiplier)

	// Rect(w, h) — properties
	rect := rage.NewClass("Rect").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("_w", args[0])
			self.Set("_h", args[1])
		}).
		Property("area", func(s *rage.State, self rage.Object) rage.Value {
			w, _ := rage.AsInt(self.Get("_w"))
			h, _ := rage.AsInt(self.Get("_h"))
			return rage.Int(w * h)
		}).
		PropertyWithSetter("width",
			func(s *rage.State, self rage.Object) rage.Value {
				return self.Get("_w")
			},
			func(s *rage.State, self rage.Object, val rage.Value) {
				self.Set("_w", val)
			},
		).
		Build(state)
	state.SetGlobal("Rect", rect)

	// Counter(n) — static method, class method, increment
	counter := rage.NewClass("Counter").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			if len(args) > 0 {
				self.Set("count", args[0])
			} else {
				self.Set("count", rage.Int(0))
			}
		}).
		Method("increment", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			n, _ := rage.AsInt(self.Get("count"))
			self.Set("count", rage.Int(n+1))
			return rage.None
		}).
		StaticMethod("from_string", func(s *rage.State, args ...rage.Value) rage.Value {
			str, _ := rage.AsString(args[0])
			return rage.Int(int64(len(str)))
		}).
		ClassMethod("class_name", func(s *rage.State, cls rage.ClassValue, args ...rage.Value) rage.Value {
			return rage.String(cls.Name())
		}).
		Build(state)
	state.SetGlobal("Counter", counter)

	// Vec2(x, y) — __add__, __str__, __repr__
	vec2 := rage.NewClass("Vec2").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("x", args[0])
			self.Set("y", args[1])
		}).
		Dunder("__add__", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			other, ok := args[0].(rage.Object)
			if !ok {
				return rage.None
			}
			x1, _ := rage.AsInt(self.Get("x"))
			y1, _ := rage.AsInt(self.Get("y"))
			x2, _ := rage.AsInt(other.Get("x"))
			y2, _ := rage.AsInt(other.Get("y"))
			return rage.List(rage.Int(x1+x2), rage.Int(y1+y2))
		}).
		Str(func(s *rage.State, self rage.Object) string {
			x, _ := rage.AsInt(self.Get("x"))
			y, _ := rage.AsInt(self.Get("y"))
			return fmt.Sprintf("Vec2(%d, %d)", x, y)
		}).
		Repr(func(s *rage.State, self rage.Object) string {
			x, _ := rage.AsInt(self.Get("x"))
			y, _ := rage.AsInt(self.Get("y"))
			return fmt.Sprintf("Vec2(%d, %d)", x, y)
		}).
		Build(state)
	state.SetGlobal("Vec2", vec2)

	// GoBase(value) — simple base class for Python to inherit from
	goBase := rage.NewClass("GoBase").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			self.Set("value", args[0])
		}).
		Method("get_value", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			return self.Get("value")
		}).
		Build(state)
	state.SetGlobal("GoBase", goBase)

	// Store() — __setitem__, __getitem__
	store := rage.NewClass("Store").
		Init(func(s *rage.State, self rage.Object, args ...rage.Value) {
			// no-op
		}).
		SetItem(func(s *rage.State, self rage.Object, key, val rage.Value) {
			k, _ := rage.AsString(key)
			self.Set("_item_"+k, val)
		}).
		GetItem(func(s *rage.State, self rage.Object, key rage.Value) rage.Value {
			k, _ := rage.AsString(key)
			return self.Get("_item_" + k)
		}).
		Build(state)
	state.SetGlobal("Store", store)

	// Config instance — created from Go via NewInstance() (no __init__)
	config := rage.NewClass("Config").
		Method("get", func(s *rage.State, self rage.Object, args ...rage.Value) rage.Value {
			key, _ := rage.AsString(args[0])
			return self.Get(key)
		}).
		Build(state)
	state.SetGlobal("Config", config)

	configInst := config.NewInstance()
	configInst.Set("host", rage.String("localhost"))
	configInst.Set("port", rage.Int(8080))
	state.SetGlobal("config", configInst)
}

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
