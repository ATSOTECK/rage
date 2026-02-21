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
			lst.Items = append(lst.Items[:idx], lst.Items[idx+1:]...)
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
			lst.Items = append(lst.Items, items...)
			return None, nil
		}}, nil
	case "insert":
		return &PyBuiltinFunc{Name: "list.insert", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("insert() takes exactly 2 arguments")
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
					lst.Items = append(lst.Items[:i], lst.Items[i+1:]...)
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
			if len(args) < 1 {
				return nil, fmt.Errorf("index() takes at least 1 argument")
			}
			start := 0
			end := len(lst.Items)
			if len(args) > 1 {
				start = int(vm.toInt(args[1]))
			}
			if len(args) > 2 {
				end = int(vm.toInt(args[2]))
			}
			for i := start; i < end && i < len(lst.Items); i++ {
				if vm.equal(lst.Items[i], args[0]) {
					return MakeInt(int64(i)), nil
				}
			}
			return nil, fmt.Errorf("ValueError: %s is not in list", vm.str(args[0]))
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
			var keyFn Value
			if k, ok := kwargs["key"]; ok && k != None {
				keyFn = k
			}
			reverse := false
			if r, ok := kwargs["reverse"]; ok {
				reverse = vm.truthy(r)
			}
			var sortErr error
			sort.SliceStable(lst.Items, func(i, j int) bool {
				if sortErr != nil {
					return false
				}
				a, b := lst.Items[i], lst.Items[j]
				if keyFn != nil {
					var err error
					a, err = vm.call(keyFn, []Value{a}, nil)
					if err != nil {
						sortErr = err
						return false
					}
					b, err = vm.call(keyFn, []Value{b}, nil)
					if err != nil {
						sortErr = err
						return false
					}
				}
				// For reverse sort, compare b < a instead of a < b to maintain stability.
				// Using !less would break stability because equal elements (less=false)
				// would return true, swapping their order.
				var cmpA, cmpB Value
				if reverse {
					cmpA, cmpB = b, a
				} else {
					cmpA, cmpB = a, b
				}
				result := vm.compareOp(OpCompareLt, cmpA, cmpB)
				if result == nil {
					// compareOp set vm.currentException (e.g. TypeError for incompatible types)
					if vm.currentException != nil {
						sortErr = vm.currentException
						vm.currentException = nil
					}
					return false
				}
				return vm.truthy(result)
			})
			if sortErr != nil {
				return nil, sortErr
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
			keys := make([]Value, len(d.Keys(vm)))
			copy(keys, d.Keys(vm))
			return &PyList{Items: keys}, nil
		}}, nil
	case "values":
		return &PyBuiltinFunc{Name: "dict.values", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			orderedKeys := d.Keys(vm)
			vals := make([]Value, 0, len(orderedKeys))
			for _, k := range orderedKeys {
				if v, ok := d.DictGet(k, vm); ok {
					vals = append(vals, v)
				}
			}
			return &PyList{Items: vals}, nil
		}}, nil
	case "items":
		return &PyBuiltinFunc{Name: "dict.items", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			orderedKeys := d.Keys(vm)
			items := make([]Value, 0, len(orderedKeys))
			for _, k := range orderedKeys {
				if v, ok := d.DictGet(k, vm); ok {
					items = append(items, &PyTuple{Items: []Value{k, v}})
				}
			}
			return &PyList{Items: items}, nil
		}}, nil
	case "update":
		return &PyBuiltinFunc{Name: "dict.update", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) > 0 {
				switch src := args[0].(type) {
				case *PyDict:
					for _, k := range src.Keys(vm) {
						if v, ok := src.DictGet(k, vm); ok {
							d.DictSet(k, v, vm)
						}
					}
				default:
					items, err := vm.toList(args[0])
					if err != nil {
						return nil, err
					}
					for _, item := range items {
						pair, err := vm.toList(item)
						if err != nil {
							return nil, err
						}
						if len(pair) != 2 {
							return nil, fmt.Errorf("ValueError: dictionary update sequence element has length %d; 2 is required", len(pair))
						}
						d.DictSet(pair[0], pair[1], vm)
					}
				}
			}
			for k, v := range kwargs {
				d.DictSet(&PyString{Value: k}, v, vm)
			}
			return None, nil
		}}, nil
	case "pop":
		return &PyBuiltinFunc{Name: "dict.pop", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("pop expected at least 1 argument")
			}
			key := args[0]
			val, found := d.DictGet(key, vm)
			if found {
				d.DictDelete(key, vm)
				return val, nil
			}
			if len(args) > 1 {
				return args[1], nil
			}
			return nil, fmt.Errorf("KeyError: %s", vm.str(key))
		}}, nil
	case "popitem":
		return &PyBuiltinFunc{Name: "dict.popitem", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(d.Items) == 0 {
				return nil, fmt.Errorf("KeyError: 'popitem(): dictionary is empty'")
			}
			keys := d.Keys(vm)
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
			val, found := d.DictGet(key, vm)
			if found {
				return val, nil
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
			d.Items = make(map[Value]Value)
			d.buckets = make(map[uint64][]dictEntry)
			d.orderedKeys = nil
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
	switch name {
	case "add":
		return &PyBuiltinFunc{Name: "set.add", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("add() takes exactly 1 argument")
			}
			if !isHashable(args[0]) {
				return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(args[0]))
			}
			s.SetAdd(args[0], vm)
			return None, nil
		}}, nil
	case "discard":
		return &PyBuiltinFunc{Name: "set.discard", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("discard() takes exactly 1 argument")
			}
			s.SetRemove(args[0], vm)
			return None, nil
		}}, nil
	case "remove":
		return &PyBuiltinFunc{Name: "set.remove", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("remove() takes exactly 1 argument")
			}
			if !s.SetContains(args[0], vm) {
				return nil, fmt.Errorf("KeyError: %s", vm.str(args[0]))
			}
			s.SetRemove(args[0], vm)
			return None, nil
		}}, nil
	case "pop":
		return &PyBuiltinFunc{Name: "set.pop", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(s.Items) == 0 {
				return nil, fmt.Errorf("KeyError: 'pop from an empty set'")
			}
			var item Value
			for k := range s.Items {
				item = k
				break
			}
			s.SetRemove(item, vm)
			return item, nil
		}}, nil
	case "clear":
		return &PyBuiltinFunc{Name: "set.clear", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			s.Items = make(map[Value]struct{})
			s.buckets = make(map[uint64][]setEntry)
			return None, nil
		}}, nil
	case "copy":
		return &PyBuiltinFunc{Name: "set.copy", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			cp := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			for k := range s.Items {
				cp.SetAdd(k, vm)
			}
			return cp, nil
		}}, nil
	case "update":
		return &PyBuiltinFunc{Name: "set.update", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			for _, arg := range args {
				items, err := vm.toList(arg)
				if err != nil {
					return nil, err
				}
				for _, item := range items {
					if !isHashable(item) {
						return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(item))
					}
					s.SetAdd(item, vm)
				}
			}
			return None, nil
		}}, nil
	case "union":
		return &PyBuiltinFunc{Name: "set.union", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			result := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			for k := range s.Items {
				result.SetAdd(k, vm)
			}
			for _, arg := range args {
				items, err := vm.toList(arg)
				if err != nil {
					return nil, err
				}
				for _, item := range items {
					result.SetAdd(item, vm)
				}
			}
			return result, nil
		}}, nil
	case "intersection":
		return &PyBuiltinFunc{Name: "set.intersection", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			result := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			if len(args) == 0 {
				for k := range s.Items {
					result.SetAdd(k, vm)
				}
				return result, nil
			}
			for k := range s.Items {
				inAll := true
				for _, arg := range args {
					items, err := vm.toList(arg)
					if err != nil {
						return nil, err
					}
					found := false
					for _, item := range items {
						if vm.equal(k, item) {
							found = true
							break
						}
					}
					if !found {
						inAll = false
						break
					}
				}
				if inAll {
					result.SetAdd(k, vm)
				}
			}
			return result, nil
		}}, nil
	case "difference":
		return &PyBuiltinFunc{Name: "set.difference", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			result := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			for k := range s.Items {
				result.SetAdd(k, vm)
			}
			for _, arg := range args {
				items, err := vm.toList(arg)
				if err != nil {
					return nil, err
				}
				for _, item := range items {
					result.SetRemove(item, vm)
				}
			}
			return result, nil
		}}, nil
	case "symmetric_difference":
		return &PyBuiltinFunc{Name: "set.symmetric_difference", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("symmetric_difference() takes exactly 1 argument")
			}
			other, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			result := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			for k := range s.Items {
				found := false
				for _, item := range other {
					if vm.equal(k, item) {
						found = true
						break
					}
				}
				if !found {
					result.SetAdd(k, vm)
				}
			}
			for _, item := range other {
				if !s.SetContains(item, vm) {
					result.SetAdd(item, vm)
				}
			}
			return result, nil
		}}, nil
	case "issubset":
		return &PyBuiltinFunc{Name: "set.issubset", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("issubset() takes exactly 1 argument")
			}
			other, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			for k := range s.Items {
				found := false
				for _, item := range other {
					if vm.equal(k, item) {
						found = true
						break
					}
				}
				if !found {
					return False, nil
				}
			}
			return True, nil
		}}, nil
	case "issuperset":
		return &PyBuiltinFunc{Name: "set.issuperset", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("issuperset() takes exactly 1 argument")
			}
			other, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			for _, item := range other {
				if !s.SetContains(item, vm) {
					return False, nil
				}
			}
			return True, nil
		}}, nil
	case "isdisjoint":
		return &PyBuiltinFunc{Name: "set.isdisjoint", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("isdisjoint() takes exactly 1 argument")
			}
			other, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			for _, item := range other {
				if s.SetContains(item, vm) {
					return False, nil
				}
			}
			return True, nil
		}}, nil
	case "difference_update":
		return &PyBuiltinFunc{Name: "set.difference_update", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			for _, arg := range args {
				items, err := vm.toList(arg)
				if err != nil {
					return nil, err
				}
				for _, item := range items {
					s.SetRemove(item, vm)
				}
			}
			return None, nil
		}}, nil
	case "intersection_update":
		return &PyBuiltinFunc{Name: "set.intersection_update", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return None, nil
			}
			// Collect all items to keep: those in self AND in all args
			keep := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			for k := range s.Items {
				inAll := true
				for _, arg := range args {
					items, err := vm.toList(arg)
					if err != nil {
						return nil, err
					}
					found := false
					for _, item := range items {
						if vm.equal(k, item) {
							found = true
							break
						}
					}
					if !found {
						inAll = false
						break
					}
				}
				if inAll {
					keep.SetAdd(k, vm)
				}
			}
			// Replace contents of s with keep
			s.Items = keep.Items
			s.buckets = keep.buckets
			s.size = keep.size
			return None, nil
		}}, nil
	case "symmetric_difference_update":
		return &PyBuiltinFunc{Name: "set.symmetric_difference_update", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("symmetric_difference_update() takes exactly 1 argument")
			}
			other, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			// Items in self but not in other stay
			// Items in other but not in self get added
			// Items in both get removed
			toRemove := []Value{}
			toAdd := []Value{}
			for k := range s.Items {
				for _, item := range other {
					if vm.equal(k, item) {
						toRemove = append(toRemove, k)
						break
					}
				}
			}
			for _, item := range other {
				if !s.SetContains(item, vm) {
					toAdd = append(toAdd, item)
				}
			}
			for _, item := range toRemove {
				s.SetRemove(item, vm)
			}
			for _, item := range toAdd {
				s.SetAdd(item, vm)
			}
			return None, nil
		}}, nil
	}
	return nil, fmt.Errorf("AttributeError: 'set' object has no attribute '%s'", name)
}

// getAttrFrozenSet handles attribute access on *PyFrozenSet values.
func (vm *VM) getAttrFrozenSet(fs *PyFrozenSet, name string) (Value, error) {
	if name == "copy" {
		return &PyBuiltinFunc{Name: "frozenset.copy", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			newFS := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			for k := range fs.Items {
				newFS.FrozenSetAdd(k, vm)
			}
			return newFS, nil
		}}, nil
	}
	if name == "union" {
		return &PyBuiltinFunc{Name: "frozenset.union", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			result := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			for k := range fs.Items {
				result.FrozenSetAdd(k, vm)
			}
			for _, arg := range args {
				items, err := vm.toList(arg)
				if err != nil {
					return nil, err
				}
				for _, item := range items {
					if !isHashable(item) {
						return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(item))
					}
					result.FrozenSetAdd(item, vm)
				}
			}
			return result, nil
		}}, nil
	}
	if name == "intersection" {
		return &PyBuiltinFunc{Name: "frozenset.intersection", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			result := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			if len(args) == 0 {
				for k := range fs.Items {
					result.FrozenSetAdd(k, vm)
				}
				return result, nil
			}
			// Start with items from this frozenset that are in all other sets
			for k := range fs.Items {
				inAll := true
				for _, arg := range args {
					items, err := vm.toList(arg)
					if err != nil {
						return nil, err
					}
					found := false
					for _, item := range items {
						if vm.equal(k, item) {
							found = true
							break
						}
					}
					if !found {
						inAll = false
						break
					}
				}
				if inAll {
					result.FrozenSetAdd(k, vm)
				}
			}
			return result, nil
		}}, nil
	}
	if name == "difference" {
		return &PyBuiltinFunc{Name: "frozenset.difference", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			result := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			for k := range fs.Items {
				result.FrozenSetAdd(k, vm)
			}
			for _, arg := range args {
				items, err := vm.toList(arg)
				if err != nil {
					return nil, err
				}
				for _, item := range items {
					// Remove if present
					for rk := range result.Items {
						if vm.equal(rk, item) {
							delete(result.Items, rk)
							// Also remove from buckets
							if result.buckets != nil {
								h := hashValue(rk)
								entries := result.buckets[h]
								for i, e := range entries {
									if vm.equal(e.value, rk) {
										result.buckets[h] = append(entries[:i], entries[i+1:]...)
										result.size--
										break
									}
								}
							}
							break
						}
					}
				}
			}
			return result, nil
		}}, nil
	}
	if name == "symmetric_difference" {
		return &PyBuiltinFunc{Name: "frozenset.symmetric_difference", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("symmetric_difference() takes exactly 1 argument")
			}
			other, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			result := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			// Add items from fs that are not in other
			for k := range fs.Items {
				found := false
				for _, item := range other {
					if vm.equal(k, item) {
						found = true
						break
					}
				}
				if !found {
					result.FrozenSetAdd(k, vm)
				}
			}
			// Add items from other that are not in fs
			for _, item := range other {
				if !isHashable(item) {
					return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(item))
				}
				if !fs.FrozenSetContains(item, vm) {
					result.FrozenSetAdd(item, vm)
				}
			}
			return result, nil
		}}, nil
	}
	if name == "issubset" {
		return &PyBuiltinFunc{Name: "frozenset.issubset", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("issubset() takes exactly 1 argument")
			}
			other, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			for k := range fs.Items {
				found := false
				for _, item := range other {
					if vm.equal(k, item) {
						found = true
						break
					}
				}
				if !found {
					return False, nil
				}
			}
			return True, nil
		}}, nil
	}
	if name == "issuperset" {
		return &PyBuiltinFunc{Name: "frozenset.issuperset", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("issuperset() takes exactly 1 argument")
			}
			other, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			for _, item := range other {
				if !fs.FrozenSetContains(item, vm) {
					return False, nil
				}
			}
			return True, nil
		}}, nil
	}
	if name == "isdisjoint" {
		return &PyBuiltinFunc{Name: "frozenset.isdisjoint", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("isdisjoint() takes exactly 1 argument")
			}
			other, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			for _, item := range other {
				if fs.FrozenSetContains(item, vm) {
					return False, nil
				}
			}
			return True, nil
		}}, nil
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
			start := 0
			end := len(tpl.Items)
			if len(args) > 1 {
				start = int(vm.toInt(args[1]))
			}
			if len(args) > 2 {
				end = int(vm.toInt(args[2]))
			}
			for i := start; i < end; i++ {
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
