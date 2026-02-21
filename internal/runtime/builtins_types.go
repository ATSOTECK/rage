package runtime

import (
	"fmt"
	"unicode/utf8"
)

// initBuiltinsTypes registers type constructors: int, float, complex, str, bool,
// list, tuple, dict, bytes, set, frozenset, slice, and len.
func (vm *VM) initBuiltinsTypes() {
	vm.builtins["len"] = &PyBuiltinFunc{
		Name: "len",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("len() takes exactly one argument (%d given)", len(args))
			}
			switch v := args[0].(type) {
			case *PyString:
				return MakeInt(int64(utf8.RuneCountInString(v.Value))), nil
			case *PyList:
				return MakeInt(int64(len(v.Items))), nil
			case *PyTuple:
				return MakeInt(int64(len(v.Items))), nil
			case *PyDict:
				return MakeInt(int64(len(v.Items))), nil
			case *PySet:
				return MakeInt(int64(len(v.Items))), nil
			case *PyFrozenSet:
				return MakeInt(int64(len(v.Items))), nil
			case *PyBytes:
				return MakeInt(int64(len(v.Value))), nil
			case *PyRange:
				return MakeInt(rangeLen(v)), nil
			case *PyInstance:
				// Check for __len__ method
				if result, found, err := vm.callDunder(v, "__len__"); found {
					if err != nil {
						return nil, err
					}
					if i, ok := result.(*PyInt); ok {
						return i, nil
					}
					return nil, fmt.Errorf("__len__() should return an integer")
				}
				return nil, fmt.Errorf("object of type '%s' has no len()", vm.typeName(v))
			default:
				return nil, fmt.Errorf("object of type '%s' has no len()", vm.typeName(v))
			}
		},
	}

	vm.builtins["slice"] = &PyBuiltinFunc{
		Name: "slice",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			var start, stop, step Value = None, None, None
			switch len(args) {
			case 1:
				stop = args[0]
			case 2:
				start = args[0]
				stop = args[1]
			case 3:
				start = args[0]
				stop = args[1]
				step = args[2]
			default:
				return nil, fmt.Errorf("slice expected 1 to 3 arguments, got %d", len(args))
			}
			return &PySlice{Start: start, Stop: stop, Step: step}, nil
		},
	}

	vm.builtins["int"] = &PyBuiltinFunc{
		Name: "int",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				if _, hasBase := kwargs["base"]; hasBase {
					return nil, fmt.Errorf("TypeError: int() missing string argument")
				}
				return MakeInt(0), nil
			}

			// Check for base argument
			var base int64
			hasBase := false
			if len(args) > 1 {
				b, err := vm.getIntIndex(args[1])
				if err != nil {
					return nil, fmt.Errorf("TypeError: '%s' object cannot be interpreted as an integer", vm.typeName(args[1]))
				}
				base = b
				hasBase = true
			}
			if b, ok := kwargs["base"]; ok {
				base = vm.toInt(b)
				hasBase = true
			}

			if hasBase {
				// Base conversion requires string argument
				s, ok := args[0].(*PyString)
				if !ok {
					return nil, fmt.Errorf("TypeError: int() can't convert non-string with explicit base")
				}
				return vm.intFromStringBase(s.Value, base)
			}

			return vm.tryToIntValue(args[0])
		},
	}

	vm.builtins["float"] = &PyBuiltinFunc{
		Name: "float",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return &PyFloat{Value: 0.0}, nil
			}
			f, err := vm.tryToFloat(args[0])
			if err != nil {
				return nil, err
			}
			return &PyFloat{Value: f}, nil
		},
	}

	vm.builtins["complex"] = &PyBuiltinFunc{
		Name: "complex",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return MakeComplex(0, 0), nil
			}

			// Single string argument: parse complex from string
			if len(args) == 1 {
				if s, ok := args[0].(*PyString); ok {
					return parseComplexString(s.Value)
				}
				// Try __complex__ dunder on instances
				if inst, ok := args[0].(*PyInstance); ok {
					if result, found, err := vm.callDunder(inst, "__complex__"); found {
						if err != nil {
							return nil, err
						}
						if c, ok := result.(*PyComplex); ok {
							return c, nil
						}
						return nil, fmt.Errorf("TypeError: __complex__ returned non-complex (type %s)", vm.typeName(result))
					}
				}
			}

			// String not allowed with 2 args
			if len(args) == 2 {
				if _, ok := args[0].(*PyString); ok {
					return nil, fmt.Errorf("TypeError: complex() can't take second arg if first is a string")
				}
				if _, ok := args[1].(*PyString); ok {
					return nil, fmt.Errorf("TypeError: complex() second arg can't be a string")
				}
			}

			// Convert real part
			realPart := 0.0
			imagPart := 0.0
			if len(args) >= 1 {
				switch v := args[0].(type) {
				case *PyInt:
					realPart = float64(v.Value)
				case *PyFloat:
					realPart = v.Value
				case *PyBool:
					if v.Value {
						realPart = 1
					}
				case *PyComplex:
					realPart = v.Real
					imagPart = v.Imag
				case *PyInstance:
					if result, found, err := vm.callDunder(v, "__float__"); found {
						if err != nil {
							return nil, err
						}
						if f, ok := result.(*PyFloat); ok {
							realPart = f.Value
						} else {
							return nil, fmt.Errorf("TypeError: __float__ returned non-float (type %s)", vm.typeName(result))
						}
					} else {
						return nil, fmt.Errorf("TypeError: complex() first argument must be a string or a number, not '%s'", vm.typeName(args[0]))
					}
				default:
					return nil, fmt.Errorf("TypeError: complex() first argument must be a string or a number, not '%s'", vm.typeName(args[0]))
				}
			}
			// Convert imag part
			if len(args) >= 2 {
				switch v := args[1].(type) {
				case *PyInt:
					imagPart += float64(v.Value)
				case *PyFloat:
					imagPart += v.Value
				case *PyBool:
					if v.Value {
						imagPart += 1
					}
				case *PyComplex:
					// complex(a, b) where b is complex: real += -b.Imag, imag += b.Real
					realPart -= v.Imag
					imagPart += v.Real
				case *PyInstance:
					if result, found, err := vm.callDunder(v, "__float__"); found {
						if err != nil {
							return nil, err
						}
						if f, ok := result.(*PyFloat); ok {
							imagPart += f.Value
						} else {
							return nil, fmt.Errorf("TypeError: __float__ returned non-float (type %s)", vm.typeName(result))
						}
					} else {
						return nil, fmt.Errorf("TypeError: complex() second argument must be a number, not '%s'", vm.typeName(args[1]))
					}
				default:
					return nil, fmt.Errorf("TypeError: complex() second argument must be a number, not '%s'", vm.typeName(args[1]))
				}
			}
			return MakeComplex(realPart, imagPart), nil
		},
	}

	vm.builtins["str"] = &PyBuiltinFunc{
		Name: "str",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return &PyString{Value: ""}, nil
			}
			return &PyString{Value: vm.str(args[0])}, nil
		},
	}

	vm.builtins["bool"] = &PyBuiltinFunc{
		Name: "bool",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return False, nil
			}
			if vm.truthy(args[0]) {
				return True, nil
			}
			return False, nil
		},
	}

	vm.builtins["list"] = &PyBuiltinFunc{
		Name: "list",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return &PyList{Items: []Value{}}, nil
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			return &PyList{Items: items}, nil
		},
	}

	vm.builtins["tuple"] = &PyBuiltinFunc{
		Name: "tuple",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return &PyTuple{Items: []Value{}}, nil
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			return &PyTuple{Items: items}, nil
		},
	}

	vm.builtins["dict"] = &PyBuiltinFunc{
		Name: "dict",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
			if len(args) > 0 {
				switch src := args[0].(type) {
				case *PyDict:
					for k, v := range src.Items {
						d.DictSet(k, v, vm)
					}
				default:
					// Iterable of (key, value) pairs
					items, err := vm.toList(args[0])
					if err != nil {
						return nil, err
					}
					for _, item := range items {
						pair, err := vm.toList(item)
						if err != nil {
							return nil, fmt.Errorf("TypeError: cannot convert dictionary update sequence element to a sequence")
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
			return d, nil
		},
	}

	vm.builtins["bytes"] = &PyBuiltinFunc{
		Name: "bytes",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return &PyBytes{Value: []byte{}}, nil
			}
			switch v := args[0].(type) {
			case *PyBytes:
				// bytes(b"hello") -> copy
				cp := make([]byte, len(v.Value))
				copy(cp, v.Value)
				return &PyBytes{Value: cp}, nil
			case *PyInt:
				// bytes(3) -> b'\x00\x00\x00'
				if v.Value < 0 {
					return nil, fmt.Errorf("ValueError: negative count")
				}
				return &PyBytes{Value: make([]byte, v.Value)}, nil
			case *PyString:
				// bytes("hello", "utf-8") - requires encoding argument
				encoding := ""
				if len(args) > 1 {
					if enc, ok := args[1].(*PyString); ok {
						encoding = enc.Value
					}
				}
				if enc, ok := kwargs["encoding"]; ok {
					if encStr, ok := enc.(*PyString); ok {
						encoding = encStr.Value
					}
				}
				if encoding == "" {
					return nil, fmt.Errorf("TypeError: string argument without an encoding")
				}
				// We only support utf-8/ascii/latin-1 for now
				return &PyBytes{Value: []byte(v.Value)}, nil
			case *PyList:
				// bytes([65, 66, 67]) -> b'ABC'
				result := make([]byte, len(v.Items))
				for i, item := range v.Items {
					n := vm.toInt(item)
					if n < 0 || n > 255 {
						return nil, fmt.Errorf("ValueError: bytes must be in range(0, 256)")
					}
					result[i] = byte(n)
				}
				return &PyBytes{Value: result}, nil
			case *PyTuple:
				// bytes((65, 66, 67)) -> b'ABC'
				result := make([]byte, len(v.Items))
				for i, item := range v.Items {
					n := vm.toInt(item)
					if n < 0 || n > 255 {
						return nil, fmt.Errorf("ValueError: bytes must be in range(0, 256)")
					}
					result[i] = byte(n)
				}
				return &PyBytes{Value: result}, nil
			case *PyInstance:
				// Check for __bytes__ method
				if result, found, err := vm.callDunder(v, "__bytes__"); found {
					if err != nil {
						return nil, err
					}
					if b, ok := result.(*PyBytes); ok {
						return b, nil
					}
					return nil, fmt.Errorf("TypeError: __bytes__ returned non-bytes (type %s)", vm.typeName(result))
				}
				// Fall through to iteration
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, fmt.Errorf("TypeError: cannot convert '%s' object to bytes", vm.typeName(args[0]))
				}
				result := make([]byte, len(items))
				for i, item := range items {
					n := vm.toInt(item)
					if n < 0 || n > 255 {
						return nil, fmt.Errorf("ValueError: bytes must be in range(0, 256)")
					}
					result[i] = byte(n)
				}
				return &PyBytes{Value: result}, nil
			default:
				// Try to iterate
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, fmt.Errorf("TypeError: cannot convert '%s' object to bytes", vm.typeName(args[0]))
				}
				result := make([]byte, len(items))
				for i, item := range items {
					n := vm.toInt(item)
					if n < 0 || n > 255 {
						return nil, fmt.Errorf("ValueError: bytes must be in range(0, 256)")
					}
					result[i] = byte(n)
				}
				return &PyBytes{Value: result}, nil
			}
		},
	}

	vm.builtins["set"] = &PyBuiltinFunc{
		Name: "set",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			s := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			if len(args) > 0 {
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				for _, item := range items {
					if !isHashable(item) {
						return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(item))
					}
					// Use hash-based storage for O(1) lookup
					s.SetAdd(item, vm)
				}
			}
			return s, nil
		},
	}

	vm.builtins["frozenset"] = &PyBuiltinFunc{
		Name: "frozenset",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			fs := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			if len(args) > 0 {
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				for _, item := range items {
					if !isHashable(item) {
						return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(item))
					}
					fs.FrozenSetAdd(item, vm)
				}
			}
			return fs, nil
		},
	}
}
