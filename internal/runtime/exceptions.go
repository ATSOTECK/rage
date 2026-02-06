package runtime

// Exception handling helpers

// createException creates a PyException from a value
func (vm *VM) createException(excVal Value, cause Value) *PyException {
	exc := &PyException{}

	switch v := excVal.(type) {
	case *PyException:
		// Already an exception, return as-is
		return v
	case *PyClass:
		// Exception class without arguments: raise ValueError
		if vm.isExceptionClass(v) {
			exc.ExcType = v
			exc.Args = &PyTuple{Items: []Value{}}
			exc.Message = v.Name
		} else {
			// Not an exception class
			exc.ExcType = vm.builtins["TypeError"].(*PyClass)
			exc.Args = &PyTuple{Items: []Value{&PyString{Value: "exceptions must derive from BaseException"}}}
			exc.Message = "TypeError: exceptions must derive from BaseException"
		}
	case *PyInstance:
		// Already instantiated exception
		if vm.isExceptionClass(v.Class) {
			exc.ExcType = v.Class
			if args, ok := v.Dict["args"]; ok {
				if t, ok := args.(*PyTuple); ok {
					exc.Args = t
				}
			}
			if exc.Args == nil {
				exc.Args = &PyTuple{Items: []Value{}}
			}
			exc.Message = vm.str(v)
		} else {
			exc.ExcType = vm.builtins["TypeError"].(*PyClass)
			exc.Args = &PyTuple{Items: []Value{&PyString{Value: "exceptions must derive from BaseException"}}}
			exc.Message = "TypeError: exceptions must derive from BaseException"
		}
	case *PyString:
		// String used as exception (legacy style, but we'll support it)
		exc.ExcType = vm.builtins["Exception"].(*PyClass)
		exc.Args = &PyTuple{Items: []Value{v}}
		exc.Message = v.Value
	default:
		exc.ExcType = vm.builtins["TypeError"].(*PyClass)
		exc.Args = &PyTuple{Items: []Value{&PyString{Value: "exceptions must derive from BaseException"}}}
		exc.Message = "TypeError: exceptions must derive from BaseException"
	}

	if cause != nil {
		exc.Cause = vm.createException(cause, nil)
	}

	return exc
}

// isExceptionClass checks if a class is an exception class (inherits from BaseException)
func (vm *VM) isExceptionClass(cls *PyClass) bool {
	baseExc, ok := vm.builtins["BaseException"].(*PyClass)
	if !ok {
		return false
	}
	for _, mroClass := range cls.Mro {
		if mroClass == baseExc {
			return true
		}
	}
	return false
}

// exceptionMatches checks if an exception matches a type for except clause
func (vm *VM) exceptionMatches(exc *PyException, exceptionType Value) bool {
	switch t := exceptionType.(type) {
	case *PyClass:
		// Check if exc.ExcType is t or subclass of t
		for _, mroClass := range exc.ExcType.Mro {
			if mroClass == t {
				return true
			}
		}
		return false
	case *PyTuple:
		// Tuple of exception types - match any
		for _, item := range t.Items {
			if vm.exceptionMatches(exc, item) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// buildTraceback builds a traceback from current frame stack
func (vm *VM) buildTraceback() []TracebackEntry {
	var tb []TracebackEntry
	for i := len(vm.frames) - 1; i >= 0; i-- {
		f := vm.frames[i]
		line := f.Code.LineForOffset(f.IP)
		tb = append(tb, TracebackEntry{
			Filename: f.Code.Filename,
			Line:     line,
			Function: f.Code.Name,
		})
	}
	return tb
}

// handleException unwinds the stack looking for exception handlers
// Returns (nil, nil) if a handler was found and we should continue execution
// Returns (nil, error) if no handler was found
func (vm *VM) handleException(exc *PyException) (Value, error) {
	vm.currentException = exc
	vm.lastException = exc

	for len(vm.frames) > 0 {
		frame := vm.frame

		// Search block stack for exception handler
		for len(frame.BlockStack) > 0 {
			block := frame.BlockStack[len(frame.BlockStack)-1]
			frame.BlockStack = frame.BlockStack[:len(frame.BlockStack)-1]

			switch block.Type {
			case BlockExcept:
				// Found exception handler - restore stack and jump to handler
				frame.SP = block.Level
				frame.IP = block.Handler
				vm.push(exc)    // Push exception onto stack for handler
				return nil, nil // Continue execution at handler

			case BlockFinally:
				// Must execute finally block first
				frame.SP = block.Level
				frame.IP = block.Handler
				vm.push(exc)    // Push exception for finally to potentially re-raise
				return nil, nil // Continue execution at finally

			case BlockLoop:
				// Skip loop blocks when unwinding for exception
				continue
			}
		}

		// No handler in this frame, pop frame and continue unwinding
		vm.frames = vm.frames[:len(vm.frames)-1]
		if len(vm.frames) > 0 {
			vm.frame = vm.frames[len(vm.frames)-1]
		}
	}

	// No handler found anywhere - exception propagates to caller
	return nil, exc
}

// exceptionPrefixes maps error message prefixes to exception type names
// Sorted by length descending for longest-match-first semantics
var exceptionPrefixes = []struct {
	prefix   string
	excName  string
	fallback string // Optional fallback exception name
}{
	{"ModuleNotFoundError", "ModuleNotFoundError", ""},
	{"ZeroDivisionError", "ZeroDivisionError", ""},
	{"FileNotFoundError", "FileNotFoundError", ""},
	{"PermissionError", "PermissionError", ""},
	{"FileExistsError", "FileExistsError", ""},
	{"AttributeError", "AttributeError", ""},
	{"ImportError", "ImportError", ""},
	{"IndexError", "IndexError", ""},
	{"ValueError", "ValueError", ""},
	{"TypeError", "TypeError", ""},
	{"NameError", "NameError", ""},
	{"MemoryError", "MemoryError", ""},
	{"KeyError", "KeyError", ""},
	{"IOError", "IOError", "OSError"},
	{"OSError", "OSError", ""},
}

// wrapGoError converts a Go error to a Python exception
func (vm *VM) wrapGoError(err error) *PyException {
	if pyExc, ok := err.(*PyException); ok {
		return pyExc
	}

	errStr := err.Error()

	// Find exception type using prefix matching (prefixes sorted by length desc)
	var excClass *PyClass
	for _, ep := range exceptionPrefixes {
		if len(errStr) >= len(ep.prefix) && errStr[:len(ep.prefix)] == ep.prefix {
			if exc, ok := vm.builtins[ep.excName]; ok {
				if cls, ok := exc.(*PyClass); ok {
					excClass = cls
				}
			}
			if excClass == nil && ep.fallback != "" {
				if fb, ok := vm.builtins[ep.fallback]; ok {
					if cls, ok := fb.(*PyClass); ok {
						excClass = cls
					}
				}
			}
			break
		}
	}
	if excClass == nil {
		// Fallback to RuntimeError with safe type assertion
		if re, ok := vm.builtins["RuntimeError"]; ok {
			if cls, ok := re.(*PyClass); ok {
				excClass = cls
			}
		}
	}

	return &PyException{
		ExcType: excClass,
		Args:    &PyTuple{Items: []Value{&PyString{Value: errStr}}},
		Message: errStr,
	}
}
