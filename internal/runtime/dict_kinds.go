package runtime

import (
	"fmt"
	"sort"
	"strings"
)

// DictKind identifies the dict subclasses from the collections module that
// are implemented natively on PyDict. A tagged dict behaves as a dict
// everywhere (subscripting, iteration, views, isinstance(x, dict)) and only
// differs where the subclass does: missing keys, repr, and extra methods.
type DictKind uint8

const (
	DictPlain   DictKind = iota
	DictDefault          // collections.defaultdict
	DictCounter          // collections.Counter
	DictOrdered          // collections.OrderedDict
)

// kindName returns the Python type name of a dict.
func (d *PyDict) kindName() string {
	switch d.Kind {
	case DictDefault:
		return "defaultdict"
	case DictCounter:
		return "Counter"
	case DictOrdered:
		return "OrderedDict"
	}
	return "dict"
}

// NewCollectionsDict implements the constructors Counter(...),
// defaultdict(default_factory, ...) and OrderedDict(...).
func (vm *VM) NewCollectionsDict(kind DictKind, args []Value, kwargs map[string]Value) (*PyDict, error) {
	d := &PyDict{Items: make(map[Value]Value), Kind: kind}
	switch kind {
	case DictCounter:
		if len(args) > 1 {
			return nil, fmt.Errorf("TypeError: Counter expected at most 1 argument, got %d", len(args))
		}
		if err := vm.counterUpdate(d, args, kwargs, 1); err != nil {
			return nil, err
		}
		return d, nil
	case DictDefault:
		if len(args) > 0 {
			if args[0] != None && !vm.isCallable(args[0]) {
				return nil, fmt.Errorf("TypeError: first argument must be callable or None")
			}
			if args[0] != None {
				d.DefaultFactory = args[0]
			}
			args = args[1:]
		}
	}
	if len(args) > 1 {
		return nil, fmt.Errorf("TypeError: %s expected at most 1 argument, got %d", d.kindName(), len(args))
	}
	if len(args) == 1 {
		if err := vm.dictUpdate(d, args[0]); err != nil {
			return nil, err
		}
	}
	if err := vm.dictUpdateKwargs(d, kwargs); err != nil {
		return nil, err
	}
	return d, nil
}

// isCallable reports whether v can be called.
func (vm *VM) isCallable(v Value) bool {
	switch val := v.(type) {
	case *PyFunction, *PyBuiltinFunc, *PyGoFunc, *PyClass, *PyMethod:
		return true
	case *PyInstance:
		return vm.hasMethod(val, "__call__")
	}
	return false
}

// dictMissing supplies the value for a missing key in d[key], as the
// subclass's __missing__ would: defaultdict inserts default_factory(), and
// Counter returns 0 without inserting. ok is false for plain dicts.
func (vm *VM) dictMissing(d *PyDict, key Value) (Value, bool, error) {
	switch d.Kind {
	case DictDefault:
		if d.DefaultFactory == nil {
			return nil, false, nil
		}
		v, err := vm.call(d.DefaultFactory, nil, nil)
		if err != nil {
			return nil, true, err
		}
		d.DictSet(key, v, vm)
		return v, true, nil
	case DictCounter:
		return MakeInt(0), true, nil
	}
	return nil, false, nil
}

// reprDictKind wraps a dict repr in its subclass's form.
func (vm *VM) reprDictKind(d *PyDict, body string) (string, error) {
	switch d.Kind {
	case DictDefault:
		factory := "None"
		if d.DefaultFactory != nil {
			s, err := vm.reprValue(d.DefaultFactory)
			if err != nil {
				return "", err
			}
			factory = s
		}
		return "defaultdict(" + factory + ", " + body + ")", nil
	case DictCounter, DictOrdered:
		if d.DictLen() == 0 {
			return d.kindName() + "()", nil
		}
		if d.Kind == DictCounter {
			// A Counter lists its items most common first.
			parts := make([]string, 0, d.DictLen())
			for _, t := range vm.counterItems(d) {
				k, err := vm.reprValue(t.Items[0])
				if err != nil {
					return "", err
				}
				v, err := vm.reprValue(t.Items[1])
				if err != nil {
					return "", err
				}
				parts = append(parts, k+": "+v)
			}
			body = "{" + strings.Join(parts, ", ") + "}"
		}
		return d.kindName() + "(" + body + ")", nil
	}
	return body, nil
}

// counterItems returns a Counter's items sorted by count, highest first;
// ties keep insertion order (CPython's most_common order).
func (vm *VM) counterItems(d *PyDict) []*PyTuple {
	var items []*PyTuple
	for _, k := range d.Keys(vm) {
		v, _ := d.DictGet(k, vm)
		items = append(items, &PyTuple{Items: []Value{k, v}})
	}
	var sortErr error
	sort.SliceStable(items, func(i, j int) bool {
		gt, err := vm.compareChecked(OpCompareGt, items[i].Items[1], items[j].Items[1])
		if err != nil {
			sortErr = err
			return false
		}
		return vm.truthy(gt)
	})
	_ = sortErr // counts are numbers; an incomparable count just keeps order
	return items
}

// counterUpdate adds (sign=1) or subtracts (sign=-1) counts from an
// iterable of elements or a mapping of counts, plus keyword counts.
func (vm *VM) counterUpdate(d *PyDict, args []Value, kwargs map[string]Value, sign int64) error {
	add := func(k, n Value) error {
		if !isHashable(k) {
			return fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(k))
		}
		cur, ok := d.DictGet(k, vm)
		if !ok {
			cur = MakeInt(0)
		}
		op := OpBinaryAdd
		if sign < 0 {
			op = OpBinarySubtract
		}
		r, err := vm.binaryOp(op, cur, n)
		if err != nil {
			return err
		}
		d.DictSet(k, r, vm)
		return nil
	}
	for _, arg := range args {
		if m, ok := arg.(*PyDict); ok {
			for _, k := range m.Keys(vm) {
				v, _ := m.DictGet(k, vm)
				if err := add(k, v); err != nil {
					return err
				}
			}
			continue
		}
		items, err := vm.toList(arg)
		if err != nil {
			return err
		}
		one := MakeInt(1)
		for _, item := range items {
			if err := add(item, one); err != nil {
				return err
			}
		}
	}
	for _, k := range vm.kwargNames(kwargs) {
		if err := add(&PyString{Value: k}, kwargs[k]); err != nil {
			return err
		}
	}
	return nil
}

// counterBinaryOp implements Counter + - & | Counter, keeping only
// positive counts as CPython does. handled is false for other operands.
func (vm *VM) counterBinaryOp(op Opcode, a, b Value) (Value, bool, error) {
	da, ok := a.(*PyDict)
	if !ok || da.Kind != DictCounter {
		return nil, false, nil
	}
	db, ok := b.(*PyDict)
	if !ok || db.Kind != DictCounter {
		return nil, false, nil
	}
	result := &PyDict{Items: make(map[Value]Value), Kind: DictCounter}
	zero := MakeInt(0)
	get := func(d *PyDict, k Value) Value {
		if v, ok := d.DictGet(k, vm); ok {
			return v
		}
		return zero
	}
	keys := da.Keys(vm)
	for _, k := range db.Keys(vm) {
		if !da.DictContains(k, vm) {
			keys = append(keys, k)
		}
	}
	for _, k := range keys {
		x, y := get(da, k), get(db, k)
		var v Value
		var err error
		switch op {
		case OpBinaryAdd:
			v, err = vm.binaryOp(OpBinaryAdd, x, y)
		case OpBinarySubtract:
			v, err = vm.binaryOp(OpBinarySubtract, x, y)
		case OpBinaryOr, OpBinaryAnd:
			pickY := OpCompareGt
			if op == OpBinaryAnd {
				pickY = OpCompareLt
			}
			var c Value
			c, err = vm.compareChecked(pickY, y, x)
			v = x
			if err == nil && vm.truthy(c) {
				v = y
			}
		default:
			return nil, false, nil
		}
		if err != nil {
			return nil, true, err
		}
		if pos, err := vm.compareChecked(OpCompareGt, v, zero); err != nil {
			return nil, true, err
		} else if vm.truthy(pos) {
			result.DictSet(k, v, vm)
		}
	}
	return result, true, nil
}

// dictKindAttr returns the extra methods and attributes of the collections
// dict subclasses, or nil if d has no such attribute.
func (vm *VM) dictKindAttr(d *PyDict, name string) Value {
	copyDict := func() *PyDict {
		c := &PyDict{Items: make(map[Value]Value), Kind: d.Kind, DefaultFactory: d.DefaultFactory}
		for _, k := range d.Keys(vm) {
			v, _ := d.DictGet(k, vm)
			c.DictSet(k, v, vm)
		}
		return c
	}
	method := func(fn func(args []Value, kwargs map[string]Value) (Value, error)) Value {
		return &PyBuiltinFunc{Name: d.kindName() + "." + name, Fn: fn}
	}
	if name == "copy" && d.Kind != DictPlain {
		return method(func(args []Value, kwargs map[string]Value) (Value, error) { return copyDict(), nil })
	}
	switch d.Kind {
	case DictDefault:
		if name == "default_factory" {
			if d.DefaultFactory == nil {
				return None
			}
			return d.DefaultFactory
		}
	case DictCounter:
		switch name {
		case "most_common":
			return method(func(args []Value, kwargs map[string]Value) (Value, error) {
				items := vm.counterItems(d)
				if len(args) > 0 && args[0] != None {
					n, err := vm.getIntIndex(args[0])
					if err != nil {
						return nil, err
					}
					if n < 0 {
						n = 0
					}
					if int(n) < len(items) {
						items = items[:n]
					}
				}
				out := make([]Value, len(items))
				for i, t := range items {
					out[i] = t
				}
				return &PyList{Items: out}, nil
			})
		case "elements":
			return method(func(args []Value, kwargs map[string]Value) (Value, error) {
				var out []Value
				for _, k := range d.Keys(vm) {
					v, _ := d.DictGet(k, vm)
					if n, ok := v.(*PyInt); ok && n.BigValue == nil {
						for i := int64(0); i < n.Value; i++ {
							out = append(out, k)
						}
					}
				}
				return &PyIterator{Items: out}, nil
			})
		case "update", "subtract":
			sign := int64(1)
			if name == "subtract" {
				sign = -1
			}
			return method(func(args []Value, kwargs map[string]Value) (Value, error) {
				return None, vm.counterUpdate(d, args, kwargs, sign)
			})
		case "total":
			return method(func(args []Value, kwargs map[string]Value) (Value, error) {
				var total Value = MakeInt(0)
				for _, k := range d.Keys(vm) {
					v, _ := d.DictGet(k, vm)
					t, err := vm.binaryOp(OpBinaryAdd, total, v)
					if err != nil {
						return nil, err
					}
					total = t
				}
				return total, nil
			})
		}
	case DictOrdered:
		switch name {
		case "move_to_end":
			return method(func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: move_to_end() takes exactly 1 positional argument")
				}
				v, ok := d.DictGet(args[0], vm)
				if !ok {
					return nil, vm.newKeyError(args[0])
				}
				last := true
				if l, ok := kwargs["last"]; ok {
					last = vm.truthy(l)
				}
				keys := d.Keys(vm)
				d.DictDelete(args[0], vm)
				if last {
					d.DictSet(args[0], v, vm)
					return None, nil
				}
				// Rebuild with the key first.
				rest := make([][2]Value, 0, len(keys))
				for _, k := range keys {
					if vm.equal(k, args[0]) {
						continue
					}
					kv, _ := d.DictGet(k, vm)
					rest = append(rest, [2]Value{k, kv})
				}
				d.DictClear()
				d.DictSet(args[0], v, vm)
				for _, kv := range rest {
					d.DictSet(kv[0], kv[1], vm)
				}
				return None, nil
			})
		case "popitem":
			return method(func(args []Value, kwargs map[string]Value) (Value, error) {
				keys := d.Keys(vm)
				if len(keys) == 0 {
					return nil, fmt.Errorf("KeyError: 'dictionary is empty'")
				}
				last := true
				if len(args) > 0 {
					last = vm.truthy(args[0])
				} else if l, ok := kwargs["last"]; ok {
					last = vm.truthy(l)
				}
				k := keys[len(keys)-1]
				if !last {
					k = keys[0]
				}
				v, _ := d.DictGet(k, vm)
				d.DictDelete(k, vm)
				return &PyTuple{Items: []Value{k, v}}, nil
			})
		}
	}
	return nil
}

