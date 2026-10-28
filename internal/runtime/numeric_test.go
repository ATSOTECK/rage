package runtime

import (
	"math"
	"testing"
)

func TestIntBinaryOpPromotesOnOverflow(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		op   Opcode
		a, b int64
		want string
	}{
		{OpBinaryAdd, math.MaxInt64, 1, "9223372036854775808"},
		{OpBinarySubtract, math.MinInt64, 1, "-9223372036854775809"},
		{OpBinaryMultiply, math.MaxInt64, 2, "18446744073709551614"},
		{OpBinaryFloorDiv, math.MinInt64, -1, "9223372036854775808"},
		{OpBinaryModulo, math.MinInt64, -1, "0"},
		{OpBinaryPower, 2, 100, "1267650600228229401496703205376"},
		{OpBinaryLShift, 1, 70, "1180591620717411303424"},
		{OpBinaryFloorDiv, -7, 2, "-4"},
		{OpBinaryModulo, -7, 2, "1"},
	}
	for _, tt := range tests {
		got, err := vm.binaryOp(tt.op, MakeInt(tt.a), MakeInt(tt.b))
		if err != nil {
			t.Fatalf("%v %d %d: %v", tt.op, tt.a, tt.b, err)
		}
		if s := got.(*PyInt).String(); s != tt.want {
			t.Errorf("%v %d %d = %s, want %s", tt.op, tt.a, tt.b, s, tt.want)
		}
	}
}

func TestBigIntResultsNormalize(t *testing.T) {
	vm := NewVM()
	big, _ := vm.binaryOp(OpBinaryPower, MakeInt(2), MakeInt(64))
	back, _ := vm.binaryOp(OpBinaryFloorDiv, big, MakeInt(1<<32))
	if back.(*PyInt).IsBig() || back.(*PyInt).Value != 1<<32 {
		t.Errorf("2**64 // 2**32 = %v, want small int 4294967296", back)
	}
	if s := AddInts(MakeInt(math.MaxInt64), MakeInt(1)).String(); s != "9223372036854775808" {
		t.Errorf("AddInts overflow = %s", s)
	}
	if s := MulInts(MakeInt(math.MinInt64), MakeInt(-1)).String(); s != "9223372036854775808" {
		t.Errorf("MulInts overflow = %s", s)
	}
	if CompareInts(MakeInt(5), big.(*PyInt)) != -1 {
		t.Errorf("CompareInts(5, 2**64) should be -1")
	}
}

func TestHashMatchesCPython(t *testing.T) {
	tests := []struct {
		v    Value
		want int64
	}{
		{MakeInt(5), 5},
		{MakeInt(-1), -2},
		{&PyFloat{Value: 1.5}, 1152921504606846977},
		{&PyFloat{Value: 2.0}, 2},
		{True, 1},
		{None, 4238894112},
		{&PyTuple{Items: []Value{MakeInt(1), MakeInt(2)}}, -3550055125485641917},
		{MakeComplex(1, 2), 2000007},
	}
	for _, tt := range tests {
		if got := int64(hashValue(tt.v)); got != tt.want {
			t.Errorf("hash(%v) = %d, want %d", tt.v, got, tt.want)
		}
	}
}

func TestSetIterationOrderMatchesCPython(t *testing.T) {
	vm := NewVM()
	s := &PySet{}
	for _, v := range []int64{10, 3, 7, 100, 5} {
		s.SetAdd(MakeInt(v), vm)
	}
	want := []int64{3, 100, 5, 7, 10}
	got := s.Values()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i, v := range got {
		if v.(*PyInt).Value != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}
