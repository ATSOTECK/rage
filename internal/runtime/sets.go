package runtime

import "fmt"

// Set algorithms shared by set and frozenset. Each follows the corresponding
// routine in CPython's Objects/setobject.c (which operand is iterated, when a
// copy is made, how tables are presized) so that results come out in the same
// iteration order as CPython's.

// asSetTable returns the hash table backing a set or frozenset, or nil.
func asSetTable(v Value) *setTable {
	switch s := v.(type) {
	case *PySet:
		return &s.setTable
	case *PyFrozenSet:
		return &s.setTable
	}
	return nil
}

// setElemHash hashes a value about to be stored in or looked up in a set,
// raising TypeError for unhashable values.
func (vm *VM) setElemHash(v Value) (uint64, error) {
	if !isHashable(v) {
		return 0, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(v))
	}
	return vm.hashValueVM(v), nil
}

// setKeyHash is setElemHash for lookups and removals, where CPython accepts a
// set argument by treating it as the equivalent frozenset.
func (vm *VM) setKeyHash(v Value) (uint64, error) {
	if s, ok := v.(*PySet); ok {
		return frozenSetHash(&s.setTable), nil
	}
	return vm.setElemHash(v)
}

// setAddValue adds v to t, enforcing hashability and the collection size limit.
func (vm *VM) setAddValue(t *setTable, v Value) error {
	h, err := vm.setElemHash(v)
	if err != nil {
		return err
	}
	if vm.maxCollectionSize > 0 && t.lookup(v, h, vm) < 0 {
		if err := vm.checkCollectionSize(int64(t.used), "set"); err != nil {
			return err
		}
	}
	t.add(v, h, vm)
	return nil
}

// isConstantLike reports whether v is a value CPython's compiler could have
// folded into a constant.
func isConstantLike(v Value) bool {
	switch t := v.(type) {
	case *PyInt, *PyFloat, *PyComplex, *PyString, *PyBytes, *PyBool, *PyNone:
		return true
	case *PyTuple:
		for _, item := range t.Items {
			if !isConstantLike(item) {
				return false
			}
		}
		return true
	}
	return false
}

// buildSet creates a set from the items of a set display, in order.
func (vm *VM) buildSet(items []Value) (*PySet, error) {
	s := &PySet{}
	if len(items) >= 3 {
		allConst := true
		for _, item := range items {
			if !isConstantLike(item) {
				allConst = false
				break
			}
		}
		if allConst {
			// CPython compiles a display of three or more constants into a
			// frozenset constant merged into an empty set, which sizes the
			// table differently from adding the items one at a time.
			var fs setTable
			for _, item := range items {
				if err := vm.setAddValue(&fs, item); err != nil {
					return nil, err
				}
			}
			s.merge(&fs, vm)
			return s, nil
		}
	}
	for _, item := range items {
		if err := vm.setAddValue(&s.setTable, item); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// setUpdate adds every element of iterable to t (CPython's set_update_internal).
func (vm *VM) setUpdate(t *setTable, iterable Value) error {
	if ot := asSetTable(iterable); ot != nil {
		if vm.maxCollectionSize == 0 || int64(t.used+ot.used) <= vm.maxCollectionSize {
			t.merge(ot, vm)
			return nil
		}
		for _, v := range ot.Values() {
			if err := vm.setAddValue(t, v); err != nil {
				return err
			}
		}
		return nil
	}
	if d, ok := iterable.(*PyDict); ok {
		keys := d.Keys(vm)
		t.ensureTable()
		if (t.fill+len(keys))*5 >= (len(t.slots)-1)*3 {
			t.resize((t.used + len(keys)) * 2)
		}
		for _, k := range keys {
			if err := vm.setAddValue(t, k); err != nil {
				return err
			}
		}
		return nil
	}
	items, err := vm.toList(iterable)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := vm.setAddValue(t, item); err != nil {
			return err
		}
	}
	return nil
}

// newSetTable builds a table holding the elements of iterable.
func (vm *VM) newSetTable(iterable Value) (setTable, error) {
	var t setTable
	err := vm.setUpdate(&t, iterable)
	return t, err
}

// setIntersection returns so & other (CPython's set_intersection).
func (vm *VM) setIntersection(so *setTable, other Value) (setTable, error) {
	var result setTable
	if ot := asSetTable(other); ot != nil {
		if ot == so {
			return so.copyTable(), nil
		}
		// Iterate over the smaller table, probing the larger one
		big, small := so, ot
		if small.used > big.used {
			big, small = small, big
		}
		for i := range small.slots {
			e := small.slots[i]
			if e.key != nil && big.lookup(e.key, e.hash, vm) >= 0 {
				result.add(e.key, e.hash, vm)
			}
		}
		return result, nil
	}
	items, err := vm.toList(other)
	if err != nil {
		return result, err
	}
	for _, item := range items {
		h, err := vm.setElemHash(item)
		if err != nil {
			return result, err
		}
		if so.lookup(item, h, vm) >= 0 {
			result.add(item, h, vm)
		}
	}
	return result, nil
}

// setDifference returns so - other (CPython's set_difference).
func (vm *VM) setDifference(so *setTable, other Value) (setTable, error) {
	ot := asSetTable(other)
	d, isDict := other.(*PyDict)
	otherSize := 0
	switch {
	case ot != nil:
		otherSize = ot.used
	case isDict:
		otherSize = d.DictLen()
	default:
		return vm.setCopyAndDifference(so, other)
	}
	// When so is much larger than other, copying and discarding is cheaper
	if so.used>>2 > otherSize {
		return vm.setCopyAndDifference(so, other)
	}
	var result setTable
	for i := range so.slots {
		e := so.slots[i]
		if e.key == nil {
			continue
		}
		var found bool
		if ot != nil {
			found = ot.lookup(e.key, e.hash, vm) >= 0
		} else {
			_, found = d.DictGet(e.key, vm)
		}
		if !found {
			result.add(e.key, e.hash, vm)
		}
	}
	return result, nil
}

func (vm *VM) setCopyAndDifference(so *setTable, other Value) (setTable, error) {
	result := so.copyTable()
	err := vm.setDifferenceUpdate(&result, other)
	return result, err
}

// setDifferenceUpdate removes every element of other from so.
func (vm *VM) setDifferenceUpdate(so *setTable, other Value) error {
	if ot := asSetTable(other); ot != nil {
		if ot == so {
			so.clear()
			return nil
		}
		for i := range ot.slots {
			e := ot.slots[i]
			if e.key != nil {
				so.discard(e.key, e.hash, vm)
			}
		}
	} else {
		items, err := vm.toList(other)
		if err != nil {
			return err
		}
		for _, item := range items {
			h, err := vm.setElemHash(item)
			if err != nil {
				return err
			}
			so.discard(item, h, vm)
		}
	}
	// If more than 1/4th are dummies, resize them away
	if so.fill-so.used <= (len(so.slots)-1)/4 {
		return nil
	}
	so.resize(so.growSize())
	return nil
}

// setSymmetricDifferenceUpdate toggles membership of every element of other.
func (vm *VM) setSymmetricDifferenceUpdate(so *setTable, other Value) error {
	if asSetTable(other) == so {
		so.clear()
		return nil
	}
	if d, ok := other.(*PyDict); ok {
		for _, k := range d.Keys(vm) {
			h, err := vm.setElemHash(k)
			if err != nil {
				return err
			}
			if !so.discard(k, h, vm) {
				so.add(k, h, vm)
			}
		}
		return nil
	}
	ot := asSetTable(other)
	if ot == nil {
		tmp, err := vm.newSetTable(other)
		if err != nil {
			return err
		}
		ot = &tmp
	}
	for i := range ot.slots {
		e := ot.slots[i]
		if e.key != nil && !so.discard(e.key, e.hash, vm) {
			so.add(e.key, e.hash, vm)
		}
	}
	return nil
}

// setSymmetricDifference returns so ^ other: a copy of other updated with so.
func (vm *VM) setSymmetricDifference(so *setTable, other Value) (setTable, error) {
	result, err := vm.newSetTable(other)
	if err != nil {
		return result, err
	}
	for i := range so.slots {
		e := so.slots[i]
		if e.key != nil && !result.discard(e.key, e.hash, vm) {
			result.add(e.key, e.hash, vm)
		}
	}
	return result, nil
}

// setIsSubset reports whether every element of so is in other.
func (vm *VM) setIsSubset(so *setTable, other Value) (bool, error) {
	ot := asSetTable(other)
	if ot == nil {
		tmp, err := vm.newSetTable(other)
		if err != nil {
			return false, err
		}
		ot = &tmp
	}
	if so.used > ot.used {
		return false, nil
	}
	for i := range so.slots {
		e := so.slots[i]
		if e.key != nil && ot.lookup(e.key, e.hash, vm) < 0 {
			return false, nil
		}
	}
	return true, nil
}

// setIsSuperset reports whether every element of other is in so.
func (vm *VM) setIsSuperset(so *setTable, other Value) (bool, error) {
	if ot := asSetTable(other); ot != nil {
		return vm.setIsSubset(ot, so.asValue())
	}
	items, err := vm.toList(other)
	if err != nil {
		return false, err
	}
	for _, item := range items {
		h, err := vm.setElemHash(item)
		if err != nil {
			return false, err
		}
		if so.lookup(item, h, vm) < 0 {
			return false, nil
		}
	}
	return true, nil
}

// setIsDisjoint reports whether so and other have no elements in common.
func (vm *VM) setIsDisjoint(so *setTable, other Value) (bool, error) {
	if ot := asSetTable(other); ot != nil {
		big, small := so, ot
		if small.used > big.used {
			big, small = small, big
		}
		for i := range small.slots {
			e := small.slots[i]
			if e.key != nil && big.lookup(e.key, e.hash, vm) >= 0 {
				return false, nil
			}
		}
		return true, nil
	}
	items, err := vm.toList(other)
	if err != nil {
		return false, err
	}
	for _, item := range items {
		h, err := vm.setElemHash(item)
		if err != nil {
			return false, err
		}
		if so.lookup(item, h, vm) >= 0 {
			return false, nil
		}
	}
	return true, nil
}

// setTableSubset reports whether every element of a is in b.
func setTableSubset(a, b *setTable, vm *VM) bool {
	if a.used > b.used {
		return false
	}
	for i := range a.slots {
		e := a.slots[i]
		if e.key != nil && b.lookup(e.key, e.hash, vm) < 0 {
			return false
		}
	}
	return true
}

// setTablesEqual reports whether a and b hold the same elements.
func setTablesEqual(a, b *setTable, vm *VM) bool {
	return a.used == b.used && setTableSubset(a, b, vm)
}

// asValue wraps a bare table as a frozenset so it can be passed where a Value
// is expected.
func (t *setTable) asValue() Value {
	return &PyFrozenSet{setTable: *t}
}

// setBinaryOp implements |, &, - and ^ between sets and frozensets. The result
// has the type of the left operand. ok is false if either operand is not a set.
func (vm *VM) setBinaryOp(op Opcode, a, b Value) (result Value, ok bool, err error) {
	at, bt := asSetTable(a), asSetTable(b)
	if at == nil || bt == nil {
		return nil, false, nil
	}
	var r setTable
	switch op {
	case OpBinaryOr:
		r = at.copyTable()
		if bt != at {
			r.merge(bt, vm)
		}
	case OpBinaryAnd:
		r, err = vm.setIntersection(at, b)
	case OpBinarySubtract:
		r, err = vm.setDifference(at, b)
	case OpBinaryXor:
		r, err = vm.setSymmetricDifference(at, b)
	default:
		return nil, false, nil
	}
	if err != nil {
		return nil, true, err
	}
	if _, frozen := a.(*PyFrozenSet); frozen {
		return &PyFrozenSet{setTable: r}, true, nil
	}
	return &PySet{setTable: r}, true, nil
}

// setInplaceOp implements |=, &=, -= and ^= on a mutable set.
func (vm *VM) setInplaceOp(op Opcode, s *PySet, b Value) (bool, error) {
	if asSetTable(b) == nil {
		return false, nil
	}
	switch op {
	case OpInplaceOr:
		return true, vm.setUpdate(&s.setTable, b)
	case OpInplaceAnd:
		r, err := vm.setIntersection(&s.setTable, b)
		if err != nil {
			return true, err
		}
		s.setTable = r
		return true, nil
	case OpInplaceSubtract:
		return true, vm.setDifferenceUpdate(&s.setTable, b)
	case OpInplaceXor:
		return true, vm.setSymmetricDifferenceUpdate(&s.setTable, b)
	}
	return false, nil
}

// setMethod returns the bound method name of a set or frozenset, or nil if
// there is no such method. Mutating methods are only offered on sets.
func (vm *VM) setMethod(self Value, t *setTable, name string) Value {
	_, frozen := self.(*PyFrozenSet)
	typeName := "set"
	if frozen {
		typeName = "frozenset"
	}
	wrap := func(r setTable) Value {
		if frozen {
			return &PyFrozenSet{setTable: r}
		}
		return &PySet{setTable: r}
	}
	method := func(fn func(args []Value) (Value, error)) Value {
		return &PyBuiltinFunc{Name: typeName + "." + name, Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(kwargs) > 0 {
				return nil, fmt.Errorf("TypeError: %s.%s() takes no keyword arguments", typeName, name)
			}
			return fn(args)
		}}
	}
	argCount := func(args []Value, n int) error {
		if len(args) != n {
			return fmt.Errorf("TypeError: %s.%s() takes exactly one argument (%d given)", typeName, name, len(args))
		}
		return nil
	}
	boolResult := func(b bool, err error) (Value, error) {
		if err != nil {
			return nil, err
		}
		if b {
			return True, nil
		}
		return False, nil
	}

	switch name {
	case "copy":
		return method(func(args []Value) (Value, error) {
			if frozen {
				return self, nil
			}
			return wrap(t.copyTable()), nil
		})
	case "union":
		return method(func(args []Value) (Value, error) {
			r := t.copyTable()
			for _, arg := range args {
				if asSetTable(arg) == t {
					continue
				}
				if err := vm.setUpdate(&r, arg); err != nil {
					return nil, err
				}
			}
			return wrap(r), nil
		})
	case "intersection":
		return method(func(args []Value) (Value, error) {
			if len(args) == 0 {
				return wrap(t.copyTable()), nil
			}
			cur := t
			for _, arg := range args {
				r, err := vm.setIntersection(cur, arg)
				if err != nil {
					return nil, err
				}
				cur = &r
			}
			return wrap(*cur), nil
		})
	case "difference":
		return method(func(args []Value) (Value, error) {
			if len(args) == 0 {
				return wrap(t.copyTable()), nil
			}
			r, err := vm.setDifference(t, args[0])
			if err != nil {
				return nil, err
			}
			for _, arg := range args[1:] {
				if err := vm.setDifferenceUpdate(&r, arg); err != nil {
					return nil, err
				}
			}
			return wrap(r), nil
		})
	case "symmetric_difference":
		return method(func(args []Value) (Value, error) {
			if err := argCount(args, 1); err != nil {
				return nil, err
			}
			r, err := vm.setSymmetricDifference(t, args[0])
			if err != nil {
				return nil, err
			}
			return wrap(r), nil
		})
	case "issubset":
		return method(func(args []Value) (Value, error) {
			if err := argCount(args, 1); err != nil {
				return nil, err
			}
			return boolResult(vm.setIsSubset(t, args[0]))
		})
	case "issuperset":
		return method(func(args []Value) (Value, error) {
			if err := argCount(args, 1); err != nil {
				return nil, err
			}
			return boolResult(vm.setIsSuperset(t, args[0]))
		})
	case "isdisjoint":
		return method(func(args []Value) (Value, error) {
			if err := argCount(args, 1); err != nil {
				return nil, err
			}
			return boolResult(vm.setIsDisjoint(t, args[0]))
		})
	case "__contains__":
		return method(func(args []Value) (Value, error) {
			if err := argCount(args, 1); err != nil {
				return nil, err
			}
			h, err := vm.setKeyHash(args[0])
			if err != nil {
				return nil, err
			}
			return boolResult(t.lookup(args[0], h, vm) >= 0, nil)
		})
	case "__len__":
		return method(func(args []Value) (Value, error) {
			return MakeInt(int64(t.used)), nil
		})
	case "__iter__":
		return method(func(args []Value) (Value, error) {
			return &PyIterator{Items: t.Values()}, nil
		})
	}
	if frozen {
		return nil
	}

	// Mutating methods (set only)
	switch name {
	case "add":
		return method(func(args []Value) (Value, error) {
			if err := argCount(args, 1); err != nil {
				return nil, err
			}
			return None, vm.setAddValue(t, args[0])
		})
	case "discard", "remove":
		return method(func(args []Value) (Value, error) {
			if err := argCount(args, 1); err != nil {
				return nil, err
			}
			h, err := vm.setKeyHash(args[0])
			if err != nil {
				return nil, err
			}
			if !t.discard(args[0], h, vm) && name == "remove" {
				return nil, vm.newKeyError(args[0])
			}
			return None, nil
		})
	case "pop":
		return method(func(args []Value) (Value, error) {
			if err := argCount(args, 0); err != nil {
				return nil, err
			}
			v, ok := t.pop()
			if !ok {
				return nil, fmt.Errorf("KeyError: 'pop from an empty set'")
			}
			return v, nil
		})
	case "clear":
		return method(func(args []Value) (Value, error) {
			t.clear()
			return None, nil
		})
	case "update", "__ior__":
		return method(func(args []Value) (Value, error) {
			for _, arg := range args {
				if err := vm.setUpdate(t, arg); err != nil {
					return nil, err
				}
			}
			if name == "__ior__" {
				return self, nil
			}
			return None, nil
		})
	case "difference_update", "__isub__":
		return method(func(args []Value) (Value, error) {
			for _, arg := range args {
				if err := vm.setDifferenceUpdate(t, arg); err != nil {
					return nil, err
				}
			}
			if name == "__isub__" {
				return self, nil
			}
			return None, nil
		})
	case "intersection_update", "__iand__":
		return method(func(args []Value) (Value, error) {
			cur := t
			for _, arg := range args {
				r, err := vm.setIntersection(cur, arg)
				if err != nil {
					return nil, err
				}
				cur = &r
			}
			if cur != t {
				*t = *cur
			}
			if name == "__iand__" {
				return self, nil
			}
			return None, nil
		})
	case "symmetric_difference_update", "__ixor__":
		return method(func(args []Value) (Value, error) {
			if err := argCount(args, 1); err != nil {
				return nil, err
			}
			if err := vm.setSymmetricDifferenceUpdate(t, args[0]); err != nil {
				return nil, err
			}
			if name == "__ixor__" {
				return self, nil
			}
			return None, nil
		})
	}
	return nil
}
