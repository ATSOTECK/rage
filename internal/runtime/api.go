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
	Value     interface{}
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

// NewString creates a Python str from a Go string
func NewString(v string) *PyString {
	return &PyString{Value: v}
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

// NewUserData creates a new userdata wrapping a Go value
func NewUserData(v interface{}) *PyUserData {
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
	return len(vm.frame.Stack)
}

// SetTop sets the stack top to a specific index
func (vm *VM) SetTop(n int) {
	if n < 0 {
		vm.frame.Stack = vm.frame.Stack[:len(vm.frame.Stack)+n+1]
	} else if n < len(vm.frame.Stack) {
		vm.frame.Stack = vm.frame.Stack[:n]
	} else {
		for len(vm.frame.Stack) < n {
			vm.frame.Stack = append(vm.frame.Stack, None)
		}
	}
}

// Get returns the value at the given stack index (1-based, like Lua)
// Negative indices count from top
func (vm *VM) Get(idx int) Value {
	if idx > 0 {
		// Positive index: from bottom (1-based)
		if idx <= len(vm.frame.Stack) {
			return vm.frame.Stack[idx-1]
		}
	} else if idx < 0 {
		// Negative index: from top
		realIdx := len(vm.frame.Stack) + idx
		if realIdx >= 0 {
			return vm.frame.Stack[realIdx]
		}
	}
	return None
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

// TypeMetatable holds methods for a user-defined type
type TypeMetatable struct {
	Name    string
	Methods map[string]GoFunction
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
func (vm *VM) NewUserDataWithMeta(v interface{}, typeName string) *PyUserData {
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
	switch v.(type) {
	case *PyFunction, *PyBuiltinFunc, *PyGoFunc, *PyClass, *PyMethod:
		return true
	default:
		return false
	}
}

// =====================================
// Go Value Conversion
// =====================================

// ToGoValue converts a Python value to a Go value
func ToGoValue(v Value) interface{} {
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
		result := make([]interface{}, len(val.Items))
		for i, item := range val.Items {
			result[i] = ToGoValue(item)
		}
		return result
	case *PyTuple:
		result := make([]interface{}, len(val.Items))
		for i, item := range val.Items {
			result[i] = ToGoValue(item)
		}
		return result
	case *PyDict:
		result := make(map[interface{}]interface{})
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
func FromGoValue(v interface{}) Value {
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
// Error Helpers
// =====================================

// ArgError raises an argument error
func (vm *VM) ArgError(n int, msg string) {
	panic(fmt.Sprintf("bad argument #%d: %s", n, msg))
}

// TypeError raises a type error
func (vm *VM) TypeError(expected string, got Value) {
	panic(fmt.Sprintf("expected %s, got %s", expected, vm.typeName(got)))
}

// RaiseError raises a Python-style error
func (vm *VM) RaiseError(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
