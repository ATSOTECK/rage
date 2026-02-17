package runtime

import (
	"testing"
)

// =====================================
// MakeInt: Small Integer Cache
// =====================================

func TestMakeIntSmallCache(t *testing.T) {
	// Values -5 to 256 should return the same pointer
	for i := int64(-5); i <= 256; i++ {
		a := MakeInt(i)
		b := MakeInt(i)
		if a != b {
			t.Errorf("MakeInt(%d) returned different pointers for cached range", i)
		}
	}
}

func TestMakeIntOutsideCache(t *testing.T) {
	// Values outside cache should return different pointers
	a := MakeInt(1000)
	b := MakeInt(1000)
	if a == b {
		t.Error("MakeInt(1000) should return different pointers (outside cache)")
	}
	// But values should be equal
	if a.Value != b.Value {
		t.Error("values should be equal")
	}

	// Negative outside cache
	a = MakeInt(-10)
	b = MakeInt(-10)
	if a == b {
		t.Error("MakeInt(-10) should return different pointers (outside cache)")
	}
}

func TestMakeIntCacheValues(t *testing.T) {
	// Verify cache boundaries
	if MakeInt(-5).Value != -5 {
		t.Error("MakeInt(-5) value wrong")
	}
	if MakeInt(0).Value != 0 {
		t.Error("MakeInt(0) value wrong")
	}
	if MakeInt(256).Value != 256 {
		t.Error("MakeInt(256) value wrong")
	}
}

// =====================================
// isHashable
// =====================================

func TestIsHashable(t *testing.T) {
	tests := []struct {
		name string
		v    Value
		want bool
	}{
		{"int", MakeInt(1), true},
		{"float", &PyFloat{Value: 1.0}, true},
		{"string", &PyString{Value: "a"}, true},
		{"bool", True, true},
		{"None", None, true},
		{"tuple", &PyTuple{Items: []Value{MakeInt(1)}}, true},
		{"frozenset", &PyFrozenSet{Items: make(map[Value]struct{})}, true},
		{"bytes", &PyBytes{Value: []byte{1}}, true},
		{"complex", MakeComplex(1, 2), true},
		{"list", &PyList{Items: []Value{}}, false},
		{"dict", &PyDict{}, false},
		{"set", &PySet{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isHashable(tt.v)
			if got != tt.want {
				t.Errorf("isHashable(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

// =====================================
// hashValue: Consistency
// =====================================

func TestHashValueBoolIntConsistency(t *testing.T) {
	// hash(True) == hash(1)
	if hashValue(True) != hashValue(MakeInt(1)) {
		t.Error("hash(True) should equal hash(1)")
	}
	// hash(False) == hash(0)
	if hashValue(False) != hashValue(MakeInt(0)) {
		t.Error("hash(False) should equal hash(0)")
	}
}

func TestHashValueIntFloatConsistency(t *testing.T) {
	// hash(1) == hash(1.0)
	if hashValue(MakeInt(1)) != hashValue(&PyFloat{Value: 1.0}) {
		t.Error("hash(1) should equal hash(1.0)")
	}
	// hash(0) == hash(0.0)
	if hashValue(MakeInt(0)) != hashValue(&PyFloat{Value: 0.0}) {
		t.Error("hash(0) should equal hash(0.0)")
	}
	// hash(42) == hash(42.0)
	if hashValue(MakeInt(42)) != hashValue(&PyFloat{Value: 42.0}) {
		t.Error("hash(42) should equal hash(42.0)")
	}
}

func TestHashValueComplexConsistency(t *testing.T) {
	// hash(complex(1, 0)) == hash(1) for real-valued complex
	if hashValue(MakeComplex(1, 0)) != hashValue(MakeInt(1)) {
		t.Error("hash(complex(1, 0)) should equal hash(1)")
	}
}

func TestHashValueStrings(t *testing.T) {
	// Same string should have same hash
	h1 := hashValue(&PyString{Value: "hello"})
	h2 := hashValue(&PyString{Value: "hello"})
	if h1 != h2 {
		t.Error("hash('hello') should be consistent")
	}

	// Different strings should (usually) have different hash
	h3 := hashValue(&PyString{Value: "world"})
	if h1 == h3 {
		t.Log("hash collision for 'hello' and 'world' (unlikely but possible)")
	}
}

func TestHashValueTuples(t *testing.T) {
	a := &PyTuple{Items: []Value{MakeInt(1), MakeInt(2)}}
	b := &PyTuple{Items: []Value{MakeInt(1), MakeInt(2)}}
	if hashValue(a) != hashValue(b) {
		t.Error("hash of equal tuples should match")
	}

	c := &PyTuple{Items: []Value{MakeInt(2), MakeInt(1)}}
	if hashValue(a) == hashValue(c) {
		t.Log("hash collision for (1,2) and (2,1) - order matters")
	}
}

func TestHashValueNone(t *testing.T) {
	h := hashValue(None)
	if h == 0 {
		t.Error("hash(None) should not be 0")
	}
}

// =====================================
// MakeComplex / formatComplex
// =====================================

func TestMakeComplex(t *testing.T) {
	c := MakeComplex(1, 2)
	if c.Real != 1 || c.Imag != 2 {
		t.Errorf("MakeComplex(1, 2) = (%f, %f), want (1, 2)", c.Real, c.Imag)
	}

	c = MakeComplex(0, 0)
	if c.Real != 0 || c.Imag != 0 {
		t.Errorf("MakeComplex(0, 0) = (%f, %f), want (0, 0)", c.Real, c.Imag)
	}
}

func TestFormatComplex(t *testing.T) {
	tests := []struct {
		name       string
		real, imag float64
		want       string
	}{
		{"pure imaginary", 0, 3, "3j"},
		{"pure negative imag", 0, -1, "-1j"},
		{"full form", 1, 2, "(1+2j)"},
		{"negative imag", 1, -2, "(1-2j)"},
		{"zero imag", 1, 0, "(1+0j)"},
		{"pure zero", 0, 0, "0j"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatComplex(tt.real, tt.imag)
			if got != tt.want {
				t.Errorf("formatComplex(%f, %f) = %q, want %q", tt.real, tt.imag, got, tt.want)
			}
		})
	}
}

// =====================================
// PyRange
// =====================================

func TestPyRangeLen(t *testing.T) {
	tests := []struct {
		name              string
		start, stop, step int64
		want              int64
	}{
		{"range(5)", 0, 5, 1, 5},
		{"range(0)", 0, 0, 1, 0},
		{"range(1, 10)", 1, 10, 1, 9},
		{"range(0, 10, 2)", 0, 10, 2, 5},
		{"range(0, 10, 3)", 0, 10, 3, 4},
		{"range(10, 0, -1)", 10, 0, -1, 10},
		{"range(10, 0, -3)", 10, 0, -3, 4},
		{"range(5, 5, 1)", 5, 5, 1, 0},
		{"range(5, 0, 1)", 5, 0, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &PyRange{Start: tt.start, Stop: tt.stop, Step: tt.step}
			got := r.Len()
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPyRangeContains(t *testing.T) {
	tests := []struct {
		name              string
		start, stop, step int64
		val               int64
		want              bool
	}{
		{"in range(5)", 0, 5, 1, 3, true},
		{"start of range", 0, 5, 1, 0, true},
		{"end excluded", 0, 5, 1, 5, false},
		{"out of range", 0, 5, 1, 6, false},
		{"negative out", 0, 5, 1, -1, false},
		{"step match", 0, 10, 2, 4, true},
		{"step no match", 0, 10, 2, 3, false},
		{"negative step", 10, 0, -1, 5, true},
		{"neg step start", 10, 0, -1, 10, true},
		{"neg step end excluded", 10, 0, -1, 0, false},
		{"neg step out", 10, 0, -1, 11, false},
		{"neg step match", 10, 0, -2, 4, true},
		{"neg step no match", 10, 0, -2, 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &PyRange{Start: tt.start, Stop: tt.stop, Step: tt.step}
			got := r.Contains(tt.val)
			if got != tt.want {
				t.Errorf("range(%d,%d,%d).Contains(%d) = %v, want %v",
					tt.start, tt.stop, tt.step, tt.val, got, tt.want)
			}
		})
	}
}

// =====================================
// PyDict operations
// =====================================

func TestPyDictOperations(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}

	// Set and get
	d.DictSet(&PyString{Value: "a"}, MakeInt(1), vm)
	d.DictSet(&PyString{Value: "b"}, MakeInt(2), vm)
	if d.DictLen() != 2 {
		t.Errorf("expected len 2, got %d", d.DictLen())
	}

	val, ok := d.DictGet(&PyString{Value: "a"}, vm)
	if !ok || val.(*PyInt).Value != 1 {
		t.Error("expected to find 'a' = 1")
	}

	// Overwrite
	d.DictSet(&PyString{Value: "a"}, MakeInt(10), vm)
	val, ok = d.DictGet(&PyString{Value: "a"}, vm)
	if !ok || val.(*PyInt).Value != 10 {
		t.Error("expected 'a' = 10 after overwrite")
	}
	if d.DictLen() != 2 {
		t.Errorf("expected len 2 after overwrite, got %d", d.DictLen())
	}

	// Contains
	if !d.DictContains(&PyString{Value: "b"}, vm) {
		t.Error("expected 'b' to be in dict")
	}
	if d.DictContains(&PyString{Value: "c"}, vm) {
		t.Error("expected 'c' not to be in dict")
	}

	// Delete
	deleted := d.DictDelete(&PyString{Value: "a"}, vm)
	if !deleted {
		t.Error("expected delete to return true")
	}
	if d.DictLen() != 1 {
		t.Errorf("expected len 1 after delete, got %d", d.DictLen())
	}
	if d.DictContains(&PyString{Value: "a"}, vm) {
		t.Error("'a' should be deleted")
	}

	// Delete missing
	deleted = d.DictDelete(&PyString{Value: "missing"}, vm)
	if deleted {
		t.Error("expected delete of missing key to return false")
	}
}

// =====================================
// PySet operations
// =====================================

func TestPySetOperations(t *testing.T) {
	vm := NewVM()
	s := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}

	s.SetAdd(MakeInt(1), vm)
	s.SetAdd(MakeInt(2), vm)
	s.SetAdd(MakeInt(1), vm) // duplicate

	if s.SetLen() != 2 {
		t.Errorf("expected len 2, got %d", s.SetLen())
	}

	if !s.SetContains(MakeInt(1), vm) {
		t.Error("expected 1 to be in set")
	}
	if s.SetContains(MakeInt(3), vm) {
		t.Error("expected 3 not to be in set")
	}

	s.SetRemove(MakeInt(1), vm)
	if s.SetContains(MakeInt(1), vm) {
		t.Error("1 should be removed")
	}
	if s.SetLen() != 1 {
		t.Errorf("expected len 1, got %d", s.SetLen())
	}
}

// =====================================
// PyBool singleton
// =====================================

func TestPyBoolSingleton(t *testing.T) {
	if True.Value != true {
		t.Error("True.Value should be true")
	}
	if False.Value != false {
		t.Error("False.Value should be false")
	}
	if True == False {
		t.Error("True and False should be different pointers")
	}
}

// =====================================
// PyNone singleton
// =====================================

func TestPyNoneSingleton(t *testing.T) {
	if None == nil {
		t.Error("None should not be Go nil")
	}
	n1 := None
	n2 := None
	if n1 != n2 {
		t.Error("None should be singleton")
	}
}

// =====================================
// String interning
// =====================================

func TestInternString(t *testing.T) {
	a := InternString("hello")
	b := InternString("hello")
	if a != b {
		t.Error("InternString should return same pointer for same string")
	}
	if a.Value != "hello" {
		t.Errorf("got %q, want %q", a.Value, "hello")
	}
}
