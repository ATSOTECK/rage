package runtime

import (
	"strings"
	"testing"
)

// =====================================
// getItem: List
// =====================================

func TestGetItemList(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{MakeInt(10), MakeInt(20), MakeInt(30)}}

	tests := []struct {
		name  string
		index Value
		want  int64
	}{
		{"index 0", MakeInt(0), 10},
		{"index 1", MakeInt(1), 20},
		{"index 2", MakeInt(2), 30},
		{"index -1", MakeInt(-1), 30},
		{"index -3", MakeInt(-3), 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.getItem(list, tt.index)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := result.(*PyInt).Value
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetItemListOutOfRange(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{MakeInt(1)}}
	_, err := vm.getItem(list, MakeInt(5))
	if err == nil {
		t.Fatal("expected index out of range error")
	}
}

func TestGetItemListWrongType(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{MakeInt(1)}}
	_, err := vm.getItem(list, &PyString{Value: "a"})
	if err == nil {
		t.Fatal("expected type error for string index on list")
	}
}

// =====================================
// getItem: Tuple
// =====================================

func TestGetItemTuple(t *testing.T) {
	vm := NewVM()
	tup := &PyTuple{Items: []Value{&PyString{Value: "a"}, &PyString{Value: "b"}, &PyString{Value: "c"}}}

	result, err := vm.getItem(tup, MakeInt(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.(*PyString).Value != "b" {
		t.Errorf("got %v, want 'b'", result)
	}

	// Negative index
	result, err = vm.getItem(tup, MakeInt(-1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.(*PyString).Value != "c" {
		t.Errorf("got %v, want 'c'", result)
	}
}

func TestGetItemTupleOutOfRange(t *testing.T) {
	vm := NewVM()
	tup := &PyTuple{Items: []Value{MakeInt(1)}}
	_, err := vm.getItem(tup, MakeInt(5))
	if err == nil {
		t.Fatal("expected index out of range error")
	}
}

// =====================================
// getItem: String
// =====================================

func TestGetItemString(t *testing.T) {
	vm := NewVM()
	s := &PyString{Value: "hello"}

	result, err := vm.getItem(s, MakeInt(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.(*PyString).Value != "h" {
		t.Errorf("got %q, want 'h'", result.(*PyString).Value)
	}

	result, err = vm.getItem(s, MakeInt(-1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.(*PyString).Value != "o" {
		t.Errorf("got %q, want 'o'", result.(*PyString).Value)
	}
}

func TestGetItemStringOutOfRange(t *testing.T) {
	vm := NewVM()
	s := &PyString{Value: "hi"}
	_, err := vm.getItem(s, MakeInt(5))
	if err == nil {
		t.Fatal("expected index out of range error")
	}
}

// =====================================
// getItem: Bytes
// =====================================

func TestGetItemBytes(t *testing.T) {
	vm := NewVM()
	b := &PyBytes{Value: []byte{65, 66, 67}}

	result, err := vm.getItem(b, MakeInt(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.(*PyInt).Value != 65 {
		t.Errorf("got %v, want 65", result)
	}
}

// =====================================
// getItem: Dict
// =====================================

func TestGetItemDict(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	d.DictSet(&PyString{Value: "key"}, MakeInt(42), vm)

	result, err := vm.getItem(d, &PyString{Value: "key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.(*PyInt).Value != 42 {
		t.Errorf("got %v, want 42", result)
	}
}

func TestGetItemDictMissing(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	_, err := vm.getItem(d, &PyString{Value: "missing"})
	if err == nil {
		t.Fatal("expected KeyError")
	}
}

// =====================================
// setItem: List
// =====================================

func TestSetItemList(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{MakeInt(1), MakeInt(2), MakeInt(3)}}

	err := vm.setItem(list, MakeInt(1), MakeInt(99))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if list.Items[1].(*PyInt).Value != 99 {
		t.Errorf("got %v, want 99", list.Items[1])
	}

	// Negative index
	err = vm.setItem(list, MakeInt(-1), MakeInt(88))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if list.Items[2].(*PyInt).Value != 88 {
		t.Errorf("got %v, want 88", list.Items[2])
	}
}

func TestSetItemListOutOfRange(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{MakeInt(1)}}
	err := vm.setItem(list, MakeInt(5), MakeInt(99))
	if err == nil {
		t.Fatal("expected index out of range error")
	}
}

// =====================================
// setItem: Dict
// =====================================

func TestSetItemDict(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}

	err := vm.setItem(d, &PyString{Value: "key"}, MakeInt(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := d.DictGet(&PyString{Value: "key"}, vm)
	if !ok {
		t.Fatal("key not found after setItem")
	}
	if val.(*PyInt).Value != 42 {
		t.Errorf("got %v, want 42", val)
	}
}

// =====================================
// delItem: List
// =====================================

func TestDelItemList(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{MakeInt(1), MakeInt(2), MakeInt(3)}}

	err := vm.delItem(list, MakeInt(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list.Items))
	}
	if list.Items[0].(*PyInt).Value != 1 || list.Items[1].(*PyInt).Value != 3 {
		t.Error("unexpected items after deletion")
	}
}

func TestDelItemListOutOfRange(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{MakeInt(1)}}
	err := vm.delItem(list, MakeInt(5))
	if err == nil {
		t.Fatal("expected index out of range error")
	}
}

// =====================================
// delItem: Dict
// =====================================

func TestDelItemDict(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	d.DictSet(&PyString{Value: "key"}, MakeInt(1), vm)

	err := vm.delItem(d, &PyString{Value: "key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.DictContains(&PyString{Value: "key"}, vm) {
		t.Error("key should be deleted")
	}
}

func TestDelItemDictMissing(t *testing.T) {
	vm := NewVM()
	d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	err := vm.delItem(d, &PyString{Value: "missing"})
	if err == nil {
		t.Fatal("expected KeyError for missing key")
	}
}

// =====================================
// Slice: List
// =====================================

func TestSliceList(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{MakeInt(0), MakeInt(1), MakeInt(2), MakeInt(3), MakeInt(4)}}

	tests := []struct {
		name       string
		start      Value
		stop       Value
		step       Value
		wantValues []int64
	}{
		{"[1:3]", MakeInt(1), MakeInt(3), nil, []int64{1, 2}},
		{"[:2]", nil, MakeInt(2), nil, []int64{0, 1}},
		{"[3:]", MakeInt(3), nil, nil, []int64{3, 4}},
		{"[::2]", nil, nil, MakeInt(2), []int64{0, 2, 4}},
		{"[::-1]", nil, nil, MakeInt(-1), []int64{4, 3, 2, 1, 0}},
		{"[-2:]", MakeInt(-2), nil, nil, []int64{3, 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slice := &PySlice{Start: tt.start, Stop: tt.stop, Step: tt.step}
			result, err := vm.sliceSequence(list, slice)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, ok := result.(*PyList)
			if !ok {
				t.Fatalf("expected *PyList, got %T", result)
			}
			if len(got.Items) != len(tt.wantValues) {
				t.Fatalf("expected %d items, got %d", len(tt.wantValues), len(got.Items))
			}
			for i, want := range tt.wantValues {
				if got.Items[i].(*PyInt).Value != want {
					t.Errorf("item[%d] = %d, want %d", i, got.Items[i].(*PyInt).Value, want)
				}
			}
		})
	}
}

// =====================================
// Slice: String
// =====================================

func TestSliceString(t *testing.T) {
	vm := NewVM()
	s := &PyString{Value: "hello"}

	tests := []struct {
		name  string
		start Value
		stop  Value
		step  Value
		want  string
	}{
		{"[1:3]", MakeInt(1), MakeInt(3), nil, "el"},
		{"[::-1]", nil, nil, MakeInt(-1), "olleh"},
		{"[::2]", nil, nil, MakeInt(2), "hlo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slice := &PySlice{Start: tt.start, Stop: tt.stop, Step: tt.step}
			result, err := vm.sliceSequence(s, slice)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := result.(*PyString).Value
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// =====================================
// Slice: Tuple
// =====================================

func TestSliceTuple(t *testing.T) {
	vm := NewVM()
	tup := &PyTuple{Items: []Value{MakeInt(10), MakeInt(20), MakeInt(30), MakeInt(40)}}

	slice := &PySlice{Start: MakeInt(1), Stop: MakeInt(3), Step: nil}
	result, err := vm.sliceSequence(tup, slice)
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

// =====================================
// Slice: Bytes
// =====================================

func TestSliceBytes(t *testing.T) {
	vm := NewVM()
	b := &PyBytes{Value: []byte{1, 2, 3, 4, 5}}

	slice := &PySlice{Start: MakeInt(1), Stop: MakeInt(4), Step: nil}
	result, err := vm.sliceSequence(b, slice)
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
	if got.Value[0] != 2 || got.Value[1] != 3 || got.Value[2] != 4 {
		t.Errorf("got %v, want [2,3,4]", got.Value)
	}
}

// =====================================
// Slice step 0 error
// =====================================

func TestSliceStepZero(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{MakeInt(1), MakeInt(2)}}
	slice := &PySlice{Start: nil, Stop: nil, Step: MakeInt(0)}
	_, err := vm.sliceSequence(list, slice)
	if err == nil {
		t.Fatal("expected slice step cannot be zero error")
	}
	if !strings.Contains(err.Error(), "step cannot be zero") {
		t.Errorf("unexpected error: %v", err)
	}
}

// =====================================
// setSlice: List
// =====================================

func TestSetSliceList(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{MakeInt(0), MakeInt(1), MakeInt(2), MakeInt(3), MakeInt(4)}}

	// Replace [1:3] with [10, 20, 30] (different length)
	slice := &PySlice{Start: MakeInt(1), Stop: MakeInt(3), Step: nil}
	replacement := &PyList{Items: []Value{MakeInt(10), MakeInt(20), MakeInt(30)}}
	err := vm.setSlice(list, slice, replacement)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 6 {
		t.Fatalf("expected 6 items, got %d", len(list.Items))
	}
	expected := []int64{0, 10, 20, 30, 3, 4}
	for i, want := range expected {
		if list.Items[i].(*PyInt).Value != want {
			t.Errorf("item[%d] = %d, want %d", i, list.Items[i].(*PyInt).Value, want)
		}
	}
}

// =====================================
// delSlice: List
// =====================================

func TestDelSliceList(t *testing.T) {
	vm := NewVM()
	list := &PyList{Items: []Value{MakeInt(0), MakeInt(1), MakeInt(2), MakeInt(3), MakeInt(4)}}

	// Delete [1:3]
	slice := &PySlice{Start: MakeInt(1), Stop: MakeInt(3), Step: nil}
	err := vm.delSlice(list, slice)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list.Items))
	}
	expected := []int64{0, 3, 4}
	for i, want := range expected {
		if list.Items[i].(*PyInt).Value != want {
			t.Errorf("item[%d] = %d, want %d", i, list.Items[i].(*PyInt).Value, want)
		}
	}
}

// =====================================
// computeSliceIndices
// =====================================

func TestComputeSliceIndices(t *testing.T) {
	getInt := func(v Value, def int) int {
		if v == nil || v == None {
			return def
		}
		if i, ok := v.(*PyInt); ok {
			return int(i.Value)
		}
		return def
	}

	tests := []struct {
		name               string
		start, stop, step  Value
		length             int
		wantStart          int
		wantStop           int
		wantStep           int
		wantErr            bool
	}{
		{"basic [1:3]", MakeInt(1), MakeInt(3), nil, 5, 1, 3, 1, false},
		{"full [::]", nil, nil, nil, 5, 0, 5, 1, false},
		{"step 2", nil, nil, MakeInt(2), 5, 0, 5, 2, false},
		{"negative step", nil, nil, MakeInt(-1), 5, 4, -6, -1, false},
		{"negative index start", MakeInt(-2), nil, nil, 5, 3, 5, 1, false},
		{"negative index stop", nil, MakeInt(-1), nil, 5, 0, 4, 1, false},
		{"step zero", nil, nil, MakeInt(0), 5, 0, 0, 0, true},
		{"clamp start < 0", MakeInt(-10), MakeInt(3), nil, 5, 0, 3, 1, false},
		{"clamp stop > len", nil, MakeInt(100), nil, 5, 0, 5, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slice := &PySlice{Start: tt.start, Stop: tt.stop, Step: tt.step}
			start, stop, step, err := computeSliceIndices(slice, tt.length, getInt)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if start != tt.wantStart {
				t.Errorf("start = %d, want %d", start, tt.wantStart)
			}
			if stop != tt.wantStop {
				t.Errorf("stop = %d, want %d", stop, tt.wantStop)
			}
			if step != tt.wantStep {
				t.Errorf("step = %d, want %d", step, tt.wantStep)
			}
		})
	}
}

// =====================================
// collectSliceIndices
// =====================================

func TestCollectSliceIndices(t *testing.T) {
	tests := []struct {
		name               string
		start, stop, step  int
		want               []int
	}{
		{"forward", 0, 5, 1, []int{0, 1, 2, 3, 4}},
		{"step 2", 0, 5, 2, []int{0, 2, 4}},
		{"reverse", 4, -1, -1, []int{4, 3, 2, 1, 0}},
		{"empty forward", 3, 1, 1, nil},
		{"empty reverse", 1, 3, -1, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectSliceIndices(tt.start, tt.stop, tt.step)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d indices, want %d: %v vs %v", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("index[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}
