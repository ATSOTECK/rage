package runtime

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Dict view kinds
const (
	dictViewKeys = iota
	dictViewValues
	dictViewItems
)

// PyDictView is a live view of a dict's keys, values or items, as returned by
// dict.keys(), dict.values() and dict.items(). It reflects later changes to
// the dict.
type PyDictView struct {
	Dict *PyDict
	Kind int
	vm   *VM // used by String() to repr elements
}

func (v *PyDictView) Type() string {
	switch v.Kind {
	case dictViewKeys:
		return "dict_keys"
	case dictViewValues:
		return "dict_values"
	}
	return "dict_items"
}

func (v *PyDictView) String() string {
	items := v.Items()
	parts := make([]string, len(items))
	for i, item := range items {
		if v.vm != nil {
			parts[i] = v.vm.repr(item)
		} else {
			parts[i] = fmt.Sprint(item)
		}
	}
	return v.Type() + "([" + strings.Join(parts, ", ") + "])"
}

// Items returns the view's current elements in dict order.
func (v *PyDictView) Items() []Value {
	keys := v.Dict.Keys(v.vm)
	if v.Kind == dictViewKeys {
		return keys
	}
	out := make([]Value, 0, len(keys))
	for _, k := range keys {
		val, ok := v.Dict.DictGet(k, v.vm)
		if !ok {
			continue
		}
		if v.Kind == dictViewValues {
			out = append(out, val)
		} else {
			out = append(out, &PyTuple{Items: []Value{k, val}})
		}
	}
	return out
}

// isSetLikeView reports whether v is a keys or items view, which support set
// operations and set comparisons.
func isSetLikeView(v Value) bool {
	dv, ok := v.(*PyDictView)
	return ok && dv.Kind != dictViewValues
}

// viewContains implements `x in view`.
func (vm *VM) viewContains(v *PyDictView, x Value) (bool, error) {
	switch v.Kind {
	case dictViewKeys:
		if !isHashable(x) {
			return false, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(x))
		}
		return v.Dict.DictContains(x, vm), nil
	case dictViewItems:
		t, ok := x.(*PyTuple)
		if !ok || len(t.Items) != 2 || !isHashable(t.Items[0]) {
			return false, nil
		}
		val, found := v.Dict.DictGet(t.Items[0], vm)
		return found && (val == t.Items[1] || vm.equal(val, t.Items[1])), nil
	}
	for _, val := range v.Items() {
		if vm.containsIdentityOrEqual(val, x) {
			return true, nil
		}
	}
	return false, nil
}

// setTableOf returns a set table holding the elements of a set, frozenset or
// set-like dict view, or nil for anything else.
func (vm *VM) setTableOf(v Value) (*setTable, error) {
	if t := asSetTable(v); t != nil {
		return t, nil
	}
	if dv, ok := v.(*PyDictView); ok && dv.Kind != dictViewValues {
		var t setTable
		for _, item := range dv.Items() {
			if err := vm.setAddValue(&t, item); err != nil {
				return nil, err
			}
		}
		return &t, nil
	}
	return nil, nil
}

// viewSetOp implements |, &, - and ^ when either operand is a keys or items
// view. The other operand may be any iterable; the result is a set.
func (vm *VM) viewSetOp(op Opcode, a, b Value) (Value, bool, error) {
	if !isSetLikeView(a) && !isSetLikeView(b) {
		return nil, false, nil
	}
	if op != OpBinaryOr && op != OpBinaryAnd && op != OpBinarySubtract && op != OpBinaryXor {
		return nil, false, nil
	}
	toSet := func(v Value) (*PySet, error) {
		s := &PySet{}
		if dv, ok := v.(*PyDictView); ok {
			v = &PyList{Items: dv.Items()}
		}
		if err := vm.setUpdate(&s.setTable, v); err != nil {
			return nil, err
		}
		return s, nil
	}
	left, err := toSet(a)
	if err != nil {
		return nil, true, err
	}
	right, err := toSet(b)
	if err != nil {
		return nil, true, err
	}
	result, _, err := vm.setBinaryOp(op, left, right)
	return result, true, err
}

// dictViewMethod returns the named method of a dict view, or nil.
func (vm *VM) dictViewMethod(v *PyDictView, name string) Value {
	method := func(fn func(args []Value) (Value, error)) Value {
		return &PyBuiltinFunc{Name: v.Type() + "." + name, Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return fn(args)
		}}
	}
	switch name {
	case "__len__":
		return method(func(args []Value) (Value, error) {
			return MakeInt(int64(v.Dict.DictLen())), nil
		})
	case "__iter__":
		return method(func(args []Value) (Value, error) {
			return &PyIterator{Items: v.Items()}, nil
		})
	case "__reversed__":
		return method(func(args []Value) (Value, error) {
			items := v.Items()
			for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
				items[i], items[j] = items[j], items[i]
			}
			return &PyIterator{Items: items}, nil
		})
	case "__contains__":
		return method(func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: __contains__() takes exactly one argument (%d given)", len(args))
			}
			found, err := vm.viewContains(v, args[0])
			if err != nil {
				return nil, err
			}
			if found {
				return True, nil
			}
			return False, nil
		})
	case "mapping":
		return v.Dict
	case "isdisjoint":
		if v.Kind == dictViewValues {
			return nil
		}
		return method(func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: isdisjoint() takes exactly one argument (%d given)", len(args))
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			for _, item := range items {
				found, err := vm.viewContains(v, item)
				if err != nil {
					return nil, err
				}
				if found {
					return False, nil
				}
			}
			return True, nil
		})
	}
	return nil
}

// dictUpdate merges src into d like dict.update(src): src may be a dict, an
// object with a keys() method, or an iterable of key/value pairs.
func (vm *VM) dictUpdate(d *PyDict, src Value) error {
	set := func(k, v Value) error {
		if !isHashable(k) {
			return fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(k))
		}
		if vm.maxCollectionSize > 0 && !d.DictContains(k, vm) {
			if err := vm.checkCollectionSize(int64(d.DictLen()), "dict"); err != nil {
				return err
			}
		}
		d.DictSet(k, v, vm)
		return nil
	}
	switch s := src.(type) {
	case *PyDict:
		if s == d {
			return nil
		}
		for _, k := range s.Keys(vm) {
			if v, ok := s.DictGet(k, vm); ok {
				if err := set(k, v); err != nil {
					return err
				}
			}
		}
		return nil
	case *PyInstance:
		// Mapping protocol: anything with keys() and __getitem__
		if keysFn, err := vm.getAttr(s, "keys"); err == nil {
			keysVal, err := vm.call(keysFn, nil, nil)
			if err != nil {
				return err
			}
			keys, err := vm.toList(keysVal)
			if err != nil {
				return err
			}
			for _, k := range keys {
				v, err := vm.getItem(s, k)
				if err != nil {
					return err
				}
				if err := set(k, v); err != nil {
					return err
				}
			}
			return nil
		}
	}
	items, err := vm.toList(src)
	if err != nil {
		return err
	}
	for i, item := range items {
		var pair []Value
		switch p := item.(type) {
		case *PyTuple:
			pair = p.Items
		case *PyList:
			pair = p.Items
		case *PyString, *PyBytes, *PyDict, *PySet, *PyFrozenSet, *PyDictView, *PyRange, *PyInstance, *PyIterator, *PyGenerator:
			pair, err = vm.toList(p)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("TypeError: object is not iterable")
		}
		if len(pair) != 2 {
			return fmt.Errorf("ValueError: dictionary update sequence element #%d has length %d; 2 is required", i, len(pair))
		}
		if err := set(pair[0], pair[1]); err != nil {
			return err
		}
	}
	return nil
}

// copyDictWithout returns a copy of d, in order, minus the given keys (the
// **rest capture of a mapping pattern).
func (vm *VM) copyDictWithout(d *PyDict, remove []Value) *PyDict {
	result := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	for _, k := range d.Keys(vm) {
		skip := false
		for _, r := range remove {
			if vm.equal(k, r) {
				skip = true
				break
			}
		}
		if !skip {
			v, _ := d.DictGet(k, vm)
			result.DictSet(k, v, vm)
		}
	}
	return result
}

// localsDict builds a dict of the bound local variables, in definition order.
func (vm *VM) localsDict(names []string, locals []Value) *PyDict {
	result := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	for i, name := range names {
		if i < len(locals) && locals[i] != nil {
			result.DictSet(&PyString{Value: name}, locals[i], vm)
		}
	}
	return result
}

// dictUpdateKwargs stores keyword arguments into d in call order.
func (vm *VM) dictUpdateKwargs(d *PyDict, kwargs map[string]Value) error {
	for _, name := range vm.kwargNames(kwargs) {
		key := &PyString{Value: name}
		if vm.maxCollectionSize > 0 && !d.DictContains(key, vm) {
			if err := vm.checkCollectionSize(int64(d.DictLen()), "dict"); err != nil {
				return err
			}
		}
		d.DictSet(key, kwargs[name], vm)
	}
	return nil
}

// kwargNames returns the names of keyword arguments in the order they were
// passed when that is known, and in sorted order otherwise, so the result is
// always deterministic.
func (vm *VM) kwargNames(kwargs map[string]Value) []string {
	names := make([]string, 0, len(kwargs))
	if order := vm.kwargOrder(kwargs); order != nil {
		return append(names, order...)
	}
	for name := range kwargs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// kwargOrderRecord remembers the order in which a kwargs map was filled.
type kwargOrderRecord struct {
	m     uintptr
	names []string
}

// RecordKwargOrder notes the order of the keyword arguments in kwargs, so
// that consumers which build dicts from it (**kwargs parameters, dict(...))
// can preserve Python's call order. Call sites that build a kwargs map from
// keyword names should call this; the most recent few maps are remembered.
func (vm *VM) RecordKwargOrder(kwargs map[string]Value, names []string) {
	if len(kwargs) < 2 {
		return
	}
	vm.kwargOrders[vm.kwargOrderNext] = kwargOrderRecord{m: reflect.ValueOf(kwargs).Pointer(), names: names}
	vm.kwargOrderNext = (vm.kwargOrderNext + 1) % len(vm.kwargOrders)
}

// RecordKwargNames is RecordKwargOrder for a tuple of keyword names, as used
// by the CALL_KW instruction.
func (vm *VM) RecordKwargNames(kwargs map[string]Value, kwNames *PyTuple) {
	if len(kwargs) < 2 {
		return
	}
	names := make([]string, 0, len(kwNames.Items))
	for _, n := range kwNames.Items {
		if s, ok := n.(*PyString); ok {
			names = append(names, s.Value)
		}
	}
	vm.RecordKwargOrder(kwargs, names)
}

// kwargOrder returns the recorded call order of a kwargs map, or nil.
func (vm *VM) kwargOrder(kwargs map[string]Value) []string {
	if len(kwargs) < 2 {
		return nil
	}
	p := reflect.ValueOf(kwargs).Pointer()
	for _, rec := range vm.kwargOrders {
		if rec.m != p || len(rec.names) != len(kwargs) {
			continue
		}
		// The map may have been reused or modified since; check the names
		valid := true
		for _, n := range rec.names {
			if _, ok := kwargs[n]; !ok {
				valid = false
				break
			}
		}
		if valid {
			return rec.names
		}
	}
	return nil
}
