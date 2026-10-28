package runtime

import (
	"context"
	"fmt"
	"strings"
)

// Exception handling helpers

// builtinClass safely retrieves a builtin class by name, returning nil if not found or wrong type.
func (vm *VM) builtinClass(name string) *PyClass {
	if v, ok := vm.builtins[name]; ok {
		if cls, ok := v.(*PyClass); ok {
			return cls
		}
	}
	return nil
}

// createException creates a PyException from a value
func (vm *VM) createException(excVal Value, cause Value) *PyException {
	exc := &PyException{}

	switch v := excVal.(type) {
	case *PyException:
		if cause == nil {
			// No cause: return existing exception unchanged.
			return v
		}
		// Attach cause without mutating the input — it may be shared
		// (e.g. vm.lastException), and `raise existing from new` should
		// not rewrite history on the original.
		clone := *v
		if cause == None {
			clone.Cause = nil
		} else {
			clone.Cause = vm.createException(cause, nil)
		}
		clone.SuppressContext = true
		return &clone
	case *PyClass:
		// Exception class without arguments: raise ValueError
		if vm.isExceptionClass(v) {
			exc.ExcType = v
			exc.Args = &PyTuple{Items: []Value{}}
			exc.Message = v.Name
		} else {
			// Not an exception class
			exc.ExcType = vm.builtinClass("TypeError")
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
			// Link back to the instance so user-defined attributes, methods,
			// and __str__ remain reachable from the caught exception.
			exc.Instance = v
			// Copy __notes__ from instance to PyException
			if notes, ok := v.Dict["__notes__"]; ok {
				if notesList, ok := notes.(*PyList); ok {
					exc.Notes = notesList
				}
			}
		} else {
			exc.ExcType = vm.builtinClass("TypeError")
			exc.Args = &PyTuple{Items: []Value{&PyString{Value: "exceptions must derive from BaseException"}}}
			exc.Message = "TypeError: exceptions must derive from BaseException"
		}
	case *PyString:
		// String used as exception (legacy style, but we'll support it)
		exc.ExcType = vm.builtinClass("Exception")
		exc.Args = &PyTuple{Items: []Value{v}}
		exc.Message = v.Value
	default:
		exc.ExcType = vm.builtinClass("TypeError")
		exc.Args = &PyTuple{Items: []Value{&PyString{Value: "exceptions must derive from BaseException"}}}
		exc.Message = "TypeError: exceptions must derive from BaseException"
	}

	if cause != nil {
		if cause == None {
			exc.Cause = nil
		} else {
			exc.Cause = vm.createException(cause, nil)
		}
		exc.SuppressContext = true
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
		// If exc has a typed ExcType, check class hierarchy
		if exc.ExcType != nil {
			for _, mroClass := range exc.ExcType.Mro {
				if mroClass == t {
					return true
				}
			}
			return false
		}
		// Fall back to name matching for exceptions created without ExcType
		return vm.exceptionNameMatches(exc.TypeName, t)

	case *PyTuple:
		// Tuple of exception types - match any
		for _, item := range t.Items {
			if vm.exceptionMatches(exc, item) {
				return true
			}
		}
		return false
	case *UnionType:
		// PEP 604: union of exception types - match any
		for _, item := range t.Args {
			if vm.exceptionMatches(exc, item) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// exceptionNameMatches checks if an exception type name matches a class by walking the MRO
func (vm *VM) exceptionNameMatches(typeName string, cls *PyClass) bool {
	// Direct name match
	if cls.Name == typeName {
		return true
	}
	// Check if the exception's type name matches any parent in the class hierarchy
	// Look up the exception class by name and check MRO
	if excClass, ok := vm.builtins[typeName]; ok {
		if excPyClass, ok := excClass.(*PyClass); ok {
			for _, mroClass := range excPyClass.Mro {
				if mroClass == cls {
					return true
				}
			}
		}
	}
	return false
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

	if len(vm.frames) == 0 {
		return nil, exc
	}

	// A timed-out or cancelled execution must not be resumable from Python:
	// skip every handler (including except:/finally) and unwind the frame.
	if abortErr := vm.contextDoneError(); abortErr != nil {
		vm.popFrame()
		return nil, abortErr
	}

	frame := vm.frame

	// Search block stack for exception handler in the current frame only.
	// We don't cross function call boundaries because Go-level callers
	// (like iterNext) need a chance to catch exceptions (e.g. StopIteration)
	// before Python-level handlers in calling frames see them.
	for len(frame.BlockStack) > 0 {
		block := frame.BlockStack[len(frame.BlockStack)-1]
		frame.BlockStack = frame.BlockStack[:len(frame.BlockStack)-1]

		// A handler set up outside a running finally body means the exception
		// escaped that body, abandoning the return or break/continue it was
		// running for. Handlers inside the finally body leave it pending.
		if fin := frame.fin; fin != nil && block.Type != BlockLoop && len(frame.BlockStack) < fin.pendingDepth {
			fin.hasPendingReturn = false
			fin.pendingReturn = nil
			fin.hasPendingJump = false
			fin.pendingDepth = 0
		}

		switch block.Type {
		case BlockExcept:
			// Found exception handler - restore stack and jump to handler
			frame.SP = block.Level
			frame.IP = block.Handler
			// Restore excHandlerStack to the level when this try was set up.
			// This cleans up entries from any nested handlers that were abandoned
			// when the exception propagated through them.
			if block.ExcStackLevel < len(vm.excHandlerStack) {
				// Clear references in truncated portion for GC
				for i := block.ExcStackLevel; i < len(vm.excHandlerStack); i++ {
					vm.excHandlerStack[i] = nil
				}
				vm.excHandlerStack = vm.excHandlerStack[:block.ExcStackLevel]
			}
			vm.push(exc)    // Push exception onto stack for handler
			return nil, nil // Continue execution at handler

		case BlockFinally:
			// Must execute finally block first
			frame.SP = block.Level
			frame.IP = block.Handler
			vm.push(exc)                                             // Push exception for finally to potentially re-raise
			vm.excHandlerStack = append(vm.excHandlerStack, exc)     // Track for __context__ on new exceptions in finally
			return nil, nil // Continue execution at finally

		case BlockWith:
			// With block - must run __exit__ with exception info
			frame.SP = block.Level
			frame.IP = block.Handler
			vm.push(exc)    // Push exception for WITH_CLEANUP
			return nil, nil // Continue execution at cleanup handler

		case BlockExceptStar:
			// except* handler — wrap exception in ExceptionGroup if needed,
			// push onto stack and initialize exceptStarState
			frame.SP = block.Level
			frame.IP = block.Handler
			var leafExcs []*PyException
			var msg string
			var isBase bool
			if vm.isExceptionGroup(exc) && exc.Instance != nil {
				leafExcs = vm.getEGExceptions(exc.Instance)
				if m, ok := exc.Instance.Dict["message"].(*PyString); ok {
					msg = m.Value
				}
				// Only use BaseExceptionGroup if the class is exactly BaseExceptionGroup
				// (not ExceptionGroup which also has BaseExceptionGroup in MRO)
				isBase = vm.isOnlyBaseExceptionGroup(exc.Instance.Class)
			} else {
				leafExcs = []*PyException{exc}
				msg = exc.Type()
				isBase = false
			}
			vm.push(exc) // Push exception onto stack for handler
			vm.exceptStarStack = append(vm.exceptStarStack, ExceptStarState{
				Remaining: leafExcs,
				Message:   msg,
				IsBase:    isBase,
			})
			return nil, nil

		case BlockLoop:
			// Skip loop blocks when unwinding for exception
			continue
		}
	}

	// No handler in this frame, pop frame and propagate to caller.
	vm.popFrame()
	return nil, exc
}

// popFrame removes the current frame from the frame stack and makes the
// caller's frame current. The vacated slot is nilled so the popped frame can
// be garbage collected.
func (vm *VM) popFrame() {
	vm.frames[len(vm.frames)-1] = nil
	vm.frames = vm.frames[:len(vm.frames)-1]
	if len(vm.frames) > 0 {
		vm.frame = vm.frames[len(vm.frames)-1]
	}
}

// contextDoneError returns the TimeoutError or CancelledError for an
// execution whose context has finished, or nil while it is still running.
func (vm *VM) contextDoneError() error {
	if vm.ctx == nil {
		return nil
	}
	switch vm.ctx.Err() {
	case nil:
		return nil
	case context.DeadlineExceeded:
		return &TimeoutError{}
	default:
		return &CancelledError{}
	}
}

// exceptionPrefixes maps error message prefixes to exception type names
// Sorted by length descending for longest-match-first semantics
var exceptionPrefixes = []struct {
	prefix   string
	excName  string
	fallback string // Optional fallback exception name
}{
	// Sorted by length descending for longest-match-first semantics
	{"ConnectionRefusedError", "ConnectionRefusedError", ""},
	{"ConnectionAbortedError", "ConnectionAbortedError", ""},
	{"PendingDeprecationWarning", "PendingDeprecationWarning", ""},
	{"ConnectionResetError", "ConnectionResetError", ""},
	{"UnicodeTranslateError", "UnicodeTranslateError", ""},
	{"StopAsyncIteration", "StopAsyncIteration", ""},
	{"ModuleNotFoundError", "ModuleNotFoundError", ""},
	{"NotImplementedError", "NotImplementedError", ""},
	{"UnicodeDecodeError", "UnicodeDecodeError", ""},
	{"UnicodeEncodeError", "UnicodeEncodeError", ""},
	{"IsADirectoryError", "IsADirectoryError", ""},
	{"ProcessLookupError", "ProcessLookupError", ""},
	{"NotADirectoryError", "NotADirectoryError", ""},
	{"DeprecationWarning", "DeprecationWarning", ""},
	{"FileNotFoundError", "FileNotFoundError", ""},
	{"ChildProcessError", "ChildProcessError", ""},
	{"ZeroDivisionError", "ZeroDivisionError", ""},
	{"UnboundLocalError", "UnboundLocalError", ""},
	{"FloatingPointError", "FloatingPointError", ""},
	{"ConnectionError", "ConnectionError", ""},
	{"InterruptedError", "InterruptedError", ""},
	{"BlockingIOError", "BlockingIOError", ""},
	{"BrokenPipeError", "BrokenPipeError", ""},
	{"EncodingWarning", "EncodingWarning", ""},
	{"ResourceWarning", "ResourceWarning", ""},
	{"UnicodeWarning", "UnicodeWarning", ""},
	{"RuntimeWarning", "RuntimeWarning", ""},
	{"FileExistsError", "FileExistsError", ""},
	{"ArithmeticError", "ArithmeticError", ""},
	{"AttributeError", "AttributeError", ""},
	{"PermissionError", "PermissionError", ""},
	{"OverflowError", "OverflowError", ""},
	{"ImportWarning", "ImportWarning", ""},
	{"FutureWarning", "FutureWarning", ""},
	{"SyntaxWarning", "SyntaxWarning", ""},
	{"RecursionError", "RecursionError", ""},
	{"UnicodeError", "UnicodeError", ""},
	{"RuntimeError", "RuntimeError", ""},
	{"AssertionError", "AssertionError", ""},
	{"StopIteration", "StopIteration", ""},
	{"GeneratorExit", "GeneratorExit", ""},
	{"TimeoutError", "TimeoutError", ""},
	{"BytesWarning", "BytesWarning", ""},
	{"LookupError", "LookupError", ""},
	{"SyntaxError", "SyntaxError", ""},
	{"ImportError", "ImportError", ""},
	{"BufferError", "BufferError", ""},
	{"IndexError", "IndexError", ""},
	{"ValueError", "ValueError", ""},
	{"TypeError", "TypeError", ""},
	{"NameError", "NameError", ""},
	{"MemoryError", "MemoryError", ""},
	{"UserWarning", "UserWarning", ""},
	{"KeyError", "KeyError", ""},
	{"EOFError", "EOFError", ""},
	{"IOError", "IOError", "OSError"},
	{"OSError", "OSError", ""},
	{"Warning", "Warning", ""},
}

// opSymbol returns the Python operator symbol for a binary, in-place,
// or unary opcode ("+" for OpBinaryAdd, "+=" for OpInplaceAdd), falling back
// to the opcode name.
func opSymbol(op Opcode) string {
	switch {
	case op >= OpBinaryAdd && op <= OpBinaryXor:
		return binaryOpSymbol(op)
	case op == OpInplacePower:
		return "**="
	case op >= OpInplaceAdd && op <= OpInplaceXor:
		return binaryOpSymbol(op-OpInplaceAdd+OpBinaryAdd) + "="
	case op == OpUnaryPositive:
		return "+"
	case op == OpUnaryNegative:
		return "-"
	case op == OpUnaryInvert:
		return "~"
	}
	return op.String()
}

// unprefixedErrorPatterns classifies Go errors whose message carries no
// exception-type prefix. These messages mirror CPython's wording, so the
// wording identifies the exception type; without this they would all surface
// as RuntimeError and `except TypeError:` / `except NameError:` would miss them.
// Checked in order; the first pattern contained in the message wins.
var unprefixedErrorPatterns = []struct {
	substr  string
	excName string
}{
	{"' is not defined", "NameError"},
	{"has no attribute '", "AttributeError"},
	{"' is read-only", "AttributeError"},
	{"' has no getter", "AttributeError"},
	{"' has no setter", "AttributeError"},
	{"' has no deleter", "AttributeError"},
	{"cannot import name '", "ImportError"},
	{"values to unpack", "ValueError"},
	{"slice step cannot be zero", "ValueError"},
	{"unsupported operand type(s)", "TypeError"},
	{"not supported between instances of", "TypeError"},
	{"bad operand type for", "TypeError"},
	{"can only concatenate", "TypeError"},
	{"unhashable type", "TypeError"},
	{"has no len()", "TypeError"},
	{"object is not subscriptable", "TypeError"},
	{"object is not iterable", "TypeError"},
	{"object is not callable", "TypeError"},
	{"object is not an iterator", "TypeError"},
	{"is not async iterable", "TypeError"},
	{"is not an async iterator", "TypeError"},
	{"is not awaitable", "TypeError"},
	{"should return an integer", "TypeError"},
	{"attribute name must be string", "TypeError"},
	{"must be None or a string", "TypeError"},
	{"descriptor '", "TypeError"},
	{"() takes ", "TypeError"},
	{"() missing ", "TypeError"},
	{"() got an unexpected keyword argument", "TypeError"},
	{"() got multiple values for argument", "TypeError"},
	{"() requires at least", "TypeError"},
	{"() requires an", "TypeError"},
	{" expected at least ", "TypeError"},
	{" expected at most ", "TypeError"},
	{" expected 1 argument", "TypeError"},
	{" expected exactly ", "TypeError"},
	{"pop expected at least", "TypeError"},
}

// classifyUnprefixedError returns the exception class name implied by an
// unprefixed error message, or "" if the message isn't recognized.
func classifyUnprefixedError(msg string) string {
	for _, p := range unprefixedErrorPatterns {
		if strings.Contains(msg, p.substr) {
			return p.excName
		}
	}
	return ""
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
		if name := classifyUnprefixedError(errStr); name != "" {
			excClass, _ = vm.builtins[name].(*PyClass)
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

	// Extract just the message after the type prefix (e.g., "ValueError: msg" -> "msg")
	msg := errStr
	for _, ep := range exceptionPrefixes {
		if len(errStr) > len(ep.prefix)+2 && errStr[:len(ep.prefix)] == ep.prefix && errStr[len(ep.prefix):len(ep.prefix)+2] == ": " {
			msg = errStr[len(ep.prefix)+2:]
			break
		}
	}

	// Go code reports a missing key as "KeyError: 'key'"; store the bare key
	// so str(e) (repr of the key, for KeyError) does not quote it twice.
	if excClass != nil && excClass.Name == "KeyError" && len(msg) >= 2 &&
		(msg[0] == '\'' || msg[0] == '"') && msg[len(msg)-1] == msg[0] {
		msg = msg[1 : len(msg)-1]
	}

	return &PyException{
		ExcType: excClass,
		Args:    &PyTuple{Items: []Value{&PyString{Value: msg}}},
		Message: msg,
	}
}

// newKeyError creates a KeyError for key. Its single argument is the key
// itself, so str(e) shows repr(key) like CPython.
func (vm *VM) newKeyError(key Value) *PyException {
	return &PyException{
		ExcType:  vm.builtinClass("KeyError"),
		TypeName: "KeyError",
		Args:     &PyTuple{Items: []Value{key}},
		Message:  plainRepr(key),
	}
}

// pyExceptionList is a Value wrapper for []*PyException stored in ExceptionGroup instances
type pyExceptionList struct {
	items []*PyException
}

func (p *pyExceptionList) Type() string   { return "exception_list" }
func (p *pyExceptionList) String() string { return fmt.Sprintf("<exception_list: %d>", len(p.items)) }

// getEGExceptions retrieves the []*PyException list from an ExceptionGroup instance
func (vm *VM) getEGExceptions(inst *PyInstance) []*PyException {
	if el, ok := inst.Dict["__eg_exceptions__"].(*pyExceptionList); ok {
		return el.items
	}
	return nil
}

// isBaseExceptionGroup checks if a class has BaseExceptionGroup in its MRO
func (vm *VM) isBaseExceptionGroup(cls *PyClass) bool {
	beg, ok := vm.builtins["BaseExceptionGroup"].(*PyClass)
	if !ok {
		return false
	}
	for _, mroClass := range cls.Mro {
		if mroClass == beg {
			return true
		}
	}
	return false
}

// isOnlyBaseExceptionGroup checks if a class is BaseExceptionGroup but NOT ExceptionGroup
func (vm *VM) isOnlyBaseExceptionGroup(cls *PyClass) bool {
	eg, ok := vm.builtins["ExceptionGroup"].(*PyClass)
	if !ok {
		return true // no ExceptionGroup class; fall back to Base
	}
	for _, mroClass := range cls.Mro {
		if mroClass == eg {
			return false // it's an ExceptionGroup or subclass
		}
	}
	return vm.isBaseExceptionGroup(cls)
}

// isExceptionGroup checks if a PyException wraps an ExceptionGroup
func (vm *VM) isExceptionGroup(exc *PyException) bool {
	if exc.Instance != nil {
		return vm.isBaseExceptionGroup(exc.Instance.Class)
	}
	if exc.ExcType != nil {
		return vm.isBaseExceptionGroup(exc.ExcType)
	}
	return false
}

// buildExceptionGroup creates a new ExceptionGroup PyException
func (vm *VM) buildExceptionGroup(message string, exceptions []*PyException, isBase bool) (Value, error) {
	className := "ExceptionGroup"
	if isBase {
		className = "BaseExceptionGroup"
	}
	cls, ok := vm.builtins[className].(*PyClass)
	if !ok {
		return None, fmt.Errorf("RuntimeError: %s class not found", className)
	}

	// Build instance
	inst := &PyInstance{
		Class: cls,
		Dict:  make(map[string]Value),
	}
	tupleItems := make([]Value, len(exceptions))
	for i, e := range exceptions {
		tupleItems[i] = e
	}
	inst.Dict["message"] = &PyString{Value: message}
	inst.Dict["exceptions"] = &PyTuple{Items: tupleItems}
	inst.Dict["args"] = &PyTuple{Items: []Value{&PyString{Value: message}}}
	inst.Dict["__eg_exceptions__"] = &pyExceptionList{items: exceptions}

	// Create PyException wrapping this instance
	exc := &PyException{
		ExcType:  cls,
		Args:     &PyTuple{Items: []Value{&PyString{Value: message}}},
		Message:  message,
		Instance: inst,
	}
	return exc, nil
}

// compareChecked runs compareOp and returns any exception it raised as an
// error. compareOp reports errors through vm.currentException, which may
// already hold an exception pending for an enclosing finally body or
// __exit__ call; that pending exception is set aside for the comparison and
// restored afterwards, so it isn't mistaken for (or clobbered by) this one.
func (vm *VM) compareChecked(op Opcode, a, b Value) (Value, error) {
	pending := vm.currentException
	vm.currentException = nil
	result := vm.compareOp(op, a, b)
	if exc := vm.currentException; exc != nil {
		vm.currentException = nil
		return nil, exc
	}
	vm.currentException = pending
	return result, nil
}

// truthyChecked is truthy with the same error isolation as compareChecked,
// for __bool__/__len__ methods that raise.
func (vm *VM) truthyChecked(v Value) (bool, error) {
	pending := vm.currentException
	vm.currentException = nil
	result := vm.truthy(v)
	if exc := vm.currentException; exc != nil {
		vm.currentException = nil
		return false, exc
	}
	vm.currentException = pending
	return result, nil
}
