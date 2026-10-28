package runtime

import "fmt"

// Incremental collection-building opcodes used for unpacking in displays and
// calls: [*a, b], (*a,), {*a}, {**d, k: v}, f(*args, **kwargs). They are
// shared by the main dispatch loop and the generator dispatcher.
//
// Like LIST_APPEND/SET_ADD/MAP_ADD, each pops its operand and then updates the
// collection at peek(arg) (arg 0 = the new top of stack).

// opListExtend implements LIST_EXTEND: list.extend(iterable).
func (vm *VM) opListExtend(arg int) error {
	iterable := vm.pop()
	list := vm.peek(arg).(*PyList)
	items, err := vm.toList(iterable)
	if err != nil {
		return fmt.Errorf("TypeError: Value after * must be an iterable, not %s", vm.typeName(iterable))
	}
	if err := vm.checkCollectionSize(int64(len(list.Items)+len(items)), "list"); err != nil {
		return err
	}
	list.Items = append(list.Items, items...)
	return nil
}

// opSetUpdate implements SET_UPDATE: set.update(iterable), with CPython's
// insertion order (merging set operands, presizing for dicts).
func (vm *VM) opSetUpdate(arg int) error {
	iterable := vm.pop()
	set := vm.peek(arg).(*PySet)
	return vm.setUpdate(&set.setTable, iterable)
}

// opListToTuple implements LIST_TO_TUPLE: replaces the list on TOS with a
// tuple of the same items.
func (vm *VM) opListToTuple() {
	list := vm.pop().(*PyList)
	vm.push(&PyTuple{Items: list.Items})
}

// opDictUpdate implements DICT_UPDATE: dict.update(mapping), used by {**m}.
func (vm *VM) opDictUpdate(arg int) error {
	mapping := vm.pop()
	dict := vm.peek(arg).(*PyDict)
	return vm.mergeMapping(dict, mapping, false)
}

// opDictMerge implements DICT_MERGE: like DICT_UPDATE, but a key that is
// already present is an error, used to build the kwargs of f(**a, **b).
func (vm *VM) opDictMerge(arg int) error {
	mapping := vm.pop()
	dict := vm.peek(arg).(*PyDict)
	return vm.mergeMapping(dict, mapping, true)
}

// mergeMapping copies every key/value of mapping into dict, in the mapping's
// iteration order. Non-dict mappings must provide keys() and __getitem__.
// With forKwargs, keys must be strings and must not already be present.
func (vm *VM) mergeMapping(dict *PyDict, mapping Value, forKwargs bool) error {
	var keys []Value
	var get func(k Value) (Value, error)
	switch m := mapping.(type) {
	case *PyDict:
		keys = m.Keys(vm)
		get = func(k Value) (Value, error) {
			v, _ := m.DictGet(k, vm)
			return v, nil
		}
	default:
		keysFn, err := vm.getAttr(mapping, "keys")
		if err != nil {
			if forKwargs {
				return fmt.Errorf("TypeError: argument after ** must be a mapping, not %s", vm.typeName(mapping))
			}
			return fmt.Errorf("TypeError: '%s' object is not a mapping", vm.typeName(mapping))
		}
		keysVal, err := vm.call(keysFn, nil, nil)
		if err != nil {
			return err
		}
		if keys, err = vm.toList(keysVal); err != nil {
			return err
		}
		get = func(k Value) (Value, error) { return vm.getItem(mapping, k) }
	}

	for _, k := range keys {
		if forKwargs {
			ks, ok := k.(*PyString)
			if !ok {
				return fmt.Errorf("TypeError: keywords must be strings")
			}
			if dict.DictContains(k, vm) {
				return fmt.Errorf("TypeError: got multiple values for keyword argument '%s'", ks.Value)
			}
		}
		v, err := get(k)
		if err != nil {
			return err
		}
		dict.DictSet(k, v, vm)
	}
	return nil
}
