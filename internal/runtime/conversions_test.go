package runtime

import (
	"math"
	"strings"
	"testing"
)

// =====================================
// tryToInt
// =====================================

func TestTryToInt(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name    string
		input   Value
		want    int64
		wantErr bool
	}{
		{"int", MakeInt(42), 42, false},
		{"int zero", MakeInt(0), 0, false},
		{"int negative", MakeInt(-10), -10, false},
		{"float truncate", &PyFloat{Value: 3.9}, 3, false},
		{"float negative", &PyFloat{Value: -2.7}, -2, false},
		{"bool true", True, 1, false},
		{"bool false", False, 0, false},
		{"string", &PyString{Value: "123"}, 123, false},
		{"string negative", &PyString{Value: "-456"}, -456, false},
		{"string spaces", &PyString{Value: "  42  "}, 42, false},
		{"string underscored", &PyString{Value: "1_000"}, 1000, false},
		{"complex error", MakeComplex(1, 2), 0, true},
		{"string empty", &PyString{Value: ""}, 0, true},
		{"string invalid", &PyString{Value: "abc"}, 0, true},
		{"string float", &PyString{Value: "3.14"}, 0, true},
		{"none error", None, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vm.tryToInt(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

// =====================================
// tryToFloat
// =====================================

func TestTryToFloat(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name    string
		input   Value
		want    float64
		wantErr bool
		isNaN   bool
		isInf   int // 0=not inf, 1=+inf, -1=-inf
	}{
		{"int", MakeInt(42), 42.0, false, false, 0},
		{"float", &PyFloat{Value: 3.14}, 3.14, false, false, 0},
		{"bool true", True, 1.0, false, false, 0},
		{"bool false", False, 0.0, false, false, 0},
		{"string", &PyString{Value: "3.14"}, 3.14, false, false, 0},
		{"string int", &PyString{Value: "42"}, 42.0, false, false, 0},
		{"string inf", &PyString{Value: "inf"}, 0, false, false, 1},
		{"string -inf", &PyString{Value: "-inf"}, 0, false, false, -1},
		{"string infinity", &PyString{Value: "infinity"}, 0, false, false, 1},
		{"string nan", &PyString{Value: "nan"}, 0, false, true, 0},
		{"complex error", MakeComplex(1, 2), 0, true, false, 0},
		{"string empty", &PyString{Value: ""}, 0, true, false, 0},
		{"string invalid", &PyString{Value: "abc"}, 0, true, false, 0},
		{"none error", None, 0, true, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vm.tryToFloat(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.isNaN {
				if !math.IsNaN(got) {
					t.Errorf("got %f, want NaN", got)
				}
				return
			}
			if tt.isInf != 0 {
				if !math.IsInf(got, tt.isInf) {
					t.Errorf("got %f, want Inf(%d)", got, tt.isInf)
				}
				return
			}
			if math.Abs(got-tt.want) > 1e-10 {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

// =====================================
// getIntIndex
// =====================================

func TestGetIntIndex(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name    string
		input   Value
		want    int64
		wantErr bool
	}{
		{"int", MakeInt(5), 5, false},
		{"int negative", MakeInt(-1), -1, false},
		{"bool true", True, 1, false},
		{"bool false", False, 0, false},
		{"float error", &PyFloat{Value: 1.0}, 0, true},
		{"string error", &PyString{Value: "1"}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vm.getIntIndex(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

// =====================================
// intFromStringBase
// =====================================

func TestIntFromStringBase(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name    string
		s       string
		base    int64
		want    int64
		wantErr bool
	}{
		{"base10", "42", 10, 42, false},
		{"base10 negative", "-42", 10, -42, false},
		{"base16", "ff", 16, 255, false},
		{"base16 prefix", "0xff", 16, 255, false},
		{"base8", "77", 8, 63, false},
		{"base8 prefix", "0o77", 8, 63, false},
		{"base2", "1010", 2, 10, false},
		{"base2 prefix", "0b1010", 2, 10, false},
		{"base0 hex", "0xff", 0, 255, false},
		{"base0 oct", "0o10", 0, 8, false},
		{"base0 bin", "0b10", 0, 2, false},
		{"base0 dec", "42", 0, 42, false},
		{"underscores", "1_000", 10, 1000, false},
		{"spaces", "  42  ", 10, 42, false},
		{"empty error", "", 10, 0, true},
		{"invalid base", "42", 1, 0, true},
		{"invalid base high", "42", 37, 0, true},
		{"invalid chars", "xyz", 10, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.intFromStringBase(tt.s, tt.base)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
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
// truthy
// =====================================

func TestTruthy(t *testing.T) {
	vm := NewVM()
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
		{"complex 0+0j", MakeComplex(0, 0), false},
		{"complex 1+0j", MakeComplex(1, 0), true},
		{"complex 0+1j", MakeComplex(0, 1), true},
		{"empty string", &PyString{Value: ""}, false},
		{"non-empty string", &PyString{Value: "a"}, true},
		{"empty list", &PyList{Items: []Value{}}, false},
		{"non-empty list", &PyList{Items: []Value{MakeInt(1)}}, true},
		{"empty tuple", &PyTuple{Items: []Value{}}, false},
		{"non-empty tuple", &PyTuple{Items: []Value{MakeInt(1)}}, true},
		{"empty dict", &PyDict{Items: make(map[Value]Value)}, false},
		{"empty set", &PySet{Items: make(map[Value]struct{})}, false},
		{"empty bytes", &PyBytes{Value: []byte{}}, false},
		{"non-empty bytes", &PyBytes{Value: []byte{1}}, true},
		{"range empty", &PyRange{Start: 0, Stop: 0, Step: 1}, false},
		{"range non-empty", &PyRange{Start: 0, Stop: 5, Step: 1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := vm.truthy(tt.v)
			if got != tt.want {
				t.Errorf("truthy(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

// =====================================
// str
// =====================================

func TestStr(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		v    Value
		want string
	}{
		{"None", None, "None"},
		{"True", True, "True"},
		{"False", False, "False"},
		{"int", MakeInt(42), "42"},
		{"int zero", MakeInt(0), "0"},
		{"int negative", MakeInt(-5), "-5"},
		{"float", &PyFloat{Value: 3.14}, "3.14"},
		{"float int-like", &PyFloat{Value: 1.0}, "1.0"},
		{"string", &PyString{Value: "hello"}, "hello"},
		{"empty string", &PyString{Value: ""}, ""},
		{"empty list", &PyList{Items: []Value{}}, "[]"},
		{"list", &PyList{Items: []Value{MakeInt(1), MakeInt(2)}}, "[1, 2]"},
		{"empty tuple", &PyTuple{Items: []Value{}}, "()"},
		{"single tuple", &PyTuple{Items: []Value{MakeInt(1)}}, "(1,)"},
		{"tuple", &PyTuple{Items: []Value{MakeInt(1), MakeInt(2)}}, "(1, 2)"},
		{"function", &PyFunction{Name: "foo"}, "<function foo>"},
		{"builtin", &PyBuiltinFunc{Name: "len"}, "<built-in function len>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := vm.str(tt.v)
			if got != tt.want {
				t.Errorf("str(%v) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}

func TestStrComplex(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name       string
		real, imag float64
		want       string
	}{
		{"pure imag", 0, 3, "3j"},
		{"full form", 1, 2, "(1+2j)"},
		{"negative imag", 1, -2, "(1-2j)"},
		{"pure neg imag", 0, -1, "-1j"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := vm.str(MakeComplex(tt.real, tt.imag))
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// =====================================
// repr
// =====================================

func TestRepr(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		v    Value
		want string
	}{
		{"string gets quotes", &PyString{Value: "hello"}, "'hello'"},
		{"None", None, "None"},
		{"True", True, "True"},
		{"False", False, "False"},
		{"int", MakeInt(42), "42"},
		{"float", &PyFloat{Value: 3.14}, "3.14"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := vm.repr(tt.v)
			if got != tt.want {
				t.Errorf("repr(%v) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}

func TestReprNestedContainers(t *testing.T) {
	vm := NewVM()
	// List with strings
	list := &PyList{Items: []Value{&PyString{Value: "a"}, &PyString{Value: "b"}}}
	got := vm.repr(list)
	if got != "['a', 'b']" {
		t.Errorf("got %q, want %q", got, "['a', 'b']")
	}

	// Tuple with strings
	tup := &PyTuple{Items: []Value{&PyString{Value: "x"}}}
	got = vm.repr(tup)
	if got != "('x',)" {
		t.Errorf("got %q, want %q", got, "('x',)")
	}
}

// =====================================
// typeName
// =====================================

func TestTypeName(t *testing.T) {
	vm := NewVM()
	tests := []struct {
		name string
		v    Value
		want string
	}{
		{"None", None, "NoneType"},
		{"bool", True, "bool"},
		{"int", MakeInt(1), "int"},
		{"float", &PyFloat{Value: 1.0}, "float"},
		{"complex", MakeComplex(1, 2), "complex"},
		{"str", &PyString{Value: ""}, "str"},
		{"bytes", &PyBytes{Value: nil}, "bytes"},
		{"list", &PyList{Items: nil}, "list"},
		{"tuple", &PyTuple{Items: nil}, "tuple"},
		{"dict", &PyDict{}, "dict"},
		{"set", &PySet{}, "set"},
		{"range", &PyRange{Start: 0, Stop: 1, Step: 1}, "range"},
		{"function", &PyFunction{Name: "f"}, "function"},
		{"builtin", &PyBuiltinFunc{Name: "len"}, "builtin_function_or_method"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := vm.typeName(tt.v)
			if got != tt.want {
				t.Errorf("typeName(%v) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}

// =====================================
// toList
// =====================================

func TestToList(t *testing.T) {
	vm := NewVM()

	// List passthrough
	items, err := vm.toList(&PyList{Items: []Value{MakeInt(1), MakeInt(2)}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Tuple
	items, err = vm.toList(&PyTuple{Items: []Value{MakeInt(3)}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].(*PyInt).Value != 3 {
		t.Error("expected [3]")
	}

	// String
	items, err = vm.toList(&PyString{Value: "abc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0].(*PyString).Value != "a" {
		t.Error("expected first char 'a'")
	}

	// Range
	items, err = vm.toList(&PyRange{Start: 0, Stop: 3, Step: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// Bytes
	items, err = vm.toList(&PyBytes{Value: []byte{65, 66, 67}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0].(*PyInt).Value != 65 {
		t.Error("expected first byte 65")
	}

	// Not iterable
	_, err = vm.toList(MakeInt(42))
	if err == nil {
		t.Fatal("expected error for non-iterable")
	}
	if !strings.Contains(err.Error(), "not iterable") {
		t.Errorf("expected 'not iterable' error, got: %v", err)
	}
}
