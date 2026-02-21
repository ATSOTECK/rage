package runtime

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"sync"
)

// Value represents a Python value
type Value any

// PyObject is the base interface for all Python objects
type PyObject interface {
	Type() string
	String() string
}

// PyNone represents Python's None
type PyNone struct{}

func (n *PyNone) Type() string   { return "NoneType" }
func (n *PyNone) String() string { return "None" }

// None is the singleton None value
var None = &PyNone{}

// PyNotImplementedType represents Python's NotImplemented singleton
type PyNotImplementedType struct{}

func (n *PyNotImplementedType) Type() string   { return "NotImplementedType" }
func (n *PyNotImplementedType) String() string { return "NotImplemented" }

// NotImplemented is the singleton NotImplemented value
var NotImplemented = &PyNotImplementedType{}

// PyBool represents a Python boolean
type PyBool struct {
	Value bool
}

func (b *PyBool) Type() string { return "bool" }
func (b *PyBool) String() string {
	if b.Value {
		return "True"
	}
	return "False"
}

// True and False are singleton boolean values
var (
	True  = &PyBool{Value: true}
	False = &PyBool{Value: false}
)

// PyInt represents a Python integer
type PyInt struct {
	Value    int64
	BigValue *big.Int // non-nil for integers that overflow int64
}

func (i *PyInt) Type() string { return "int" }
func (i *PyInt) String() string {
	if i.BigValue != nil {
		return i.BigValue.String()
	}
	return fmt.Sprintf("%d", i.Value)
}

// IsBig returns true if this integer uses big.Int representation
func (i *PyInt) IsBig() bool {
	return i.BigValue != nil
}

// BigIntValue returns the big.Int representation of this integer
func (i *PyInt) BigIntValue() *big.Int {
	if i.BigValue != nil {
		return i.BigValue
	}
	return big.NewInt(i.Value)
}

// MakeBigInt returns a PyInt from a big.Int value
func MakeBigInt(v *big.Int) *PyInt {
	// If it fits in int64, use the regular representation
	if v.IsInt64() {
		return MakeInt(v.Int64())
	}
	return &PyInt{BigValue: v}
}

// Small integer cache for common values (-5 to 256)
// This avoids allocations for frequently used integers
const (
	smallIntMin   = -5
	smallIntMax   = 256
	smallIntCount = smallIntMax - smallIntMin + 1
)

var smallIntCache [smallIntCount]*PyInt

func init() {
	for i := 0; i < smallIntCount; i++ {
		smallIntCache[i] = &PyInt{Value: int64(i + smallIntMin)}
	}
}

// MakeInt returns a PyInt, using cached values for small integers
func MakeInt(v int64) *PyInt {
	if v >= smallIntMin && v <= smallIntMax {
		return smallIntCache[v-smallIntMin]
	}
	return &PyInt{Value: v}
}

// String interning for common/short strings
// This reduces memory usage and speeds up string comparisons
var (
	stringInternPool     = make(map[string]*PyString)
	stringInternMaxLen   = 64 // Only intern strings up to this length
	stringInternPoolLock sync.RWMutex
)

// Common strings that are pre-interned
var commonStrings = []string{
	"", "__init__", "__str__", "__repr__", "__eq__", "__ne__", "__lt__", "__le__",
	"__gt__", "__ge__", "__hash__", "__len__", "__iter__", "__next__", "__getitem__",
	"__setitem__", "__delitem__", "__contains__", "__call__", "__add__", "__sub__",
	"__mul__", "__truediv__", "__floordiv__", "__mod__", "__pow__", "__and__",
	"__or__", "__xor__", "__neg__", "__pos__", "__abs__", "__invert__",
	"self", "None", "True", "False", "args", "kwargs", "name", "value",
	"key", "item", "index", "result", "data", "error", "message", "type",
}

func init() {
	// Pre-intern common strings
	for _, s := range commonStrings {
		stringInternPool[s] = &PyString{Value: s}
	}
}

// InternString returns an interned PyString for short strings
// This allows pointer comparison for equality in many cases
func InternString(s string) *PyString {
	if len(s) > stringInternMaxLen {
		return &PyString{Value: s}
	}

	// Use write lock directly to avoid race between read unlock and write lock.
	// The slight performance cost is acceptable for correctness.
	stringInternPoolLock.Lock()
	defer stringInternPoolLock.Unlock()

	if interned, ok := stringInternPool[s]; ok {
		return interned
	}
	interned := &PyString{Value: s}
	stringInternPool[s] = interned
	return interned
}

// intPow computes base^exp using binary exponentiation to avoid float precision loss
func intPow(base, exp int64) int64 {
	if exp == 0 {
		return 1
	}
	result := int64(1)
	for exp > 0 {
		if exp&1 == 1 {
			result *= base
		}
		base *= base
		exp >>= 1
	}
	return result
}

// ptrValue returns a uintptr for a Value, used for cycle detection in equality
func ptrValue(v Value) uintptr {
	return reflect.ValueOf(v).Pointer()
}

// isHashable checks if a value can be used as a dict key or set member
// In Python, mutable types (list, dict, set) are not hashable
func isHashable(v Value) bool {
	switch v.(type) {
	case *PyList, *PyDict, *PySet:
		return false
	default:
		return true
	}
}

// hashValue computes a hash for a Python value
// Only hashable types should be passed to this function
func hashValue(v Value) uint64 {
	switch val := v.(type) {
	case *PyNone:
		return 0x9e3779b97f4a7c15 // FNV offset basis
	case *PyBool:
		// Bool hash must match int hash: hash(True) == hash(1), hash(False) == hash(0)
		h := uint64(0)
		if val.Value {
			h = 1
		}
		h ^= h >> 33
		h *= 0xff51afd7ed558ccd
		h ^= h >> 33
		h *= 0xc4ceb9fe1a85ec53
		h ^= h >> 33
		return h
	case *PyInt:
		// Use FNV-1a style hashing for integers
		h := uint64(val.Value)
		h ^= h >> 33
		h *= 0xff51afd7ed558ccd
		h ^= h >> 33
		h *= 0xc4ceb9fe1a85ec53
		h ^= h >> 33
		return h
	case *PyFloat:
		// If the float is a whole number, hash it the same as the equivalent int
		// This ensures hash(1) == hash(1.0) as required by Python
		if val.Value == math.Trunc(val.Value) && !math.IsInf(val.Value, 0) && !math.IsNaN(val.Value) {
			intVal := int64(val.Value)
			h := uint64(intVal)
			h ^= h >> 33
			h *= 0xff51afd7ed558ccd
			h ^= h >> 33
			h *= 0xc4ceb9fe1a85ec53
			h ^= h >> 33
			return h
		}
		// Hash the bit representation of the float
		bits := math.Float64bits(val.Value)
		h := bits
		h ^= h >> 33
		h *= 0xff51afd7ed558ccd
		h ^= h >> 33
		return h
	case *PyComplex:
		// hash(complex(x, 0)) must equal hash(x) for real numbers
		if val.Imag == 0 {
			if val.Real == math.Trunc(val.Real) && !math.IsInf(val.Real, 0) && !math.IsNaN(val.Real) {
				return hashValue(MakeInt(int64(val.Real)))
			}
			return hashValue(&PyFloat{Value: val.Real})
		}
		// Combine real and imag hashes
		hr := hashValue(&PyFloat{Value: val.Real})
		hi := hashValue(&PyFloat{Value: val.Imag})
		return hr ^ (hi * 1000003)
	case *PyString:
		// FNV-1a hash for strings
		h := uint64(0xcbf29ce484222325)
		for i := 0; i < len(val.Value); i++ {
			h ^= uint64(val.Value[i])
			h *= 0x100000001b3
		}
		return h
	case *PyBytes:
		h := uint64(0xcbf29ce484222325)
		for i := 0; i < len(val.Value); i++ {
			h ^= uint64(val.Value[i])
			h *= 0x100000001b3
		}
		return h
	case *PyTuple:
		// Hash tuple by combining element hashes
		h := uint64(0xcbf29ce484222325)
		for _, item := range val.Items {
			itemHash := hashValue(item)
			h ^= itemHash
			h *= 0x100000001b3
		}
		return h
	case *PyClass:
		// Classes hash by identity (pointer)
		return uint64(ptrValue(v))
	case *PyInstance:
		// Instances hash by identity by default
		// Note: __hash__ support is handled in hashValueVM method
		return uint64(ptrValue(v))
	case *PyFunction:
		return uint64(ptrValue(v))
	case *PyBuiltinFunc:
		return uint64(ptrValue(v))
	case *PyFrozenSet:
		// Hash frozenset by XORing element hashes (order-independent)
		h := uint64(0xcbf29ce484222325)
		for item := range val.Items {
			itemHash := hashValue(item)
			h ^= itemHash
		}
		return h
	default:
		// Default to pointer hash for other types
		return uint64(ptrValue(v))
	}
}

// hashValueVM computes a hash with VM context, supporting __hash__ dunder method
func (vm *VM) hashValueVM(v Value) uint64 {
	if inst, ok := v.(*PyInstance); ok {
		// Check for __hash__ method
		if result, found, err := vm.callDunder(inst, "__hash__"); found && err == nil {
			if i, ok := result.(*PyInt); ok {
				return uint64(i.Value)
			}
		}
	}
	return hashValue(v)
}

// PyFloat represents a Python float
type PyFloat struct {
	Value float64
}

func (f *PyFloat) Type() string   { return "float" }
func (f *PyFloat) String() string { return fmt.Sprintf("%g", f.Value) }

// PyComplex represents a Python complex number
type PyComplex struct {
	Real float64
	Imag float64
}

func (c *PyComplex) Type() string   { return "complex" }
func (c *PyComplex) String() string { return formatComplex(c.Real, c.Imag) }

// MakeComplex creates a PyComplex value
func MakeComplex(real, imag float64) *PyComplex {
	return &PyComplex{Real: real, Imag: imag}
}

// formatComplex formats a complex number matching CPython output
func formatComplex(real, imag float64) string {
	formatPart := func(v float64) string {
		s := fmt.Sprintf("%g", v)
		return s
	}

	if real == 0 && !math.Signbit(real) {
		// Pure imaginary: just show the imaginary part
		return formatPart(imag) + "j"
	}
	// Full form with parens: (real+imagj)
	imagStr := formatPart(imag)
	sign := "+"
	if imag < 0 || math.IsNaN(imag) {
		sign = ""
	}
	return "(" + formatPart(real) + sign + imagStr + "j)"
}

// PyString represents a Python string
type PyString struct {
	Value string
}

func (s *PyString) Type() string   { return "str" }
func (s *PyString) String() string { return s.Value }

// PyBytes represents Python bytes
type PyBytes struct {
	Value []byte
}

func (b *PyBytes) Type() string   { return "bytes" }
func (b *PyBytes) String() string { return fmt.Sprintf("b'%s'", string(b.Value)) }
