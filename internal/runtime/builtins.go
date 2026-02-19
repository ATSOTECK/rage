package runtime

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (vm *VM) initBuiltins() {
	vm.builtins["print"] = &PyBuiltinFunc{
		Name: "print",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			sep := " "
			end := "\n"
			if v, ok := kwargs["sep"]; ok {
				if v == None {
					// sep=None means default
				} else if s, ok := v.(*PyString); ok {
					sep = s.Value
				} else {
					return nil, fmt.Errorf("sep must be None or a string, not %s", vm.typeName(v))
				}
			}
			if v, ok := kwargs["end"]; ok {
				if v == None {
					// end=None means default
				} else if s, ok := v.(*PyString); ok {
					end = s.Value
				} else {
					return nil, fmt.Errorf("end must be None or a string, not %s", vm.typeName(v))
				}
			}
			if v, ok := kwargs["flush"]; ok {
				_ = v // accepted but no-op (stdout is unbuffered)
			}
			if v, ok := kwargs["file"]; ok {
				_ = v // accepted but ignored (no file I/O yet)
			}
			for i, arg := range args {
				if i > 0 {
					fmt.Print(sep)
				}
				fmt.Print(vm.str(arg))
			}
			fmt.Print(end)
			return None, nil
		},
	}

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

	vm.builtins["range"] = &PyBuiltinFunc{
		Name: "range",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			var start, stop, step int64 = 0, 0, 1
			// Validate all arguments are integers (not floats)
			for i, arg := range args {
				switch arg.(type) {
				case *PyInt, *PyBool:
					// ok
				default:
					_ = i
					return nil, fmt.Errorf("TypeError: '%s' object cannot be interpreted as an integer", vm.typeName(arg))
				}
			}
			switch len(args) {
			case 1:
				stop = vm.toInt(args[0])
			case 2:
				start = vm.toInt(args[0])
				stop = vm.toInt(args[1])
			case 3:
				start = vm.toInt(args[0])
				stop = vm.toInt(args[1])
				step = vm.toInt(args[2])
			default:
				return nil, fmt.Errorf("range expected 1 to 3 arguments, got %d", len(args))
			}
			if step == 0 {
				return nil, fmt.Errorf("ValueError: range() arg 3 must not be zero")
			}
			return &PyRange{Start: start, Stop: stop, Step: step}, nil
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

	vm.builtins["repr"] = &PyBuiltinFunc{
		Name: "repr",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("repr() takes exactly 1 argument (%d given)", len(args))
			}
			return &PyString{Value: vm.repr(args[0])}, nil
		},
	}

	vm.builtins["ascii"] = &PyBuiltinFunc{
		Name: "ascii",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("ascii() takes exactly 1 argument (%d given)", len(args))
			}
			return &PyString{Value: vm.ascii(args[0])}, nil
		},
	}

	vm.builtins["__import__"] = &PyBuiltinFunc{
		Name: "__import__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			// __import__(name, globals=None, locals=None, fromlist=(), level=0)
			if len(args) < 1 {
				return nil, fmt.Errorf("TypeError: __import__() missing required argument: 'name'")
			}
			nameStr, ok := args[0].(*PyString)
			if !ok {
				return nil, fmt.Errorf("TypeError: __import__() argument 1 must be str, not %s", vm.typeName(args[0]))
			}
			moduleName := nameStr.Value

			// Extract globals (arg 1 or kwarg) for relative import resolution
			var globalsDict map[string]Value
			if len(args) > 1 {
				if d, ok := args[1].(*PyDict); ok {
					globalsDict = make(map[string]Value)
					for k, v := range d.Items {
						if ks, ok := k.(*PyString); ok {
							globalsDict[ks.Value] = v
						}
					}
				}
			}
			if globalsDict == nil && vm.frame != nil {
				globalsDict = vm.frame.Globals
			}

			// Extract fromlist (arg 3 or kwarg)
			var fromlistItems []string
			var fromlistVal Value
			if len(args) > 3 {
				fromlistVal = args[3]
			}
			if v, ok := kwargs["fromlist"]; ok {
				fromlistVal = v
			}
			if fromlistVal != nil && fromlistVal != None {
				switch fl := fromlistVal.(type) {
				case *PyTuple:
					for _, item := range fl.Items {
						if s, ok := item.(*PyString); ok {
							fromlistItems = append(fromlistItems, s.Value)
						}
					}
				case *PyList:
					for _, item := range fl.Items {
						if s, ok := item.(*PyString); ok {
							fromlistItems = append(fromlistItems, s.Value)
						}
					}
				}
			}

			// Extract level (arg 4 or kwarg)
			level := 0
			if len(args) > 4 {
				if li, ok := args[4].(*PyInt); ok {
					level = int(li.Value)
				}
			}
			if v, ok := kwargs["level"]; ok {
				if li, ok := v.(*PyInt); ok {
					level = int(li.Value)
				}
			}

			// Resolve relative imports
			resolvedName := moduleName
			if level > 0 {
				packageName := ""
				if globalsDict != nil {
					if pkgVal, ok := globalsDict["__package__"]; ok {
						if pkgStr, ok := pkgVal.(*PyString); ok {
							packageName = pkgStr.Value
						}
					}
					if packageName == "" {
						if nameVal, ok := globalsDict["__name__"]; ok {
							if nameStr, ok := nameVal.(*PyString); ok {
								packageName = nameStr.Value
							}
						}
					}
				}
				resolved, err := ResolveRelativeImport(moduleName, level, packageName)
				if err != nil {
					return nil, err
				}
				resolvedName = resolved
			}

			// Import each part of dotted name
			var rootMod, targetMod *PyModule
			parts := splitModuleName(resolvedName)
			for i := range parts {
				partialName := joinModuleName(parts[:i+1])
				mod, err := vm.ImportModule(partialName)
				if err != nil {
					return nil, err
				}
				if i == 0 {
					rootMod = mod
				}
				targetMod = mod
			}

			// fromlist non-empty → return target module; empty → return root module
			if len(fromlistItems) > 0 {
				return targetMod, nil
			}
			return rootMod, nil
		},
	}

	vm.builtins["format"] = &PyBuiltinFunc{
		Name: "format",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("TypeError: format() takes 1 or 2 arguments (%d given)", len(args))
			}
			spec := ""
			if len(args) == 2 {
				if s, ok := args[1].(*PyString); ok {
					spec = s.Value
				} else {
					return nil, fmt.Errorf("TypeError: format() argument 2 must be str, not %s", vm.typeName(args[1]))
				}
			}
			result, err := vm.formatValue(args[0], spec)
			if err != nil {
				return nil, err
			}
			return &PyString{Value: result}, nil
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

	vm.builtins["isinstance"] = &PyBuiltinFunc{
		Name: "isinstance",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("isinstance() takes exactly 2 arguments")
			}
			obj := args[0]
			classInfo := args[1]

			// Check for __instancecheck__ on metaclass (search MRO)
			if cls, ok := classInfo.(*PyClass); ok && cls.Metaclass != nil {
				typeClass, _ := vm.builtins["type"].(*PyClass)
				for _, metaCls := range cls.Metaclass.Mro {
					if metaCls == typeClass || metaCls.Name == "object" {
						continue
					}
					if method, hasMethod := metaCls.Dict["__instancecheck__"]; hasMethod {
						allArgs := []Value{cls, obj}
						var result Value
						var err error
						switch fn := method.(type) {
						case *PyFunction:
							result, err = vm.callFunction(fn, allArgs, nil)
						case *PyBuiltinFunc:
							result, err = fn.Fn(allArgs, nil)
						}
						if err != nil {
							return nil, err
						}
						if result != nil {
							if vm.truthy(result) {
								return True, nil
							}
							return False, nil
						}
						break
					}
				}
			}

			// Helper to check if obj is instance of a type by name
			checkTypeName := func(typeName string) bool {
				switch o := obj.(type) {
				case *PyBool:
					// bool is a subclass of int in Python
					return typeName == "bool" || typeName == "int" || typeName == "object"
				case *PyInt:
					return typeName == "int" || typeName == "object"
				case *PyFloat:
					return typeName == "float" || typeName == "object"
				case *PyComplex:
					return typeName == "complex" || typeName == "object"
				case *PyString:
					return typeName == "str" || typeName == "object"
				case *PyList:
					return typeName == "list" || typeName == "object"
				case *PyTuple:
					return typeName == "tuple" || typeName == "object"
				case *PyDict:
					return typeName == "dict" || typeName == "object"
				case *PySet:
					return typeName == "set" || typeName == "object"
				case *PyFrozenSet:
					return typeName == "frozenset" || typeName == "object"
				case *PyBytes:
					return typeName == "bytes" || typeName == "object"
				case *PyNone:
					return typeName == "NoneType" || typeName == "object"
				case *PyInstance:
					return typeName == o.Class.Name || typeName == "object"
				case *PyException:
					return typeName == o.Type() || typeName == "Exception" || typeName == "BaseException" || typeName == "object"
				}
				return false
			}

			// Helper to check if obj is instance of a single class
			checkClass := func(cls *PyClass) bool {
				switch o := obj.(type) {
				case *PyInstance:
					if vm.isInstanceOf(o, cls) {
						return true
					}
					// Check registered virtual subclasses
					if len(cls.RegisteredSubclasses) > 0 {
						for _, reg := range cls.RegisteredSubclasses {
							if vm.isInstanceOf(o, reg) {
								return true
							}
						}
					}
					return false
				case *PyException:
					if vm.isExceptionClass(cls) {
						return vm.exceptionMatches(o, cls)
					}
				}
				// For built-in types, check by name
				return checkTypeName(cls.Name)
			}

			// Recursive check for a single type specification (handles nested tuples)
			var checkType func(typeSpec Value) (bool, error)
			checkType = func(typeSpec Value) (bool, error) {
				switch t := typeSpec.(type) {
				case *PyClass:
					return checkClass(t), nil
				case *PyBuiltinFunc:
					return checkTypeName(t.Name), nil
				case *PyTuple:
					for _, item := range t.Items {
						match, err := checkType(item)
						if err != nil {
							return false, err
						}
						if match {
							return true, nil
						}
					}
					return false, nil
				default:
					return false, fmt.Errorf("TypeError: isinstance() arg 2 must be a type, a tuple of types, or a union")
				}
			}

			match, err := checkType(classInfo)
			if err != nil {
				return nil, err
			}
			if match {
				return True, nil
			}
			return False, nil
		},
	}

	vm.builtins["abs"] = &PyBuiltinFunc{
		Name: "abs",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("abs() takes exactly one argument")
			}
			switch v := args[0].(type) {
			case *PyInt:
				if v.Value < 0 {
					return MakeInt(-v.Value), nil
				}
				return v, nil
			case *PyFloat:
				return &PyFloat{Value: math.Abs(v.Value)}, nil
			case *PyComplex:
				return &PyFloat{Value: math.Sqrt(v.Real*v.Real + v.Imag*v.Imag)}, nil
			case *PyInstance:
				// Check for __abs__ method
				if result, found, err := vm.callDunder(v, "__abs__"); found {
					return result, err
				}
				return nil, fmt.Errorf("bad operand type for abs(): '%s'", vm.typeName(v))
			default:
				return nil, fmt.Errorf("bad operand type for abs()")
			}
		},
	}

	vm.builtins["hash"] = &PyBuiltinFunc{
		Name: "hash",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("hash() takes exactly one argument (%d given)", len(args))
			}
			if !isHashable(args[0]) {
				return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(args[0]))
			}
			return MakeInt(int64(vm.hashValueVM(args[0]))), nil
		},
	}

	vm.builtins["min"] = &PyBuiltinFunc{
		Name: "min",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("min expected at least 1 argument")
			}
			keyFn, hasKey := kwargs["key"]
			defaultVal, hasDefault := kwargs["default"]
			if len(args) == 1 {
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				args = items
			}
			if len(args) == 0 {
				if hasDefault {
					return defaultVal, nil
				}
				return nil, &PyException{TypeName: "ValueError", Message: "min() arg is an empty sequence"}
			}
			minVal := args[0]
			for _, v := range args[1:] {
				if hasKey {
					kv, err := vm.call(keyFn, []Value{v}, nil)
					if err != nil {
						return nil, err
					}
					km, err := vm.call(keyFn, []Value{minVal}, nil)
					if err != nil {
						return nil, err
					}
					if vm.compare(kv, km) < 0 {
						minVal = v
					}
				} else if vm.compare(v, minVal) < 0 {
					minVal = v
				}
			}
			return minVal, nil
		},
	}

	vm.builtins["max"] = &PyBuiltinFunc{
		Name: "max",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("max expected at least 1 argument")
			}
			keyFn, hasKey := kwargs["key"]
			defaultVal, hasDefault := kwargs["default"]
			if len(args) == 1 {
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				args = items
			}
			if len(args) == 0 {
				if hasDefault {
					return defaultVal, nil
				}
				return nil, &PyException{TypeName: "ValueError", Message: "max() arg is an empty sequence"}
			}
			maxVal := args[0]
			for _, v := range args[1:] {
				if hasKey {
					kv, err := vm.call(keyFn, []Value{v}, nil)
					if err != nil {
						return nil, err
					}
					km, err := vm.call(keyFn, []Value{maxVal}, nil)
					if err != nil {
						return nil, err
					}
					if vm.compare(kv, km) > 0 {
						maxVal = v
					}
				} else if vm.compare(v, maxVal) > 0 {
					maxVal = v
				}
			}
			return maxVal, nil
		},
	}

	vm.builtins["sum"] = &PyBuiltinFunc{
		Name: "sum",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("sum expected at least 1 argument")
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			var start Value = MakeInt(0)
			if len(args) > 1 {
				start = args[1]
			}
			result := start
			for _, item := range items {
				result, err = vm.binaryOp(OpBinaryAdd, result, item)
				if err != nil {
					return nil, err
				}
			}
			return result, nil
		},
	}

	vm.builtins["input"] = &PyBuiltinFunc{
		Name: "input",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) > 0 {
				fmt.Print(vm.str(args[0]))
			}
			var line string
			fmt.Scanln(&line)
			return &PyString{Value: line}, nil
		},
	}

	vm.builtins["ord"] = &PyBuiltinFunc{
		Name: "ord",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: ord() takes exactly one argument")
			}
			s, ok := args[0].(*PyString)
			if !ok {
				return nil, fmt.Errorf("TypeError: ord() expected string of length 1, but %s found", vm.typeName(args[0]))
			}
			runes := []rune(s.Value)
			if len(runes) != 1 {
				return nil, fmt.Errorf("TypeError: ord() expected a character, but string of length %d found", len(runes))
			}
			return MakeInt(int64(runes[0])), nil
		},
	}

	vm.builtins["chr"] = &PyBuiltinFunc{
		Name: "chr",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("chr() takes exactly one argument")
			}
			i := vm.toInt(args[0])
			return &PyString{Value: string(rune(i))}, nil
		},
	}

	// enumerate(iterable, start=0) - returns iterator of (index, value) tuples
	vm.builtins["enumerate"] = &PyBuiltinFunc{
		Name: "enumerate",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("enumerate expected 1 to 2 arguments, got %d", len(args))
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			var start int64 = 0
			if len(args) == 2 {
				start = vm.toInt(args[1])
			}
			if s, ok := kwargs["start"]; ok {
				start = vm.toInt(s)
			}
			result := make([]Value, len(items))
			for i, item := range items {
				result[i] = &PyTuple{Items: []Value{MakeInt(start + int64(i)), item}}
			}
			return &PyIterator{Items: result, Index: 0}, nil
		},
	}

	// zip(*iterables) - returns iterator of tuples
	vm.builtins["zip"] = &PyBuiltinFunc{
		Name: "zip",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return &PyIterator{Items: []Value{}, Index: 0}, nil
			}
			// Get iterators for all args
			iters := make([]Value, len(args))
			for i, arg := range args {
				it, err := vm.getIter(arg)
				if err != nil {
					return nil, fmt.Errorf("zip argument #%d is not iterable", i+1)
				}
				iters[i] = it
			}
			// Iterate lazily, collecting tuples until any iterator is exhausted
			var result []Value
			for {
				tuple := make([]Value, len(iters))
				for j, it := range iters {
					val, done, err := vm.iterNext(it)
					if err != nil {
						return nil, err
					}
					if done {
						// This iterator is exhausted, stop
						return &PyIterator{Items: result, Index: 0}, nil
					}
					tuple[j] = val
				}
				result = append(result, &PyTuple{Items: tuple})
			}
		},
	}

	// map(func, *iterables) - applies function to items from iterables
	vm.builtins["map"] = &PyBuiltinFunc{
		Name: "map",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("map() must have at least two arguments")
			}
			fn := args[0]
			// Convert remaining args to lists
			lists := make([][]Value, len(args)-1)
			minLen := -1
			for i, arg := range args[1:] {
				items, err := vm.toList(arg)
				if err != nil {
					return nil, err
				}
				lists[i] = items
				if minLen == -1 || len(items) < minLen {
					minLen = len(items)
				}
			}
			// Apply function to each set of items
			result := make([]Value, minLen)
			for i := 0; i < minLen; i++ {
				fnArgs := make([]Value, len(lists))
				for j, list := range lists {
					fnArgs[j] = list[i]
				}
				val, err := vm.call(fn, fnArgs, nil)
				if err != nil {
					return nil, err
				}
				result[i] = val
			}
			return &PyIterator{Items: result, Index: 0}, nil
		},
	}

	// filter(func, iterable) - filters items based on function
	vm.builtins["filter"] = &PyBuiltinFunc{
		Name: "filter",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("filter expected 2 arguments, got %d", len(args))
			}
			fn := args[0]
			items, err := vm.toList(args[1])
			if err != nil {
				return nil, err
			}
			var result []Value
			for _, item := range items {
				var keep bool
				if fn == None {
					// If function is None, filter by truthiness
					keep = vm.truthy(item)
				} else {
					val, err := vm.call(fn, []Value{item}, nil)
					if err != nil {
						return nil, err
					}
					keep = vm.truthy(val)
				}
				if keep {
					result = append(result, item)
				}
			}
			return &PyIterator{Items: result, Index: 0}, nil
		},
	}

	// reversed(seq) - returns reversed iterator
	vm.builtins["reversed"] = &PyBuiltinFunc{
		Name: "reversed",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("reversed() takes exactly one argument (%d given)", len(args))
			}
			// Try __reversed__ dunder on PyInstance first
			if inst, ok := args[0].(*PyInstance); ok {
				if result, found, err := vm.callDunder(inst, "__reversed__"); found {
					if err != nil {
						return nil, err
					}
					return result, nil
				}
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			// Create reversed copy
			result := make([]Value, len(items))
			for i, item := range items {
				result[len(items)-1-i] = item
			}
			return &PyIterator{Items: result, Index: 0}, nil
		},
	}

	// sorted(iterable, key=None, reverse=False) - returns sorted list
	vm.builtins["sorted"] = &PyBuiltinFunc{
		Name: "sorted",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("sorted expected 1 argument, got %d", len(args))
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			// Make a copy to avoid modifying the original
			result := make([]Value, len(items))
			copy(result, items)

			// Check for key function
			var keyFn Value
			if k, ok := kwargs["key"]; ok && k != None {
				keyFn = k
			}

			// Check for reverse flag
			reverse := false
			if r, ok := kwargs["reverse"]; ok {
				reverse = vm.truthy(r)
			}

			// Sort using comparison
			var sortErr error
			sort.SliceStable(result, func(i, j int) bool {
				if sortErr != nil {
					return false
				}
				a, b := result[i], result[j]
				// Apply key function if provided
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
				cmpResult := vm.compareOp(OpCompareLt, cmpA, cmpB)
				if cmpResult == nil {
					// compareOp set vm.currentException (e.g. TypeError for incompatible types)
					if vm.currentException != nil {
						sortErr = vm.currentException
						vm.currentException = nil
					}
					return false
				}
				return vm.truthy(cmpResult)
			})
			if sortErr != nil {
				return nil, sortErr
			}
			return &PyList{Items: result}, nil
		},
	}

	// all(iterable) - returns True if all elements are truthy
	vm.builtins["all"] = &PyBuiltinFunc{
		Name: "all",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("all() takes exactly one argument (%d given)", len(args))
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			for _, item := range items {
				if !vm.truthy(item) {
					return False, nil
				}
			}
			return True, nil
		},
	}

	// any(iterable) - returns True if any element is truthy
	vm.builtins["any"] = &PyBuiltinFunc{
		Name: "any",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("any() takes exactly one argument (%d given)", len(args))
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			for _, item := range items {
				if vm.truthy(item) {
					return True, nil
				}
			}
			return False, nil
		},
	}

	// hasattr(obj, name) - returns True if object has the attribute
	vm.builtins["hasattr"] = &PyBuiltinFunc{
		Name: "hasattr",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("hasattr() takes exactly 2 arguments (%d given)", len(args))
			}
			name, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("attribute name must be string, not '%s'", vm.typeName(args[1]))
			}
			_, err := vm.getAttr(args[0], name.Value)
			if err != nil {
				// Check if builtin type has this dunder
				if typeName := builtinValueTypeName(args[0]); typeName != "" {
					if builtinHasDunder(typeName, name.Value) {
						return True, nil
					}
				}
				return False, nil
			}
			return True, nil
		},
	}

	// dir([obj]) - list attributes of object or names in current scope
	vm.builtins["dir"] = &PyBuiltinFunc{
		Name: "dir",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) > 1 {
				return nil, fmt.Errorf("dir expected at most 1 argument, got %d", len(args))
			}
			if len(args) == 0 {
				// No argument: return sorted names from current scope
				names := make(map[string]bool)
				if vm.frame != nil {
					for k := range vm.frame.Globals {
						names[k] = true
					}
					for k := range vm.builtins {
						names[k] = true
					}
				}
				return vm.sortedNameList(names), nil
			}

			obj := args[0]

			// Check for __dir__ on instances
			if inst, ok := obj.(*PyInstance); ok {
				if result, found, err := vm.callDunder(inst, "__dir__"); found {
					if err != nil {
						return nil, err
					}
					return result, nil
				}
			}

			// Default: collect attributes
			names := make(map[string]bool)
			switch o := obj.(type) {
			case *PyInstance:
				if o.Dict != nil {
					for k := range o.Dict {
						names[k] = true
					}
				}
				if o.Slots != nil {
					for k := range o.Slots {
						names[k] = true
					}
				}
				for _, cls := range o.Class.Mro {
					for k := range cls.Dict {
						names[k] = true
					}
				}
			case *PyClass:
				for _, cls := range o.Mro {
					for k := range cls.Dict {
						names[k] = true
					}
				}
			case *PyModule:
				for k := range o.Dict {
					names[k] = true
				}
			case *PyDict:
				for _, item := range o.Items {
					if ks, ok := item.(*PyString); ok {
						names[ks.Value] = true
					}
				}
			}
			return vm.sortedNameList(names), nil
		},
	}

	// getattr(obj, name[, default]) - get attribute from object
	vm.builtins["getattr"] = &PyBuiltinFunc{
		Name: "getattr",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 2 || len(args) > 3 {
				return nil, fmt.Errorf("getattr expected 2 or 3 arguments, got %d", len(args))
			}
			name, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("attribute name must be string, not '%s'", vm.typeName(args[1]))
			}
			val, err := vm.getAttr(args[0], name.Value)
			if err != nil {
				if len(args) == 3 {
					return args[2], nil // Return default
				}
				return nil, fmt.Errorf("'%s' object has no attribute '%s'", vm.typeName(args[0]), name.Value)
			}
			return val, nil
		},
	}

	// setattr(obj, name, value) - set attribute on object
	vm.builtins["setattr"] = &PyBuiltinFunc{
		Name: "setattr",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("setattr() takes exactly 3 arguments (%d given)", len(args))
			}
			name, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("attribute name must be string, not '%s'", vm.typeName(args[1]))
			}
			err := vm.setAttr(args[0], name.Value, args[2])
			if err != nil {
				return nil, err
			}
			return None, nil
		},
	}

	// delattr(obj, name) - delete attribute from object
	vm.builtins["delattr"] = &PyBuiltinFunc{
		Name: "delattr",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("delattr() takes exactly 2 arguments (%d given)", len(args))
			}
			name, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("attribute name must be string, not '%s'", vm.typeName(args[1]))
			}
			if err := vm.delAttr(args[0], name.Value); err != nil {
				return nil, err
			}
			return None, nil
		},
	}

	// pow(base, exp[, mod]) - power with optional modulo
	vm.builtins["pow"] = &PyBuiltinFunc{
		Name: "pow",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 2 || len(args) > 3 {
				return nil, fmt.Errorf("pow expected 2 or 3 arguments, got %d", len(args))
			}
			base := vm.toFloat(args[0])
			exp := vm.toFloat(args[1])

			if len(args) == 3 {
				// Modular exponentiation - requires integers
				baseInt := vm.toInt(args[0])
				expInt := vm.toInt(args[1])
				modInt := vm.toInt(args[2])
				if modInt == 0 {
					return nil, fmt.Errorf("pow() 3rd argument cannot be 0")
				}
				// Simple modular exponentiation
				result := int64(1)
				baseInt = baseInt % modInt
				for expInt > 0 {
					if expInt%2 == 1 {
						result = (result * baseInt) % modInt
					}
					expInt = expInt / 2
					baseInt = (baseInt * baseInt) % modInt
				}
				return MakeInt(result), nil
			}

			result := math.Pow(base, exp)
			// Return int if result is a whole number and inputs were ints
			_, baseIsInt := args[0].(*PyInt)
			_, expIsInt := args[1].(*PyInt)
			if baseIsInt && expIsInt && result == float64(int64(result)) {
				return MakeInt(int64(result)), nil
			}
			return &PyFloat{Value: result}, nil
		},
	}

	// divmod(a, b) - returns (quotient, remainder)
	vm.builtins["divmod"] = &PyBuiltinFunc{
		Name: "divmod",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("divmod expected 2 arguments, got %d", len(args))
			}
			// Check if both are integers
			aInt, aIsInt := args[0].(*PyInt)
			bInt, bIsInt := args[1].(*PyInt)
			if aIsInt && bIsInt {
				if bInt.Value == 0 {
					return nil, &PyException{TypeName: "ZeroDivisionError", Message: "integer division or modulo by zero"}
				}
				q := aInt.Value / bInt.Value
				r := aInt.Value % bInt.Value
				// Python's divmod uses floor division
				if r != 0 && (aInt.Value < 0) != (bInt.Value < 0) {
					q--
					r += bInt.Value
				}
				return &PyTuple{Items: []Value{MakeInt(q), MakeInt(r)}}, nil
			}
			// Float division
			a := vm.toFloat(args[0])
			b := vm.toFloat(args[1])
			if b == 0 {
				return nil, &PyException{TypeName: "ZeroDivisionError", Message: "float division by zero"}
			}
			q := math.Floor(a / b)
			r := a - q*b
			return &PyTuple{Items: []Value{&PyFloat{Value: q}, &PyFloat{Value: r}}}, nil
		},
	}

	// hex(x) - convert integer to hexadecimal string
	vm.builtins["hex"] = &PyBuiltinFunc{
		Name: "hex",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("hex() takes exactly one argument (%d given)", len(args))
			}
			n, err := vm.getIntIndex(args[0])
			if err != nil {
				return nil, err
			}
			if n < 0 {
				return &PyString{Value: fmt.Sprintf("-0x%x", -n)}, nil
			}
			return &PyString{Value: fmt.Sprintf("0x%x", n)}, nil
		},
	}

	// oct(x) - convert integer to octal string
	vm.builtins["oct"] = &PyBuiltinFunc{
		Name: "oct",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("oct() takes exactly one argument (%d given)", len(args))
			}
			n, err := vm.getIntIndex(args[0])
			if err != nil {
				return nil, err
			}
			if n < 0 {
				return &PyString{Value: fmt.Sprintf("-0o%o", -n)}, nil
			}
			return &PyString{Value: fmt.Sprintf("0o%o", n)}, nil
		},
	}

	// bin(x) - convert integer to binary string
	vm.builtins["bin"] = &PyBuiltinFunc{
		Name: "bin",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("bin() takes exactly one argument (%d given)", len(args))
			}
			n, err := vm.getIntIndex(args[0])
			if err != nil {
				return nil, err
			}
			if n < 0 {
				return &PyString{Value: fmt.Sprintf("-0b%b", -n)}, nil
			}
			return &PyString{Value: fmt.Sprintf("0b%b", n)}, nil
		},
	}

	// round(number[, ndigits]) - round to given precision
	vm.builtins["round"] = &PyBuiltinFunc{
		Name: "round",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("round() takes 1 or 2 arguments (%d given)", len(args))
			}
			// Try __round__ dunder on PyInstance first
			if inst, ok := args[0].(*PyInstance); ok {
				var dunderArgs []Value
				if len(args) == 2 {
					dunderArgs = []Value{args[1]}
				}
				if result, found, err := vm.callDunder(inst, "__round__", dunderArgs...); found {
					if err != nil {
						return nil, err
					}
					return result, nil
				}
				return nil, fmt.Errorf("TypeError: type %s doesn't define __round__ method", vm.typeName(args[0]))
			}
			num := vm.toFloat(args[0])

			if len(args) == 1 {
				// Round to integer - use banker's rounding (round half to even)
				rounded := math.RoundToEven(num)
				return MakeInt(int64(rounded)), nil
			}

			// Round to ndigits decimal places
			ndigits := vm.toInt(args[1])
			multiplier := math.Pow(10, float64(ndigits))
			rounded := math.RoundToEven(num*multiplier) / multiplier

			// If ndigits is negative, return int
			if ndigits < 0 {
				return MakeInt(int64(rounded)), nil
			}
			return &PyFloat{Value: rounded}, nil
		},
	}

	// callable(obj) - check if object is callable
	vm.builtins["callable"] = &PyBuiltinFunc{
		Name: "callable",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("callable() takes exactly one argument (%d given)", len(args))
			}
			switch args[0].(type) {
			case *PyFunction, *PyBuiltinFunc, *PyGoFunc, *PyMethod, *PyClass:
				return True, nil
			case *PyInstance:
				// Check if instance has __call__ method
				inst := args[0].(*PyInstance)
				for _, cls := range inst.Class.Mro {
					if _, ok := cls.Dict["__call__"]; ok {
						return True, nil
					}
				}
				return False, nil
			default:
				return False, nil
			}
		},
	}

	// next(iterator[, default]) - retrieve next item from iterator
	vm.builtins["next"] = &PyBuiltinFunc{
		Name: "next",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("TypeError: next expected 1 or 2 arguments, got %d", len(args))
			}
			hasDefault := len(args) == 2

			switch it := args[0].(type) {
			case *PyGenerator:
				val, done, err := vm.GeneratorSend(it, None)
				if done || err != nil {
					if hasDefault {
						return args[1], nil
					}
					if err != nil {
						return nil, err
					}
					return nil, &PyException{TypeName: "StopIteration", Message: ""}
				}
				return val, nil
			case *PyIterator:
				items := it.Items
				if it.Source != nil {
					items = it.Source.Items
				}
				if it.Index >= len(items) {
					if hasDefault {
						return args[1], nil
					}
					return nil, &PyException{TypeName: "StopIteration", Message: ""}
				}
				val := items[it.Index]
				it.Index++
				return val, nil
			default:
				// Try __next__ method
				nextMethod, err := vm.getAttr(args[0], "__next__")
				if err != nil {
					return nil, fmt.Errorf("TypeError: '%s' object is not an iterator", vm.typeName(args[0]))
				}
				result, err := vm.call(nextMethod, nil, nil)
				if err != nil {
					if hasDefault {
						if pyExc, ok := err.(*PyException); ok && pyExc.Type() == "StopIteration" {
							return args[1], nil
						}
					}
					return nil, err
				}
				return result, nil
			}
		},
	}

	// iter(object) - get an iterator from an object
	vm.builtins["iter"] = &PyBuiltinFunc{
		Name: "iter",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: iter expected 1 argument, got %d", len(args))
			}
			return vm.getIter(args[0])
		},
	}

	// issubclass(cls, classinfo) - check if class is a subclass
	vm.builtins["issubclass"] = &PyBuiltinFunc{
		Name: "issubclass",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("TypeError: issubclass expected 2 arguments, got %d", len(args))
			}

			// Get the class name from arg 1
			getClassName := func(v Value) (string, bool) {
				switch t := v.(type) {
				case *PyClass:
					return t.Name, true
				case *PyBuiltinFunc:
					return t.Name, true
				}
				return "", false
			}

			// Builtin type hierarchy
			builtinSubclass := func(child, parent string) bool {
				if child == parent {
					return true
				}
				// bool is a subclass of int
				if child == "bool" && parent == "int" {
					return true
				}
				// Everything is a subclass of object
				if parent == "object" {
					return true
				}
				return false
			}

			// Check for __subclasscheck__ on metaclass of arg 2 (search MRO)
			if targetCls, ok := args[1].(*PyClass); ok && targetCls.Metaclass != nil {
				typeClass, _ := vm.builtins["type"].(*PyClass)
				for _, metaCls := range targetCls.Metaclass.Mro {
					if metaCls == typeClass || metaCls.Name == "object" {
						continue
					}
					if method, hasMethod := metaCls.Dict["__subclasscheck__"]; hasMethod {
						allArgs := []Value{targetCls, args[0]}
						var result Value
						var err error
						switch fn := method.(type) {
						case *PyFunction:
							result, err = vm.callFunction(fn, allArgs, nil)
						case *PyBuiltinFunc:
							result, err = fn.Fn(allArgs, nil)
						}
						if err != nil {
							return nil, err
						}
						if result != nil {
							if vm.truthy(result) {
								return True, nil
							}
							return False, nil
						}
						break
					}
				}
			}

			clsName, ok := getClassName(args[0])
			if !ok {
				return nil, fmt.Errorf("TypeError: issubclass() arg 1 must be a class")
			}

			// Check single target class
			checkSingleTarget := func(cls *PyClass, target *PyClass) bool {
				for _, mroClass := range cls.Mro {
					if mroClass == target {
						return true
					}
				}
				if len(target.RegisteredSubclasses) > 0 {
					for _, reg := range target.RegisteredSubclasses {
						for _, mroClass := range cls.Mro {
							if mroClass == reg {
								return true
							}
						}
					}
				}
				return false
			}

			// Recursive check that handles nested tuples
			var checkTarget func(arg1 Value, arg2 Value) (Value, error)
			checkTarget = func(arg1 Value, arg2 Value) (Value, error) {
				switch target := arg2.(type) {
				case *PyClass:
					if cls, ok := arg1.(*PyClass); ok {
						if checkSingleTarget(cls, target) {
							return True, nil
						}
						return vm.toValue(builtinSubclass(clsName, target.Name)), nil
					}
					return vm.toValue(builtinSubclass(clsName, target.Name)), nil
				case *PyBuiltinFunc:
					return vm.toValue(builtinSubclass(clsName, target.Name)), nil
				case *PyTuple:
					for _, item := range target.Items {
						result, err := checkTarget(arg1, item)
						if err != nil {
							return nil, err
						}
						if result == True {
							return True, nil
						}
					}
					return False, nil
				default:
					return nil, fmt.Errorf("TypeError: issubclass() arg 2 must be a class or tuple of classes")
				}
			}

			return checkTarget(args[0], args[1])
		},
	}

	vm.builtins["None"] = None
	vm.builtins["True"] = True
	vm.builtins["False"] = False
	vm.builtins["NotImplemented"] = NotImplemented

	// __build_class__ is used to create classes
	vm.builtins["__build_class__"] = &PyBuiltinFunc{
		Name: "__build_class__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("__build_class__: not enough arguments")
			}

			// First arg is the class body function
			bodyFunc, ok := args[0].(*PyFunction)
			if !ok {
				return nil, fmt.Errorf("__build_class__: first argument must be a function")
			}

			// Second arg is the class name
			nameVal, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("__build_class__: second argument must be a string")
			}
			className := nameVal.Value

			// Remaining args are base classes — resolve __mro_entries__ for non-class bases
			originalBases := args[2:]
			var bases []*PyClass
			for _, baseArg := range originalBases {
				if base, ok := baseArg.(*PyClass); ok {
					bases = append(bases, base)
					continue
				}
				// Try __mro_entries__ for non-class bases (e.g. GenericAlias)
				origTuple := &PyTuple{Items: make([]Value, len(originalBases))}
				copy(origTuple.Items, originalBases)
				if mroEntries, err := vm.getAttr(baseArg, "__mro_entries__"); err == nil {
					if result, callErr := vm.call(mroEntries, []Value{origTuple}, nil); callErr == nil {
						if tup, ok := result.(*PyTuple); ok {
							for _, entry := range tup.Items {
								if cls, ok := entry.(*PyClass); ok {
									bases = append(bases, cls)
								}
							}
							continue
						}
						if lst, ok := result.(*PyList); ok {
							for _, entry := range lst.Items {
								if cls, ok := entry.(*PyClass); ok {
									bases = append(bases, cls)
								}
							}
							continue
						}
					}
				}
				// If __mro_entries__ not found or failed, skip non-class base
			}

			// If no bases specified, implicitly inherit from object (Python 3 behavior)
			objectClass := vm.builtins["object"].(*PyClass)
			if len(bases) == 0 {
				bases = []*PyClass{objectClass}
			}

			// Execute the class body to get the namespace and cells
			classDict, cells, err := vm.callClassBody(bodyFunc)
			if err != nil {
				return nil, fmt.Errorf("__build_class__: error executing class body: %w", err)
			}

			// Check for metaclass kwarg
			typeClass := vm.builtins["type"].(*PyClass)
			var metaclass *PyClass
			if mc, ok := kwargs["metaclass"]; ok {
				if mcClass, ok := mc.(*PyClass); ok {
					metaclass = mcClass
				}
			}

			// If no explicit metaclass, infer from bases
			if metaclass == nil {
				for _, base := range bases {
					if base.Metaclass != nil && base.Metaclass != typeClass {
						if metaclass == nil {
							metaclass = base.Metaclass
						} else {
							// Check if new metaclass is subclass of current
							for _, m := range base.Metaclass.Mro {
								if m == metaclass {
									metaclass = base.Metaclass
									break
								}
							}
						}
					}
				}
			}

			var class *PyClass

			if metaclass != nil && metaclass != typeClass {
				// Metaclass-based class creation: call metaclass.__new__ and __init__

				// Convert to Python values for metaclass methods
				basesItems := make([]Value, len(bases))
				for i, b := range bases {
					basesItems[i] = b
				}
				basesTuple := &PyTuple{Items: basesItems}
				nsDict := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
				for k, v := range classDict {
					nsDict.DictSet(&PyString{Value: k}, v, vm)
				}
				nameStr := &PyString{Value: className}

				// Call metaclass.__new__(mcs, name, bases, namespace) via MRO
				var newResult Value
				for _, cls := range metaclass.Mro {
					if newMethod, ok := cls.Dict["__new__"]; ok {
						newArgs := []Value{metaclass, nameStr, basesTuple, nsDict}
						switch nm := newMethod.(type) {
						case *PyFunction:
							newResult, err = vm.callFunction(nm, newArgs, kwargs)
						case *PyBuiltinFunc:
							newResult, err = nm.Fn(newArgs, kwargs)
						case *PyStaticMethod:
							newResult, err = vm.call(nm.Func, newArgs, kwargs)
						}
						if err != nil {
							return nil, err
						}
						break
					}
				}

				if newResult == nil {
					return nil, fmt.Errorf("TypeError: metaclass __new__ did not return a value")
				}

				if cls, ok := newResult.(*PyClass); ok {
					class = cls
					class.Metaclass = metaclass
					// Extract __slots__ if defined
					if class.Slots == nil {
						slots := extractSlots(class.Dict, bases)
						if slots != nil {
							class.Slots = slots
						}
					}

					// Call metaclass.__init__(cls, name, bases, namespace) via MRO
					for _, mroClass := range metaclass.Mro {
						if initMethod, ok := mroClass.Dict["__init__"]; ok {
							initArgs := []Value{class, nameStr, basesTuple, nsDict}
							switch im := initMethod.(type) {
							case *PyFunction:
								_, err = vm.callFunction(im, initArgs, kwargs)
							case *PyBuiltinFunc:
								_, err = im.Fn(initArgs, kwargs)
							}
							if err != nil {
								return nil, err
							}
							break
						}
					}
				} else {
					// If __new__ didn't return a *PyClass, just return it
					return newResult, nil
				}
			} else {
				// Standard class creation (no custom metaclass)
				slots := extractSlots(classDict, bases)
				class = &PyClass{
					Name:      className,
					Bases:     bases,
					Dict:      classDict,
					Metaclass: typeClass,
					Slots:     slots,
				}

				// Build MRO using C3 linearization for proper multiple inheritance
				mro, err := vm.ComputeC3MRO(class, bases)
				if err != nil {
					return nil, err
				}
				class.Mro = mro

				// Check if this class should use ABC abstract method checking
				for _, base := range bases {
					if base.IsABC {
						class.IsABC = true
						break
					}
				}

				// Collect abstract methods if this is an ABC class
				if class.IsABC {
					abstractMethods := make(map[string]bool)
					// Scan MRO (excluding current class) for abstract methods
					for _, cls := range mro[1:] {
						for name, val := range cls.Dict {
							if isAbstractValue(val) {
								abstractMethods[name] = true
							}
						}
					}
					// Scan current class: abstract methods add, concrete methods remove
					for name, val := range classDict {
						if isAbstractValue(val) {
							abstractMethods[name] = true
						} else {
							delete(abstractMethods, name)
						}
					}
					// Store as a PySet of strings for the instantiation guard
					if len(abstractMethods) > 0 {
						items := make([]Value, 0, len(abstractMethods))
						for name := range abstractMethods {
							items = append(items, &PyString{Value: name})
						}
						class.Dict["__abstractmethods__"] = &PyList{Items: items}
					}

					// Inject register() method for ABC classes (if not already defined)
					if _, hasRegister := class.Dict["register"]; !hasRegister {
						thisClass := class // capture for closure
						class.Dict["register"] = &PyBuiltinFunc{
							Name: "register",
							Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
								if len(args) != 1 {
									return nil, fmt.Errorf("TypeError: register() takes exactly 1 argument (%d given)", len(args))
								}
								subcls, ok := args[0].(*PyClass)
								if !ok {
									return nil, fmt.Errorf("TypeError: register() argument must be a class")
								}
								for _, existing := range thisClass.RegisteredSubclasses {
									if existing == subcls {
										return subcls, nil
									}
								}
								thisClass.RegisteredSubclasses = append(thisClass.RegisteredSubclasses, subcls)
								return subcls, nil
							},
						}
					}
				}
			}

			// Call __set_name__ on descriptors in the class dict
			if err := vm.callSetName(class); err != nil {
				return nil, err
			}

			// Call __init_subclass__ on parent classes
			if err := vm.callInitSubclass(class, kwargs); err != nil {
				return nil, err
			}

			// Populate the __class__ cell if present (for zero-argument super() support)
			// The __class__ cell is created by the compiler when methods use super()
			for i, cellName := range bodyFunc.Code.CellVars {
				if cellName == "__class__" && i < len(cells) && cells[i] != nil {
					cells[i].Value = class
					break
				}
			}

			return class, nil
		},
	}

	// property() creates a property descriptor
	vm.builtins["property"] = &PyBuiltinFunc{
		Name: "property",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			prop := &PyProperty{}

			// Handle positional args: property(fget=None, fset=None, fdel=None, doc=None)
			if len(args) > 0 && args[0] != None {
				prop.Fget = args[0]
			}
			if len(args) > 1 && args[1] != None {
				prop.Fset = args[1]
			}
			if len(args) > 2 && args[2] != None {
				prop.Fdel = args[2]
			}
			if len(args) > 3 {
				if s, ok := args[3].(*PyString); ok {
					prop.Doc = s.Value
				}
			}

			// Handle keyword args
			if fget, ok := kwargs["fget"]; ok && fget != None {
				prop.Fget = fget
			}
			if fset, ok := kwargs["fset"]; ok && fset != None {
				prop.Fset = fset
			}
			if fdel, ok := kwargs["fdel"]; ok && fdel != None {
				prop.Fdel = fdel
			}
			if doc, ok := kwargs["doc"]; ok {
				if s, ok := doc.(*PyString); ok {
					prop.Doc = s.Value
				}
			}

			return prop, nil
		},
	}

	// classmethod() wraps a function to bind the class as first argument
	vm.builtins["classmethod"] = &PyBuiltinFunc{
		Name: "classmethod",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("classmethod expected 1 argument, got %d", len(args))
			}
			return &PyClassMethod{Func: args[0]}, nil
		},
	}

	// staticmethod() wraps a function to prevent binding
	vm.builtins["staticmethod"] = &PyBuiltinFunc{
		Name: "staticmethod",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("staticmethod expected 1 argument, got %d", len(args))
			}
			return &PyStaticMethod{Func: args[0]}, nil
		},
	}

	// super() returns a proxy object for MRO-based method lookup
	vm.builtins["super"] = &PyBuiltinFunc{
		Name: "super",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			var thisClass *PyClass
			var instance Value

			if len(args) == 0 {
				// Zero-argument form: super()
				// Need to find __class__ from the enclosing scope and self from first arg
				// Look up the call stack to find the method context
				// We need to look at the caller's frame, not the current one (which is the builtin call)
				callerFrame := vm.frame
				if len(vm.frames) >= 1 {
					// The frames stack contains previous frames; vm.frame is current
					// For builtin calls, we want the calling frame
					callerFrame = vm.frame
				}
				if callerFrame != nil && callerFrame.Code != nil {
					// Try to get __class__ from closure cells
					for i, name := range callerFrame.Code.FreeVars {
						if name == "__class__" && i < len(callerFrame.Cells) {
							if cls, ok := callerFrame.Cells[i].Value.(*PyClass); ok {
								thisClass = cls
							}
						}
					}
					// Try to get self from first local variable (slot 0)
					if len(callerFrame.Code.VarNames) > 0 && len(callerFrame.Locals) > 0 {
						instance = callerFrame.Locals[0]
					}
				}
				if thisClass == nil {
					return nil, fmt.Errorf("super(): __class__ cell not found")
				}
				if instance == nil {
					return nil, fmt.Errorf("super(): self argument not found")
				}
			} else if len(args) == 2 {
				// Two-argument form: super(type, object-or-type)
				var ok bool
				thisClass, ok = args[0].(*PyClass)
				if !ok {
					return nil, fmt.Errorf("super() argument 1 must be type, not %s", vm.typeName(args[0]))
				}
				instance = args[1]
			} else if len(args) == 1 {
				// One-argument form: super(type) - unbound super
				var ok bool
				thisClass, ok = args[0].(*PyClass)
				if !ok {
					return nil, fmt.Errorf("super() argument 1 must be type, not %s", vm.typeName(args[0]))
				}
				instance = nil
			} else {
				return nil, fmt.Errorf("super() takes 0, 1, or 2 arguments (%d given)", len(args))
			}

			// Find the index of thisClass in the MRO of the instance's class
			var mro []*PyClass
			if inst, ok := instance.(*PyInstance); ok {
				mro = inst.Class.Mro
			} else if cls, ok := instance.(*PyClass); ok {
				// For class instances, check if thisClass is in the metaclass MRO
				// (used when super() is called inside a metaclass method)
				useMetaMro := false
				if cls.Metaclass != nil {
					for _, mc := range cls.Metaclass.Mro {
						if mc == thisClass {
							useMetaMro = true
							break
						}
					}
				}
				if useMetaMro {
					mro = cls.Metaclass.Mro
				} else {
					mro = cls.Mro
				}
			} else if instance != nil {
				return nil, fmt.Errorf("super(type, obj): obj must be an instance or subtype of type")
			}

			startIdx := 0
			if mro != nil {
				for i, cls := range mro {
					if cls == thisClass {
						startIdx = i + 1 // Start searching from the next class in MRO
						break
					}
				}
			}

			return &PySuper{
				ThisClass: thisClass,
				Instance:  instance,
				StartIdx:  startIdx,
			}, nil
		},
	}

	// object is the base class for all classes
	vm.builtins["object"] = &PyClass{
		Name:  "object",
		Bases: nil,
		Dict:  make(map[string]Value),
		Mro:   nil,
	}
	// Set object's MRO to just itself
	objectClass := vm.builtins["object"].(*PyClass)
	objectClass.Mro = []*PyClass{objectClass}

	// object.__getattribute__(self, name) - default attribute lookup (descriptor protocol)
	objectClass.Dict["__getattribute__"] = &PyBuiltinFunc{
		Name: "object.__getattribute__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("object.__getattribute__() takes exactly 2 arguments (%d given)", len(args))
			}
			inst, ok := args[0].(*PyInstance)
			if !ok {
				return nil, fmt.Errorf("descriptor '__getattribute__' requires a 'object' instance")
			}
			name, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("attribute name must be string, not '%s'", vm.typeName(args[1]))
			}
			return vm.defaultGetAttribute(inst, name.Value)
		},
	}

	// object.__setattr__(self, name, value) - direct instance dict assignment (bypasses user __setattr__)
	objectClass.Dict["__setattr__"] = &PyBuiltinFunc{
		Name: "object.__setattr__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("object.__setattr__() takes exactly 3 arguments (%d given)", len(args))
			}
			inst, ok := args[0].(*PyInstance)
			if !ok {
				return nil, fmt.Errorf("descriptor '__setattr__' requires a 'object' instance")
			}
			name, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("attribute name must be string, not '%s'", vm.typeName(args[1]))
			}
			// Respect property setters in MRO
			for _, cls := range inst.Class.Mro {
				if clsVal, ok := cls.Dict[name.Value]; ok {
					if prop, ok := clsVal.(*PyProperty); ok {
						if prop.Fset == nil {
							return nil, fmt.Errorf("property '%s' has no setter", name.Value)
						}
						_, err := vm.call(prop.Fset, []Value{inst, args[2]}, nil)
						if err != nil {
							return nil, err
						}
						return None, nil
					}
					break
				}
			}
			if inst.Slots != nil {
				if !isValidSlot(inst.Class, name.Value) {
					return nil, fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", inst.Class.Name, name.Value)
				}
				inst.Slots[name.Value] = args[2]
			} else {
				inst.Dict[name.Value] = args[2]
			}
			return None, nil
		},
	}

	// object.__delattr__(self, name) - direct instance dict deletion (bypasses user __delattr__)
	objectClass.Dict["__delattr__"] = &PyBuiltinFunc{
		Name: "object.__delattr__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("object.__delattr__() takes exactly 2 arguments (%d given)", len(args))
			}
			inst, ok := args[0].(*PyInstance)
			if !ok {
				return nil, fmt.Errorf("descriptor '__delattr__' requires a 'object' instance")
			}
			name, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("attribute name must be string, not '%s'", vm.typeName(args[1]))
			}
			// Respect property deleters and custom descriptor __delete__ in MRO
			for _, cls := range inst.Class.Mro {
				if clsVal, ok := cls.Dict[name.Value]; ok {
					if prop, ok := clsVal.(*PyProperty); ok {
						if prop.Fdel == nil {
							return nil, fmt.Errorf("property '%s' has no deleter", name.Value)
						}
						_, err := vm.call(prop.Fdel, []Value{inst}, nil)
						if err != nil {
							return nil, err
						}
						return None, nil
					}
					// Check for custom descriptor with __delete__
					if descInst, ok := clsVal.(*PyInstance); ok {
						if _, found, err := vm.callDunder(descInst, "__delete__", inst); found {
							if err != nil {
								return nil, err
							}
							return None, nil
						}
					}
					break
				}
			}
			if inst.Slots != nil {
				if _, exists := inst.Slots[name.Value]; !exists {
					return nil, fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", inst.Class.Name, name.Value)
				}
				delete(inst.Slots, name.Value)
			} else {
				if _, exists := inst.Dict[name.Value]; !exists {
					return nil, fmt.Errorf("AttributeError: '%s' object has no attribute '%s'", inst.Class.Name, name.Value)
				}
				delete(inst.Dict, name.Value)
			}
			return None, nil
		},
	}

	// object.__init_subclass__(cls, **kwargs) - default hook, does nothing
	objectClass.Dict["__sizeof__"] = &PyBuiltinFunc{
		Name: "object.__sizeof__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("descriptor '__sizeof__' requires an argument")
			}
			inst, ok := args[0].(*PyInstance)
			if !ok {
				return MakeInt(64), nil
			}
			// Base size for the instance struct
			var size int64 = 64
			if inst.Dict != nil {
				// 8 bytes per key-value pair estimate
				size += int64(len(inst.Dict) * 16)
			}
			if inst.Slots != nil {
				size += int64(len(inst.Slots) * 16)
			}
			return MakeInt(size), nil
		},
	}

	objectClass.Dict["__init_subclass__"] = &PyClassMethod{Func: &PyBuiltinFunc{
		Name: "__init_subclass__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return None, nil
		},
	}}

	// object.__new__(cls) - create a new instance of cls
	objectClass.Dict["__new__"] = &PyBuiltinFunc{
		Name: "object.__new__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("object.__new__(): not enough arguments")
			}
			cls, ok := args[0].(*PyClass)
			if !ok {
				return nil, fmt.Errorf("object.__new__(X): X is not a type object (%s)", vm.typeName(args[0]))
			}
			if cls.Slots != nil {
				return &PyInstance{
					Class: cls,
					Slots: make(map[string]Value),
				}, nil
			}
			return &PyInstance{
				Class: cls,
				Dict:  make(map[string]Value),
			}, nil
		},
	}

	// type is a proper PyClass (the metaclass of all classes)
	typeClass := &PyClass{
		Name:  "type",
		Bases: []*PyClass{objectClass},
		Dict:  make(map[string]Value),
	}
	typeClass.Mro = []*PyClass{typeClass, objectClass}
	vm.builtins["type"] = typeClass

	// type.__new__(mcs, name_or_obj, bases, namespace) - static method
	typeClass.Dict["__new__"] = &PyStaticMethod{Func: &PyBuiltinFunc{
		Name: "type.__new__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			// 2-arg form: type.__new__(type, x) -> type of x (called as type(x))
			if len(args) == 2 {
				switch v := args[1].(type) {
				case *PyInstance:
					return v.Class, nil
				case *PyClass:
					return typeClass, nil
				default:
					// Return a class with the type name
					typeName := vm.typeName(args[1])
					cls := &PyClass{Name: typeName}
					cls.Mro = []*PyClass{cls}
					return cls, nil
				}
			}
			// 4-arg form: type.__new__(mcs, name, bases, namespace)
			if len(args) == 4 {
				mcs, ok := args[0].(*PyClass)
				if !ok {
					return nil, fmt.Errorf("TypeError: type.__new__(X): X is not a type object (%s)", vm.typeName(args[0]))
				}
				nameStr, ok := args[1].(*PyString)
				if !ok {
					return nil, fmt.Errorf("TypeError: type.__new__() argument 1 must be str, not %s", vm.typeName(args[1]))
				}
				basesTuple, ok := args[2].(*PyTuple)
				if !ok {
					return nil, fmt.Errorf("TypeError: type.__new__() argument 2 must be tuple, not %s", vm.typeName(args[2]))
				}
				nsDict, ok := args[3].(*PyDict)
				if !ok {
					return nil, fmt.Errorf("TypeError: type.__new__() argument 3 must be dict, not %s", vm.typeName(args[3]))
				}

				// Convert bases tuple to []*PyClass
				var bases []*PyClass
				for _, b := range basesTuple.Items {
					if bc, ok := b.(*PyClass); ok {
						bases = append(bases, bc)
					}
				}
				if len(bases) == 0 {
					bases = []*PyClass{objectClass}
				}

				// Convert namespace dict to map[string]Value
				classDict := make(map[string]Value)
				for k, v := range nsDict.Items {
					if ks, ok := k.(*PyString); ok {
						classDict[ks.Value] = v
					}
				}

				slots := extractSlots(classDict, bases)
				cls := &PyClass{
					Name:      nameStr.Value,
					Bases:     bases,
					Dict:      classDict,
					Metaclass: mcs,
					Slots:     slots,
				}

				// Compute C3 MRO
				mro, err := vm.ComputeC3MRO(cls, bases)
				if err != nil {
					return nil, err
				}
				cls.Mro = mro

				// Handle ABC abstract method tracking
				if mcs.IsABC {
					cls.IsABC = true
				}
				if !cls.IsABC {
					for _, base := range bases {
						if base.IsABC {
							cls.IsABC = true
							break
						}
					}
				}
				if cls.IsABC {
					abstractMethods := make(map[string]bool)
					for _, mroClass := range mro[1:] {
						for name, val := range mroClass.Dict {
							if isAbstractValue(val) {
								abstractMethods[name] = true
							}
						}
					}
					for name, val := range classDict {
						if isAbstractValue(val) {
							abstractMethods[name] = true
						} else {
							delete(abstractMethods, name)
						}
					}
					if len(abstractMethods) > 0 {
						items := make([]Value, 0, len(abstractMethods))
						for name := range abstractMethods {
							items = append(items, &PyString{Value: name})
						}
						cls.Dict["__abstractmethods__"] = &PyList{Items: items}
					}
				}

				// Call __set_name__ on descriptors
				if err := vm.callSetName(cls); err != nil {
					return nil, err
				}

				return cls, nil
			}
			return nil, fmt.Errorf("type.__new__() takes 2 or 4 arguments (%d given)", len(args))
		},
	}}

	// type.__init__(cls, name, bases, namespace) - no-op
	typeClass.Dict["__init__"] = &PyBuiltinFunc{
		Name: "type.__init__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return None, nil
		},
	}

	// type.__call__(cls, *args, **kwargs) - default class instantiation
	typeClass.Dict["__call__"] = &PyBuiltinFunc{
		Name: "type.__call__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("TypeError: type.__call__() requires at least 1 argument")
			}
			cls, ok := args[0].(*PyClass)
			if !ok {
				return nil, fmt.Errorf("TypeError: descriptor '__call__' requires a 'type' object")
			}
			return vm.defaultClassCall(cls, args[1:], kwargs)
		},
	}

	// Initialize exception class hierarchy
	vm.initExceptionClasses()
}

// parseComplexString parses a string like "1+2j", "3j", "-1-2j", "1", etc. into a PyComplex
func parseComplexString(s string) (*PyComplex, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("ValueError: complex() arg is a malformed string")
	}

	// Remove surrounding parens if present
	if len(s) >= 2 && s[0] == '(' && s[len(s)-1] == ')' {
		s = s[1 : len(s)-1]
		s = strings.TrimSpace(s)
	}

	// Pure imaginary: ends with 'j' or 'J'
	if s[len(s)-1] == 'j' || s[len(s)-1] == 'J' {
		body := s[:len(s)-1]
		if body == "" || body == "+" {
			return MakeComplex(0, 1), nil
		}
		if body == "-" {
			return MakeComplex(0, -1), nil
		}

		// Find the split point between real and imaginary parts
		// Look for + or - that is NOT after e/E (scientific notation)
		splitIdx := -1
		for i := len(body) - 1; i > 0; i-- {
			if (body[i] == '+' || body[i] == '-') && body[i-1] != 'e' && body[i-1] != 'E' {
				splitIdx = i
				break
			}
		}

		if splitIdx == -1 {
			// Pure imaginary like "3j" or "-3j"
			imag, err := strconv.ParseFloat(body, 64)
			if err != nil {
				return nil, fmt.Errorf("ValueError: complex() arg is a malformed string")
			}
			return MakeComplex(0, imag), nil
		}

		// Has both real and imaginary: "1+2j" or "1-2j"
		realStr := body[:splitIdx]
		imagStr := body[splitIdx:]
		if imagStr == "+" {
			imagStr = "+1"
		} else if imagStr == "-" {
			imagStr = "-1"
		}

		real, err := strconv.ParseFloat(realStr, 64)
		if err != nil {
			return nil, fmt.Errorf("ValueError: complex() arg is a malformed string")
		}
		imag, err := strconv.ParseFloat(imagStr, 64)
		if err != nil {
			return nil, fmt.Errorf("ValueError: complex() arg is a malformed string")
		}
		return MakeComplex(real, imag), nil
	}

	// No 'j' — pure real
	real, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, fmt.Errorf("ValueError: complex() arg is a malformed string")
	}
	return MakeComplex(real, 0), nil
}

// extractSlots checks classDict for __slots__ and returns the slot names.
// Returns nil if __slots__ is not defined. Removes __slots__ from classDict.
func extractSlots(classDict map[string]Value, bases []*PyClass) []string {
	slotsVal, ok := classDict["__slots__"]
	if !ok {
		return nil
	}
	delete(classDict, "__slots__")

	var slotNames []string
	switch s := slotsVal.(type) {
	case *PyList:
		for _, item := range s.Items {
			if str, ok := item.(*PyString); ok {
				slotNames = append(slotNames, str.Value)
			}
		}
	case *PyTuple:
		for _, item := range s.Items {
			if str, ok := item.(*PyString); ok {
				slotNames = append(slotNames, str.Value)
			}
		}
	}

	// Collect slots from base classes that define __slots__
	for _, base := range bases {
		if base.Slots != nil {
			for _, s := range base.Slots {
				slotNames = append(slotNames, s)
			}
		}
	}

	if slotNames == nil {
		slotNames = []string{} // empty __slots__ = () should be non-nil empty slice
	}
	return slotNames
}

// isValidSlot checks whether name is in the class's allowed slots (including base class slots via MRO).
func isValidSlot(cls *PyClass, name string) bool {
	for _, mroClass := range cls.Mro {
		if mroClass.Slots != nil {
			for _, s := range mroClass.Slots {
				if s == name {
					return true
				}
			}
		}
	}
	return false
}

// initExceptionClasses sets up the exception class hierarchy
func (vm *VM) initExceptionClasses() {
	// BaseException is the root of all exceptions
	baseException := &PyClass{
		Name:  "BaseException",
		Bases: nil,
		Dict:  make(map[string]Value),
	}
	baseException.Mro = []*PyClass{baseException}
	// add_note(note) — appends a note string to __notes__
	baseException.Dict["add_note"] = &PyBuiltinFunc{
		Name: "BaseException.add_note",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("add_note() missing required argument: 'note'")
			}
			self, ok := args[0].(*PyInstance)
			if !ok {
				return nil, fmt.Errorf("add_note() requires an exception instance")
			}
			note, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("TypeError: note must be a str, not '%s'", vm.typeName(args[1]))
			}
			notes, exists := self.Dict["__notes__"]
			if !exists {
				notesList := &PyList{Items: []Value{note}}
				self.Dict["__notes__"] = notesList
			} else if notesList, ok := notes.(*PyList); ok {
				notesList.Items = append(notesList.Items, note)
			}
			return None, nil
		},
	}
	vm.builtins["BaseException"] = baseException

	// Exception inherits from BaseException (most exceptions derive from this)
	exception := &PyClass{
		Name:  "Exception",
		Bases: []*PyClass{baseException},
		Dict:  make(map[string]Value),
	}
	exception.Mro = []*PyClass{exception, baseException}
	vm.builtins["Exception"] = exception

	// Helper to create exception class inheriting from Exception
	makeExc := func(name string, parent *PyClass) *PyClass {
		cls := &PyClass{
			Name:  name,
			Bases: []*PyClass{parent},
			Dict:  make(map[string]Value),
		}
		// Build MRO by prepending self to parent's MRO
		cls.Mro = append([]*PyClass{cls}, parent.Mro...)
		vm.builtins[name] = cls
		return cls
	}

	// Standard exceptions inheriting from Exception
	makeExc("ValueError", exception)
	makeExc("TypeError", exception)
	makeExc("KeyError", exception)
	makeExc("IndexError", exception)
	attrError := makeExc("AttributeError", exception)
	makeExc("FrozenInstanceError", attrError)
	nameError := makeExc("NameError", exception)
	makeExc("UnboundLocalError", nameError)
	makeExc("RuntimeError", exception)
	makeExc("ZeroDivisionError", exception)
	makeExc("AssertionError", exception)
	makeExc("StopIteration", exception)
	makeExc("NotImplementedError", exception)
	makeExc("RecursionError", exception)
	makeExc("MemoryError", exception)

	// OSError and its subclasses
	osError := makeExc("OSError", exception)
	makeExc("FileNotFoundError", osError)
	makeExc("PermissionError", osError)
	makeExc("FileExistsError", osError)
	makeExc("IOError", osError) // IOError is an alias for OSError in Python 3

	// ImportError and its subclass
	importError := makeExc("ImportError", exception)
	makeExc("ModuleNotFoundError", importError)

	// LookupError (base for KeyError and IndexError - for compatibility)
	lookupError := makeExc("LookupError", exception)
	_ = lookupError // We already created KeyError and IndexError above

	// ArithmeticError (base for ZeroDivisionError - for compatibility)
	arithmeticError := makeExc("ArithmeticError", exception)
	_ = arithmeticError // We already created ZeroDivisionError above

	// GeneratorExit inherits from BaseException (not Exception)
	makeExc("GeneratorExit", baseException)

	// SystemExit inherits from BaseException (not Exception)
	makeExc("SystemExit", baseException)

	// KeyboardInterrupt inherits from BaseException (not Exception)
	makeExc("KeyboardInterrupt", baseException)

	// BaseExceptionGroup inherits from BaseException
	baseExcGroup := &PyClass{
		Name:  "BaseExceptionGroup",
		Bases: []*PyClass{baseException},
		Dict:  make(map[string]Value),
	}
	baseExcGroup.Mro = []*PyClass{baseExcGroup, baseException}
	vm.builtins["BaseExceptionGroup"] = baseExcGroup

	// ExceptionGroup inherits from Exception and BaseExceptionGroup
	excGroup := &PyClass{
		Name:  "ExceptionGroup",
		Bases: []*PyClass{exception, baseExcGroup},
		Dict:  make(map[string]Value),
	}
	excGroup.Mro = []*PyClass{excGroup, exception, baseExcGroup, baseException}
	vm.builtins["ExceptionGroup"] = excGroup

	// Shared __init__ for both exception group classes
	egInit := &PyBuiltinFunc{Name: "ExceptionGroup.__init__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("TypeError: ExceptionGroup.__init__() requires at least 2 arguments (message, exceptions)")
		}
		self := args[0]
		inst, ok := self.(*PyInstance)
		if !ok {
			return nil, fmt.Errorf("TypeError: ExceptionGroup.__init__() expected instance")
		}
		msgVal := args[1]
		msgStr, ok := msgVal.(*PyString)
		if !ok {
			return nil, fmt.Errorf("TypeError: ExceptionGroup message must be a string")
		}
		excsVal := args[2]
		var excItems []Value
		switch ev := excsVal.(type) {
		case *PyList:
			excItems = ev.Items
		case *PyTuple:
			excItems = ev.Items
		default:
			return nil, fmt.Errorf("TypeError: ExceptionGroup exceptions must be a list or tuple")
		}
		if len(excItems) == 0 {
			return nil, fmt.Errorf("ValueError: ExceptionGroup exceptions must be non-empty")
		}
		// Convert to []*PyException for VM use and validate
		pyExcs := make([]*PyException, len(excItems))
		tupleItems := make([]Value, len(excItems))
		for i, item := range excItems {
			switch e := item.(type) {
			case *PyException:
				pyExcs[i] = e
				tupleItems[i] = e
			case *PyInstance:
				if vm.isExceptionClass(e.Class) {
					pyExcs[i] = vm.createException(e, nil)
					tupleItems[i] = e
				} else {
					return nil, fmt.Errorf("TypeError: exceptions must be instances of BaseException")
				}
			default:
				return nil, fmt.Errorf("TypeError: exceptions must be instances of BaseException")
			}
		}
		inst.Dict["message"] = msgStr
		inst.Dict["exceptions"] = &PyTuple{Items: tupleItems}
		inst.Dict["args"] = &PyTuple{Items: []Value{msgStr}}
		inst.Dict["__eg_exceptions__"] = &pyExceptionList{items: pyExcs}
		return None, nil
	}}

	// __str__ for exception groups
	egStr := &PyBuiltinFunc{Name: "ExceptionGroup.__str__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 1 {
			return &PyString{Value: "ExceptionGroup()"}, nil
		}
		inst, ok := args[0].(*PyInstance)
		if !ok {
			return &PyString{Value: "ExceptionGroup()"}, nil
		}
		msg := "ExceptionGroup"
		if m, ok := inst.Dict["message"].(*PyString); ok {
			msg = m.Value
		}
		excsTuple, _ := inst.Dict["exceptions"].(*PyTuple)
		count := 0
		if excsTuple != nil {
			count = len(excsTuple.Items)
		}
		sub := "sub-exception"
		if count != 1 {
			sub = "sub-exceptions"
		}
		return &PyString{Value: fmt.Sprintf("%s (%d %s)", msg, count, sub)}, nil
	}}

	// __repr__ for exception groups
	egRepr := &PyBuiltinFunc{Name: "ExceptionGroup.__repr__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 1 {
			return &PyString{Value: "ExceptionGroup()"}, nil
		}
		inst, ok := args[0].(*PyInstance)
		if !ok {
			return &PyString{Value: "ExceptionGroup()"}, nil
		}
		msg := ""
		if m, ok := inst.Dict["message"].(*PyString); ok {
			msg = m.Value
		}
		excsTuple, _ := inst.Dict["exceptions"].(*PyTuple)
		className := "ExceptionGroup"
		if inst.Class != nil {
			className = inst.Class.Name
		}
		excReprs := "[]"
		if excsTuple != nil {
			parts := make([]string, len(excsTuple.Items))
			for i, e := range excsTuple.Items {
				parts[i] = vm.repr(e)
			}
			excReprs = "[" + strings.Join(parts, ", ") + "]"
		}
		return &PyString{Value: fmt.Sprintf("%s('%s', %s)", className, msg, excReprs)}, nil
	}}

	// subgroup(condition) — filter exceptions matching type
	egSubgroup := &PyBuiltinFunc{Name: "ExceptionGroup.subgroup", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("TypeError: subgroup() requires 1 argument")
		}
		inst, ok := args[0].(*PyInstance)
		if !ok {
			return None, nil
		}
		condition := args[1]
		egExcs := vm.getEGExceptions(inst)
		if egExcs == nil {
			return None, nil
		}
		var matched []*PyException
		for _, exc := range egExcs {
			if vm.exceptionMatches(exc, condition) {
				matched = append(matched, exc)
			}
		}
		if len(matched) == 0 {
			return None, nil
		}
		msg := ""
		if m, ok := inst.Dict["message"].(*PyString); ok {
			msg = m.Value
		}
		return vm.buildExceptionGroup(msg, matched, vm.isBaseExceptionGroup(inst.Class))
	}}

	// split(condition) — return (matched, rest) tuple
	egSplit := &PyBuiltinFunc{Name: "ExceptionGroup.split", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("TypeError: split() requires 1 argument")
		}
		inst, ok := args[0].(*PyInstance)
		if !ok {
			return &PyTuple{Items: []Value{None, None}}, nil
		}
		condition := args[1]
		egExcs := vm.getEGExceptions(inst)
		if egExcs == nil {
			return &PyTuple{Items: []Value{None, None}}, nil
		}
		var matched, rest []*PyException
		for _, exc := range egExcs {
			if vm.exceptionMatches(exc, condition) {
				matched = append(matched, exc)
			} else {
				rest = append(rest, exc)
			}
		}
		msg := ""
		if m, ok := inst.Dict["message"].(*PyString); ok {
			msg = m.Value
		}
		isBase := vm.isBaseExceptionGroup(inst.Class)
		var matchGroup, restGroup Value
		if len(matched) > 0 {
			matchGroup, _ = vm.buildExceptionGroup(msg, matched, isBase)
		} else {
			matchGroup = None
		}
		if len(rest) > 0 {
			restGroup, _ = vm.buildExceptionGroup(msg, rest, isBase)
		} else {
			restGroup = None
		}
		return &PyTuple{Items: []Value{matchGroup, restGroup}}, nil
	}}

	// derive(excs) — create new group of same class
	egDerive := &PyBuiltinFunc{Name: "ExceptionGroup.derive", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("TypeError: derive() requires 1 argument")
		}
		inst, ok := args[0].(*PyInstance)
		if !ok {
			return None, nil
		}
		excsVal := args[1]
		var excItems []Value
		switch ev := excsVal.(type) {
		case *PyList:
			excItems = ev.Items
		case *PyTuple:
			excItems = ev.Items
		default:
			return nil, fmt.Errorf("TypeError: derive() argument must be a list or tuple")
		}
		pyExcs := make([]*PyException, len(excItems))
		for i, item := range excItems {
			switch e := item.(type) {
			case *PyException:
				pyExcs[i] = e
			default:
				pyExcs[i] = vm.createException(e, nil)
			}
		}
		msg := ""
		if m, ok := inst.Dict["message"].(*PyString); ok {
			msg = m.Value
		}
		return vm.buildExceptionGroup(msg, pyExcs, vm.isBaseExceptionGroup(inst.Class))
	}}

	// Register methods on both classes
	for _, cls := range []*PyClass{baseExcGroup, excGroup} {
		cls.Dict["__init__"] = egInit
		cls.Dict["__str__"] = egStr
		cls.Dict["__repr__"] = egRepr
		cls.Dict["subgroup"] = egSubgroup
		cls.Dict["split"] = egSplit
		cls.Dict["derive"] = egDerive
	}
}

// ComputeC3MRO computes the Method Resolution Order using C3 linearization algorithm.
// This properly handles multiple inheritance and detects inconsistent hierarchies.
func (vm *VM) ComputeC3MRO(class *PyClass, bases []*PyClass) ([]*PyClass, error) {
	// Base case: no bases
	if len(bases) == 0 {
		return []*PyClass{class}, nil
	}

	// Collect linearizations of all bases plus the list of bases itself
	// We need to copy slices to avoid modifying the originals
	var toMerge [][]*PyClass
	for _, base := range bases {
		// Copy the base's MRO
		baseMRO := make([]*PyClass, len(base.Mro))
		copy(baseMRO, base.Mro)
		toMerge = append(toMerge, baseMRO)
	}
	// Add the list of direct bases
	basesCopy := make([]*PyClass, len(bases))
	copy(basesCopy, bases)
	toMerge = append(toMerge, basesCopy)

	// Start with the class itself
	result := []*PyClass{class}

	// Merge until all lists are empty
	for {
		// Remove empty lists
		var nonEmpty [][]*PyClass
		for _, list := range toMerge {
			if len(list) > 0 {
				nonEmpty = append(nonEmpty, list)
			}
		}
		toMerge = nonEmpty

		if len(toMerge) == 0 {
			break
		}

		// Find a good head: a class that is not in the tail of any list
		var candidate *PyClass
		for _, list := range toMerge {
			head := list[0]
			inTail := false
			for _, other := range toMerge {
				// Check if head appears in the tail (positions 1+) of other
				for i := 1; i < len(other); i++ {
					if other[i] == head {
						inTail = true
						break
					}
				}
				if inTail {
					break
				}
			}
			if !inTail {
				candidate = head
				break
			}
		}

		if candidate == nil {
			// No valid candidate found - inconsistent hierarchy
			msg := fmt.Sprintf("Cannot create a consistent method resolution order (MRO) for bases %s",
				vm.formatBases(bases))
			return nil, &PyException{
				ExcType:  vm.builtins["TypeError"].(*PyClass),
				Args:     &PyTuple{Items: []Value{&PyString{Value: msg}}},
				Message:  "TypeError: " + msg,
				TypeName: "TypeError",
			}
		}

		// Add candidate to result
		result = append(result, candidate)

		// Remove candidate from the head of all lists where it appears
		for i := range toMerge {
			if len(toMerge[i]) > 0 && toMerge[i][0] == candidate {
				toMerge[i] = toMerge[i][1:]
			}
		}
	}

	return result, nil
}

// sortedNameList converts a set of names to a sorted PyList of PyStrings.
func (vm *VM) sortedNameList(names map[string]bool) *PyList {
	sorted := make([]string, 0, len(names))
	for k := range names {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	items := make([]Value, len(sorted))
	for i, s := range sorted {
		items[i] = &PyString{Value: s}
	}
	return &PyList{Items: items}
}

// isAbstractValue checks if a value is marked as abstract
func isAbstractValue(v Value) bool {
	switch val := v.(type) {
	case *PyFunction:
		return val.IsAbstract
	case *PyProperty:
		if fn, ok := val.Fget.(*PyFunction); ok {
			return fn.IsAbstract
		}
	case *PyClassMethod:
		if fn, ok := val.Func.(*PyFunction); ok {
			return fn.IsAbstract
		}
	case *PyStaticMethod:
		if fn, ok := val.Func.(*PyFunction); ok {
			return fn.IsAbstract
		}
	}
	return false
}

// formatBases formats a list of base classes for error messages
func (vm *VM) formatBases(bases []*PyClass) string {
	if len(bases) == 0 {
		return ""
	}
	names := make([]string, len(bases))
	for i, b := range bases {
		names[i] = b.Name
	}
	result := names[0]
	for i := 1; i < len(names); i++ {
		result += ", " + names[i]
	}
	return result
}

// callInitSubclass calls __init_subclass__ on the parent class after a new class is created.
// It walks the MRO starting from index 1 (skipping the new class itself) to find the hook.
// callSetName iterates through the class dict and calls __set_name__(owner, name)
// on any descriptor that defines it. Called during class creation.
func (vm *VM) callSetName(class *PyClass) error {
	for name, val := range class.Dict {
		if inst, ok := val.(*PyInstance); ok {
			_, found, err := vm.callDunder(inst, "__set_name__", class, &PyString{Value: name})
			if err != nil {
				return fmt.Errorf("RuntimeError: __set_name__ of '%s' descriptor '%s' raised: %w",
					inst.Class.Name, name, err)
			}
			_ = found
		}
	}
	return nil
}

func (vm *VM) callInitSubclass(class *PyClass, kwargs map[string]Value) error {
	// Filter out "metaclass" from kwargs
	var filteredKwargs map[string]Value
	if len(kwargs) > 0 {
		filteredKwargs = make(map[string]Value, len(kwargs))
		for k, v := range kwargs {
			if k != "metaclass" {
				filteredKwargs[k] = v
			}
		}
	}

	// Walk MRO starting from index 1 (skip the new class itself)
	for i := 1; i < len(class.Mro); i++ {
		if method, ok := class.Mro[i].Dict["__init_subclass__"]; ok {
			args := []Value{class}
			var err error
			switch m := method.(type) {
			case *PyClassMethod:
				_, err = vm.call(m.Func, args, filteredKwargs)
			case *PyFunction:
				_, err = vm.callFunction(m, args, filteredKwargs)
			case *PyBuiltinFunc:
				_, err = m.Fn(args, filteredKwargs)
			}
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}
