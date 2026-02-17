package runtime

import (
	"math"
	"testing"
)

// =====================================
// Binary Op: Integer Arithmetic
// =====================================

func TestBinaryOpIntArithmetic(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		op   Opcode
		a, b int64
		want int64
	}{
		{"add", OpBinaryAdd, 3, 4, 7},
		{"add negative", OpBinaryAdd, -3, 4, 1},
		{"sub", OpBinarySubtract, 10, 3, 7},
		{"sub negative result", OpBinarySubtract, 3, 10, -7},
		{"mul", OpBinaryMultiply, 6, 7, 42},
		{"mul zero", OpBinaryMultiply, 100, 0, 0},
		{"mul negative", OpBinaryMultiply, -3, 4, -12},
		{"floordiv", OpBinaryFloorDiv, 10, 3, 3},
		{"floordiv exact", OpBinaryFloorDiv, 9, 3, 3},
		{"floordiv negative", OpBinaryFloorDiv, -7, 2, -4},
		{"mod", OpBinaryModulo, 10, 3, 1},
		{"mod no remainder", OpBinaryModulo, 9, 3, 0},
		{"pow", OpBinaryPower, 2, 10, 1024},
		{"pow zero", OpBinaryPower, 5, 0, 1},
		{"lshift", OpBinaryLShift, 1, 8, 256},
		{"rshift", OpBinaryRShift, 256, 4, 16},
		{"bitand", OpBinaryAnd, 0xFF, 0x0F, 0x0F},
		{"bitor", OpBinaryOr, 0xF0, 0x0F, 0xFF},
		{"bitxor", OpBinaryXor, 0xFF, 0x0F, 0xF0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.binaryOp(tt.op, MakeInt(tt.a), MakeInt(tt.b))
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

func TestBinaryOpIntDivisionByZero(t *testing.T) {
	vm := NewVM()
	ops := []struct {
		name string
		op   Opcode
	}{
		{"true divide", OpBinaryDivide},
		{"floor divide", OpBinaryFloorDiv},
		{"modulo", OpBinaryModulo},
	}

	for _, tt := range ops {
		t.Run(tt.name, func(t *testing.T) {
			_, err := vm.binaryOp(tt.op, MakeInt(10), MakeInt(0))
			if err == nil {
				t.Fatal("expected division by zero error")
			}
		})
	}
}

func TestBinaryOpTrueDivide(t *testing.T) {
	vm := NewVM()
	result, err := vm.binaryOp(OpBinaryDivide, MakeInt(7), MakeInt(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyFloat)
	if !ok {
		t.Fatalf("expected *PyFloat, got %T", result)
	}
	if got.Value != 3.5 {
		t.Errorf("got %f, want 3.5", got.Value)
	}
}

// =====================================
// Binary Op: Float Arithmetic
// =====================================

func TestBinaryOpFloatArithmetic(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		op   Opcode
		a, b float64
		want float64
	}{
		{"add", OpBinaryAdd, 1.5, 2.5, 4.0},
		{"sub", OpBinarySubtract, 5.5, 3.0, 2.5},
		{"mul", OpBinaryMultiply, 2.0, 3.5, 7.0},
		{"div", OpBinaryDivide, 7.0, 2.0, 3.5},
		{"floordiv", OpBinaryFloorDiv, 7.5, 2.0, 3.0},
		{"mod", OpBinaryModulo, 7.5, 2.0, 1.5},
		{"pow", OpBinaryPower, 2.0, 3.0, 8.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.binaryOp(tt.op, &PyFloat{Value: tt.a}, &PyFloat{Value: tt.b})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, ok := result.(*PyFloat)
			if !ok {
				t.Fatalf("expected *PyFloat, got %T", result)
			}
			if math.Abs(got.Value-tt.want) > 1e-10 {
				t.Errorf("got %f, want %f", got.Value, tt.want)
			}
		})
	}
}

func TestBinaryOpFloatDivisionByZero(t *testing.T) {
	vm := NewVM()
	_, err := vm.binaryOp(OpBinaryDivide, &PyFloat{Value: 1.0}, &PyFloat{Value: 0.0})
	if err == nil {
		t.Fatal("expected division by zero error")
	}
}

// =====================================
// Binary Op: Mixed Int/Float Promotion
// =====================================

func TestBinaryOpMixedIntFloat(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		op   Opcode
		a    Value
		b    Value
		want float64
	}{
		{"int+float", OpBinaryAdd, MakeInt(1), &PyFloat{Value: 2.5}, 3.5},
		{"float+int", OpBinaryAdd, &PyFloat{Value: 2.5}, MakeInt(1), 3.5},
		{"int*float", OpBinaryMultiply, MakeInt(3), &PyFloat{Value: 2.5}, 7.5},
		{"int-float", OpBinarySubtract, MakeInt(5), &PyFloat{Value: 1.5}, 3.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.binaryOp(tt.op, tt.a, tt.b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, ok := result.(*PyFloat)
			if !ok {
				t.Fatalf("expected *PyFloat, got %T", result)
			}
			if math.Abs(got.Value-tt.want) > 1e-10 {
				t.Errorf("got %f, want %f", got.Value, tt.want)
			}
		})
	}
}

// =====================================
// Binary Op: String Operations
// =====================================

func TestBinaryOpStringConcat(t *testing.T) {
	vm := NewVM()
	result, err := vm.binaryOp(OpBinaryAdd, &PyString{Value: "hello "}, &PyString{Value: "world"})
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

func TestBinaryOpStringRepetition(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		a    Value
		b    Value
		want string
	}{
		{"str*int", &PyString{Value: "ab"}, MakeInt(3), "ababab"},
		{"int*str", MakeInt(2), &PyString{Value: "xy"}, "xyxy"},
		{"str*0", &PyString{Value: "ab"}, MakeInt(0), ""},
		{"str*neg", &PyString{Value: "ab"}, MakeInt(-1), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.binaryOp(OpBinaryMultiply, tt.a, tt.b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, ok := result.(*PyString)
			if !ok {
				t.Fatalf("expected *PyString, got %T", result)
			}
			if got.Value != tt.want {
				t.Errorf("got %q, want %q", got.Value, tt.want)
			}
		})
	}
}

// =====================================
// Binary Op: List/Tuple Operations
// =====================================

func TestBinaryOpListConcat(t *testing.T) {
	vm := NewVM()
	a := &PyList{Items: []Value{MakeInt(1), MakeInt(2)}}
	b := &PyList{Items: []Value{MakeInt(3), MakeInt(4)}}
	result, err := vm.binaryOp(OpBinaryAdd, a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyList)
	if !ok {
		t.Fatalf("expected *PyList, got %T", result)
	}
	if len(got.Items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(got.Items))
	}
	for i, want := range []int64{1, 2, 3, 4} {
		if v := got.Items[i].(*PyInt).Value; v != want {
			t.Errorf("item[%d] = %d, want %d", i, v, want)
		}
	}
}

func TestBinaryOpListRepetition(t *testing.T) {
	vm := NewVM()
	a := &PyList{Items: []Value{MakeInt(1), MakeInt(2)}}
	result, err := vm.binaryOp(OpBinaryMultiply, a, MakeInt(3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyList)
	if !ok {
		t.Fatalf("expected *PyList, got %T", result)
	}
	if len(got.Items) != 6 {
		t.Fatalf("expected 6 items, got %d", len(got.Items))
	}
}

func TestBinaryOpTupleConcat(t *testing.T) {
	vm := NewVM()
	a := &PyTuple{Items: []Value{MakeInt(1)}}
	b := &PyTuple{Items: []Value{MakeInt(2)}}
	result, err := vm.binaryOp(OpBinaryAdd, a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyTuple)
	if !ok {
		t.Fatalf("expected *PyTuple, got %T", result)
	}
	if len(got.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got.Items))
	}
}

func TestBinaryOpTupleRepetition(t *testing.T) {
	vm := NewVM()
	a := &PyTuple{Items: []Value{MakeInt(1)}}
	result, err := vm.binaryOp(OpBinaryMultiply, a, MakeInt(3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyTuple)
	if !ok {
		t.Fatalf("expected *PyTuple, got %T", result)
	}
	if len(got.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(got.Items))
	}
}

// =====================================
// Binary Op: Complex Arithmetic
// =====================================

func TestBinaryOpComplexArithmetic(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name     string
		op       Opcode
		a, b     Value
		wantReal float64
		wantImag float64
	}{
		{"add", OpBinaryAdd, MakeComplex(1, 2), MakeComplex(3, 4), 4, 6},
		{"sub", OpBinarySubtract, MakeComplex(5, 7), MakeComplex(3, 4), 2, 3},
		{"mul", OpBinaryMultiply, MakeComplex(1, 2), MakeComplex(3, 4), -5, 10},
		{"int+complex", OpBinaryAdd, MakeInt(1), MakeComplex(2, 3), 3, 3},
		{"float+complex", OpBinaryAdd, &PyFloat{Value: 1.5}, MakeComplex(2, 3), 3.5, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.binaryOp(tt.op, tt.a, tt.b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, ok := result.(*PyComplex)
			if !ok {
				t.Fatalf("expected *PyComplex, got %T", result)
			}
			if math.Abs(got.Real-tt.wantReal) > 1e-10 {
				t.Errorf("real: got %f, want %f", got.Real, tt.wantReal)
			}
			if math.Abs(got.Imag-tt.wantImag) > 1e-10 {
				t.Errorf("imag: got %f, want %f", got.Imag, tt.wantImag)
			}
		})
	}
}

// =====================================
// Binary Op: Bool Coercion
// =====================================

func TestBinaryOpBoolCoercion(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		op   Opcode
		a, b Value
		want int64
	}{
		{"True+True", OpBinaryAdd, True, True, 2},
		{"True+1", OpBinaryAdd, True, MakeInt(1), 2},
		{"False+0", OpBinaryAdd, False, MakeInt(0), 0},
		{"True*5", OpBinaryMultiply, True, MakeInt(5), 5},
		{"False*5", OpBinaryMultiply, False, MakeInt(5), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.binaryOp(tt.op, tt.a, tt.b)
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

// =====================================
// Binary Op: Type Mismatch Errors
// =====================================

func TestBinaryOpTypeMismatch(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		op   Opcode
		a, b Value
	}{
		{"str+int", OpBinaryAdd, &PyString{Value: "a"}, MakeInt(1)},
		{"int+str", OpBinaryAdd, MakeInt(1), &PyString{Value: "a"}},
		{"str-str", OpBinarySubtract, &PyString{Value: "a"}, &PyString{Value: "b"}},
		{"list+int", OpBinaryAdd, &PyList{Items: []Value{}}, MakeInt(1)},
		{"list-list", OpBinarySubtract, &PyList{Items: []Value{}}, &PyList{Items: []Value{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := vm.binaryOp(tt.op, tt.a, tt.b)
			if err == nil {
				t.Fatal("expected unsupported operand error")
			}
		})
	}
}

// =====================================
// Unary Operations
// =====================================

func TestUnaryOpNegative(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		a    Value
		want Value
	}{
		{"int", MakeInt(5), MakeInt(-5)},
		{"neg int", MakeInt(-3), MakeInt(3)},
		{"zero", MakeInt(0), MakeInt(0)},
		{"float", &PyFloat{Value: 3.14}, &PyFloat{Value: -3.14}},
		{"bool True", True, MakeInt(-1)},
		{"bool False", False, MakeInt(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.unaryOp(OpUnaryNegative, tt.a)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			switch want := tt.want.(type) {
			case *PyInt:
				got, ok := result.(*PyInt)
				if !ok {
					t.Fatalf("expected *PyInt, got %T", result)
				}
				if got.Value != want.Value {
					t.Errorf("got %d, want %d", got.Value, want.Value)
				}
			case *PyFloat:
				got, ok := result.(*PyFloat)
				if !ok {
					t.Fatalf("expected *PyFloat, got %T", result)
				}
				if got.Value != want.Value {
					t.Errorf("got %f, want %f", got.Value, want.Value)
				}
			}
		})
	}
}

func TestUnaryOpNegativeComplex(t *testing.T) {
	vm := NewVM()
	result, err := vm.unaryOp(OpUnaryNegative, MakeComplex(1, 2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := result.(*PyComplex)
	if !ok {
		t.Fatalf("expected *PyComplex, got %T", result)
	}
	if got.Real != -1 || got.Imag != -2 {
		t.Errorf("got (%f, %f), want (-1, -2)", got.Real, got.Imag)
	}
}

func TestUnaryOpInvert(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		a    Value
		want int64
	}{
		{"~0", MakeInt(0), -1},
		{"~42", MakeInt(42), -43},
		{"~-1", MakeInt(-1), 0},
		{"~True", True, -2},
		{"~False", False, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.unaryOp(OpUnaryInvert, tt.a)
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

func TestUnaryOpInvertFloat(t *testing.T) {
	vm := NewVM()
	_, err := vm.unaryOp(OpUnaryInvert, &PyFloat{Value: 1.0})
	if err == nil {
		t.Fatal("expected error for ~float")
	}
}

func TestUnaryOpPositive(t *testing.T) {
	vm := NewVM()
	result, err := vm.unaryOp(OpUnaryPositive, MakeInt(5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := result.(*PyInt)
	if got.Value != 5 {
		t.Errorf("got %d, want 5", got.Value)
	}
}

// =====================================
// Comparison Operations
// =====================================

func TestCompareOpInts(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		op   Opcode
		a, b int64
		want bool
	}{
		{"eq true", OpCompareEq, 5, 5, true},
		{"eq false", OpCompareEq, 5, 6, false},
		{"ne true", OpCompareNe, 5, 6, true},
		{"ne false", OpCompareNe, 5, 5, false},
		{"lt true", OpCompareLt, 3, 5, true},
		{"lt false", OpCompareLt, 5, 3, false},
		{"lt equal", OpCompareLt, 5, 5, false},
		{"le true", OpCompareLe, 3, 5, true},
		{"le equal", OpCompareLe, 5, 5, true},
		{"le false", OpCompareLe, 6, 5, false},
		{"gt true", OpCompareGt, 5, 3, true},
		{"gt false", OpCompareGt, 3, 5, false},
		{"ge true", OpCompareGe, 5, 3, true},
		{"ge equal", OpCompareGe, 5, 5, true},
		{"ge false", OpCompareGe, 3, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.compareOp(tt.op, MakeInt(tt.a), MakeInt(tt.b))
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

func TestCompareOpFloats(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		op   Opcode
		a, b float64
		want bool
	}{
		{"eq", OpCompareEq, 3.14, 3.14, true},
		{"ne", OpCompareNe, 3.14, 2.71, true},
		{"lt", OpCompareLt, 2.71, 3.14, true},
		{"gt", OpCompareGt, 3.14, 2.71, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.compareOp(tt.op, &PyFloat{Value: tt.a}, &PyFloat{Value: tt.b})
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

func TestCompareOpMixedIntFloat(t *testing.T) {
	vm := NewVM()
	// 1 == 1.0 should be true
	result := vm.compareOp(OpCompareEq, MakeInt(1), &PyFloat{Value: 1.0})
	got, ok := result.(*PyBool)
	if !ok {
		t.Fatalf("expected *PyBool, got %T", result)
	}
	if !got.Value {
		t.Error("expected 1 == 1.0 to be true")
	}

	// 1 < 1.5 should be true
	result = vm.compareOp(OpCompareLt, MakeInt(1), &PyFloat{Value: 1.5})
	got = result.(*PyBool)
	if !got.Value {
		t.Error("expected 1 < 1.5 to be true")
	}
}

func TestCompareOpStrings(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		op   Opcode
		a, b string
		want bool
	}{
		{"eq true", OpCompareEq, "abc", "abc", true},
		{"eq false", OpCompareEq, "abc", "def", false},
		{"lt true", OpCompareLt, "abc", "abd", true},
		{"lt false", OpCompareLt, "abd", "abc", false},
		{"gt true", OpCompareGt, "b", "a", true},
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

// =====================================
// Equality: Cross-type
// =====================================

func TestEqualCrossType(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		a, b Value
		want bool
	}{
		{"True == 1", True, MakeInt(1), true},
		{"False == 0", False, MakeInt(0), true},
		{"1 == 1.0", MakeInt(1), &PyFloat{Value: 1.0}, true},
		{"'1' != 1", &PyString{Value: "1"}, MakeInt(1), false},
		{"None == None", None, None, true},
		{"None != 0", None, MakeInt(0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := vm.equal(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("equal(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestEqualCollections(t *testing.T) {
	vm := NewVM()

	// Lists
	a := &PyList{Items: []Value{MakeInt(1), MakeInt(2), MakeInt(3)}}
	b := &PyList{Items: []Value{MakeInt(1), MakeInt(2), MakeInt(3)}}
	c := &PyList{Items: []Value{MakeInt(1), MakeInt(2)}}
	if !vm.equal(a, b) {
		t.Error("[1,2,3] should equal [1,2,3]")
	}
	if vm.equal(a, c) {
		t.Error("[1,2,3] should not equal [1,2]")
	}

	// Tuples
	ta := &PyTuple{Items: []Value{MakeInt(1), MakeInt(2)}}
	tb := &PyTuple{Items: []Value{MakeInt(1), MakeInt(2)}}
	if !vm.equal(ta, tb) {
		t.Error("(1,2) should equal (1,2)")
	}

	// Nested lists
	na := &PyList{Items: []Value{&PyList{Items: []Value{MakeInt(1)}}}}
	nb := &PyList{Items: []Value{&PyList{Items: []Value{MakeInt(1)}}}}
	if !vm.equal(na, nb) {
		t.Error("[[1]] should equal [[1]]")
	}
}

// =====================================
// Containment: "in" operator
// =====================================

func TestContainsList(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{MakeInt(1), MakeInt(2), MakeInt(3)}}
	if !vm.contains(list, MakeInt(2)) {
		t.Error("2 should be in [1,2,3]")
	}
	if vm.contains(list, MakeInt(4)) {
		t.Error("4 should not be in [1,2,3]")
	}
}

func TestContainsString(t *testing.T) {
	vm := NewVM()
	s := &PyString{Value: "hello world"}
	if !vm.contains(s, &PyString{Value: "hello"}) {
		t.Error("'hello' should be in 'hello world'")
	}
	if vm.contains(s, &PyString{Value: "xyz"}) {
		t.Error("'xyz' should not be in 'hello world'")
	}
}

func TestContainsTuple(t *testing.T) {
	vm := NewVM()
	tup := &PyTuple{Items: []Value{&PyString{Value: "a"}, &PyString{Value: "b"}}}
	if !vm.contains(tup, &PyString{Value: "a"}) {
		t.Error("'a' should be in ('a', 'b')")
	}
	if vm.contains(tup, &PyString{Value: "c"}) {
		t.Error("'c' should not be in ('a', 'b')")
	}
}

func TestContainsDict(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	d.DictSet(&PyString{Value: "key"}, MakeInt(1), vm)
	if !vm.contains(d, &PyString{Value: "key"}) {
		t.Error("'key' should be in dict")
	}
	if vm.contains(d, &PyString{Value: "missing"}) {
		t.Error("'missing' should not be in dict")
	}
}

func TestContainsSet(t *testing.T) {
	vm := NewVM()
	s := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
	s.SetAdd(MakeInt(42), vm)
	if !vm.contains(s, MakeInt(42)) {
		t.Error("42 should be in set")
	}
	if vm.contains(s, MakeInt(99)) {
		t.Error("99 should not be in set")
	}
}

func TestContainsRange(t *testing.T) {
	vm := NewVM()
	r := &PyRange{Start: 0, Stop: 10, Step: 2}
	if !vm.contains(r, MakeInt(4)) {
		t.Error("4 should be in range(0, 10, 2)")
	}
	if vm.contains(r, MakeInt(3)) {
		t.Error("3 should not be in range(0, 10, 2)")
	}
	if vm.contains(r, MakeInt(10)) {
		t.Error("10 should not be in range(0, 10, 2)")
	}
}

func TestContainsBytes(t *testing.T) {
	vm := NewVM()
	b := &PyBytes{Value: []byte{1, 2, 3, 4}}
	if !vm.contains(b, MakeInt(3)) {
		t.Error("3 should be in bytes")
	}
	if vm.contains(b, MakeInt(5)) {
		t.Error("5 should not be in bytes")
	}
}

// =====================================
// Is / Is Not
// =====================================

func TestCompareIs(t *testing.T) {
	vm := NewVM()
	// Same object
	result := vm.compareOp(OpCompareIs, None, None)
	if !result.(*PyBool).Value {
		t.Error("None is None should be True")
	}

	// Different objects
	a := MakeInt(1000)
	b := MakeInt(1000)
	result = vm.compareOp(OpCompareIsNot, a, b)
	// They are different pointers (1000 > 256, outside cache)
	if !result.(*PyBool).Value {
		t.Error("MakeInt(1000) should not be MakeInt(1000) (different pointers)")
	}

	// Cached ints are the same pointer
	c := MakeInt(5)
	d := MakeInt(5)
	result = vm.compareOp(OpCompareIs, c, d)
	if !result.(*PyBool).Value {
		t.Error("MakeInt(5) is MakeInt(5) should be True (cached)")
	}
}

// =====================================
// Compare In / Not In via compareOp
// =====================================

func TestCompareInNotIn(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{MakeInt(1), MakeInt(2), MakeInt(3)}}

	result := vm.compareOp(OpCompareIn, MakeInt(2), list)
	if !result.(*PyBool).Value {
		t.Error("2 in [1,2,3] should be True")
	}

	result = vm.compareOp(OpCompareNotIn, MakeInt(4), list)
	if !result.(*PyBool).Value {
		t.Error("4 not in [1,2,3] should be True")
	}
}
