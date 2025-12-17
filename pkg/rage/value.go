package rage

import (
	"fmt"

	"github.com/ATSOTECK/oink/internal/runtime"
)

// Value represents a Python value.
// Use the type assertion or helper methods to access the underlying Go value.
type Value interface {
	// Type returns the Python type name (e.g., "int", "str", "list")
	Type() string

	// String returns a string representation of the value
	String() string

	// GoValue returns the underlying Go value
	GoValue() interface{}

	// internal conversion
	toRuntime() runtime.Value
}

// =====================================
// Concrete Value Types
// =====================================

// NoneValue represents Python's None
type NoneValue struct{}

func (v NoneValue) Type() string            { return "NoneType" }
func (v NoneValue) String() string          { return "None" }
func (v NoneValue) GoValue() interface{}    { return nil }
func (v NoneValue) toRuntime() runtime.Value { return runtime.None }

// None is the singleton None value
var None Value = NoneValue{}

// BoolValue represents a Python bool
type BoolValue struct {
	value bool
}

func (v BoolValue) Type() string            { return "bool" }
func (v BoolValue) String() string          { return fmt.Sprintf("%v", v.value) }
func (v BoolValue) GoValue() interface{}    { return v.value }
func (v BoolValue) Bool() bool              { return v.value }
func (v BoolValue) toRuntime() runtime.Value { return runtime.NewBool(v.value) }

// True and False are the singleton bool values
var (
	True  Value = BoolValue{value: true}
	False Value = BoolValue{value: false}
)

// IntValue represents a Python int
type IntValue struct {
	value int64
}

func (v IntValue) Type() string            { return "int" }
func (v IntValue) String() string          { return fmt.Sprintf("%d", v.value) }
func (v IntValue) GoValue() interface{}    { return v.value }
func (v IntValue) Int() int64              { return v.value }
func (v IntValue) toRuntime() runtime.Value { return runtime.NewInt(v.value) }

// FloatValue represents a Python float
type FloatValue struct {
	value float64
}

func (v FloatValue) Type() string            { return "float" }
func (v FloatValue) String() string          { return fmt.Sprintf("%g", v.value) }
func (v FloatValue) GoValue() interface{}    { return v.value }
func (v FloatValue) Float() float64          { return v.value }
func (v FloatValue) toRuntime() runtime.Value { return runtime.NewFloat(v.value) }

// StringValue represents a Python str
type StringValue struct {
	value string
}

func (v StringValue) Type() string            { return "str" }
func (v StringValue) String() string          { return v.value }
func (v StringValue) GoValue() interface{}    { return v.value }
func (v StringValue) Str() string             { return v.value }
func (v StringValue) toRuntime() runtime.Value { return runtime.NewString(v.value) }

// ListValue represents a Python list
type ListValue struct {
	items []Value
}

func (v ListValue) Type() string { return "list" }
func (v ListValue) String() string {
	return fmt.Sprintf("%v", v.GoValue())
}
func (v ListValue) GoValue() interface{} {
	result := make([]interface{}, len(v.items))
	for i, item := range v.items {
		result[i] = item.GoValue()
	}
	return result
}
func (v ListValue) Items() []Value { return v.items }
func (v ListValue) Len() int       { return len(v.items) }
func (v ListValue) Get(i int) Value {
	if i >= 0 && i < len(v.items) {
		return v.items[i]
	}
	return None
}
func (v ListValue) toRuntime() runtime.Value {
	items := make([]runtime.Value, len(v.items))
	for i, item := range v.items {
		items[i] = toRuntime(item)
	}
	return runtime.NewList(items)
}

// TupleValue represents a Python tuple
type TupleValue struct {
	items []Value
}

func (v TupleValue) Type() string { return "tuple" }
func (v TupleValue) String() string {
	return fmt.Sprintf("%v", v.GoValue())
}
func (v TupleValue) GoValue() interface{} {
	result := make([]interface{}, len(v.items))
	for i, item := range v.items {
		result[i] = item.GoValue()
	}
	return result
}
func (v TupleValue) Items() []Value { return v.items }
func (v TupleValue) Len() int       { return len(v.items) }
func (v TupleValue) Get(i int) Value {
	if i >= 0 && i < len(v.items) {
		return v.items[i]
	}
	return None
}
func (v TupleValue) toRuntime() runtime.Value {
	items := make([]runtime.Value, len(v.items))
	for i, item := range v.items {
		items[i] = toRuntime(item)
	}
	return runtime.NewTuple(items)
}

// DictValue represents a Python dict
type DictValue struct {
	items map[string]Value
}

func (v DictValue) Type() string { return "dict" }
func (v DictValue) String() string {
	return fmt.Sprintf("%v", v.GoValue())
}
func (v DictValue) GoValue() interface{} {
	result := make(map[string]interface{})
	for k, val := range v.items {
		result[k] = val.GoValue()
	}
	return result
}
func (v DictValue) Items() map[string]Value { return v.items }
func (v DictValue) Len() int                { return len(v.items) }
func (v DictValue) Get(key string) Value {
	if val, ok := v.items[key]; ok {
		return val
	}
	return None
}
func (v DictValue) toRuntime() runtime.Value {
	d := runtime.NewDict()
	for k, val := range v.items {
		d.Items[runtime.NewString(k)] = toRuntime(val)
	}
	return d
}

// UserDataValue wraps arbitrary Go values
type UserDataValue struct {
	value interface{}
}

func (v UserDataValue) Type() string            { return "userdata" }
func (v UserDataValue) String() string          { return fmt.Sprintf("<userdata %T>", v.value) }
func (v UserDataValue) GoValue() interface{}    { return v.value }
func (v UserDataValue) toRuntime() runtime.Value { return runtime.NewUserData(v.value) }

// FunctionValue represents a Python function (for introspection)
type FunctionValue struct {
	name string
	rv   runtime.Value
}

func (v FunctionValue) Type() string            { return "function" }
func (v FunctionValue) String() string          { return fmt.Sprintf("<function %s>", v.name) }
func (v FunctionValue) GoValue() interface{}    { return nil }
func (v FunctionValue) Name() string            { return v.name }
func (v FunctionValue) toRuntime() runtime.Value { return v.rv }

// =====================================
// Value Constructors
// =====================================

// Bool creates a Python bool value
func Bool(v bool) Value {
	if v {
		return True
	}
	return False
}

// Int creates a Python int value
func Int(v int64) Value {
	return IntValue{value: v}
}

// Float creates a Python float value
func Float(v float64) Value {
	return FloatValue{value: v}
}

// String creates a Python str value
func String(v string) Value {
	return StringValue{value: v}
}

// List creates a Python list from Values
func List(items ...Value) Value {
	return ListValue{items: items}
}

// Tuple creates a Python tuple from Values
func Tuple(items ...Value) Value {
	return TupleValue{items: items}
}

// Dict creates a Python dict from string keys and Values
func Dict(pairs ...interface{}) Value {
	items := make(map[string]Value)
	for i := 0; i+1 < len(pairs); i += 2 {
		if key, ok := pairs[i].(string); ok {
			if val, ok := pairs[i+1].(Value); ok {
				items[key] = val
			} else {
				items[key] = FromGo(pairs[i+1])
			}
		}
	}
	return DictValue{items: items}
}

// UserData wraps a Go value for use in Python
func UserData(v interface{}) Value {
	return UserDataValue{value: v}
}

// FromGo converts a Go value to a Python Value
func FromGo(v interface{}) Value {
	if v == nil {
		return None
	}

	switch val := v.(type) {
	case Value:
		return val
	case bool:
		return Bool(val)
	case int:
		return Int(int64(val))
	case int8:
		return Int(int64(val))
	case int16:
		return Int(int64(val))
	case int32:
		return Int(int64(val))
	case int64:
		return Int(val)
	case uint:
		return Int(int64(val))
	case uint8:
		return Int(int64(val))
	case uint16:
		return Int(int64(val))
	case uint32:
		return Int(int64(val))
	case uint64:
		return Int(int64(val))
	case float32:
		return Float(float64(val))
	case float64:
		return Float(val)
	case string:
		return String(val)
	case []interface{}:
		items := make([]Value, len(val))
		for i, item := range val {
			items[i] = FromGo(item)
		}
		return ListValue{items: items}
	case map[string]interface{}:
		items := make(map[string]Value)
		for k, v := range val {
			items[k] = FromGo(v)
		}
		return DictValue{items: items}
	default:
		return UserData(v)
	}
}

// =====================================
// Type Checking Helpers
// =====================================

// IsNone returns true if the value is None
func IsNone(v Value) bool {
	_, ok := v.(NoneValue)
	return ok
}

// IsBool returns true if the value is a bool
func IsBool(v Value) bool {
	_, ok := v.(BoolValue)
	return ok
}

// IsInt returns true if the value is an int
func IsInt(v Value) bool {
	_, ok := v.(IntValue)
	return ok
}

// IsFloat returns true if the value is a float
func IsFloat(v Value) bool {
	_, ok := v.(FloatValue)
	return ok
}

// IsString returns true if the value is a string
func IsString(v Value) bool {
	_, ok := v.(StringValue)
	return ok
}

// IsList returns true if the value is a list
func IsList(v Value) bool {
	_, ok := v.(ListValue)
	return ok
}

// IsTuple returns true if the value is a tuple
func IsTuple(v Value) bool {
	_, ok := v.(TupleValue)
	return ok
}

// IsDict returns true if the value is a dict
func IsDict(v Value) bool {
	_, ok := v.(DictValue)
	return ok
}

// IsUserData returns true if the value is userdata
func IsUserData(v Value) bool {
	_, ok := v.(UserDataValue)
	return ok
}

// =====================================
// Type Assertion Helpers
// =====================================

// AsBool returns the bool value or false if not a bool
func AsBool(v Value) (bool, bool) {
	if bv, ok := v.(BoolValue); ok {
		return bv.value, true
	}
	return false, false
}

// AsInt returns the int value or 0 if not an int
func AsInt(v Value) (int64, bool) {
	if iv, ok := v.(IntValue); ok {
		return iv.value, true
	}
	return 0, false
}

// AsFloat returns the float value or 0 if not a float
func AsFloat(v Value) (float64, bool) {
	if fv, ok := v.(FloatValue); ok {
		return fv.value, true
	}
	// Also accept int as float
	if iv, ok := v.(IntValue); ok {
		return float64(iv.value), true
	}
	return 0, false
}

// AsString returns the string value or "" if not a string
func AsString(v Value) (string, bool) {
	if sv, ok := v.(StringValue); ok {
		return sv.value, true
	}
	return "", false
}

// AsList returns the list value or nil if not a list
func AsList(v Value) ([]Value, bool) {
	if lv, ok := v.(ListValue); ok {
		return lv.items, true
	}
	return nil, false
}

// AsTuple returns the tuple value or nil if not a tuple
func AsTuple(v Value) ([]Value, bool) {
	if tv, ok := v.(TupleValue); ok {
		return tv.items, true
	}
	return nil, false
}

// AsDict returns the dict value or nil if not a dict
func AsDict(v Value) (map[string]Value, bool) {
	if dv, ok := v.(DictValue); ok {
		return dv.items, true
	}
	return nil, false
}

// AsUserData returns the userdata value or nil if not userdata
func AsUserData(v Value) (interface{}, bool) {
	if uv, ok := v.(UserDataValue); ok {
		return uv.value, true
	}
	return nil, false
}

// =====================================
// Internal Conversion Functions
// =====================================

// toRuntime converts a rage.Value to a runtime.Value
func toRuntime(v Value) runtime.Value {
	if v == nil {
		return runtime.None
	}
	return v.toRuntime()
}

// fromRuntime converts a runtime.Value to a rage.Value
func fromRuntime(v runtime.Value) Value {
	if v == nil {
		return None
	}

	switch val := v.(type) {
	case *runtime.PyNone:
		return None
	case *runtime.PyBool:
		return Bool(val.Value)
	case *runtime.PyInt:
		return Int(val.Value)
	case *runtime.PyFloat:
		return Float(val.Value)
	case *runtime.PyString:
		return String(val.Value)
	case *runtime.PyList:
		items := make([]Value, len(val.Items))
		for i, item := range val.Items {
			items[i] = fromRuntime(item)
		}
		return ListValue{items: items}
	case *runtime.PyTuple:
		items := make([]Value, len(val.Items))
		for i, item := range val.Items {
			items[i] = fromRuntime(item)
		}
		return TupleValue{items: items}
	case *runtime.PyDict:
		items := make(map[string]Value)
		for k, v := range val.Items {
			if ks, ok := k.(*runtime.PyString); ok {
				items[ks.Value] = fromRuntime(v)
			} else {
				items[fmt.Sprintf("%v", runtime.ToGoValue(k))] = fromRuntime(v)
			}
		}
		return DictValue{items: items}
	case *runtime.PyUserData:
		return UserDataValue{value: val.Value}
	case *runtime.PyFunction:
		return FunctionValue{name: val.Name, rv: val}
	case *runtime.PyBuiltinFunc:
		return FunctionValue{name: val.Name, rv: val}
	case *runtime.PyGoFunc:
		return FunctionValue{name: val.Name, rv: val}
	default:
		// For other types, wrap as userdata
		return UserDataValue{value: v}
	}
}
