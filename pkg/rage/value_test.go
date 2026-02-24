package rage

import (
	"testing"
)

// =====================================
// Value Constructor Tests
// =====================================

func TestNoneValue(t *testing.T) {
	v := None
	if v.Type() != "NoneType" {
		t.Errorf("expected type 'NoneType', got %q", v.Type())
	}
	if v.String() != "None" {
		t.Errorf("expected string 'None', got %q", v.String())
	}
	if v.GoValue() != nil {
		t.Errorf("expected GoValue nil, got %v", v.GoValue())
	}
}

func TestBoolValue(t *testing.T) {
	tr := Bool(true)
	if tr.Type() != "bool" {
		t.Errorf("expected type 'bool', got %q", tr.Type())
	}
	if tr != True {
		t.Error("Bool(true) should return True singleton")
	}

	fl := Bool(false)
	if fl != False {
		t.Error("Bool(false) should return False singleton")
	}

	bv := tr.(BoolValue)
	if !bv.Bool() {
		t.Error("True.Bool() should be true")
	}
	if bv.GoValue() != true {
		t.Error("True.GoValue() should be true")
	}
}

func TestIntValue(t *testing.T) {
	v := Int(42)
	if v.Type() != "int" {
		t.Errorf("expected type 'int', got %q", v.Type())
	}
	if v.String() != "42" {
		t.Errorf("expected string '42', got %q", v.String())
	}
	iv := v.(IntValue)
	if iv.Int() != 42 {
		t.Errorf("expected Int() = 42, got %d", iv.Int())
	}
	if iv.GoValue() != int64(42) {
		t.Errorf("expected GoValue = 42, got %v", iv.GoValue())
	}
}

func TestIntValueNegative(t *testing.T) {
	v := Int(-100)
	if v.String() != "-100" {
		t.Errorf("expected '-100', got %q", v.String())
	}
}

func TestIntValueZero(t *testing.T) {
	v := Int(0)
	iv := v.(IntValue)
	if iv.Int() != 0 {
		t.Errorf("expected 0, got %d", iv.Int())
	}
}

func TestFloatValue(t *testing.T) {
	v := Float(3.14)
	if v.Type() != "float" {
		t.Errorf("expected type 'float', got %q", v.Type())
	}
	fv := v.(FloatValue)
	if fv.Float() != 3.14 {
		t.Errorf("expected Float() = 3.14, got %f", fv.Float())
	}
	if fv.GoValue() != 3.14 {
		t.Errorf("expected GoValue = 3.14, got %v", fv.GoValue())
	}
}

func TestComplexValue(t *testing.T) {
	v := Complex(1.0, 2.0)
	if v.Type() != "complex" {
		t.Errorf("expected type 'complex', got %q", v.Type())
	}
	cv := v.(ComplexValue)
	if cv.Real() != 1.0 {
		t.Errorf("expected Real() = 1.0, got %f", cv.Real())
	}
	if cv.Imag() != 2.0 {
		t.Errorf("expected Imag() = 2.0, got %f", cv.Imag())
	}
	goVal := cv.GoValue().(complex128)
	if real(goVal) != 1.0 || imag(goVal) != 2.0 {
		t.Errorf("expected GoValue = (1+2i), got %v", goVal)
	}
}

func TestStringValue(t *testing.T) {
	v := String("hello")
	if v.Type() != "str" {
		t.Errorf("expected type 'str', got %q", v.Type())
	}
	if v.String() != "hello" {
		t.Errorf("expected 'hello', got %q", v.String())
	}
	sv := v.(StringValue)
	if sv.Str() != "hello" {
		t.Errorf("expected Str() = 'hello', got %q", sv.Str())
	}
	if sv.GoValue() != "hello" {
		t.Errorf("expected GoValue = 'hello', got %v", sv.GoValue())
	}
}

func TestStringValueEmpty(t *testing.T) {
	v := String("")
	sv := v.(StringValue)
	if sv.Str() != "" {
		t.Errorf("expected empty string, got %q", sv.Str())
	}
}

func TestListValue(t *testing.T) {
	v := List(Int(1), Int(2), Int(3))
	if v.Type() != "list" {
		t.Errorf("expected type 'list', got %q", v.Type())
	}
	lv := v.(ListValue)
	if lv.Len() != 3 {
		t.Errorf("expected Len() = 3, got %d", lv.Len())
	}

	items := lv.Items()
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	for i, want := range []int64{1, 2, 3} {
		n, ok := AsInt(items[i])
		if !ok || n != want {
			t.Errorf("item %d: expected %d, got %v", i, want, items[i])
		}
	}
}

func TestListValueGet(t *testing.T) {
	v := List(Int(10), Int(20)).(ListValue)

	val := v.Get(0)
	n, ok := AsInt(val)
	if !ok || n != 10 {
		t.Errorf("Get(0): expected 10, got %v", val)
	}

	val = v.Get(1)
	n, ok = AsInt(val)
	if !ok || n != 20 {
		t.Errorf("Get(1): expected 20, got %v", val)
	}

	// Out of bounds returns None
	val = v.Get(-1)
	if !IsNone(val) {
		t.Errorf("Get(-1): expected None, got %v", val)
	}
	val = v.Get(5)
	if !IsNone(val) {
		t.Errorf("Get(5): expected None, got %v", val)
	}
}

func TestListValueEmpty(t *testing.T) {
	v := List().(ListValue)
	if v.Len() != 0 {
		t.Errorf("expected Len() = 0, got %d", v.Len())
	}
}

func TestTupleValue(t *testing.T) {
	v := Tuple(String("a"), String("b"))
	if v.Type() != "tuple" {
		t.Errorf("expected type 'tuple', got %q", v.Type())
	}
	tv := v.(TupleValue)
	if tv.Len() != 2 {
		t.Errorf("expected Len() = 2, got %d", tv.Len())
	}

	val := tv.Get(0)
	s, ok := AsString(val)
	if !ok || s != "a" {
		t.Errorf("Get(0): expected 'a', got %v", val)
	}

	// Out of bounds
	val = tv.Get(10)
	if !IsNone(val) {
		t.Errorf("Get(10): expected None, got %v", val)
	}
}

func TestDictValue(t *testing.T) {
	v := Dict("x", Int(1), "y", Int(2))
	if v.Type() != "dict" {
		t.Errorf("expected type 'dict', got %q", v.Type())
	}
	dv := v.(DictValue)
	if dv.Len() != 2 {
		t.Errorf("expected Len() = 2, got %d", dv.Len())
	}

	items := dv.Items()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	val := dv.Get("x")
	n, ok := AsInt(val)
	if !ok || n != 1 {
		t.Errorf("Get('x'): expected 1, got %v", val)
	}

	val = dv.Get("nonexistent")
	if !IsNone(val) {
		t.Errorf("Get('nonexistent'): expected None, got %v", val)
	}
}

func TestDictValueGoConversion(t *testing.T) {
	v := Dict("key", String("value"))
	dv := v.(DictValue)
	goVal := dv.GoValue().(map[string]any)
	if goVal["key"] != "value" {
		t.Errorf("expected GoValue['key'] = 'value', got %v", goVal["key"])
	}
}

func TestDictWithNonValuePairs(t *testing.T) {
	// Dict should auto-convert non-Value pairs via FromGo
	v := Dict("count", 42, "name", "test")
	dv := v.(DictValue)

	n, ok := AsInt(dv.Get("count"))
	if !ok || n != 42 {
		t.Errorf("expected count=42, got %v", dv.Get("count"))
	}
	s, ok := AsString(dv.Get("name"))
	if !ok || s != "test" {
		t.Errorf("expected name='test', got %v", dv.Get("name"))
	}
}

func TestBytesValue(t *testing.T) {
	v := Bytes([]byte{0x48, 0x65, 0x6c, 0x6c, 0x6f})
	if v.Type() != "bytes" {
		t.Errorf("expected type 'bytes', got %q", v.Type())
	}
	bv := v.(BytesValue)
	if string(bv.Bytes()) != "Hello" {
		t.Errorf("expected 'Hello', got %q", string(bv.Bytes()))
	}
	if v.String() != "b'Hello'" {
		t.Errorf("expected \"b'Hello'\", got %q", v.String())
	}
}

func TestUserDataValue(t *testing.T) {
	type myStruct struct{ X int }
	v := UserData(&myStruct{X: 42})
	if v.Type() != "userdata" {
		t.Errorf("expected type 'userdata', got %q", v.Type())
	}
	uv := v.(UserDataValue)
	s := uv.GoValue().(*myStruct)
	if s.X != 42 {
		t.Errorf("expected X=42, got %d", s.X)
	}
}

// =====================================
// FromGo Tests
// =====================================

func TestFromGo_Nil(t *testing.T) {
	v := FromGo(nil)
	if !IsNone(v) {
		t.Errorf("expected None for nil, got %v", v)
	}
}

func TestFromGo_Value(t *testing.T) {
	original := Int(42)
	v := FromGo(original)
	if v != original {
		t.Error("FromGo(Value) should return the same Value")
	}
}

func TestFromGo_Bool(t *testing.T) {
	v := FromGo(true)
	b, ok := AsBool(v)
	if !ok || !b {
		t.Errorf("expected true, got %v", v)
	}
}

func TestFromGo_Ints(t *testing.T) {
	tests := []struct {
		input any
		want  int64
	}{
		{int(5), 5},
		{int8(8), 8},
		{int16(16), 16},
		{int32(32), 32},
		{int64(64), 64},
		{uint(10), 10},
		{uint8(8), 8},
		{uint16(16), 16},
		{uint32(32), 32},
		{uint64(64), 64},
	}
	for _, tt := range tests {
		v := FromGo(tt.input)
		n, ok := AsInt(v)
		if !ok || n != tt.want {
			t.Errorf("FromGo(%T(%v)): expected %d, got %v", tt.input, tt.input, tt.want, v)
		}
	}
}

func TestFromGo_Floats(t *testing.T) {
	v32 := FromGo(float32(1.5))
	f, ok := AsFloat(v32)
	if !ok || f != float64(float32(1.5)) {
		t.Errorf("FromGo(float32): expected 1.5, got %v", v32)
	}

	v64 := FromGo(float64(2.5))
	f, ok = AsFloat(v64)
	if !ok || f != 2.5 {
		t.Errorf("FromGo(float64): expected 2.5, got %v", v64)
	}
}

func TestFromGo_Complex(t *testing.T) {
	v := FromGo(complex(1.0, 2.0))
	r, i, ok := AsComplex(v)
	if !ok || r != 1.0 || i != 2.0 {
		t.Errorf("FromGo(complex128): expected (1+2i), got (%f+%fi)", r, i)
	}

	v32 := FromGo(complex64(complex(3.0, 4.0)))
	r, i, ok = AsComplex(v32)
	if !ok || r != 3.0 || i != 4.0 {
		t.Errorf("FromGo(complex64): expected (3+4i), got (%f+%fi)", r, i)
	}
}

func TestFromGo_String(t *testing.T) {
	v := FromGo("hello")
	s, ok := AsString(v)
	if !ok || s != "hello" {
		t.Errorf("FromGo(string): expected 'hello', got %v", v)
	}
}

func TestFromGo_Bytes(t *testing.T) {
	v := FromGo([]byte{1, 2, 3})
	bs, ok := AsBytes(v)
	if !ok || len(bs) != 3 {
		t.Errorf("FromGo([]byte): expected bytes len 3, got %v", v)
	}
}

func TestFromGo_Slice(t *testing.T) {
	v := FromGo([]any{1, "two", 3.0})
	items, ok := AsList(v)
	if !ok || len(items) != 3 {
		t.Fatalf("FromGo([]any): expected list of 3, got %v", v)
	}
	n, _ := AsInt(items[0])
	if n != 1 {
		t.Errorf("item 0: expected 1, got %v", items[0])
	}
	s, _ := AsString(items[1])
	if s != "two" {
		t.Errorf("item 1: expected 'two', got %v", items[1])
	}
	f, _ := AsFloat(items[2])
	if f != 3.0 {
		t.Errorf("item 2: expected 3.0, got %v", items[2])
	}
}

func TestFromGo_Map(t *testing.T) {
	v := FromGo(map[string]any{"a": 1, "b": "two"})
	d, ok := AsDict(v)
	if !ok || len(d) != 2 {
		t.Fatalf("FromGo(map): expected dict of 2, got %v", v)
	}
}

func TestFromGo_UnknownType(t *testing.T) {
	type custom struct{ X int }
	v := FromGo(&custom{X: 42})
	if !IsUserData(v) {
		t.Errorf("expected UserData for unknown type, got %T", v)
	}
	ud, ok := AsUserData(v)
	if !ok {
		t.Fatal("AsUserData failed")
	}
	c := ud.(*custom)
	if c.X != 42 {
		t.Errorf("expected X=42, got %d", c.X)
	}
}

// =====================================
// Type Checking Tests
// =====================================

func TestIsNone(t *testing.T) {
	if !IsNone(None) {
		t.Error("IsNone(None) should be true")
	}
	if IsNone(Int(0)) {
		t.Error("IsNone(Int(0)) should be false")
	}
}

func TestIsBool(t *testing.T) {
	if !IsBool(True) {
		t.Error("IsBool(True) should be true")
	}
	if !IsBool(False) {
		t.Error("IsBool(False) should be true")
	}
	if IsBool(Int(1)) {
		t.Error("IsBool(Int(1)) should be false")
	}
}

func TestIsInt(t *testing.T) {
	if !IsInt(Int(42)) {
		t.Error("IsInt(Int(42)) should be true")
	}
	if IsInt(Float(42.0)) {
		t.Error("IsInt(Float(42.0)) should be false")
	}
}

func TestIsFloat(t *testing.T) {
	if !IsFloat(Float(1.0)) {
		t.Error("IsFloat(Float(1.0)) should be true")
	}
	if IsFloat(Int(1)) {
		t.Error("IsFloat(Int(1)) should be false")
	}
}

func TestIsComplex(t *testing.T) {
	if !IsComplex(Complex(1, 2)) {
		t.Error("IsComplex should be true")
	}
	if IsComplex(Float(1.0)) {
		t.Error("IsComplex(Float) should be false")
	}
}

func TestIsString(t *testing.T) {
	if !IsString(String("x")) {
		t.Error("IsString should be true")
	}
	if IsString(Int(1)) {
		t.Error("IsString(Int) should be false")
	}
}

func TestIsList(t *testing.T) {
	if !IsList(List(Int(1))) {
		t.Error("IsList should be true")
	}
	if IsList(Tuple(Int(1))) {
		t.Error("IsList(Tuple) should be false")
	}
}

func TestIsTuple(t *testing.T) {
	if !IsTuple(Tuple(Int(1))) {
		t.Error("IsTuple should be true")
	}
	if IsTuple(List(Int(1))) {
		t.Error("IsTuple(List) should be false")
	}
}

func TestIsDict(t *testing.T) {
	if !IsDict(Dict("x", Int(1))) {
		t.Error("IsDict should be true")
	}
	if IsDict(List()) {
		t.Error("IsDict(List) should be false")
	}
}

func TestIsBytes(t *testing.T) {
	if !IsBytes(Bytes([]byte{1})) {
		t.Error("IsBytes should be true")
	}
	if IsBytes(String("x")) {
		t.Error("IsBytes(String) should be false")
	}
}

func TestIsUserData(t *testing.T) {
	if !IsUserData(UserData(42)) {
		t.Error("IsUserData should be true")
	}
	if IsUserData(Int(42)) {
		t.Error("IsUserData(Int) should be false")
	}
}

// =====================================
// Type Assertion Tests
// =====================================

func TestAsBool(t *testing.T) {
	b, ok := AsBool(True)
	if !ok || !b {
		t.Error("AsBool(True) should return (true, true)")
	}
	b, ok = AsBool(False)
	if !ok || b {
		t.Error("AsBool(False) should return (false, true)")
	}
	_, ok = AsBool(Int(1))
	if ok {
		t.Error("AsBool(Int) should return (_, false)")
	}
}

func TestAsInt(t *testing.T) {
	n, ok := AsInt(Int(42))
	if !ok || n != 42 {
		t.Errorf("AsInt(Int(42)): expected (42, true), got (%d, %v)", n, ok)
	}
	_, ok = AsInt(String("42"))
	if ok {
		t.Error("AsInt(String) should return (_, false)")
	}
}

func TestAsFloat(t *testing.T) {
	f, ok := AsFloat(Float(1.5))
	if !ok || f != 1.5 {
		t.Errorf("AsFloat(Float(1.5)): expected (1.5, true), got (%f, %v)", f, ok)
	}
	// AsFloat also accepts Int
	f, ok = AsFloat(Int(10))
	if !ok || f != 10.0 {
		t.Errorf("AsFloat(Int(10)): expected (10.0, true), got (%f, %v)", f, ok)
	}
	_, ok = AsFloat(String("1.0"))
	if ok {
		t.Error("AsFloat(String) should return (_, false)")
	}
}

func TestAsComplex(t *testing.T) {
	r, i, ok := AsComplex(Complex(1, 2))
	if !ok || r != 1.0 || i != 2.0 {
		t.Errorf("AsComplex: expected (1, 2, true), got (%f, %f, %v)", r, i, ok)
	}
	_, _, ok = AsComplex(Float(1.0))
	if ok {
		t.Error("AsComplex(Float) should return false")
	}
}

func TestAsString(t *testing.T) {
	s, ok := AsString(String("hello"))
	if !ok || s != "hello" {
		t.Errorf("AsString: expected ('hello', true), got (%q, %v)", s, ok)
	}
	_, ok = AsString(Int(1))
	if ok {
		t.Error("AsString(Int) should return false")
	}
}

func TestAsList(t *testing.T) {
	items, ok := AsList(List(Int(1), Int(2)))
	if !ok || len(items) != 2 {
		t.Error("AsList failed")
	}
	_, ok = AsList(Tuple(Int(1)))
	if ok {
		t.Error("AsList(Tuple) should return false")
	}
}

func TestAsTuple(t *testing.T) {
	items, ok := AsTuple(Tuple(Int(1), Int(2)))
	if !ok || len(items) != 2 {
		t.Error("AsTuple failed")
	}
	_, ok = AsTuple(List(Int(1)))
	if ok {
		t.Error("AsTuple(List) should return false")
	}
}

func TestAsDict(t *testing.T) {
	d, ok := AsDict(Dict("a", Int(1)))
	if !ok || len(d) != 1 {
		t.Error("AsDict failed")
	}
	_, ok = AsDict(List())
	if ok {
		t.Error("AsDict(List) should return false")
	}
}

func TestAsBytes(t *testing.T) {
	bs, ok := AsBytes(Bytes([]byte{1, 2}))
	if !ok || len(bs) != 2 {
		t.Error("AsBytes failed")
	}
	_, ok = AsBytes(String("x"))
	if ok {
		t.Error("AsBytes(String) should return false")
	}
}

func TestAsUserData(t *testing.T) {
	val, ok := AsUserData(UserData(42))
	if !ok || val != 42 {
		t.Error("AsUserData failed")
	}
	_, ok = AsUserData(Int(42))
	if ok {
		t.Error("AsUserData(Int) should return false")
	}
}

// =====================================
// toRuntime / fromRuntime Round-trip Tests
// =====================================

func TestToRuntimeNil(t *testing.T) {
	rv := toRuntime(nil)
	if rv == nil {
		t.Error("toRuntime(nil) should return runtime.None, not nil")
	}
}

func TestFromRuntimeNil(t *testing.T) {
	v := fromRuntime(nil)
	if !IsNone(v) {
		t.Errorf("fromRuntime(nil) should return None, got %v", v)
	}
}

func TestValueRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		val  Value
	}{
		{"None", None},
		{"True", True},
		{"False", False},
		{"Int", Int(42)},
		{"Float", Float(3.14)},
		{"String", String("hello")},
		{"Bytes", Bytes([]byte{1, 2, 3})},
		{"Complex", Complex(1, 2)},
		{"EmptyList", List()},
		{"List", List(Int(1), Int(2))},
		{"EmptyTuple", Tuple()},
		{"Tuple", Tuple(String("a"), String("b"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rv := toRuntime(tt.val)
			back := fromRuntime(rv)
			if back.Type() != tt.val.Type() {
				t.Errorf("round-trip type mismatch: %q != %q", back.Type(), tt.val.Type())
			}
		})
	}
}

// =====================================
// FunctionValue Tests
// =====================================

func TestFunctionValue(t *testing.T) {
	state := NewState()
	defer state.Close()

	_, err := state.Run(`
def my_func():
    pass
`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fn := state.GetGlobal("my_func")
	if fn.Type() != "function" {
		t.Errorf("expected type 'function', got %q", fn.Type())
	}
	fv, ok := fn.(FunctionValue)
	if !ok {
		t.Fatal("expected FunctionValue")
	}
	if fv.Name() != "my_func" {
		t.Errorf("expected name 'my_func', got %q", fv.Name())
	}
	if fv.GoValue() != nil {
		t.Error("FunctionValue.GoValue() should be nil")
	}
}

// =====================================
// GoValue for collections
// =====================================

func TestListGoValue(t *testing.T) {
	v := List(Int(1), String("two"), Float(3.0))
	goVal := v.GoValue().([]any)
	if len(goVal) != 3 {
		t.Fatalf("expected 3 items, got %d", len(goVal))
	}
	if goVal[0] != int64(1) {
		t.Errorf("item 0: expected int64(1), got %T(%v)", goVal[0], goVal[0])
	}
	if goVal[1] != "two" {
		t.Errorf("item 1: expected 'two', got %v", goVal[1])
	}
	if goVal[2] != 3.0 {
		t.Errorf("item 2: expected 3.0, got %v", goVal[2])
	}
}

func TestTupleGoValue(t *testing.T) {
	v := Tuple(Bool(true), None)
	goVal := v.GoValue().([]any)
	if len(goVal) != 2 {
		t.Fatalf("expected 2 items, got %d", len(goVal))
	}
	if goVal[0] != true {
		t.Errorf("item 0: expected true, got %v", goVal[0])
	}
	if goVal[1] != nil {
		t.Errorf("item 1: expected nil, got %v", goVal[1])
	}
}
