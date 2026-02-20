package runtime

import (
	"fmt"
	"reflect"
)

// GoFunction is the signature for Go functions callable from Python.
// Similar to gopher-lua's LGFunction.
// The function receives the VM and returns the number of values pushed onto the stack.
type GoFunction func(vm *VM) int

// PyGoFunc wraps a Go function for use in Python
type PyGoFunc struct {
	Name string
	Fn   GoFunction
}

func (g *PyGoFunc) Type() string   { return "builtin_function_or_method" }
func (g *PyGoFunc) String() string { return fmt.Sprintf("<go function %s>", g.Name) }

// PyUserData wraps arbitrary Go values for use in Python.
// Similar to gopher-lua's LUserData.
type PyUserData struct {
	Value     any
	Metatable *PyDict
}

func (u *PyUserData) Type() string   { return "userdata" }
func (u *PyUserData) String() string { return fmt.Sprintf("<userdata %T>", u.Value) }

// =====================================
// Value Conversion Methods (Go -> Python)
// =====================================

// NewInt creates a Python int from a Go int64
func NewInt(v int64) *PyInt {
	return &PyInt{Value: v}
}

// NewFloat creates a Python float from a Go float64
func NewFloat(v float64) *PyFloat {
	return &PyFloat{Value: v}
}

// NewComplex creates a Python complex from two Go float64s (real, imag)
func NewComplex(real, imag float64) *PyComplex {
	return MakeComplex(real, imag)
}

// NewString creates a Python str from a Go string, using interning for short strings
func NewString(v string) *PyString {
	return InternString(v)
}

// NewBool creates a Python bool from a Go bool
func NewBool(v bool) *PyBool {
	if v {
		return True
	}
	return False
}

// NewList creates a Python list from a slice of Values
func NewList(items []Value) *PyList {
	return &PyList{Items: items}
}

// NewTuple creates a Python tuple from a slice of Values
func NewTuple(items []Value) *PyTuple {
	return &PyTuple{Items: items}
}

// NewDict creates a Python dict
func NewDict() *PyDict {
	return &PyDict{Items: make(map[Value]Value)}
}

// NewBytes creates a Python bytes from a Go byte slice
func NewBytes(v []byte) *PyBytes {
	return &PyBytes{Value: v}
}

// NewUserData creates a new userdata wrapping a Go value
func NewUserData(v any) *PyUserData {
	return &PyUserData{Value: v}
}

// NewGoFunction creates a Python-callable function from a Go function
func NewGoFunction(name string, fn GoFunction) *PyGoFunc {
	return &PyGoFunc{Name: name, Fn: fn}
}

// =====================================
// VM API Methods (gopher-lua style)
// =====================================

// Push pushes a value onto the stack
func (vm *VM) Push(v Value) {
	vm.push(v)
}

// Pop pops and returns a value from the stack
func (vm *VM) Pop() Value {
	return vm.pop()
}

// GetTop returns the index of the top element (number of elements on stack)
func (vm *VM) GetTop() int {
	return vm.frame.SP
}

// SetTop sets the stack top to a specific index
func (vm *VM) SetTop(n int) {
	if n < 0 {
		vm.frame.SP = vm.frame.SP + n + 1
		if vm.frame.SP < 0 {
			vm.frame.SP = 0
		}
	} else {
		vm.frame.SP = n
	}
}

// Get returns the value at the given stack index (1-based, like Lua)
// Negative indices count from top
func (vm *VM) Get(idx int) Value {
	if idx > 0 {
		// Positive index: from bottom (1-based)
		if idx <= vm.frame.SP {
			return vm.frame.Stack[idx-1]
		}
	} else if idx < 0 {
		// Negative index: from top
		realIdx := vm.frame.SP + idx
		if realIdx >= 0 {
			return vm.frame.Stack[realIdx]
		}
	}
	return None
}

// =====================================
// Argument Validation Helpers
// =====================================

// RequireArgs checks that at least min arguments were passed.
// If not, it raises a TypeError with the function name and returns false.
// Usage:
//
//	if !vm.RequireArgs("loads", 1) { return 0 }
func (vm *VM) RequireArgs(name string, min int) bool {
	if vm.GetTop() >= min {
		return true
	}
	panic(&PyPanicError{
		ExcType: "TypeError",
		Message: fmt.Sprintf("%s() requires at least %d argument(s), got %d", name, min, vm.GetTop()),
	})
}

// OptionalInt returns the int at stack position pos if present and not None,
// otherwise returns def.
func (vm *VM) OptionalInt(pos int, def int64) int64 {
	if vm.GetTop() >= pos {
		v := vm.Get(pos)
		if _, ok := v.(*PyNone); !ok {
			return vm.toInt(v)
		}
	}
	return def
}

// OptionalFloat returns the float at stack position pos if present and not None,
// otherwise returns def.
func (vm *VM) OptionalFloat(pos int, def float64) float64 {
	if vm.GetTop() >= pos {
		v := vm.Get(pos)
		if _, ok := v.(*PyNone); !ok {
			return vm.toFloat(v)
		}
	}
	return def
}

// OptionalString returns the string at stack position pos if present and not None,
// otherwise returns def.
func (vm *VM) OptionalString(pos int, def string) string {
	if vm.GetTop() >= pos {
		v := vm.Get(pos)
		if _, ok := v.(*PyNone); !ok {
			if s, ok := v.(*PyString); ok {
				return s.Value
			}
			return vm.str(v)
		}
	}
	return def
}

// OptionalBool returns the bool at stack position pos if present and not None,
// otherwise returns def.
func (vm *VM) OptionalBool(pos int, def bool) bool {
	if vm.GetTop() >= pos {
		v := vm.Get(pos)
		if _, ok := v.(*PyNone); !ok {
			return vm.truthy(v)
		}
	}
	return def
}

// =====================================
// Type Checking Methods
// =====================================

// CheckInt gets argument n as int64, raises error if not convertible
func (vm *VM) CheckInt(n int) int64 {
	v := vm.Get(n)
	return vm.toInt(v)
}

// CheckFloat gets argument n as float64, raises error if not convertible
func (vm *VM) CheckFloat(n int) float64 {
	v := vm.Get(n)
	return vm.toFloat(v)
}

// CheckString gets argument n as string, raises error if not a string
func (vm *VM) CheckString(n int) string {
	v := vm.Get(n)
	if s, ok := v.(*PyString); ok {
		return s.Value
	}
	return vm.str(v)
}

// CheckBool gets argument n as bool
func (vm *VM) CheckBool(n int) bool {
	v := vm.Get(n)
	return vm.truthy(v)
}

// CheckList gets argument n as a list
func (vm *VM) CheckList(n int) *PyList {
	v := vm.Get(n)
	if l, ok := v.(*PyList); ok {
		return l
	}
	return nil
}

// CheckDict gets argument n as a dict
func (vm *VM) CheckDict(n int) *PyDict {
	v := vm.Get(n)
	if d, ok := v.(*PyDict); ok {
		return d
	}
	return nil
}

// CheckUserData gets argument n as userdata, checking the metatable name
func (vm *VM) CheckUserData(n int, typeName string) *PyUserData {
	v := vm.Get(n)
	if ud, ok := v.(*PyUserData); ok {
		// Optionally check metatable type
		return ud
	}
	return nil
}

// ToInt converts argument n to int64, returns 0 if not convertible
func (vm *VM) ToInt(n int) int64 {
	return vm.toInt(vm.Get(n))
}

// ToFloat converts argument n to float64
func (vm *VM) ToFloat(n int) float64 {
	return vm.toFloat(vm.Get(n))
}

// ToString converts argument n to string
func (vm *VM) ToString(n int) string {
	return vm.str(vm.Get(n))
}

// ToBool converts argument n to bool
func (vm *VM) ToBool(n int) bool {
	return vm.truthy(vm.Get(n))
}

// ToUserData gets argument n as userdata
func (vm *VM) ToUserData(n int) *PyUserData {
	v := vm.Get(n)
	if ud, ok := v.(*PyUserData); ok {
		return ud
	}
	return nil
}

// =====================================
// Global Variable Access
// =====================================

// SetGlobal sets a global variable
func (vm *VM) SetGlobal(name string, v Value) {
	vm.Globals[name] = v
}

// GetGlobal gets a global variable
func (vm *VM) GetGlobal(name string) Value {
	if v, ok := vm.Globals[name]; ok {
		return v
	}
	return None
}

// SetBuiltin registers a builtin function
func (vm *VM) SetBuiltin(name string, v Value) {
	vm.builtins[name] = v
}

// GetBuiltin gets a builtin
func (vm *VM) GetBuiltin(name string) Value {
	if v, ok := vm.builtins[name]; ok {
		return v
	}
	return None
}

// =====================================
// Function Registration Helpers
// =====================================

// Register registers a Go function as a global
func (vm *VM) Register(name string, fn GoFunction) {
	vm.SetGlobal(name, NewGoFunction(name, fn))
}

// RegisterBuiltin registers a Go function as a builtin
func (vm *VM) RegisterBuiltin(name string, fn GoFunction) {
	vm.SetBuiltin(name, NewGoFunction(name, fn))
}

// RegisterFuncs registers multiple functions at once
func (vm *VM) RegisterFuncs(funcs map[string]GoFunction) {
	for name, fn := range funcs {
		vm.Register(name, fn)
	}
}

// =====================================
// Type Registration (gopher-lua style metatables)
// =====================================

// TypeMetatable holds methods and properties for a user-defined type
type TypeMetatable struct {
	Name       string
	Methods    map[string]GoFunction
	Properties map[string]GoFunction // Getters that are called automatically (like Python @property)
}

var typeMetatables = make(map[string]*TypeMetatable)

// NewTypeMetatable creates a new type metatable
func (vm *VM) NewTypeMetatable(typeName string) *TypeMetatable {
	mt := &TypeMetatable{
		Name:    typeName,
		Methods: make(map[string]GoFunction),
	}
	typeMetatables[typeName] = mt
	return mt
}

// GetTypeMetatable retrieves a registered type metatable
func (vm *VM) GetTypeMetatable(typeName string) *TypeMetatable {
	return typeMetatables[typeName]
}

// RegisterTypeMetatable registers a type metatable globally (without VM instance)
func RegisterTypeMetatable(typeName string, mt *TypeMetatable) {
	typeMetatables[typeName] = mt
}

// GetRegisteredTypeMetatable retrieves a type metatable globally (without VM instance)
func GetRegisteredTypeMetatable(typeName string) *TypeMetatable {
	return typeMetatables[typeName]
}

// SetMethod sets a method on a type metatable
func (mt *TypeMetatable) SetMethod(name string, fn GoFunction) {
	mt.Methods[name] = fn
}

// SetMethods sets multiple methods on a type metatable
func (mt *TypeMetatable) SetMethods(methods map[string]GoFunction) {
	for name, fn := range methods {
		mt.Methods[name] = fn
	}
}

// ResetTypeMetatables clears the type metatable registry (called by ResetModules)
func ResetTypeMetatables() {
	typeMetatables = make(map[string]*TypeMetatable)
}

// =====================================
// Pending Builtins Registry
// =====================================

// pendingBuiltins holds builtins to be registered on new VMs.
// This allows stdlib modules (initialized before VM creation) to register builtins.
var pendingBuiltins = make(map[string]GoFunction)

// RegisterPendingBuiltin registers a builtin to be added to new VMs.
// Use this from stdlib modules that need to add builtin functions.
func RegisterPendingBuiltin(name string, fn GoFunction) {
	pendingBuiltins[name] = fn
}

// GetPendingBuiltins returns all pending builtins (called by NewVM)
func GetPendingBuiltins() map[string]GoFunction {
	return pendingBuiltins
}

// ResetPendingBuiltins clears the pending builtins registry (called by ResetModules)
func ResetPendingBuiltins() {
	pendingBuiltins = make(map[string]GoFunction)
}

// ApplyPendingBuiltins applies any pending builtins to an existing VM.
// This is useful when enabling modules after VM creation.
func (vm *VM) ApplyPendingBuiltins() {
	for name, fn := range pendingBuiltins {
		if _, exists := vm.builtins[name]; !exists {
			vm.builtins[name] = NewGoFunction(name, fn)
		}
	}
}

// RegisterType is a convenience function to register a Go type with methods
// Usage:
//
//	vm.RegisterType("person", map[string]GoFunction{
//	    "new": newPerson,
//	    "name": personGetName,
//	    "set_name": personSetName,
//	})
func (vm *VM) RegisterType(typeName string, constructor GoFunction, methods map[string]GoFunction) {
	mt := vm.NewTypeMetatable(typeName)
	mt.SetMethods(methods)

	// Create the type table
	typeTable := NewDict()

	// Add constructor as "new" or as callable
	if constructor != nil {
		vm.SetGlobal(typeName, NewGoFunction(typeName, constructor))
	}

	// Store metatable reference
	typeTable.Items[NewString("__name__")] = NewString(typeName)
	vm.SetGlobal("__"+typeName+"_mt__", typeTable)
}

// =====================================
// UserData with Metatable Support
// =====================================

// NewUserDataWithMeta creates userdata with a metatable attached
func (vm *VM) NewUserDataWithMeta(v any, typeName string) *PyUserData {
	ud := &PyUserData{Value: v}

	// Create metatable dict with method lookups
	mt := vm.GetTypeMetatable(typeName)
	if mt != nil {
		ud.Metatable = NewDict()
		ud.Metatable.Items[NewString("__type__")] = NewString(typeName)
	}

	return ud
}

// =====================================
// Value Type Helpers
// =====================================

// IsNone checks if a value is None
func IsNone(v Value) bool {
	_, ok := v.(*PyNone)
	return ok
}

// IsInt checks if a value is an int
func IsInt(v Value) bool {
	_, ok := v.(*PyInt)
	return ok
}

// IsFloat checks if a value is a float
func IsFloat(v Value) bool {
	_, ok := v.(*PyFloat)
	return ok
}

// IsString checks if a value is a string
func IsString(v Value) bool {
	_, ok := v.(*PyString)
	return ok
}

// IsBool checks if a value is a bool
func IsBool(v Value) bool {
	_, ok := v.(*PyBool)
	return ok
}

// IsList checks if a value is a list
func IsList(v Value) bool {
	_, ok := v.(*PyList)
	return ok
}

// IsDict checks if a value is a dict
func IsDict(v Value) bool {
	_, ok := v.(*PyDict)
	return ok
}

// IsTuple checks if a value is a tuple
func IsTuple(v Value) bool {
	_, ok := v.(*PyTuple)
	return ok
}

// IsUserData checks if a value is userdata
func IsUserData(v Value) bool {
	_, ok := v.(*PyUserData)
	return ok
}

// IsCallable checks if a value is callable
func IsCallable(v Value) bool {
	switch val := v.(type) {
	case *PyFunction, *PyBuiltinFunc, *PyGoFunc, *PyClass, *PyMethod:
		return true
	case *PyUserData:
		// Check for __call__ method in metatable
		if val.Metatable != nil {
			var typeName string
			for k, v := range val.Metatable.Items {
				if ks, ok := k.(*PyString); ok && ks.Value == "__type__" {
					if ts, ok := v.(*PyString); ok {
						typeName = ts.Value
					}
					break
				}
			}
			if typeName != "" {
				if mt := GetRegisteredTypeMetatable(typeName); mt != nil {
					if _, ok := mt.Methods["__call__"]; ok {
						return true
					}
				}
			}
		}
		return false
	default:
		return false
	}
}

// =====================================
// Go Value Conversion
// =====================================

// ToGoValue converts a Python value to a Go value
func ToGoValue(v Value) any {
	switch val := v.(type) {
	case *PyNone:
		return nil
	case *PyBool:
		return val.Value
	case *PyInt:
		return val.Value
	case *PyFloat:
		return val.Value
	case *PyString:
		return val.Value
	case *PyBytes:
		return val.Value
	case *PyList:
		result := make([]any, len(val.Items))
		for i, item := range val.Items {
			result[i] = ToGoValue(item)
		}
		return result
	case *PyTuple:
		result := make([]any, len(val.Items))
		for i, item := range val.Items {
			result[i] = ToGoValue(item)
		}
		return result
	case *PyDict:
		result := make(map[any]any)
		for k, v := range val.Items {
			result[ToGoValue(k)] = ToGoValue(v)
		}
		return result
	case *PyUserData:
		return val.Value
	default:
		return v
	}
}

// FromGoValue converts a Go value to a Python value
func FromGoValue(v any) Value {
	if v == nil {
		return None
	}

	val := reflect.ValueOf(v)

	switch val.Kind() {
	case reflect.Bool:
		return NewBool(val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return NewInt(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return NewInt(int64(val.Uint()))
	case reflect.Float32, reflect.Float64:
		return NewFloat(val.Float())
	case reflect.String:
		return NewString(val.String())
	case reflect.Slice:
		if val.Type().Elem().Kind() == reflect.Uint8 {
			// []byte -> PyBytes
			return &PyBytes{Value: val.Bytes()}
		}
		items := make([]Value, val.Len())
		for i := 0; i < val.Len(); i++ {
			items[i] = FromGoValue(val.Index(i).Interface())
		}
		return NewList(items)
	case reflect.Array:
		items := make([]Value, val.Len())
		for i := 0; i < val.Len(); i++ {
			items[i] = FromGoValue(val.Index(i).Interface())
		}
		return NewTuple(items)
	case reflect.Map:
		d := NewDict()
		iter := val.MapRange()
		for iter.Next() {
			k := FromGoValue(iter.Key().Interface())
			v := FromGoValue(iter.Value().Interface())
			d.Items[k] = v
		}
		return d
	case reflect.Ptr, reflect.Struct, reflect.Interface:
		// Wrap as userdata
		return NewUserData(v)
	default:
		// Wrap as userdata for unknown types
		return NewUserData(v)
	}
}

// =====================================
// Function Calling
// =====================================

// Call calls a Python callable with the given arguments
// This is the exported version of the internal call method
func (vm *VM) Call(callable Value, args []Value, kwargs map[string]Value) (Value, error) {
	return vm.call(callable, args, kwargs)
}

// IsTrue returns whether a Python value is truthy (without needing a VM reference).
// This handles common cases; for instances with __bool__/__len__, use vm.Truthy instead.
func IsTrue(v Value) bool {
	switch val := v.(type) {
	case *PyNone:
		return false
	case *PyBool:
		return val.Value
	case *PyInt:
		return val.Value != 0
	case *PyFloat:
		return val.Value != 0.0
	case *PyString:
		return len(val.Value) > 0
	case *PyList:
		return len(val.Items) > 0
	case *PyTuple:
		return len(val.Items) > 0
	case *PyDict:
		return len(val.Items) > 0 || val.DictLen() > 0
	case *PySet:
		return len(val.Items) > 0
	default:
		return v != nil
	}
}

// Truthy returns whether a Python value is truthy (exported wrapper)
func (vm *VM) Truthy(v Value) bool {
	return vm.truthy(v)
}

// Equal tests equality of two Python values (exported wrapper)
func (vm *VM) Equal(a, b Value) bool {
	return vm.equal(a, b)
}

// CompareOp performs a comparison operation (exported wrapper)
func (vm *VM) CompareOp(op Opcode, a, b Value) Value {
	return vm.compareOp(op, a, b)
}

// HashValue returns the hash of a Python value (exported wrapper)
func (vm *VM) HashValue(v Value) uint64 {
	return vm.hashValueVM(v)
}

// ToList converts a Python iterable to a Go slice of Values (exported wrapper).
// Handles lists, tuples, strings, ranges, sets, dicts, iterators, generators, etc.
func (vm *VM) ToList(v Value) ([]Value, error) {
	return vm.toList(v)
}

// =====================================
// Error Helpers
// =====================================

// PyPanicError is used to propagate Python exceptions through Go panics.
// This type is caught by the VM's recover and converted to a proper PyException.
type PyPanicError struct {
	ExcType string // Exception type name (e.g., "TypeError", "ValueError")
	Message string // Error message
}

func (e *PyPanicError) Error() string {
	return fmt.Sprintf("%s: %s", e.ExcType, e.Message)
}

// ArgError raises an argument error
func (vm *VM) ArgError(n int, msg string) {
	panic(&PyPanicError{
		ExcType: "TypeError",
		Message: fmt.Sprintf("bad argument #%d: %s", n, msg),
	})
}

// TypeError raises a type error
func (vm *VM) TypeError(expected string, got Value) {
	panic(&PyPanicError{
		ExcType: "TypeError",
		Message: fmt.Sprintf("expected %s, got %s", expected, vm.typeName(got)),
	})
}

// CallDunder looks up and calls a dunder method on a PyInstance via MRO.
// Returns (result, found). If found is false, no method was found.
// Panics with RaiseError on call errors (matching GoFunction conventions).
func (vm *VM) CallDunder(inst *PyInstance, name string, args ...Value) (Value, bool) {
	result, found, err := vm.callDunder(inst, name, args...)
	if err != nil {
		vm.RaiseError("%s", err.Error())
		return nil, false
	}
	return result, found
}

// CallFunction calls a PyFunction with the given arguments.
func (vm *VM) CallFunction(fn *PyFunction, args []Value, kwargs map[string]Value) (Value, error) {
	return vm.callFunction(fn, args, kwargs)
}

// IsInstanceOf checks if an instance is an instance of a class (including subclasses via MRO).
func (vm *VM) IsInstanceOf(inst *PyInstance, cls *PyClass) bool {
	return vm.isInstanceOf(inst, cls)
}

// CallDunderWithError looks up and calls a dunder method on a PyInstance via MRO.
// Returns (result, found, error) - for use in PyBuiltinFunc where errors are returned.
func (vm *VM) CallDunderWithError(inst *PyInstance, name string, args ...Value) (Value, bool, error) {
	return vm.callDunder(inst, name, args...)
}

// TypeNameOf returns the Python type name for a value.
func (vm *VM) TypeNameOf(v Value) string {
	return vm.typeName(v)
}

// GetIntIndex exports getIntIndex for use by stdlib packages.
func (vm *VM) GetIntIndex(v Value) (int64, error) {
	return vm.getIntIndex(v)
}

// RaiseError raises a Python-style error.
// The format string can optionally start with an exception type prefix like "ValueError: ".
func (vm *VM) RaiseError(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	excType := "RuntimeError"

	// Parse exception type from message if present (e.g., "ValueError: message")
	for _, prefix := range []string{
		"TypeError", "ValueError", "KeyError", "IndexError", "AttributeError",
		"RuntimeError", "StopIteration", "NotImplementedError", "OSError",
		"FileNotFoundError", "PermissionError", "FileExistsError", "IOError",
		"ZeroDivisionError", "OverflowError", "RecursionError", "SyntaxError",
		"LookupError", "ArithmeticError", "FloatingPointError", "EOFError",
		"BufferError", "TimeoutError", "ConnectionError", "ConnectionRefusedError",
		"ConnectionResetError", "ConnectionAbortedError", "BrokenPipeError",
		"IsADirectoryError", "NotADirectoryError", "InterruptedError",
		"BlockingIOError", "ChildProcessError", "ProcessLookupError",
		"UnicodeError", "UnicodeDecodeError", "UnicodeEncodeError",
		"UnicodeTranslateError", "ImportError", "ModuleNotFoundError",
		"UnboundLocalError", "NameError", "MemoryError", "AssertionError",
		"Warning", "DeprecationWarning", "RuntimeWarning", "UserWarning",
		"FutureWarning", "SyntaxWarning", "ImportWarning", "UnicodeWarning",
		"BytesWarning", "ResourceWarning", "EncodingWarning",
		"PendingDeprecationWarning", "StopAsyncIteration",
	} {
		if len(msg) > len(prefix)+2 && msg[:len(prefix)] == prefix && msg[len(prefix)] == ':' {
			excType = prefix
			msg = msg[len(prefix)+1:]
			// Trim leading space
			if len(msg) > 0 && msg[0] == ' ' {
				msg = msg[1:]
			}
			break
		}
	}

	panic(&PyPanicError{
		ExcType: excType,
		Message: msg,
	})
}

// =====================================
// Compile Function Bridge
// =====================================

// CompileFunc is set by the rage package to provide compilation capability
// for exec/eval/compile builtins. This avoids import cycles between
// runtime and compiler packages.
var CompileFunc func(source, filename, mode string) (*CodeObject, error)
