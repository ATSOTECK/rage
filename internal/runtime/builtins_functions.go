package runtime

import (
	"fmt"
	"math"
	"math/big"
	"strings"
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
			flush := false
			if v, ok := kwargs["flush"]; ok {
				if b, ok := v.(*PyBool); ok && b.Value {
					flush = true
				}
			}

			var fileObj Value
			if v, ok := kwargs["file"]; ok && !IsNone(v) {
				fileObj = v
			}
			// Honor sys.stdout if a contextmanager (e.g. contextlib.redirect_stdout)
			// has swapped it. Skip if the user passed an explicit file= kwarg.
			if fileObj == nil {
				if mod, ok := vm.GetModule("sys"); ok {
					if v, ok := mod.Dict["stdout"]; ok && !IsNone(v) {
						fileObj = v
					}
				}
			}

			if fileObj != nil {
				// Write to file object via its write() method
				writeMethod, err := vm.GetAttr(fileObj, "write")
				if err != nil {
					return nil, fmt.Errorf("TypeError: argument 'file' must have a 'write' attribute")
				}
				var buf strings.Builder
				for i, arg := range args {
					if i > 0 {
						buf.WriteString(sep)
					}
					s, err := vm.strValue(arg)
					if err != nil {
						return nil, err
					}
					buf.WriteString(s)
				}
				buf.WriteString(end)
				_, err = vm.Call(writeMethod, []Value{NewString(buf.String())}, nil)
				if err != nil {
					return nil, err
				}
				if flush {
					if flushMethod, err := vm.GetAttr(fileObj, "flush"); err == nil {
						_, _ = vm.Call(flushMethod, nil, nil)
					}
				}
			} else {
				// Write to stdout (default). Convert every argument first so
				// an exception from __str__ leaves no partial output.
				var buf strings.Builder
				for i, arg := range args {
					if i > 0 {
						buf.WriteString(sep)
					}
					s, err := vm.strValue(arg)
					if err != nil {
						return nil, err
					}
					buf.WriteString(s)
				}
				buf.WriteString(end)
				fmt.Print(buf.String())
			}
			return None, nil
		},
	}

	vm.builtins["range"] = &PyBuiltinFunc{
		Name: "range",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			var start int64
			var stop int64
			var step int64 = 1
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
			s, err := vm.reprValue(args[0])
			if err != nil {
				return nil, err
			}
			return &PyString{Value: s}, nil
		},
	}

	vm.builtins["ascii"] = &PyBuiltinFunc{
		Name: "ascii",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("ascii() takes exactly 1 argument (%d given)", len(args))
			}
			s, err := vm.asciiValue(args[0])
			if err != nil {
				return nil, err
			}
			return &PyString{Value: s}, nil
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
				case *UnionType:
					for _, item := range t.Args {
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
				return intAbs(v), nil
			case *PyBool:
				return boolToInt(v), nil
			case *PyFloat:
				return &PyFloat{Value: math.Abs(v.Value)}, nil
			case *PyComplex:
				r := math.Hypot(v.Real, v.Imag)
				if math.IsInf(r, 0) && !math.IsInf(v.Real, 0) && !math.IsInf(v.Imag, 0) {
					return nil, fmt.Errorf("OverflowError: absolute value too large")
				}
				return &PyFloat{Value: r}, nil
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
			return vm.minMax("min", args, kwargs)
		},
	}

	vm.builtins["max"] = &PyBuiltinFunc{
		Name: "max",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return vm.minMax("max", args, kwargs)
		},
	}

	vm.builtins["sum"] = &PyBuiltinFunc{
		Name: "sum",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 || len(args) > 2 {
				return nil, fmt.Errorf("TypeError: sum() takes at most 2 arguments (%d given)", len(args))
			}
			var start Value = MakeInt(0)
			if len(args) > 1 {
				start = args[1]
			} else if s, ok := kwargs["start"]; ok {
				start = s
			}
			switch start.(type) {
			case *PyString:
				return nil, fmt.Errorf("TypeError: sum() can't sum strings [use ''.join(seq) instead]")
			case *PyBytes:
				return nil, fmt.Errorf("TypeError: sum() can't sum bytes [use b''.join(seq) instead]")
			}
			iter, err := vm.getIter(args[0])
			if err != nil {
				return nil, err
			}
			return vm.sumValues(start, func() (Value, bool, error) {
				v, done, err := vm.iterNext(iter)
				return v, done, err
			})
		},
	}

	vm.builtins["input"] = &PyBuiltinFunc{
		Name: "input",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) > 0 {
				fmt.Print(vm.str(args[0]))
			}
			var line string
			if _, err := fmt.Scanln(&line); err != nil {
				return &PyString{Value: ""}, nil
			}
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
			strict := false
			for k, v := range kwargs {
				if k != "strict" {
					return nil, fmt.Errorf("TypeError: zip() got an unexpected keyword argument '%s'", k)
				}
				strict = vm.truthy(v)
			}
			if len(args) == 0 {
				return &PyIterator{Items: []Value{}, Index: 0}, nil
			}
			// Get iterators for all args
			iters := make([]Value, len(args))
			for i, arg := range args {
				it, err := vm.getIter(arg)
				if err != nil {
					return nil, fmt.Errorf("TypeError: 'zip' argument #%d must support iteration", i+1)
				}
				iters[i] = it
			}
			// argsRange describes the arguments before index j, e.g. "argument 1"
			// or "arguments 1-2", for strict-mode errors
			argsRange := func(j int) string {
				if j == 1 {
					return "argument 1"
				}
				return fmt.Sprintf("arguments 1-%d", j)
			}
			// Collect tuples until any iterator is exhausted
			var result []Value
			for {
				tuple := make([]Value, len(iters))
				for j, it := range iters {
					val, done, err := vm.iterNext(it)
					if err != nil {
						return nil, err
					}
					if done {
						if strict {
							if j > 0 {
								return nil, fmt.Errorf("ValueError: zip() argument %d is shorter than %s", j+1, argsRange(j))
							}
							// The first iterator ended: all others must end too
							for k := 1; k < len(iters); k++ {
								_, kDone, err := vm.iterNext(iters[k])
								if err != nil {
									return nil, err
								}
								if !kDone {
									return nil, fmt.Errorf("ValueError: zip() argument %d is longer than %s", k+1, argsRange(k))
								}
							}
						}
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
				var val Value
				var err error
				if len(lists) == 1 {
					val, err = vm.call1(fn, lists[0][i])
				} else {
					fnArgs := make([]Value, len(lists))
					for j, list := range lists {
						fnArgs[j] = list[i]
					}
					val, err = vm.call(fn, fnArgs, nil)
				}
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
					val, err := vm.call1(fn, item)
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
				return nil, fmt.Errorf("TypeError: sorted expected 1 argument, got %d", len(args))
			}
			keyFn, reverse, err := vm.sortKwargs("sorted", kwargs)
			if err != nil {
				return nil, err
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			// Make a copy to avoid modifying the original
			result := make([]Value, len(items))
			copy(result, items)
			if err := vm.sortValues(result, keyFn, reverse); err != nil {
				return nil, err
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
			args = append([]Value(nil), args...)
			for i, name := range []string{"base", "exp", "mod"} {
				if v, ok := kwargs[name]; ok {
					for len(args) <= i {
						args = append(args, nil)
					}
					args[i] = v
				}
			}
			if len(args) < 2 || len(args) > 3 || args[0] == nil || args[1] == nil {
				return nil, fmt.Errorf("TypeError: pow() missing required argument")
			}
			if len(args) == 3 && args[2] != nil && args[2] != None {
				return vm.modPow(args[0], args[1], args[2])
			}
			return vm.binaryOp(OpBinaryPower, args[0], args[1])
		},
	}

	// divmod(a, b) - returns (quotient, remainder)
	vm.builtins["divmod"] = &PyBuiltinFunc{
		Name: "divmod",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("TypeError: divmod expected 2 arguments, got %d", len(args))
			}
			return vm.divmod(args[0], args[1])
		},
	}

	// hex(x), oct(x), bin(x) - integer to string in base 16, 8, 2
	for _, conv := range []struct {
		name, prefix string
		base         int
	}{{"hex", "0x", 16}, {"oct", "0o", 8}, {"bin", "0b", 2}} {
		conv := conv
		vm.builtins[conv.name] = &PyBuiltinFunc{
			Name: conv.name,
			Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("TypeError: %s() takes exactly one argument (%d given)", conv.name, len(args))
				}
				n, err := vm.indexInt(args[0])
				if err != nil {
					return nil, err
				}
				x := n.BigIntValue()
				if x.Sign() < 0 {
					return &PyString{Value: "-" + conv.prefix + new(big.Int).Neg(x).Text(conv.base)}, nil
				}
				return &PyString{Value: conv.prefix + x.Text(conv.base)}, nil
			},
		}
	}

	// round(number[, ndigits]) - round to given precision
	vm.builtins["round"] = &PyBuiltinFunc{
		Name: "round",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if nd, ok := kwargs["ndigits"]; ok && len(args) == 1 {
				args = []Value{args[0], nd}
			}
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("TypeError: round() takes 1 or 2 arguments (%d given)", len(args))
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
			var ndigits Value
			if len(args) == 2 && args[1] != None {
				ndigits = args[1]
			}
			return vm.round(args[0], ndigits)
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
			case lazyIterator:
				val, ok, err := it.advance(vm)
				if err != nil {
					return nil, err
				}
				if !ok {
					if hasDefault {
						return args[1], nil
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
				case *UnionType:
					for _, item := range target.Args {
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

// sortValues sorts items in place by Python "<" ordering. The sort is stable.
// When keyFn is non-nil it is called exactly once per item (as in CPython) and
// the resulting keys are compared instead of the items themselves. When
// reverse is true the order is descending, with equal elements keeping their
// original relative order.
func (vm *VM) sortValues(items []Value, keyFn Value, reverse bool) error {
	n := len(items)
	keys := items
	if keyFn != nil {
		keys = make([]Value, n)
		for i, item := range items {
			k, err := vm.call1(keyFn, item)
			if err != nil {
				return err
			}
			keys[i] = k
		}
	}
	if n < 2 {
		return nil
	}

	var sortErr error
	less := func(i, j int) bool {
		if sortErr != nil {
			return false
		}
		// For reverse sort, compare b < a instead of a < b to maintain stability.
		// Using !less would break stability because equal elements (less=false)
		// would return true, swapping their order.
		a, b := keys[i], keys[j]
		if reverse {
			a, b = b, a
		}
		result, err := vm.compareChecked(OpCompareLt, a, b)
		if err != nil {
			sortErr = err
			return false
		}
		return vm.truthy(result)
	}

	perm := make([]int, n)
	for i := range perm {
		perm[i] = i
	}
	mergeSortIndices(perm, make([]int, n/2+1), less)
	if sortErr != nil {
		return sortErr
	}

	sorted := make([]Value, n)
	for i, p := range perm {
		sorted[i] = items[p]
	}
	copy(items, sorted)
	return nil
}

// mergeSortIndices stably sorts a using less, which compares the elements the
// indices refer to. buf must have room for at least half of a. Merge sort keeps
// the number of comparisons (each potentially a Python-level __lt__ call) close
// to n*log2(n), and is linear for input that is already sorted.
func mergeSortIndices(a, buf []int, less func(i, j int) bool) {
	n := len(a)
	if n <= 8 {
		// Binary insertion sort; equal elements are inserted after existing
		// ones, keeping the sort stable.
		for i := 1; i < n; i++ {
			x := a[i]
			lo, hi := 0, i
			for lo < hi {
				mid := int(uint(lo+hi) >> 1)
				if less(x, a[mid]) {
					hi = mid
				} else {
					lo = mid + 1
				}
			}
			copy(a[lo+1:i+1], a[lo:i])
			a[lo] = x
		}
		return
	}
	mid := n / 2
	mergeSortIndices(a[:mid], buf, less)
	mergeSortIndices(a[mid:], buf, less)
	if !less(a[mid], a[mid-1]) {
		return // Halves are already in order
	}
	left := buf[:mid]
	copy(left, a[:mid])
	i, j, k := 0, mid, 0
	for i < mid && j < n {
		if less(a[j], left[i]) {
			a[k] = a[j]
			j++
		} else {
			a[k] = left[i]
			i++
		}
		k++
	}
	copy(a[k:], left[i:])
}
