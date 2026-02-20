package runtime

import (
	"fmt"
	"sort"
	"strings"
)

// PyCode wraps a CodeObject for Python access via compile()
type PyCode struct {
	Code *CodeObject
}

func (c *PyCode) Type() string   { return "code" }
func (c *PyCode) String() string { return fmt.Sprintf("<code object %s at %p>", c.Code.Name, c) }

// =====================================
// Reflection Builtins
// =====================================

// BuiltinRepr implements repr(obj)
func BuiltinRepr(vm *VM) int {
	nargs := vm.GetTop()
	if nargs != 1 {
		vm.RaiseError("TypeError: repr() takes exactly one argument (%d given)", nargs)
		return 0
	}

	obj := vm.Get(1)
	result := vm.Repr(obj)
	vm.Push(NewString(result))
	return 1
}

// Repr returns the repr() string for a value
func (vm *VM) Repr(v Value) string {
	switch val := v.(type) {
	case *PyNone:
		return "None"
	case *PyBool:
		if val.Value {
			return "True"
		}
		return "False"
	case *PyInt:
		return fmt.Sprintf("%d", val.Value)
	case *PyFloat:
		return fmt.Sprintf("%g", val.Value)
	case *PyString:
		// repr() adds quotes around strings and escapes special chars
		return fmt.Sprintf("'%s'", escapeString(val.Value))
	case *PyBytes:
		return fmt.Sprintf("b'%s'", escapeBytes(val.Value))
	case *PyList:
		var items []string
		for _, item := range val.Items {
			items = append(items, vm.Repr(item))
		}
		return "[" + strings.Join(items, ", ") + "]"
	case *PyTuple:
		var items []string
		for _, item := range val.Items {
			items = append(items, vm.Repr(item))
		}
		if len(items) == 1 {
			return "(" + items[0] + ",)"
		}
		return "(" + strings.Join(items, ", ") + ")"
	case *PyDict:
		var items []string
		for k, v := range val.Items {
			items = append(items, vm.Repr(k)+": "+vm.Repr(v))
		}
		return "{" + strings.Join(items, ", ") + "}"
	case *PySet:
		if len(val.Items) == 0 {
			return "set()"
		}
		var items []string
		for k := range val.Items {
			items = append(items, vm.Repr(k))
		}
		return "{" + strings.Join(items, ", ") + "}"
	case *PyInstance:
		// Check for __repr__ method
		if reprMethod, err := vm.getAttr(val, "__repr__"); err == nil && reprMethod != nil {
			if result, err := vm.call(reprMethod, nil, nil); err == nil {
				if s, ok := result.(*PyString); ok {
					return s.Value
				}
			}
		}
		return fmt.Sprintf("<%s object at %p>", val.Class.Name, val)
	case *PyClass:
		return fmt.Sprintf("<class '%s'>", val.Name)
	case *PyFunction:
		return fmt.Sprintf("<function %s at %p>", val.Name, val)
	case *PyBuiltinFunc:
		return fmt.Sprintf("<built-in function %s>", val.Name)
	case *PyGoFunc:
		return fmt.Sprintf("<built-in function %s>", val.Name)
	case *PyModule:
		return fmt.Sprintf("<module '%s'>", val.Name)
	case *PyCode:
		return fmt.Sprintf("<code object %s at %p>", val.Code.Name, val)
	default:
		return fmt.Sprintf("<%s object at %p>", vm.typeName(v), v)
	}
}

// escapeString escapes special characters for repr()
func escapeString(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString("\\\\")
		case '\'':
			b.WriteString("\\'")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		default:
			if r < 32 || r > 126 {
				b.WriteString(fmt.Sprintf("\\x%02x", r))
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// escapeBytes escapes bytes for repr()
func escapeBytes(data []byte) string {
	var b strings.Builder
	for _, c := range data {
		switch c {
		case '\\':
			b.WriteString("\\\\")
		case '\'':
			b.WriteString("\\'")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		default:
			if c < 32 || c > 126 {
				b.WriteString(fmt.Sprintf("\\x%02x", c))
			} else {
				b.WriteByte(c)
			}
		}
	}
	return b.String()
}

// BuiltinDir implements dir([obj])
func BuiltinDir(vm *VM) int {
	nargs := vm.GetTop()

	if nargs == 0 {
		// dir() with no args - return names in current scope
		names := make(map[string]bool)

		// Find the caller's frame
		callerFrame := vm.getCallerFrame()
		if callerFrame != nil {
			// Add local variable names
			if callerFrame.Code != nil {
				for _, name := range callerFrame.Code.VarNames {
					names[name] = true
				}
			}
			// Add globals
			for name := range callerFrame.Globals {
				names[name] = true
			}
			// Add builtins
			for name := range callerFrame.Builtins {
				names[name] = true
			}
		}
		// Also add top-level globals
		for name := range vm.Globals {
			names[name] = true
		}

		result := sortedStringList(names)
		vm.Push(&PyList{Items: result})
		return 1
	}

	if nargs != 1 {
		vm.RaiseError("TypeError: dir() takes at most 1 argument (%d given)", nargs)
		return 0
	}

	// dir(obj) - return attributes of object
	obj := vm.Get(1)
	names := vm.getObjectDir(obj)
	vm.Push(&PyList{Items: names})
	return 1
}

// getObjectDir returns the attributes of an object for dir()
func (vm *VM) getObjectDir(obj Value) []Value {
	names := make(map[string]bool)

	switch v := obj.(type) {
	case *PyInstance:
		// Instance attributes
		for name := range v.Dict {
			names[name] = true
		}
		// Class attributes
		for name := range v.Class.Dict {
			names[name] = true
		}
		// Walk MRO for inherited attributes
		for _, cls := range v.Class.Mro {
			for name := range cls.Dict {
				names[name] = true
			}
		}
	case *PyClass:
		for name := range v.Dict {
			names[name] = true
		}
		for _, cls := range v.Mro {
			for name := range cls.Dict {
				names[name] = true
			}
		}
	case *PyModule:
		for name := range v.Dict {
			names[name] = true
		}
	case *PyDict:
		// For dicts, return the keys that are strings
		for k := range v.Items {
			if s, ok := k.(*PyString); ok {
				names[s.Value] = true
			}
		}
		for _, name := range []string{"clear", "copy", "fromkeys", "get", "items", "keys", "pop", "popitem", "setdefault", "update", "values"} {
			names[name] = true
		}
	case *PyList:
		// List methods
		for _, name := range []string{"append", "clear", "copy", "count", "extend", "index", "insert", "pop", "remove", "reverse", "sort"} {
			names[name] = true
		}
	case *PyString:
		// String methods
		for _, name := range []string{"capitalize", "casefold", "center", "count", "encode", "endswith", "expandtabs", "find", "format", "format_map", "index", "isalnum", "isalpha", "isascii", "isdecimal", "isdigit", "isidentifier", "islower", "isnumeric", "isprintable", "isspace", "istitle", "isupper", "join", "ljust", "lower", "lstrip", "maketrans", "partition", "removeprefix", "removesuffix", "replace", "rfind", "rindex", "rjust", "rpartition", "rsplit", "rstrip", "split", "splitlines", "startswith", "strip", "swapcase", "title", "translate", "upper", "zfill"} {
			names[name] = true
		}
	case *PyInt:
		for _, name := range []string{"bit_length", "bit_count", "conjugate", "as_integer_ratio", "to_bytes", "from_bytes", "real", "imag", "numerator", "denominator"} {
			names[name] = true
		}
	case *PyFloat:
		for _, name := range []string{"is_integer", "hex", "fromhex", "as_integer_ratio", "conjugate", "real", "imag"} {
			names[name] = true
		}
	case *PySet:
		for _, name := range []string{"add", "clear", "copy", "difference", "difference_update", "discard", "intersection", "intersection_update", "isdisjoint", "issubset", "issuperset", "pop", "remove", "symmetric_difference", "symmetric_difference_update", "union", "update"} {
			names[name] = true
		}
	case *PyFrozenSet:
		for _, name := range []string{"copy", "difference", "intersection", "isdisjoint", "issubset", "issuperset", "symmetric_difference", "union"} {
			names[name] = true
		}
	case *PyTuple:
		for _, name := range []string{"count", "index"} {
			names[name] = true
		}
	case *PyBytes:
		for _, name := range []string{"capitalize", "center", "count", "decode", "endswith", "expandtabs", "find", "hex", "index", "isalnum", "isalpha", "isascii", "isdigit", "islower", "isspace", "istitle", "isupper", "join", "ljust", "lower", "lstrip", "maketrans", "partition", "removeprefix", "removesuffix", "replace", "rfind", "rindex", "rjust", "rpartition", "rsplit", "rstrip", "split", "splitlines", "startswith", "strip", "swapcase", "title", "translate", "upper", "zfill"} {
			names[name] = true
		}
	case *PyRange:
		for _, name := range []string{"count", "index", "start", "stop", "step"} {
			names[name] = true
		}
	case *PyComplex:
		for _, name := range []string{"conjugate", "imag", "real"} {
			names[name] = true
		}
	}

	return sortedStringList(names)
}

// sortedStringList converts a map of strings to a sorted list of PyStrings
func sortedStringList(names map[string]bool) []Value {
	sorted := make([]string, 0, len(names))
	for name := range names {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)

	result := make([]Value, len(sorted))
	for i, name := range sorted {
		result[i] = NewString(name)
	}
	return result
}

// BuiltinGlobals implements globals()
func BuiltinGlobals(vm *VM) int {
	nargs := vm.GetTop()
	if nargs != 0 {
		vm.RaiseError("TypeError: globals() takes no arguments (%d given)", nargs)
		return 0
	}

	globalsDict := &PyDict{Items: make(map[Value]Value)}

	// Find the caller's frame to get the right globals
	callerFrame := vm.getCallerFrame()
	if callerFrame != nil && callerFrame.Globals != nil {
		for name, value := range callerFrame.Globals {
			globalsDict.Items[NewString(name)] = value
		}
	} else {
		// Fallback to vm.Globals
		for name, value := range vm.Globals {
			globalsDict.Items[NewString(name)] = value
		}
	}

	vm.Push(globalsDict)
	return 1
}

// BuiltinLocals implements locals()
func BuiltinLocals(vm *VM) int {
	nargs := vm.GetTop()
	if nargs != 0 {
		vm.RaiseError("TypeError: locals() takes no arguments (%d given)", nargs)
		return 0
	}

	locals := &PyDict{Items: make(map[Value]Value)}

	// Find the caller's frame (skip the builtin's temporary frame)
	callerFrame := vm.getCallerFrame()
	if callerFrame != nil && callerFrame.Code != nil {
		for i, name := range callerFrame.Code.VarNames {
			if i < len(callerFrame.Locals) && callerFrame.Locals[i] != nil {
				locals.Items[NewString(name)] = callerFrame.Locals[i]
			}
		}
	}

	vm.Push(locals)
	return 1
}

// getCallerFrame returns the frame of the Python code that called the current builtin
func (vm *VM) getCallerFrame() *Frame {
	// If there are frames on the call stack, find the first one with Code set
	// (skip temporary frames created for Go function calls)
	for i := len(vm.frames) - 1; i >= 0; i-- {
		if vm.frames[i].Code != nil {
			return vm.frames[i]
		}
	}
	// Fall back to current frame if it has Code
	if vm.frame != nil && vm.frame.Code != nil {
		return vm.frame
	}
	return nil
}

// BuiltinVars implements vars([obj])
func BuiltinVars(vm *VM) int {
	nargs := vm.GetTop()

	if nargs == 0 {
		// vars() with no args is equivalent to locals()
		return BuiltinLocals(vm)
	}

	if nargs != 1 {
		vm.RaiseError("TypeError: vars() takes at most 1 argument (%d given)", nargs)
		return 0
	}

	obj := vm.Get(1)

	// Try to get __dict__ from the object
	switch v := obj.(type) {
	case *PyInstance:
		result := &PyDict{Items: make(map[Value]Value)}
		for name, value := range v.Dict {
			result.Items[NewString(name)] = value
		}
		vm.Push(result)
		return 1
	case *PyClass:
		result := &PyDict{Items: make(map[Value]Value)}
		for name, value := range v.Dict {
			result.Items[NewString(name)] = value
		}
		vm.Push(result)
		return 1
	case *PyModule:
		result := &PyDict{Items: make(map[Value]Value)}
		for name, value := range v.Dict {
			result.Items[NewString(name)] = value
		}
		vm.Push(result)
		return 1
	default:
		vm.RaiseError("TypeError: vars() argument must have __dict__ attribute")
		return 0
	}
}

// =====================================
// Execution Builtins
// =====================================

// BuiltinCompile implements compile(source, filename, mode)
func BuiltinCompile(vm *VM) int {
	if CompileFunc == nil {
		vm.RaiseError("RuntimeError: compile() not available - compiler not registered")
		return 0
	}

	nargs := vm.GetTop()
	if nargs < 3 {
		vm.RaiseError("TypeError: compile() requires at least 3 arguments: source, filename, mode")
		return 0
	}

	sourceArg := vm.Get(1)
	source, ok := sourceArg.(*PyString)
	if !ok {
		vm.RaiseError("TypeError: compile() expected string for source, got %s", vm.typeName(sourceArg))
		return 0
	}

	filenameArg := vm.Get(2)
	filename, ok := filenameArg.(*PyString)
	if !ok {
		vm.RaiseError("TypeError: compile() expected string for filename, got %s", vm.typeName(filenameArg))
		return 0
	}

	modeArg := vm.Get(3)
	mode, ok := modeArg.(*PyString)
	if !ok {
		vm.RaiseError("TypeError: compile() expected string for mode, got %s", vm.typeName(modeArg))
		return 0
	}

	// Validate mode
	if mode.Value != "exec" && mode.Value != "eval" && mode.Value != "single" {
		vm.RaiseError("ValueError: compile() mode must be 'exec', 'eval', or 'single'")
		return 0
	}

	code, err := CompileFunc(source.Value, filename.Value, mode.Value)
	if err != nil {
		vm.RaiseError("SyntaxError: %s", err.Error())
		return 0
	}

	vm.Push(&PyCode{Code: code})
	return 1
}

// BuiltinExec implements exec(code, globals=None, locals=None)
func BuiltinExec(vm *VM) int {
	nargs := vm.GetTop()
	if nargs < 1 {
		vm.RaiseError("TypeError: exec() missing required argument: 'source'")
		return 0
	}

	codeArg := vm.Get(1)

	// Get the caller's frame for default globals/locals
	callerFrame := vm.getCallerFrame()

	// Track the original PyDicts so we can copy changes back
	var originalGlobals *PyDict
	var originalLocals *PyDict

	// Determine if we should use caller's globals directly
	var globalsDict map[string]Value

	if nargs >= 2 {
		arg2 := vm.Get(2)
		if _, isNone := arg2.(*PyNone); !isNone {
			if g, ok := arg2.(*PyDict); ok {
				originalGlobals = g
				globalsDict = dictToStringMap(g)
			} else {
				vm.RaiseError("TypeError: exec() globals must be a dict")
				return 0
			}
		}
	}
	if globalsDict == nil {
		if callerFrame != nil && callerFrame.Globals != nil {
			globalsDict = callerFrame.Globals
		} else {
			globalsDict = vm.Globals
		}
	}

	// Get locals dict (arg 3) or use caller's locals
	var localsDict map[string]Value
	if nargs >= 3 {
		arg3 := vm.Get(3)
		if _, isNone := arg3.(*PyNone); !isNone {
			if l, ok := arg3.(*PyDict); ok {
				originalLocals = l
				localsDict = dictToStringMap(l)
			} else {
				vm.RaiseError("TypeError: exec() locals must be a dict")
				return 0
			}
		}
	}
	if localsDict == nil {
		// Build locals from caller's frame if available
		if callerFrame != nil && callerFrame.Code != nil {
			localsDict = make(map[string]Value)
			// Copy globals first
			for k, v := range globalsDict {
				localsDict[k] = v
			}
			// Then add/override with actual local variables
			for i, name := range callerFrame.Code.VarNames {
				if i < len(callerFrame.Locals) && callerFrame.Locals[i] != nil {
					localsDict[name] = callerFrame.Locals[i]
				}
			}
		} else {
			localsDict = globalsDict
		}
		// If no explicit locals, changes go to globals (or original globals dict)
		if originalGlobals != nil {
			originalLocals = originalGlobals
		}
	}

	// Get or compile the code
	var code *CodeObject
	switch c := codeArg.(type) {
	case *PyCode:
		code = c.Code
	case *PyString:
		if CompileFunc == nil {
			vm.RaiseError("RuntimeError: exec() cannot compile - compiler not registered")
			return 0
		}
		var err error
		code, err = CompileFunc(c.Value, "<string>", "exec")
		if err != nil {
			vm.RaiseError("SyntaxError: %s", err.Error())
			return 0
		}
	default:
		vm.RaiseError("TypeError: exec() arg 1 must be a string or code object, not %s", vm.typeName(codeArg))
		return 0
	}

	// Execute with the specified namespace
	err := vm.ExecuteInNamespace(code, globalsDict, localsDict)
	if err != nil {
		vm.RaiseError("%s", err.Error())
		return 0
	}

	// Copy changes back to original PyDicts
	if originalGlobals != nil {
		for k, v := range globalsDict {
			originalGlobals.Items[NewString(k)] = v
		}
	}
	if originalLocals != nil && originalLocals != originalGlobals {
		for k, v := range localsDict {
			originalLocals.Items[NewString(k)] = v
		}
	}

	vm.Push(None)
	return 1
}

// BuiltinEval implements eval(expression, globals=None, locals=None)
func BuiltinEval(vm *VM) int {
	nargs := vm.GetTop()
	if nargs < 1 {
		vm.RaiseError("TypeError: eval() missing required argument: 'expression'")
		return 0
	}

	exprArg := vm.Get(1)

	// Get the caller's frame for default globals/locals
	callerFrame := vm.getCallerFrame()

	// Get globals dict (arg 2) or use caller's globals
	var globalsDict map[string]Value
	if nargs >= 2 {
		arg2 := vm.Get(2)
		if _, isNone := arg2.(*PyNone); !isNone {
			if g, ok := arg2.(*PyDict); ok {
				globalsDict = dictToStringMap(g)
			} else {
				vm.RaiseError("TypeError: eval() globals must be a dict")
				return 0
			}
		}
	}
	if globalsDict == nil {
		if callerFrame != nil && callerFrame.Globals != nil {
			globalsDict = callerFrame.Globals
		} else {
			globalsDict = vm.Globals
		}
	}

	// Get locals dict (arg 3) or use caller's locals
	var localsDict map[string]Value
	if nargs >= 3 {
		arg3 := vm.Get(3)
		if _, isNone := arg3.(*PyNone); !isNone {
			if l, ok := arg3.(*PyDict); ok {
				localsDict = dictToStringMap(l)
			} else {
				vm.RaiseError("TypeError: eval() locals must be a dict")
				return 0
			}
		}
	}
	if localsDict == nil {
		// Build locals from caller's frame if available
		if callerFrame != nil && callerFrame.Code != nil {
			localsDict = make(map[string]Value)
			// Copy globals first (like Python's eval does)
			for k, v := range globalsDict {
				localsDict[k] = v
			}
			// Then add/override with actual local variables
			for i, name := range callerFrame.Code.VarNames {
				if i < len(callerFrame.Locals) && callerFrame.Locals[i] != nil {
					localsDict[name] = callerFrame.Locals[i]
				}
			}
		} else {
			localsDict = globalsDict
		}
	}

	// Get or compile the code
	var code *CodeObject
	switch c := exprArg.(type) {
	case *PyCode:
		code = c.Code
	case *PyString:
		if CompileFunc == nil {
			vm.RaiseError("RuntimeError: eval() cannot compile - compiler not registered")
			return 0
		}
		var err error
		code, err = CompileFunc(c.Value, "<string>", "eval")
		if err != nil {
			vm.RaiseError("SyntaxError: %s", err.Error())
			return 0
		}
	default:
		vm.RaiseError("TypeError: eval() arg 1 must be a string or code object, not %s", vm.typeName(exprArg))
		return 0
	}

	// Execute and return the result
	result, err := vm.EvalInNamespace(code, globalsDict, localsDict)
	if err != nil {
		vm.RaiseError("%s", err.Error())
		return 0
	}

	if result == nil {
		result = None
	}
	vm.Push(result)
	return 1
}

// =====================================
// Helper Functions
// =====================================

// dictToStringMap converts a PyDict to a map[string]Value
func dictToStringMap(d *PyDict) map[string]Value {
	result := make(map[string]Value)
	for k, v := range d.Items {
		if s, ok := k.(*PyString); ok {
			result[s.Value] = v
		}
	}
	return result
}

// ExecuteInNamespace executes code with custom globals/locals
func (vm *VM) ExecuteInNamespace(code *CodeObject, globals, locals map[string]Value) error {
	// For exec, we need to merge locals into the globals namespace because
	// the compiled code uses OpLoadGlobal for all name lookups (since it doesn't
	// know about the caller's local variables at compile time).
	mergedGlobals := make(map[string]Value)
	for k, v := range globals {
		mergedGlobals[k] = v
	}
	// Locals override globals (Python's LEGB rule)
	for k, v := range locals {
		mergedGlobals[k] = v
	}

	// Create a new frame for execution
	frame := &Frame{
		Code:     code,
		IP:       0,
		Stack:    make([]Value, code.StackSize+16),
		SP:       0,
		Locals:   make([]Value, len(code.VarNames)),
		Globals:  mergedGlobals,
		Builtins: vm.builtins,
	}

	// Initialize locals from the provided dict
	for i, name := range code.VarNames {
		if val, ok := locals[name]; ok {
			frame.Locals[i] = val
		}
	}

	// Save current state
	oldFrame := vm.frame
	oldFrames := vm.frames

	// Set up new frame
	vm.frames = []*Frame{frame}
	vm.frame = frame

	// Execute
	_, err := vm.run()

	// Update locals dict with any changes from frame.Locals
	for i, name := range code.VarNames {
		if i < len(frame.Locals) && frame.Locals[i] != nil {
			locals[name] = frame.Locals[i]
		}
	}

	// Copy new global assignments back to the original globals dict
	// This handles cases like `exec("x = 100")` where x is stored via OpStoreGlobal
	for k, v := range mergedGlobals {
		globals[k] = v
	}

	// Restore state
	vm.frame = oldFrame
	vm.frames = oldFrames

	return err
}

// EvalInNamespace evaluates code and returns the result
func (vm *VM) EvalInNamespace(code *CodeObject, globals, locals map[string]Value) (Value, error) {
	// For eval, we need to merge locals into the globals namespace because
	// the compiled code uses OpLoadGlobal for all name lookups (since it doesn't
	// know about the caller's local variables at compile time).
	mergedGlobals := make(map[string]Value)
	for k, v := range globals {
		mergedGlobals[k] = v
	}
	// Locals override globals (Python's LEGB rule: Local > Enclosing > Global > Builtin)
	for k, v := range locals {
		mergedGlobals[k] = v
	}

	// Create a new frame for execution
	frame := &Frame{
		Code:     code,
		IP:       0,
		Stack:    make([]Value, code.StackSize+16),
		SP:       0,
		Locals:   make([]Value, len(code.VarNames)),
		Globals:  mergedGlobals,
		Builtins: vm.builtins,
	}

	// Initialize locals from the provided dict
	for i, name := range code.VarNames {
		if val, ok := locals[name]; ok {
			frame.Locals[i] = val
		}
	}

	// Save current state
	oldFrame := vm.frame
	oldFrames := vm.frames

	// Set up new frame
	vm.frames = []*Frame{frame}
	vm.frame = frame

	// Execute
	result, err := vm.run()

	// Restore state
	vm.frame = oldFrame
	vm.frames = oldFrames

	if err != nil {
		return nil, err
	}

	// Check for the eval result variable (set by compileForBuiltin for eval mode)
	// It will be in mergedGlobals since that's what the frame used
	if evalResult, ok := mergedGlobals["__eval_result__"]; ok {
		// Clean up the temp variable
		delete(mergedGlobals, "__eval_result__")
		return evalResult, nil
	}

	// Check if there's a value left on the stack (expression result)
	if frame.SP > 0 {
		result = frame.Stack[frame.SP-1]
	}

	return result, nil
}
