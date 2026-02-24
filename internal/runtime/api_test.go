package runtime

import (
	"testing"
)

// =====================================
// Value Constructors: NewInt, NewFloat, NewString, etc.
// =====================================

func TestNewInt(t *testing.T) {
	v := NewInt(42)
	if v.Value != 42 {
		t.Errorf("NewInt(42).Value = %d, want 42", v.Value)
	}

	v = NewInt(0)
	if v.Value != 0 {
		t.Errorf("NewInt(0).Value = %d, want 0", v.Value)
	}

	v = NewInt(-100)
	if v.Value != -100 {
		t.Errorf("NewInt(-100).Value = %d, want -100", v.Value)
	}
}

func TestNewFloat(t *testing.T) {
	v := NewFloat(3.14)
	if v.Value != 3.14 {
		t.Errorf("NewFloat(3.14).Value = %f, want 3.14", v.Value)
	}

	v = NewFloat(0.0)
	if v.Value != 0.0 {
		t.Errorf("NewFloat(0.0).Value = %f, want 0.0", v.Value)
	}

	v = NewFloat(-1.5)
	if v.Value != -1.5 {
		t.Errorf("NewFloat(-1.5).Value = %f, want -1.5", v.Value)
	}
}

func TestNewComplex(t *testing.T) {
	v := NewComplex(1.0, 2.0)
	if v.Real != 1.0 || v.Imag != 2.0 {
		t.Errorf("NewComplex(1,2) = (%f, %f), want (1, 2)", v.Real, v.Imag)
	}

	v = NewComplex(0, 0)
	if v.Real != 0 || v.Imag != 0 {
		t.Errorf("NewComplex(0,0) = (%f, %f), want (0, 0)", v.Real, v.Imag)
	}
}

func TestNewString(t *testing.T) {
	v := NewString("hello")
	if v.Value != "hello" {
		t.Errorf("NewString('hello').Value = %q", v.Value)
	}

	v = NewString("")
	if v.Value != "" {
		t.Errorf("NewString('').Value = %q", v.Value)
	}
}

func TestNewStringInterning(t *testing.T) {
	// Short strings should be interned (same pointer)
	a := NewString("test")
	b := NewString("test")
	if a != b {
		t.Error("NewString should return interned string for short strings")
	}
}

func TestNewBool(t *testing.T) {
	v := NewBool(true)
	if v != True {
		t.Error("NewBool(true) should return True singleton")
	}

	v = NewBool(false)
	if v != False {
		t.Error("NewBool(false) should return False singleton")
	}
}

func TestNewList(t *testing.T) {
	items := []Value{MakeInt(1), MakeInt(2), MakeInt(3)}
	v := NewList(items)
	if len(v.Items) != 3 {
		t.Fatalf("NewList len = %d, want 3", len(v.Items))
	}
	for i, want := range []int64{1, 2, 3} {
		if v.Items[i].(*PyInt).Value != want {
			t.Errorf("Items[%d] = %v, want %d", i, v.Items[i], want)
		}
	}
}

func TestNewListEmpty(t *testing.T) {
	v := NewList(nil)
	if v.Items != nil {
		t.Error("NewList(nil) should have nil Items")
	}

	v = NewList([]Value{})
	if len(v.Items) != 0 {
		t.Error("NewList([]Value{}) should have empty Items")
	}
}

func TestNewTuple(t *testing.T) {
	items := []Value{&PyString{Value: "a"}, &PyString{Value: "b"}}
	v := NewTuple(items)
	if len(v.Items) != 2 {
		t.Fatalf("NewTuple len = %d, want 2", len(v.Items))
	}
}

func TestNewDict(t *testing.T) {
	v := NewDict()
	if v == nil {
		t.Fatal("NewDict returned nil")
	}
	if v.Items == nil {
		t.Error("Items map should be initialized")
	}
	if len(v.Items) != 0 {
		t.Errorf("new dict should be empty, got %d items", len(v.Items))
	}
}

func TestNewBytes(t *testing.T) {
	data := []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}
	v := NewBytes(data)
	if len(v.Value) != 5 {
		t.Fatalf("NewBytes len = %d, want 5", len(v.Value))
	}
	if string(v.Value) != "Hello" {
		t.Errorf("NewBytes value = %q, want 'Hello'", string(v.Value))
	}
}

func TestNewBytesEmpty(t *testing.T) {
	v := NewBytes([]byte{})
	if len(v.Value) != 0 {
		t.Error("NewBytes([]byte{}) should be empty")
	}
}

// =====================================
// SetGlobal / GetGlobal
// =====================================

func TestSetGetGlobal(t *testing.T) {
	vm := NewVM()

	vm.SetGlobal("x", MakeInt(42))
	val := vm.GetGlobal("x")
	if v, ok := val.(*PyInt); !ok || v.Value != 42 {
		t.Errorf("GetGlobal('x') = %v, want 42", val)
	}
}

func TestGetGlobalMissing(t *testing.T) {
	vm := NewVM()
	val := vm.GetGlobal("nonexistent")
	if _, ok := val.(*PyNone); !ok {
		t.Errorf("GetGlobal('nonexistent') = %v, want None", val)
	}
}

func TestSetGlobalOverwrite(t *testing.T) {
	vm := NewVM()

	vm.SetGlobal("x", MakeInt(1))
	vm.SetGlobal("x", MakeInt(2))
	val := vm.GetGlobal("x")
	if v := val.(*PyInt).Value; v != 2 {
		t.Errorf("GetGlobal('x') = %d, want 2 after overwrite", v)
	}
}

// =====================================
// SetBuiltin / GetBuiltin
// =====================================

func TestSetGetBuiltin(t *testing.T) {
	vm := NewVM()

	vm.SetBuiltin("my_func", NewGoFunction("my_func", func(vm *VM) int { return 0 }))
	val := vm.GetBuiltin("my_func")
	if _, ok := val.(*PyGoFunc); !ok {
		t.Errorf("GetBuiltin('my_func') = %T, expected *PyGoFunc", val)
	}
}

func TestGetBuiltinMissing(t *testing.T) {
	vm := NewVM()
	val := vm.GetBuiltin("definitely_not_a_builtin")
	if _, ok := val.(*PyNone); !ok {
		t.Errorf("GetBuiltin(missing) = %v, want None", val)
	}
}

// =====================================
// IsNone
// =====================================

func TestIsNone(t *testing.T) {
	if !IsNone(None) {
		t.Error("IsNone(None) should be true")
	}
	if IsNone(MakeInt(0)) {
		t.Error("IsNone(0) should be false")
	}
	if IsNone(False) {
		t.Error("IsNone(False) should be false")
	}
	if IsNone(&PyString{Value: ""}) {
		t.Error("IsNone('') should be false")
	}
}

// =====================================
// IsTrue (standalone function)
// =====================================

func TestIsTrueFunction(t *testing.T) {
	tests := []struct {
		name string
		v    Value
		want bool
	}{
		{"None", None, false},
		{"True", True, true},
		{"False", False, false},
		{"int 0", MakeInt(0), false},
		{"int 1", MakeInt(1), true},
		{"int -1", MakeInt(-1), true},
		{"float 0.0", &PyFloat{Value: 0.0}, false},
		{"float 1.0", &PyFloat{Value: 1.0}, true},
		{"empty string", &PyString{Value: ""}, false},
		{"nonempty string", &PyString{Value: "x"}, true},
		{"empty list", &PyList{Items: []Value{}}, false},
		{"nonempty list", &PyList{Items: []Value{MakeInt(1)}}, true},
		{"empty tuple", &PyTuple{Items: []Value{}}, false},
		{"nonempty tuple", &PyTuple{Items: []Value{MakeInt(1)}}, true},
		{"empty dict", &PyDict{Items: make(map[Value]Value)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTrue(tt.v)
			if got != tt.want {
				t.Errorf("IsTrue(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

// =====================================
// Stack API: Push, Pop, GetTop, SetTop, Get
// =====================================

// newAPIVM creates a VM with a frame suitable for API-level stack tests.
func newAPIVM(stackSize int) *VM {
	vm := NewVM()
	frame := &Frame{
		Code: &CodeObject{
			Name:      "<api_test>",
			Filename:  "<api_test>",
			StackSize: stackSize,
		},
		Stack: make([]Value, stackSize),
		SP:    0,
	}
	vm.frames = []*Frame{frame}
	vm.frame = frame
	return vm
}

func TestPushPopAPI(t *testing.T) {
	vm := newAPIVM(16)

	vm.Push(MakeInt(10))
	vm.Push(MakeInt(20))

	val := vm.Pop()
	if v, ok := val.(*PyInt); !ok || v.Value != 20 {
		t.Errorf("Pop() = %v, want 20", val)
	}
	val = vm.Pop()
	if v, ok := val.(*PyInt); !ok || v.Value != 10 {
		t.Errorf("Pop() = %v, want 10", val)
	}
}

func TestGetTop(t *testing.T) {
	vm := newAPIVM(16)
	if vm.GetTop() != 0 {
		t.Errorf("GetTop() = %d, want 0", vm.GetTop())
	}

	vm.Push(MakeInt(1))
	if vm.GetTop() != 1 {
		t.Errorf("GetTop() = %d, want 1", vm.GetTop())
	}

	vm.Push(MakeInt(2))
	if vm.GetTop() != 2 {
		t.Errorf("GetTop() = %d, want 2", vm.GetTop())
	}
}

func TestSetTopPositive(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(1))
	vm.Push(MakeInt(2))
	vm.Push(MakeInt(3))

	vm.SetTop(1) // Reduce stack to 1 element
	if vm.GetTop() != 1 {
		t.Errorf("GetTop() = %d, want 1 after SetTop(1)", vm.GetTop())
	}
}

func TestSetTopNegative(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(1))
	vm.Push(MakeInt(2))
	vm.Push(MakeInt(3))

	vm.SetTop(-1) // Remove top element (SP = SP + (-1) + 1 = SP)
	if vm.GetTop() != 3 {
		t.Errorf("GetTop() = %d, want 3 after SetTop(-1)", vm.GetTop())
	}

	vm.SetTop(-2) // SP = SP + (-2) + 1 = SP - 1
	if vm.GetTop() != 2 {
		t.Errorf("GetTop() = %d, want 2 after SetTop(-2)", vm.GetTop())
	}
}

func TestSetTopZero(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(1))
	vm.Push(MakeInt(2))

	vm.SetTop(0) // Clear stack
	if vm.GetTop() != 0 {
		t.Errorf("GetTop() = %d, want 0 after SetTop(0)", vm.GetTop())
	}
}

func TestGetPositiveIndex(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(10)) // index 1 (1-based)
	vm.Push(MakeInt(20)) // index 2
	vm.Push(MakeInt(30)) // index 3

	val := vm.Get(1)
	if v, ok := val.(*PyInt); !ok || v.Value != 10 {
		t.Errorf("Get(1) = %v, want 10", val)
	}

	val = vm.Get(2)
	if v, ok := val.(*PyInt); !ok || v.Value != 20 {
		t.Errorf("Get(2) = %v, want 20", val)
	}

	val = vm.Get(3)
	if v, ok := val.(*PyInt); !ok || v.Value != 30 {
		t.Errorf("Get(3) = %v, want 30", val)
	}
}

func TestGetNegativeIndex(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(10))
	vm.Push(MakeInt(20))
	vm.Push(MakeInt(30))

	// -1 = top of stack
	val := vm.Get(-1)
	if v, ok := val.(*PyInt); !ok || v.Value != 30 {
		t.Errorf("Get(-1) = %v, want 30", val)
	}

	val = vm.Get(-2)
	if v, ok := val.(*PyInt); !ok || v.Value != 20 {
		t.Errorf("Get(-2) = %v, want 20", val)
	}

	val = vm.Get(-3)
	if v, ok := val.(*PyInt); !ok || v.Value != 10 {
		t.Errorf("Get(-3) = %v, want 10", val)
	}
}

func TestGetOutOfRangeReturnsNone(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(1))

	val := vm.Get(5) // Out of range
	if _, ok := val.(*PyNone); !ok {
		t.Errorf("Get(5) = %v, want None for out of range", val)
	}

	val = vm.Get(0) // 0 index is invalid (1-based, and not negative)
	if _, ok := val.(*PyNone); !ok {
		t.Errorf("Get(0) = %v, want None for zero index", val)
	}

	val = vm.Get(-5) // Too far negative
	if _, ok := val.(*PyNone); !ok {
		t.Errorf("Get(-5) = %v, want None for out-of-range negative", val)
	}
}

// =====================================
// Type checking: CheckInt, CheckFloat, CheckString, CheckBool, CheckList, CheckDict
// =====================================

func TestCheckInt(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(42))

	got := vm.CheckInt(1) // 1-based index
	if got != 42 {
		t.Errorf("CheckInt(1) = %d, want 42", got)
	}
}

func TestCheckFloat(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(&PyFloat{Value: 3.14})

	got := vm.CheckFloat(1)
	if got != 3.14 {
		t.Errorf("CheckFloat(1) = %f, want 3.14", got)
	}
}

func TestCheckFloatFromInt(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(5))

	// toFloat should convert int to float
	got := vm.CheckFloat(1)
	if got != 5.0 {
		t.Errorf("CheckFloat(1) from int = %f, want 5.0", got)
	}
}

func TestCheckString(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(&PyString{Value: "hello"})

	got := vm.CheckString(1)
	if got != "hello" {
		t.Errorf("CheckString(1) = %q, want %q", got, "hello")
	}
}

func TestCheckBool(t *testing.T) {
	vm := newAPIVM(16)

	vm.Push(True)
	if !vm.CheckBool(1) {
		t.Error("CheckBool(1) with True should be true")
	}

	vm.SetTop(0)
	vm.Push(False)
	if vm.CheckBool(1) {
		t.Error("CheckBool(1) with False should be false")
	}
}

func TestCheckList(t *testing.T) {
	vm := newAPIVM(16)
	list := &PyList{Items: []Value{MakeInt(1)}}
	vm.Push(list)

	got := vm.CheckList(1)
	if got != list {
		t.Error("CheckList(1) should return the same list")
	}
}

func TestCheckListNonList(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(42))

	got := vm.CheckList(1)
	if got != nil {
		t.Error("CheckList on non-list should return nil")
	}
}

func TestCheckDict(t *testing.T) {
	vm := newAPIVM(16)
	d := NewDict()
	vm.Push(d)

	got := vm.CheckDict(1)
	if got != d {
		t.Error("CheckDict(1) should return the same dict")
	}
}

func TestCheckDictNonDict(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(&PyString{Value: "not a dict"})

	got := vm.CheckDict(1)
	if got != nil {
		t.Error("CheckDict on non-dict should return nil")
	}
}

// =====================================
// Conversion: ToInt, ToFloat, ToString, ToBool
// =====================================

func TestToInt(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(42))

	got := vm.ToInt(1)
	if got != 42 {
		t.Errorf("ToInt(1) = %d, want 42", got)
	}
}

func TestToIntFromFloat(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(&PyFloat{Value: 3.7})

	got := vm.ToInt(1)
	if got != 3 {
		t.Errorf("ToInt(1) from 3.7 = %d, want 3 (truncated)", got)
	}
}

func TestToIntFromBool(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(True)
	if vm.ToInt(1) != 1 {
		t.Error("ToInt(True) should be 1")
	}

	vm.SetTop(0)
	vm.Push(False)
	if vm.ToInt(1) != 0 {
		t.Error("ToInt(False) should be 0")
	}
}

func TestToFloat(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(&PyFloat{Value: 2.5})

	got := vm.ToFloat(1)
	if got != 2.5 {
		t.Errorf("ToFloat(1) = %f, want 2.5", got)
	}
}

func TestToFloatFromInt(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(7))

	got := vm.ToFloat(1)
	if got != 7.0 {
		t.Errorf("ToFloat(1) from int = %f, want 7.0", got)
	}
}

func TestToString(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(&PyString{Value: "hello"})

	got := vm.ToString(1)
	if got != "hello" {
		t.Errorf("ToString(1) = %q, want %q", got, "hello")
	}
}

func TestToStringFromInt(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(42))

	got := vm.ToString(1)
	if got != "42" {
		t.Errorf("ToString(1) from int = %q, want %q", got, "42")
	}
}

func TestToStringFromNone(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(None)

	got := vm.ToString(1)
	if got != "None" {
		t.Errorf("ToString(1) from None = %q, want %q", got, "None")
	}
}

func TestToBool(t *testing.T) {
	vm := newAPIVM(16)

	vm.Push(True)
	if !vm.ToBool(1) {
		t.Error("ToBool(True) should be true")
	}

	vm.SetTop(0)
	vm.Push(MakeInt(0))
	if vm.ToBool(1) {
		t.Error("ToBool(0) should be false")
	}

	vm.SetTop(0)
	vm.Push(MakeInt(1))
	if !vm.ToBool(1) {
		t.Error("ToBool(1) should be true")
	}

	vm.SetTop(0)
	vm.Push(None)
	if vm.ToBool(1) {
		t.Error("ToBool(None) should be false")
	}
}

// =====================================
// RequireArgs
// =====================================

func TestRequireArgsSuccess(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(1))
	vm.Push(MakeInt(2))

	// Should not panic
	ok := vm.RequireArgs("test_func", 2)
	if !ok {
		t.Error("RequireArgs should return true when enough args")
	}
}

func TestRequireArgsMoreThanRequired(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(1))
	vm.Push(MakeInt(2))
	vm.Push(MakeInt(3))

	ok := vm.RequireArgs("test_func", 2)
	if !ok {
		t.Error("RequireArgs should return true when more than enough args")
	}
}

func TestRequireArgsFails(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(1))

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic from RequireArgs when not enough args")
		}
		pe, ok := r.(*PyPanicError)
		if !ok {
			t.Errorf("expected *PyPanicError, got %T", r)
			return
		}
		if pe.ExcType != "TypeError" {
			t.Errorf("ExcType = %q, want %q", pe.ExcType, "TypeError")
		}
	}()

	vm.RequireArgs("my_func", 3)
}

// =====================================
// OptionalInt
// =====================================

func TestOptionalIntPresent(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(99))

	got := vm.OptionalInt(1, 42)
	if got != 99 {
		t.Errorf("OptionalInt(1, 42) = %d, want 99", got)
	}
}

func TestOptionalIntMissing(t *testing.T) {
	vm := newAPIVM(16)
	// Stack is empty

	got := vm.OptionalInt(1, 42)
	if got != 42 {
		t.Errorf("OptionalInt(1, 42) = %d, want 42 (default)", got)
	}
}

func TestOptionalIntNone(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(None)

	got := vm.OptionalInt(1, 42)
	if got != 42 {
		t.Errorf("OptionalInt(1, 42) with None = %d, want 42 (default)", got)
	}
}

// =====================================
// OptionalFloat
// =====================================

func TestOptionalFloatPresent(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(&PyFloat{Value: 2.5})

	got := vm.OptionalFloat(1, 1.0)
	if got != 2.5 {
		t.Errorf("OptionalFloat(1, 1.0) = %f, want 2.5", got)
	}
}

func TestOptionalFloatMissing(t *testing.T) {
	vm := newAPIVM(16)

	got := vm.OptionalFloat(1, 1.0)
	if got != 1.0 {
		t.Errorf("OptionalFloat(1, 1.0) = %f, want 1.0 (default)", got)
	}
}

func TestOptionalFloatNone(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(None)

	got := vm.OptionalFloat(1, 1.0)
	if got != 1.0 {
		t.Errorf("OptionalFloat(1, 1.0) with None = %f, want 1.0 (default)", got)
	}
}

// =====================================
// OptionalString
// =====================================

func TestOptionalStringPresent(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(&PyString{Value: "hello"})

	got := vm.OptionalString(1, "default")
	if got != "hello" {
		t.Errorf("OptionalString = %q, want %q", got, "hello")
	}
}

func TestOptionalStringMissing(t *testing.T) {
	vm := newAPIVM(16)

	got := vm.OptionalString(1, "default")
	if got != "default" {
		t.Errorf("OptionalString = %q, want %q", got, "default")
	}
}

func TestOptionalStringNone(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(None)

	got := vm.OptionalString(1, "default")
	if got != "default" {
		t.Errorf("OptionalString with None = %q, want %q", got, "default")
	}
}

func TestOptionalStringNonStringValue(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(42))

	// Non-string value should be converted via str()
	got := vm.OptionalString(1, "default")
	if got != "42" {
		t.Errorf("OptionalString with int = %q, want %q", got, "42")
	}
}

// =====================================
// OptionalBool
// =====================================

func TestOptionalBoolPresent(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(True)

	got := vm.OptionalBool(1, false)
	if !got {
		t.Error("OptionalBool(1, false) with True should be true")
	}
}

func TestOptionalBoolMissing(t *testing.T) {
	vm := newAPIVM(16)

	got := vm.OptionalBool(1, true)
	if !got {
		t.Error("OptionalBool(1, true) missing should return default true")
	}
}

func TestOptionalBoolNone(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(None)

	got := vm.OptionalBool(1, true)
	if !got {
		t.Error("OptionalBool(1, true) with None should return default true")
	}
}

// =====================================
// RaiseError
// =====================================

func TestRaiseErrorValueError(t *testing.T) {
	vm := newAPIVM(16)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from RaiseError")
		}
		pe, ok := r.(*PyPanicError)
		if !ok {
			t.Fatalf("expected *PyPanicError, got %T", r)
		}
		if pe.ExcType != "ValueError" {
			t.Errorf("ExcType = %q, want %q", pe.ExcType, "ValueError")
		}
		if pe.Message != "invalid input" {
			t.Errorf("Message = %q, want %q", pe.Message, "invalid input")
		}
	}()

	vm.RaiseError("ValueError: invalid input")
}

func TestRaiseErrorTypeError(t *testing.T) {
	vm := newAPIVM(16)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from RaiseError")
		}
		pe := r.(*PyPanicError)
		if pe.ExcType != "TypeError" {
			t.Errorf("ExcType = %q, want %q", pe.ExcType, "TypeError")
		}
	}()

	vm.RaiseError("TypeError: expected int, got str")
}

func TestRaiseErrorDefaultRuntimeError(t *testing.T) {
	vm := newAPIVM(16)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from RaiseError")
		}
		pe := r.(*PyPanicError)
		if pe.ExcType != "RuntimeError" {
			t.Errorf("ExcType = %q, want %q", pe.ExcType, "RuntimeError")
		}
		if pe.Message != "something went wrong" {
			t.Errorf("Message = %q, want %q", pe.Message, "something went wrong")
		}
	}()

	vm.RaiseError("something went wrong")
}

func TestRaiseErrorWithFormat(t *testing.T) {
	vm := newAPIVM(16)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from RaiseError")
		}
		pe := r.(*PyPanicError)
		if pe.ExcType != "IndexError" {
			t.Errorf("ExcType = %q, want %q", pe.ExcType, "IndexError")
		}
		if pe.Message != "index 5 out of range" {
			t.Errorf("Message = %q", pe.Message)
		}
	}()

	vm.RaiseError("IndexError: index %d out of range", 5)
}

// =====================================
// Type predicates: IsInt, IsFloat, IsString, etc.
// =====================================

func TestTypePredicates(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		isInt, isFloat, isString, isBool, isList, isDict, isTuple bool
	}{
		{"int", MakeInt(1), true, false, false, false, false, false, false},
		{"float", &PyFloat{Value: 1.0}, false, true, false, false, false, false, false},
		{"string", &PyString{Value: "x"}, false, false, true, false, false, false, false},
		{"bool", True, false, false, false, true, false, false, false},
		{"list", &PyList{Items: nil}, false, false, false, false, true, false, false},
		{"dict", &PyDict{}, false, false, false, false, false, true, false},
		{"tuple", &PyTuple{Items: nil}, false, false, false, false, false, false, true},
		{"None", None, false, false, false, false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsInt(tt.value) != tt.isInt {
				t.Errorf("IsInt = %v, want %v", IsInt(tt.value), tt.isInt)
			}
			if IsFloat(tt.value) != tt.isFloat {
				t.Errorf("IsFloat = %v, want %v", IsFloat(tt.value), tt.isFloat)
			}
			if IsString(tt.value) != tt.isString {
				t.Errorf("IsString = %v, want %v", IsString(tt.value), tt.isString)
			}
			if IsBool(tt.value) != tt.isBool {
				t.Errorf("IsBool = %v, want %v", IsBool(tt.value), tt.isBool)
			}
			if IsList(tt.value) != tt.isList {
				t.Errorf("IsList = %v, want %v", IsList(tt.value), tt.isList)
			}
			if IsDict(tt.value) != tt.isDict {
				t.Errorf("IsDict = %v, want %v", IsDict(tt.value), tt.isDict)
			}
			if IsTuple(tt.value) != tt.isTuple {
				t.Errorf("IsTuple = %v, want %v", IsTuple(tt.value), tt.isTuple)
			}
		})
	}
}

// =====================================
// IsCallable
// =====================================

func TestIsCallable(t *testing.T) {
	tests := []struct {
		name string
		v    Value
		want bool
	}{
		{"PyFunction", &PyFunction{Name: "f"}, true},
		{"PyBuiltinFunc", &PyBuiltinFunc{Name: "b"}, true},
		{"PyGoFunc", &PyGoFunc{Name: "g"}, true},
		{"PyClass", &PyClass{Name: "C"}, true},
		{"PyMethod", &PyMethod{Func: &PyFunction{Name: "m"}}, true},
		{"int", MakeInt(1), false},
		{"string", &PyString{Value: "x"}, false},
		{"None", None, false},
		{"list", &PyList{Items: nil}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCallable(tt.v)
			if got != tt.want {
				t.Errorf("IsCallable(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// =====================================
// Register / RegisterBuiltin / RegisterFuncs
// =====================================

func TestRegister(t *testing.T) {
	vm := NewVM()

	vm.Register("my_func", func(vm *VM) int {
		return 0
	})

	val := vm.GetGlobal("my_func")
	fn, ok := val.(*PyGoFunc)
	if !ok {
		t.Fatalf("expected *PyGoFunc, got %T", val)
	}
	if fn.Name != "my_func" {
		t.Errorf("Name = %q", fn.Name)
	}
}

func TestRegisterBuiltin(t *testing.T) {
	vm := NewVM()

	vm.RegisterBuiltin("my_builtin", func(vm *VM) int { return 0 })

	val := vm.GetBuiltin("my_builtin")
	if _, ok := val.(*PyGoFunc); !ok {
		t.Errorf("expected *PyGoFunc, got %T", val)
	}
}

func TestRegisterFuncs(t *testing.T) {
	vm := NewVM()

	vm.RegisterFuncs(map[string]GoFunction{
		"fn1": func(vm *VM) int { return 0 },
		"fn2": func(vm *VM) int { return 0 },
	})

	if _, ok := vm.GetGlobal("fn1").(*PyGoFunc); !ok {
		t.Error("fn1 not registered as global")
	}
	if _, ok := vm.GetGlobal("fn2").(*PyGoFunc); !ok {
		t.Error("fn2 not registered as global")
	}
}

// =====================================
// ToGoValue
// =====================================

func TestToGoValueBasicTypes(t *testing.T) {
	if ToGoValue(None) != nil {
		t.Error("ToGoValue(None) should be nil")
	}
	if ToGoValue(True) != true {
		t.Error("ToGoValue(True) should be true")
	}
	if ToGoValue(False) != false {
		t.Error("ToGoValue(False) should be false")
	}
	if ToGoValue(MakeInt(42)).(int64) != 42 {
		t.Error("ToGoValue(42) should be int64(42)")
	}
	if ToGoValue(&PyFloat{Value: 3.14}).(float64) != 3.14 {
		t.Error("ToGoValue(3.14) should be 3.14")
	}
	if ToGoValue(&PyString{Value: "hello"}).(string) != "hello" {
		t.Error("ToGoValue('hello') should be 'hello'")
	}
	if string(ToGoValue(&PyBytes{Value: []byte("hi")}).([]byte)) != "hi" {
		t.Error("ToGoValue(bytes) should be []byte")
	}
}

func TestToGoValueList(t *testing.T) {
	list := &PyList{Items: []Value{MakeInt(1), MakeInt(2)}}
	result := ToGoValue(list).([]any)
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	if result[0].(int64) != 1 || result[1].(int64) != 2 {
		t.Errorf("got %v, want [1, 2]", result)
	}
}

// =====================================
// FromGoValue
// =====================================

func TestFromGoValue(t *testing.T) {
	// nil -> None
	if _, ok := FromGoValue(nil).(*PyNone); !ok {
		t.Error("FromGoValue(nil) should be None")
	}

	// bool
	if v, ok := FromGoValue(true).(*PyBool); !ok || !v.Value {
		t.Error("FromGoValue(true) should be True")
	}

	// int
	if v, ok := FromGoValue(42).(*PyInt); !ok || v.Value != 42 {
		t.Error("FromGoValue(42) should be int 42")
	}

	// float
	if v, ok := FromGoValue(3.14).(*PyFloat); !ok || v.Value != 3.14 {
		t.Error("FromGoValue(3.14) should be float 3.14")
	}

	// string
	if v, ok := FromGoValue("hello").(*PyString); !ok || v.Value != "hello" {
		t.Error("FromGoValue('hello') should be string 'hello'")
	}

	// []byte -> bytes
	if v, ok := FromGoValue([]byte{1, 2}).(*PyBytes); !ok || len(v.Value) != 2 {
		t.Error("FromGoValue([]byte{1,2}) should be bytes")
	}

	// slice -> list
	if v, ok := FromGoValue([]int{1, 2, 3}).(*PyList); !ok || len(v.Items) != 3 {
		t.Error("FromGoValue([]int{1,2,3}) should be list of 3 items")
	}
}

// =====================================
// NewGoFunction
// =====================================

func TestNewGoFunction(t *testing.T) {
	fn := NewGoFunction("test", func(vm *VM) int { return 0 })
	if fn.Name != "test" {
		t.Errorf("Name = %q, want %q", fn.Name, "test")
	}
	if fn.Fn == nil {
		t.Error("Fn should not be nil")
	}
	if fn.Type() != "builtin_function_or_method" {
		t.Errorf("Type() = %q", fn.Type())
	}
}

// =====================================
// NewUserData
// =====================================

func TestNewUserData(t *testing.T) {
	ud := NewUserData("custom data")
	if ud.Value != "custom data" {
		t.Errorf("Value = %v", ud.Value)
	}
	if ud.Type() != "userdata" {
		t.Errorf("Type() = %q", ud.Type())
	}
	if IsUserData(ud) != true {
		t.Error("IsUserData should be true")
	}
	if IsUserData(MakeInt(1)) != false {
		t.Error("IsUserData on int should be false")
	}
}

// =====================================
// Truthy / Equal / CompareOp (VM methods)
// =====================================

func TestVMTruthy(t *testing.T) {
	vm := NewVM()
	if !vm.Truthy(MakeInt(1)) {
		t.Error("Truthy(1) should be true")
	}
	if vm.Truthy(MakeInt(0)) {
		t.Error("Truthy(0) should be false")
	}
	if vm.Truthy(None) {
		t.Error("Truthy(None) should be false")
	}
}

func TestVMEqual(t *testing.T) {
	vm := NewVM()
	if !vm.Equal(MakeInt(1), MakeInt(1)) {
		t.Error("Equal(1, 1) should be true")
	}
	if vm.Equal(MakeInt(1), MakeInt(2)) {
		t.Error("Equal(1, 2) should be false")
	}
	if !vm.Equal(&PyString{Value: "abc"}, &PyString{Value: "abc"}) {
		t.Error("Equal('abc', 'abc') should be true")
	}
}

func TestVMHashValue(t *testing.T) {
	vm := NewVM()
	h1 := vm.HashValue(MakeInt(42))
	h2 := vm.HashValue(MakeInt(42))
	if h1 != h2 {
		t.Error("HashValue should be consistent for same value")
	}
}

// =====================================
// ArgError / TypeError helpers
// =====================================

func TestArgError(t *testing.T) {
	vm := newAPIVM(16)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from ArgError")
		}
		pe, ok := r.(*PyPanicError)
		if !ok {
			t.Fatalf("expected *PyPanicError, got %T", r)
		}
		if pe.ExcType != "TypeError" {
			t.Errorf("ExcType = %q, want %q", pe.ExcType, "TypeError")
		}
	}()

	vm.ArgError(1, "expected integer")
}

func TestTypeErrorHelper(t *testing.T) {
	vm := newAPIVM(16)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic from TypeError")
		}
		pe, ok := r.(*PyPanicError)
		if !ok {
			t.Fatalf("expected *PyPanicError, got %T", r)
		}
		if pe.ExcType != "TypeError" {
			t.Errorf("ExcType = %q, want %q", pe.ExcType, "TypeError")
		}
	}()

	vm.TypeError("int", MakeInt(42))
}

// =====================================
// CheckUserData
// =====================================

func TestCheckUserData(t *testing.T) {
	vm := newAPIVM(16)
	ud := NewUserData(42)
	vm.Push(ud)

	got := vm.CheckUserData(1, "")
	if got != ud {
		t.Error("CheckUserData should return the userdata")
	}
}

func TestCheckUserDataNonUserData(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(MakeInt(1))

	got := vm.CheckUserData(1, "")
	if got != nil {
		t.Error("CheckUserData on non-userdata should return nil")
	}
}

// =====================================
// ToUserData
// =====================================

func TestToUserData(t *testing.T) {
	vm := newAPIVM(16)
	ud := NewUserData("data")
	vm.Push(ud)

	got := vm.ToUserData(1)
	if got != ud {
		t.Error("ToUserData should return the userdata")
	}
}

func TestToUserDataNonUserData(t *testing.T) {
	vm := newAPIVM(16)
	vm.Push(&PyString{Value: "not ud"})

	got := vm.ToUserData(1)
	if got != nil {
		t.Error("ToUserData on non-userdata should return nil")
	}
}

// =====================================
// TypeNameOf
// =====================================

func TestTypeNameOf(t *testing.T) {
	vm := NewVM()

	tests := []struct {
		v    Value
		want string
	}{
		{None, "NoneType"},
		{True, "bool"},
		{MakeInt(1), "int"},
		{&PyFloat{Value: 1.0}, "float"},
		{&PyString{Value: "x"}, "str"},
		{&PyList{Items: nil}, "list"},
		{&PyDict{}, "dict"},
		{&PyTuple{Items: nil}, "tuple"},
		{MakeComplex(1, 2), "complex"},
		{&PyBytes{Value: nil}, "bytes"},
	}

	for _, tt := range tests {
		got := vm.TypeNameOf(tt.v)
		if got != tt.want {
			t.Errorf("TypeNameOf(%v) = %q, want %q", tt.v, got, tt.want)
		}
	}
}
