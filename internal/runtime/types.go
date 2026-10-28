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

// PyEllipsisType represents Python's Ellipsis singleton (the `...` literal).
type PyEllipsisType struct{}

func (e *PyEllipsisType) Type() string   { return "ellipsis" }
func (e *PyEllipsisType) String() string { return "Ellipsis" }

// Ellipsis is the singleton Ellipsis value
var Ellipsis = &PyEllipsisType{}

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
// In Python, mutable types (list, dict, set) are not hashable, a tuple is
// hashable only if its items are, and an instance is unhashable when its
// class sets __hash__ = None (implicitly so when it defines __eq__ alone).
func isHashable(v Value) bool {
	switch val := v.(type) {
	case *PyList, *PyDict, *PySet:
		return false
	case *PyTuple:
		for _, item := range val.Items {
			if !isHashable(item) {
				return false
			}
		}
		return true
	case *PyInstance:
		for _, cls := range val.Class.Mro {
			if h, ok := cls.Dict["__hash__"]; ok {
				return h != None
			}
		}
		return true
	default:
		return true
	}
}

// Constants from CPython's Include/cpython/pyhash.h. Numeric hashes follow
// CPython exactly so that hash() results and set iteration order match.
const (
	pyHashBits    = 61
	pyHashModulus = (uint64(1) << pyHashBits) - 1
	pyHashInf     = 314159
	pyHashImag    = 1000003
	pyHashNone    = 0xFCA86420
)

// fixHash maps the reserved hash value -1 to -2, as CPython does.
func fixHash(h int64) uint64 {
	if h == -1 {
		h = -2
	}
	return uint64(h)
}

// hashInt64 returns CPython's hash of an int: the value reduced modulo 2**61-1,
// keeping its sign.
func hashInt64(v int64) uint64 {
	if v >= 0 {
		return fixHash(int64(uint64(v) % pyHashModulus))
	}
	// uint64(-v) is the correct magnitude even for math.MinInt64
	return fixHash(-int64(uint64(-v) % pyHashModulus))
}

// hashBigInt returns CPython's hash of an arbitrary-precision int.
func hashBigInt(b *big.Int) uint64 {
	m := new(big.Int).Abs(b)
	m.Mod(m, new(big.Int).SetUint64(pyHashModulus))
	h := int64(m.Uint64())
	if b.Sign() < 0 {
		h = -h
	}
	return fixHash(h)
}

// hashFloat returns CPython's hash of a finite or infinite float. NaN is
// hashed by identity in CPython, so callers handle it separately.
func hashFloat(v float64) uint64 {
	if math.IsInf(v, 1) {
		return pyHashInf
	}
	if math.IsInf(v, -1) {
		return fixHash(-pyHashInf)
	}
	m, e := math.Frexp(v)
	sign := int64(1)
	if m < 0 {
		sign = -1
		m = -m
	}
	var x uint64
	for m != 0 {
		x = ((x << 28) & pyHashModulus) | x>>(pyHashBits-28)
		m *= 268435456.0 // 2**28
		e -= 28
		y := uint64(m)
		m -= float64(y)
		x += y
		if x >= pyHashModulus {
			x -= pyHashModulus
		}
	}
	if e >= 0 {
		e = e % pyHashBits
	} else {
		e = pyHashBits - 1 - ((-1 - e) % pyHashBits)
	}
	x = ((x << uint(e)) & pyHashModulus) | x>>(pyHashBits-uint(e))
	return fixHash(int64(x) * sign)
}

// hashTupleItems combines element hashes using CPython's xxHash-based tuple hash.
func hashTupleItems(items []Value, hashItem func(Value) uint64) uint64 {
	const (
		prime1 = 11400714785074694791
		prime2 = 14029467366897019727
		prime5 = 2870177450012600261
	)
	acc := uint64(prime5)
	for _, item := range items {
		acc += hashItem(item) * prime2
		acc = (acc << 31) | (acc >> 33)
		acc *= prime1
	}
	acc += uint64(len(items)) ^ (prime5 ^ 3527539)
	if acc == ^uint64(0) {
		return 1546275796
	}
	return acc
}

// hashValue computes a hash for a Python value
// Only hashable types should be passed to this function
func hashValue(v Value) uint64 {
	switch val := v.(type) {
	case *PyNone:
		return pyHashNone
	case *PyBool:
		// Bool hash must match int hash: hash(True) == hash(1), hash(False) == hash(0)
		if val.Value {
			return 1
		}
		return 0
	case *PyInt:
		if val.BigValue != nil {
			return hashBigInt(val.BigValue)
		}
		return hashInt64(val.Value)
	case *PyFloat:
		if math.IsNaN(val.Value) {
			return uint64(ptrValue(v))
		}
		return hashFloat(val.Value)
	case *PyComplex:
		// hash(complex(x, 0)) == hash(x) falls out of the formula
		if math.IsNaN(val.Real) || math.IsNaN(val.Imag) {
			return uint64(ptrValue(v))
		}
		h := hashFloat(val.Real) + pyHashImag*hashFloat(val.Imag)
		return fixHash(int64(h))
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
		return hashTupleItems(val.Items, hashValue)
	case *PyFrozenSet:
		return frozenSetHash(&val.setTable)
	default:
		// Classes, functions, instances and other objects hash by identity
		return uint64(ptrValue(v))
	}
}

// hashValueVM computes a hash with VM context, supporting __hash__ dunder method
func (vm *VM) hashValueVM(v Value) uint64 {
	switch val := v.(type) {
	case *PyInstance:
		// Check for __hash__ method
		if result, found, err := vm.callDunder(val, "__hash__"); found && err == nil {
			if i, ok := result.(*PyInt); ok {
				// CPython reduces the returned int like any other int hash
				return hashValue(i)
			}
		}
	case *PyTuple:
		return hashTupleItems(val.Items, vm.hashValueVM)
	}
	return hashValue(v)
}

// PyFloat represents a Python float
type PyFloat struct {
	Value float64
}

func (f *PyFloat) Type() string   { return "float" }
func (f *PyFloat) String() string { return floatRepr(f.Value) }

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
func (b *PyBytes) String() string { return bytesRepr(b.Value) }
