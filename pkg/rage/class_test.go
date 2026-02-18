package rage

import (
	"fmt"
	"testing"
)

// helper to evaluate an expression via a temporary variable
func eval(t *testing.T, state *State, expr string) Value {
	t.Helper()
	_, err := state.Run(fmt.Sprintf("__test_result__ = %s", expr))
	if err != nil {
		t.Fatalf("error evaluating %q: %v", expr, err)
	}
	return state.GetGlobal("__test_result__")
}

func TestClassBuilder_BasicClass(t *testing.T) {
	state := NewState()
	defer state.Close()

	person := NewClass("Person").
		Init(func(s *State, self Object, args ...Value) {
			self.Set("name", args[0])
			self.Set("age", args[1])
		}).
		Method("greet", func(s *State, self Object, args ...Value) Value {
			name, _ := AsString(self.Get("name"))
			return String("Hello, I'm " + name)
		}).
		Build(state)

	state.SetGlobal("Person", person)

	_, err := state.Run(`p = Person("Alice", 30)`)
	if err != nil {
		t.Fatalf("unexpected error creating instance: %v", err)
	}

	result := eval(t, state, `p.greet()`)
	if str, ok := AsString(result); !ok || str != "Hello, I'm Alice" {
		t.Errorf("expected 'Hello, I'm Alice', got %v", result)
	}
}

func TestClassBuilder_Str(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Point").
		Init(func(s *State, self Object, args ...Value) {
			self.Set("x", args[0])
			self.Set("y", args[1])
		}).
		Str(func(s *State, self Object) string {
			x, _ := AsInt(self.Get("x"))
			y, _ := AsInt(self.Get("y"))
			return fmt.Sprintf("Point(%d, %d)", x, y)
		}).
		Build(state)

	state.SetGlobal("Point", cls)

	result := eval(t, state, `str(Point(3, 4))`)
	if str, ok := AsString(result); !ok || str != "Point(3, 4)" {
		t.Errorf("expected 'Point(3, 4)', got %v", result)
	}
}

func TestClassBuilder_Repr(t *testing.T) {
	state := NewState()
	defer state.Close()
	state.EnableBuiltin(BuiltinRepr)

	cls := NewClass("Foo").
		Repr(func(s *State, self Object) string {
			return "Foo()"
		}).
		Build(state)

	state.SetGlobal("Foo", cls)

	result := eval(t, state, `repr(Foo())`)
	if str, ok := AsString(result); !ok || str != "Foo()" {
		t.Errorf("expected 'Foo()', got %v", result)
	}
}

func TestClassBuilder_Eq(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Val").
		Init(func(s *State, self Object, args ...Value) {
			self.Set("v", args[0])
		}).
		Eq(func(s *State, self Object, other Value) bool {
			otherObj, ok := other.(Object)
			if !ok {
				return false
			}
			a, _ := AsInt(self.Get("v"))
			b, _ := AsInt(otherObj.Get("v"))
			return a == b
		}).
		Build(state)

	state.SetGlobal("Val", cls)

	result := eval(t, state, `Val(10) == Val(10)`)
	if b, ok := AsBool(result); !ok || !b {
		t.Errorf("expected True, got %v", result)
	}

	result = eval(t, state, `Val(10) == Val(20)`)
	if b, ok := AsBool(result); !ok || b {
		t.Errorf("expected False, got %v", result)
	}
}

func TestClassBuilder_Len(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Bag").
		Init(func(s *State, self Object, args ...Value) {
			self.Set("count", args[0])
		}).
		Len(func(s *State, self Object) int64 {
			n, _ := AsInt(self.Get("count"))
			return n
		}).
		Build(state)

	state.SetGlobal("Bag", cls)

	result := eval(t, state, `len(Bag(5))`)
	if n, ok := AsInt(result); !ok || n != 5 {
		t.Errorf("expected 5, got %v", result)
	}
}

func TestClassBuilder_GetItem(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("MyList").
		Init(func(s *State, self Object, args ...Value) {
			self.Set("items", args[0])
		}).
		GetItem(func(s *State, self Object, key Value) Value {
			items, _ := AsList(self.Get("items"))
			idx, _ := AsInt(key)
			if int(idx) < len(items) {
				return items[idx]
			}
			return None
		}).
		Build(state)

	state.SetGlobal("MyList", cls)

	result := eval(t, state, `MyList([10, 20, 30])[1]`)
	if n, ok := AsInt(result); !ok || n != 20 {
		t.Errorf("expected 20, got %v", result)
	}
}

func TestClassBuilder_Contains(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Set3").
		Contains(func(s *State, self Object, item Value) bool {
			n, _ := AsInt(item)
			return n >= 1 && n <= 3
		}).
		Build(state)

	state.SetGlobal("Set3", cls)

	result := eval(t, state, `2 in Set3()`)
	if b, ok := AsBool(result); !ok || !b {
		t.Errorf("expected True, got %v", result)
	}

	result = eval(t, state, `5 in Set3()`)
	if b, ok := AsBool(result); !ok || b {
		t.Errorf("expected False, got %v", result)
	}
}

func TestClassBuilder_BoolDunder(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Falsy").
		Bool(func(s *State, self Object) bool {
			return false
		}).
		Build(state)

	state.SetGlobal("Falsy", cls)

	result := eval(t, state, `bool(Falsy())`)
	if b, ok := AsBool(result); !ok || b {
		t.Errorf("expected False, got %v", result)
	}
}

func TestClassBuilder_CallDunder(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Adder").
		Init(func(s *State, self Object, args ...Value) {
			self.Set("base", args[0])
		}).
		Call(func(s *State, self Object, args ...Value) Value {
			base, _ := AsInt(self.Get("base"))
			n, _ := AsInt(args[0])
			return Int(base + n)
		}).
		Build(state)

	state.SetGlobal("Adder", cls)

	result := eval(t, state, `Adder(10)(5)`)
	if n, ok := AsInt(result); !ok || n != 15 {
		t.Errorf("expected 15, got %v", result)
	}
}

func TestClassBuilder_Property(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Circle").
		Init(func(s *State, self Object, args ...Value) {
			self.Set("_radius", args[0])
		}).
		Property("radius", func(s *State, self Object) Value {
			return self.Get("_radius")
		}).
		Build(state)

	state.SetGlobal("Circle", cls)

	result := eval(t, state, `Circle(5).radius`)
	if n, ok := AsInt(result); !ok || n != 5 {
		t.Errorf("expected 5, got %v", result)
	}
}

func TestClassBuilder_PropertyWithSetter(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Box").
		Init(func(s *State, self Object, args ...Value) {
			self.Set("_size", args[0])
		}).
		PropertyWithSetter("size",
			func(s *State, self Object) Value {
				return self.Get("_size")
			},
			func(s *State, self Object, val Value) {
				self.Set("_size", val)
			},
		).
		Build(state)

	state.SetGlobal("Box", cls)

	_, err := state.Run(`
b = Box(10)
b.size = 20
_prop_result = b.size
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("_prop_result")
	if n, ok := AsInt(result); !ok || n != 20 {
		t.Errorf("expected 20, got %v", result)
	}
}

func TestClassBuilder_StaticMethod(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("MathHelper").
		StaticMethod("add", func(s *State, args ...Value) Value {
			a, _ := AsInt(args[0])
			b, _ := AsInt(args[1])
			return Int(a + b)
		}).
		Build(state)

	state.SetGlobal("MathHelper", cls)

	result := eval(t, state, `MathHelper.add(3, 4)`)
	if n, ok := AsInt(result); !ok || n != 7 {
		t.Errorf("expected 7, got %v", result)
	}
}

func TestClassBuilder_ClassMethod(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Named").
		ClassMethod("class_name", func(s *State, cls ClassValue, args ...Value) Value {
			return String(cls.Name())
		}).
		Build(state)

	state.SetGlobal("Named", cls)

	result := eval(t, state, `Named.class_name()`)
	if str, ok := AsString(result); !ok || str != "Named" {
		t.Errorf("expected 'Named', got %v", result)
	}
}

func TestClassBuilder_Isinstance(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("MyClass").Build(state)
	state.SetGlobal("MyClass", cls)

	result := eval(t, state, `isinstance(MyClass(), MyClass)`)
	if b, ok := AsBool(result); !ok || !b {
		t.Errorf("expected True, got %v", result)
	}
}

func TestClassBuilder_Inheritance(t *testing.T) {
	state := NewState()
	defer state.Close()

	animal := NewClass("Animal").
		Init(func(s *State, self Object, args ...Value) {
			self.Set("name", args[0])
		}).
		Method("speak", func(s *State, self Object, args ...Value) Value {
			return String("...")
		}).
		Build(state)

	dog := NewClass("Dog").
		Base(animal).
		Method("speak", func(s *State, self Object, args ...Value) Value {
			name, _ := AsString(self.Get("name"))
			return String(name + " says Woof!")
		}).
		Build(state)

	state.SetGlobal("Animal", animal)
	state.SetGlobal("Dog", dog)

	// Dog inherits Animal's __init__
	result := eval(t, state, `Dog("Rex").speak()`)
	if str, ok := AsString(result); !ok || str != "Rex says Woof!" {
		t.Errorf("expected 'Rex says Woof!', got %v", result)
	}

	// isinstance check
	result = eval(t, state, `isinstance(Dog("Rex"), Animal)`)
	if b, ok := AsBool(result); !ok || !b {
		t.Errorf("expected isinstance(Dog, Animal) = True, got %v", result)
	}
}

func TestClassBuilder_DunderArbitrary(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Addable").
		Init(func(s *State, self Object, args ...Value) {
			self.Set("val", args[0])
		}).
		Dunder("__add__", func(s *State, self Object, args ...Value) Value {
			a, _ := AsInt(self.Get("val"))
			if other, ok := args[0].(Object); ok {
				b, _ := AsInt(other.Get("val"))
				return Int(a + b)
			}
			return None
		}).
		Build(state)

	state.SetGlobal("Addable", cls)

	result := eval(t, state, `Addable(3) + Addable(4)`)
	if n, ok := AsInt(result); !ok || n != 7 {
		t.Errorf("expected 7, got %v", result)
	}
}

func TestObject_GetSetHasDelete(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Obj").Build(state)
	obj := cls.NewInstance()

	if obj.Has("x") {
		t.Error("expected Has('x') = false")
	}

	obj.Set("x", Int(42))
	if !obj.Has("x") {
		t.Error("expected Has('x') = true")
	}
	if n, ok := AsInt(obj.Get("x")); !ok || n != 42 {
		t.Errorf("expected 42, got %v", obj.Get("x"))
	}

	obj.Delete("x")
	if obj.Has("x") {
		t.Error("expected Has('x') = false after delete")
	}
}

func TestObject_ClassName(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("TestClass").Build(state)
	obj := cls.NewInstance()

	if obj.ClassName() != "TestClass" {
		t.Errorf("expected 'TestClass', got %q", obj.ClassName())
	}
}

func TestClassValue_Name(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("MyClass").Build(state)
	if cls.Name() != "MyClass" {
		t.Errorf("expected 'MyClass', got %q", cls.Name())
	}
}

func TestClassValue_Type(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("MyClass").Build(state)
	if cls.Type() != "type" {
		t.Errorf("expected 'type', got %q", cls.Type())
	}
}

func TestState_Call(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Doubler").
		Init(func(s *State, self Object, args ...Value) {
			self.Set("v", args[0])
		}).
		Method("double", func(s *State, self Object, args ...Value) Value {
			n, _ := AsInt(self.Get("v"))
			return Int(n * 2)
		}).
		Build(state)

	result, err := state.Call(cls, Int(21))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj, ok := result.(Object)
	if !ok {
		t.Fatalf("expected Object, got %T", result)
	}
	if n, ok := AsInt(obj.Get("v")); !ok || n != 21 {
		t.Errorf("expected v=21, got %v", obj.Get("v"))
	}
}

func TestState_CallFunction(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.Run(`
def add(a, b):
    return a + b
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	addFn := state.GetGlobal("add")
	result, err := state.Call(addFn, Int(3), Int(4))
	if err != nil {
		t.Fatalf("unexpected error calling add: %v", err)
	}
	if n, ok := AsInt(result); !ok || n != 7 {
		t.Errorf("expected 7, got %v", result)
	}
}

func TestClassBuilder_SetItem(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Store").
		Init(func(s *State, self Object, args ...Value) {
			// nothing
		}).
		SetItem(func(s *State, self Object, key, val Value) {
			k, _ := AsString(key)
			self.Set("_item_"+k, val)
		}).
		GetItem(func(s *State, self Object, key Value) Value {
			k, _ := AsString(key)
			return self.Get("_item_" + k)
		}).
		Build(state)

	state.SetGlobal("Store", cls)

	_, err := state.Run(`
s = Store()
s["x"] = 42
_setitem_result = s["x"]
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("_setitem_result")
	if n, ok := AsInt(result); !ok || n != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestFromRuntime_ClassAndInstance(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.Run(`
class Foo:
    pass
obj = Foo()
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fooClass := state.GetGlobal("Foo")
	if _, ok := fooClass.(ClassValue); !ok {
		t.Errorf("expected ClassValue for Foo, got %T", fooClass)
	}

	fooObj := state.GetGlobal("obj")
	if _, ok := fooObj.(Object); !ok {
		t.Errorf("expected Object for obj, got %T", fooObj)
	}
}

func TestClassBuilder_MethodAccessesState(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Runner").
		Method("run_code", func(s *State, self Object, args ...Value) Value {
			code, _ := AsString(args[0])
			result, err := s.Run(code)
			if err != nil {
				return None
			}
			if result == nil {
				return None
			}
			return result
		}).
		Build(state)

	state.SetGlobal("Runner", cls)

	// Use Run to assign to a variable, since Run on statements returns nil
	_, err := state.Run(`_runner_result = Runner().run_code("1 + 2")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// run_code runs "1 + 2" as a statement which returns None because there's no assignment
	// So let's test with something that uses State properly
	_, err = state.Run(`_runner_result = Runner().run_code("__x = 1 + 2")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("__x")
	if n, ok := AsInt(result); !ok || n != 3 {
		t.Errorf("expected 3, got %v", result)
	}
}
