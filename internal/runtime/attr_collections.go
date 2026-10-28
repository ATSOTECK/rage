package runtime

import (
	"fmt"
	"sort"
)

// getAttrList handles attribute access on *PyList values.
func (vm *VM) getAttrList(lst *PyList, name string) (Value, error) {
	switch name {
	case "append":
		return &PyBuiltinFunc{Name: "list.append", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("append() takes exactly 1 argument")
			}
			if err := vm.checkCollectionSize(int64(len(lst.Items)), "list"); err != nil {
				return nil, err
			}
			lst.Items = append(lst.Items, args[0])
			return None, nil
		}}, nil
	case "pop":
		return &PyBuiltinFunc{Name: "list.pop", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(lst.Items) == 0 {
				return nil, fmt.Errorf("IndexError: pop from empty list")
			}
			idx := len(lst.Items) - 1
			if len(args) > 0 {
				idx = int(vm.toInt(args[0]))
				if idx < 0 {
					idx += len(lst.Items)
				}
			}
			if idx < 0 || idx >= len(lst.Items) {
				return nil, fmt.Errorf("IndexError: pop index out of range")
			}
			val := lst.Items[idx]
			copy(lst.Items[idx:], lst.Items[idx+1:])
			lst.Items[len(lst.Items)-1] = nil // release moved-out tail ref for GC
			lst.Items = lst.Items[:len(lst.Items)-1]
			return val, nil
		}}, nil
	case "extend":
		return &PyBuiltinFunc{Name: "list.extend", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("extend() takes exactly 1 argument")
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			if vm.maxCollectionSize > 0 && int64(len(lst.Items)+len(items)) > vm.maxCollectionSize {
				return nil, fmt.Errorf("MemoryError: list size limit exceeded (limit is %d)", vm.maxCollectionSize)
			}
			lst.Items = append(lst.Items, items...)
			return None, nil
		}}, nil
	case "insert":
		return &PyBuiltinFunc{Name: "list.insert", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("insert() takes exactly 2 arguments")
			}
			if err := vm.checkCollectionSize(int64(len(lst.Items)), "list"); err != nil {
				return nil, err
			}
			idx := int(vm.toInt(args[0]))
			if idx < 0 {
				idx += len(lst.Items)
				if idx < 0 {
					idx = 0
				}
			}
			if idx >= len(lst.Items) {
				lst.Items = append(lst.Items, args[1])
			} else {
				lst.Items = append(lst.Items, nil)
				copy(lst.Items[idx+1:], lst.Items[idx:])
				lst.Items[idx] = args[1]
			}
			return None, nil
		}}, nil
	case "remove":
		return &PyBuiltinFunc{Name: "list.remove", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("remove() takes exactly 1 argument")
			}
			for i, item := range lst.Items {
				if vm.equal(item, args[0]) {
					copy(lst.Items[i:], lst.Items[i+1:])
					lst.Items[len(lst.Items)-1] = nil // release moved-out tail ref for GC
					lst.Items = lst.Items[:len(lst.Items)-1]
					return None, nil
				}
			}
			return nil, fmt.Errorf("ValueError: list.remove(x): x not in list")
		}}, nil
	case "clear":
		return &PyBuiltinFunc{Name: "list.clear", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			lst.Items = []Value{}
			return None, nil
		}}, nil
	case "index":
		return &PyBuiltinFunc{Name: "list.index", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 || len(args) > 3 {
				return nil, fmt.Errorf("TypeError: index expected at least 1 argument, got %d", len(args))
			}
			start, end, err := vm.strSliceBounds(args[1:], len(lst.Items))
			if err != nil {
				return nil, err
			}
			for i := start; i < end && i < len(lst.Items); i++ {
				if vm.containsIdentityOrEqual(lst.Items[i], args[0]) {
					return MakeInt(int64(i)), nil
				}
			}
			return nil, fmt.Errorf("ValueError: list.index(x): x not in list")
		}}, nil
	case "count":
		return &PyBuiltinFunc{Name: "list.count", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("count() takes exactly 1 argument")
			}
			count := 0
			for _, item := range lst.Items {
				if vm.equal(item, args[0]) {
					count++
				}
			}
			return MakeInt(int64(count)), nil
		}}, nil
	case "reverse":
		return &PyBuiltinFunc{Name: "list.reverse", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			for i, j := 0, len(lst.Items)-1; i < j; i, j = i+1, j-1 {
				lst.Items[i], lst.Items[j] = lst.Items[j], lst.Items[i]
			}
			return None, nil
		}}, nil
	case "sort":
		return &PyBuiltinFunc{Name: "list.sort", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) > 0 {
				return nil, fmt.Errorf("TypeError: sort() takes no positional arguments")
			}
			keyFn, reverse, err := vm.sortKwargs("sort", kwargs)
			if err != nil {
				return nil, err
			}
			if err := vm.sortValues(lst.Items, keyFn, reverse); err != nil {
				return nil, err
			}
			return None, nil
		}}, nil
	case "copy":
		return &PyBuiltinFunc{Name: "list.copy", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			cp := make([]Value, len(lst.Items))
			copy(cp, lst.Items)
			return &PyList{Items: cp}, nil
		}}, nil
	}
	return nil, fmt.Errorf("AttributeError: 'list' object has no attribute '%s'", name)
}

// getAttrDict handles attribute access on *PyDict values.
func (vm *VM) getAttrDict(d *PyDict, name string) (Value, error) {
	switch name {
	case "get":
		return &PyBuiltinFunc{Name: "dict.get", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("get() requires at least 1 argument")
			}
			key := args[0]
			def := Value(None)
			if len(args) > 1 {
				def = args[1]
			}
			val, found := d.DictGet(key, vm)
			if found {
				return val, nil
			}
			return def, nil
		}}, nil
	case "keys":
		return &PyBuiltinFunc{Name: "dict.keys", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return &PyDictView{Dict: d, Kind: dictViewKeys, vm: vm}, nil
		}}, nil
	case "values":
		return &PyBuiltinFunc{Name: "dict.values", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return &PyDictView{Dict: d, Kind: dictViewValues, vm: vm}, nil
		}}, nil
	case "items":
		return &PyBuiltinFunc{Name: "dict.items", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return &PyDictView{Dict: d, Kind: dictViewItems, vm: vm}, nil
		}}, nil
	case "update":
		return &PyBuiltinFunc{Name: "dict.update", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) > 1 {
				return nil, fmt.Errorf("TypeError: update expected at most 1 argument, got %d", len(args))
			}
			if len(args) > 0 {
				if err := vm.dictUpdate(d, args[0]); err != nil {
					return nil, err
				}
			}
			if err := vm.dictUpdateKwargs(d, kwargs); err != nil {
				return nil, err
			}
			return None, nil
		}}, nil
	case "pop":
		return &PyBuiltinFunc{Name: "dict.pop", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("TypeError: pop expected at least 1 argument, got %d", len(args))
			}
			key := args[0]
			if !isHashable(key) {
				return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(key))
			}
			val, found := d.DictGet(key, vm)
			if found {
				d.DictDelete(key, vm)
				return val, nil
			}
			if len(args) > 1 {
				return args[1], nil
			}
			return nil, vm.newKeyError(key)
		}}, nil
	case "popitem":
		return &PyBuiltinFunc{Name: "dict.popitem", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			keys := d.Keys(vm)
			if len(keys) == 0 {
				return nil, fmt.Errorf("KeyError: 'popitem(): dictionary is empty'")
			}
			lastKey := keys[len(keys)-1]
			lastVal, _ := d.DictGet(lastKey, vm)
			d.DictDelete(lastKey, vm)
			return &PyTuple{Items: []Value{lastKey, lastVal}}, nil
		}}, nil
	case "setdefault":
		return &PyBuiltinFunc{Name: "dict.setdefault", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("setdefault() takes at least 1 argument")
			}
			key := args[0]
			if !isHashable(key) {
				return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(key))
			}
			val, found := d.DictGet(key, vm)
			if found {
				return val, nil
			}
			if err := vm.checkCollectionSize(int64(d.DictLen()), "dict"); err != nil {
				return nil, err
			}
			def := Value(None)
			if len(args) > 1 {
				def = args[1]
			}
			d.DictSet(key, def, vm)
			return def, nil
		}}, nil
	case "clear":
		return &PyBuiltinFunc{Name: "dict.clear", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			d.DictClear()
			if d.instanceOwner != nil {
				for k := range d.instanceOwner.Dict {
					delete(d.instanceOwner.Dict, k)
				}
			}
			return None, nil
		}}, nil
	case "copy":
		return &PyBuiltinFunc{Name: "dict.copy", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			cp := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
			for _, k := range d.Keys(vm) {
				if v, ok := d.DictGet(k, vm); ok {
					cp.DictSet(k, v, vm)
				}
			}
			return cp, nil
		}}, nil
	case "fromkeys":
		return &PyBuiltinFunc{Name: "dict.fromkeys", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("fromkeys() requires at least 1 argument")
			}
			keys, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			if vm.maxCollectionSize > 0 && int64(len(keys)) > vm.maxCollectionSize {
				return nil, fmt.Errorf("MemoryError: dict size limit exceeded (limit is %d)", vm.maxCollectionSize)
			}
			def := Value(None)
			if len(args) > 1 {
				def = args[1]
			}
			result := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
			for _, k := range keys {
				result.DictSet(k, def, vm)
			}
			return result, nil
		}}, nil
	}
	return nil, fmt.Errorf("AttributeError: 'dict' object has no attribute '%s'", name)
}

// getAttrSet handles attribute access on *PySet values.
func (vm *VM) getAttrSet(s *PySet, name string) (Value, error) {
	if m := vm.setMethod(s, &s.setTable, name); m != nil {
		return m, nil
	}
	return nil, fmt.Errorf("AttributeError: 'set' object has no attribute '%s'", name)
}

// getAttrFrozenSet handles attribute access on *PyFrozenSet values.
func (vm *VM) getAttrFrozenSet(fs *PyFrozenSet, name string) (Value, error) {
	if m := vm.setMethod(fs, &fs.setTable, name); m != nil {
		return m, nil
	}
	return nil, fmt.Errorf("AttributeError: 'frozenset' object has no attribute '%s'", name)
}

// getAttrTuple handles attribute access on *PyTuple values.
func (vm *VM) getAttrTuple(tpl *PyTuple, name string) (Value, error) {
	switch name {
	case "count":
		return &PyBuiltinFunc{Name: "tuple.count", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("count() takes exactly 1 argument")
			}
			count := 0
			for _, item := range tpl.Items {
				if vm.equal(item, args[0]) {
					count++
				}
			}
			return MakeInt(int64(count)), nil
		}}, nil
	case "index":
		return &PyBuiltinFunc{Name: "tuple.index", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("index() takes at least 1 argument")
			}
			n := len(tpl.Items)
			start := 0
			end := n
			if len(args) > 1 {
				start = int(vm.toInt(args[1]))
				if start < 0 {
					start += n
				}
				if start < 0 {
					start = 0
				}
			}
			if len(args) > 2 {
				end = int(vm.toInt(args[2]))
				if end < 0 {
					end += n
				}
				if end < 0 {
					end = 0
				}
			}
			for i := start; i < end && i < n; i++ {
				if vm.equal(tpl.Items[i], args[0]) {
					return MakeInt(int64(i)), nil
				}
			}
			return nil, fmt.Errorf("ValueError: tuple.index(x): x not in tuple")
		}}, nil
	}
	return nil, fmt.Errorf("AttributeError: 'tuple' object has no attribute '%s'", name)
}

// Ensure unused import is satisfied.
var _ = sort.SliceStable
