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
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("name", args[0])
			self.Set("age", args[1])
			return nil
		}).
		Method("greet", func(s *State, self Object, args ...Value) (Value, error) {
			name, _ := AsString(self.Get("name"))
			return String("Hello, I'm " + name), nil
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
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("x", args[0])
			self.Set("y", args[1])
			return nil
		}).
		Str(func(s *State, self Object) (string, error) {
			x, _ := AsInt(self.Get("x"))
			y, _ := AsInt(self.Get("y"))
			return fmt.Sprintf("Point(%d, %d)", x, y), nil
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
		Repr(func(s *State, self Object) (string, error) {
			return "Foo()", nil
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
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("v", args[0])
			return nil
		}).
		Eq(func(s *State, self Object, other Value) (bool, error) {
			otherObj, ok := AsObject(other)
			if !ok {
				return false, nil
			}
			a, _ := AsInt(self.Get("v"))
			b, _ := AsInt(otherObj.Get("v"))
			return a == b, nil
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
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("count", args[0])
			return nil
		}).
		Len(func(s *State, self Object) (int64, error) {
			n, _ := AsInt(self.Get("count"))
			return n, nil
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
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("items", args[0])
			return nil
		}).
		GetItem(func(s *State, self Object, key Value) (Value, error) {
			items, _ := AsList(self.Get("items"))
			idx, _ := AsInt(key)
			if int(idx) < len(items) {
				return items[idx], nil
			}
			return nil, IndexError("index out of range")
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
		Contains(func(s *State, self Object, item Value) (bool, error) {
			n, _ := AsInt(item)
			return n >= 1 && n <= 3, nil
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
		Bool(func(s *State, self Object) (bool, error) {
			return false, nil
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
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("base", args[0])
			return nil
		}).
		Call(func(s *State, self Object, args ...Value) (Value, error) {
			base, _ := AsInt(self.Get("base"))
			n, _ := AsInt(args[0])
			return Int(base + n), nil
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
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("_radius", args[0])
			return nil
		}).
		Property("radius", func(s *State, self Object) (Value, error) {
			return self.Get("_radius"), nil
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
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("_size", args[0])
			return nil
		}).
		PropertyWithSetter("size",
			func(s *State, self Object) (Value, error) {
				return self.Get("_size"), nil
			},
			func(s *State, self Object, val Value) error {
				self.Set("_size", val)
				return nil
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
		StaticMethod("add", func(s *State, args ...Value) (Value, error) {
			a, _ := AsInt(args[0])
			b, _ := AsInt(args[1])
			return Int(a + b), nil
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
		ClassMethod("class_name", func(s *State, cls ClassValue, args ...Value) (Value, error) {
			return String(cls.Name()), nil
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
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("name", args[0])
			return nil
		}).
		Method("speak", func(s *State, self Object, args ...Value) (Value, error) {
			return String("..."), nil
		}).
		Build(state)

	dog := NewClass("Dog").
		Base(animal).
		Method("speak", func(s *State, self Object, args ...Value) (Value, error) {
			name, _ := AsString(self.Get("name"))
			return String(name + " says Woof!"), nil
		}).
		Build(state)

	state.SetGlobal("Animal", animal)
	state.SetGlobal("Dog", dog)

	result := eval(t, state, `Dog("Rex").speak()`)
	if str, ok := AsString(result); !ok || str != "Rex says Woof!" {
		t.Errorf("expected 'Rex says Woof!', got %v", result)
	}

	result = eval(t, state, `isinstance(Dog("Rex"), Animal)`)
	if b, ok := AsBool(result); !ok || !b {
		t.Errorf("expected isinstance(Dog, Animal) = True, got %v", result)
	}
}

func TestClassBuilder_DunderArbitrary(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Addable").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("val", args[0])
			return nil
		}).
		Dunder("__add__", func(s *State, self Object, args ...Value) (Value, error) {
			a, _ := AsInt(self.Get("val"))
			if other, ok := AsObject(args[0]); ok {
				b, _ := AsInt(other.Get("val"))
				return Int(a + b), nil
			}
			return None, nil
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

func TestObject_Class(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("TestClass").Build(state)
	obj := cls.NewInstance()

	if obj.Class().Name() != "TestClass" {
		t.Errorf("expected Class().Name() = 'TestClass', got %q", obj.Class().Name())
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
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("v", args[0])
			return nil
		}).
		Method("double", func(s *State, self Object, args ...Value) (Value, error) {
			n, _ := AsInt(self.Get("v"))
			return Int(n * 2), nil
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
		Init(func(s *State, self Object, args ...Value) error {
			return nil
		}).
		SetItem(func(s *State, self Object, key, val Value) error {
			k, _ := AsString(key)
			self.Set("_item_"+k, val)
			return nil
		}).
		GetItem(func(s *State, self Object, key Value) (Value, error) {
			k, _ := AsString(key)
			return self.Get("_item_" + k), nil
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
	if !IsClass(fooClass) {
		t.Errorf("expected IsClass(Foo) = true, got %T", fooClass)
	}

	fooObj := state.GetGlobal("obj")
	if !IsObject(fooObj) {
		t.Errorf("expected IsObject(obj) = true, got %T", fooObj)
	}
}

func TestClassBuilder_MethodAccessesState(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Runner").
		Method("run_code", func(s *State, self Object, args ...Value) (Value, error) {
			code, _ := AsString(args[0])
			_, err := s.Run(code)
			if err != nil {
				return None, nil
			}
			return None, nil
		}).
		Build(state)

	state.SetGlobal("Runner", cls)

	_, err := state.Run(`_runner_result = Runner().run_code("__x = 1 + 2")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("__x")
	if n, ok := AsInt(result); !ok || n != 3 {
		t.Errorf("expected 3, got %v", result)
	}
}

func TestClassBuilder_ErrorReturn(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Strict").
		Init(func(s *State, self Object, args ...Value) error {
			if len(args) == 0 {
				return ValueError("requires at least one argument")
			}
			self.Set("v", args[0])
			return nil
		}).
		Build(state)

	state.SetGlobal("Strict", cls)

	// Should raise ValueError
	_, err := state.Run(`
try:
    Strict()
    _err_result = "no error"
except ValueError as e:
    _err_result = "caught: " + str(e)
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("_err_result")
	str, ok := AsString(result)
	if !ok || str != "caught: requires at least one argument" {
		t.Errorf("expected 'caught: requires at least one argument', got %v", result)
	}
}

func TestClassBuilder_Comparisons(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Num").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("v", args[0])
			return nil
		}).
		Lt(func(s *State, self Object, other Value) (bool, error) {
			o, ok := AsObject(other)
			if !ok {
				return false, TypeError("unsupported operand type")
			}
			a, _ := AsInt(self.Get("v"))
			b, _ := AsInt(o.Get("v"))
			return a < b, nil
		}).
		Le(func(s *State, self Object, other Value) (bool, error) {
			o, ok := AsObject(other)
			if !ok {
				return false, TypeError("unsupported operand type")
			}
			a, _ := AsInt(self.Get("v"))
			b, _ := AsInt(o.Get("v"))
			return a <= b, nil
		}).
		Gt(func(s *State, self Object, other Value) (bool, error) {
			o, ok := AsObject(other)
			if !ok {
				return false, TypeError("unsupported operand type")
			}
			a, _ := AsInt(self.Get("v"))
			b, _ := AsInt(o.Get("v"))
			return a > b, nil
		}).
		Ge(func(s *State, self Object, other Value) (bool, error) {
			o, ok := AsObject(other)
			if !ok {
				return false, TypeError("unsupported operand type")
			}
			a, _ := AsInt(self.Get("v"))
			b, _ := AsInt(o.Get("v"))
			return a >= b, nil
		}).
		Build(state)

	state.SetGlobal("Num", cls)

	tests := []struct {
		expr string
		want bool
	}{
		{`Num(1) < Num(2)`, true},
		{`Num(2) < Num(1)`, false},
		{`Num(1) <= Num(1)`, true},
		{`Num(2) <= Num(1)`, false},
		{`Num(2) > Num(1)`, true},
		{`Num(1) > Num(2)`, false},
		{`Num(1) >= Num(1)`, true},
		{`Num(1) >= Num(2)`, false},
	}
	for _, tt := range tests {
		result := eval(t, state, tt.expr)
		b, ok := AsBool(result)
		if !ok || b != tt.want {
			t.Errorf("%s: expected %v, got %v", tt.expr, tt.want, result)
		}
	}
}

func TestClassBuilder_Hash(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Hashable").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("v", args[0])
			return nil
		}).
		Hash(func(s *State, self Object) (int64, error) {
			n, _ := AsInt(self.Get("v"))
			return n * 31, nil
		}).
		Eq(func(s *State, self Object, other Value) (bool, error) {
			o, ok := AsObject(other)
			if !ok {
				return false, nil
			}
			a, _ := AsInt(self.Get("v"))
			b, _ := AsInt(o.Get("v"))
			return a == b, nil
		}).
		Build(state)

	state.SetGlobal("Hashable", cls)

	result := eval(t, state, `hash(Hashable(10))`)
	n, ok := AsInt(result)
	if !ok || n != 310 {
		t.Errorf("expected 310, got %v", result)
	}
}

func TestClassBuilder_Iterator(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Range3").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("_idx", Int(0))
			return nil
		}).
		Iter(func(s *State, self Object) (Value, error) {
			return self, nil
		}).
		Next(func(s *State, self Object) (Value, error) {
			idx, _ := AsInt(self.Get("_idx"))
			if idx >= 3 {
				return nil, ErrStopIteration
			}
			self.Set("_idx", Int(idx+1))
			return Int(idx), nil
		}).
		Build(state)

	state.SetGlobal("Range3", cls)

	_, err := state.Run(`_iter_result = [x for x in Range3()]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("_iter_result")
	items, ok := AsList(result)
	if !ok || len(items) != 3 {
		t.Fatalf("expected list of 3, got %v", result)
	}
	for i, want := range []int64{0, 1, 2} {
		n, _ := AsInt(items[i])
		if n != want {
			t.Errorf("item %d: expected %d, got %d", i, want, n)
		}
	}
}

func TestClassBuilder_DelItem(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("DStore").
		Init(func(s *State, self Object, args ...Value) error {
			return nil
		}).
		SetItem(func(s *State, self Object, key, val Value) error {
			k, _ := AsString(key)
			self.Set("_"+k, val)
			return nil
		}).
		GetItem(func(s *State, self Object, key Value) (Value, error) {
			k, _ := AsString(key)
			return self.Get("_" + k), nil
		}).
		DelItem(func(s *State, self Object, key Value) error {
			k, _ := AsString(key)
			self.Delete("_" + k)
			return nil
		}).
		Build(state)

	state.SetGlobal("DStore", cls)

	_, err := state.Run(`
ds = DStore()
ds["x"] = 42
del ds["x"]
_del_result = ds["x"]
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("_del_result")
	if !IsNone(result) {
		t.Errorf("expected None after del, got %v", result)
	}
}

func TestClassBuilder_ContextManager(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("CM").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("entered", False)
			self.Set("exited", False)
			return nil
		}).
		Enter(func(s *State, self Object) (Value, error) {
			self.Set("entered", True)
			return self, nil
		}).
		Exit(func(s *State, self Object, excType, excVal, excTb Value) (bool, error) {
			self.Set("exited", True)
			return false, nil
		}).
		Build(state)

	state.SetGlobal("CM", cls)

	_, err := state.Run(`
cm = CM()
with cm as ctx:
    _in_with = cm.entered
_after_with = cm.exited
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	inWith := state.GetGlobal("_in_with")
	if b, ok := AsBool(inWith); !ok || !b {
		t.Error("expected entered=True inside with block")
	}
	afterWith := state.GetGlobal("_after_with")
	if b, ok := AsBool(afterWith); !ok || !b {
		t.Error("expected exited=True after with block")
	}
}

func TestClassBuilder_MultipleBases(t *testing.T) {
	state := NewState()
	defer state.Close()

	flyable := NewClass("Flyable").
		Method("fly", func(s *State, self Object, args ...Value) (Value, error) {
			return String("flying"), nil
		}).
		Build(state)

	swimmable := NewClass("Swimmable").
		Method("swim", func(s *State, self Object, args ...Value) (Value, error) {
			return String("swimming"), nil
		}).
		Build(state)

	duck := NewClass("Duck").
		Bases(flyable, swimmable).
		Build(state)

	state.SetGlobal("Flyable", flyable)
	state.SetGlobal("Swimmable", swimmable)
	state.SetGlobal("Duck", duck)

	result := eval(t, state, `Duck().fly()`)
	if str, ok := AsString(result); !ok || str != "flying" {
		t.Errorf("expected 'flying', got %v", result)
	}

	result = eval(t, state, `Duck().swim()`)
	if str, ok := AsString(result); !ok || str != "swimming" {
		t.Errorf("expected 'swimming', got %v", result)
	}
}

func TestAsObjectAsClass(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("X").Build(state)
	obj := cls.NewInstance()

	if _, ok := AsObject(obj); !ok {
		t.Error("AsObject should succeed on Object")
	}
	if _, ok := AsClass(cls); !ok {
		t.Error("AsClass should succeed on ClassValue")
	}
	if _, ok := AsObject(cls); ok {
		t.Error("AsObject should fail on ClassValue")
	}
	if _, ok := AsClass(obj); ok {
		t.Error("AsClass should fail on Object")
	}
}

func TestClassBuilder_Attr(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("WithAttr").
		Attr("VERSION", Int(42)).
		Attr("NAME", String("test")).
		Build(state)

	state.SetGlobal("WithAttr", cls)

	result := eval(t, state, `WithAttr.VERSION`)
	if n, ok := AsInt(result); !ok || n != 42 {
		t.Errorf("expected 42, got %v", result)
	}

	result = eval(t, state, `WithAttr.NAME`)
	if s, ok := AsString(result); !ok || s != "test" {
		t.Errorf("expected 'test', got %v", result)
	}

	// Attrs should also be accessible on instances
	result = eval(t, state, `WithAttr().VERSION`)
	if n, ok := AsInt(result); !ok || n != 42 {
		t.Errorf("expected 42 on instance, got %v", result)
	}
}

func TestClassBuilder_AddSubMul(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Num").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("v", args[0])
			return nil
		}).
		Add(func(s *State, self Object, other Value) (Value, error) {
			a, _ := AsInt(self.Get("v"))
			o, ok := AsObject(other)
			if !ok {
				return None, nil
			}
			b, _ := AsInt(o.Get("v"))
			return Int(a + b), nil
		}).
		Sub(func(s *State, self Object, other Value) (Value, error) {
			a, _ := AsInt(self.Get("v"))
			o, ok := AsObject(other)
			if !ok {
				return None, nil
			}
			b, _ := AsInt(o.Get("v"))
			return Int(a - b), nil
		}).
		Mul(func(s *State, self Object, other Value) (Value, error) {
			a, _ := AsInt(self.Get("v"))
			o, ok := AsObject(other)
			if !ok {
				return None, nil
			}
			b, _ := AsInt(o.Get("v"))
			return Int(a * b), nil
		}).
		Build(state)

	state.SetGlobal("Num", cls)

	result := eval(t, state, `Num(10) + Num(5)`)
	if n, ok := AsInt(result); !ok || n != 15 {
		t.Errorf("Add: expected 15, got %v", result)
	}

	result = eval(t, state, `Num(10) - Num(3)`)
	if n, ok := AsInt(result); !ok || n != 7 {
		t.Errorf("Sub: expected 7, got %v", result)
	}

	result = eval(t, state, `Num(4) * Num(5)`)
	if n, ok := AsInt(result); !ok || n != 20 {
		t.Errorf("Mul: expected 20, got %v", result)
	}
}

func TestClassBuilder_NegAbs(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Num").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("v", args[0])
			return nil
		}).
		Neg(func(s *State, self Object) (Value, error) {
			n, _ := AsInt(self.Get("v"))
			return Int(-n), nil
		}).
		Abs(func(s *State, self Object) (Value, error) {
			n, _ := AsInt(self.Get("v"))
			if n < 0 {
				return Int(-n), nil
			}
			return Int(n), nil
		}).
		Build(state)

	state.SetGlobal("Num", cls)

	result := eval(t, state, `-Num(5)`)
	if n, ok := AsInt(result); !ok || n != -5 {
		t.Errorf("Neg: expected -5, got %v", result)
	}

	result = eval(t, state, `abs(Num(-7))`)
	if n, ok := AsInt(result); !ok || n != 7 {
		t.Errorf("Abs: expected 7, got %v", result)
	}
}

func TestClassBuilder_GetAttrDunder(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Dynamic").
		GetAttr(func(s *State, self Object, name string) (Value, error) {
			return String("dynamic:" + name), nil
		}).
		Build(state)

	state.SetGlobal("Dynamic", cls)

	result := eval(t, state, `Dynamic().foo`)
	if s, ok := AsString(result); !ok || s != "dynamic:foo" {
		t.Errorf("expected 'dynamic:foo', got %v", result)
	}

	result = eval(t, state, `Dynamic().bar`)
	if s, ok := AsString(result); !ok || s != "dynamic:bar" {
		t.Errorf("expected 'dynamic:bar', got %v", result)
	}
}

func TestClassBuilder_SetAttrDunder(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Tracked").
		SetAttr(func(s *State, self Object, name string, val Value) error {
			// Store under a prefixed key
			self.Set("_tracked_"+name, val)
			return nil
		}).
		Build(state)

	state.SetGlobal("Tracked", cls)

	_, err := state.Run(`
t = Tracked()
t.x = 42
_setattr_result = t._tracked_x
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("_setattr_result")
	if n, ok := AsInt(result); !ok || n != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestClassBuilder_DelAttrDunder(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Deletable").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("x", Int(10))
			self.Set("_deleted", False)
			return nil
		}).
		DelAttr(func(s *State, self Object, name string) error {
			self.Delete(name)
			self.Set("_deleted", True)
			return nil
		}).
		Build(state)

	state.SetGlobal("Deletable", cls)

	_, err := state.Run(`
d = Deletable()
del d.x
_delattr_result = d._deleted
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("_delattr_result")
	if b, ok := AsBool(result); !ok || !b {
		t.Errorf("expected True, got %v", result)
	}
}

func TestClassBuilder_New(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Singleton").
		New(func(s *State, cls ClassValue, args ...Value) (Object, error) {
			inst := cls.NewInstance()
			inst.Set("created_by_new", True)
			return inst, nil
		}).
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("initialized", True)
			return nil
		}).
		Build(state)

	state.SetGlobal("Singleton", cls)

	_, err := state.Run(`
s = Singleton()
_new_created = s.created_by_new
_new_init = s.initialized
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("_new_created")
	if b, ok := AsBool(result); !ok || !b {
		t.Errorf("expected created_by_new=True, got %v", result)
	}
	result = state.GetGlobal("_new_init")
	if b, ok := AsBool(result); !ok || !b {
		t.Errorf("expected initialized=True, got %v", result)
	}
}

func TestClassBuilder_IntConv(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("IntLike").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("v", args[0])
			return nil
		}).
		IntConv(func(s *State, self Object) (int64, error) {
			f, _ := AsFloat(self.Get("v"))
			return int64(f), nil
		}).
		Build(state)

	state.SetGlobal("IntLike", cls)

	result := eval(t, state, `int(IntLike(3.7))`)
	if n, ok := AsInt(result); !ok || n != 3 {
		t.Errorf("expected 3, got %v", result)
	}
}

func TestClassBuilder_FloatConv(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("FloatLike").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("v", args[0])
			return nil
		}).
		FloatConv(func(s *State, self Object) (float64, error) {
			n, _ := AsInt(self.Get("v"))
			return float64(n) + 0.5, nil
		}).
		Build(state)

	state.SetGlobal("FloatLike", cls)

	result := eval(t, state, `float(FloatLike(3))`)
	if f, ok := AsFloat(result); !ok || f != 3.5 {
		t.Errorf("expected 3.5, got %v", result)
	}
}

func TestClassBuilder_Index(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Idx").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("v", args[0])
			return nil
		}).
		Index(func(s *State, self Object) (int64, error) {
			n, _ := AsInt(self.Get("v"))
			return n, nil
		}).
		Build(state)

	state.SetGlobal("Idx", cls)

	// __index__ allows use as a list index
	result := eval(t, state, `[10, 20, 30][Idx(1)]`)
	if n, ok := AsInt(result); !ok || n != 20 {
		t.Errorf("expected 20, got %v", result)
	}
}

func TestClassBuilder_Format(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Fmt").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("v", args[0])
			return nil
		}).
		Format(func(s *State, self Object, spec string) (string, error) {
			n, _ := AsInt(self.Get("v"))
			if spec == "hex" {
				return fmt.Sprintf("0x%x", n), nil
			}
			return fmt.Sprintf("%d", n), nil
		}).
		Build(state)

	state.SetGlobal("Fmt", cls)

	result := eval(t, state, `format(Fmt(255), "hex")`)
	if s, ok := AsString(result); !ok || s != "0xff" {
		t.Errorf("expected '0xff', got %v", result)
	}

	result = eval(t, state, `format(Fmt(42), "")`)
	if s, ok := AsString(result); !ok || s != "42" {
		t.Errorf("expected '42', got %v", result)
	}
}

func TestClassBuilder_Missing(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("DefaultDict").
		Init(func(s *State, self Object, args ...Value) error {
			return nil
		}).
		GetItem(func(s *State, self Object, key Value) (Value, error) {
			k, _ := AsString(key)
			v := self.Get("_item_" + k)
			if IsNone(v) {
				return String("default"), nil
			}
			return v, nil
		}).
		SetItem(func(s *State, self Object, key, val Value) error {
			k, _ := AsString(key)
			self.Set("_item_"+k, val)
			return nil
		}).
		Missing(func(s *State, self Object, key Value) (Value, error) {
			return String("missing"), nil
		}).
		Build(state)

	state.SetGlobal("DefaultDict", cls)

	result := eval(t, state, `DefaultDict()["nonexistent"]`)
	if s, ok := AsString(result); !ok || s != "default" {
		t.Errorf("expected 'default', got %v", result)
	}
}

func TestClassBuilder_InPlaceOps(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Acc").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("v", args[0])
			return nil
		}).
		IAdd(func(s *State, self Object, other Value) (Value, error) {
			a, _ := AsInt(self.Get("v"))
			b, _ := AsInt(other)
			self.Set("v", Int(a+b))
			return self, nil
		}).
		Str(func(s *State, self Object) (string, error) {
			n, _ := AsInt(self.Get("v"))
			return fmt.Sprintf("Acc(%d)", n), nil
		}).
		Build(state)

	state.SetGlobal("Acc", cls)

	_, err := state.Run(`
a = Acc(10)
a += 5
_iadd_result = a.v
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("_iadd_result")
	if n, ok := AsInt(result); !ok || n != 15 {
		t.Errorf("expected 15, got %v", result)
	}
}

func TestClassBuilder_DescGet(t *testing.T) {
	state := NewState()
	defer state.Close()

	// Build a descriptor class
	desc := NewClass("Desc").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("val", args[0])
			return nil
		}).
		DescGet(func(s *State, self Object, instance, owner Value) (Value, error) {
			if IsNone(instance) {
				return self, nil // class access returns descriptor itself
			}
			return self.Get("val"), nil
		}).
		DescSet(func(s *State, self Object, instance, val Value) error {
			self.Set("val", val)
			return nil
		}).
		Build(state)

	state.SetGlobal("Desc", desc)

	_, err := state.Run(`
class MyClass:
    attr = Desc(42)

obj = MyClass()
_desc_get_result = obj.attr
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("_desc_get_result")
	if n, ok := AsInt(result); !ok || n != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestClassBuilder_BytesValue(t *testing.T) {
	// Test BytesValue type
	b := Bytes([]byte("hello"))
	if b.Type() != "bytes" {
		t.Errorf("expected type 'bytes', got %q", b.Type())
	}
	if !IsBytes(b) {
		t.Error("expected IsBytes = true")
	}
	bs, ok := AsBytes(b)
	if !ok || string(bs) != "hello" {
		t.Errorf("expected 'hello', got %v", bs)
	}
}

func TestClassBuilder_ReverseOps(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("RNum").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("v", args[0])
			return nil
		}).
		RMul(func(s *State, self Object, other Value) (Value, error) {
			a, _ := AsInt(self.Get("v"))
			b, _ := AsInt(other)
			return Int(a * b), nil
		}).
		Build(state)

	state.SetGlobal("RNum", cls)

	result := eval(t, state, `3 * RNum(5)`)
	if n, ok := AsInt(result); !ok || n != 15 {
		t.Errorf("expected 15, got %v", result)
	}
}

func TestClassBuilder_CallKw(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("KwCallable").
		CallKw(func(s *State, self Object, args []Value, kwargs map[string]Value) (Value, error) {
			// Return the value of the "key" kwarg, or the first arg
			if v, ok := kwargs["key"]; ok {
				return v, nil
			}
			if len(args) > 0 {
				return args[0], nil
			}
			return None, nil
		}).
		Build(state)

	state.SetGlobal("KwCallable", cls)

	result := eval(t, state, `KwCallable()(key=42)`)
	if n, ok := AsInt(result); !ok || n != 42 {
		t.Errorf("expected 42, got %v", result)
	}

	result = eval(t, state, `KwCallable()(99)`)
	if n, ok := AsInt(result); !ok || n != 99 {
		t.Errorf("expected 99, got %v", result)
	}
}

func TestClassBuilder_Reversed(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("RevList").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("items", args[0])
			self.Set("_idx", None)
			return nil
		}).
		Reversed(func(s *State, self Object) (Value, error) {
			items, _ := AsList(self.Get("items"))
			reversed := make([]Value, len(items))
			for i, v := range items {
				reversed[len(items)-1-i] = v
			}
			return List(reversed...), nil
		}).
		Build(state)

	state.SetGlobal("RevList", cls)

	result := eval(t, state, `list(reversed(RevList([1, 2, 3])))`)
	items, ok := AsList(result)
	if !ok || len(items) != 3 {
		t.Fatalf("expected list of 3, got %v", result)
	}
	for i, want := range []int64{3, 2, 1} {
		n, _ := AsInt(items[i])
		if n != want {
			t.Errorf("item %d: expected %d, got %d", i, want, n)
		}
	}
}

func TestClassBuilder_Round(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("RoundNum").
		Init(func(s *State, self Object, args ...Value) error {
			self.Set("v", args[0])
			return nil
		}).
		Round(func(s *State, self Object, ndigits Value) (Value, error) {
			f, _ := AsFloat(self.Get("v"))
			if IsNone(ndigits) {
				return Int(int64(f + 0.5)), nil
			}
			// For simplicity, just return the int version
			return Int(int64(f + 0.5)), nil
		}).
		Build(state)

	state.SetGlobal("RoundNum", cls)

	result := eval(t, state, `round(RoundNum(3.7))`)
	if n, ok := AsInt(result); !ok || n != 4 {
		t.Errorf("expected 4, got %v", result)
	}
}

func TestClassBuilder_ClassGetItem(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("Generic").
		ClassGetItem(func(s *State, cls ClassValue, key Value) (Value, error) {
			k, _ := AsString(key)
			return String(cls.Name() + "[" + k + "]"), nil
		}).
		Build(state)

	state.SetGlobal("Generic", cls)

	result := eval(t, state, `Generic["int"]`)
	if s, ok := AsString(result); !ok || s != "Generic[int]" {
		t.Errorf("expected 'Generic[int]', got %v", result)
	}
}

func TestClassBuilder_SetName(t *testing.T) {
	state := NewState()
	defer state.Close()

	cls := NewClass("NamedDesc").
		Init(func(s *State, self Object, args ...Value) error {
			return nil
		}).
		SetName(func(s *State, self Object, owner Value, name string) error {
			self.Set("attr_name", String(name))
			return nil
		}).
		DescGet(func(s *State, self Object, instance, owner Value) (Value, error) {
			return self.Get("attr_name"), nil
		}).
		Build(state)

	state.SetGlobal("NamedDesc", cls)

	_, err := state.Run(`
class Host:
    my_field = NamedDesc()

_setname_result = Host().my_field
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := state.GetGlobal("_setname_result")
	if s, ok := AsString(result); !ok || s != "my_field" {
		t.Errorf("expected 'my_field', got %v", result)
	}
}
