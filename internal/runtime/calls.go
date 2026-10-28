package runtime

import "fmt"

// Function calls

// callFunctionWithSelf calls fn with self followed by args. callFunction
// copies its arguments into the new frame without retaining the slice, so for
// small arities the combined argument list is built on the Go stack instead
// of the heap.
func (vm *VM) callFunctionWithSelf(fn *PyFunction, self Value, args []Value, kwargs map[string]Value) (Value, error) {
	if len(args) < 4 {
		var buf [4]Value
		buf[0] = self
		n := copy(buf[1:], args)
		return vm.callFunction(fn, buf[:n+1], kwargs)
	}
	return vm.callFunction(fn, prependArg(self, args), kwargs)
}

// call1 calls callable with a single positional argument, avoiding a heap
// allocated argument slice when callable is a Python function.
func (vm *VM) call1(callable Value, arg Value) (Value, error) {
	if fn, ok := callable.(*PyFunction); ok {
		args := [1]Value{arg}
		return vm.callFunction(fn, args[:], nil)
	}
	return vm.call(callable, []Value{arg}, nil)
}

func (vm *VM) call(callable Value, args []Value, kwargs map[string]Value) (Value, error) {
	switch fn := callable.(type) {
	case *PyBuiltinFunc:
		return fn.Fn(args, kwargs)

	case *PyGoFunc:
		// Call Go function with gopher-lua style stack-based API
		return vm.callGoFunction(fn, args, kwargs)

	case *PyFunction:
		return vm.callFunction(fn, args, kwargs)

	case *PyMethod:
		// Prepend instance to args
		return vm.callFunctionWithSelf(fn.Func, fn.Instance, args, kwargs)

	case *PyClass:
		slotArgs := make([]Value, len(args)+1)
		copy(slotArgs[1:], args)
		return vm.callClass(fn, slotArgs, kwargs)

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
					return vm.callGoFunction(&PyGoFunc{Name: "__call__", Fn: callMethod}, allArgs, kwargs)
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
				if callBuiltin, ok := method.(*PyBuiltinFunc); ok {
					allArgs := append([]Value{fn}, args...)
					return callBuiltin.Fn(allArgs, kwargs)
				}
			}
		}
		return nil, fmt.Errorf("'%s' object is not callable", vm.typeName(callable))
	}
	return nil, fmt.Errorf("TypeError: '%s' object is not callable", vm.typeName(callable))
}

// callGoFunction calls a Go function with stack-based argument passing.
// kwargs (may be nil) are exposed to the Go function via vm.Kwarg/vm.OptionalArg.
func (vm *VM) callGoFunction(fn *PyGoFunc, args []Value, kwargs map[string]Value) (Value, error) {
	// Save current frame state
	oldFrame := vm.frame

	// Create a temporary frame for the Go function call
	tempFrame := &Frame{
		Stack:    make([]Value, len(args)+16),
		SP:       0,
		Globals:  vm.Globals,
		Builtins: vm.builtins,
		Kwargs:   kwargs,
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
	var panicExc *PyException
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Check if it's a typed PyPanicError for better exception mapping
				if pe, ok := r.(*PyPanicError); ok {
					panicErr = pe
				} else if exc, ok := r.(*PyException); ok {
					// A Python exception raised as-is (e.g. from a user __repr__)
					panicExc = exc
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
		if panicExc != nil {
			return nil, panicExc
		}
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

	// Check recursion depth before pushing frame
	if vm.maxRecursionDepth > 0 && int64(len(vm.frames)) >= vm.maxRecursionDepth {
		return nil, fmt.Errorf("RecursionError: maximum recursion depth exceeded")
	}

	// Create new frame for regular function call
	frame, err := vm.createFunctionFrame(fn, args, kwargs)
	if err != nil {
		return nil, err
	}

	// Track frame memory: ~16 bytes per stack slot, ~16 bytes per local, ~24 bytes per cell
	if vm.maxMemoryBytes > 0 {
		frameBytes := int64(len(frame.Stack))*16 + int64(len(frame.Locals))*16 + int64(len(frame.Cells))*24
		if err := vm.trackAlloc(frameBytes); err != nil {
			return nil, err
		}
	}

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

	// A frame that returned normally (RETURN_VALUE popped it off vm.frames)
	// is no longer referenced by anything and can be recycled.
	if err == nil && vm.frame != frame && (len(vm.frames) == 0 || vm.frames[len(vm.frames)-1] != frame) {
		vm.releaseFrame(frame)
	}

	return result, err
}

// maxFramePool bounds the number of finished frames kept for reuse per VM.
const maxFramePool = 64

// newFrame returns a zeroed frame whose slots buffer holds n nil values,
// reusing a frame recycled by releaseFrame when one is available.
func (vm *VM) newFrame(n int) *Frame {
	if last := len(vm.framePool) - 1; last >= 0 {
		f := vm.framePool[last]
		vm.framePool[last] = nil
		vm.framePool = vm.framePool[:last]
		if cap(f.slots) >= n {
			f.slots = f.slots[:n] // cleared by releaseFrame
		} else {
			f.slots = make([]Value, n)
		}
		return f
	}
	return &Frame{slots: make([]Value, n)}
}

// releaseFrame resets a finished function frame and returns it to the pool.
// The caller must guarantee nothing references the frame or its slots any
// more: only frames that completed via RETURN_VALUE in callFunction qualify
// (generator/coroutine frames and frames unwound by exceptions never do).
func (vm *VM) releaseFrame(f *Frame) {
	if len(vm.framePool) >= maxFramePool {
		return
	}
	clear(f.slots)
	*f = Frame{slots: f.slots, BlockStack: f.BlockStack[:0]}
	vm.framePool = append(vm.framePool, f)
}

// prependArg returns a new slice holding first followed by args.
func prependArg(first Value, args []Value) []Value {
	all := make([]Value, len(args)+1)
	all[0] = first
	copy(all[1:], args)
	return all
}

// callClass calls a class object. slotArgs[0] is scratch space that is
// overwritten with the class or the new instance, and slotArgs[1:] holds the
// positional arguments. slotArgs may alias the caller's operand stack, so it
// is only passed on to callees that copy their arguments before running any
// other code (Python functions and object.__new__); other callees get a copy.
func (vm *VM) callClass(fn *PyClass, slotArgs []Value, kwargs map[string]Value) (Value, error) {
	// Check for custom metaclass __call__ override
	typeClass, _ := vm.builtins["type"].(*PyClass)
	if fn.Metaclass != nil && fn.Metaclass != typeClass {
		for _, cls := range fn.Metaclass.Mro {
			if cls == typeClass {
				break
			}
			if callMethod, ok := cls.Dict["__call__"]; ok {
				switch cm := callMethod.(type) {
				case *PyFunction:
					slotArgs[0] = fn
					return vm.callFunction(cm, slotArgs, kwargs)
				case *PyBuiltinFunc:
					return cm.Fn(prependArg(fn, slotArgs[1:]), kwargs)
				}
			}
		}
	}

	return vm.classCall(fn, slotArgs, kwargs)
}

// defaultClassCall implements the default class instantiation logic:
// abstract check → __new__ → __init__
func (vm *VM) defaultClassCall(fn *PyClass, args []Value, kwargs map[string]Value) (Value, error) {
	slotArgs := make([]Value, len(args)+1)
	copy(slotArgs[1:], args)
	return vm.classCall(fn, slotArgs, kwargs)
}

// classCall is defaultClassCall with arguments in the slotArgs layout
// described on callClass.
func (vm *VM) classCall(fn *PyClass, slotArgs []Value, kwargs map[string]Value) (Value, error) {
	args := slotArgs[1:]

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
			var err error
			switch nm := newMethod.(type) {
			case *PyFunction:
				slotArgs[0] = fn
				instance, err = vm.callFunction(nm, slotArgs, kwargs)
			case *PyBuiltinFunc:
				if nm == vm.objectNew {
					// object.__new__ only reads its class argument.
					slotArgs[0] = fn
					instance, err = nm.Fn(slotArgs, kwargs)
				} else {
					instance, err = nm.Fn(prependArg(fn, args), kwargs)
				}
			case *PyStaticMethod:
				// __new__ may be explicitly decorated with @staticmethod
				instance, err = vm.call(nm.Func, prependArg(fn, args), kwargs)
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
				switch initFn := init.(type) {
				case *PyFunction:
					slotArgs[0] = inst
					_, err := vm.callFunction(initFn, slotArgs, kwargs)
					if err != nil {
						return nil, err
					}
				case *PyBuiltinFunc:
					_, err := initFn.Fn(prependArg(inst, args), kwargs)
					if err != nil {
						return nil, err
					}
				case *PyStaticMethod:
					_, err := vm.call(initFn.Func, prependArg(inst, args), kwargs)
					if err != nil {
						return nil, err
					}
				}
				break
			}
		}
	}
	return instance, nil
}

// createFunctionFrame creates a new frame for a function call without executing it.
// Returns an error for invalid call shapes (duplicate args, missing required args).
func (vm *VM) createFunctionFrame(fn *PyFunction, args []Value, kwargs map[string]Value) (*Frame, error) {
	code := fn.Code

	// Create new frame. Locals and the operand stack share a single backing
	// allocation (locals first, capacity-capped so they can never grow into
	// the stack); the stack is replaced with a fresh slice if it ever grows.
	numLocals := len(code.VarNames)
	frame := vm.newFrame(numLocals + code.StackSize + 16) // Pre-allocate
	frame.Code = code
	frame.Stack = frame.slots[numLocals:]
	frame.Locals = frame.slots[:numLocals:numLocals]
	frame.Globals = fn.Globals
	frame.Builtins = vm.builtins

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
	// This checks both positional and keyword-only parameters.
	totalNamedParams := code.ArgCount + code.KwOnlyArgCount
	for cellIdx, cellName := range code.CellVars {
		// Find if this cell var corresponds to a parameter
		for argIdx := 0; argIdx < totalNamedParams && argIdx < len(code.VarNames); argIdx++ {
			if code.VarNames[argIdx] == cellName {
				// This parameter is captured, initialize its cell with the local value
				if cellIdx < len(frame.Cells) && frame.Cells[cellIdx] != nil {
					if argIdx < len(frame.Locals) && frame.Locals[argIdx] != nil {
						frame.Cells[cellIdx].Value = frame.Locals[argIdx]
					}
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
					if i < numPositional {
						return nil, fmt.Errorf("TypeError: %s() got multiple values for argument '%s'", fn.Name, name)
					}
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
				// Keep the call order of the keyword arguments
				for _, name := range vm.kwargNames(kwargs) {
					val := kwargs[name]
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

	// Apply keyword-only defaults for missing kwonly arguments
	if fn.KwDefaults != nil {
		for kwName, kwDefault := range fn.KwDefaults {
			for i := code.ArgCount; i < code.ArgCount+code.KwOnlyArgCount && i < len(code.VarNames); i++ {
				if code.VarNames[i] == kwName && i < len(frame.Locals) && frame.Locals[i] == nil {
					frame.Locals[i] = kwDefault
					break
				}
			}
		}
	}

	// Re-initialize cells from locals: kwonly defaults and kwargs may have been applied
	// after the initial cell initialization, so update cells that still have nil values.
	for cellIdx, cellName := range code.CellVars {
		if cellIdx < len(frame.Cells) && frame.Cells[cellIdx] != nil && frame.Cells[cellIdx].Value == nil {
			for i := 0; i < totalNamedParams && i < len(code.VarNames); i++ {
				if code.VarNames[i] == cellName && i < len(frame.Locals) && frame.Locals[i] != nil {
					frame.Cells[cellIdx].Value = frame.Locals[i]
					break
				}
			}
		}
	}

	// Required positional and kw-only parameters must be bound.
	numDefaults := 0
	if fn.Defaults != nil {
		numDefaults = len(fn.Defaults.Items)
	}
	requiredPosEnd := code.ArgCount - numDefaults
	var missing []string
	for i := 0; i < requiredPosEnd && i < len(code.VarNames); i++ {
		if frame.Locals[i] == nil {
			missing = append(missing, "'"+code.VarNames[i]+"'")
		}
	}
	if len(missing) > 0 {
		noun := "argument"
		if len(missing) > 1 {
			noun = "arguments"
		}
		return nil, fmt.Errorf("TypeError: %s() missing %d required positional %s: %s",
			fn.Name, len(missing), noun, joinMissing(missing))
	}
	for i := code.ArgCount; i < code.ArgCount+code.KwOnlyArgCount && i < len(code.VarNames); i++ {
		if frame.Locals[i] == nil {
			if fn.KwDefaults == nil || fn.KwDefaults[code.VarNames[i]] == nil {
				return nil, fmt.Errorf("TypeError: %s() missing required keyword-only argument: '%s'",
					fn.Name, code.VarNames[i])
			}
		}
	}

	return frame, nil
}

// joinMissing joins a list of quoted parameter names into a CPython-style
// "and"-list, e.g. "'a' and 'b'" or "'a', 'b', and 'c'".
func joinMissing(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	case 2:
		return names[0] + " and " + names[1]
	}
	s := ""
	for i, n := range names {
		if i == len(names)-1 {
			s += "and " + n
		} else {
			s += n + ", "
		}
	}
	return s
}

// createGenerator creates a new generator object from a generator function
func (vm *VM) createGenerator(fn *PyFunction, args []Value, kwargs map[string]Value) (*PyGenerator, error) {
	frame, err := vm.createFunctionFrame(fn, args, kwargs)
	if err != nil {
		return nil, err
	}
	return &PyGenerator{
		Frame: frame,
		Code:  fn.Code,
		Name:  fn.Name,
		State: GenCreated,
	}, nil
}

// createCoroutine creates a new coroutine object from an async function
func (vm *VM) createCoroutine(fn *PyFunction, args []Value, kwargs map[string]Value) (*PyCoroutine, error) {
	frame, err := vm.createFunctionFrame(fn, args, kwargs)
	if err != nil {
		return nil, err
	}
	return &PyCoroutine{
		Frame: frame,
		Code:  fn.Code,
		Name:  fn.Name,
		State: GenCreated,
	}, nil
}
