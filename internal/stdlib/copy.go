package stdlib

import (
	"fmt"
	"reflect"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// InitCopyModule registers the copy module.
func InitCopyModule() {
	runtime.RegisterModule("copy", func(vm *runtime.VM) *runtime.PyModule {
		mod := runtime.NewModule("copy")

		mod.Dict["copy"] = &runtime.PyBuiltinFunc{
			Name: "copy",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: copy() takes exactly 1 argument (%d given)", len(args))
				}
				return shallowCopy(vm, args[0])
			},
		}

		mod.Dict["deepcopy"] = &runtime.PyBuiltinFunc{
			Name: "deepcopy",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				if len(args) < 1 || len(args) > 2 {
					return nil, fmt.Errorf("TypeError: deepcopy() takes 1 or 2 arguments (%d given)", len(args))
				}
				memo := make(map[uintptr]runtime.Value)
				return deepCopy(vm, args[0], memo)
			},
		}

		// Also expose the Error class
		mod.Dict["Error"] = vm.GetBuiltin("Exception")

		return mod
	})
}

// shallowCopy creates a shallow copy of a value.
// For instances with __copy__, calls __copy__().
func shallowCopy(vm *runtime.VM, val runtime.Value) (runtime.Value, error) {
	switch v := val.(type) {
	case *runtime.PyInstance:
		// Check for __copy__
		for _, cls := range v.Class.Mro {
			if method, ok := cls.Dict["__copy__"]; ok {
				args := []runtime.Value{v}
				return vm.Call(method, args, nil)
			}
		}
		// Default: shallow copy of instance dict/slots
		inst := &runtime.PyInstance{Class: v.Class}
		if v.Dict != nil {
			inst.Dict = make(map[string]runtime.Value, len(v.Dict))
			for k, val := range v.Dict {
				inst.Dict[k] = val
			}
		}
		if v.Slots != nil {
			inst.Slots = make(map[string]runtime.Value, len(v.Slots))
			for k, val := range v.Slots {
				inst.Slots[k] = val
			}
		}
		return inst, nil

	case *runtime.PyList:
		items := make([]runtime.Value, len(v.Items))
		copy(items, v.Items)
		return &runtime.PyList{Items: items}, nil

	case *runtime.PyDict:
		d := &runtime.PyDict{Items: make(map[runtime.Value]runtime.Value, len(v.Items))}
		for _, k := range v.Keys(vm) {
			if val, ok := v.DictGet(k, vm); ok {
				d.DictSet(k, val, vm)
			}
		}
		return d, nil

	case *runtime.PySet:
		s := &runtime.PySet{Items: make(map[runtime.Value]struct{}, len(v.Items))}
		for k := range v.Items {
			s.Items[k] = struct{}{}
		}
		return s, nil

	case *runtime.PyTuple:
		// Tuples are immutable — return the same object
		return v, nil

	case *runtime.PyFrozenSet:
		// Frozen sets are immutable — return the same object
		return v, nil

	default:
		// Immutable types (int, float, str, bool, None, bytes) — return as-is
		return val, nil
	}
}

// ptrOf returns the pointer identity of a Value for cycle detection in deepCopy.
func ptrOf(v runtime.Value) uintptr {
	return reflect.ValueOf(v).Pointer()
}

// deepCopy creates a deep copy of a value.
// For instances with __deepcopy__, calls __deepcopy__(memo).
func deepCopy(vm *runtime.VM, val runtime.Value, memo map[uintptr]runtime.Value) (runtime.Value, error) {
	// Check memo for already-copied objects (cycle detection)
	switch val.(type) {
	case *runtime.PyList, *runtime.PyDict, *runtime.PySet, *runtime.PyTuple, *runtime.PyInstance:
		key := ptrOf(val)
		if cached, ok := memo[key]; ok {
			return cached, nil
		}
	}

	switch v := val.(type) {
	case *runtime.PyInstance:
		// Check for __deepcopy__
		for _, cls := range v.Class.Mro {
			if method, ok := cls.Dict["__deepcopy__"]; ok {
				// Pass memo as a PyDict
				memoDict := &runtime.PyDict{Items: make(map[runtime.Value]runtime.Value)}
				args := []runtime.Value{v, memoDict}
				return vm.Call(method, args, nil)
			}
		}
		// Default: deep copy of instance dict/slots
		inst := &runtime.PyInstance{Class: v.Class}
		// Store in memo before recursing to break cycles
		memo[ptrOf(v)] = inst
		if v.Dict != nil {
			inst.Dict = make(map[string]runtime.Value, len(v.Dict))
			for k, origVal := range v.Dict {
				copied, err := deepCopy(vm, origVal, memo)
				if err != nil {
					return nil, err
				}
				inst.Dict[k] = copied
			}
		}
		if v.Slots != nil {
			inst.Slots = make(map[string]runtime.Value, len(v.Slots))
			for k, origVal := range v.Slots {
				copied, err := deepCopy(vm, origVal, memo)
				if err != nil {
					return nil, err
				}
				inst.Slots[k] = copied
			}
		}
		return inst, nil

	case *runtime.PyList:
		result := &runtime.PyList{Items: make([]runtime.Value, len(v.Items))}
		// Store in memo before recursing to break cycles
		memo[ptrOf(v)] = result
		for i, item := range v.Items {
			copied, err := deepCopy(vm, item, memo)
			if err != nil {
				return nil, err
			}
			result.Items[i] = copied
		}
		return result, nil

	case *runtime.PyDict:
		d := &runtime.PyDict{Items: make(map[runtime.Value]runtime.Value, len(v.Items))}
		memo[ptrOf(v)] = d
		for _, k := range v.Keys(vm) {
			origVal, _ := v.DictGet(k, vm)
			copiedVal, err := deepCopy(vm, origVal, memo)
			if err != nil {
				return nil, err
			}
			d.DictSet(k, copiedVal, vm)
		}
		return d, nil

	case *runtime.PySet:
		s := &runtime.PySet{Items: make(map[runtime.Value]struct{}, len(v.Items))}
		memo[ptrOf(v)] = s
		for k := range v.Items {
			s.Items[k] = struct{}{}
		}
		return s, nil

	case *runtime.PyTuple:
		// Deep copy tuple contents (tuples are immutable but may contain mutable objects)
		result := &runtime.PyTuple{Items: make([]runtime.Value, len(v.Items))}
		memo[ptrOf(v)] = result
		for i, item := range v.Items {
			copied, err := deepCopy(vm, item, memo)
			if err != nil {
				return nil, err
			}
			result.Items[i] = copied
		}
		return result, nil

	case *runtime.PyFrozenSet:
		return v, nil

	default:
		// Immutable types (int, float, str, bool, None, bytes) — return as-is
		return val, nil
	}
}
