package runtime

import (
	"fmt"
	"strings"
)

// NewNamedTupleClass builds the class returned by collections.namedtuple.
// Instances keep their fields as attributes and behave as a tuple of those
// values for indexing, iteration, len(), comparison, hashing and
// unpacking. defaults apply to the rightmost fields.
func (vm *VM) NewNamedTupleClass(typename string, fields []string, defaults []Value) (*PyClass, error) {
	if len(defaults) > len(fields) {
		return nil, fmt.Errorf("TypeError: Got more default values than field names")
	}
	seen := make(map[string]bool, len(fields))
	for _, f := range fields {
		if seen[f] {
			return nil, fmt.Errorf("ValueError: Encountered duplicate field name: '%s'", f)
		}
		seen[f] = true
	}

	fieldVals := make([]Value, len(fields))
	for i, f := range fields {
		fieldVals[i] = &PyString{Value: f}
	}
	fieldsTuple := &PyTuple{Items: fieldVals}

	cls := &PyClass{Name: typename, Dict: make(map[string]Value)}
	object, _ := vm.builtins["object"].(*PyClass)
	if object != nil {
		cls.Bases = []*PyClass{object}
	}
	mro, err := vm.ComputeC3MRO(cls, cls.Bases)
	if err != nil {
		return nil, err
	}
	cls.Mro = mro

	// values returns an instance's field values in order.
	values := func(self Value) ([]Value, error) {
		inst, ok := self.(*PyInstance)
		if !ok {
			return nil, fmt.Errorf("TypeError: descriptor requires a '%s' object", typename)
		}
		out := make([]Value, len(fields))
		for i, f := range fields {
			out[i] = inst.Dict[f]
		}
		return out, nil
	}
	// asTuple returns other's values when it is a tuple or a namedtuple.
	asTuple := func(other Value) (*PyTuple, bool) {
		switch o := other.(type) {
		case *PyTuple:
			return o, true
		case *PyInstance:
			if _, ok := o.Class.Dict["_fields"]; ok {
				if f, ok := o.Class.Dict["_fields"].(*PyTuple); ok {
					items := make([]Value, len(f.Items))
					for i, name := range f.Items {
						items[i] = o.Dict[name.(*PyString).Value]
					}
					return &PyTuple{Items: items}, true
				}
			}
		}
		return nil, false
	}
	method := func(name string, fn func(args []Value, kwargs map[string]Value) (Value, error)) {
		cls.Dict[name] = &PyBuiltinFunc{Name: typename + "." + name, Fn: fn}
	}

	method("__init__", func(args []Value, kwargs map[string]Value) (Value, error) {
		inst, ok := args[0].(*PyInstance)
		if !ok {
			return nil, fmt.Errorf("TypeError: %s.__init__ requires an instance", typename)
		}
		pos := args[1:]
		if len(pos) > len(fields) {
			return nil, fmt.Errorf("TypeError: %s() takes %d positional arguments but %d were given", typename, len(fields)+1, len(pos)+1)
		}
		set := make([]bool, len(fields))
		for i, v := range pos {
			inst.Dict[fields[i]] = v
			set[i] = true
		}
		for _, k := range vm.kwargNames(kwargs) {
			idx := -1
			for i, f := range fields {
				if f == k {
					idx = i
					break
				}
			}
			if idx < 0 {
				return nil, fmt.Errorf("TypeError: %s() got an unexpected keyword argument '%s'", typename, k)
			}
			if set[idx] {
				return nil, fmt.Errorf("TypeError: %s() got multiple values for argument '%s'", typename, k)
			}
			inst.Dict[k] = kwargs[k]
			set[idx] = true
		}
		firstDefault := len(fields) - len(defaults)
		var missing []string
		for i, f := range fields {
			if set[i] {
				continue
			}
			if i >= firstDefault {
				inst.Dict[f] = defaults[i-firstDefault]
				continue
			}
			missing = append(missing, "'"+f+"'")
		}
		if len(missing) > 0 {
			noun := "argument"
			if len(missing) > 1 {
				noun = "arguments"
			}
			return nil, fmt.Errorf("TypeError: %s() missing %d required positional %s: %s", typename, len(missing), noun, strings.Join(missing, " and "))
		}
		return None, nil
	})
	method("__repr__", func(args []Value, kwargs map[string]Value) (Value, error) {
		vals, err := values(args[0])
		if err != nil {
			return nil, err
		}
		parts := make([]string, len(fields))
		for i, f := range fields {
			s, err := vm.reprValue(vals[i])
			if err != nil {
				return nil, err
			}
			parts[i] = f + "=" + s
		}
		return &PyString{Value: typename + "(" + strings.Join(parts, ", ") + ")"}, nil
	})
	method("__getitem__", func(args []Value, kwargs map[string]Value) (Value, error) {
		vals, err := values(args[0])
		if err != nil {
			return nil, err
		}
		return vm.getItem(&PyTuple{Items: vals}, args[1])
	})
	method("__iter__", func(args []Value, kwargs map[string]Value) (Value, error) {
		vals, err := values(args[0])
		if err != nil {
			return nil, err
		}
		return &PyIterator{Items: vals}, nil
	})
	method("__len__", func(args []Value, kwargs map[string]Value) (Value, error) {
		return MakeInt(int64(len(fields))), nil
	})
	method("__hash__", func(args []Value, kwargs map[string]Value) (Value, error) {
		vals, err := values(args[0])
		if err != nil {
			return nil, err
		}
		return MakeInt(int64(vm.hashValueVM(&PyTuple{Items: vals}) >> 1)), nil
	})
	for _, c := range []struct {
		name string
		op   Opcode
	}{{"__eq__", OpCompareEq}, {"__ne__", OpCompareNe}, {"__lt__", OpCompareLt}, {"__le__", OpCompareLe}, {"__gt__", OpCompareGt}, {"__ge__", OpCompareGe}} {
		op := c.op
		method(c.name, func(args []Value, kwargs map[string]Value) (Value, error) {
			self, _ := asTuple(args[0])
			other, ok := asTuple(args[1])
			if self == nil || !ok {
				return NotImplemented, nil
			}
			return vm.compareChecked(op, self, other)
		})
	}
	method("_asdict", func(args []Value, kwargs map[string]Value) (Value, error) {
		vals, err := values(args[0])
		if err != nil {
			return nil, err
		}
		d := &PyDict{Items: make(map[Value]Value)}
		for i, f := range fieldVals {
			d.DictSet(f, vals[i], vm)
		}
		return d, nil
	})
	method("_replace", func(args []Value, kwargs map[string]Value) (Value, error) {
		vals, err := values(args[0])
		if err != nil {
			return nil, err
		}
		for _, k := range vm.kwargNames(kwargs) {
			if !seen[k] {
				return nil, fmt.Errorf("ValueError: Got unexpected field names: ['%s']", k)
			}
		}
		newVals := make([]Value, len(fields))
		for i, f := range fields {
			if v, ok := kwargs[f]; ok {
				newVals[i] = v
			} else {
				newVals[i] = vals[i]
			}
		}
		return vm.call(cls, newVals, nil)
	})
	cls.Dict["_make"] = &PyClassMethod{Func: &PyBuiltinFunc{Name: typename + "._make", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("TypeError: _make() takes exactly one argument")
		}
		items, err := vm.toList(args[1])
		if err != nil {
			return nil, err
		}
		return vm.call(args[0], items, nil)
	}}}
	cls.Dict["_fields"] = fieldsTuple
	cls.Dict["__match_args__"] = fieldsTuple
	defaultsDict := &PyDict{Items: make(map[Value]Value)}
	for i, v := range defaults {
		defaultsDict.DictSet(fieldVals[len(fields)-len(defaults)+i], v, vm)
	}
	cls.Dict["_field_defaults"] = defaultsDict
	return cls, nil
}
