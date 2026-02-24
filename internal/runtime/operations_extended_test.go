package runtime

import (
	"math"
	"strings"
	"testing"
)

// =====================================
// compareOp: extended type combinations
// =====================================

func TestCompareOpBoolOrdering(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		op   Opcode
		a, b Value
		want bool
	}{
		{"True > False", OpCompareGt, True, False, true},
		{"False < True", OpCompareLt, False, True, true},
		{"True >= True", OpCompareGe, True, True, true},
		{"False <= False", OpCompareLe, False, False, true},
		{"True == True", OpCompareEq, True, True, true},
		{"True != False", OpCompareNe, True, False, true},
		// Bool vs Int ordering
		{"True < 2", OpCompareLt, True, MakeInt(2), true},
		{"False >= 0", OpCompareGe, False, MakeInt(0), true},
		{"True <= 1", OpCompareLe, True, MakeInt(1), true},
		{"False > -1", OpCompareGt, False, MakeInt(-1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.compareOp(tt.op, tt.a, tt.b)
			got, ok := result.(*PyBool)
			if !ok {
				t.Fatalf("expected *PyBool, got %T (result: %v)", result, result)
			}
			if got.Value != tt.want {
				t.Errorf("got %v, want %v", got.Value, tt.want)
			}
		})
	}
}

func TestCompareOpNoneEquality(t *testing.T) {
	vm := NewVM()
	// None == None
	result := vm.compareOp(OpCompareEq, None, None)
	if !result.(*PyBool).Value {
		t.Error("None == None should be True")
	}
	// None != 0
	result = vm.compareOp(OpCompareNe, None, MakeInt(0))
	if !result.(*PyBool).Value {
		t.Error("None != 0 should be True")
	}
	// None != ""
	result = vm.compareOp(OpCompareNe, None, &PyString{Value: ""})
	if !result.(*PyBool).Value {
		t.Error("None != '' should be True")
	}
	// None != False
	result = vm.compareOp(OpCompareNe, None, False)
	if !result.(*PyBool).Value {
		t.Error("None != False should be True")
	}
}

func TestCompareOpNoneNotOrderable(t *testing.T) {
	vm := NewVM()
	result := vm.compareOp(OpCompareLt, None, MakeInt(0))
	if result != nil {
		t.Error("expected nil result for None < 0 (TypeError)")
	}
	if vm.currentException == nil {
		t.Error("expected currentException to be set for None < 0")
	}
	vm.currentException = nil
}

func TestCompareOpTupleComparison(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		op   Opcode
		a, b *PyTuple
		want bool
	}{
		{"eq same", OpCompareEq,
			&PyTuple{Items: []Value{MakeInt(1), MakeInt(2)}},
			&PyTuple{Items: []Value{MakeInt(1), MakeInt(2)}}, true},
		{"eq different len", OpCompareEq,
			&PyTuple{Items: []Value{MakeInt(1)}},
			&PyTuple{Items: []Value{MakeInt(1), MakeInt(2)}}, false},
		{"lt first elem", OpCompareLt,
			&PyTuple{Items: []Value{MakeInt(1), MakeInt(2)}},
			&PyTuple{Items: []Value{MakeInt(2), MakeInt(0)}}, true},
		{"lt shorter prefix", OpCompareLt,
			&PyTuple{Items: []Value{MakeInt(1)}},
			&PyTuple{Items: []Value{MakeInt(1), MakeInt(2)}}, true},
		{"gt longer", OpCompareGt,
			&PyTuple{Items: []Value{MakeInt(1), MakeInt(2)}},
			&PyTuple{Items: []Value{MakeInt(1)}}, true},
		{"le equal", OpCompareLe,
			&PyTuple{Items: []Value{MakeInt(1)}},
			&PyTuple{Items: []Value{MakeInt(1)}}, true},
		{"ge equal", OpCompareGe,
			&PyTuple{Items: []Value{MakeInt(1)}},
			&PyTuple{Items: []Value{MakeInt(1)}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.compareOp(tt.op, tt.a, tt.b)
			got, ok := result.(*PyBool)
			if !ok {
				t.Fatalf("expected *PyBool, got %T", result)
			}
			if got.Value != tt.want {
				t.Errorf("got %v, want %v", got.Value, tt.want)
			}
		})
	}
}

func TestCompareOpListComparison(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		op   Opcode
		a, b *PyList
		want bool
	}{
		{"eq same", OpCompareEq,
			&PyList{Items: []Value{MakeInt(1), MakeInt(2)}},
			&PyList{Items: []Value{MakeInt(1), MakeInt(2)}}, true},
		{"ne different", OpCompareNe,
			&PyList{Items: []Value{MakeInt(1)}},
			&PyList{Items: []Value{MakeInt(2)}}, true},
		{"lt first elem", OpCompareLt,
			&PyList{Items: []Value{MakeInt(1)}},
			&PyList{Items: []Value{MakeInt(2)}}, true},
		{"gt first elem", OpCompareGt,
			&PyList{Items: []Value{MakeInt(3)}},
			&PyList{Items: []Value{MakeInt(2)}}, true},
		{"le shorter", OpCompareLe,
			&PyList{Items: []Value{MakeInt(1)}},
			&PyList{Items: []Value{MakeInt(1), MakeInt(0)}}, true},
		{"ge longer", OpCompareGe,
			&PyList{Items: []Value{MakeInt(1), MakeInt(0)}},
			&PyList{Items: []Value{MakeInt(1)}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.compareOp(tt.op, tt.a, tt.b)
			got, ok := result.(*PyBool)
			if !ok {
				t.Fatalf("expected *PyBool, got %T", result)
			}
			if got.Value != tt.want {
				t.Errorf("got %v, want %v", got.Value, tt.want)
			}
		})
	}
}

func TestCompareOpStringOrdering(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		op   Opcode
		a, b string
		want bool
	}{
		{"le true", OpCompareLe, "abc", "abd", true},
		{"le equal", OpCompareLe, "abc", "abc", true},
		{"ge true", OpCompareGe, "xyz", "abc", true},
		{"ge equal", OpCompareGe, "abc", "abc", true},
		{"ne true", OpCompareNe, "abc", "xyz", true},
		{"ne false", OpCompareNe, "abc", "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.compareOp(tt.op, &PyString{Value: tt.a}, &PyString{Value: tt.b})
			got, ok := result.(*PyBool)
			if !ok {
				t.Fatalf("expected *PyBool, got %T", result)
			}
			if got.Value != tt.want {
				t.Errorf("got %v, want %v", got.Value, tt.want)
			}
		})
	}
}

func TestCompareOpIntVsFloat(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		op   Opcode
		a, b Value
		want bool
	}{
		{"int lt float", OpCompareLt, MakeInt(1), &PyFloat{Value: 1.5}, true},
		{"float lt int", OpCompareLt, &PyFloat{Value: 0.5}, MakeInt(1), true},
		{"int ge float", OpCompareGe, MakeInt(2), &PyFloat{Value: 1.9}, true},
		{"float ge int", OpCompareGe, &PyFloat{Value: 2.1}, MakeInt(2), true},
		{"int le float equal", OpCompareLe, MakeInt(3), &PyFloat{Value: 3.0}, true},
		{"float gt int", OpCompareGt, &PyFloat{Value: 5.1}, MakeInt(5), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.compareOp(tt.op, tt.a, tt.b)
			got, ok := result.(*PyBool)
			if !ok {
				t.Fatalf("expected *PyBool, got %T", result)
			}
			if got.Value != tt.want {
				t.Errorf("got %v, want %v", got.Value, tt.want)
			}
		})
	}
}

func TestCompareOpComplexEquality(t *testing.T) {
	vm := NewVM()
	// Complex supports == and !=
	result := vm.compareOp(OpCompareEq, MakeComplex(1, 2), MakeComplex(1, 2))
	if !result.(*PyBool).Value {
		t.Error("(1+2j) == (1+2j) should be True")
	}
	result = vm.compareOp(OpCompareNe, MakeComplex(1, 2), MakeComplex(3, 4))
	if !result.(*PyBool).Value {
		t.Error("(1+2j) != (3+4j) should be True")
	}
}

func TestCompareOpComplexNotOrderable(t *testing.T) {
	vm := NewVM()
	ops := []struct {
		name string
		op   Opcode
	}{
		{"lt", OpCompareLt},
		{"le", OpCompareLe},
		{"gt", OpCompareGt},
		{"ge", OpCompareGe},
	}
	for _, tt := range ops {
		t.Run(tt.name, func(t *testing.T) {
			vm.currentException = nil
			result := vm.compareOp(tt.op, MakeComplex(1, 2), MakeComplex(3, 4))
			if result != nil {
				t.Error("expected nil result for complex ordering")
			}
			if vm.currentException == nil {
				t.Error("expected TypeError for complex ordering")
			}
			vm.currentException = nil
		})
	}
}

// =====================================
// compareOp: set comparisons
// =====================================

func TestCompareOpSetSubset(t *testing.T) {
	vm := NewVM()
	a := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	a.SetAdd(MakeInt(1), vm)
	b := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	b.SetAdd(MakeInt(1), vm)
	b.SetAdd(MakeInt(2), vm)

	// a <= b (subset)
	result := vm.compareOp(OpCompareLe, a, b)
	if !result.(*PyBool).Value {
		t.Error("{1} <= {1,2} should be True")
	}

	// a < b (proper subset)
	result = vm.compareOp(OpCompareLt, a, b)
	if !result.(*PyBool).Value {
		t.Error("{1} < {1,2} should be True")
	}

	// b >= a (superset)
	result = vm.compareOp(OpCompareGe, b, a)
	if !result.(*PyBool).Value {
		t.Error("{1,2} >= {1} should be True")
	}

	// b > a (proper superset)
	result = vm.compareOp(OpCompareGt, b, a)
	if !result.(*PyBool).Value {
		t.Error("{1,2} > {1} should be True")
	}
}

func TestCompareOpSetEquality(t *testing.T) {
	vm := NewVM()
	a := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	a.SetAdd(MakeInt(1), vm)
	a.SetAdd(MakeInt(2), vm)
	b := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	b.SetAdd(MakeInt(1), vm)
	b.SetAdd(MakeInt(2), vm)

	result := vm.compareOp(OpCompareEq, a, b)
	if !result.(*PyBool).Value {
		t.Error("{1,2} == {1,2} should be True")
	}
	result = vm.compareOp(OpCompareNe, a, b)
	if result.(*PyBool).Value {
		t.Error("{1,2} != {1,2} should be False")
	}
}

func TestCompareOpSetNotProperSubset(t *testing.T) {
	vm := NewVM()
	a := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	a.SetAdd(MakeInt(1), vm)
	a.SetAdd(MakeInt(2), vm)
	b := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	b.SetAdd(MakeInt(1), vm)
	b.SetAdd(MakeInt(2), vm)

	// a < b (proper subset) should be false when equal
	result := vm.compareOp(OpCompareLt, a, b)
	if result.(*PyBool).Value {
		t.Error("{1,2} < {1,2} should be False")
	}

	// a > b (proper superset) should be false when equal
	result = vm.compareOp(OpCompareGt, a, b)
	if result.(*PyBool).Value {
		t.Error("{1,2} > {1,2} should be False")
	}
}

// =====================================
// contains: extended cases
// =====================================

func TestContainsStringInString(t *testing.T) {
	vm := NewVM()
	s := &PyString{Value: "abcdef"}
	if !vm.contains(s, &PyString{Value: "cd"}) {
		t.Error("'cd' should be in 'abcdef'")
	}
	if !vm.contains(s, &PyString{Value: ""}) {
		t.Error("'' should be in 'abcdef'")
	}
	if vm.contains(s, &PyString{Value: "xyz"}) {
		t.Error("'xyz' should not be in 'abcdef'")
	}
}

func TestContainsNonStringInStringTypeError(t *testing.T) {
	vm := NewVM()
	vm.currentException = nil
	vm.contains(&PyString{Value: "hello"}, MakeInt(1))
	if vm.currentException == nil {
		t.Error("expected TypeError for int in string")
	}
	if !strings.Contains(vm.currentException.Message, "requires string") {
		t.Errorf("expected 'requires string' message, got: %s", vm.currentException.Message)
	}
	vm.currentException = nil
}

func TestContainsFrozenSet(t *testing.T) {
	vm := NewVM()
	fs := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	fs.FrozenSetAdd(MakeInt(10), vm)
	fs.FrozenSetAdd(MakeInt(20), vm)

	if !vm.contains(fs, MakeInt(10)) {
		t.Error("10 should be in frozenset")
	}
	if vm.contains(fs, MakeInt(30)) {
		t.Error("30 should not be in frozenset")
	}
}

func TestContainsBytesSubsequence(t *testing.T) {
	vm := NewVM()
	b := &PyBytes{Value: []byte{1, 2, 3, 4, 5}}
	// Check individual byte
	if !vm.contains(b, MakeInt(3)) {
		t.Error("3 should be in bytes")
	}
	// Check bytes subsequence
	sub := &PyBytes{Value: []byte{2, 3}}
	if !vm.contains(b, sub) {
		t.Error("b'\\x02\\x03' should be in bytes")
	}
	// Check bytes subsequence not found
	sub2 := &PyBytes{Value: []byte{3, 2}}
	if vm.contains(b, sub2) {
		t.Error("b'\\x03\\x02' should not be in bytes")
	}
	// Empty bytes always contained
	empty := &PyBytes{Value: []byte{}}
	if !vm.contains(b, empty) {
		t.Error("empty bytes should be in bytes")
	}
}

func TestContainsRangeWithBool(t *testing.T) {
	vm := NewVM()
	r := &PyRange{Start: 0, Stop: 5, Step: 1}
	if !vm.contains(r, True) {
		t.Error("True (1) should be in range(0, 5)")
	}
	if !vm.contains(r, False) {
		t.Error("False (0) should be in range(0, 5)")
	}
}

func TestContainsEmptyList(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{}}
	if vm.contains(list, MakeInt(1)) {
		t.Error("1 should not be in empty list")
	}
}

func TestContainsListWithNone(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{None, MakeInt(1)}}
	if !vm.contains(list, None) {
		t.Error("None should be in [None, 1]")
	}
}

func TestContainsTupleWithMixedTypes(t *testing.T) {
	vm := NewVM()
	tup := &PyTuple{Items: []Value{MakeInt(1), &PyString{Value: "hello"}, True}}
	if !vm.contains(tup, &PyString{Value: "hello"}) {
		t.Error("'hello' should be in tuple")
	}
	if vm.contains(tup, &PyString{Value: "world"}) {
		t.Error("'world' should not be in tuple")
	}
}

func TestContainsDictChecksKeys(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	d.DictSet(&PyString{Value: "key"}, &PyString{Value: "value"}, vm)
	// "key" is in dict (checks keys)
	if !vm.contains(d, &PyString{Value: "key"}) {
		t.Error("'key' should be in dict")
	}
	// "value" is NOT in dict (it's a value, not a key)
	if vm.contains(d, &PyString{Value: "value"}) {
		t.Error("'value' should not be in dict (it's a value, not a key)")
	}
}

// =====================================
// equalWithCycleDetection: extended
// =====================================

func TestEqualDicts(t *testing.T) {
	vm := NewVM()
	a := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	a.DictSet(&PyString{Value: "x"}, MakeInt(1), vm)
	a.DictSet(&PyString{Value: "y"}, MakeInt(2), vm)

	b := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	b.DictSet(&PyString{Value: "x"}, MakeInt(1), vm)
	b.DictSet(&PyString{Value: "y"}, MakeInt(2), vm)

	if !vm.equal(a, b) {
		t.Error("equal dicts should be equal")
	}
}

func TestEqualDictsDifferentValues(t *testing.T) {
	vm := NewVM()
	a := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	a.DictSet(&PyString{Value: "x"}, MakeInt(1), vm)

	b := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	b.DictSet(&PyString{Value: "x"}, MakeInt(2), vm)

	if vm.equal(a, b) {
		t.Error("dicts with different values should not be equal")
	}
}

func TestEqualDictsDifferentSize(t *testing.T) {
	vm := NewVM()
	a := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	a.DictSet(&PyString{Value: "x"}, MakeInt(1), vm)

	b := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	b.DictSet(&PyString{Value: "x"}, MakeInt(1), vm)
	b.DictSet(&PyString{Value: "y"}, MakeInt(2), vm)

	if vm.equal(a, b) {
		t.Error("dicts with different sizes should not be equal")
	}
}

func TestEqualSets(t *testing.T) {
	vm := NewVM()
	a := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	a.SetAdd(MakeInt(1), vm)
	a.SetAdd(MakeInt(2), vm)

	b := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	b.SetAdd(MakeInt(1), vm)
	b.SetAdd(MakeInt(2), vm)

	if !vm.equal(a, b) {
		t.Error("equal sets should be equal")
	}
}

func TestEqualSetsDifferent(t *testing.T) {
	vm := NewVM()
	a := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	a.SetAdd(MakeInt(1), vm)

	b := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	b.SetAdd(MakeInt(2), vm)

	if vm.equal(a, b) {
		t.Error("different sets should not be equal")
	}
}

func TestEqualFrozenSets(t *testing.T) {
	vm := NewVM()
	a := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	a.FrozenSetAdd(MakeInt(1), vm)

	b := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	b.FrozenSetAdd(MakeInt(1), vm)

	if !vm.equal(a, b) {
		t.Error("equal frozen sets should be equal")
	}
}

func TestEqualSetAndFrozenSet(t *testing.T) {
	vm := NewVM()
	a := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	a.SetAdd(MakeInt(1), vm)

	b := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	b.FrozenSetAdd(MakeInt(1), vm)

	if !vm.equal(a, b) {
		t.Error("set {1} should equal frozenset {1}")
	}
}

func TestEqualFrozenSetAndSet(t *testing.T) {
	vm := NewVM()
	a := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	a.FrozenSetAdd(MakeInt(1), vm)

	b := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	b.SetAdd(MakeInt(1), vm)

	if !vm.equal(a, b) {
		t.Error("frozenset {1} should equal set {1}")
	}
}

func TestEqualRanges(t *testing.T) {
	vm := NewVM()
	// Same range
	a := &PyRange{Start: 0, Stop: 10, Step: 2}
	b := &PyRange{Start: 0, Stop: 10, Step: 2}
	if !vm.equal(a, b) {
		t.Error("equal ranges should be equal")
	}

	// Empty ranges are equal regardless of params
	c := &PyRange{Start: 0, Stop: 0, Step: 1}
	d := &PyRange{Start: 5, Stop: 5, Step: 3}
	if !vm.equal(c, d) {
		t.Error("empty ranges should be equal")
	}

	// Single element ranges with same start
	e := &PyRange{Start: 0, Stop: 1, Step: 1}
	f := &PyRange{Start: 0, Stop: 5, Step: 10}
	if !vm.equal(e, f) {
		t.Error("single element ranges with same start should be equal")
	}

	// Different step, different range
	g := &PyRange{Start: 0, Stop: 10, Step: 2}
	h := &PyRange{Start: 0, Stop: 10, Step: 3}
	if vm.equal(g, h) {
		t.Error("ranges with different steps should not be equal")
	}
}

func TestEqualBytes(t *testing.T) {
	vm := NewVM()
	a := &PyBytes{Value: []byte{1, 2, 3}}
	b := &PyBytes{Value: []byte{1, 2, 3}}
	if !vm.equal(a, b) {
		t.Error("equal bytes should be equal")
	}

	c := &PyBytes{Value: []byte{1, 2, 4}}
	if vm.equal(a, c) {
		t.Error("different bytes should not be equal")
	}

	d := &PyBytes{Value: []byte{1, 2}}
	if vm.equal(a, d) {
		t.Error("different length bytes should not be equal")
	}
}

func TestEqualComplex(t *testing.T) {
	vm := NewVM()
	if !vm.equal(MakeComplex(1, 2), MakeComplex(1, 2)) {
		t.Error("equal complex should be equal")
	}
	if vm.equal(MakeComplex(1, 2), MakeComplex(1, 3)) {
		t.Error("different complex should not be equal")
	}
}

func TestEqualComplexAndInt(t *testing.T) {
	vm := NewVM()
	// 5+0j == 5
	if !vm.equal(MakeComplex(5, 0), MakeInt(5)) {
		t.Error("5+0j should equal 5")
	}
	// 5 == 5+0j
	if !vm.equal(MakeInt(5), MakeComplex(5, 0)) {
		t.Error("5 should equal 5+0j")
	}
	// 5+1j != 5
	if vm.equal(MakeComplex(5, 1), MakeInt(5)) {
		t.Error("5+1j should not equal 5")
	}
}

func TestEqualComplexAndFloat(t *testing.T) {
	vm := NewVM()
	if !vm.equal(MakeComplex(3.14, 0), &PyFloat{Value: 3.14}) {
		t.Error("3.14+0j should equal 3.14")
	}
	if !vm.equal(&PyFloat{Value: 3.14}, MakeComplex(3.14, 0)) {
		t.Error("3.14 should equal 3.14+0j")
	}
}

func TestEqualComplexAndBool(t *testing.T) {
	vm := NewVM()
	if !vm.equal(MakeComplex(1, 0), True) {
		t.Error("1+0j should equal True")
	}
	if !vm.equal(MakeComplex(0, 0), False) {
		t.Error("0+0j should equal False")
	}
}

func TestEqualBoolAndFloat(t *testing.T) {
	vm := NewVM()
	if !vm.equal(True, &PyFloat{Value: 1.0}) {
		t.Error("True should equal 1.0")
	}
	if !vm.equal(False, &PyFloat{Value: 0.0}) {
		t.Error("False should equal 0.0")
	}
}

func TestEqualMixedTypesFalse(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		a, b Value
	}{
		{"string vs int", &PyString{Value: "1"}, MakeInt(1)},
		{"string vs list", &PyString{Value: "a"}, &PyList{Items: []Value{&PyString{Value: "a"}}}},
		{"list vs tuple", &PyList{Items: []Value{MakeInt(1)}}, &PyTuple{Items: []Value{MakeInt(1)}}},
		{"int vs None", MakeInt(0), None},
		{"string vs None", &PyString{Value: ""}, None},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if vm.equal(tt.a, tt.b) {
				t.Errorf("expected %T and %T to not be equal", tt.a, tt.b)
			}
		})
	}
}

func TestEqualClassIdentity(t *testing.T) {
	vm := NewVM()
	cls1 := &PyClass{Name: "A", Dict: map[string]Value{}}
	cls2 := &PyClass{Name: "A", Dict: map[string]Value{}}
	if !vm.equal(cls1, cls1) {
		t.Error("same class should be equal to itself")
	}
	if vm.equal(cls1, cls2) {
		t.Error("different class objects should not be equal (identity comparison)")
	}
}

// =====================================
// binaryOp: extended untested branches
// =====================================

func TestBinaryOpBytesConcatenation(t *testing.T) {
	vm := NewVM()
	a := &PyBytes{Value: []byte{1, 2}}
	b := &PyBytes{Value: []byte{3, 4}}
	result, err := vm.binaryOp(OpBinaryAdd, a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyBytes)
	if !ok {
		t.Fatalf("expected *PyBytes, got %T", result)
	}
	if len(got.Value) != 4 || got.Value[0] != 1 || got.Value[3] != 4 {
		t.Errorf("expected [1,2,3,4], got %v", got.Value)
	}
}

func TestBinaryOpBytesRepetition(t *testing.T) {
	vm := NewVM()
	a := &PyBytes{Value: []byte{0xAB}}
	result, err := vm.binaryOp(OpBinaryMultiply, a, MakeInt(3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyBytes)
	if !ok {
		t.Fatalf("expected *PyBytes, got %T", result)
	}
	if len(got.Value) != 3 {
		t.Fatalf("expected 3 bytes, got %d", len(got.Value))
	}
	for _, b := range got.Value {
		if b != 0xAB {
			t.Errorf("expected 0xAB, got 0x%X", b)
		}
	}
}

func TestBinaryOpBytesRepetitionReverse(t *testing.T) {
	vm := NewVM()
	result, err := vm.binaryOp(OpBinaryMultiply, MakeInt(2), &PyBytes{Value: []byte{0xCD}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyBytes)
	if len(got.Value) != 2 {
		t.Fatalf("expected 2 bytes, got %d", len(got.Value))
	}
}

func TestBinaryOpBytesRepetitionZero(t *testing.T) {
	vm := NewVM()
	result, err := vm.binaryOp(OpBinaryMultiply, &PyBytes{Value: []byte{1, 2}}, MakeInt(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyBytes)
	if len(got.Value) != 0 {
		t.Errorf("expected empty bytes, got %v", got.Value)
	}
}

func TestBinaryOpBytesRepetitionNegative(t *testing.T) {
	vm := NewVM()
	result, err := vm.binaryOp(OpBinaryMultiply, &PyBytes{Value: []byte{1}}, MakeInt(-5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyBytes)
	if len(got.Value) != 0 {
		t.Errorf("expected empty bytes, got %v", got.Value)
	}
}

func TestBinaryOpTupleRepetitionReverse(t *testing.T) {
	vm := NewVM()
	result, err := vm.binaryOp(OpBinaryMultiply, MakeInt(2), &PyTuple{Items: []Value{MakeInt(1)}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyTuple)
	if len(got.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got.Items))
	}
}

func TestBinaryOpTupleRepetitionZero(t *testing.T) {
	vm := NewVM()
	result, err := vm.binaryOp(OpBinaryMultiply, &PyTuple{Items: []Value{MakeInt(1)}}, MakeInt(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyTuple)
	if len(got.Items) != 0 {
		t.Errorf("expected empty tuple, got %d items", len(got.Items))
	}
}

func TestBinaryOpListRepetitionReverse(t *testing.T) {
	vm := NewVM()
	result, err := vm.binaryOp(OpBinaryMultiply, MakeInt(3), &PyList{Items: []Value{MakeInt(1)}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyList)
	if len(got.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(got.Items))
	}
}

func TestBinaryOpListRepetitionZero(t *testing.T) {
	vm := NewVM()
	result, err := vm.binaryOp(OpBinaryMultiply, &PyList{Items: []Value{MakeInt(1)}}, MakeInt(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyList)
	if len(got.Items) != 0 {
		t.Errorf("expected empty list, got %d items", len(got.Items))
	}
}

func TestBinaryOpListRepetitionNeg(t *testing.T) {
	vm := NewVM()
	result, err := vm.binaryOp(OpBinaryMultiply, &PyList{Items: []Value{MakeInt(1)}}, MakeInt(-1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyList)
	if len(got.Items) != 0 {
		t.Errorf("expected empty list, got %d items", len(got.Items))
	}
}

func TestBinaryOpTupleRepetitionNeg(t *testing.T) {
	vm := NewVM()
	result, err := vm.binaryOp(OpBinaryMultiply, &PyTuple{Items: []Value{MakeInt(1)}}, MakeInt(-2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyTuple)
	if len(got.Items) != 0 {
		t.Errorf("expected empty tuple, got %d items", len(got.Items))
	}
}

func TestBinaryOpStringModFormat(t *testing.T) {
	vm := NewVM()
	result, err := vm.binaryOp(OpBinaryModulo, &PyString{Value: "hello %s"}, &PyString{Value: "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyString)
	if !ok {
		t.Fatalf("expected *PyString, got %T", result)
	}
	if got.Value != "hello world" {
		t.Errorf("got %q, want %q", got.Value, "hello world")
	}
}

func TestBinaryOpDictMerge(t *testing.T) {
	vm := NewVM()
	a := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	a.DictSet(&PyString{Value: "a"}, MakeInt(1), vm)
	b := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	b.DictSet(&PyString{Value: "b"}, MakeInt(2), vm)

	result, err := vm.binaryOp(OpBinaryOr, a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyDict)
	if !ok {
		t.Fatalf("expected *PyDict, got %T", result)
	}
	v1, ok1 := got.DictGet(&PyString{Value: "a"}, vm)
	v2, ok2 := got.DictGet(&PyString{Value: "b"}, vm)
	if !ok1 || !ok2 {
		t.Fatal("merged dict should have both keys")
	}
	if v1.(*PyInt).Value != 1 || v2.(*PyInt).Value != 2 {
		t.Error("merged dict values incorrect")
	}
}

func TestBinaryOpComplexDivision(t *testing.T) {
	vm := NewVM()
	// (4+2j) / (2+0j) = (2+1j)
	result, err := vm.binaryOp(OpBinaryDivide, MakeComplex(4, 2), MakeComplex(2, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyComplex)
	if math.Abs(got.Real-2) > 1e-10 || math.Abs(got.Imag-1) > 1e-10 {
		t.Errorf("got (%f, %f), want (2, 1)", got.Real, got.Imag)
	}
}

func TestBinaryOpComplexDivisionByZero(t *testing.T) {
	vm := NewVM()
	_, err := vm.binaryOp(OpBinaryDivide, MakeComplex(1, 2), MakeComplex(0, 0))
	if err == nil {
		t.Fatal("expected division by zero error")
	}
	if !strings.Contains(err.Error(), "ZeroDivisionError") {
		t.Errorf("expected ZeroDivisionError, got: %v", err)
	}
}

func TestBinaryOpComplexFloorDivError(t *testing.T) {
	vm := NewVM()
	_, err := vm.binaryOp(OpBinaryFloorDiv, MakeComplex(1, 2), MakeComplex(3, 4))
	if err == nil {
		t.Fatal("expected TypeError for complex floor div")
	}
	if !strings.Contains(err.Error(), "TypeError") {
		t.Errorf("expected TypeError, got: %v", err)
	}
}

func TestBinaryOpComplexModuloError(t *testing.T) {
	vm := NewVM()
	_, err := vm.binaryOp(OpBinaryModulo, MakeComplex(1, 2), MakeComplex(3, 4))
	if err == nil {
		t.Fatal("expected TypeError for complex modulo")
	}
}

func TestBinaryOpComplexPower(t *testing.T) {
	vm := NewVM()
	result, err := vm.binaryOp(OpBinaryPower, MakeComplex(0, 1), MakeComplex(2, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyComplex)
	// i^2 = -1 (with possible floating point error)
	if math.Abs(got.Real+1) > 1e-10 {
		t.Errorf("real part: got %f, want -1", got.Real)
	}
}

func TestBinaryOpComplexPromotionFromFloat(t *testing.T) {
	vm := NewVM()
	// complex + float should promote float to complex
	result, err := vm.binaryOp(OpBinaryAdd, MakeComplex(1, 2), &PyFloat{Value: 3.0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyComplex)
	if got.Real != 4.0 || got.Imag != 2.0 {
		t.Errorf("got (%f, %f), want (4.0, 2.0)", got.Real, got.Imag)
	}
}

func TestBinaryOpIntNegativePower(t *testing.T) {
	vm := NewVM()
	// 2 ** -1 returns float 0.5
	result, err := vm.binaryOp(OpBinaryPower, MakeInt(2), MakeInt(-1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyFloat)
	if !ok {
		t.Fatalf("expected *PyFloat for negative exponent, got %T", result)
	}
	if got.Value != 0.5 {
		t.Errorf("got %f, want 0.5", got.Value)
	}
}

func TestBinaryOpIntNegativeShiftCount(t *testing.T) {
	vm := NewVM()
	_, err := vm.binaryOp(OpBinaryLShift, MakeInt(1), MakeInt(-1))
	if err == nil {
		t.Fatal("expected error for negative shift count")
	}
	if !strings.Contains(err.Error(), "negative shift count") {
		t.Errorf("expected 'negative shift count' error, got: %v", err)
	}
	_, err = vm.binaryOp(OpBinaryRShift, MakeInt(1), MakeInt(-1))
	if err == nil {
		t.Fatal("expected error for negative rshift count")
	}
}

func TestBinaryOpIntLargeShift(t *testing.T) {
	vm := NewVM()
	// Left shift by 64 should give 0
	result, err := vm.binaryOp(OpBinaryLShift, MakeInt(1), MakeInt(64))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyInt).Value
	if got != 0 {
		t.Errorf("1 << 64 = %d, want 0", got)
	}
	// Right shift by 64 should give 0 for positive
	result, err = vm.binaryOp(OpBinaryRShift, MakeInt(100), MakeInt(64))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got = result.(*PyInt).Value
	if got != 0 {
		t.Errorf("100 >> 64 = %d, want 0", got)
	}
	// Right shift by 64 should give -1 for negative
	result, err = vm.binaryOp(OpBinaryRShift, MakeInt(-1), MakeInt(64))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got = result.(*PyInt).Value
	if got != -1 {
		t.Errorf("-1 >> 64 = %d, want -1", got)
	}
}

func TestBinaryOpFloatModByZero(t *testing.T) {
	vm := NewVM()
	_, err := vm.binaryOp(OpBinaryModulo, &PyFloat{Value: 1.0}, &PyFloat{Value: 0.0})
	if err == nil {
		t.Fatal("expected error for float modulo by zero")
	}
}

func TestBinaryOpFloatFloorDivByZero(t *testing.T) {
	vm := NewVM()
	_, err := vm.binaryOp(OpBinaryFloorDiv, &PyFloat{Value: 1.0}, &PyFloat{Value: 0.0})
	if err == nil {
		t.Fatal("expected error for float floor div by zero")
	}
}

func TestBinaryOpSetUnion(t *testing.T) {
	vm := NewVM()
	a := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	a.SetAdd(MakeInt(1), vm)
	b := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	b.SetAdd(MakeInt(2), vm)

	result, err := vm.binaryOp(OpBinaryOr, a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PySet)
	if !ok {
		t.Fatalf("expected *PySet, got %T", result)
	}
	if len(got.Items) != 2 {
		t.Errorf("expected 2 items in union, got %d", len(got.Items))
	}
}

func TestBinaryOpSetIntersection(t *testing.T) {
	vm := NewVM()
	a := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	a.SetAdd(MakeInt(1), vm)
	a.SetAdd(MakeInt(2), vm)
	b := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	b.SetAdd(MakeInt(2), vm)
	b.SetAdd(MakeInt(3), vm)

	result, err := vm.binaryOp(OpBinaryAnd, a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PySet)
	if !ok {
		t.Fatalf("expected *PySet, got %T", result)
	}
	if len(got.Items) != 1 {
		t.Errorf("expected 1 item in intersection, got %d", len(got.Items))
	}
}

func TestBinaryOpSetDifference(t *testing.T) {
	vm := NewVM()
	a := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	a.SetAdd(MakeInt(1), vm)
	a.SetAdd(MakeInt(2), vm)
	b := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	b.SetAdd(MakeInt(2), vm)
	b.SetAdd(MakeInt(3), vm)

	result, err := vm.binaryOp(OpBinarySubtract, a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PySet)
	if !ok {
		t.Fatalf("expected *PySet, got %T", result)
	}
	if len(got.Items) != 1 {
		t.Errorf("expected 1 item in difference, got %d", len(got.Items))
	}
}

func TestBinaryOpSetSymmetricDifference(t *testing.T) {
	vm := NewVM()
	a := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	a.SetAdd(MakeInt(1), vm)
	a.SetAdd(MakeInt(2), vm)
	b := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	b.SetAdd(MakeInt(2), vm)
	b.SetAdd(MakeInt(3), vm)

	result, err := vm.binaryOp(OpBinaryXor, a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PySet)
	if !ok {
		t.Fatalf("expected *PySet, got %T", result)
	}
	if len(got.Items) != 2 {
		t.Errorf("expected 2 items in symmetric difference, got %d", len(got.Items))
	}
}

func TestBinaryOpIntPythonFloorDiv(t *testing.T) {
	vm := NewVM()
	// Python floor division rounds toward negative infinity
	// -7 // 2 = -4 (not -3)
	result, err := vm.binaryOp(OpBinaryFloorDiv, MakeInt(-7), MakeInt(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyInt).Value
	if got != -4 {
		t.Errorf("-7 // 2 = %d, want -4", got)
	}
}

func TestBinaryOpIntPythonModulo(t *testing.T) {
	vm := NewVM()
	// Python modulo: result has same sign as divisor
	// -7 % 3 = 2 (not -1)
	result, err := vm.binaryOp(OpBinaryModulo, MakeInt(-7), MakeInt(3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyInt).Value
	if got != 2 {
		t.Errorf("-7 %% 3 = %d, want 2", got)
	}
}

func TestBinaryOpFloatPythonModulo(t *testing.T) {
	vm := NewVM()
	// Python float modulo: -7.0 % 3.0 = 2.0
	result, err := vm.binaryOp(OpBinaryModulo, &PyFloat{Value: -7.0}, &PyFloat{Value: 3.0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyFloat).Value
	if math.Abs(got-2.0) > 1e-10 {
		t.Errorf("-7.0 %% 3.0 = %f, want 2.0", got)
	}
}

// =====================================
// unaryOp: extended branches
// =====================================

func TestUnaryOpPositiveFloat(t *testing.T) {
	vm := NewVM()
	result, err := vm.unaryOp(OpUnaryPositive, &PyFloat{Value: -3.14})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyFloat)
	if got.Value != -3.14 {
		t.Errorf("+(-3.14) = %f, want -3.14", got.Value)
	}
}

func TestUnaryOpPositiveComplex(t *testing.T) {
	vm := NewVM()
	result, err := vm.unaryOp(OpUnaryPositive, MakeComplex(1, -2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyComplex)
	if got.Real != 1 || got.Imag != -2 {
		t.Errorf("+(1-2j) = (%f, %f), want (1, -2)", got.Real, got.Imag)
	}
}

func TestUnaryOpNegativeFloat(t *testing.T) {
	vm := NewVM()
	result, err := vm.unaryOp(OpUnaryNegative, &PyFloat{Value: 2.5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyFloat)
	if got.Value != -2.5 {
		t.Errorf("-(2.5) = %f, want -2.5", got.Value)
	}
}

func TestUnaryOpInvertString(t *testing.T) {
	vm := NewVM()
	_, err := vm.unaryOp(OpUnaryInvert, &PyString{Value: "abc"})
	if err == nil {
		t.Fatal("expected error for ~string")
	}
	if !strings.Contains(err.Error(), "TypeError") {
		t.Errorf("expected TypeError, got: %v", err)
	}
}

func TestUnaryOpNegativeString(t *testing.T) {
	vm := NewVM()
	_, err := vm.unaryOp(OpUnaryNegative, &PyString{Value: "abc"})
	if err == nil {
		t.Fatal("expected error for -string")
	}
	if !strings.Contains(err.Error(), "TypeError") {
		t.Errorf("expected TypeError, got: %v", err)
	}
}

func TestUnaryOpPositiveString(t *testing.T) {
	vm := NewVM()
	_, err := vm.unaryOp(OpUnaryPositive, &PyString{Value: "abc"})
	if err == nil {
		t.Fatal("expected error for +string")
	}
}

func TestUnaryOpInvertComplex(t *testing.T) {
	vm := NewVM()
	_, err := vm.unaryOp(OpUnaryInvert, MakeComplex(1, 2))
	if err == nil {
		t.Fatal("expected error for ~complex")
	}
}

// =====================================
// compare: bytes comparison
// =====================================

func TestCompareOpBytesOrdering(t *testing.T) {
	vm := NewVM()
	a := &PyBytes{Value: []byte{1, 2, 3}}
	b := &PyBytes{Value: []byte{1, 2, 4}}

	result := vm.compareOp(OpCompareLt, a, b)
	if !result.(*PyBool).Value {
		t.Error("b'\\x01\\x02\\x03' < b'\\x01\\x02\\x04' should be True")
	}

	result = vm.compareOp(OpCompareGt, b, a)
	if !result.(*PyBool).Value {
		t.Error("b'\\x01\\x02\\x04' > b'\\x01\\x02\\x03' should be True")
	}

	// Same prefix, different length
	c := &PyBytes{Value: []byte{1, 2}}
	result = vm.compareOp(OpCompareLt, c, a)
	if !result.(*PyBool).Value {
		t.Error("shorter prefix should be less")
	}

	result = vm.compareOp(OpCompareEq, a, &PyBytes{Value: []byte{1, 2, 3}})
	if !result.(*PyBool).Value {
		t.Error("equal bytes should be equal")
	}
}

// =====================================
// bytesContains: edge cases
// =====================================

func TestBytesContainsEdgeCases(t *testing.T) {
	// sub longer than data
	if bytesContains([]byte{1}, []byte{1, 2}) {
		t.Error("sub longer than data should return false")
	}
	// exact match
	if !bytesContains([]byte{1, 2, 3}, []byte{1, 2, 3}) {
		t.Error("exact match should return true")
	}
	// not found
	if bytesContains([]byte{1, 2, 3}, []byte{4}) {
		t.Error("should return false when not found")
	}
	// at end
	if !bytesContains([]byte{1, 2, 3}, []byte{2, 3}) {
		t.Error("should find at end")
	}
}

// =====================================
// areBuiltinOrderable: coverage
// =====================================

func TestAreBuiltinOrderable(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		a, b Value
		want bool
	}{
		{"int int", MakeInt(1), MakeInt(2), true},
		{"int float", MakeInt(1), &PyFloat{Value: 2.0}, true},
		{"float float", &PyFloat{Value: 1.0}, &PyFloat{Value: 2.0}, true},
		{"float int", &PyFloat{Value: 1.0}, MakeInt(2), true},
		{"str str", &PyString{Value: "a"}, &PyString{Value: "b"}, true},
		{"list list", &PyList{Items: nil}, &PyList{Items: nil}, true},
		{"tuple tuple", &PyTuple{Items: nil}, &PyTuple{Items: nil}, true},
		{"bytes bytes", &PyBytes{Value: nil}, &PyBytes{Value: nil}, true},
		{"bool bool", True, False, true},
		{"bool int", True, MakeInt(1), true},
		{"bool float", False, &PyFloat{Value: 0.0}, true},
		{"str int", &PyString{Value: "a"}, MakeInt(1), false},
		{"list str", &PyList{Items: nil}, &PyString{Value: ""}, false},
		{"int str", MakeInt(1), &PyString{Value: "a"}, false},
		{"None int", None, MakeInt(0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := vm.areBuiltinOrderable(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("areBuiltinOrderable(%T, %T) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// =====================================
// containsIdentityOrEqual: coverage
// =====================================

func TestContainsIdentityOrEqual(t *testing.T) {
	vm := NewVM()
	// Identity check: same pointer
	x := MakeInt(999)
	if !vm.containsIdentityOrEqual(x, x) {
		t.Error("same pointer should return true")
	}
	// Equal but different pointers
	if !vm.containsIdentityOrEqual(MakeInt(999), MakeInt(999)) {
		t.Error("equal values should return true")
	}
	// Not equal
	if vm.containsIdentityOrEqual(MakeInt(1), MakeInt(2)) {
		t.Error("different values should return false")
	}
}

// =====================================
// compare: bool coercion in compare
// =====================================

func TestCompareBoolCoercion(t *testing.T) {
	vm := NewVM()
	// True (1) > False (0)
	cmp := vm.compare(True, False)
	if cmp != 1 {
		t.Errorf("compare(True, False) = %d, want 1", cmp)
	}
	// False (0) < MakeInt(1)
	cmp = vm.compare(False, MakeInt(1))
	if cmp != -1 {
		t.Errorf("compare(False, 1) = %d, want -1", cmp)
	}
	// True (1) == MakeInt(1)
	cmp = vm.compare(True, MakeInt(1))
	if cmp != 0 {
		t.Errorf("compare(True, 1) = %d, want 0", cmp)
	}
}

// =====================================
// compare: float vs int in compare
// =====================================

func TestCompareFloatVsInt(t *testing.T) {
	vm := NewVM()
	cmp := vm.compare(&PyFloat{Value: 1.5}, MakeInt(1))
	if cmp != 1 {
		t.Errorf("compare(1.5, 1) = %d, want 1", cmp)
	}
	cmp = vm.compare(&PyFloat{Value: 0.5}, MakeInt(1))
	if cmp != -1 {
		t.Errorf("compare(0.5, 1) = %d, want -1", cmp)
	}
	cmp = vm.compare(&PyFloat{Value: 1.0}, MakeInt(1))
	if cmp != 0 {
		t.Errorf("compare(1.0, 1) = %d, want 0", cmp)
	}
}

func TestCompareIntVsFloat(t *testing.T) {
	vm := NewVM()
	cmp := vm.compare(MakeInt(2), &PyFloat{Value: 1.5})
	if cmp != 1 {
		t.Errorf("compare(2, 1.5) = %d, want 1", cmp)
	}
	cmp = vm.compare(MakeInt(1), &PyFloat{Value: 1.5})
	if cmp != -1 {
		t.Errorf("compare(1, 1.5) = %d, want -1", cmp)
	}
	cmp = vm.compare(MakeInt(1), &PyFloat{Value: 1.0})
	if cmp != 0 {
		t.Errorf("compare(1, 1.0) = %d, want 0", cmp)
	}
}

// =====================================
// compare: string comparison
// =====================================

func TestCompareStrings(t *testing.T) {
	vm := NewVM()
	cmp := vm.compare(&PyString{Value: "abc"}, &PyString{Value: "abd"})
	if cmp != -1 {
		t.Errorf("compare('abc', 'abd') = %d, want -1", cmp)
	}
	cmp = vm.compare(&PyString{Value: "abd"}, &PyString{Value: "abc"})
	if cmp != 1 {
		t.Errorf("compare('abd', 'abc') = %d, want 1", cmp)
	}
	cmp = vm.compare(&PyString{Value: "abc"}, &PyString{Value: "abc"})
	if cmp != 0 {
		t.Errorf("compare('abc', 'abc') = %d, want 0", cmp)
	}
}
