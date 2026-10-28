package runtime

import (
	"fmt"
	"math"
	"strings"
)

// Item access

func (vm *VM) getItem(obj Value, index Value) (Value, error) {
	// Handle slice objects for built-in types (not PyInstance, which should dispatch to __getitem__)
	if slice, ok := index.(*PySlice); ok {
		switch obj.(type) {
		case *PyInstance:
			// Fall through to PyInstance handling below, passing the slice to __getitem__
		default:
			return vm.sliceSequence(obj, slice)
		}
	}

	switch o := obj.(type) {
	case *PyList:
		i, err := vm.getIntIndex(index)
		if err != nil {
			return nil, err
		}
		idx := int(i)
		if idx < 0 {
			idx = len(o.Items) + idx
		}
		if idx < 0 || idx >= len(o.Items) {
			return nil, fmt.Errorf("IndexError: list index out of range")
		}
		return o.Items[idx], nil
	case *PyTuple:
		i, err := vm.getIntIndex(index)
		if err != nil {
			return nil, err
		}
		idx := int(i)
		if idx < 0 {
			idx = len(o.Items) + idx
		}
		if idx < 0 || idx >= len(o.Items) {
			return nil, fmt.Errorf("IndexError: tuple index out of range")
		}
		return o.Items[idx], nil
	case *PyString:
		// Convert to runes for proper UTF-8 character indexing
		runes := []rune(o.Value)
		i, err := vm.getIntIndex(index)
		if err != nil {
			return nil, err
		}
		idx := int(i)
		if idx < 0 {
			idx = len(runes) + idx
		}
		if idx < 0 || idx >= len(runes) {
			return nil, fmt.Errorf("IndexError: string index out of range")
		}
		return &PyString{Value: string(runes[idx])}, nil
	case *PyBytes:
		i, err := vm.getIntIndex(index)
		if err != nil {
			return nil, err
		}
		idx := int(i)
		if idx < 0 {
			idx = len(o.Value) + idx
		}
		if idx < 0 || idx >= len(o.Value) {
			return nil, fmt.Errorf("IndexError: index out of range")
		}
		return MakeInt(int64(o.Value[idx])), nil
	case *PyDict:
		if !isHashable(index) {
			return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(index))
		}
		// Use hash-based lookup for O(1) average case
		if val, found := o.DictGet(index, vm); found {
			return val, nil
		}
		return nil, vm.newKeyError(index)
	case *PyRange:
		i, err := vm.getIntIndex(index)
		if err != nil {
			return nil, fmt.Errorf("TypeError: range indices must be integers or slices, not %s", vm.typeName(index))
		}
		n := rangeLen(o)
		if i < 0 {
			i += n
		}
		if i < 0 || i >= n {
			return nil, fmt.Errorf("IndexError: range object index out of range")
		}
		return MakeInt(o.Start + i*o.Step), nil
	case *PyUserData:
		// Check for __getitem__ method in metatable
		if o.Metatable != nil {
			var typeName string
			for k, v := range o.Metatable.Items {
				if ks, ok := k.(*PyString); ok && ks.Value == "__type__" {
					typeName = vm.str(v)
					break
				}
			}
			if typeName != "" {
				mt := typeMetatables[typeName]
				if mt != nil {
					if method, ok := mt.Methods["__getitem__"]; ok {
						// Call __getitem__ with userdata and index
						oldStack := vm.frame.Stack
						oldSP := vm.frame.SP
						vm.frame.Stack = make([]Value, 17)
						vm.frame.Stack[0] = o
						vm.frame.Stack[1] = index
						vm.frame.SP = 2
						n := method(vm)
						var result Value = None
						if n > 0 {
							result = vm.frame.Stack[vm.frame.SP-1]
						}
						vm.frame.Stack = oldStack
						vm.frame.SP = oldSP
						return result, nil
					}
				}
			}
		}
		return nil, fmt.Errorf("'%s' object is not subscriptable", vm.typeName(obj))
	case *PyInstance:
		// Check for __getitem__ method
		if result, found, err := vm.callDunder(o, "__getitem__", index); found {
			if err != nil && vm.isKeyError(err) {
				// KeyError from __getitem__: try __missing__
				// Clear exception state so __missing__ can execute cleanly
				vm.currentException = nil
				if missing, mFound, mErr := vm.callDunder(o, "__missing__", index); mFound {
					return missing, mErr
				}
			}
			return result, err
		}
		return nil, fmt.Errorf("'%s' object is not subscriptable", vm.typeName(obj))
	case *PyClass:
		// Class subscript: cls[item] → cls.__class_getitem__(item)
		for _, cls := range o.Mro {
			if method, ok := cls.Dict["__class_getitem__"]; ok {
				// Build args: (cls, item)
				args := []Value{o, index}
				switch fn := method.(type) {
				case *PyFunction:
					result, err := vm.callFunction(fn, args, nil)
					return result, err
				case *PyBuiltinFunc:
					result, err := fn.Fn(args, nil)
					return result, err
				case *PyClassMethod:
					// __class_getitem__ is implicitly a classmethod
					args := []Value{o, index}
					return vm.call(fn.Func, args, nil)
				}
			}
		}
		return nil, fmt.Errorf("TypeError: type '%s' is not subscriptable", o.Name)
	case *PyBuiltinFunc:
		// Built-in type subscript: list[int], dict[str, int], etc.
		switch o.Name {
		case "list", "dict", "tuple", "set", "frozenset", "type", "bytearray", "memoryview":
			args := vm.unpackSubscriptArgs(index)
			return &GenericAlias{Origin: o, Args: args}, nil
		}
		return nil, fmt.Errorf("TypeError: '%s' is not subscriptable", o.Name)
	}
	return nil, fmt.Errorf("'%s' object is not subscriptable", vm.typeName(obj))
}

// unpackSubscriptArgs converts a subscript index into a slice of type arguments.
// For dict[str, int], the compiler produces a tuple (str, int) as the index.
func (vm *VM) unpackSubscriptArgs(index Value) []Value {
	if t, ok := index.(*PyTuple); ok {
		return t.Items
	}
	return []Value{index}
}

// isKeyError checks if an error is a KeyError (either a *PyException or a Go error string).
func (vm *VM) isKeyError(err error) bool {
	if pyExc, ok := err.(*PyException); ok {
		return pyExc.Type() == "KeyError"
	}
	return strings.HasPrefix(err.Error(), "KeyError:")
}

// getAttrSlice handles attribute access on *PySlice values.
func (vm *VM) getAttrSlice(s *PySlice, name string) (Value, error) {
	orNone := func(v Value) Value {
		if v == nil {
			return None
		}
		return v
	}
	switch name {
	case "start":
		return orNone(s.Start), nil
	case "stop":
		return orNone(s.Stop), nil
	case "step":
		return orNone(s.Step), nil
	case "indices":
		return &PyBuiltinFunc{Name: "slice.indices", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: slice.indices() takes exactly one argument (%d given)", len(args))
			}
			length, err := vm.getIntIndex(args[0])
			if err != nil {
				return nil, err
			}
			if length < 0 {
				return nil, fmt.Errorf("ValueError: length should not be negative")
			}
			start, stop, step, _, err := vm.sliceIndices(s, int(length))
			if err != nil {
				return nil, err
			}
			return &PyTuple{Items: []Value{MakeInt(int64(start)), MakeInt(int64(stop)), MakeInt(int64(step))}}, nil
		}}, nil
	}
	return nil, fmt.Errorf("AttributeError: 'slice' object has no attribute '%s'", name)
}

// unpackCountError builds the ValueError raised when unpacking a sequence of
// the wrong length into `expected` targets.
func unpackCountError(expected, got int) error {
	if got > expected {
		return fmt.Errorf("ValueError: too many values to unpack (expected %d, got %d)", expected, got)
	}
	return fmt.Errorf("ValueError: not enough values to unpack (expected %d, got %d)", expected, got)
}

// computeSliceIndices normalizes start/stop/step for a sequence of the given
// length, following CPython's PySlice_AdjustIndices: out-of-range bounds are
// clamped, and for a negative step a stop of -1 means "through index 0".
func computeSliceIndices(slice *PySlice, length int, getInt func(v Value, def int) int) (start, stop, step int, err error) {
	step = getInt(slice.Step, 1)
	if step == 0 {
		return 0, 0, 0, fmt.Errorf("ValueError: slice step cannot be zero")
	}
	// Defaults when a bound is None
	defStart, defStop := 0, length
	if step < 0 {
		defStart, defStop = length-1, -1
	}
	adjust := func(v Value, def int) int {
		if v == nil || v == None {
			return def
		}
		i := getInt(v, def)
		if i < 0 {
			i += length
			if i < 0 {
				if step < 0 {
					return -1
				}
				return 0
			}
		} else if i >= length {
			if step < 0 {
				return length - 1
			}
			return length
		}
		return i
	}
	return adjust(slice.Start, defStart), adjust(slice.Stop, defStop), step, nil
}

// sliceLength returns the number of elements selected by normalized indices.
func sliceLength(start, stop, step int) int {
	if step < 0 {
		if stop < start {
			return (start-stop-1)/(-step) + 1
		}
	} else if start < stop {
		return (stop-start-1)/step + 1
	}
	return 0
}

// collectSliceIndices returns the sequence of indices selected by normalized
// slice bounds (as produced by computeSliceIndices).
func collectSliceIndices(start, stop, step int) []int {
	n := sliceLength(start, stop, step)
	if n == 0 {
		return nil
	}
	indices := make([]int, n)
	for i := range indices {
		indices[i] = start + i*step
	}
	return indices
}

// sliceBound converts a slice bound to an int. Ints too large for an index are
// clamped, as CPython does for slices.
func (vm *VM) sliceBound(v Value, def int) (int, error) {
	if v == nil || v == None {
		return def, nil
	}
	if i, ok := v.(*PyInt); ok && i.BigValue != nil {
		if i.BigValue.Sign() < 0 {
			return math.MinInt32, nil
		}
		return math.MaxInt32, nil
	}
	i, err := vm.getIntIndex(v)
	if err != nil {
		return 0, fmt.Errorf("TypeError: slice indices must be integers or None or have an __index__ method")
	}
	// Clamp far-out-of-range values so length arithmetic cannot overflow
	if i > math.MaxInt32 {
		i = math.MaxInt32
	} else if i < math.MinInt32 {
		i = math.MinInt32
	}
	return int(i), nil
}

// sliceIndices resolves a slice against a sequence length, returning the
// normalized bounds and the number of selected elements.
func (vm *VM) sliceIndices(slice *PySlice, length int) (start, stop, step, n int, err error) {
	var boundErr error
	getInt := func(v Value, def int) int {
		i, err := vm.sliceBound(v, def)
		if err != nil && boundErr == nil {
			boundErr = err
		}
		return i
	}
	start, stop, step, err = computeSliceIndices(slice, length, getInt)
	if boundErr != nil {
		return 0, 0, 0, 0, boundErr
	}
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return start, stop, step, sliceLength(start, stop, step), nil
}

// sliceSequence handles slicing for lists, tuples, bytes, strings and ranges
func (vm *VM) sliceSequence(obj Value, slice *PySlice) (Value, error) {
	switch o := obj.(type) {
	case *PyList:
		start, stop, step, _, err := vm.sliceIndices(slice, len(o.Items))
		if err != nil {
			return nil, err
		}
		indices := collectSliceIndices(start, stop, step)
		result := make([]Value, len(indices))
		for i, idx := range indices {
			result[i] = o.Items[idx]
		}
		return &PyList{Items: result}, nil

	case *PyTuple:
		start, stop, step, _, err := vm.sliceIndices(slice, len(o.Items))
		if err != nil {
			return nil, err
		}
		indices := collectSliceIndices(start, stop, step)
		result := make([]Value, len(indices))
		for i, idx := range indices {
			result[i] = o.Items[idx]
		}
		return &PyTuple{Items: result}, nil

	case *PyBytes:
		start, stop, step, _, err := vm.sliceIndices(slice, len(o.Value))
		if err != nil {
			return nil, err
		}
		indices := collectSliceIndices(start, stop, step)
		result := make([]byte, len(indices))
		for i, idx := range indices {
			result[i] = o.Value[idx]
		}
		return &PyBytes{Value: result}, nil

	case *PyString:
		runes := []rune(o.Value)
		start, stop, step, _, err := vm.sliceIndices(slice, len(runes))
		if err != nil {
			return nil, err
		}
		indices := collectSliceIndices(start, stop, step)
		result := make([]rune, len(indices))
		for i, idx := range indices {
			result[i] = runes[idx]
		}
		return &PyString{Value: string(result)}, nil

	case *PyRange:
		start, stop, step, _, err := vm.sliceIndices(slice, int(rangeLen(o)))
		if err != nil {
			return nil, err
		}
		return &PyRange{
			Start: o.Start + int64(start)*o.Step,
			Stop:  o.Start + int64(stop)*o.Step,
			Step:  o.Step * int64(step),
		}, nil
	}

	return nil, fmt.Errorf("TypeError: '%s' object is not subscriptable", vm.typeName(obj))
}

func (vm *VM) setItem(obj Value, index Value, val Value) error {
	// Handle slice assignment on lists (instances receive the slice object)
	if slice, ok := index.(*PySlice); ok {
		if _, isInst := obj.(*PyInstance); !isInst {
			return vm.setSlice(obj, slice, val)
		}
	}
	switch o := obj.(type) {
	case *PyList:
		i, err := vm.getIntIndex(index)
		if err != nil {
			return fmt.Errorf("TypeError: list indices must be integers or slices, not %s", vm.typeName(index))
		}
		idx := int(i)
		if idx < 0 {
			idx = len(o.Items) + idx
		}
		if idx < 0 || idx >= len(o.Items) {
			return fmt.Errorf("IndexError: list assignment index out of range")
		}
		o.Items[idx] = val
		return nil
	case *PyDict:
		if !isHashable(index) {
			return fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(index))
		}
		// Use hash-based storage for O(1) average case
		o.DictSet(index, val, vm)
		return nil
	case *PyInstance:
		// Check for __setitem__ method
		if _, found, err := vm.callDunder(o, "__setitem__", index, val); found {
			return err
		}
		return fmt.Errorf("TypeError: '%s' object does not support item assignment", vm.typeName(obj))
	}
	return fmt.Errorf("TypeError: '%s' object does not support item assignment", vm.typeName(obj))
}

// setSlice implements lst[slice] = iterable.
func (vm *VM) setSlice(obj Value, slice *PySlice, val Value) error {
	lst, ok := obj.(*PyList)
	if !ok {
		return fmt.Errorf("TypeError: '%s' object does not support item assignment", vm.typeName(obj))
	}
	start, stop, step, n, err := vm.sliceIndices(slice, len(lst.Items))
	if err != nil {
		return err
	}
	var newItems []Value
	if src, isList := val.(*PyList); isList {
		// Copy first so that a[::] = a and similar self-assignments work
		newItems = append([]Value(nil), src.Items...)
	} else {
		newItems, err = vm.toList(val)
		if err != nil {
			if step == 1 {
				return fmt.Errorf("TypeError: can only assign an iterable")
			}
			return fmt.Errorf("TypeError: must assign iterable to extended slice")
		}
	}

	if step == 1 {
		if stop < start {
			stop = start
		}
		newLen := len(lst.Items) - (stop - start) + len(newItems)
		if vm.maxCollectionSize > 0 && int64(newLen) > vm.maxCollectionSize {
			return fmt.Errorf("MemoryError: list size limit exceeded (limit is %d)", vm.maxCollectionSize)
		}
		result := make([]Value, 0, newLen)
		result = append(result, lst.Items[:start]...)
		result = append(result, newItems...)
		result = append(result, lst.Items[stop:]...)
		lst.Items = result
		return nil
	}

	// Extended slice: sizes must match exactly
	if len(newItems) != n {
		return fmt.Errorf("ValueError: attempt to assign sequence of size %d to extended slice of size %d", len(newItems), n)
	}
	for i, idx := range collectSliceIndices(start, stop, step) {
		lst.Items[idx] = newItems[i]
	}
	return nil
}

// delSlice implements del lst[slice].
func (vm *VM) delSlice(obj Value, slice *PySlice) error {
	lst, ok := obj.(*PyList)
	if !ok {
		return fmt.Errorf("TypeError: '%s' object does not support item deletion", vm.typeName(obj))
	}
	start, stop, step, n, err := vm.sliceIndices(slice, len(lst.Items))
	if err != nil {
		return err
	}
	if n == 0 {
		return nil
	}
	if step == 1 {
		lst.Items = append(lst.Items[:start:start], lst.Items[stop:]...)
		return nil
	}
	toDelete := make(map[int]bool, n)
	for _, idx := range collectSliceIndices(start, stop, step) {
		toDelete[idx] = true
	}
	result := make([]Value, 0, len(lst.Items)-n)
	for i, item := range lst.Items {
		if !toDelete[i] {
			result = append(result, item)
		}
	}
	lst.Items = result
	return nil
}

func (vm *VM) delItem(obj Value, index Value) error {
	// Handle slice deletion on lists (instances receive the slice object)
	if slice, ok := index.(*PySlice); ok {
		if _, isInst := obj.(*PyInstance); !isInst {
			return vm.delSlice(obj, slice)
		}
	}
	switch o := obj.(type) {
	case *PyList:
		i, err := vm.getIntIndex(index)
		if err != nil {
			return fmt.Errorf("TypeError: list indices must be integers or slices, not %s", vm.typeName(index))
		}
		idx := int(i)
		if idx < 0 {
			idx = len(o.Items) + idx
		}
		if idx < 0 || idx >= len(o.Items) {
			return fmt.Errorf("IndexError: list assignment index out of range")
		}
		copy(o.Items[idx:], o.Items[idx+1:])
		o.Items[len(o.Items)-1] = nil // release moved-out tail ref for GC
		o.Items = o.Items[:len(o.Items)-1]
		return nil
	case *PyDict:
		if !isHashable(index) {
			return fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(index))
		}
		// Use hash-based deletion for O(1) average case
		if !o.DictDelete(index, vm) {
			return vm.newKeyError(index)
		}
		return nil
	case *PyInstance:
		// Check for __delitem__ method
		if _, found, err := vm.callDunder(o, "__delitem__", index); found {
			return err
		}
		return fmt.Errorf("TypeError: '%s' object does not support item deletion", vm.typeName(obj))
	}
	return fmt.Errorf("TypeError: '%s' object does not support item deletion", vm.typeName(obj))
}
