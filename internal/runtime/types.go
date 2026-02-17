package runtime

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strings"
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

// PyList represents a Python list
type PyList struct {
	Items []Value
}

func (l *PyList) Type() string { return "list" }
func (l *PyList) String() string {
	return fmt.Sprintf("%v", l.Items)
}

// PyTuple represents a Python tuple
type PyTuple struct {
	Items []Value
}

func (t *PyTuple) Type() string { return "tuple" }
func (t *PyTuple) String() string {
	return fmt.Sprintf("%v", t.Items)
}

// dictEntry represents a key-value pair in a PyDict
type dictEntry struct {
	key   Value
	value Value
}

// PyDict represents a Python dictionary with hash-based lookups
type PyDict struct {
	Items         map[Value]Value        // Legacy field for compatibility
	buckets       map[uint64][]dictEntry // Hash buckets for O(1) lookup
	size          int
	orderedKeys   []Value // Insertion-ordered keys for Python 3.7+ dict ordering
	instanceOwner *PyInstance            // if non-nil, sync mutations back to instance's Dict
}

func (d *PyDict) Type() string { return "dict" }
func (d *PyDict) String() string {
	return fmt.Sprintf("%v", d.Items)
}

// DictGet retrieves a value by key using hash-based lookup
func (d *PyDict) DictGet(key Value, vm *VM) (Value, bool) {
	if d.buckets == nil {
		// Fall back to legacy Items lookup
		if val, ok := d.Items[key]; ok {
			return val, true
		}
		for k, v := range d.Items {
			if vm.equal(k, key) {
				return v, true
			}
		}
		return nil, false
	}
	h := vm.hashValueVM(key)
	entries := d.buckets[h]
	for _, e := range entries {
		if vm.equal(e.key, key) {
			return e.value, true
		}
	}
	return nil, false
}

// DictSet sets a key-value pair using hash-based storage
func (d *PyDict) DictSet(key, value Value, vm *VM) {
	if d.buckets == nil {
		d.buckets = make(map[uint64][]dictEntry)
	}
	h := vm.hashValueVM(key)
	entries := d.buckets[h]
	for i, e := range entries {
		if vm.equal(e.key, key) {
			d.buckets[h][i].value = value
			// Update legacy Items using original key from bucket (value equality)
			if d.Items != nil {
				d.deleteItemByEquality(e.key, vm)
				d.Items[e.key] = value
			}
			// Update value in orderedKeys entry (key stays the same, just update bucket value)
			return
		}
	}
	d.buckets[h] = append(entries, dictEntry{key: key, value: value})
	d.size++
	d.orderedKeys = append(d.orderedKeys, key)
	// Also update legacy Items for compatibility
	if d.Items == nil {
		d.Items = make(map[Value]Value)
	}
	d.deleteItemByEquality(key, vm) // Remove any existing entry with equivalent key
	d.Items[key] = value
	// Sync back to instance dict if this is a __dict__ proxy
	if d.instanceOwner != nil {
		if ks, ok := key.(*PyString); ok {
			d.instanceOwner.Dict[ks.Value] = value
		}
	}
}

// deleteItemByEquality removes a key from legacy Items using value equality
func (d *PyDict) deleteItemByEquality(key Value, vm *VM) {
	for k := range d.Items {
		if vm.equal(k, key) {
			delete(d.Items, k)
			return
		}
	}
}

// removeOrderedKey removes a key from the orderedKeys slice using value equality
func (d *PyDict) removeOrderedKey(key Value, vm *VM) {
	for i, k := range d.orderedKeys {
		if vm.equal(k, key) {
			d.orderedKeys = append(d.orderedKeys[:i], d.orderedKeys[i+1:]...)
			return
		}
	}
}

// Keys returns keys in insertion order
func (d *PyDict) Keys(vm *VM) []Value {
	if len(d.orderedKeys) > 0 {
		return d.orderedKeys
	}
	// Fallback for dicts created without ordered tracking
	var keys []Value
	for k := range d.Items {
		keys = append(keys, k)
	}
	return keys
}

// DictDelete removes a key using hash-based lookup
func (d *PyDict) DictDelete(key Value, vm *VM) bool {
	if d.buckets == nil {
		// Use value equality for legacy Items
		for k := range d.Items {
			if vm.equal(k, key) {
				delete(d.Items, k)
				d.removeOrderedKey(key, vm)
				if d.instanceOwner != nil {
					if ks, ok := key.(*PyString); ok {
						delete(d.instanceOwner.Dict, ks.Value)
					}
				}
				return true
			}
		}
		return false
	}
	h := vm.hashValueVM(key)
	entries := d.buckets[h]
	for i, e := range entries {
		if vm.equal(e.key, key) {
			// Remove entry by replacing with last and truncating
			d.buckets[h] = append(entries[:i], entries[i+1:]...)
			d.size--
			d.deleteItemByEquality(e.key, vm)
			d.removeOrderedKey(key, vm)
			if d.instanceOwner != nil {
				if ks, ok := key.(*PyString); ok {
					delete(d.instanceOwner.Dict, ks.Value)
				}
			}
			return true
		}
	}
	return false
}

// DictContains checks if a key exists using hash-based lookup
func (d *PyDict) DictContains(key Value, vm *VM) bool {
	_, found := d.DictGet(key, vm)
	return found
}

// DictLen returns the number of items
func (d *PyDict) DictLen() int {
	if d.buckets != nil {
		return d.size
	}
	return len(d.Items)
}

// setEntry represents a value in a PySet
type setEntry struct {
	value Value
}

// PySet represents a Python set with hash-based lookups
type PySet struct {
	Items   map[Value]struct{}    // Legacy field for compatibility
	buckets map[uint64][]setEntry // Hash buckets for O(1) lookup
	size    int
}

func (s *PySet) Type() string { return "set" }
func (s *PySet) String() string {
	return fmt.Sprintf("%v", s.Items)
}

// SetAdd adds a value to the set using hash-based storage
func (s *PySet) SetAdd(value Value, vm *VM) {
	if s.buckets == nil {
		s.buckets = make(map[uint64][]setEntry)
	}
	h := vm.hashValueVM(value)
	entries := s.buckets[h]
	for _, e := range entries {
		if vm.equal(e.value, value) {
			return // Already exists
		}
	}
	s.buckets[h] = append(entries, setEntry{value: value})
	s.size++
	// Also update legacy Items for compatibility
	if s.Items == nil {
		s.Items = make(map[Value]struct{})
	}
	s.Items[value] = struct{}{}
}

// SetContains checks if a value exists using hash-based lookup
func (s *PySet) SetContains(value Value, vm *VM) bool {
	if s.buckets == nil {
		// Fall back to legacy Items lookup
		if _, ok := s.Items[value]; ok {
			return true
		}
		for k := range s.Items {
			if vm.equal(k, value) {
				return true
			}
		}
		return false
	}
	h := vm.hashValueVM(value)
	entries := s.buckets[h]
	for _, e := range entries {
		if vm.equal(e.value, value) {
			return true
		}
	}
	return false
}

// SetRemove removes a value from the set
func (s *PySet) SetRemove(value Value, vm *VM) bool {
	if s.buckets == nil {
		delete(s.Items, value)
		return true
	}
	h := vm.hashValueVM(value)
	entries := s.buckets[h]
	for i, e := range entries {
		if vm.equal(e.value, value) {
			s.buckets[h] = append(entries[:i], entries[i+1:]...)
			s.size--
			delete(s.Items, value)
			return true
		}
	}
	return false
}

// SetLen returns the number of items
func (s *PySet) SetLen() int {
	if s.buckets != nil {
		return s.size
	}
	return len(s.Items)
}

// PyFrozenSet represents an immutable Python frozenset
type PyFrozenSet struct {
	Items   map[Value]struct{}    // Legacy field for compatibility
	buckets map[uint64][]setEntry // Hash buckets for O(1) lookup
	size    int
}

func (s *PyFrozenSet) Type() string { return "frozenset" }
func (s *PyFrozenSet) String() string {
	if len(s.Items) == 0 {
		return "frozenset()"
	}
	return fmt.Sprintf("frozenset(%v)", s.Items)
}

// FrozenSetAdd adds a value to the frozenset (used during construction)
func (s *PyFrozenSet) FrozenSetAdd(value Value, vm *VM) {
	if s.buckets == nil {
		s.buckets = make(map[uint64][]setEntry)
	}
	h := vm.hashValueVM(value)
	entries := s.buckets[h]
	for _, e := range entries {
		if vm.equal(e.value, value) {
			return // Already exists
		}
	}
	s.buckets[h] = append(entries, setEntry{value: value})
	s.size++
	if s.Items == nil {
		s.Items = make(map[Value]struct{})
	}
	s.Items[value] = struct{}{}
}

// FrozenSetContains checks if a value exists using hash-based lookup
func (s *PyFrozenSet) FrozenSetContains(value Value, vm *VM) bool {
	if s.buckets == nil {
		if _, ok := s.Items[value]; ok {
			return true
		}
		for k := range s.Items {
			if vm.equal(k, value) {
				return true
			}
		}
		return false
	}
	h := vm.hashValueVM(value)
	entries := s.buckets[h]
	for _, e := range entries {
		if vm.equal(e.value, value) {
			return true
		}
	}
	return false
}

// FrozenSetLen returns the number of items
func (s *PyFrozenSet) FrozenSetLen() int {
	if s.buckets != nil {
		return s.size
	}
	return len(s.Items)
}

// PyFunction represents a Python function
type PyFunction struct {
	Code       *CodeObject
	Globals    map[string]Value
	Defaults   *PyTuple
	Closure    []*PyCell
	Name       string
	IsAbstract bool // Set by @abstractmethod decorator
}

func (f *PyFunction) Type() string   { return "function" }
func (f *PyFunction) String() string { return fmt.Sprintf("<function %s>", f.Name) }

// PyCell represents a cell for closures
type PyCell struct {
	Value Value
}

// PyMethod represents a bound method
type PyMethod struct {
	Func     *PyFunction
	Instance Value
}

func (m *PyMethod) Type() string   { return "method" }
func (m *PyMethod) String() string { return fmt.Sprintf("<bound method %s>", m.Func.Name) }

// PyBuiltinFunc represents a built-in function
type PyBuiltinFunc struct {
	Name string
	Fn   func(args []Value, kwargs map[string]Value) (Value, error)
}

func (b *PyBuiltinFunc) Type() string   { return "builtin_function_or_method" }
func (b *PyBuiltinFunc) String() string { return fmt.Sprintf("<built-in function %s>", b.Name) }

// PyClass represents a Python class
type PyClass struct {
	Name                 string
	Bases                []*PyClass
	Dict                 map[string]Value
	Mro                  []*PyClass // Method Resolution Order
	IsABC                bool       // True if class uses ABC abstract method checking
	RegisteredSubclasses []*PyClass // Virtual subclasses registered via ABC.register()
	Metaclass            *PyClass   // Custom metaclass (if any)
	Slots                []string   // nil means no __slots__ (dict allowed); non-nil restricts instance attrs
}

func (c *PyClass) Type() string   { return "type" }
func (c *PyClass) String() string { return fmt.Sprintf("<class '%s'>", c.Name) }

// PyInstance represents an instance of a class
type PyInstance struct {
	Class *PyClass
	Dict  map[string]Value   // nil when class defines __slots__
	Slots map[string]Value   // non-nil when class defines __slots__
}

func (i *PyInstance) Type() string   { return i.Class.Name }
func (i *PyInstance) String() string { return fmt.Sprintf("<%s object>", i.Class.Name) }

// PyProperty represents a Python property descriptor
type PyProperty struct {
	Fget Value
	Fset Value
	Fdel Value
	Doc  string
}

func (p *PyProperty) Type() string   { return "property" }
func (p *PyProperty) String() string { return "<property object>" }

// PyClassMethod wraps a function to bind class as first argument
type PyClassMethod struct {
	Func Value
}

func (c *PyClassMethod) Type() string   { return "classmethod" }
func (c *PyClassMethod) String() string { return "<classmethod object>" }

// PyStaticMethod wraps a function to prevent binding
type PyStaticMethod struct {
	Func Value
}

func (s *PyStaticMethod) Type() string   { return "staticmethod" }
func (s *PyStaticMethod) String() string { return "<staticmethod object>" }

// PySuper represents Python's super() proxy object
type PySuper struct {
	ThisClass *PyClass // The class where super() was called (__class__)
	Instance  Value    // The instance (self) or class
	StartIdx  int      // Index in MRO to start searching from (after ThisClass)
}

func (s *PySuper) Type() string   { return "super" }
func (s *PySuper) String() string { return "<super object>" }

// PyException represents a Python exception
type PyException struct {
	ExcType   *PyClass         // Exception class (e.g., ValueError, TypeError)
	TypeName  string           // Exception type name (used when ExcType is nil)
	Args      *PyTuple         // Exception arguments
	Message   string           // String representation
	Cause     *PyException     // __cause__ for chained exceptions (raise X from Y)
	Context   *PyException     // __context__ for implicit chaining
	Traceback []TracebackEntry // Traceback frames
}

func (e *PyException) Type() string {
	if e.ExcType != nil {
		return e.ExcType.Name
	}
	return e.TypeName
}

func (e *PyException) String() string {
	return e.formatError()
}

func (e *PyException) Error() string {
	return e.formatError()
}

// formatError creates a properly formatted error message for the exception
func (e *PyException) formatError() string {
	typeName := e.Type()
	var msg string

	// If we have args, format with them
	if e.Args != nil && len(e.Args.Items) > 0 {
		if len(e.Args.Items) == 1 {
			if s, ok := e.Args.Items[0].(*PyString); ok {
				msg = fmt.Sprintf("%s: %s", typeName, s.Value)
			} else {
				msg = fmt.Sprintf("%s: %v", typeName, e.Args.Items[0])
			}
		} else {
			// Multiple args
			parts := make([]string, len(e.Args.Items))
			for i, item := range e.Args.Items {
				if s, ok := item.(*PyString); ok {
					parts[i] = s.Value
				} else {
					parts[i] = fmt.Sprintf("%v", item)
				}
			}
			msg = fmt.Sprintf("%s: (%s)", typeName, strings.Join(parts, ", "))
		}
	} else if e.Message != "" && e.Message != typeName {
		// If we have a message that's different from just the type name
		// Avoid duplicating the type name if message already contains it
		if strings.HasPrefix(e.Message, typeName+":") {
			msg = e.Message
		} else {
			msg = fmt.Sprintf("%s: %s", typeName, e.Message)
		}
	} else {
		msg = typeName
	}

	// Add location info from traceback if available
	if len(e.Traceback) > 0 {
		// Try to find the most relevant frame:
		// - Skip frames in test_framework.py (they're not useful for debugging)
		// - Prefer frames with actual filenames over <string>
		var bestFrame *TracebackEntry
		for i := range e.Traceback {
			tb := &e.Traceback[i]
			if tb.Line <= 0 {
				continue
			}
			// Skip test framework internal frames
			if strings.HasSuffix(tb.Filename, "test_framework.py") {
				continue
			}
			bestFrame = tb
			break
		}

		// If no good frame found, use the first one with a line number
		if bestFrame == nil {
			for i := range e.Traceback {
				if e.Traceback[i].Line > 0 {
					bestFrame = &e.Traceback[i]
					break
				}
			}
		}

		if bestFrame != nil {
			location := fmt.Sprintf(" (line %d", bestFrame.Line)
			if bestFrame.Filename != "" && bestFrame.Filename != "<string>" {
				location = fmt.Sprintf(" (%s:%d", bestFrame.Filename, bestFrame.Line)
			}
			if bestFrame.Function != "" && bestFrame.Function != "<module>" {
				location += fmt.Sprintf(" in %s)", bestFrame.Function)
			} else {
				location += ")"
			}
			msg += location
		}
	}

	return msg
}

// TracebackEntry represents a single frame in a traceback
type TracebackEntry struct {
	Filename string
	Line     int
	Function string
}

// Frame represents a call frame
type Frame struct {
	Code             *CodeObject
	IP               int              // Instruction pointer
	Stack            []Value          // Operand stack (pre-allocated)
	SP               int              // Stack pointer (index of next free slot)
	Locals           []Value          // Local variables
	Globals          map[string]Value // Global variables
	EnclosingGlobals map[string]Value // Enclosing globals (for class bodies)
	Builtins         map[string]Value // Built-in functions
	Cells            []*PyCell        // Closure cells
	BlockStack       []Block          // Block stack for try/except/finally
}

// Block represents a control flow block
type Block struct {
	Type    BlockType
	Handler int // Handler address
	Level   int // Stack level
}

// BlockType identifies the type of block
type BlockType int

const (
	BlockLoop BlockType = iota
	BlockExcept
	BlockFinally
	BlockWith
)



// PyRange represents a range object
type PyRange struct {
	Start, Stop, Step int64
}

func (r *PyRange) Type() string   { return "range" }
func (r *PyRange) String() string { return fmt.Sprintf("range(%d, %d, %d)", r.Start, r.Stop, r.Step) }

// Len returns the number of elements in the range
func (r *PyRange) Len() int64 {
	return rangeLen(r)
}

// Contains returns whether val is in the range
func (r *PyRange) Contains(val int64) bool {
	if r.Step > 0 {
		if val < r.Start || val >= r.Stop {
			return false
		}
		return (val-r.Start)%r.Step == 0
	} else {
		if val > r.Start || val <= r.Stop {
			return false
		}
		return (r.Start-val)%(-r.Step) == 0
	}
}

func rangeLen(r *PyRange) int64 {
	if r.Step > 0 {
		if r.Stop <= r.Start {
			return 0
		}
		return (r.Stop - r.Start + r.Step - 1) / r.Step
	} else {
		if r.Stop >= r.Start {
			return 0
		}
		return (r.Start - r.Stop - r.Step - 1) / (-r.Step)
	}
}

// PySlice represents a slice object for slicing sequences
type PySlice struct {
	Start Value // Can be nil (None) or int
	Stop  Value // Can be nil (None) or int
	Step  Value // Can be nil (None) or int
}

func (s *PySlice) Type() string { return "slice" }
func (s *PySlice) String() string {
	start, stop, step := "None", "None", "None"
	if s.Start != nil && s.Start != None {
		start = fmt.Sprintf("%v", s.Start.(*PyInt).Value)
	}
	if s.Stop != nil && s.Stop != None {
		stop = fmt.Sprintf("%v", s.Stop.(*PyInt).Value)
	}
	if s.Step != nil && s.Step != None {
		step = fmt.Sprintf("%v", s.Step.(*PyInt).Value)
	}
	return fmt.Sprintf("slice(%s, %s, %s)", start, stop, step)
}

// PyIterator wraps an iterator
type PyIterator struct {
	Items []Value
	Index int
}

func (i *PyIterator) Type() string   { return "iterator" }
func (i *PyIterator) String() string { return "<iterator>" }

// GeneratorState represents the state of a generator/coroutine
type GeneratorState int

const (
	GenCreated   GeneratorState = iota // Generator created but not started
	GenRunning                         // Generator is currently executing
	GenSuspended                       // Generator is suspended at yield
	GenClosed                          // Generator has finished or was closed
)

// PyGenerator represents a Python generator object
type PyGenerator struct {
	Frame      *Frame         // Suspended frame state
	Code       *CodeObject    // The generator's code object
	Name       string         // Generator function name
	State      GeneratorState // Current state
	YieldValue Value          // Value to send into generator on resume
}

func (g *PyGenerator) Type() string   { return "generator" }
func (g *PyGenerator) String() string { return fmt.Sprintf("<generator object %s>", g.Name) }

// PyCoroutine represents a Python coroutine object (async def)
type PyCoroutine struct {
	Frame      *Frame         // Suspended frame state
	Code       *CodeObject    // The coroutine's code object
	Name       string         // Coroutine function name
	State      GeneratorState // Current state (reuses generator states)
	YieldValue Value          // Value to send into coroutine on resume
}

func (c *PyCoroutine) Type() string   { return "coroutine" }
func (c *PyCoroutine) String() string { return fmt.Sprintf("<coroutine object %s>", c.Name) }

// GenericAlias represents a parameterized type like list[int] or dict[str, int]
type GenericAlias struct {
	Origin Value   // The original class/builtin (e.g., list, dict, or a PyClass)
	Args   []Value // The type arguments
}

func (g *GenericAlias) Type() string { return "GenericAlias" }
func (g *GenericAlias) String() string {
	return g.formatRepr()
}

func (g *GenericAlias) formatRepr() string {
	// Get origin name
	var originName string
	switch o := g.Origin.(type) {
	case *PyClass:
		originName = o.Name
	case *PyBuiltinFunc:
		originName = o.Name
	default:
		originName = fmt.Sprintf("%v", o)
	}

	if len(g.Args) == 0 {
		return originName
	}

	args := make([]string, len(g.Args))
	for i, arg := range g.Args {
		args[i] = genericAliasArgRepr(arg)
	}
	return fmt.Sprintf("%s[%s]", originName, strings.Join(args, ", "))
}

func genericAliasArgRepr(v Value) string {
	switch a := v.(type) {
	case *PyClass:
		return a.Name
	case *PyBuiltinFunc:
		return a.Name
	case *PyNone:
		return "None"
	case *GenericAlias:
		return a.formatRepr()
	default:
		return fmt.Sprintf("%v", a)
	}
}

