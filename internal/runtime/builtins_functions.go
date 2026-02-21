package runtime

import (
	"fmt"
	"math"
	"sort"
)

// initBuiltinsFunctions registers builtin functions: print, range, enumerate, zip,
// map, filter, reversed, sorted, next, iter, all, any, abs, hash, min, max, sum,
// pow, divmod, round, ord, chr, hex, oct, bin, isinstance, issubclass, callable,
// hasattr, dir, getattr, setattr, delattr, repr, ascii, format, input.
func (vm *VM) initBuiltinsFunctions() {
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
}
