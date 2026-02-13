package runtime

import "fmt"

// Function calls

func (vm *VM) call(callable Value, args []Value, kwargs map[string]Value) (Value, error) {
	switch fn := callable.(type) {
	case *PyBuiltinFunc:
		return fn.Fn(args, kwargs)

	case *PyGoFunc:
		// Call Go function with gopher-lua style stack-based API
		return vm.callGoFunction(fn, args)

	case *PyFunction:
		return vm.callFunction(fn, args, kwargs)

	case *PyMethod:
		// Prepend instance to args
		allArgs := append([]Value{fn.Instance}, args...)
		return vm.callFunction(fn.Func, allArgs, kwargs)

	case *PyClass:
		// Check for custom metaclass __call__ override
		if fn.Metaclass != nil {
			for _, cls := range fn.Metaclass.Mro {
				if callMethod, ok := cls.Dict["__call__"]; ok {
					// Skip if this is the base 'type' class (default behavior)
					if cls.Name == "type" {
						break
					}
					// Invoke metaclass __call__ with (cls, *args, **kwargs)
					mcArgs := append([]Value{fn}, args...)
					switch cm := callMethod.(type) {
					case *PyFunction:
						return vm.callFunction(cm, mcArgs, kwargs)
					case *PyBuiltinFunc:
						return cm.Fn(mcArgs, kwargs)
					}
				}
			}
		}

		// Check for abstract methods - prevent instantiation
		if abstractMethods, ok := fn.Dict["__abstractmethods__"]; ok {
			if absList, ok := abstractMethods.(*PyList); ok && len(absList.Items) > 0 {
				names := make([]string, len(absList.Items))
				for i, item := range absList.Items {
					if s, ok := item.(*PyString); ok {
						names[i] = s.Value
					}
				}
				// Sort for consistent error messages
				for i := 0; i < len(names); i++ {
					for j := i + 1; j < len(names); j++ {
						if names[i] > names[j] {
							names[i], names[j] = names[j], names[i]
						}
					}
				}
				plural := ""
				if len(names) > 1 {
					plural = "s"
				}
				methodList := names[0]
				for k := 1; k < len(names); k++ {
					if k == len(names)-1 {
						methodList += " and " + names[k]
					} else {
						methodList += ", " + names[k]
					}
				}
				return nil, fmt.Errorf("TypeError: Can't instantiate abstract class %s with abstract method%s %s", fn.Name, plural, methodList)
			}
		}

		// Step 1: Call __new__ to create the instance
		var instance Value
		newFound := false
		for _, cls := range fn.Mro {
			if newMethod, ok := cls.Dict["__new__"]; ok {
				newArgs := append([]Value{fn}, args...)
				var err error
				switch nm := newMethod.(type) {
				case *PyFunction:
					instance, err = vm.callFunction(nm, newArgs, kwargs)
				case *PyBuiltinFunc:
					instance, err = nm.Fn(newArgs, kwargs)
				case *PyStaticMethod:
					// __new__ may be explicitly decorated with @staticmethod
					instance, err = vm.call(nm.Func, newArgs, kwargs)
				}
				if err != nil {
					return nil, err
				}
				newFound = true
				break
			}
		}
		if !newFound {
			// Fallback (shouldn't happen if object is always in MRO)
			instance = &PyInstance{Class: fn, Dict: make(map[string]Value)}
		}

		// Step 2: If __new__ returned an instance of this class, call __init__
		if inst, ok := instance.(*PyInstance); ok && inst.Class == fn {
			// Special handling for exception classes - set args attribute
			if vm.isExceptionClass(fn) {
				tupleItems := make([]Value, len(args))
				copy(tupleItems, args)
				inst.Dict["args"] = &PyTuple{Items: tupleItems}
			}

			// Look for __init__ in class MRO
			for _, cls := range fn.Mro {
				if init, ok := cls.Dict["__init__"]; ok {
					if initFn, ok := init.(*PyFunction); ok {
						allArgs := append([]Value{inst}, args...)
						_, err := vm.callFunction(initFn, allArgs, kwargs)
						if err != nil {
							return nil, err
						}
					}
					break
				}
			}
		}
		return instance, nil

	case *PyUserData:
		// Check for __call__ method in metatable
		var typeName string
		if fn.Metatable != nil {
			for k, v := range fn.Metatable.Items {
				if ks, ok := k.(*PyString); ok && ks.Value == "__type__" {
					typeName = vm.str(v)
					break
				}
			}
		}
		if typeName != "" {
			if mt := typeMetatables[typeName]; mt != nil {
				if callMethod, ok := mt.Methods["__call__"]; ok {
					// Call the __call__ method with the userdata as first argument
					allArgs := append([]Value{fn}, args...)
					return vm.callGoFunction(&PyGoFunc{Name: "__call__", Fn: callMethod}, allArgs)
				}
			}
		}
		return nil, fmt.Errorf("'%s' object is not callable", vm.typeName(callable))

	case *PyInstance:
		// Check for __call__ method
		// We need to look up __call__ via MRO and call it with kwargs support
		for _, cls := range fn.Class.Mro {
			if method, ok := cls.Dict["__call__"]; ok {
				if callFn, ok := method.(*PyFunction); ok {
					// Prepend instance as self
					allArgs := append([]Value{fn}, args...)
					return vm.callFunction(callFn, allArgs, kwargs)
				}
				if callBuiltin, ok := method.(*PyBuiltinFunc); ok {
					allArgs := append([]Value{fn}, args...)
					return callBuiltin.Fn(allArgs, kwargs)
				}
			}
		}
		// Fallback: check class dict directly if MRO is empty
		if len(fn.Class.Mro) == 0 {
			if method, ok := fn.Class.Dict["__call__"]; ok {
				if callFn, ok := method.(*PyFunction); ok {
					allArgs := append([]Value{fn}, args...)
					return vm.callFunction(callFn, allArgs, kwargs)
				}
			}
		}
		return nil, fmt.Errorf("'%s' object is not callable", vm.typeName(callable))
	}
	return nil, fmt.Errorf("'%s' object is not callable", vm.typeName(callable))
}

// callGoFunction calls a Go function with stack-based argument passing
func (vm *VM) callGoFunction(fn *PyGoFunc, args []Value) (Value, error) {
	// Save current frame state
	oldFrame := vm.frame

	// Create a temporary frame for the Go function call
	tempFrame := &Frame{
		Stack:    make([]Value, len(args)+16),
		SP:       0,
		Globals:  vm.Globals,
		Builtins: vm.builtins,
	}

	// Push arguments onto the temporary frame's stack
	for _, arg := range args {
		tempFrame.Stack[tempFrame.SP] = arg
		tempFrame.SP++
	}

	vm.frame = tempFrame

	// Call the Go function - it returns number of results
	var nResults int
	var panicErr *PyPanicError
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Check if it's a typed PyPanicError for better exception mapping
				if pe, ok := r.(*PyPanicError); ok {
					panicErr = pe
				} else {
					// Generic panic - convert to string
					panicErr = &PyPanicError{
						ExcType: "RuntimeError",
						Message: fmt.Sprintf("%v", r),
					}
				}
				nResults = -1 // Indicate error
			}
		}()
		nResults = fn.Fn(vm)
	}()

	// Restore frame
	vm.frame = oldFrame

	// Handle error case
	if nResults < 0 {
		return nil, fmt.Errorf("%s: %s", panicErr.ExcType, panicErr.Message)
	}

	// Get results from stack
	if nResults == 0 {
		return None, nil
	} else if nResults == 1 {
		return tempFrame.Stack[tempFrame.SP-1], nil
	} else {
		// Multiple returns - return as tuple
		results := make([]Value, nResults)
		for i := 0; i < nResults; i++ {
			results[i] = tempFrame.Stack[tempFrame.SP-nResults+i]
		}
		return &PyTuple{Items: results}, nil
	}
}

func (vm *VM) callFunction(fn *PyFunction, args []Value, kwargs map[string]Value) (Value, error) {
	code := fn.Code

	// Check if this is a generator or coroutine - if so, create the appropriate object
	// instead of executing immediately
	if code.Flags&FlagGenerator != 0 {
		return vm.createGenerator(fn, args, kwargs)
	}
	if code.Flags&FlagCoroutine != 0 {
		return vm.createCoroutine(fn, args, kwargs)
	}
	if code.Flags&FlagAsyncGenerator != 0 {
		// For now, treat async generators like coroutines
		return vm.createCoroutine(fn, args, kwargs)
	}

	// Create new frame for regular function call
	frame := vm.createFunctionFrame(fn, args, kwargs)

	// Push frame
	vm.frames = append(vm.frames, frame)
	oldFrame := vm.frame
	vm.frame = frame

	// Execute
	result, err := vm.run()

	// Pop frame - but only if exception handling didn't already unwind past us
	if err != errExceptionHandledInOuterFrame {
		vm.frame = oldFrame
	}

	return result, err
}

// createFunctionFrame creates a new frame for a function call without executing it
func (vm *VM) createFunctionFrame(fn *PyFunction, args []Value, kwargs map[string]Value) *Frame {
	code := fn.Code

	// Create new frame
	frame := &Frame{
		Code:     code,
		IP:       0,
		Stack:    make([]Value, code.StackSize+16), // Pre-allocate
		SP:       0,
		Locals:   make([]Value, len(code.VarNames)),
		Globals:  fn.Globals,
		Builtins: vm.builtins,
	}

	// Set up closure cells
	// Cells include CellVars (our variables captured by inner functions) and FreeVars (from closure)
	numCells := len(code.CellVars) + len(code.FreeVars)
	if numCells > 0 || len(fn.Closure) > 0 {
		frame.Cells = make([]*PyCell, numCells)
		// CellVars are new cells for our locals that will be captured
		for i := 0; i < len(code.CellVars); i++ {
			frame.Cells[i] = &PyCell{}
		}
		// FreeVars come from the function's closure
		for i, cell := range fn.Closure {
			frame.Cells[len(code.CellVars)+i] = cell
		}
	}

	// Bind positional arguments to locals (up to ArgCount)
	numPositional := code.ArgCount
	if numPositional > len(args) {
		numPositional = len(args)
	}
	for i := 0; i < numPositional; i++ {
		frame.Locals[i] = args[i]
	}

	// Handle *args: collect extra positional arguments into a tuple
	if code.Flags&FlagVarArgs != 0 {
		varArgsIdx := code.ArgCount + code.KwOnlyArgCount
		if varArgsIdx < len(frame.Locals) {
			if len(args) > code.ArgCount {
				extraArgs := args[code.ArgCount:]
				items := make([]Value, len(extraArgs))
				copy(items, extraArgs)
				frame.Locals[varArgsIdx] = &PyTuple{Items: items}
			} else {
				frame.Locals[varArgsIdx] = &PyTuple{Items: []Value{}}
			}
		}
	}

	// If any arguments are CellVars (captured by inner functions), initialize cells with args
	// CellVars contain names of captured parameters - match by name against VarNames
	for cellIdx, cellName := range code.CellVars {
		// Find if this cell var corresponds to a parameter (in first ArgCount VarNames)
		for argIdx := 0; argIdx < code.ArgCount && argIdx < len(code.VarNames); argIdx++ {
			if code.VarNames[argIdx] == cellName && argIdx < len(args) {
				// This parameter is captured, initialize its cell with the argument
				if cellIdx < len(frame.Cells) && frame.Cells[cellIdx] != nil {
					frame.Cells[cellIdx].Value = args[argIdx]
				}
				break
			}
		}
	}

	// Apply keyword arguments to the appropriate local slots
	if kwargs != nil {
		for name, val := range kwargs {
			// Find the parameter index by name in positional and kw-only params
			for i, varName := range code.VarNames {
				if varName == name && i < code.ArgCount+code.KwOnlyArgCount {
					frame.Locals[i] = val
					break
				}
			}
		}
	}

	// Handle **kwargs: collect extra keyword arguments into a dict
	if code.Flags&FlagVarKeywords != 0 {
		kwArgsIdx := code.ArgCount + code.KwOnlyArgCount
		if code.Flags&FlagVarArgs != 0 {
			kwArgsIdx++ // Skip the *args slot
		}
		if kwArgsIdx < len(frame.Locals) {
			kwargsDict := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
			if kwargs != nil {
				for name, val := range kwargs {
					// Check if this kwarg matches a named parameter
					isNamedParam := false
					for i := 0; i < code.ArgCount+code.KwOnlyArgCount && i < len(code.VarNames); i++ {
						if code.VarNames[i] == name {
							isNamedParam = true
							break
						}
					}
					if !isNamedParam {
						kwargsDict.DictSet(&PyString{Value: name}, val, vm)
					}
				}
			}
			frame.Locals[kwArgsIdx] = kwargsDict
		}
	}

	// Apply defaults for missing arguments
	if fn.Defaults != nil {
		numDefaults := len(fn.Defaults.Items)
		startDefault := code.ArgCount - numDefaults
		for i := 0; i < numDefaults; i++ {
			argIdx := startDefault + i
			if argIdx < len(frame.Locals) && frame.Locals[argIdx] == nil {
				frame.Locals[argIdx] = fn.Defaults.Items[i]
			}
		}
	}

	return frame
}

// createGenerator creates a new generator object from a generator function
func (vm *VM) createGenerator(fn *PyFunction, args []Value, kwargs map[string]Value) (*PyGenerator, error) {
	frame := vm.createFunctionFrame(fn, args, kwargs)
	return &PyGenerator{
		Frame: frame,
		Code:  fn.Code,
		Name:  fn.Name,
		State: GenCreated,
	}, nil
}

// createCoroutine creates a new coroutine object from an async function
func (vm *VM) createCoroutine(fn *PyFunction, args []Value, kwargs map[string]Value) (*PyCoroutine, error) {
	frame := vm.createFunctionFrame(fn, args, kwargs)
	return &PyCoroutine{
		Frame: frame,
		Code:  fn.Code,
		Name:  fn.Name,
		State: GenCreated,
	}, nil
}
