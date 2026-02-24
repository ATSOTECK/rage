package runtime

import (
	"math"
	"strings"
	"testing"
)

// =====================================
// toList: extended conversions
// =====================================

func TestToListDict(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	d.DictSet(&PyString{Value: "a"}, MakeInt(1), vm)
	d.DictSet(&PyString{Value: "b"}, MakeInt(2), vm)

	items, err := vm.toList(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	// Dict toList returns keys
	found := map[string]bool{}
	for _, item := range items {
		s, ok := item.(*PyString)
		if !ok {
			t.Fatalf("expected *PyString key, got %T", item)
		}
		found[s.Value] = true
	}
	if !found["a"] || !found["b"] {
		t.Error("expected keys 'a' and 'b'")
	}
}

func TestToListSet(t *testing.T) {
	vm := NewVM()
	s := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	s.SetAdd(MakeInt(10), vm)
	s.SetAdd(MakeInt(20), vm)

	items, err := vm.toList(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestToListFrozenSet(t *testing.T) {
	vm := NewVM()
	fs := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	fs.FrozenSetAdd(MakeInt(5), vm)
	fs.FrozenSetAdd(MakeInt(15), vm)

	items, err := vm.toList(fs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestToListRangeNegativeStep(t *testing.T) {
	vm := NewVM()
	r := &PyRange{Start: 10, Stop: 0, Step: -3}
	items, err := vm.toList(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 10, 7, 4, 1
	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(items))
	}
	expected := []int64{10, 7, 4, 1}
	for i, want := range expected {
		got := items[i].(*PyInt).Value
		if got != want {
			t.Errorf("item[%d] = %d, want %d", i, got, want)
		}
	}
}

func TestToListEmptyRange(t *testing.T) {
	vm := NewVM()
	r := &PyRange{Start: 5, Stop: 5, Step: 1}
	items, err := vm.toList(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestToListStringUnicode(t *testing.T) {
	vm := NewVM()
	items, err := vm.toList(&PyString{Value: "hi\u00e9"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "hie\u0301" = 3 runes: 'h', 'i', '\u00e9'
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[2].(*PyString).Value != "\u00e9" {
		t.Errorf("expected third char to be e-acute, got %q", items[2].(*PyString).Value)
	}
}

func TestToListEmptyString(t *testing.T) {
	vm := NewVM()
	items, err := vm.toList(&PyString{Value: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestToListNoneNotIterable(t *testing.T) {
	vm := NewVM()
	_, err := vm.toList(None)
	if err == nil {
		t.Fatal("expected error for None")
	}
	if !strings.Contains(err.Error(), "not iterable") {
		t.Errorf("expected 'not iterable' error, got: %v", err)
	}
}

func TestToListFloatNotIterable(t *testing.T) {
	vm := NewVM()
	_, err := vm.toList(&PyFloat{Value: 3.14})
	if err == nil {
		t.Fatal("expected error for float")
	}
	if !strings.Contains(err.Error(), "not iterable") {
		t.Errorf("expected 'not iterable' error, got: %v", err)
	}
}

func TestToListBoolNotIterable(t *testing.T) {
	vm := NewVM()
	_, err := vm.toList(True)
	if err == nil {
		t.Fatal("expected error for bool")
	}
	if !strings.Contains(err.Error(), "not iterable") {
		t.Errorf("expected 'not iterable' error, got: %v", err)
	}
}

func TestToListIterator(t *testing.T) {
	vm := NewVM()
	iter := &PyIterator{
		Items: []Value{MakeInt(10), MakeInt(20), MakeInt(30)},
		Index: 1,
	}
	items, err := vm.toList(iter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Index is 1, so we get items[1:]
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].(*PyInt).Value != 20 {
		t.Errorf("expected first item to be 20, got %d", items[0].(*PyInt).Value)
	}
}

func TestToListEmptyBytes(t *testing.T) {
	vm := NewVM()
	items, err := vm.toList(&PyBytes{Value: []byte{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestToListEmptyList(t *testing.T) {
	vm := NewVM()
	items, err := vm.toList(&PyList{Items: []Value{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestToListEmptyTuple(t *testing.T) {
	vm := NewVM()
	items, err := vm.toList(&PyTuple{Items: []Value{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestToListEmptyDict(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	items, err := vm.toList(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

// =====================================
// truthy: extended edge cases
// =====================================

func TestTruthyFrozenSet(t *testing.T) {
	vm := NewVM()
	empty := &PyFrozenSet{Items: make(map[Value]struct{})}
	if vm.truthy(empty) {
		t.Error("empty frozenset should be falsy")
	}
	nonEmpty := &PyFrozenSet{Items: map[Value]struct{}{MakeInt(1): {}}}
	if !vm.truthy(nonEmpty) {
		t.Error("non-empty frozenset should be truthy")
	}
}

func TestTruthyNonEmptyDict(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	d.DictSet(&PyString{Value: "x"}, MakeInt(1), vm)
	if !vm.truthy(d) {
		t.Error("non-empty dict should be truthy")
	}
}

func TestTruthyNonEmptySet(t *testing.T) {
	vm := NewVM()
	s := &PySet{Items: map[Value]struct{}{MakeInt(1): {}}}
	if !vm.truthy(s) {
		t.Error("non-empty set should be truthy")
	}
}

func TestTruthyNegativeFloat(t *testing.T) {
	vm := NewVM()
	if !vm.truthy(&PyFloat{Value: -0.1}) {
		t.Error("negative non-zero float should be truthy")
	}
}

func TestTruthyComplexBothParts(t *testing.T) {
	vm := NewVM()
	if !vm.truthy(MakeComplex(1, 1)) {
		t.Error("complex(1,1) should be truthy")
	}
}

func TestTruthyInstance(t *testing.T) {
	// Instance with no __bool__ or __len__ should default to true
	vm := NewVM()
	cls := &PyClass{
		Name: "Foo",
		Dict: map[string]Value{},
		Mro:  []*PyClass{},
	}
	cls.Mro = []*PyClass{cls}
	inst := &PyInstance{Class: cls, Dict: map[string]Value{}}
	if !vm.truthy(inst) {
		t.Error("instance with no __bool__/__len__ should be truthy")
	}
}

func TestTruthyFunctionIsTrue(t *testing.T) {
	vm := NewVM()
	// Functions and other objects default to true
	fn := &PyFunction{Name: "f"}
	if !vm.truthy(fn) {
		t.Error("function should be truthy")
	}
}

func TestTruthyBuiltinFuncIsTrue(t *testing.T) {
	vm := NewVM()
	bf := &PyBuiltinFunc{Name: "len"}
	if !vm.truthy(bf) {
		t.Error("builtin function should be truthy")
	}
}

func TestTruthyRangeNegativeStep(t *testing.T) {
	vm := NewVM()
	// range(10, 0, -2) has elements: 10, 8, 6, 4, 2 => truthy
	r := &PyRange{Start: 10, Stop: 0, Step: -2}
	if !vm.truthy(r) {
		t.Error("non-empty range (negative step) should be truthy")
	}
	// range(0, 10, -1) is empty => falsy
	r2 := &PyRange{Start: 0, Stop: 10, Step: -1}
	if vm.truthy(r2) {
		t.Error("empty range (negative step) should be falsy")
	}
}

// =====================================
// str: extended types
// =====================================

func TestStrDict(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	d.DictSet(&PyString{Value: "a"}, MakeInt(1), vm)
	got := vm.str(d)
	if got != "{'a': 1}" {
		t.Errorf("str(dict) = %q, want %q", got, "{'a': 1}")
	}
}

func TestStrEmptyDict(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	got := vm.str(d)
	if got != "{}" {
		t.Errorf("str(empty dict) = %q, want %q", got, "{}")
	}
}

func TestStrEmptySet(t *testing.T) {
	vm := NewVM()
	s := &PySet{Items: make(map[Value]struct{})}
	got := vm.str(s)
	if got != "set()" {
		t.Errorf("str(empty set) = %q, want %q", got, "set()")
	}
}

func TestStrNonEmptySet(t *testing.T) {
	vm := NewVM()
	s := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	s.SetAdd(MakeInt(42), vm)
	got := vm.str(s)
	if got != "{42}" {
		t.Errorf("str(set with 42) = %q, want %q", got, "{42}")
	}
}

func TestStrEmptyFrozenSet(t *testing.T) {
	vm := NewVM()
	fs := &PyFrozenSet{Items: make(map[Value]struct{})}
	got := vm.str(fs)
	if got != "frozenset()" {
		t.Errorf("str(empty frozenset) = %q, want %q", got, "frozenset()")
	}
}

func TestStrNonEmptyFrozenSet(t *testing.T) {
	vm := NewVM()
	fs := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	fs.FrozenSetAdd(MakeInt(7), vm)
	got := vm.str(fs)
	if got != "frozenset({7})" {
		t.Errorf("str(frozenset with 7) = %q, want %q", got, "frozenset({7})")
	}
}

func TestStrBytes(t *testing.T) {
	vm := NewVM()
	b := &PyBytes{Value: []byte("hello")}
	got := vm.str(b)
	if got != "b'hello'" {
		t.Errorf("str(bytes) = %q, want %q", got, "b'hello'")
	}
}

func TestStrBytesSpecialChars(t *testing.T) {
	vm := NewVM()
	b := &PyBytes{Value: []byte{'\t', '\n', '\r', '\\'}}
	got := vm.str(b)
	if got != `b'\t\n\r\\'` {
		t.Errorf("str(bytes with special chars) = %q, want %q", got, `b'\t\n\r\\'`)
	}
}

func TestStrBytesNonASCII(t *testing.T) {
	vm := NewVM()
	b := &PyBytes{Value: []byte{0x00, 0x80, 0xFF}}
	got := vm.str(b)
	if got != `b'\x00\x80\xff'` {
		t.Errorf("str(bytes non-ascii) = %q, want %q", got, `b'\x00\x80\xff'`)
	}
}

func TestStrClass(t *testing.T) {
	vm := NewVM()
	cls := &PyClass{Name: "MyClass", Dict: map[string]Value{}}
	got := vm.str(cls)
	if got != "<class 'MyClass'>" {
		t.Errorf("str(class) = %q, want %q", got, "<class 'MyClass'>")
	}
}

func TestStrInstance(t *testing.T) {
	vm := NewVM()
	cls := &PyClass{
		Name: "Point",
		Dict: map[string]Value{},
		Mro:  []*PyClass{},
	}
	cls.Mro = []*PyClass{cls}
	inst := &PyInstance{Class: cls, Dict: map[string]Value{}}
	got := vm.str(inst)
	if got != "<Point object>" {
		t.Errorf("str(instance) = %q, want %q", got, "<Point object>")
	}
}

func TestStrModule(t *testing.T) {
	vm := NewVM()
	mod := &PyModule{Name: "math"}
	got := vm.str(mod)
	if got != "<module 'math'>" {
		t.Errorf("str(module) = %q, want %q", got, "<module 'math'>")
	}
}

func TestStrGoFunc(t *testing.T) {
	vm := NewVM()
	gf := &PyGoFunc{Name: "myfunc"}
	got := vm.str(gf)
	if got != "<go function myfunc>" {
		t.Errorf("str(gofunc) = %q, want %q", got, "<go function myfunc>")
	}
}

func TestStrNotImplemented(t *testing.T) {
	vm := NewVM()
	got := vm.str(NotImplemented)
	if got != "NotImplemented" {
		t.Errorf("str(NotImplemented) = %q, want %q", got, "NotImplemented")
	}
}

func TestStrFloatScientific(t *testing.T) {
	vm := NewVM()
	// Very large float should use scientific notation
	got := vm.str(&PyFloat{Value: 1e100})
	if got != "1e+100" {
		t.Errorf("str(1e100) = %q, want %q", got, "1e+100")
	}
}

func TestStrFloatNegativeZero(t *testing.T) {
	vm := NewVM()
	got := vm.str(&PyFloat{Value: -0.0})
	// Go's fmt.Sprintf("%g", -0.0) outputs "0.0" (no negative sign)
	// This differs from CPython which prints "-0.0"
	if got != "0.0" && got != "-0.0" {
		t.Errorf("str(-0.0) = %q, want %q or %q", got, "0.0", "-0.0")
	}
}

func TestStrListWithStrings(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{&PyString{Value: "x"}, &PyString{Value: "y"}}}
	got := vm.str(list)
	if got != "['x', 'y']" {
		t.Errorf("str(list of strings) = %q, want %q", got, "['x', 'y']")
	}
}

func TestStrTupleMultiple(t *testing.T) {
	vm := NewVM()
	tup := &PyTuple{Items: []Value{MakeInt(1), &PyString{Value: "a"}, True}}
	got := vm.str(tup)
	if got != "(1, 'a', True)" {
		t.Errorf("str(tuple) = %q, want %q", got, "(1, 'a', True)")
	}
}

func TestStrComplexZeroReal(t *testing.T) {
	vm := NewVM()
	got := vm.str(MakeComplex(0, 0))
	if got != "0j" {
		t.Errorf("str(0+0j) = %q, want %q", got, "0j")
	}
}

func TestStrUserData(t *testing.T) {
	vm := NewVM()
	ud := &PyUserData{Value: 42}
	got := vm.str(ud)
	if !strings.Contains(got, "userdata") {
		t.Errorf("str(userdata) = %q, expected to contain 'userdata'", got)
	}
}

// =====================================
// repr: extended types
// =====================================

func TestReprComplexNumber(t *testing.T) {
	vm := NewVM()
	got := vm.repr(MakeComplex(3, -4))
	if got != "(3-4j)" {
		t.Errorf("repr(3-4j) = %q, want %q", got, "(3-4j)")
	}
}

func TestReprBytes(t *testing.T) {
	vm := NewVM()
	got := vm.repr(&PyBytes{Value: []byte("abc")})
	if got != "b'abc'" {
		t.Errorf("repr(bytes) = %q, want %q", got, "b'abc'")
	}
}

func TestReprEmptyList(t *testing.T) {
	vm := NewVM()
	got := vm.repr(&PyList{Items: []Value{}})
	if got != "[]" {
		t.Errorf("repr(empty list) = %q, want %q", got, "[]")
	}
}

func TestReprEmptyTuple(t *testing.T) {
	vm := NewVM()
	got := vm.repr(&PyTuple{Items: []Value{}})
	if got != "()" {
		t.Errorf("repr(empty tuple) = %q, want %q", got, "()")
	}
}

func TestReprDict(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	d.DictSet(MakeInt(1), &PyString{Value: "one"}, vm)
	got := vm.repr(d)
	if got != "{1: 'one'}" {
		t.Errorf("repr(dict) = %q, want %q", got, "{1: 'one'}")
	}
}

func TestReprEmptyDict(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	got := vm.repr(d)
	if got != "{}" {
		t.Errorf("repr(empty dict) = %q, want %q", got, "{}")
	}
}

func TestReprEmptySet(t *testing.T) {
	vm := NewVM()
	s := &PySet{Items: make(map[Value]struct{})}
	got := vm.repr(s)
	if got != "set()" {
		t.Errorf("repr(empty set) = %q, want %q", got, "set()")
	}
}

func TestReprNonEmptySet(t *testing.T) {
	vm := NewVM()
	s := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	s.SetAdd(MakeInt(99), vm)
	got := vm.repr(s)
	if got != "{99}" {
		t.Errorf("repr(set with 99) = %q, want %q", got, "{99}")
	}
}

func TestReprInstance(t *testing.T) {
	vm := NewVM()
	cls := &PyClass{
		Name: "Obj",
		Dict: map[string]Value{},
		Mro:  []*PyClass{},
	}
	cls.Mro = []*PyClass{cls}
	inst := &PyInstance{Class: cls, Dict: map[string]Value{}}
	got := vm.repr(inst)
	if got != "<Obj object>" {
		t.Errorf("repr(instance) = %q, want %q", got, "<Obj object>")
	}
}

func TestReprNestedList(t *testing.T) {
	vm := NewVM()
	inner := &PyList{Items: []Value{MakeInt(1), MakeInt(2)}}
	outer := &PyList{Items: []Value{inner, MakeInt(3)}}
	got := vm.repr(outer)
	if got != "[[1, 2], 3]" {
		t.Errorf("repr(nested list) = %q, want %q", got, "[[1, 2], 3]")
	}
}

func TestReprNestedTuple(t *testing.T) {
	vm := NewVM()
	inner := &PyTuple{Items: []Value{MakeInt(1)}}
	outer := &PyTuple{Items: []Value{inner}}
	got := vm.repr(outer)
	if got != "((1,),)" {
		t.Errorf("repr(nested tuple) = %q, want %q", got, "((1,),)")
	}
}

// =====================================
// typeName: extended types
// =====================================

func TestTypeNameFrozenSet(t *testing.T) {
	vm := NewVM()
	got := vm.typeName(&PyFrozenSet{})
	if got != "frozenset" {
		t.Errorf("typeName(frozenset) = %q, want %q", got, "frozenset")
	}
}

func TestTypeNameIterator(t *testing.T) {
	vm := NewVM()
	got := vm.typeName(&PyIterator{})
	if got != "iterator" {
		t.Errorf("typeName(iterator) = %q, want %q", got, "iterator")
	}
}

func TestTypeNameGoFunc(t *testing.T) {
	vm := NewVM()
	got := vm.typeName(&PyGoFunc{Name: "gf"})
	if got != "builtin_function_or_method" {
		t.Errorf("typeName(gofunc) = %q, want %q", got, "builtin_function_or_method")
	}
}

func TestTypeNameClass(t *testing.T) {
	vm := NewVM()
	got := vm.typeName(&PyClass{Name: "Foo"})
	if got != "type" {
		t.Errorf("typeName(class) = %q, want %q", got, "type")
	}
}

func TestTypeNameInstance(t *testing.T) {
	vm := NewVM()
	cls := &PyClass{Name: "Bar"}
	inst := &PyInstance{Class: cls}
	got := vm.typeName(inst)
	if got != "Bar" {
		t.Errorf("typeName(instance) = %q, want %q", got, "Bar")
	}
}

func TestTypeNameModule(t *testing.T) {
	vm := NewVM()
	got := vm.typeName(&PyModule{Name: "os"})
	if got != "module" {
		t.Errorf("typeName(module) = %q, want %q", got, "module")
	}
}

func TestTypeNameUserData(t *testing.T) {
	vm := NewVM()
	got := vm.typeName(&PyUserData{Value: 42})
	if got != "userdata" {
		t.Errorf("typeName(userdata) = %q, want %q", got, "userdata")
	}
}

func TestTypeNameNotImplemented(t *testing.T) {
	vm := NewVM()
	got := vm.typeName(NotImplemented)
	if got != "NotImplementedType" {
		t.Errorf("typeName(NotImplemented) = %q, want %q", got, "NotImplementedType")
	}
}

// =====================================
// tryToInt: extended edge cases
// =====================================

func TestTryToIntWhitespaceOnlyString(t *testing.T) {
	vm := NewVM()
	_, err := vm.tryToInt(&PyString{Value: "   "})
	if err == nil {
		t.Fatal("expected error for whitespace-only string")
	}
	if !strings.Contains(err.Error(), "ValueError") {
		t.Errorf("expected ValueError, got: %v", err)
	}
}

func TestTryToIntListError(t *testing.T) {
	vm := NewVM()
	_, err := vm.tryToInt(&PyList{Items: []Value{}})
	if err == nil {
		t.Fatal("expected error for list")
	}
	if !strings.Contains(err.Error(), "TypeError") {
		t.Errorf("expected TypeError, got: %v", err)
	}
}

// =====================================
// tryToFloat: extended edge cases
// =====================================

func TestTryToFloatUnderscoreString(t *testing.T) {
	vm := NewVM()
	got, err := vm.tryToFloat(&PyString{Value: "1_000.5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(got-1000.5) > 1e-10 {
		t.Errorf("got %f, want 1000.5", got)
	}
}

func TestTryToFloatPositiveInfString(t *testing.T) {
	vm := NewVM()
	got, err := vm.tryToFloat(&PyString{Value: "+inf"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !math.IsInf(got, 1) {
		t.Errorf("expected +Inf, got %f", got)
	}
}

func TestTryToFloatNegInfinityString(t *testing.T) {
	vm := NewVM()
	got, err := vm.tryToFloat(&PyString{Value: "-infinity"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !math.IsInf(got, -1) {
		t.Errorf("expected -Inf, got %f", got)
	}
}

func TestTryToFloatPositiveInfinityString(t *testing.T) {
	vm := NewVM()
	got, err := vm.tryToFloat(&PyString{Value: "+infinity"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !math.IsInf(got, 1) {
		t.Errorf("expected +Inf, got %f", got)
	}
}

func TestTryToFloatListError(t *testing.T) {
	vm := NewVM()
	_, err := vm.tryToFloat(&PyList{Items: []Value{}})
	if err == nil {
		t.Fatal("expected error for list")
	}
	if !strings.Contains(err.Error(), "TypeError") {
		t.Errorf("expected TypeError, got: %v", err)
	}
}

// =====================================
// intFromStringBase: extended edge cases
// =====================================

func TestIntFromStringBaseAutoDetect(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		s    string
		base int64
		want int64
	}{
		{"base0 hex uppercase", "0XFF", 0, 255},
		{"base0 oct uppercase", "0O10", 0, 8},
		{"base0 bin uppercase", "0B10", 0, 2},
		{"base8 uppercase prefix", "0O77", 8, 63},
		{"base16 uppercase prefix", "0XFF", 16, 255},
		{"base2 uppercase prefix", "0B1010", 2, 10},
		{"positive sign", "+42", 10, 42},
		{"base0 single zero", "0", 0, 0},
		{"base0 multi zeros", "000", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.intFromStringBase(tt.s, tt.base)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, ok := result.(*PyInt)
			if !ok {
				t.Fatalf("expected *PyInt, got %T", result)
			}
			if got.Value != tt.want {
				t.Errorf("got %d, want %d", got.Value, tt.want)
			}
		})
	}
}

func TestIntFromStringBaseErrors(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		s    string
		base int64
	}{
		{"empty after sign", "-", 10},
		{"empty after prefix", "0x", 16},
		{"base0 leading zeros mixed", "0123", 0},
		{"base 0 with just sign", "+", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := vm.intFromStringBase(tt.s, tt.base)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestIntFromStringBaseNegative(t *testing.T) {
	vm := NewVM()
	result, err := vm.intFromStringBase("-ff", 16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyInt)
	if !ok {
		t.Fatalf("expected *PyInt, got %T", result)
	}
	if got.Value != -255 {
		t.Errorf("got %d, want -255", got.Value)
	}
}

func TestIntFromStringBaseUnderscores(t *testing.T) {
	vm := NewVM()
	result, err := vm.intFromStringBase("1_0_0", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyInt).Value
	if got != 100 {
		t.Errorf("got %d, want 100", got)
	}
}

// =====================================
// tryToIntValue: extended
// =====================================

func TestTryToIntValueFromInt(t *testing.T) {
	vm := NewVM()
	result, err := vm.tryToIntValue(MakeInt(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyInt)
	if !ok {
		t.Fatalf("expected *PyInt, got %T", result)
	}
	if got.Value != 42 {
		t.Errorf("got %d, want 42", got.Value)
	}
}

func TestTryToIntValueFromString(t *testing.T) {
	vm := NewVM()
	result, err := vm.tryToIntValue(&PyString{Value: "  100  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyInt)
	if !ok {
		t.Fatalf("expected *PyInt, got %T", result)
	}
	if got.Value != 100 {
		t.Errorf("got %d, want 100", got.Value)
	}
}

func TestTryToIntValueFromFloat(t *testing.T) {
	vm := NewVM()
	result, err := vm.tryToIntValue(&PyFloat{Value: 9.9})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyInt)
	if !ok {
		t.Fatalf("expected *PyInt, got %T", result)
	}
	if got.Value != 9 {
		t.Errorf("got %d, want 9", got.Value)
	}
}

func TestTryToIntValueFromBool(t *testing.T) {
	vm := NewVM()
	result, err := vm.tryToIntValue(True)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyInt)
	if !ok {
		t.Fatalf("expected *PyInt, got %T", result)
	}
	if got.Value != 1 {
		t.Errorf("got %d, want 1", got.Value)
	}
}

func TestTryToIntValueInvalidString(t *testing.T) {
	vm := NewVM()
	_, err := vm.tryToIntValue(&PyString{Value: "abc"})
	if err == nil {
		t.Fatal("expected error for invalid string")
	}
}

func TestTryToIntValueEmptyString(t *testing.T) {
	vm := NewVM()
	_, err := vm.tryToIntValue(&PyString{Value: ""})
	if err == nil {
		t.Fatal("expected error for empty string")
	}
}

// =====================================
// getIntIndex: extended
// =====================================

func TestGetIntIndexNoneError(t *testing.T) {
	vm := NewVM()
	_, err := vm.getIntIndex(None)
	if err == nil {
		t.Fatal("expected error for None")
	}
	if !strings.Contains(err.Error(), "cannot be interpreted as an integer") {
		t.Errorf("expected type error, got: %v", err)
	}
}

func TestGetIntIndexListError(t *testing.T) {
	vm := NewVM()
	_, err := vm.getIntIndex(&PyList{Items: []Value{}})
	if err == nil {
		t.Fatal("expected error for list")
	}
}

// =====================================
// toValue: various Go types
// =====================================

func TestToValueNil(t *testing.T) {
	vm := NewVM()
	got := vm.toValue(nil)
	if got != None {
		t.Errorf("toValue(nil) should be None, got %T", got)
	}
}

func TestToValueBool(t *testing.T) {
	vm := NewVM()
	got := vm.toValue(true)
	if got != True {
		t.Errorf("toValue(true) should be True")
	}
	got = vm.toValue(false)
	if got != False {
		t.Errorf("toValue(false) should be False")
	}
}

func TestToValueInt(t *testing.T) {
	vm := NewVM()
	got := vm.toValue(42)
	pyInt, ok := got.(*PyInt)
	if !ok {
		t.Fatalf("expected *PyInt, got %T", got)
	}
	if pyInt.Value != 42 {
		t.Errorf("got %d, want 42", pyInt.Value)
	}
}

func TestToValueInt64(t *testing.T) {
	vm := NewVM()
	got := vm.toValue(int64(100))
	pyInt, ok := got.(*PyInt)
	if !ok {
		t.Fatalf("expected *PyInt, got %T", got)
	}
	if pyInt.Value != 100 {
		t.Errorf("got %d, want 100", pyInt.Value)
	}
}

func TestToValueFloat64(t *testing.T) {
	vm := NewVM()
	got := vm.toValue(3.14)
	pyFloat, ok := got.(*PyFloat)
	if !ok {
		t.Fatalf("expected *PyFloat, got %T", got)
	}
	if math.Abs(pyFloat.Value-3.14) > 1e-10 {
		t.Errorf("got %f, want 3.14", pyFloat.Value)
	}
}

func TestToValueString(t *testing.T) {
	vm := NewVM()
	got := vm.toValue("hello")
	pyStr, ok := got.(*PyString)
	if !ok {
		t.Fatalf("expected *PyString, got %T", got)
	}
	if pyStr.Value != "hello" {
		t.Errorf("got %q, want %q", pyStr.Value, "hello")
	}
}

func TestToValueByteSlice(t *testing.T) {
	vm := NewVM()
	got := vm.toValue([]byte{1, 2, 3})
	pyBytes, ok := got.(*PyBytes)
	if !ok {
		t.Fatalf("expected *PyBytes, got %T", got)
	}
	if len(pyBytes.Value) != 3 {
		t.Errorf("expected 3 bytes, got %d", len(pyBytes.Value))
	}
}

func TestToValueStringSlice(t *testing.T) {
	vm := NewVM()
	got := vm.toValue([]string{"a", "b", "c"})
	pyTup, ok := got.(*PyTuple)
	if !ok {
		t.Fatalf("expected *PyTuple, got %T", got)
	}
	if len(pyTup.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(pyTup.Items))
	}
	if pyTup.Items[0].(*PyString).Value != "a" {
		t.Errorf("first item should be 'a'")
	}
}

func TestToValuePyValue(t *testing.T) {
	vm := NewVM()
	original := MakeInt(99)
	got := vm.toValue(original)
	if got != original {
		t.Error("toValue(PyInt) should return same value")
	}
}

// =====================================
// ascii: coverage
// =====================================

func TestAsciiString(t *testing.T) {
	vm := NewVM()
	got := vm.ascii(&PyString{Value: "hello"})
	if got != "'hello'" {
		t.Errorf("ascii('hello') = %q, want %q", got, "'hello'")
	}
}

func TestAsciiNonASCIIString(t *testing.T) {
	vm := NewVM()
	got := vm.ascii(&PyString{Value: "\u00e9"})
	// \u00e9 is e-acute, should be escaped as \xe9
	if got != "'\\xe9'" {
		t.Errorf("ascii(e-acute) = %q, want %q", got, "'\\xe9'")
	}
}

func TestAsciiList(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{&PyString{Value: "\u00e9"}}}
	got := vm.ascii(list)
	if got != "['\\xe9']" {
		t.Errorf("ascii(list) = %q, want %q", got, "['\\xe9']")
	}
}

func TestAsciiTuple(t *testing.T) {
	vm := NewVM()
	tup := &PyTuple{Items: []Value{&PyString{Value: "a"}, &PyString{Value: "b"}}}
	got := vm.ascii(tup)
	if got != "('a', 'b')" {
		t.Errorf("ascii(tuple) = %q, want %q", got, "('a', 'b')")
	}
}

func TestAsciiSingletonTuple(t *testing.T) {
	vm := NewVM()
	tup := &PyTuple{Items: []Value{&PyString{Value: "a"}}}
	got := vm.ascii(tup)
	if got != "('a',)" {
		t.Errorf("ascii(singleton tuple) = %q, want %q", got, "('a',)")
	}
}

func TestAsciiDict(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	d.DictSet(&PyString{Value: "k"}, MakeInt(1), vm)
	got := vm.ascii(d)
	if got != "{'k': 1}" {
		t.Errorf("ascii(dict) = %q, want %q", got, "{'k': 1}")
	}
}

func TestAsciiEmptySet(t *testing.T) {
	vm := NewVM()
	s := &PySet{Items: make(map[Value]struct{})}
	got := vm.ascii(s)
	if got != "set()" {
		t.Errorf("ascii(empty set) = %q, want %q", got, "set()")
	}
}

func TestAsciiNonEmptySet(t *testing.T) {
	vm := NewVM()
	s := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	s.SetAdd(MakeInt(1), vm)
	got := vm.ascii(s)
	if got != "{1}" {
		t.Errorf("ascii(set) = %q, want %q", got, "{1}")
	}
}

func TestAsciiInt(t *testing.T) {
	vm := NewVM()
	got := vm.ascii(MakeInt(42))
	if got != "42" {
		t.Errorf("ascii(42) = %q, want %q", got, "42")
	}
}

// =====================================
// bytesRepr: edge cases
// =====================================

func TestBytesReprSingleQuote(t *testing.T) {
	got := bytesRepr([]byte("it's"))
	if got != `b'it\'s'` {
		t.Errorf("bytesRepr with quote = %q, want %q", got, `b'it\'s'`)
	}
}

func TestBytesReprEmpty(t *testing.T) {
	got := bytesRepr([]byte{})
	if got != "b''" {
		t.Errorf("bytesRepr(empty) = %q, want %q", got, "b''")
	}
}

// =====================================
// asciiRepr: edge cases
// =====================================

func TestAsciiReprSpecialChars(t *testing.T) {
	got := asciiRepr("a\tb\nc\\d'e")
	if got != `'a\tb\nc\\d\'e'` {
		t.Errorf("asciiRepr = %q, want %q", got, `'a\tb\nc\\d\'e'`)
	}
}

func TestAsciiReprCarriageReturn(t *testing.T) {
	got := asciiRepr("a\rb")
	if got != `'a\rb'` {
		t.Errorf("asciiRepr(cr) = %q, want %q", got, `'a\rb'`)
	}
}

func TestAsciiReprHighUnicode(t *testing.T) {
	// Test a Unicode char > 0xffff (e.g. emoji)
	got := asciiRepr("\U0001F600") // grinning face
	if got != "'\\U0001f600'" {
		t.Errorf("asciiRepr(emoji) = %q, want %q", got, "'\\U0001f600'")
	}
}

func TestAsciiReprMidUnicode(t *testing.T) {
	// Test a Unicode char in range 0x100-0xffff
	got := asciiRepr("\u0100") // A with macron
	if got != "'\\u0100'" {
		t.Errorf("asciiRepr(\\u0100) = %q, want %q", got, "'\\u0100'")
	}
}

// =====================================
// formatException / formatExceptionInstance: coverage
// =====================================

func TestStrPyExceptionWithArgs(t *testing.T) {
	vm := NewVM()
	exc := &PyException{
		TypeName: "ValueError",
		Args:     &PyTuple{Items: []Value{&PyString{Value: "bad value"}}},
	}
	got := vm.str(exc)
	if got != "bad value" {
		t.Errorf("str(exc with single arg) = %q, want %q", got, "bad value")
	}
}

func TestStrPyExceptionWithMultipleArgs(t *testing.T) {
	vm := NewVM()
	exc := &PyException{
		TypeName: "ValueError",
		Args:     &PyTuple{Items: []Value{&PyString{Value: "a"}, MakeInt(42)}},
	}
	got := vm.str(exc)
	if got != "('a', 42)" {
		t.Errorf("str(exc with multi args) = %q, want %q", got, "('a', 42)")
	}
}

func TestStrPyExceptionWithMessage(t *testing.T) {
	vm := NewVM()
	exc := &PyException{
		TypeName: "RuntimeError",
		Message:  "something went wrong",
	}
	got := vm.str(exc)
	if got != "something went wrong" {
		t.Errorf("str(exc with message) = %q, want %q", got, "something went wrong")
	}
}

func TestStrPyExceptionEmpty(t *testing.T) {
	vm := NewVM()
	exc := &PyException{
		TypeName: "Exception",
	}
	got := vm.str(exc)
	if got != "" {
		t.Errorf("str(empty exc) = %q, want %q", got, "")
	}
}
