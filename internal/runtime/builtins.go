package runtime

import (
	"fmt"
	"math"
	"sort"
	"unicode/utf8"
)

func (vm *VM) initBuiltins() {
	vm.builtins["print"] = &PyBuiltinFunc{
		Name: "print",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			for i, arg := range args {
				if i > 0 {
					fmt.Print(" ")
				}
				fmt.Print(vm.str(arg))
			}
			fmt.Println()
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

			i, err := vm.tryToInt(args[0])
			if err != nil {
				return nil, err
			}
			return MakeInt(i), nil
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

	typeClassCache := make(map[string]*PyClass)
	getTypeClass := func(name string) *PyClass {
		if cls, ok := typeClassCache[name]; ok {
			return cls
		}
		cls := &PyClass{Name: name}
		cls.Mro = []*PyClass{cls}
		typeClassCache[name] = cls
		return cls
	}
	vm.builtins["type"] = &PyBuiltinFunc{
		Name: "type",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 1 {
				switch v := args[0].(type) {
				case *PyInstance:
					return v.Class, nil
				case *PyClass:
					return getTypeClass("type"), nil
				default:
					return getTypeClass(vm.typeName(args[0])), nil
				}
			}
			if len(args) == 3 {
				// 3-arg form: type(name, bases, dict) - metaclass creation
				nameStr, ok := args[0].(*PyString)
				if !ok {
					return nil, fmt.Errorf("TypeError: type() argument 1 must be str, not %s", vm.typeName(args[0]))
				}
				cls := &PyClass{Name: nameStr.Value, Dict: make(map[string]Value)}
				cls.Mro = []*PyClass{cls}
				return cls, nil
			}
			return nil, fmt.Errorf("type() takes 1 or 3 arguments")
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
					return vm.isInstanceOf(o, cls)
				case *PyException:
					if vm.isExceptionClass(cls) {
						return vm.exceptionMatches(o, cls)
					}
				}
				// For built-in types, check by name
				return checkTypeName(cls.Name)
			}

			// Helper to check a single type specification
			checkType := func(typeSpec Value) bool {
				switch t := typeSpec.(type) {
				case *PyClass:
					return checkClass(t)
				case *PyBuiltinFunc:
					// Built-in type constructors (int, str, float, etc.)
					return checkTypeName(t.Name)
				}
				return false
			}

			// classInfo can be a single class/type or a tuple of classes/types
			switch ci := classInfo.(type) {
			case *PyClass:
				if checkClass(ci) {
					return True, nil
				}
			case *PyBuiltinFunc:
				// Built-in type constructor (int, str, float, etc.)
				if checkTypeName(ci.Name) {
					return True, nil
				}
			case *PyTuple:
				for _, item := range ci.Items {
					if checkType(item) {
						return True, nil
					}
				}
			default:
				return nil, fmt.Errorf("isinstance() arg 2 must be a type or tuple of types")
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
			if len(args) == 1 {
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				args = items
			}
			minVal := args[0]
			for _, v := range args[1:] {
				if vm.compare(v, minVal) < 0 {
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
			if len(args) == 1 {
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				args = items
			}
			maxVal := args[0]
			for _, v := range args[1:] {
				if vm.compare(v, maxVal) > 0 {
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
			// Convert all args to lists
			lists := make([][]Value, len(args))
			minLen := -1
			for i, arg := range args {
				items, err := vm.toList(arg)
				if err != nil {
					return nil, err
				}
				lists[i] = items
				if minLen == -1 || len(items) < minLen {
					minLen = len(items)
				}
			}
			// Build result tuples
			result := make([]Value, minLen)
			for i := 0; i < minLen; i++ {
				tuple := make([]Value, len(lists))
				for j, list := range lists {
					tuple[j] = list[i]
				}
				result[i] = &PyTuple{Items: tuple}
			}
			return &PyIterator{Items: result, Index: 0}, nil
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
				cmp := vm.compare(a, b)
				if reverse {
					return cmp > 0
				}
				return cmp < 0
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
				return False, nil
			}
			return True, nil
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
			// Handle deletion based on object type
			switch o := args[0].(type) {
			case *PyInstance:
				if _, exists := o.Dict[name.Value]; exists {
					delete(o.Dict, name.Value)
					return None, nil
				}
				return nil, fmt.Errorf("'%s' object has no attribute '%s'", o.Class.Name, name.Value)
			case *PyModule:
				if _, exists := o.Dict[name.Value]; exists {
					delete(o.Dict, name.Value)
					return None, nil
				}
				return nil, fmt.Errorf("module '%s' has no attribute '%s'", o.Name, name.Value)
			case *PyDict:
				// Allow delattr on dict for dynamic attribute-style access
				key := &PyString{Value: name.Value}
				if o.DictDelete(key, vm) {
					return None, nil
				}
				return nil, fmt.Errorf("'dict' object has no attribute '%s'", name.Value)
			default:
				return nil, fmt.Errorf("'%s' object has no attribute '%s'", vm.typeName(args[0]), name.Value)
			}
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
					return nil, fmt.Errorf("integer division or modulo by zero")
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
				return nil, fmt.Errorf("float division by zero")
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
			n := vm.toInt(args[0])
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
			n := vm.toInt(args[0])
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
			n := vm.toInt(args[0])
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

	vm.builtins["None"] = None
	vm.builtins["True"] = True
	vm.builtins["False"] = False

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

			// Remaining args are base classes
			var bases []*PyClass
			for i := 2; i < len(args); i++ {
				if base, ok := args[i].(*PyClass); ok {
					bases = append(bases, base)
				}
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

			// Create the class
			class := &PyClass{
				Name:  className,
				Bases: bases,
				Dict:  classDict,
			}

			// Build MRO using C3 linearization for proper multiple inheritance
			mro, err := vm.computeC3MRO(class, bases)
			if err != nil {
				return nil, err
			}
			class.Mro = mro

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
				mro = cls.Mro
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
	vm.builtins["object"].(*PyClass).Mro = []*PyClass{vm.builtins["object"].(*PyClass)}

	// Initialize exception class hierarchy
	vm.initExceptionClasses()
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
	makeExc("AttributeError", exception)
	makeExc("NameError", exception)
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
}

// computeC3MRO computes the Method Resolution Order using C3 linearization algorithm.
// This properly handles multiple inheritance and detects inconsistent hierarchies.
func (vm *VM) computeC3MRO(class *PyClass, bases []*PyClass) ([]*PyClass, error) {
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
