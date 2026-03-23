package stdlib

import (
	"fmt"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// InitBisectModule registers the bisect module.
func InitBisectModule() {
	runtime.RegisterModule("bisect", func(vm *runtime.VM) *runtime.PyModule {
		mod := runtime.NewModule("bisect")

		mod.Dict["bisect_left"] = &runtime.PyBuiltinFunc{
			Name: "bisect_left",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				return bisectFunc(vm, args, kwargs, true)
			},
		}

		mod.Dict["bisect_right"] = &runtime.PyBuiltinFunc{
			Name: "bisect_right",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				return bisectFunc(vm, args, kwargs, false)
			},
		}

		// bisect is an alias for bisect_right
		mod.Dict["bisect"] = mod.Dict["bisect_right"]

		mod.Dict["insort_left"] = &runtime.PyBuiltinFunc{
			Name: "insort_left",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				return insortFunc(vm, args, kwargs, true)
			},
		}

		mod.Dict["insort_right"] = &runtime.PyBuiltinFunc{
			Name: "insort_right",
			Fn: func(args []runtime.Value, kwargs map[string]runtime.Value) (runtime.Value, error) {
				return insortFunc(vm, args, kwargs, false)
			},
		}

		// insort is an alias for insort_right
		mod.Dict["insort"] = mod.Dict["insort_right"]

		return mod
	})
}

// parseBisectArgs parses (a, x, lo=0, hi=len(a), *, key=None) from args/kwargs.
func parseBisectArgs(vm *runtime.VM, args []runtime.Value, kwargs map[string]runtime.Value) (
	a *runtime.PyList, x runtime.Value, lo, hi int, keyFn runtime.Value, err error,
) {
	if len(args) < 2 || len(args) > 4 {
		return nil, nil, 0, 0, nil, fmt.Errorf("TypeError: expected 2 to 4 positional arguments, got %d", len(args))
	}

	var ok bool
	a, ok = args[0].(*runtime.PyList)
	if !ok {
		return nil, nil, 0, 0, nil, fmt.Errorf("TypeError: first argument must be a list")
	}
	x = args[1]

	lo = 0
	hi = len(a.Items)

	if len(args) >= 3 {
		loVal, ok := args[2].(*runtime.PyInt)
		if !ok {
			return nil, nil, 0, 0, nil, fmt.Errorf("TypeError: lo must be an integer")
		}
		lo = int(loVal.Value)
	}
	if v, ok := kwargs["lo"]; ok {
		loVal, ok := v.(*runtime.PyInt)
		if !ok {
			return nil, nil, 0, 0, nil, fmt.Errorf("TypeError: lo must be an integer")
		}
		lo = int(loVal.Value)
	}

	if len(args) >= 4 {
		hiVal, ok := args[3].(*runtime.PyInt)
		if !ok {
			return nil, nil, 0, 0, nil, fmt.Errorf("TypeError: hi must be an integer")
		}
		hi = int(hiVal.Value)
	}
	if v, ok := kwargs["hi"]; ok {
		if !runtime.IsNone(v) {
			hiVal, ok := v.(*runtime.PyInt)
			if !ok {
				return nil, nil, 0, 0, nil, fmt.Errorf("TypeError: hi must be an integer")
			}
			hi = int(hiVal.Value)
		}
	}

	if v, ok := kwargs["key"]; ok {
		keyFn = v
	}

	if lo < 0 {
		return nil, nil, 0, 0, nil, fmt.Errorf("ValueError: lo must be non-negative")
	}

	return a, x, lo, hi, keyFn, nil
}

// bisectFunc implements bisect_left and bisect_right.
func bisectFunc(vm *runtime.VM, args []runtime.Value, kwargs map[string]runtime.Value, left bool) (runtime.Value, error) {
	a, x, lo, hi, keyFn, err := parseBisectArgs(vm, args, kwargs)
	if err != nil {
		return nil, err
	}

	// Apply key to x
	xKey := x
	if keyFn != nil && !runtime.IsNone(keyFn) {
		xKey, err = vm.Call(keyFn, []runtime.Value{x}, nil)
		if err != nil {
			return nil, err
		}
	}

	idx, err := bisectSearch(vm, a, xKey, lo, hi, keyFn, left)
	if err != nil {
		return nil, err
	}
	return runtime.MakeInt(int64(idx)), nil
}

// bisectSearch performs the binary search.
func bisectSearch(vm *runtime.VM, a *runtime.PyList, xKey runtime.Value, lo, hi int, keyFn runtime.Value, left bool) (int, error) {
	for lo < hi {
		mid := (lo + hi) / 2
		midVal := a.Items[mid]
		midKey := midVal
		if keyFn != nil && !runtime.IsNone(keyFn) {
			var err error
			midKey, err = vm.Call(keyFn, []runtime.Value{midVal}, nil)
			if err != nil {
				return 0, err
			}
		}

		var cond bool
		if left {
			// bisect_left: find first position where midKey >= xKey (i.e., midKey < xKey means go right)
			lt, err := heapLt(vm, midKey, xKey)
			if err != nil {
				return 0, err
			}
			cond = lt
		} else {
			// bisect_right: find first position where midKey > xKey (i.e., xKey < midKey means go left)
			lt, err := heapLt(vm, xKey, midKey)
			if err != nil {
				return 0, err
			}
			cond = !lt
		}

		if cond {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo, nil
}

// insortFunc implements insort_left and insort_right.
func insortFunc(vm *runtime.VM, args []runtime.Value, kwargs map[string]runtime.Value, left bool) (runtime.Value, error) {
	a, x, lo, hi, keyFn, err := parseBisectArgs(vm, args, kwargs)
	if err != nil {
		return nil, err
	}

	// Apply key to x
	xKey := x
	if keyFn != nil && !runtime.IsNone(keyFn) {
		xKey, err = vm.Call(keyFn, []runtime.Value{x}, nil)
		if err != nil {
			return nil, err
		}
	}

	idx, err := bisectSearch(vm, a, xKey, lo, hi, keyFn, left)
	if err != nil {
		return nil, err
	}

	// Insert x at idx
	a.Items = append(a.Items, nil)
	copy(a.Items[idx+1:], a.Items[idx:])
	a.Items[idx] = x

	return runtime.None, nil
}
