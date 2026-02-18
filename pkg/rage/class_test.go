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
