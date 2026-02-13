package runtime

import "fmt"

// Item access

func (vm *VM) getItem(obj Value, index Value) (Value, error) {
	// Handle slice objects
	if slice, ok := index.(*PySlice); ok {
		return vm.sliceSequence(obj, slice)
	}

	switch o := obj.(type) {
	case *PyList:
		idx := int(vm.toInt(index))
		if idx < 0 {
			idx = len(o.Items) + idx
		}
		if idx < 0 || idx >= len(o.Items) {
			return nil, fmt.Errorf("IndexError: list index out of range")
		}
		return o.Items[idx], nil
	case *PyTuple:
		idx := int(vm.toInt(index))
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
		idx := int(vm.toInt(index))
		if idx < 0 {
			idx = len(runes) + idx
		}
		if idx < 0 || idx >= len(runes) {
			return nil, fmt.Errorf("IndexError: string index out of range")
		}
		return &PyString{Value: string(runes[idx])}, nil
	case *PyBytes:
		idx := int(vm.toInt(index))
		if idx < 0 {
			idx = len(o.Value) + idx
		}
		if idx < 0 || idx >= len(o.Value) {
			return nil, fmt.Errorf("IndexError: index out of range")
		}
		return MakeInt(int64(o.Value[idx])), nil
	case *PyDict:
		// Use hash-based lookup for O(1) average case
		if val, found := o.DictGet(index, vm); found {
			return val, nil
		}
		return nil, fmt.Errorf("KeyError: %v", index)
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
			return result, err
		}
		return nil, fmt.Errorf("'%s' object is not subscriptable", vm.typeName(obj))
	}
	return nil, fmt.Errorf("'%s' object is not subscriptable", vm.typeName(obj))
}

// computeSliceIndices normalizes start/stop/step for a sequence of the given length.
// It handles None defaults, negative indices, and bounds clamping.
func computeSliceIndices(slice *PySlice, length int, getInt func(v Value, def int) int) (start, stop, step int, err error) {
	step = getInt(slice.Step, 1)
	if step == 0 {
		return 0, 0, 0, fmt.Errorf("slice step cannot be zero")
	}

	// Compute start/stop with correct defaults based on step direction
	if step > 0 {
		if slice.Start == nil || slice.Start == None {
			start = 0
		} else {
			start = getInt(slice.Start, 0)
		}
		if slice.Stop == nil || slice.Stop == None {
			stop = length
		} else {
			stop = getInt(slice.Stop, length)
		}
	} else {
		// For negative step, defaults are reversed
		if slice.Start == nil || slice.Start == None {
			start = length - 1
		} else {
			start = getInt(slice.Start, length-1)
		}
		if slice.Stop == nil || slice.Stop == None {
			stop = -length - 1 // Sentinel to include index 0
		} else {
			stop = getInt(slice.Stop, -length-1)
		}
	}

	// Handle negative indices
	if start < 0 && start >= -length {
		start = length + start
	}
	if stop < 0 && stop >= -length {
		stop = length + stop
	}

	// Clamp to bounds
	if step > 0 {
		if start < 0 {
			start = 0
		}
		if stop > length {
			stop = length
		}
	} else {
		if start >= length {
			start = length - 1
		}
	}

	return start, stop, step, nil
}

// collectSliceIndices returns the sequence of indices selected by a slice over a given length.
func collectSliceIndices(start, stop, step int) []int {
	var indices []int
	if step > 0 {
		for i := start; i < stop; i += step {
			indices = append(indices, i)
		}
	} else {
		// stop can be -1 to include index 0
		for i := start; i > stop && i >= 0; i += step {
			indices = append(indices, i)
		}
	}
	return indices
}

// sliceSequence handles slicing for lists, tuples, bytes, and strings
func (vm *VM) sliceSequence(obj Value, slice *PySlice) (Value, error) {
	// Helper to get int from slice component (None means use default)
	getInt := func(v Value, def int) int {
		if v == nil || v == None {
			return def
		}
		return int(vm.toInt(v))
	}

	switch o := obj.(type) {
	case *PyList:
		start, stop, step, err := computeSliceIndices(slice, len(o.Items), getInt)
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
		start, stop, step, err := computeSliceIndices(slice, len(o.Items), getInt)
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
		start, stop, step, err := computeSliceIndices(slice, len(o.Value), getInt)
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
		start, stop, step, err := computeSliceIndices(slice, len(runes), getInt)
		if err != nil {
			return nil, err
		}
		indices := collectSliceIndices(start, stop, step)
		result := make([]rune, len(indices))
		for i, idx := range indices {
			result[i] = runes[idx]
		}
		return &PyString{Value: string(result)}, nil
	}

	return nil, fmt.Errorf("'%s' object is not subscriptable", vm.typeName(obj))
}

func (vm *VM) setItem(obj Value, index Value, val Value) error {
	// Handle slice assignment
	if slice, ok := index.(*PySlice); ok {
		return vm.setSlice(obj, slice, val)
	}
	switch o := obj.(type) {
	case *PyList:
		idx := int(vm.toInt(index))
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

func (vm *VM) setSlice(obj Value, slice *PySlice, val Value) error {
	lst, ok := obj.(*PyList)
	if !ok {
		return fmt.Errorf("TypeError: '%s' object does not support slice assignment", vm.typeName(obj))
	}
	newItems, err := vm.toList(val)
	if err != nil {
		return err
	}

	length := len(lst.Items)
	start := 0
	stop := length
	if slice.Start != nil && slice.Start != None {
		start = int(vm.toInt(slice.Start))
		if start < 0 {
			start = length + start
		}
		if start < 0 {
			start = 0
		}
		if start > length {
			start = length
		}
	}
	if slice.Stop != nil && slice.Stop != None {
		stop = int(vm.toInt(slice.Stop))
		if stop < 0 {
			stop = length + stop
		}
		if stop < 0 {
			stop = 0
		}
		if stop > length {
			stop = length
		}
	}
	if start > stop {
		stop = start
	}

	// Replace lst.Items[start:stop] with newItems
	result := make([]Value, 0, start+len(newItems)+(length-stop))
	result = append(result, lst.Items[:start]...)
	result = append(result, newItems...)
	result = append(result, lst.Items[stop:]...)
	lst.Items = result
	return nil
}

func (vm *VM) delSlice(obj Value, slice *PySlice) error {
	lst, ok := obj.(*PyList)
	if !ok {
		return fmt.Errorf("TypeError: '%s' object does not support slice deletion", vm.typeName(obj))
	}

	length := len(lst.Items)
	step := 1
	if slice.Step != nil && slice.Step != None {
		step = int(vm.toInt(slice.Step))
		if step == 0 {
			return fmt.Errorf("ValueError: slice step cannot be zero")
		}
	}

	start := 0
	stop := length
	if step < 0 {
		start = length - 1
		stop = -length - 1
	}
	if slice.Start != nil && slice.Start != None {
		start = int(vm.toInt(slice.Start))
		if start < 0 {
			start = length + start
		}
		if step > 0 {
			if start < 0 {
				start = 0
			}
			if start > length {
				start = length
			}
		} else {
			if start < -1 {
				start = -1
			}
			if start >= length {
				start = length - 1
			}
		}
	}
	if slice.Stop != nil && slice.Stop != None {
		stop = int(vm.toInt(slice.Stop))
		if stop < 0 {
			stop = length + stop
		}
		if step > 0 {
			if stop < 0 {
				stop = 0
			}
			if stop > length {
				stop = length
			}
		} else {
			if stop < -length-1 {
				stop = -length - 1
			}
			if stop >= length {
				stop = length - 1
			}
		}
	}

	if step == 1 {
		// Contiguous deletion - fast path
		if start >= stop {
			return nil
		}
		result := make([]Value, 0, start+(length-stop))
		result = append(result, lst.Items[:start]...)
		result = append(result, lst.Items[stop:]...)
		lst.Items = result
		return nil
	}

	// Step-based deletion: collect indices to delete
	toDelete := make(map[int]bool)
	if step > 0 {
		for i := start; i < stop; i += step {
			toDelete[i] = true
		}
	} else {
		for i := start; i > stop; i += step {
			toDelete[i] = true
		}
	}

	result := make([]Value, 0, length-len(toDelete))
	for i, item := range lst.Items {
		if !toDelete[i] {
			result = append(result, item)
		}
	}
	lst.Items = result
	return nil
}

func (vm *VM) delItem(obj Value, index Value) error {
	// Handle slice deletion
	if slice, ok := index.(*PySlice); ok {
		return vm.delSlice(obj, slice)
	}
	switch o := obj.(type) {
	case *PyList:
		idx := int(vm.toInt(index))
		if idx < 0 {
			idx = len(o.Items) + idx
		}
		if idx < 0 || idx >= len(o.Items) {
			return fmt.Errorf("IndexError: list assignment index out of range")
		}
		o.Items = append(o.Items[:idx], o.Items[idx+1:]...)
		return nil
	case *PyDict:
		// Use hash-based deletion for O(1) average case
		if !o.DictDelete(index, vm) {
			return fmt.Errorf("KeyError: %s", vm.repr(index))
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
