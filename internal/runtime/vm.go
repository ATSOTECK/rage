package runtime

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
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
	Value int64
}

func (i *PyInt) Type() string   { return "int" }
func (i *PyInt) String() string { return fmt.Sprintf("%d", i.Value) }

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
	stringInternPoolLock.RLock()
	if interned, ok := stringInternPool[s]; ok {
		stringInternPoolLock.RUnlock()
		return interned
	}
	stringInternPoolLock.RUnlock()

	// Not in pool, add it
	stringInternPoolLock.Lock()
	defer stringInternPoolLock.Unlock()
	// Double-check after acquiring write lock
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
		if val.Value {
			return 1
		}
		return 0
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
		// Hash the bit representation of the float
		bits := math.Float64bits(val.Value)
		h := bits
		h ^= h >> 33
		h *= 0xff51afd7ed558ccd
		h ^= h >> 33
		return h
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
		return uint64(ptrValue(v))
	case *PyFunction:
		return uint64(ptrValue(v))
	case *PyBuiltinFunc:
		return uint64(ptrValue(v))
	default:
		// Default to pointer hash for other types
		return uint64(ptrValue(v))
	}
}

// PyFloat represents a Python float
type PyFloat struct {
	Value float64
}

func (f *PyFloat) Type() string   { return "float" }
func (f *PyFloat) String() string { return fmt.Sprintf("%g", f.Value) }

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
	Items   map[Value]Value      // Legacy field for compatibility
	buckets map[uint64][]dictEntry // Hash buckets for O(1) lookup
	size    int
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
	h := hashValue(key)
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
	h := hashValue(key)
	entries := d.buckets[h]
	for i, e := range entries {
		if vm.equal(e.key, key) {
			d.buckets[h][i].value = value
			// Also update legacy Items for compatibility
			if d.Items != nil {
				d.Items[key] = value
			}
			return
		}
	}
	d.buckets[h] = append(entries, dictEntry{key: key, value: value})
	d.size++
	// Also update legacy Items for compatibility
	if d.Items == nil {
		d.Items = make(map[Value]Value)
	}
	d.Items[key] = value
}

// DictDelete removes a key using hash-based lookup
func (d *PyDict) DictDelete(key Value, vm *VM) bool {
	if d.buckets == nil {
		delete(d.Items, key)
		return true
	}
	h := hashValue(key)
	entries := d.buckets[h]
	for i, e := range entries {
		if vm.equal(e.key, key) {
			// Remove entry by replacing with last and truncating
			d.buckets[h] = append(entries[:i], entries[i+1:]...)
			d.size--
			delete(d.Items, key)
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
	h := hashValue(value)
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
	h := hashValue(value)
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
	h := hashValue(value)
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

// PyFunction represents a Python function
type PyFunction struct {
	Code     *CodeObject
	Globals  map[string]Value
	Defaults *PyTuple
	Closure  []*PyCell
	Name     string
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
	Name  string
	Bases []*PyClass
	Dict  map[string]Value
	Mro   []*PyClass // Method Resolution Order
}

func (c *PyClass) Type() string   { return "type" }
func (c *PyClass) String() string { return fmt.Sprintf("<class '%s'>", c.Name) }

// PyInstance represents an instance of a class
type PyInstance struct {
	Class *PyClass
	Dict  map[string]Value
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
func (e *PyException) String() string { return e.Message }
func (e *PyException) Error() string  { return e.Message }

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
)

// VM is the Python virtual machine
type VM struct {
	frames   []*Frame
	frame    *Frame // Current frame
	Globals  map[string]Value
	builtins map[string]Value

	// Execution control
	ctx           context.Context
	checkCounter  int64 // Counts down to next context check
	checkInterval int64 // Check context every N instructions

	// Exception handling state
	currentException *PyException // Currently active exception being handled
	lastException    *PyException // Last raised exception (for bare raise)

	// Generator exception injection
	generatorThrow *PyException // Exception to throw into generator on resume
}

// TimeoutError is returned when script execution exceeds the time limit
type TimeoutError struct {
	Timeout time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("script execution timed out after %v", e.Timeout)
}

// CancelledError is returned when script execution is cancelled via context
type CancelledError struct{}

func (e *CancelledError) Error() string {
	return "script execution was cancelled"
}

// exceptionHandledInOuterFrame is a sentinel error to signal that an exception
// was handled in an outer frame and execution should continue there
type exceptionHandledInOuterFrame struct{}

func (e *exceptionHandledInOuterFrame) Error() string {
	return "exception handled in outer frame"
}

var errExceptionHandledInOuterFrame = &exceptionHandledInOuterFrame{}

// NewVM creates a new virtual machine
func NewVM() *VM {
	vm := &VM{
		Globals:       make(map[string]Value),
		builtins:      make(map[string]Value),
		checkInterval: 1000, // Check context every 1000 instructions by default
		checkCounter:  1000, // Initialize counter
	}
	vm.initBuiltins()

	// Add pending builtins registered by stdlib modules (e.g., open() from IO module)
	for name, fn := range GetPendingBuiltins() {
		vm.builtins[name] = NewGoFunction(name, fn)
	}

	return vm
}

// SetCheckInterval sets how often the VM checks for timeout/cancellation.
// Lower values are more responsive but have more overhead.
// Default is 1000 instructions.
func (vm *VM) SetCheckInterval(n int64) {
	if n < 1 {
		n = 1
	}
	vm.checkInterval = n
	vm.checkCounter = n
}

func (vm *VM) initBuiltins() {
	vm.builtins["print"] = &PyBuiltinFunc{
		Name: "print",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			for i, arg := range args {
				if i > 0 {
					fmt.Print(" ")
				}
				fmt.Print(vm.str(arg))
			}
			fmt.Println()
			return None, nil
		},
	}

	vm.builtins["len"] = &PyBuiltinFunc{
		Name: "len",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("len() takes exactly one argument (%d given)", len(args))
			}
			switch v := args[0].(type) {
			case *PyString:
				return MakeInt(int64(len(v.Value))), nil
			case *PyList:
				return MakeInt(int64(len(v.Items))), nil
			case *PyTuple:
				return MakeInt(int64(len(v.Items))), nil
			case *PyDict:
				return MakeInt(int64(len(v.Items))), nil
			case *PySet:
				return MakeInt(int64(len(v.Items))), nil
			case *PyBytes:
				return MakeInt(int64(len(v.Value))), nil
			default:
				return nil, fmt.Errorf("object of type '%s' has no len()", vm.typeName(v))
			}
		},
	}

	vm.builtins["range"] = &PyBuiltinFunc{
		Name: "range",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			var start, stop, step int64 = 0, 0, 1
			switch len(args) {
			case 1:
				stop = vm.toInt(args[0])
			case 2:
				start = vm.toInt(args[0])
				stop = vm.toInt(args[1])
			case 3:
				start = vm.toInt(args[0])
				stop = vm.toInt(args[1])
				step = vm.toInt(args[2])
			default:
				return nil, fmt.Errorf("range expected 1 to 3 arguments, got %d", len(args))
			}
			return &PyRange{Start: start, Stop: stop, Step: step}, nil
		},
	}

	vm.builtins["int"] = &PyBuiltinFunc{
		Name: "int",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return MakeInt(0), nil
			}
			i, err := vm.tryToInt(args[0])
			if err != nil {
				return nil, err
			}
			return MakeInt(i), nil
		},
	}

	vm.builtins["float"] = &PyBuiltinFunc{
		Name: "float",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return &PyFloat{Value: 0.0}, nil
			}
			f, err := vm.tryToFloat(args[0])
			if err != nil {
				return nil, err
			}
			return &PyFloat{Value: f}, nil
		},
	}

	vm.builtins["str"] = &PyBuiltinFunc{
		Name: "str",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return &PyString{Value: ""}, nil
			}
			return &PyString{Value: vm.str(args[0])}, nil
		},
	}

	vm.builtins["bool"] = &PyBuiltinFunc{
		Name: "bool",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return False, nil
			}
			if vm.truthy(args[0]) {
				return True, nil
			}
			return False, nil
		},
	}

	vm.builtins["list"] = &PyBuiltinFunc{
		Name: "list",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return &PyList{Items: []Value{}}, nil
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			return &PyList{Items: items}, nil
		},
	}

	vm.builtins["tuple"] = &PyBuiltinFunc{
		Name: "tuple",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return &PyTuple{Items: []Value{}}, nil
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			return &PyTuple{Items: items}, nil
		},
	}

	vm.builtins["dict"] = &PyBuiltinFunc{
		Name: "dict",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
			for k, v := range kwargs {
				// Use hash-based storage for O(1) lookup
				d.DictSet(&PyString{Value: k}, v, vm)
			}
			return d, nil
		},
	}

	vm.builtins["set"] = &PyBuiltinFunc{
		Name: "set",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			s := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			if len(args) > 0 {
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				for _, item := range items {
					if !isHashable(item) {
						return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(item))
					}
					// Use hash-based storage for O(1) lookup
					s.SetAdd(item, vm)
				}
			}
			return s, nil
		},
	}

	vm.builtins["type"] = &PyBuiltinFunc{
		Name: "type",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("type() takes 1 argument")
			}
			return &PyString{Value: vm.typeName(args[0])}, nil
		},
	}

	vm.builtins["isinstance"] = &PyBuiltinFunc{
		Name: "isinstance",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("isinstance() takes exactly 2 arguments")
			}
			// Simplified implementation
			return True, nil
		},
	}

	vm.builtins["abs"] = &PyBuiltinFunc{
		Name: "abs",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("abs() takes exactly one argument")
			}
			switch v := args[0].(type) {
			case *PyInt:
				if v.Value < 0 {
					return MakeInt(-v.Value), nil
				}
				return v, nil
			case *PyFloat:
				return &PyFloat{Value: math.Abs(v.Value)}, nil
			default:
				return nil, fmt.Errorf("bad operand type for abs()")
			}
		},
	}

	vm.builtins["min"] = &PyBuiltinFunc{
		Name: "min",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("min expected at least 1 argument")
			}
			if len(args) == 1 {
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				args = items
			}
			minVal := args[0]
			for _, v := range args[1:] {
				if vm.compare(v, minVal) < 0 {
					minVal = v
				}
			}
			return minVal, nil
		},
	}

	vm.builtins["max"] = &PyBuiltinFunc{
		Name: "max",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("max expected at least 1 argument")
			}
			if len(args) == 1 {
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				args = items
			}
			maxVal := args[0]
			for _, v := range args[1:] {
				if vm.compare(v, maxVal) > 0 {
					maxVal = v
				}
			}
			return maxVal, nil
		},
	}

	vm.builtins["sum"] = &PyBuiltinFunc{
		Name: "sum",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("sum expected at least 1 argument")
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			var start Value = MakeInt(0)
			if len(args) > 1 {
				start = args[1]
			}
			result := start
			for _, item := range items {
				result, err = vm.binaryOp(OpBinaryAdd, result, item)
				if err != nil {
					return nil, err
				}
			}
			return result, nil
		},
	}

	vm.builtins["input"] = &PyBuiltinFunc{
		Name: "input",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) > 0 {
				fmt.Print(vm.str(args[0]))
			}
			var line string
			fmt.Scanln(&line)
			return &PyString{Value: line}, nil
		},
	}

	vm.builtins["ord"] = &PyBuiltinFunc{
		Name: "ord",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("ord() takes exactly one argument")
			}
			s, ok := args[0].(*PyString)
			if !ok || len(s.Value) != 1 {
				return nil, fmt.Errorf("ord() expected a character")
			}
			return MakeInt(int64(s.Value[0])), nil
		},
	}

	vm.builtins["chr"] = &PyBuiltinFunc{
		Name: "chr",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("chr() takes exactly one argument")
			}
			i := vm.toInt(args[0])
			return &PyString{Value: string(rune(i))}, nil
		},
	}

	// enumerate(iterable, start=0) - returns iterator of (index, value) tuples
	vm.builtins["enumerate"] = &PyBuiltinFunc{
		Name: "enumerate",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("enumerate expected 1 to 2 arguments, got %d", len(args))
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			var start int64 = 0
			if len(args) == 2 {
				start = vm.toInt(args[1])
			}
			if s, ok := kwargs["start"]; ok {
				start = vm.toInt(s)
			}
			result := make([]Value, len(items))
			for i, item := range items {
				result[i] = &PyTuple{Items: []Value{MakeInt(start + int64(i)), item}}
			}
			return &PyIterator{Items: result, Index: 0}, nil
		},
	}

	// zip(*iterables) - returns iterator of tuples
	vm.builtins["zip"] = &PyBuiltinFunc{
		Name: "zip",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return &PyIterator{Items: []Value{}, Index: 0}, nil
			}
			// Convert all args to lists
			lists := make([][]Value, len(args))
			minLen := -1
			for i, arg := range args {
				items, err := vm.toList(arg)
				if err != nil {
					return nil, err
				}
				lists[i] = items
				if minLen == -1 || len(items) < minLen {
					minLen = len(items)
				}
			}
			// Build result tuples
			result := make([]Value, minLen)
			for i := 0; i < minLen; i++ {
				tuple := make([]Value, len(lists))
				for j, list := range lists {
					tuple[j] = list[i]
				}
				result[i] = &PyTuple{Items: tuple}
			}
			return &PyIterator{Items: result, Index: 0}, nil
		},
	}

	// map(func, *iterables) - applies function to items from iterables
	vm.builtins["map"] = &PyBuiltinFunc{
		Name: "map",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("map() must have at least two arguments")
			}
			fn := args[0]
			// Convert remaining args to lists
			lists := make([][]Value, len(args)-1)
			minLen := -1
			for i, arg := range args[1:] {
				items, err := vm.toList(arg)
				if err != nil {
					return nil, err
				}
				lists[i] = items
				if minLen == -1 || len(items) < minLen {
					minLen = len(items)
				}
			}
			// Apply function to each set of items
			result := make([]Value, minLen)
			for i := 0; i < minLen; i++ {
				fnArgs := make([]Value, len(lists))
				for j, list := range lists {
					fnArgs[j] = list[i]
				}
				val, err := vm.call(fn, fnArgs, nil)
				if err != nil {
					return nil, err
				}
				result[i] = val
			}
			return &PyIterator{Items: result, Index: 0}, nil
		},
	}

	// filter(func, iterable) - filters items based on function
	vm.builtins["filter"] = &PyBuiltinFunc{
		Name: "filter",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("filter expected 2 arguments, got %d", len(args))
			}
			fn := args[0]
			items, err := vm.toList(args[1])
			if err != nil {
				return nil, err
			}
			var result []Value
			for _, item := range items {
				var keep bool
				if fn == None {
					// If function is None, filter by truthiness
					keep = vm.truthy(item)
				} else {
					val, err := vm.call(fn, []Value{item}, nil)
					if err != nil {
						return nil, err
					}
					keep = vm.truthy(val)
				}
				if keep {
					result = append(result, item)
				}
			}
			return &PyIterator{Items: result, Index: 0}, nil
		},
	}

	// reversed(seq) - returns reversed iterator
	vm.builtins["reversed"] = &PyBuiltinFunc{
		Name: "reversed",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("reversed() takes exactly one argument (%d given)", len(args))
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			// Create reversed copy
			result := make([]Value, len(items))
			for i, item := range items {
				result[len(items)-1-i] = item
			}
			return &PyIterator{Items: result, Index: 0}, nil
		},
	}

	// sorted(iterable, key=None, reverse=False) - returns sorted list
	vm.builtins["sorted"] = &PyBuiltinFunc{
		Name: "sorted",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("sorted expected 1 argument, got %d", len(args))
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			// Make a copy to avoid modifying the original
			result := make([]Value, len(items))
			copy(result, items)

			// Check for key function
			var keyFn Value
			if k, ok := kwargs["key"]; ok && k != None {
				keyFn = k
			}

			// Check for reverse flag
			reverse := false
			if r, ok := kwargs["reverse"]; ok {
				reverse = vm.truthy(r)
			}

			// Sort using comparison
			var sortErr error
			sort.SliceStable(result, func(i, j int) bool {
				if sortErr != nil {
					return false
				}
				a, b := result[i], result[j]
				// Apply key function if provided
				if keyFn != nil {
					var err error
					a, err = vm.call(keyFn, []Value{a}, nil)
					if err != nil {
						sortErr = err
						return false
					}
					b, err = vm.call(keyFn, []Value{b}, nil)
					if err != nil {
						sortErr = err
						return false
					}
				}
				cmp := vm.compare(a, b)
				if reverse {
					return cmp > 0
				}
				return cmp < 0
			})
			if sortErr != nil {
				return nil, sortErr
			}
			return &PyList{Items: result}, nil
		},
	}

	// all(iterable) - returns True if all elements are truthy
	vm.builtins["all"] = &PyBuiltinFunc{
		Name: "all",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("all() takes exactly one argument (%d given)", len(args))
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			for _, item := range items {
				if !vm.truthy(item) {
					return False, nil
				}
			}
			return True, nil
		},
	}

	// any(iterable) - returns True if any element is truthy
	vm.builtins["any"] = &PyBuiltinFunc{
		Name: "any",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("any() takes exactly one argument (%d given)", len(args))
			}
			items, err := vm.toList(args[0])
			if err != nil {
				return nil, err
			}
			for _, item := range items {
				if vm.truthy(item) {
					return True, nil
				}
			}
			return False, nil
		},
	}

	// hasattr(obj, name) - returns True if object has the attribute
	vm.builtins["hasattr"] = &PyBuiltinFunc{
		Name: "hasattr",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("hasattr() takes exactly 2 arguments (%d given)", len(args))
			}
			name, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("attribute name must be string, not '%s'", vm.typeName(args[1]))
			}
			_, err := vm.getAttr(args[0], name.Value)
			if err != nil {
				return False, nil
			}
			return True, nil
		},
	}

	// getattr(obj, name[, default]) - get attribute from object
	vm.builtins["getattr"] = &PyBuiltinFunc{
		Name: "getattr",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 2 || len(args) > 3 {
				return nil, fmt.Errorf("getattr expected 2 or 3 arguments, got %d", len(args))
			}
			name, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("attribute name must be string, not '%s'", vm.typeName(args[1]))
			}
			val, err := vm.getAttr(args[0], name.Value)
			if err != nil {
				if len(args) == 3 {
					return args[2], nil // Return default
				}
				return nil, fmt.Errorf("'%s' object has no attribute '%s'", vm.typeName(args[0]), name.Value)
			}
			return val, nil
		},
	}

	// setattr(obj, name, value) - set attribute on object
	vm.builtins["setattr"] = &PyBuiltinFunc{
		Name: "setattr",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("setattr() takes exactly 3 arguments (%d given)", len(args))
			}
			name, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("attribute name must be string, not '%s'", vm.typeName(args[1]))
			}
			err := vm.setAttr(args[0], name.Value, args[2])
			if err != nil {
				return nil, err
			}
			return None, nil
		},
	}

	// delattr(obj, name) - delete attribute from object
	vm.builtins["delattr"] = &PyBuiltinFunc{
		Name: "delattr",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("delattr() takes exactly 2 arguments (%d given)", len(args))
			}
			name, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("attribute name must be string, not '%s'", vm.typeName(args[1]))
			}
			// Handle deletion based on object type
			switch o := args[0].(type) {
			case *PyInstance:
				if _, exists := o.Dict[name.Value]; exists {
					delete(o.Dict, name.Value)
					return None, nil
				}
				return nil, fmt.Errorf("'%s' object has no attribute '%s'", o.Class.Name, name.Value)
			case *PyModule:
				if _, exists := o.Dict[name.Value]; exists {
					delete(o.Dict, name.Value)
					return None, nil
				}
				return nil, fmt.Errorf("module '%s' has no attribute '%s'", o.Name, name.Value)
			case *PyDict:
				// Allow delattr on dict for dynamic attribute-style access
				for k := range o.Items {
					if str, ok := k.(*PyString); ok && str.Value == name.Value {
						delete(o.Items, k)
						return None, nil
					}
				}
				return nil, fmt.Errorf("'dict' object has no attribute '%s'", name.Value)
			default:
				return nil, fmt.Errorf("'%s' object has no attribute '%s'", vm.typeName(args[0]), name.Value)
			}
		},
	}

	// pow(base, exp[, mod]) - power with optional modulo
	vm.builtins["pow"] = &PyBuiltinFunc{
		Name: "pow",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 2 || len(args) > 3 {
				return nil, fmt.Errorf("pow expected 2 or 3 arguments, got %d", len(args))
			}
			base := vm.toFloat(args[0])
			exp := vm.toFloat(args[1])

			if len(args) == 3 {
				// Modular exponentiation - requires integers
				baseInt := vm.toInt(args[0])
				expInt := vm.toInt(args[1])
				modInt := vm.toInt(args[2])
				if modInt == 0 {
					return nil, fmt.Errorf("pow() 3rd argument cannot be 0")
				}
				// Simple modular exponentiation
				result := int64(1)
				baseInt = baseInt % modInt
				for expInt > 0 {
					if expInt%2 == 1 {
						result = (result * baseInt) % modInt
					}
					expInt = expInt / 2
					baseInt = (baseInt * baseInt) % modInt
				}
				return MakeInt(result), nil
			}

			result := math.Pow(base, exp)
			// Return int if result is a whole number and inputs were ints
			_, baseIsInt := args[0].(*PyInt)
			_, expIsInt := args[1].(*PyInt)
			if baseIsInt && expIsInt && result == float64(int64(result)) {
				return MakeInt(int64(result)), nil
			}
			return &PyFloat{Value: result}, nil
		},
	}

	// divmod(a, b) - returns (quotient, remainder)
	vm.builtins["divmod"] = &PyBuiltinFunc{
		Name: "divmod",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("divmod expected 2 arguments, got %d", len(args))
			}
			// Check if both are integers
			aInt, aIsInt := args[0].(*PyInt)
			bInt, bIsInt := args[1].(*PyInt)
			if aIsInt && bIsInt {
				if bInt.Value == 0 {
					return nil, fmt.Errorf("integer division or modulo by zero")
				}
				q := aInt.Value / bInt.Value
				r := aInt.Value % bInt.Value
				// Python's divmod uses floor division
				if r != 0 && (aInt.Value < 0) != (bInt.Value < 0) {
					q--
					r += bInt.Value
				}
				return &PyTuple{Items: []Value{MakeInt(q), MakeInt(r)}}, nil
			}
			// Float division
			a := vm.toFloat(args[0])
			b := vm.toFloat(args[1])
			if b == 0 {
				return nil, fmt.Errorf("float division by zero")
			}
			q := math.Floor(a / b)
			r := a - q*b
			return &PyTuple{Items: []Value{&PyFloat{Value: q}, &PyFloat{Value: r}}}, nil
		},
	}

	// hex(x) - convert integer to hexadecimal string
	vm.builtins["hex"] = &PyBuiltinFunc{
		Name: "hex",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("hex() takes exactly one argument (%d given)", len(args))
			}
			n := vm.toInt(args[0])
			if n < 0 {
				return &PyString{Value: fmt.Sprintf("-0x%x", -n)}, nil
			}
			return &PyString{Value: fmt.Sprintf("0x%x", n)}, nil
		},
	}

	// oct(x) - convert integer to octal string
	vm.builtins["oct"] = &PyBuiltinFunc{
		Name: "oct",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("oct() takes exactly one argument (%d given)", len(args))
			}
			n := vm.toInt(args[0])
			if n < 0 {
				return &PyString{Value: fmt.Sprintf("-0o%o", -n)}, nil
			}
			return &PyString{Value: fmt.Sprintf("0o%o", n)}, nil
		},
	}

	// bin(x) - convert integer to binary string
	vm.builtins["bin"] = &PyBuiltinFunc{
		Name: "bin",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("bin() takes exactly one argument (%d given)", len(args))
			}
			n := vm.toInt(args[0])
			if n < 0 {
				return &PyString{Value: fmt.Sprintf("-0b%b", -n)}, nil
			}
			return &PyString{Value: fmt.Sprintf("0b%b", n)}, nil
		},
	}

	// round(number[, ndigits]) - round to given precision
	vm.builtins["round"] = &PyBuiltinFunc{
		Name: "round",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("round() takes 1 or 2 arguments (%d given)", len(args))
			}
			num := vm.toFloat(args[0])

			if len(args) == 1 {
				// Round to integer - use banker's rounding (round half to even)
				rounded := math.RoundToEven(num)
				return MakeInt(int64(rounded)), nil
			}

			// Round to ndigits decimal places
			ndigits := vm.toInt(args[1])
			multiplier := math.Pow(10, float64(ndigits))
			rounded := math.RoundToEven(num*multiplier) / multiplier

			// If ndigits is negative, return int
			if ndigits < 0 {
				return MakeInt(int64(rounded)), nil
			}
			return &PyFloat{Value: rounded}, nil
		},
	}

	// callable(obj) - check if object is callable
	vm.builtins["callable"] = &PyBuiltinFunc{
		Name: "callable",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("callable() takes exactly one argument (%d given)", len(args))
			}
			switch args[0].(type) {
			case *PyFunction, *PyBuiltinFunc, *PyGoFunc, *PyMethod, *PyClass:
				return True, nil
			case *PyInstance:
				// Check if instance has __call__ method
				inst := args[0].(*PyInstance)
				for _, cls := range inst.Class.Mro {
					if _, ok := cls.Dict["__call__"]; ok {
						return True, nil
					}
				}
				return False, nil
			default:
				return False, nil
			}
		},
	}

	vm.builtins["None"] = None
	vm.builtins["True"] = True
	vm.builtins["False"] = False

	// __build_class__ is used to create classes
	vm.builtins["__build_class__"] = &PyBuiltinFunc{
		Name: "__build_class__",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("__build_class__: not enough arguments")
			}

			// First arg is the class body function
			bodyFunc, ok := args[0].(*PyFunction)
			if !ok {
				return nil, fmt.Errorf("__build_class__: first argument must be a function")
			}

			// Second arg is the class name
			nameVal, ok := args[1].(*PyString)
			if !ok {
				return nil, fmt.Errorf("__build_class__: second argument must be a string")
			}
			className := nameVal.Value

			// Remaining args are base classes
			var bases []*PyClass
			for i := 2; i < len(args); i++ {
				if base, ok := args[i].(*PyClass); ok {
					bases = append(bases, base)
				}
			}

			// If no bases specified, implicitly inherit from object (Python 3 behavior)
			objectClass := vm.builtins["object"].(*PyClass)
			if len(bases) == 0 {
				bases = []*PyClass{objectClass}
			}

			// Execute the class body to get the namespace
			classDict, err := vm.callClassBody(bodyFunc)
			if err != nil {
				return nil, fmt.Errorf("__build_class__: error executing class body: %w", err)
			}

			// Create the class
			class := &PyClass{
				Name:  className,
				Bases: bases,
				Dict:  classDict,
			}

			// Build MRO using C3 linearization for proper multiple inheritance
			mro, err := vm.computeC3MRO(class, bases)
			if err != nil {
				return nil, err
			}
			class.Mro = mro

			return class, nil
		},
	}

	// property() creates a property descriptor
	vm.builtins["property"] = &PyBuiltinFunc{
		Name: "property",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			prop := &PyProperty{}

			// Handle positional args: property(fget=None, fset=None, fdel=None, doc=None)
			if len(args) > 0 && args[0] != None {
				prop.Fget = args[0]
			}
			if len(args) > 1 && args[1] != None {
				prop.Fset = args[1]
			}
			if len(args) > 2 && args[2] != None {
				prop.Fdel = args[2]
			}
			if len(args) > 3 {
				if s, ok := args[3].(*PyString); ok {
					prop.Doc = s.Value
				}
			}

			// Handle keyword args
			if fget, ok := kwargs["fget"]; ok && fget != None {
				prop.Fget = fget
			}
			if fset, ok := kwargs["fset"]; ok && fset != None {
				prop.Fset = fset
			}
			if fdel, ok := kwargs["fdel"]; ok && fdel != None {
				prop.Fdel = fdel
			}
			if doc, ok := kwargs["doc"]; ok {
				if s, ok := doc.(*PyString); ok {
					prop.Doc = s.Value
				}
			}

			return prop, nil
		},
	}

	// classmethod() wraps a function to bind the class as first argument
	vm.builtins["classmethod"] = &PyBuiltinFunc{
		Name: "classmethod",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("classmethod expected 1 argument, got %d", len(args))
			}
			return &PyClassMethod{Func: args[0]}, nil
		},
	}

	// staticmethod() wraps a function to prevent binding
	vm.builtins["staticmethod"] = &PyBuiltinFunc{
		Name: "staticmethod",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("staticmethod expected 1 argument, got %d", len(args))
			}
			return &PyStaticMethod{Func: args[0]}, nil
		},
	}

	// super() returns a proxy object for MRO-based method lookup
	vm.builtins["super"] = &PyBuiltinFunc{
		Name: "super",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			var thisClass *PyClass
			var instance Value

			if len(args) == 0 {
				// Zero-argument form: super()
				// Need to find __class__ from the enclosing scope and self from first arg
				// Look up the call stack to find the method context
				// We need to look at the caller's frame, not the current one (which is the builtin call)
				callerFrame := vm.frame
				if len(vm.frames) >= 1 {
					// The frames stack contains previous frames; vm.frame is current
					// For builtin calls, we want the calling frame
					callerFrame = vm.frame
				}
				if callerFrame != nil && callerFrame.Code != nil {
					// Try to get __class__ from closure cells
					for i, name := range callerFrame.Code.FreeVars {
						if name == "__class__" && i < len(callerFrame.Cells) {
							if cls, ok := callerFrame.Cells[i].Value.(*PyClass); ok {
								thisClass = cls
							}
						}
					}
					// Try to get self from first local variable (slot 0)
					if len(callerFrame.Code.VarNames) > 0 && len(callerFrame.Locals) > 0 {
						instance = callerFrame.Locals[0]
					}
				}
				if thisClass == nil {
					return nil, fmt.Errorf("super(): __class__ cell not found")
				}
				if instance == nil {
					return nil, fmt.Errorf("super(): self argument not found")
				}
			} else if len(args) == 2 {
				// Two-argument form: super(type, object-or-type)
				var ok bool
				thisClass, ok = args[0].(*PyClass)
				if !ok {
					return nil, fmt.Errorf("super() argument 1 must be type, not %s", vm.typeName(args[0]))
				}
				instance = args[1]
			} else if len(args) == 1 {
				// One-argument form: super(type) - unbound super
				var ok bool
				thisClass, ok = args[0].(*PyClass)
				if !ok {
					return nil, fmt.Errorf("super() argument 1 must be type, not %s", vm.typeName(args[0]))
				}
				instance = nil
			} else {
				return nil, fmt.Errorf("super() takes 0, 1, or 2 arguments (%d given)", len(args))
			}

			// Find the index of thisClass in the MRO of the instance's class
			var mro []*PyClass
			if inst, ok := instance.(*PyInstance); ok {
				mro = inst.Class.Mro
			} else if cls, ok := instance.(*PyClass); ok {
				mro = cls.Mro
			} else if instance != nil {
				return nil, fmt.Errorf("super(type, obj): obj must be an instance or subtype of type")
			}

			startIdx := 0
			if mro != nil {
				for i, cls := range mro {
					if cls == thisClass {
						startIdx = i + 1 // Start searching from the next class in MRO
						break
					}
				}
			}

			return &PySuper{
				ThisClass: thisClass,
				Instance:  instance,
				StartIdx:  startIdx,
			}, nil
		},
	}

	// object is the base class for all classes
	vm.builtins["object"] = &PyClass{
		Name:  "object",
		Bases: nil,
		Dict:  make(map[string]Value),
		Mro:   nil,
	}
	// Set object's MRO to just itself
	vm.builtins["object"].(*PyClass).Mro = []*PyClass{vm.builtins["object"].(*PyClass)}

	// Initialize exception class hierarchy
	vm.initExceptionClasses()
}

// initExceptionClasses sets up the exception class hierarchy
func (vm *VM) initExceptionClasses() {
	// BaseException is the root of all exceptions
	baseException := &PyClass{
		Name:  "BaseException",
		Bases: nil,
		Dict:  make(map[string]Value),
	}
	baseException.Mro = []*PyClass{baseException}
	vm.builtins["BaseException"] = baseException

	// Exception inherits from BaseException (most exceptions derive from this)
	exception := &PyClass{
		Name:  "Exception",
		Bases: []*PyClass{baseException},
		Dict:  make(map[string]Value),
	}
	exception.Mro = []*PyClass{exception, baseException}
	vm.builtins["Exception"] = exception

	// Helper to create exception class inheriting from Exception
	makeExc := func(name string, parent *PyClass) *PyClass {
		cls := &PyClass{
			Name:  name,
			Bases: []*PyClass{parent},
			Dict:  make(map[string]Value),
		}
		// Build MRO by prepending self to parent's MRO
		cls.Mro = append([]*PyClass{cls}, parent.Mro...)
		vm.builtins[name] = cls
		return cls
	}

	// Standard exceptions inheriting from Exception
	makeExc("ValueError", exception)
	makeExc("TypeError", exception)
	makeExc("KeyError", exception)
	makeExc("IndexError", exception)
	makeExc("AttributeError", exception)
	makeExc("NameError", exception)
	makeExc("RuntimeError", exception)
	makeExc("ZeroDivisionError", exception)
	makeExc("AssertionError", exception)
	makeExc("StopIteration", exception)
	makeExc("NotImplementedError", exception)
	makeExc("RecursionError", exception)

	// OSError and its subclasses
	osError := makeExc("OSError", exception)
	makeExc("FileNotFoundError", osError)
	makeExc("PermissionError", osError)
	makeExc("FileExistsError", osError)
	makeExc("IOError", osError) // IOError is an alias for OSError in Python 3

	// ImportError and its subclass
	importError := makeExc("ImportError", exception)
	makeExc("ModuleNotFoundError", importError)

	// LookupError (base for KeyError and IndexError - for compatibility)
	lookupError := makeExc("LookupError", exception)
	_ = lookupError // We already created KeyError and IndexError above

	// ArithmeticError (base for ZeroDivisionError - for compatibility)
	arithmeticError := makeExc("ArithmeticError", exception)
	_ = arithmeticError // We already created ZeroDivisionError above
}

// computeC3MRO computes the Method Resolution Order using C3 linearization algorithm.
// This properly handles multiple inheritance and detects inconsistent hierarchies.
func (vm *VM) computeC3MRO(class *PyClass, bases []*PyClass) ([]*PyClass, error) {
	// Base case: no bases
	if len(bases) == 0 {
		return []*PyClass{class}, nil
	}

	// Collect linearizations of all bases plus the list of bases itself
	// We need to copy slices to avoid modifying the originals
	var toMerge [][]*PyClass
	for _, base := range bases {
		// Copy the base's MRO
		baseMRO := make([]*PyClass, len(base.Mro))
		copy(baseMRO, base.Mro)
		toMerge = append(toMerge, baseMRO)
	}
	// Add the list of direct bases
	basesCopy := make([]*PyClass, len(bases))
	copy(basesCopy, bases)
	toMerge = append(toMerge, basesCopy)

	// Start with the class itself
	result := []*PyClass{class}

	// Merge until all lists are empty
	for {
		// Remove empty lists
		var nonEmpty [][]*PyClass
		for _, list := range toMerge {
			if len(list) > 0 {
				nonEmpty = append(nonEmpty, list)
			}
		}
		toMerge = nonEmpty

		if len(toMerge) == 0 {
			break
		}

		// Find a good head: a class that is not in the tail of any list
		var candidate *PyClass
		for _, list := range toMerge {
			head := list[0]
			inTail := false
			for _, other := range toMerge {
				// Check if head appears in the tail (positions 1+) of other
				for i := 1; i < len(other); i++ {
					if other[i] == head {
						inTail = true
						break
					}
				}
				if inTail {
					break
				}
			}
			if !inTail {
				candidate = head
				break
			}
		}

		if candidate == nil {
			// No valid candidate found - inconsistent hierarchy
			msg := fmt.Sprintf("Cannot create a consistent method resolution order (MRO) for bases %s",
				vm.formatBases(bases))
			return nil, &PyException{
				ExcType:  vm.builtins["TypeError"].(*PyClass),
				Args:     &PyTuple{Items: []Value{&PyString{Value: msg}}},
				Message:  "TypeError: " + msg,
				TypeName: "TypeError",
			}
		}

		// Add candidate to result
		result = append(result, candidate)

		// Remove candidate from the head of all lists where it appears
		for i := range toMerge {
			if len(toMerge[i]) > 0 && toMerge[i][0] == candidate {
				toMerge[i] = toMerge[i][1:]
			}
		}
	}

	return result, nil
}

// formatBases formats a list of base classes for error messages
func (vm *VM) formatBases(bases []*PyClass) string {
	if len(bases) == 0 {
		return ""
	}
	names := make([]string, len(bases))
	for i, b := range bases {
		names[i] = b.Name
	}
	result := names[0]
	for i := 1; i < len(names); i++ {
		result += ", " + names[i]
	}
	return result
}

// PyRange represents a range object
type PyRange struct {
	Start, Stop, Step int64
}

func (r *PyRange) Type() string   { return "range" }
func (r *PyRange) String() string { return fmt.Sprintf("range(%d, %d, %d)", r.Start, r.Stop, r.Step) }

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

// Execute runs bytecode and returns the result
func (vm *VM) Execute(code *CodeObject) (Value, error) {
	return vm.ExecuteWithContext(context.Background(), code)
}

// ExecuteWithTimeout runs bytecode with a time limit.
// Returns TimeoutError if the script exceeds the specified duration.
func (vm *VM) ExecuteWithTimeout(timeout time.Duration, code *CodeObject) (Value, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return vm.ExecuteWithContext(ctx, code)
}

// ExecuteWithContext runs bytecode with a context for cancellation/timeout support.
// The context is checked periodically during execution (see SetCheckInterval).
// Returns CancelledError if the context is cancelled, or TimeoutError if it times out.
func (vm *VM) ExecuteWithContext(ctx context.Context, code *CodeObject) (Value, error) {
	frame := &Frame{
		Code:     code,
		IP:       0,
		Stack:    make([]Value, code.StackSize+16), // Pre-allocate with small buffer
		SP:       0,
		Locals:   make([]Value, len(code.VarNames)),
		Globals:  vm.Globals,
		Builtins: vm.builtins,
	}

	vm.frames = append(vm.frames, frame)
	vm.frame = frame
	vm.ctx = ctx
	vm.checkCounter = vm.checkInterval // Reset counter for new execution

	return vm.run()
}

func (vm *VM) run() (Value, error) {
	frame := vm.frame

	for frame.IP < len(frame.Code.Code) {
		// Check for timeout/cancellation periodically (counter decrement is faster than modulo)
		if vm.ctx != nil {
			vm.checkCounter--
			if vm.checkCounter <= 0 {
				vm.checkCounter = vm.checkInterval
				select {
				case <-vm.ctx.Done():
					if vm.ctx.Err() == context.DeadlineExceeded {
						// Extract timeout duration from context if possible
						if deadline, ok := vm.ctx.Deadline(); ok {
							return nil, &TimeoutError{Timeout: time.Until(deadline) * -1}
						}
						return nil, &TimeoutError{}
					}
					return nil, &CancelledError{}
				default:
					// Context not done, continue execution
				}
			}
		}

		op := Opcode(frame.Code.Code[frame.IP])
		frame.IP++

		var arg int
		if op.HasArg() {
			arg = int(frame.Code.Code[frame.IP]) | int(frame.Code.Code[frame.IP+1])<<8
			frame.IP += 2
		}

		var err error
		switch op {
		case OpPop:
			vm.pop()

		case OpDup:
			vm.push(vm.top())

		case OpRot2:
			a := vm.pop()
			b := vm.pop()
			vm.push(a)
			vm.push(b)

		case OpRot3:
			a := vm.pop()
			b := vm.pop()
			c := vm.pop()
			vm.push(a)
			vm.push(c)
			vm.push(b)

		case OpLoadConst:
			// Inline push for constant load - grow stack if needed
			if frame.SP >= len(frame.Stack) {
				vm.ensureStack(1)
			}
			frame.Stack[frame.SP] = vm.toValue(frame.Code.Constants[arg])
			frame.SP++

		case OpLoadName:
			name := frame.Code.Names[arg]
			if val, ok := frame.Globals[name]; ok {
				vm.push(val)
			} else if frame.EnclosingGlobals != nil {
				if val, ok := frame.EnclosingGlobals[name]; ok {
					vm.push(val)
				} else if val, ok := frame.Builtins[name]; ok {
					vm.push(val)
				} else {
					return nil, fmt.Errorf("name '%s' is not defined", name)
				}
			} else if val, ok := frame.Builtins[name]; ok {
				vm.push(val)
			} else {
				return nil, fmt.Errorf("name '%s' is not defined", name)
			}

		case OpStoreName:
			name := frame.Code.Names[arg]
			frame.Globals[name] = vm.pop()

		case OpDeleteName:
			name := frame.Code.Names[arg]
			delete(frame.Globals, name)

		case OpLoadFast:
			// Inline push for local variable load
			frame.Stack[frame.SP] = frame.Locals[arg]
			frame.SP++

		case OpStoreFast:
			// Inline pop for local variable store
			frame.SP--
			frame.Locals[arg] = frame.Stack[frame.SP]

		case OpDeleteFast:
			frame.Locals[arg] = nil

		case OpLoadDeref:
			// Load from closure cell
			// arg indexes into cells: first CellVars, then FreeVars
			if arg < len(frame.Cells) {
				cell := frame.Cells[arg]
				if cell != nil {
					vm.push(cell.Value)
				} else {
					return nil, fmt.Errorf("cell is nil at index %d", arg)
				}
			} else {
				return nil, fmt.Errorf("cell index %d out of range (have %d cells)", arg, len(frame.Cells))
			}

		case OpStoreDeref:
			// Store to closure cell
			val := vm.pop()
			if arg < len(frame.Cells) {
				if frame.Cells[arg] == nil {
					frame.Cells[arg] = &PyCell{}
				}
				frame.Cells[arg].Value = val
			} else {
				return nil, fmt.Errorf("cell index %d out of range for store (have %d cells)", arg, len(frame.Cells))
			}

		case OpLoadClosure:
			// Load a cell to build a closure tuple
			// arg indexes into cells (CellVars first, then FreeVars)
			if arg < len(frame.Cells) {
				vm.push(frame.Cells[arg])
			} else {
				return nil, fmt.Errorf("closure cell index %d out of range", arg)
			}

		// ==========================================
		// Specialized opcodes (no argument fetch needed)
		// ==========================================

		case OpLoadFast0:
			frame.Stack[frame.SP] = frame.Locals[0]
			frame.SP++

		case OpLoadFast1:
			frame.Stack[frame.SP] = frame.Locals[1]
			frame.SP++

		case OpLoadFast2:
			frame.Stack[frame.SP] = frame.Locals[2]
			frame.SP++

		case OpLoadFast3:
			frame.Stack[frame.SP] = frame.Locals[3]
			frame.SP++

		case OpStoreFast0:
			frame.SP--
			frame.Locals[0] = frame.Stack[frame.SP]

		case OpStoreFast1:
			frame.SP--
			frame.Locals[1] = frame.Stack[frame.SP]

		case OpStoreFast2:
			frame.SP--
			frame.Locals[2] = frame.Stack[frame.SP]

		case OpStoreFast3:
			frame.SP--
			frame.Locals[3] = frame.Stack[frame.SP]

		case OpLoadNone:
			frame.Stack[frame.SP] = None
			frame.SP++

		case OpLoadTrue:
			frame.Stack[frame.SP] = True
			frame.SP++

		case OpLoadFalse:
			frame.Stack[frame.SP] = False
			frame.SP++

		case OpLoadZero:
			frame.Stack[frame.SP] = MakeInt(0)
			frame.SP++

		case OpLoadOne:
			frame.Stack[frame.SP] = MakeInt(1)
			frame.SP++

		case OpIncrementFast:
			// Increment local variable by 1
			if v, ok := frame.Locals[arg].(*PyInt); ok {
				frame.Locals[arg] = MakeInt(v.Value + 1)
			} else {
				// Fallback for non-int
				result, err := vm.binaryOp(OpBinaryAdd, frame.Locals[arg], MakeInt(1))
				if err != nil {
					return nil, err
				}
				frame.Locals[arg] = result
			}

		case OpDecrementFast:
			// Decrement local variable by 1
			if v, ok := frame.Locals[arg].(*PyInt); ok {
				frame.Locals[arg] = MakeInt(v.Value - 1)
			} else {
				result, err := vm.binaryOp(OpBinarySubtract, frame.Locals[arg], MakeInt(1))
				if err != nil {
					return nil, err
				}
				frame.Locals[arg] = result
			}

		case OpNegateFast:
			// Negate local variable in place: sign = -sign
			if v, ok := frame.Locals[arg].(*PyInt); ok {
				frame.Locals[arg] = MakeInt(-v.Value)
			} else if v, ok := frame.Locals[arg].(*PyFloat); ok {
				frame.Locals[arg] = &PyFloat{Value: -v.Value}
			} else {
				// Fallback
				result, err := vm.unaryOp(OpUnaryNegative, frame.Locals[arg])
				if err != nil {
					return nil, err
				}
				frame.Locals[arg] = result
			}

		case OpAddConstFast:
			// Add constant to local: x = x + const
			// arg contains packed indices: low byte = local index, high byte = const index
			localIdx := arg & 0xFF
			constIdx := (arg >> 8) & 0xFF
			constVal := vm.toValue(frame.Code.Constants[constIdx])
			localVal := frame.Locals[localIdx]
			if li, ok := localVal.(*PyInt); ok {
				if ci, ok := constVal.(*PyInt); ok {
					frame.Locals[localIdx] = MakeInt(li.Value + ci.Value)
					break
				}
			}
			// Fallback
			result, err := vm.binaryOp(OpBinaryAdd, localVal, constVal)
			if err != nil {
				return nil, err
			}
			frame.Locals[localIdx] = result

		case OpAccumulateFast:
			// Accumulate TOS to local: local[arg] = local[arg] + TOS
			frame.SP--
			addend := frame.Stack[frame.SP]
			localVal := frame.Locals[arg]
			// Fast path for float accumulation (common in numerical code)
			if lf, ok := localVal.(*PyFloat); ok {
				if af, ok := addend.(*PyFloat); ok {
					frame.Locals[arg] = &PyFloat{Value: lf.Value + af.Value}
					break
				}
				if ai, ok := addend.(*PyInt); ok {
					frame.Locals[arg] = &PyFloat{Value: lf.Value + float64(ai.Value)}
					break
				}
			}
			// Fast path for int accumulation
			if li, ok := localVal.(*PyInt); ok {
				if ai, ok := addend.(*PyInt); ok {
					frame.Locals[arg] = MakeInt(li.Value + ai.Value)
					break
				}
				if af, ok := addend.(*PyFloat); ok {
					frame.Locals[arg] = &PyFloat{Value: float64(li.Value) + af.Value}
					break
				}
			}
			// Fallback
			result, err := vm.binaryOp(OpBinaryAdd, localVal, addend)
			if err != nil {
				return nil, err
			}
			frame.Locals[arg] = result

		case OpLoadFastLoadFast:
			// Load two locals: arg contains packed indices (low byte = first, high byte = second)
			idx1 := arg & 0xFF
			idx2 := (arg >> 8) & 0xFF
			vm.ensureStack(2) // Ensure space for two pushes
			frame.Stack[frame.SP] = frame.Locals[idx1]
			frame.SP++
			frame.Stack[frame.SP] = frame.Locals[idx2]
			frame.SP++

		case OpLoadFastLoadConst:
			// Load local then const: arg contains packed indices
			localIdx := arg & 0xFF
			constIdx := (arg >> 8) & 0xFF
			vm.ensureStack(2) // Ensure space for two pushes
			frame.Stack[frame.SP] = frame.Locals[localIdx]
			frame.SP++
			frame.Stack[frame.SP] = vm.toValue(frame.Code.Constants[constIdx])
			frame.SP++

		case OpStoreFastLoadFast:
			// Store to local then load another: arg contains packed indices
			storeIdx := arg & 0xFF
			loadIdx := (arg >> 8) & 0xFF
			frame.SP--
			frame.Locals[storeIdx] = frame.Stack[frame.SP]
			frame.Stack[frame.SP] = frame.Locals[loadIdx]
			frame.SP++

		case OpBinaryAddInt:
			// Optimized int + int (assumes both are ints, falls back if not)
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					frame.Stack[frame.SP] = MakeInt(ai.Value + bi.Value)
					frame.SP++
					break
				}
			}
			// Fallback
			result, err := vm.binaryOp(OpBinaryAdd, a, b)
			if err != nil {
				return nil, err
			}
			frame.Stack[frame.SP] = result
			frame.SP++

		case OpBinarySubtractInt:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					frame.Stack[frame.SP] = MakeInt(ai.Value - bi.Value)
					frame.SP++
					break
				}
			}
			result, err := vm.binaryOp(OpBinarySubtract, a, b)
			if err != nil {
				return nil, err
			}
			frame.Stack[frame.SP] = result
			frame.SP++

		case OpBinaryMultiplyInt:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					frame.Stack[frame.SP] = MakeInt(ai.Value * bi.Value)
					frame.SP++
					break
				}
			}
			result, err := vm.binaryOp(OpBinaryMultiply, a, b)
			if err != nil {
				return nil, err
			}
			frame.Stack[frame.SP] = result
			frame.SP++

		case OpBinaryDivideFloat:
			// Optimized true division (always returns float)
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			// Fast path for int/int division
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					if bi.Value == 0 {
						return nil, fmt.Errorf("division by zero")
					}
					frame.Stack[frame.SP] = &PyFloat{Value: float64(ai.Value) / float64(bi.Value)}
					frame.SP++
					break
				}
				if bf, ok := b.(*PyFloat); ok {
					if bf.Value == 0 {
						return nil, fmt.Errorf("division by zero")
					}
					frame.Stack[frame.SP] = &PyFloat{Value: float64(ai.Value) / bf.Value}
					frame.SP++
					break
				}
			}
			if af, ok := a.(*PyFloat); ok {
				if bi, ok := b.(*PyInt); ok {
					if bi.Value == 0 {
						return nil, fmt.Errorf("division by zero")
					}
					frame.Stack[frame.SP] = &PyFloat{Value: af.Value / float64(bi.Value)}
					frame.SP++
					break
				}
				if bf, ok := b.(*PyFloat); ok {
					if bf.Value == 0 {
						return nil, fmt.Errorf("division by zero")
					}
					frame.Stack[frame.SP] = &PyFloat{Value: af.Value / bf.Value}
					frame.SP++
					break
				}
			}
			// Fallback
			result, err := vm.binaryOp(OpBinaryDivide, a, b)
			if err != nil {
				return nil, err
			}
			frame.Stack[frame.SP] = result
			frame.SP++

		case OpBinaryAddFloat:
			// Optimized float addition
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if af, ok := a.(*PyFloat); ok {
				if bf, ok := b.(*PyFloat); ok {
					frame.Stack[frame.SP] = &PyFloat{Value: af.Value + bf.Value}
					frame.SP++
					break
				}
				if bi, ok := b.(*PyInt); ok {
					frame.Stack[frame.SP] = &PyFloat{Value: af.Value + float64(bi.Value)}
					frame.SP++
					break
				}
			}
			if ai, ok := a.(*PyInt); ok {
				if bf, ok := b.(*PyFloat); ok {
					frame.Stack[frame.SP] = &PyFloat{Value: float64(ai.Value) + bf.Value}
					frame.SP++
					break
				}
			}
			// Fallback
			result, err := vm.binaryOp(OpBinaryAdd, a, b)
			if err != nil {
				return nil, err
			}
			frame.Stack[frame.SP] = result
			frame.SP++

		case OpCompareLtInt:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					if ai.Value < bi.Value {
						frame.Stack[frame.SP] = True
					} else {
						frame.Stack[frame.SP] = False
					}
					frame.SP++
					break
				}
			}
			frame.Stack[frame.SP] = vm.compareOp(OpCompareLt, a, b)
			frame.SP++

		case OpCompareLeInt:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					if ai.Value <= bi.Value {
						frame.Stack[frame.SP] = True
					} else {
						frame.Stack[frame.SP] = False
					}
					frame.SP++
					break
				}
			}
			frame.Stack[frame.SP] = vm.compareOp(OpCompareLe, a, b)
			frame.SP++

		case OpCompareGtInt:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					if ai.Value > bi.Value {
						frame.Stack[frame.SP] = True
					} else {
						frame.Stack[frame.SP] = False
					}
					frame.SP++
					break
				}
			}
			frame.Stack[frame.SP] = vm.compareOp(OpCompareGt, a, b)
			frame.SP++

		case OpCompareGeInt:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					if ai.Value >= bi.Value {
						frame.Stack[frame.SP] = True
					} else {
						frame.Stack[frame.SP] = False
					}
					frame.SP++
					break
				}
			}
			frame.Stack[frame.SP] = vm.compareOp(OpCompareGe, a, b)
			frame.SP++

		case OpCompareEqInt:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					if ai.Value == bi.Value {
						frame.Stack[frame.SP] = True
					} else {
						frame.Stack[frame.SP] = False
					}
					frame.SP++
					break
				}
			}
			frame.Stack[frame.SP] = vm.compareOp(OpCompareEq, a, b)
			frame.SP++

		case OpCompareNeInt:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					if ai.Value != bi.Value {
						frame.Stack[frame.SP] = True
					} else {
						frame.Stack[frame.SP] = False
					}
					frame.SP++
					break
				}
			}
			frame.Stack[frame.SP] = vm.compareOp(OpCompareNe, a, b)
			frame.SP++

		// ==========================================
		// Empty collection opcodes
		// ==========================================

		case OpLoadEmptyList:
			frame.Stack[frame.SP] = &PyList{Items: []Value{}}
			frame.SP++

		case OpLoadEmptyTuple:
			frame.Stack[frame.SP] = &PyTuple{Items: []Value{}}
			frame.SP++

		case OpLoadEmptyDict:
			frame.Stack[frame.SP] = &PyDict{Items: make(map[Value]Value)}
			frame.SP++

		// ==========================================
		// Combined compare+jump opcodes
		// ==========================================

		case OpCompareLtJump:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			result := false
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					result = ai.Value < bi.Value
				} else {
					result = vm.truthy(vm.compareOp(OpCompareLt, a, b))
				}
			} else {
				result = vm.truthy(vm.compareOp(OpCompareLt, a, b))
			}
			if !result {
				frame.IP = arg
			}

		case OpCompareLeJump:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			result := false
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					result = ai.Value <= bi.Value
				} else {
					result = vm.truthy(vm.compareOp(OpCompareLe, a, b))
				}
			} else {
				result = vm.truthy(vm.compareOp(OpCompareLe, a, b))
			}
			if !result {
				frame.IP = arg
			}

		case OpCompareGtJump:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			result := false
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					result = ai.Value > bi.Value
				} else {
					result = vm.truthy(vm.compareOp(OpCompareGt, a, b))
				}
			} else {
				result = vm.truthy(vm.compareOp(OpCompareGt, a, b))
			}
			if !result {
				frame.IP = arg
			}

		case OpCompareGeJump:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			result := false
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					result = ai.Value >= bi.Value
				} else {
					result = vm.truthy(vm.compareOp(OpCompareGe, a, b))
				}
			} else {
				result = vm.truthy(vm.compareOp(OpCompareGe, a, b))
			}
			if !result {
				frame.IP = arg
			}

		case OpCompareEqJump:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			result := vm.equal(a, b)
			if !result {
				frame.IP = arg
			}

		case OpCompareNeJump:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			result := !vm.equal(a, b)
			if !result {
				frame.IP = arg
			}

		case OpCompareLtLocalJump:
			// Ultra-optimized: compare two locals and jump if false
			// arg format: bits 0-7 = local1, bits 8-15 = local2, bits 16+ = jump offset
			local1 := arg & 0xFF
			local2 := (arg >> 8) & 0xFF
			jumpOffset := arg >> 16
			a := frame.Locals[local1]
			b := frame.Locals[local2]
			// Fast path for ints
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					if ai.Value >= bi.Value {
						frame.IP = jumpOffset
					}
					break
				}
			}
			// Fallback to generic comparison
			cmp := vm.compareOp(OpCompareLt, a, b)
			if cmp == False || cmp == nil {
				frame.IP = jumpOffset
			}

		// ==========================================
		// Inline len() opcodes
		// ==========================================

		case OpLenList:
			frame.SP--
			obj := frame.Stack[frame.SP]
			if list, ok := obj.(*PyList); ok {
				frame.Stack[frame.SP] = MakeInt(int64(len(list.Items)))
				frame.SP++
			} else {
				return nil, fmt.Errorf("object of type '%s' has no len()", vm.typeName(obj))
			}

		case OpLenString:
			frame.SP--
			obj := frame.Stack[frame.SP]
			if str, ok := obj.(*PyString); ok {
				frame.Stack[frame.SP] = MakeInt(int64(len(str.Value)))
				frame.SP++
			} else {
				return nil, fmt.Errorf("object of type '%s' has no len()", vm.typeName(obj))
			}

		case OpLenTuple:
			frame.SP--
			obj := frame.Stack[frame.SP]
			if tup, ok := obj.(*PyTuple); ok {
				frame.Stack[frame.SP] = MakeInt(int64(len(tup.Items)))
				frame.SP++
			} else {
				return nil, fmt.Errorf("object of type '%s' has no len()", vm.typeName(obj))
			}

		case OpLenDict:
			frame.SP--
			obj := frame.Stack[frame.SP]
			if dict, ok := obj.(*PyDict); ok {
				frame.Stack[frame.SP] = MakeInt(int64(len(dict.Items)))
				frame.SP++
			} else {
				return nil, fmt.Errorf("object of type '%s' has no len()", vm.typeName(obj))
			}

		case OpLenGeneric:
			frame.SP--
			obj := frame.Stack[frame.SP]
			var length int64
			switch v := obj.(type) {
			case *PyString:
				length = int64(len(v.Value))
			case *PyList:
				length = int64(len(v.Items))
			case *PyTuple:
				length = int64(len(v.Items))
			case *PyDict:
				length = int64(len(v.Items))
			case *PySet:
				length = int64(len(v.Items))
			case *PyBytes:
				length = int64(len(v.Value))
			default:
				return nil, fmt.Errorf("object of type '%s' has no len()", vm.typeName(obj))
			}
			frame.Stack[frame.SP] = MakeInt(length)
			frame.SP++

		// ==========================================
		// More superinstructions
		// ==========================================

		case OpLoadConstLoadFast:
			// Load const then local: arg contains packed indices (high byte = const, low byte = local)
			constIdx := (arg >> 8) & 0xFF
			localIdx := arg & 0xFF
			vm.ensureStack(2) // Ensure space for two pushes
			frame.Stack[frame.SP] = vm.toValue(frame.Code.Constants[constIdx])
			frame.SP++
			frame.Stack[frame.SP] = frame.Locals[localIdx]
			frame.SP++

		case OpLoadGlobalLoadFast:
			// Load global then local: arg contains packed indices (high byte = name, low byte = local)
			nameIdx := (arg >> 8) & 0xFF
			localIdx := arg & 0xFF
			name := frame.Code.Names[nameIdx]
			if val, ok := frame.Globals[name]; ok {
				frame.Stack[frame.SP] = val
			} else if val, ok := frame.Builtins[name]; ok {
				frame.Stack[frame.SP] = val
			} else {
				return nil, fmt.Errorf("name '%s' is not defined", name)
			}
			frame.SP++
			frame.Stack[frame.SP] = frame.Locals[localIdx]
			frame.SP++

		case OpLoadGlobal:
			name := frame.Code.Names[arg]
			if val, ok := frame.Globals[name]; ok {
				vm.push(val)
			} else if val, ok := frame.Builtins[name]; ok {
				vm.push(val)
			} else {
				return nil, fmt.Errorf("name '%s' is not defined", name)
			}

		case OpStoreGlobal:
			name := frame.Code.Names[arg]
			frame.Globals[name] = vm.pop()

		case OpLoadAttr:
			name := frame.Code.Names[arg]
			obj := vm.pop()
			val, err := vm.getAttr(obj, name)
			if err != nil {
				return nil, err
			}
			vm.push(val)

		case OpStoreAttr:
			name := frame.Code.Names[arg]
			obj := vm.pop()
			val := vm.pop()
			err = vm.setAttr(obj, name, val)
			if err != nil {
				return nil, err
			}

		case OpBinarySubscr:
			index := vm.pop()
			obj := vm.pop()
			val, err := vm.getItem(obj, index)
			if err != nil {
				return nil, err
			}
			vm.push(val)

		case OpStoreSubscr:
			index := vm.pop()
			obj := vm.pop()
			val := vm.pop()
			err = vm.setItem(obj, index, val)
			if err != nil {
				return nil, err
			}

		case OpDeleteSubscr:
			index := vm.pop()
			obj := vm.pop()
			err = vm.delItem(obj, index)
			if err != nil {
				return nil, err
			}

		case OpUnaryPositive:
			a := vm.pop()
			vm.push(a) // Usually a no-op for numbers

		case OpUnaryNegative:
			a := vm.pop()
			result, err := vm.unaryOp(op, a)
			if err != nil {
				return nil, err
			}
			vm.push(result)

		case OpUnaryNot:
			a := vm.pop()
			if vm.truthy(a) {
				vm.push(False)
			} else {
				vm.push(True)
			}

		case OpUnaryInvert:
			a := vm.pop()
			result, err := vm.unaryOp(op, a)
			if err != nil {
				return nil, err
			}
			vm.push(result)

		case OpBinaryAdd:
			// Inline fast path for int + int (most common case)
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					frame.Stack[frame.SP] = MakeInt(ai.Value + bi.Value)
					frame.SP++
					break
				}
			}
			// Fall back to general case
			result, err := vm.binaryOp(op, a, b)
			if err != nil {
				return nil, err
			}
			frame.Stack[frame.SP] = result
			frame.SP++

		case OpBinarySubtract:
			// Inline fast path for int - int
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					frame.Stack[frame.SP] = MakeInt(ai.Value - bi.Value)
					frame.SP++
					break
				}
			}
			result, err := vm.binaryOp(op, a, b)
			if err != nil {
				return nil, err
			}
			frame.Stack[frame.SP] = result
			frame.SP++

		case OpBinaryMultiply:
			// Inline fast path for int * int
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					frame.Stack[frame.SP] = MakeInt(ai.Value * bi.Value)
					frame.SP++
					break
				}
			}
			result, err := vm.binaryOp(op, a, b)
			if err != nil {
				return nil, err
			}
			frame.Stack[frame.SP] = result
			frame.SP++

		case OpBinaryDivide, OpBinaryFloorDiv, OpBinaryModulo, OpBinaryPower, OpBinaryMatMul,
			OpBinaryLShift, OpBinaryRShift, OpBinaryAnd, OpBinaryOr, OpBinaryXor:
			b := vm.pop()
			a := vm.pop()
			result, err := vm.binaryOp(op, a, b)
			if err != nil {
				return nil, err
			}
			vm.push(result)

		case OpInplaceAdd, OpInplaceSubtract, OpInplaceMultiply, OpInplaceDivide,
			OpInplaceFloorDiv, OpInplaceModulo, OpInplacePower, OpInplaceMatMul,
			OpInplaceLShift, OpInplaceRShift, OpInplaceAnd, OpInplaceOr, OpInplaceXor:
			b := vm.pop()
			a := vm.pop()
			// For now, inplace ops are the same as binary ops
			binOp := op - OpInplaceAdd + OpBinaryAdd
			result, err := vm.binaryOp(binOp, a, b)
			if err != nil {
				return nil, err
			}
			vm.push(result)

		case OpCompareLt:
			// Inline fast path for int < int (very common in loops)
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					if ai.Value < bi.Value {
						frame.Stack[frame.SP] = True
					} else {
						frame.Stack[frame.SP] = False
					}
					frame.SP++
					break
				}
			}
			result := vm.compareOp(op, a, b)
			frame.Stack[frame.SP] = result
			frame.SP++

		case OpCompareEq, OpCompareNe, OpCompareLe,
			OpCompareGt, OpCompareGe, OpCompareIs, OpCompareIsNot,
			OpCompareIn, OpCompareNotIn:
			b := vm.pop()
			a := vm.pop()
			result := vm.compareOp(op, a, b)
			vm.push(result)

		case OpJump:
			frame.IP = arg

		case OpJumpIfTrue:
			if vm.truthy(vm.top()) {
				frame.IP = arg
			}

		case OpJumpIfFalse:
			if !vm.truthy(vm.top()) {
				frame.IP = arg
			}

		case OpPopJumpIfTrue:
			if vm.truthy(vm.pop()) {
				frame.IP = arg
			}

		case OpPopJumpIfFalse:
			// Inline pop and fast path for PyBool (result of comparisons)
			frame.SP--
			val := frame.Stack[frame.SP]
			if b, ok := val.(*PyBool); ok {
				if !b.Value {
					frame.IP = arg
				}
			} else if !vm.truthy(val) {
				frame.IP = arg
			}

		case OpJumpIfTrueOrPop:
			if vm.truthy(vm.top()) {
				frame.IP = arg
			} else {
				vm.pop()
			}

		case OpJumpIfFalseOrPop:
			if !vm.truthy(vm.top()) {
				frame.IP = arg
			} else {
				vm.pop()
			}

		case OpGetIter:
			obj := vm.pop()
			iter, err := vm.getIter(obj)
			if err != nil {
				return nil, err
			}
			vm.push(iter)

		case OpForIter:
			iter := vm.top()
			val, done, err := vm.iterNext(iter)
			if err != nil {
				return nil, err
			}
			if done {
				vm.pop() // Pop iterator
				frame.IP = arg
			} else {
				vm.push(val)
			}

		case OpBuildTuple:
			items := make([]Value, arg)
			for i := arg - 1; i >= 0; i-- {
				items[i] = vm.pop()
			}
			vm.push(&PyTuple{Items: items})

		case OpBuildList:
			items := make([]Value, arg)
			for i := arg - 1; i >= 0; i-- {
				items[i] = vm.pop()
			}
			vm.push(&PyList{Items: items})

		case OpBuildSet:
			s := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			for i := 0; i < arg; i++ {
				val := vm.pop()
				if !isHashable(val) {
					return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(val))
				}
				// Use hash-based storage for O(1) lookup
				s.SetAdd(val, vm)
			}
			vm.push(s)

		case OpBuildMap:
			d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
			for i := 0; i < arg; i++ {
				val := vm.pop()
				key := vm.pop()
				if !isHashable(key) {
					return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(key))
				}
				// Use hash-based storage for O(1) lookup
				d.DictSet(key, val, vm)
			}
			vm.push(d)

		case OpUnpackSequence:
			seq := vm.pop()
			items, err := vm.toList(seq)
			if err != nil {
				return nil, err
			}
			if len(items) != arg {
				return nil, fmt.Errorf("not enough values to unpack (expected %d, got %d)", arg, len(items))
			}
			for i := len(items) - 1; i >= 0; i-- {
				vm.push(items[i])
			}

		case OpListAppend:
			val := vm.pop()
			list := vm.peek(arg).(*PyList)
			list.Items = append(list.Items, val)

		case OpSetAdd:
			val := vm.pop()
			set := vm.peek(arg).(*PySet)
			if !isHashable(val) {
				return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(val))
			}
			// Use hash-based storage for O(1) lookup
			set.SetAdd(val, vm)

		case OpMapAdd:
			val := vm.pop()
			key := vm.pop()
			dict := vm.peek(arg).(*PyDict)
			if !isHashable(key) {
				return nil, fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(key))
			}
			// Use hash-based storage for O(1) lookup
			dict.DictSet(key, val, vm)

		case OpMakeFunction:
			name := vm.pop().(*PyString).Value
			code := vm.pop().(*CodeObject)
			var defaults *PyTuple
			var closure []*PyCell
			if arg&8 != 0 {
				// Has closure - pop tuple of cells
				closureTuple := vm.pop().(*PyTuple)
				closure = make([]*PyCell, len(closureTuple.Items))
				for i, item := range closureTuple.Items {
					closure[i] = item.(*PyCell)
				}
			}
			if arg&1 != 0 {
				defaults = vm.pop().(*PyTuple)
			}

			// If the code has FreeVars but we don't have a closure tuple,
			// capture the values from the current frame
			if closure == nil && len(code.FreeVars) > 0 {
				closure = make([]*PyCell, len(code.FreeVars))
				for i, varName := range code.FreeVars {
					// Look up the variable in the enclosing scope
					var val Value

					// Check if it's in the current frame's CellVars
					found := false
					for j, cellName := range frame.Code.CellVars {
						if cellName == varName && j < len(frame.Cells) && frame.Cells[j] != nil {
							cell := frame.Cells[j]
							// If cell value is nil, check if it was stored in locals instead
							// (this happens when the bytecode uses STORE_FAST before we knew
							// the variable would be captured)
							if cell.Value == nil {
								for k, localName := range frame.Code.VarNames {
									if localName == varName && k < len(frame.Locals) && frame.Locals[k] != nil {
										cell.Value = frame.Locals[k]
										break
									}
								}
							}
							// Share the same cell
							closure[i] = cell
							found = true
							break
						}
					}
					if found {
						continue
					}

					// Check if it's in the current frame's FreeVars (cells after CellVars)
					numCellVars := len(frame.Code.CellVars)
					for j, freeName := range frame.Code.FreeVars {
						cellIdx := numCellVars + j
						if freeName == varName && cellIdx < len(frame.Cells) && frame.Cells[cellIdx] != nil {
							// Share the same cell (pass through)
							closure[i] = frame.Cells[cellIdx]
							found = true
							break
						}
					}
					if found {
						continue
					}

					// Check locals
					for j, localName := range frame.Code.VarNames {
						if localName == varName && j < len(frame.Locals) {
							val = frame.Locals[j]
							break
						}
					}

					// Check globals if not found in locals
					if val == nil {
						val = frame.Globals[varName]
					}
					if val == nil && frame.EnclosingGlobals != nil {
						val = frame.EnclosingGlobals[varName]
					}

					closure[i] = &PyCell{Value: val}
				}
			}

			// Use enclosing globals if available (for methods in class bodies)
			// so they can access module-level variables
			fnGlobals := frame.Globals
			if frame.EnclosingGlobals != nil {
				fnGlobals = frame.EnclosingGlobals
			}
			fn := &PyFunction{
				Code:     code,
				Globals:  fnGlobals,
				Defaults: defaults,
				Closure:  closure,
				Name:     name,
			}
			vm.push(fn)

		case OpCall:
			args := make([]Value, arg)
			for i := arg - 1; i >= 0; i-- {
				args[i] = vm.pop()
			}
			callable := vm.pop()
			result, err := vm.call(callable, args, nil)
			if err != nil {
				// Check if exception was already handled in an outer frame
				if err == errExceptionHandledInOuterFrame {
					// Check if the handler is in THIS frame
					if vm.frame == frame {
						// Handler is in this frame, continue from handler
						continue
					}
					// Handler is in an even outer frame, propagate
					return nil, err
				}
				// Check if it's a Python exception that can be handled
				if pyExc, ok := err.(*PyException); ok {
					_, handleErr := vm.handleException(pyExc)
					if handleErr != nil {
						// No handler found, propagate exception
						return nil, handleErr
					}
					// Handler found - check if it's in this frame
					if vm.frame != frame {
						return nil, errExceptionHandledInOuterFrame
					}
					// Handler is in current frame, continue
					continue
				}
				return nil, err
			}
			vm.push(result)

		case OpCallKw:
			kwNames := vm.pop().(*PyTuple)
			totalArgs := arg
			kwargs := make(map[string]Value)
			for i := len(kwNames.Items) - 1; i >= 0; i-- {
				name := kwNames.Items[i].(*PyString).Value
				kwargs[name] = vm.pop()
				totalArgs--
			}
			args := make([]Value, totalArgs)
			for i := totalArgs - 1; i >= 0; i-- {
				args[i] = vm.pop()
			}
			callable := vm.pop()
			result, err := vm.call(callable, args, kwargs)
			if err != nil {
				// Check if exception was already handled in an outer frame
				if err == errExceptionHandledInOuterFrame {
					if vm.frame == frame {
						continue
					}
					return nil, err
				}
				// Check if it's a Python exception that can be handled
				if pyExc, ok := err.(*PyException); ok {
					_, handleErr := vm.handleException(pyExc)
					if handleErr != nil {
						return nil, handleErr
					}
					if vm.frame != frame {
						return nil, errExceptionHandledInOuterFrame
					}
					continue
				}
				return nil, err
			}
			vm.push(result)

		case OpReturn:
			// Inline pop for return
			frame.SP--
			result := frame.Stack[frame.SP]
			vm.frames = vm.frames[:len(vm.frames)-1]
			if len(vm.frames) > 0 {
				vm.frame = vm.frames[len(vm.frames)-1]
			}
			return result, nil

		case OpYieldValue:
			// Yield is only valid in generators - should not reach here in regular run()
			// If we get here, it means a generator's frame is being run directly
			value := vm.pop()
			return value, nil

		case OpYieldFrom:
			// Yield from delegates to a sub-iterator
			// This should not be reached in regular run() - handled by runGenerator()
			return nil, fmt.Errorf("yield from outside generator")

		case OpGenStart:
			// No-op marker for generator start
			// This opcode is used to mark the beginning of a generator function

		case OpGetAwaitable:
			// Get an awaitable object from the top of the stack
			obj := vm.pop()
			// Check if it's already a coroutine or has __await__
			switch v := obj.(type) {
			case *PyCoroutine:
				vm.push(v)
			case *PyGenerator:
				// Generators are iterable but not awaitable by default
				// Only coroutine generators (from async def) are awaitable
				vm.push(v)
			default:
				// Try to call __await__ method
				awaitMethod, err := vm.getAttr(obj, "__await__")
				if err != nil {
					return nil, fmt.Errorf("object %T is not awaitable", obj)
				}
				awaitable, err := vm.call(awaitMethod, nil, nil)
				if err != nil {
					return nil, err
				}
				vm.push(awaitable)
			}

		case OpGetAIter:
			// Get an async iterator from the top of the stack
			obj := vm.pop()
			aiterMethod, err := vm.getAttr(obj, "__aiter__")
			if err != nil {
				return nil, fmt.Errorf("object %T is not async iterable", obj)
			}
			aiter, err := vm.call(aiterMethod, nil, nil)
			if err != nil {
				return nil, err
			}
			vm.push(aiter)

		case OpGetANext:
			// Get next from an async iterator
			aiter := vm.top()
			anextMethod, err := vm.getAttr(aiter, "__anext__")
			if err != nil {
				return nil, fmt.Errorf("async iterator has no __anext__ method")
			}
			anext, err := vm.call(anextMethod, nil, nil)
			if err != nil {
				return nil, err
			}
			vm.push(anext)

		// Pattern matching opcodes
		case OpMatchSequence:
			// Check if TOS is a sequence with the expected length
			// arg = expected length (-1 means any length is ok)
			// Does NOT pop subject - leaves it for element access
			subject := vm.top()
			expectedLen := arg

			// Check if it's a matchable sequence (list or tuple, not string/bytes)
			var length int
			isSequence := false
			switch s := subject.(type) {
			case *PyList:
				length = len(s.Items)
				isSequence = true
			case *PyTuple:
				length = len(s.Items)
				isSequence = true
			}

			// expectedLen of 65535 (0xFFFF) means any length is acceptable (star pattern)
			anyLength := expectedLen == 65535 || expectedLen == -1

			if !isSequence {
				vm.push(False)
			} else if !anyLength && length != expectedLen {
				vm.push(False)
			} else {
				vm.push(True)
			}

		case OpMatchStar:
			// Check minimum length for star pattern
			// arg = minLength (beforeStar + afterStar)
			// Does NOT pop subject
			subject := vm.top()
			minLen := arg

			var length int
			switch s := subject.(type) {
			case *PyList:
				length = len(s.Items)
			case *PyTuple:
				length = len(s.Items)
			default:
				vm.push(False)
				continue
			}

			if length < minLen {
				vm.push(False)
			} else {
				vm.push(True)
			}

		case OpExtractStar:
			// Extract star slice from sequence
			// arg = beforeStar << 8 | afterStar
			beforeStar := arg >> 8
			afterStar := arg & 0xFF
			subject := vm.top()

			var items []Value
			switch s := subject.(type) {
			case *PyList:
				items = s.Items
			case *PyTuple:
				items = s.Items
			default:
				vm.push(&PyList{Items: nil})
				continue
			}

			// Extract slice: items[beforeStar : len(items) - afterStar]
			start := beforeStar
			end := len(items) - afterStar
			if end < start {
				end = start
			}
			slice := make([]Value, end-start)
			copy(slice, items[start:end])
			vm.push(&PyList{Items: slice})

		case OpMatchMapping:
			// Check if TOS is a mapping (dict)
			subject := vm.top()
			if _, ok := subject.(*PyDict); ok {
				vm.push(True)
			} else {
				vm.pop()
				vm.push(False)
			}

		case OpMatchKeys:
			// Check if mapping has all required keys
			// Stack: subject, key1, key2, ..., keyN, N
			// (True from MATCH_MAPPING was already consumed by POP_JUMP_IF_FALSE)
			keyCount := int(vm.pop().(*PyInt).Value)
			keys := make([]Value, keyCount)
			for i := keyCount - 1; i >= 0; i-- {
				keys[i] = vm.pop()
			}

			subject := vm.top()
			dict, ok := subject.(*PyDict)
			if !ok {
				vm.pop()
				vm.push(False)
				continue
			}

			// Check all keys exist and collect values
			values := make([]Value, keyCount)
			allPresent := true
			for i, key := range keys {
				found := false
				for k, v := range dict.Items {
					if vm.equal(k, key) {
						values[i] = v
						found = true
						break
					}
				}
				if !found {
					allPresent = false
					break
				}
			}

			if !allPresent {
				vm.pop() // remove subject
				vm.push(False)
			} else {
				// Push values in reverse order (so values[0] is on top after True is popped)
				// Stack after: [subject, valueN, ..., value2, value1, True]
				// After PopJumpIfFalse: [subject, valueN, ..., value2, value1]
				// Now value1 is on TOS for the first pattern
				for i := len(values) - 1; i >= 0; i-- {
					vm.push(values[i])
				}
				vm.push(True)
			}

		case OpCopyDict:
			// Copy dict, optionally removing specified keys
			// Stack: subject, key1, ..., keyN, N
			// (True was already consumed by POP_JUMP_IF_FALSE)
			keyCount := int(vm.pop().(*PyInt).Value)
			keysToRemove := make([]Value, keyCount)
			for i := keyCount - 1; i >= 0; i-- {
				keysToRemove[i] = vm.pop()
			}

			subject := vm.top()
			dict := subject.(*PyDict)

			// Create a copy without the specified keys
			newDict := &PyDict{Items: make(map[Value]Value)}
			for k, v := range dict.Items {
				shouldRemove := false
				for _, removeKey := range keysToRemove {
					if vm.equal(k, removeKey) {
						shouldRemove = true
						break
					}
				}
				if !shouldRemove {
					newDict.Items[k] = v
				}
			}
			vm.push(newDict)

		case OpMatchClass:
			// Match class pattern
			// Stack: subject, class
			// arg = number of positional patterns
			cls := vm.pop()
			subject := vm.top()
			positionalCount := arg

			// Check isinstance
			isInstance := false
			var matchedClass *PyClass
			switch c := cls.(type) {
			case *PyClass:
				matchedClass = c
				if inst, ok := subject.(*PyInstance); ok {
					// Check if instance's class matches or is subclass
					isInstance = vm.isInstanceOf(inst, c)
				}
			}

			if !isInstance {
				vm.pop() // remove subject
				vm.push(False)
				continue
			}

			// Get __match_args__ for positional pattern mapping
			var matchArgs []string
			if matchedClass != nil && positionalCount > 0 {
				if maVal, ok := matchedClass.Dict["__match_args__"]; ok {
					if maTuple, ok := maVal.(*PyTuple); ok {
						for _, item := range maTuple.Items {
							if s, ok := item.(*PyString); ok {
								matchArgs = append(matchArgs, s.Value)
							}
						}
					} else if maList, ok := maVal.(*PyList); ok {
						for _, item := range maList.Items {
							if s, ok := item.(*PyString); ok {
								matchArgs = append(matchArgs, s.Value)
							}
						}
					}
				}
			}

			// Extract attributes based on positional patterns
			if positionalCount > len(matchArgs) && matchedClass != nil {
				vm.pop()
				vm.push(False)
				continue
			}

			// Get attribute values
			attrs := make([]Value, positionalCount)
			allFound := true
			inst, isInst := subject.(*PyInstance)
			for i := 0; i < positionalCount; i++ {
				if i < len(matchArgs) {
					attrName := matchArgs[i]
					if isInst {
						if val, ok := inst.Dict[attrName]; ok {
							attrs[i] = val
						} else {
							allFound = false
							break
						}
					} else {
						allFound = false
						break
					}
				} else {
					allFound = false
					break
				}
			}

			if !allFound {
				vm.pop()
				vm.push(False)
			} else {
				// Push attrs in reverse order first
				for i := len(attrs) - 1; i >= 0; i-- {
					vm.push(attrs[i])
				}
				// Push True last so it's on top for POP_JUMP_IF_FALSE
				vm.push(True)
			}

		case OpGetLen:
			// Get length of TOS without consuming it (for pattern matching length checks)
			subject := vm.top()
			var length int64
			switch s := subject.(type) {
			case *PyList:
				length = int64(len(s.Items))
			case *PyTuple:
				length = int64(len(s.Items))
			case *PyString:
				length = int64(len(s.Value))
			case *PyDict:
				length = int64(len(s.Items))
			default:
				return nil, fmt.Errorf("object of type '%s' has no len()", vm.typeName(subject))
			}
			vm.push(MakeInt(length))

		case OpLoadBuildClass:
			vm.push(vm.builtins["__build_class__"])

		case OpLoadMethod:
			name := frame.Code.Names[arg]
			obj := vm.pop()
			method, err := vm.getAttr(obj, name)
			if err != nil {
				return nil, err
			}
			// Push object and method for CALL_METHOD
			vm.push(obj)
			vm.push(method)

		case OpCallMethod:
			args := make([]Value, arg)
			for i := arg - 1; i >= 0; i-- {
				args[i] = vm.pop()
			}
			method := vm.pop()
			obj := vm.pop()
			var result Value
			var err error
			// Check if method is already bound (PyMethod)
			if _, isBound := method.(*PyMethod); isBound {
				// Bound method already has self, don't prepend again
				result, err = vm.call(method, args, nil)
			} else {
				// Raw function, need to prepend self
				allArgs := append([]Value{obj}, args...)
				result, err = vm.call(method, allArgs, nil)
			}
			if err != nil {
				return nil, err
			}
			vm.push(result)

		case OpLoadLocals:
			locals := &PyDict{Items: make(map[Value]Value)}
			for i, name := range frame.Code.VarNames {
				if frame.Locals[i] != nil {
					locals.Items[&PyString{Value: name}] = frame.Locals[i]
				}
			}
			vm.push(locals)

		case OpSetupExcept:
			// Push exception handler block onto block stack
			block := Block{
				Type:    BlockExcept,
				Handler: arg,
				Level:   frame.SP,
			}
			frame.BlockStack = append(frame.BlockStack, block)

		case OpSetupFinally:
			// Push finally handler block onto block stack
			block := Block{
				Type:    BlockFinally,
				Handler: arg,
				Level:   frame.SP,
			}
			frame.BlockStack = append(frame.BlockStack, block)

		case OpPopExcept:
			// Pop exception handler from block stack (try block completed normally)
			if len(frame.BlockStack) > 0 {
				frame.BlockStack = frame.BlockStack[:len(frame.BlockStack)-1]
			}
			vm.currentException = nil

		case OpClearException:
			// Clear the current exception state (when handler catches exception)
			// Does NOT pop the block stack (block was already popped by handleException)
			vm.currentException = nil

		case OpEndFinally:
			// End finally block - re-raise exception if one was active
			if vm.currentException != nil {
				exc := vm.currentException
				vm.currentException = nil
				// Try to find an exception handler
				result, err := vm.handleException(exc)
				if err != nil {
					// No handler found, propagate exception
					return nil, err
				}
				// Handler found, continue execution
				_ = result
			}

		case OpExceptionMatch:
			// Check if exception matches type for except clause
			// Stack: [..., exception, type] -> [..., exception, bool]
			excType := vm.pop()
			exc := vm.top() // Peek, don't pop
			if pyExc, ok := exc.(*PyException); ok {
				if vm.exceptionMatches(pyExc, excType) {
					vm.push(True)
				} else {
					vm.push(False)
				}
			} else {
				vm.push(False)
			}

		case OpRaiseVarargs:
			var exc *PyException
			if arg == 0 {
				// Bare raise - re-raise current/last exception
				if vm.lastException != nil {
					exc = vm.lastException
				} else {
					return nil, fmt.Errorf("RuntimeError: No active exception to re-raise")
				}
			} else if arg == 1 {
				// raise exc
				excVal := vm.pop()
				exc = vm.createException(excVal, nil)
			} else if arg >= 2 {
				// raise exc from cause
				cause := vm.pop()
				excVal := vm.pop()
				exc = vm.createException(excVal, cause)
			}

			// Build traceback
			exc.Traceback = vm.buildTraceback()

			// Try to find an exception handler
			_, err = vm.handleException(exc)
			if err != nil {
				// No handler found, propagate exception
				return nil, err
			}
			// Handler found - check if we're still in the same frame
			if vm.frame != frame {
				// Handler is in an outer frame, return sentinel to signal caller
				return nil, errExceptionHandledInOuterFrame
			}
			// Handler is in current frame, update and continue
			frame = vm.frame

		case OpImportName:
			name := frame.Code.Names[arg]
			fromlist := vm.pop() // fromlist (list of names to import, or nil)
			_ = vm.pop()         // level (for relative imports, not yet used)

			// Try to import the module
			mod, err := vm.ImportModule(name)
			if err != nil {
				return nil, err
			}
			vm.push(mod)

			// If fromlist is provided and non-empty, we're doing "from X import Y"
			// The actual attribute extraction is done by IMPORT_FROM
			_ = fromlist

		case OpImportFrom:
			name := frame.Code.Names[arg]
			// Top of stack is the module
			mod := vm.top()
			pyMod, ok := mod.(*PyModule)
			if !ok {
				return nil, fmt.Errorf("import from requires a module, got %s", vm.typeName(mod))
			}

			// Get the attribute from the module
			value, exists := pyMod.Get(name)
			if !exists {
				return nil, fmt.Errorf("cannot import name '%s' from '%s'", name, pyMod.Name)
			}
			vm.push(value)

		case OpImportStar:
			// Top of stack is the module (don't pop - the compiler emits a POP after)
			mod := vm.top()
			pyMod, ok := mod.(*PyModule)
			if !ok {
				return nil, fmt.Errorf("import * requires a module, got %s", vm.typeName(mod))
			}

			// Import all public names (not starting with _) into globals
			for name, value := range pyMod.Dict {
				if len(name) == 0 || name[0] != '_' {
					frame.Globals[name] = value
				}
			}

		case OpNop:
			// Do nothing

		default:
			return nil, fmt.Errorf("unimplemented opcode: %s", op.String())
		}
	}

	// Implicit return None
	return None, nil
}

// Stack operations - using stack pointer with pre-allocated slice

func (vm *VM) push(v Value) {
	f := vm.frame
	// Grow stack if needed
	if f.SP >= len(f.Stack) {
		newStack := make([]Value, len(f.Stack)*2)
		copy(newStack, f.Stack)
		f.Stack = newStack
	}
	f.Stack[f.SP] = v
	f.SP++
}

// ensureStack ensures the stack has at least n additional slots available
func (vm *VM) ensureStack(n int) {
	f := vm.frame
	needed := f.SP + n
	if needed > len(f.Stack) {
		newSize := len(f.Stack) * 2
		if newSize < needed {
			newSize = needed + 16
		}
		newStack := make([]Value, newSize)
		copy(newStack, f.Stack)
		f.Stack = newStack
	}
}

func (vm *VM) pop() Value {
	if vm.frame.SP <= 0 {
		panic("stack underflow: cannot pop from empty stack")
	}
	vm.frame.SP--
	return vm.frame.Stack[vm.frame.SP]
}

func (vm *VM) top() Value {
	if vm.frame.SP <= 0 {
		panic("stack underflow: cannot access top of empty stack")
	}
	return vm.frame.Stack[vm.frame.SP-1]
}

func (vm *VM) peek(n int) Value {
	idx := vm.frame.SP - 1 - n
	if idx < 0 || idx >= vm.frame.SP {
		panic("stack underflow: invalid peek index")
	}
	return vm.frame.Stack[idx]
}

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

// Type conversions

func (vm *VM) toValue(v any) Value {
	if v == nil {
		return None
	}
	switch val := v.(type) {
	case bool:
		if val {
			return True
		}
		return False
	case int:
		return MakeInt(int64(val))
	case int64:
		return MakeInt(val)
	case float64:
		return &PyFloat{Value: val}
	case string:
		return &PyString{Value: val}
	case []byte:
		return &PyBytes{Value: val}
	case []string:
		items := make([]Value, len(val))
		for i, s := range val {
			items[i] = &PyString{Value: s}
		}
		return &PyTuple{Items: items}
	case *CodeObject:
		return val
	case Value:
		return val
	default:
		return &PyString{Value: fmt.Sprintf("%v", v)}
	}
}

func (vm *VM) toInt(v Value) int64 {
	i, _ := vm.tryToInt(v)
	return i
}

// tryToInt converts a value to int64, returning an error if conversion fails.
// Use this for Python's int() builtin where ValueError should be raised on failure.
func (vm *VM) tryToInt(v Value) (int64, error) {
	switch val := v.(type) {
	case *PyInt:
		return val.Value, nil
	case *PyFloat:
		return int64(val.Value), nil
	case *PyBool:
		if val.Value {
			return 1, nil
		}
		return 0, nil
	case *PyString:
		s := strings.TrimSpace(val.Value)
		if s == "" {
			return 0, fmt.Errorf("ValueError: invalid literal for int() with base 10: %q", val.Value)
		}
		// Try parsing as integer
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			// Try parsing as float then converting
			f, ferr := strconv.ParseFloat(s, 64)
			if ferr != nil {
				return 0, fmt.Errorf("ValueError: invalid literal for int() with base 10: %q", val.Value)
			}
			return int64(f), nil
		}
		return i, nil
	default:
		return 0, fmt.Errorf("TypeError: int() argument must be a string or a number, not '%s'", vm.typeName(v))
	}
}

func (vm *VM) toFloat(v Value) float64 {
	f, _ := vm.tryToFloat(v)
	return f
}

// tryToFloat converts a value to float64, returning an error if conversion fails.
// Use this for Python's float() builtin where ValueError should be raised on failure.
func (vm *VM) tryToFloat(v Value) (float64, error) {
	switch val := v.(type) {
	case *PyInt:
		return float64(val.Value), nil
	case *PyFloat:
		return val.Value, nil
	case *PyBool:
		if val.Value {
			return 1.0, nil
		}
		return 0.0, nil
	case *PyString:
		s := strings.TrimSpace(val.Value)
		if s == "" {
			return 0, fmt.Errorf("ValueError: could not convert string to float: %q", val.Value)
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, fmt.Errorf("ValueError: could not convert string to float: %q", val.Value)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("TypeError: float() argument must be a string or a number, not '%s'", vm.typeName(v))
	}
}

func (vm *VM) toList(v Value) ([]Value, error) {
	switch val := v.(type) {
	case *PyList:
		return val.Items, nil
	case *PyTuple:
		return val.Items, nil
	case *PyString:
		items := make([]Value, len(val.Value))
		for i, ch := range val.Value {
			items[i] = &PyString{Value: string(ch)}
		}
		return items, nil
	case *PyRange:
		var items []Value
		for i := val.Start; (val.Step > 0 && i < val.Stop) || (val.Step < 0 && i > val.Stop); i += val.Step {
			items = append(items, MakeInt(i))
		}
		return items, nil
	case *PySet:
		var items []Value
		for k := range val.Items {
			items = append(items, k)
		}
		return items, nil
	case *PyDict:
		var items []Value
		for k := range val.Items {
			items = append(items, k)
		}
		return items, nil
	case *PyIterator:
		return val.Items[val.Index:], nil
	default:
		return nil, fmt.Errorf("'%s' object is not iterable", vm.typeName(v))
	}
}

func (vm *VM) truthy(v Value) bool {
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
		return len(val.Items) > 0
	case *PySet:
		return len(val.Items) > 0
	default:
		return true
	}
}

func (vm *VM) str(v Value) string {
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
		return val.Value
	case *PyBytes:
		return fmt.Sprintf("b'%s'", string(val.Value))
	case *PyList:
		return fmt.Sprintf("%v", val.Items)
	case *PyTuple:
		return fmt.Sprintf("%v", val.Items)
	case *PyDict:
		return fmt.Sprintf("%v", val.Items)
	case *PySet:
		return fmt.Sprintf("%v", val.Items)
	case *PyFunction:
		return fmt.Sprintf("<function %s>", val.Name)
	case *PyBuiltinFunc:
		return fmt.Sprintf("<built-in function %s>", val.Name)
	case *PyGoFunc:
		return fmt.Sprintf("<go function %s>", val.Name)
	case *PyUserData:
		return fmt.Sprintf("<userdata %T>", val.Value)
	case *PyModule:
		return fmt.Sprintf("<module '%s'>", val.Name)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (vm *VM) typeName(v Value) string {
	switch val := v.(type) {
	case *PyNone:
		return "NoneType"
	case *PyBool:
		return "bool"
	case *PyInt:
		return "int"
	case *PyFloat:
		return "float"
	case *PyString:
		return "str"
	case *PyBytes:
		return "bytes"
	case *PyList:
		return "list"
	case *PyTuple:
		return "tuple"
	case *PyDict:
		return "dict"
	case *PySet:
		return "set"
	case *PyFunction:
		return "function"
	case *PyBuiltinFunc:
		return "builtin_function_or_method"
	case *PyGoFunc:
		return "builtin_function_or_method"
	case *PyClass:
		return "type"
	case *PyInstance:
		return val.Class.Name
	case *PyRange:
		return "range"
	case *PyIterator:
		return "iterator"
	case *PyUserData:
		if val.Metatable != nil {
			// Find __type__ key in metatable (iterate because Value keys use pointers)
			for k, v := range val.Metatable.Items {
				if ks, ok := k.(*PyString); ok && ks.Value == "__type__" {
					return vm.str(v)
				}
			}
		}
		return "userdata"
	case *PyModule:
		return "module"
	default:
		return "object"
	}
}

// isInstanceOf checks if an instance is an instance of a class (including subclasses)
func (vm *VM) isInstanceOf(inst *PyInstance, cls *PyClass) bool {
	// Direct class match
	if inst.Class == cls {
		return true
	}
	// Check base classes recursively
	return vm.isSubclassOf(inst.Class, cls)
}

// isSubclassOf checks if cls is a subclass of target
func (vm *VM) isSubclassOf(cls, target *PyClass) bool {
	if cls == target {
		return true
	}
	for _, base := range cls.Bases {
		if vm.isSubclassOf(base, target) {
			return true
		}
	}
	return false
}

// Operations

func (vm *VM) unaryOp(op Opcode, a Value) (Value, error) {
	switch op {
	case OpUnaryNegative:
		switch v := a.(type) {
		case *PyInt:
			return MakeInt(-v.Value), nil
		case *PyFloat:
			return &PyFloat{Value: -v.Value}, nil
		}
	case OpUnaryInvert:
		if v, ok := a.(*PyInt); ok {
			return MakeInt(^v.Value), nil
		}
	}
	return nil, fmt.Errorf("bad operand type for unary %s: '%s'", op.String(), vm.typeName(a))
}

func (vm *VM) binaryOp(op Opcode, a, b Value) (Value, error) {
	// Fast path: int op int (most common case in numeric code)
	if ai, ok := a.(*PyInt); ok {
		if bi, ok := b.(*PyInt); ok {
			switch op {
			case OpBinaryAdd:
				return MakeInt(ai.Value + bi.Value), nil
			case OpBinarySubtract:
				return MakeInt(ai.Value - bi.Value), nil
			case OpBinaryMultiply:
				return MakeInt(ai.Value * bi.Value), nil
			case OpBinaryDivide:
				if bi.Value == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return &PyFloat{Value: float64(ai.Value) / float64(bi.Value)}, nil
			case OpBinaryFloorDiv:
				if bi.Value == 0 {
					return nil, fmt.Errorf("integer division by zero")
				}
				// Python floor division: always rounds toward negative infinity
				result := ai.Value / bi.Value
				// If signs differ and there's a remainder, adjust toward -inf
				if (ai.Value < 0) != (bi.Value < 0) && ai.Value%bi.Value != 0 {
					result--
				}
				return MakeInt(result), nil
			case OpBinaryModulo:
				if bi.Value == 0 {
					return nil, fmt.Errorf("integer modulo by zero")
				}
				// Python modulo: result has same sign as divisor
				result := ai.Value % bi.Value
				if result != 0 && (result < 0) != (bi.Value < 0) {
					result += bi.Value
				}
				return MakeInt(result), nil
			case OpBinaryPower:
				// Use integer exponentiation to avoid float precision loss
				if bi.Value < 0 {
					// Negative exponent returns float
					return &PyFloat{Value: math.Pow(float64(ai.Value), float64(bi.Value))}, nil
				}
				return MakeInt(intPow(ai.Value, bi.Value)), nil
			case OpBinaryLShift:
				if bi.Value < 0 {
					return nil, fmt.Errorf("ValueError: negative shift count")
				}
				if bi.Value > 63 {
					// For very large shifts, result is 0 (or overflow behavior)
					if ai.Value == 0 {
						return MakeInt(0), nil
					}
					return MakeInt(0), nil // Simplified: large left shifts overflow to 0
				}
				return MakeInt(ai.Value << uint(bi.Value)), nil
			case OpBinaryRShift:
				if bi.Value < 0 {
					return nil, fmt.Errorf("ValueError: negative shift count")
				}
				if bi.Value > 63 {
					// Right shift by large amount gives 0 or -1 (for negative numbers)
					if ai.Value < 0 {
						return MakeInt(-1), nil
					}
					return MakeInt(0), nil
				}
				return MakeInt(ai.Value >> uint(bi.Value)), nil
			case OpBinaryAnd:
				return MakeInt(ai.Value & bi.Value), nil
			case OpBinaryOr:
				return MakeInt(ai.Value | bi.Value), nil
			case OpBinaryXor:
				return MakeInt(ai.Value ^ bi.Value), nil
			}
		}
	}

	// Handle string concatenation
	if op == OpBinaryAdd {
		if as, ok := a.(*PyString); ok {
			if bs, ok := b.(*PyString); ok {
				return &PyString{Value: as.Value + bs.Value}, nil
			}
		}
		// List concatenation
		if al, ok := a.(*PyList); ok {
			if bl, ok := b.(*PyList); ok {
				items := make([]Value, len(al.Items)+len(bl.Items))
				copy(items, al.Items)
				copy(items[len(al.Items):], bl.Items)
				return &PyList{Items: items}, nil
			}
		}
	}

	// String repetition - use strings.Repeat for O(n) instead of O(n²)
	// Limit maximum result size to 100MB to prevent memory exhaustion
	const maxStringRepeatSize = 100 * 1024 * 1024
	if op == OpBinaryMultiply {
		if as, ok := a.(*PyString); ok {
			if bi, ok := b.(*PyInt); ok {
				if bi.Value <= 0 {
					return &PyString{Value: ""}, nil
				}
				resultSize := int64(len(as.Value)) * bi.Value
				if resultSize > maxStringRepeatSize {
					return nil, fmt.Errorf("MemoryError: string repetition result too large")
				}
				return &PyString{Value: strings.Repeat(as.Value, int(bi.Value))}, nil
			}
		}
		if as, ok := b.(*PyString); ok {
			if ai, ok := a.(*PyInt); ok {
				if ai.Value <= 0 {
					return &PyString{Value: ""}, nil
				}
				resultSize := int64(len(as.Value)) * ai.Value
				if resultSize > maxStringRepeatSize {
					return nil, fmt.Errorf("MemoryError: string repetition result too large")
				}
				return &PyString{Value: strings.Repeat(as.Value, int(ai.Value))}, nil
			}
		}
		// List repetition - pre-allocate for efficiency
		// Limit maximum result size to 10M items to prevent memory exhaustion
		const maxListRepeatItems = 10 * 1024 * 1024
		if al, ok := a.(*PyList); ok {
			if bi, ok := b.(*PyInt); ok {
				if bi.Value <= 0 {
					return &PyList{Items: []Value{}}, nil
				}
				resultItems := int64(len(al.Items)) * bi.Value
				if resultItems > maxListRepeatItems {
					return nil, fmt.Errorf("MemoryError: list repetition result too large")
				}
				count := int(bi.Value)
				items := make([]Value, 0, len(al.Items)*count)
				for i := 0; i < count; i++ {
					items = append(items, al.Items...)
				}
				return &PyList{Items: items}, nil
			}
		}
	}

	// Float operations (including int+float and float+int)
	af, aIsFloat := a.(*PyFloat)
	bf, bIsFloat := b.(*PyFloat)
	ai, aIsInt := a.(*PyInt)
	bi, bIsInt := b.(*PyInt)

	// Convert to float if mixed types
	if aIsInt && bIsFloat {
		af = &PyFloat{Value: float64(ai.Value)}
		aIsFloat = true
	}
	if aIsFloat && bIsInt {
		bf = &PyFloat{Value: float64(bi.Value)}
		bIsFloat = true
	}

	if aIsFloat && bIsFloat {
		switch op {
		case OpBinaryAdd:
			return &PyFloat{Value: af.Value + bf.Value}, nil
		case OpBinarySubtract:
			return &PyFloat{Value: af.Value - bf.Value}, nil
		case OpBinaryMultiply:
			return &PyFloat{Value: af.Value * bf.Value}, nil
		case OpBinaryDivide:
			if bf.Value == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return &PyFloat{Value: af.Value / bf.Value}, nil
		case OpBinaryFloorDiv:
			if bf.Value == 0 {
				return nil, fmt.Errorf("float floor division by zero")
			}
			return &PyFloat{Value: math.Floor(af.Value / bf.Value)}, nil
		case OpBinaryModulo:
			if bf.Value == 0 {
				return nil, fmt.Errorf("float modulo by zero")
			}
			return &PyFloat{Value: math.Mod(af.Value, bf.Value)}, nil
		case OpBinaryPower:
			return &PyFloat{Value: math.Pow(af.Value, bf.Value)}, nil
		}
	}

	return nil, fmt.Errorf("unsupported operand type(s) for %s: '%s' and '%s'",
		op.String(), vm.typeName(a), vm.typeName(b))
}

func (vm *VM) compareOp(op Opcode, a, b Value) Value {
	// Fast path: int vs int comparisons (most common case)
	if ai, ok := a.(*PyInt); ok {
		if bi, ok := b.(*PyInt); ok {
			switch op {
			case OpCompareEq:
				if ai.Value == bi.Value {
					return True
				}
				return False
			case OpCompareNe:
				if ai.Value != bi.Value {
					return True
				}
				return False
			case OpCompareLt:
				if ai.Value < bi.Value {
					return True
				}
				return False
			case OpCompareLe:
				if ai.Value <= bi.Value {
					return True
				}
				return False
			case OpCompareGt:
				if ai.Value > bi.Value {
					return True
				}
				return False
			case OpCompareGe:
				if ai.Value >= bi.Value {
					return True
				}
				return False
			}
		}
	}

	switch op {
	case OpCompareEq:
		if vm.equal(a, b) {
			return True
		}
		return False
	case OpCompareNe:
		if !vm.equal(a, b) {
			return True
		}
		return False
	case OpCompareLt:
		if vm.compare(a, b) < 0 {
			return True
		}
		return False
	case OpCompareLe:
		if vm.compare(a, b) <= 0 {
			return True
		}
		return False
	case OpCompareGt:
		if vm.compare(a, b) > 0 {
			return True
		}
		return False
	case OpCompareGe:
		if vm.compare(a, b) >= 0 {
			return True
		}
		return False
	case OpCompareIs:
		if a == b {
			return True
		}
		return False
	case OpCompareIsNot:
		if a != b {
			return True
		}
		return False
	case OpCompareIn:
		if vm.contains(b, a) {
			return True
		}
		return False
	case OpCompareNotIn:
		if !vm.contains(b, a) {
			return True
		}
		return False
	}
	return False
}

func (vm *VM) equal(a, b Value) bool {
	return vm.equalWithCycleDetection(a, b, make(map[uintptr]map[uintptr]bool))
}

// equalWithCycleDetection compares values with cycle detection to prevent stack overflow
func (vm *VM) equalWithCycleDetection(a, b Value, seen map[uintptr]map[uintptr]bool) bool {
	switch av := a.(type) {
	case *PyNone:
		_, ok := b.(*PyNone)
		return ok
	case *PyBool:
		if bv, ok := b.(*PyBool); ok {
			return av.Value == bv.Value
		}
	case *PyInt:
		switch bv := b.(type) {
		case *PyInt:
			return av.Value == bv.Value
		case *PyFloat:
			return float64(av.Value) == bv.Value
		}
	case *PyFloat:
		switch bv := b.(type) {
		case *PyFloat:
			return av.Value == bv.Value
		case *PyInt:
			return av.Value == float64(bv.Value)
		}
	case *PyString:
		if bv, ok := b.(*PyString); ok {
			return av.Value == bv.Value
		}
	case *PyList:
		if bv, ok := b.(*PyList); ok {
			if len(av.Items) != len(bv.Items) {
				return false
			}
			// Cycle detection: check if we've seen this pair
			ptrA := ptrValue(a)
			ptrB := ptrValue(b)
			if seen[ptrA] != nil && seen[ptrA][ptrB] {
				return true // Already comparing these, assume equal to break cycle
			}
			if seen[ptrA] == nil {
				seen[ptrA] = make(map[uintptr]bool)
			}
			seen[ptrA][ptrB] = true
			for i := range av.Items {
				if !vm.equalWithCycleDetection(av.Items[i], bv.Items[i], seen) {
					return false
				}
			}
			return true
		}
	case *PyTuple:
		if bv, ok := b.(*PyTuple); ok {
			if len(av.Items) != len(bv.Items) {
				return false
			}
			for i := range av.Items {
				if !vm.equalWithCycleDetection(av.Items[i], bv.Items[i], seen) {
					return false
				}
			}
			return true
		}
	case *PyDict:
		if bv, ok := b.(*PyDict); ok {
			if len(av.Items) != len(bv.Items) {
				return false
			}
			// Cycle detection for dicts
			ptrA := ptrValue(a)
			ptrB := ptrValue(b)
			if seen[ptrA] != nil && seen[ptrA][ptrB] {
				return true
			}
			if seen[ptrA] == nil {
				seen[ptrA] = make(map[uintptr]bool)
			}
			seen[ptrA][ptrB] = true
			for k, v := range av.Items {
				found := false
				for k2, v2 := range bv.Items {
					if vm.equalWithCycleDetection(k, k2, seen) {
						if vm.equalWithCycleDetection(v, v2, seen) {
							found = true
							break
						}
					}
				}
				if !found {
					return false
				}
			}
			return true
		}
	case *PyClass:
		// Classes are compared by identity
		return a == b
	case *PyInstance:
		// Instances are compared by identity by default
		return a == b
	}
	return false
}

func (vm *VM) compare(a, b Value) int {
	switch av := a.(type) {
	case *PyInt:
		switch bv := b.(type) {
		case *PyInt:
			if av.Value < bv.Value {
				return -1
			} else if av.Value > bv.Value {
				return 1
			}
			return 0
		case *PyFloat:
			af := float64(av.Value)
			if af < bv.Value {
				return -1
			} else if af > bv.Value {
				return 1
			}
			return 0
		}
	case *PyFloat:
		switch bv := b.(type) {
		case *PyFloat:
			if av.Value < bv.Value {
				return -1
			} else if av.Value > bv.Value {
				return 1
			}
			return 0
		case *PyInt:
			bf := float64(bv.Value)
			if av.Value < bf {
				return -1
			} else if av.Value > bf {
				return 1
			}
			return 0
		}
	case *PyString:
		if bv, ok := b.(*PyString); ok {
			if av.Value < bv.Value {
				return -1
			} else if av.Value > bv.Value {
				return 1
			}
			return 0
		}
	}
	return 0
}

func (vm *VM) contains(container, item Value) bool {
	switch c := container.(type) {
	case *PyString:
		if i, ok := item.(*PyString); ok {
			// Use strings.Contains for optimized O(n+m) substring search
			return strings.Contains(c.Value, i.Value)
		}
	case *PyList:
		for _, v := range c.Items {
			if vm.equal(v, item) {
				return true
			}
		}
	case *PyTuple:
		for _, v := range c.Items {
			if vm.equal(v, item) {
				return true
			}
		}
	case *PySet:
		// Use hash-based lookup for O(1) average case
		return c.SetContains(item, vm)
	case *PyDict:
		// Use hash-based lookup for O(1) average case
		return c.DictContains(item, vm)
	}
	return false
}

// Attribute access

func (vm *VM) getAttr(obj Value, name string) (Value, error) {
	switch o := obj.(type) {
	case *PyGenerator:
		gen := o
		switch name {
		case "__iter__":
			return &PyBuiltinFunc{Name: "generator.__iter__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return gen, nil
			}}, nil
		case "__next__":
			return &PyBuiltinFunc{Name: "generator.__next__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				val, done, err := vm.GeneratorSend(gen, None)
				if err != nil {
					return nil, err
				}
				if done {
					return nil, &PyException{TypeName: "StopIteration", Message: ""}
				}
				return val, nil
			}}, nil
		case "send":
			return &PyBuiltinFunc{Name: "generator.send", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var sendVal Value = None
				if len(args) > 0 {
					sendVal = args[0]
				}
				val, done, err := vm.GeneratorSend(gen, sendVal)
				if err != nil {
					return nil, err
				}
				if done {
					return nil, &PyException{TypeName: "StopIteration", Message: ""}
				}
				return val, nil
			}}, nil
		case "throw":
			return &PyBuiltinFunc{Name: "generator.throw", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var excType, excValue Value = &PyString{Value: "Exception"}, None
				if len(args) > 0 {
					excType = args[0]
				}
				if len(args) > 1 {
					excValue = args[1]
				}
				val, done, err := vm.GeneratorThrow(gen, excType, excValue)
				if err != nil {
					return nil, err
				}
				if done {
					return nil, &PyException{TypeName: "StopIteration", Message: ""}
				}
				return val, nil
			}}, nil
		case "close":
			return &PyBuiltinFunc{Name: "generator.close", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				err := vm.GeneratorClose(gen)
				if err != nil {
					return nil, err
				}
				return None, nil
			}}, nil
		}
		return nil, fmt.Errorf("'generator' object has no attribute '%s'", name)

	case *PyCoroutine:
		coro := o
		switch name {
		case "__await__":
			return &PyBuiltinFunc{Name: "coroutine.__await__", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				return coro, nil
			}}, nil
		case "send":
			return &PyBuiltinFunc{Name: "coroutine.send", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var sendVal Value = None
				if len(args) > 0 {
					sendVal = args[0]
				}
				val, done, err := vm.CoroutineSend(coro, sendVal)
				if err != nil {
					return nil, err
				}
				if done {
					return nil, &PyException{TypeName: "StopIteration", Message: ""}
				}
				return val, nil
			}}, nil
		case "throw":
			return &PyBuiltinFunc{Name: "coroutine.throw", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				// Similar to generator.throw
				return nil, &PyException{TypeName: "NotImplementedError", Message: "coroutine.throw not yet implemented"}
			}}, nil
		case "close":
			return &PyBuiltinFunc{Name: "coroutine.close", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				coro.State = GenClosed
				return None, nil
			}}, nil
		}
		return nil, fmt.Errorf("'coroutine' object has no attribute '%s'", name)

	case *PyModule:
		if val, ok := o.Dict[name]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("module '%s' has no attribute '%s'", o.Name, name)
	case *PyUserData:
		// Look up method in metatable by type name
		if o.Metatable != nil {
			// Find __type__ key in metatable (iterate because Value keys use pointers)
			var typeName string
			for k, v := range o.Metatable.Items {
				if ks, ok := k.(*PyString); ok && ks.Value == "__type__" {
					typeName = vm.str(v)
					break
				}
			}
			if typeName != "" {
				mt := typeMetatables[typeName]
				if mt != nil {
					// Check properties first (like Python @property - called automatically)
					if mt.Properties != nil {
						if propGetter, ok := mt.Properties[name]; ok {
							// Call the property getter with userdata as arg 1
							ud := o
							oldStack := vm.frame.Stack
							oldSP := vm.frame.SP
							vm.frame.Stack = make([]Value, 17)
							vm.frame.Stack[0] = ud
							vm.frame.SP = 1
							n := propGetter(vm)
							var result Value = None
							if n > 0 {
								// Result was pushed onto stack after ud
								result = vm.frame.Stack[vm.frame.SP-1]
							}
							vm.frame.Stack = oldStack
							vm.frame.SP = oldSP
							return result, nil
						}
					}
					if method, ok := mt.Methods[name]; ok {
						// Capture the userdata and method in closure
						ud := o
						m := method
						// Return a bound method that includes the userdata as first arg
						return &PyGoFunc{
							Name: name,
							Fn: func(vm *VM) int {
								// Shift stack to insert userdata as first argument
								top := vm.GetTop()
								newStack := make([]Value, top+17) // Extra space for stack operations
								newStack[0] = ud
								for i := 0; i < top; i++ {
									newStack[i+1] = vm.Get(i + 1)
								}
								vm.frame.Stack = newStack
								vm.frame.SP = top + 1
								return m(vm)
							},
						}, nil
					}
				}
			}
		}
		return nil, fmt.Errorf("'%s' object has no attribute '%s'", vm.typeName(obj), name)
	case *PyProperty:
		prop := o
		switch name {
		case "setter":
			return &PyBuiltinFunc{
				Name: "property.setter",
				Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("setter() takes exactly 1 argument")
					}
					return &PyProperty{
						Fget: prop.Fget,
						Fset: args[0],
						Fdel: prop.Fdel,
						Doc:  prop.Doc,
					}, nil
				},
			}, nil
		case "deleter":
			return &PyBuiltinFunc{
				Name: "property.deleter",
				Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("deleter() takes exactly 1 argument")
					}
					return &PyProperty{
						Fget: prop.Fget,
						Fset: prop.Fset,
						Fdel: args[0],
						Doc:  prop.Doc,
					}, nil
				},
			}, nil
		case "getter":
			return &PyBuiltinFunc{
				Name: "property.getter",
				Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("getter() takes exactly 1 argument")
					}
					return &PyProperty{
						Fget: args[0],
						Fset: prop.Fset,
						Fdel: prop.Fdel,
						Doc:  prop.Doc,
					}, nil
				},
			}, nil
		case "fget":
			if prop.Fget != nil {
				return prop.Fget, nil
			}
			return None, nil
		case "fset":
			if prop.Fset != nil {
				return prop.Fset, nil
			}
			return None, nil
		case "fdel":
			if prop.Fdel != nil {
				return prop.Fdel, nil
			}
			return None, nil
		case "__doc__":
			return &PyString{Value: prop.Doc}, nil
		}
		return nil, fmt.Errorf("'property' object has no attribute '%s'", name)
	case *PySuper:
		// super() proxy - look up attribute in MRO starting after ThisClass
		if o.Instance == nil {
			return nil, fmt.Errorf("'super' object has no attribute '%s'", name)
		}

		// Get the MRO to search
		var mro []*PyClass
		var instance Value = o.Instance
		if inst, ok := o.Instance.(*PyInstance); ok {
			mro = inst.Class.Mro
		} else if cls, ok := o.Instance.(*PyClass); ok {
			mro = cls.Mro
			instance = cls
		}

		// Search MRO starting from StartIdx
		for i := o.StartIdx; i < len(mro); i++ {
			cls := mro[i]
			if val, ok := cls.Dict[name]; ok {
				// Handle classmethod - bind class
				if cm, ok := val.(*PyClassMethod); ok {
					if fn, ok := cm.Func.(*PyFunction); ok {
						if inst, ok := o.Instance.(*PyInstance); ok {
							return &PyMethod{Func: fn, Instance: inst.Class}, nil
						}
						return &PyMethod{Func: fn, Instance: instance}, nil
					}
				}

				// Handle staticmethod - return unwrapped function
				if sm, ok := val.(*PyStaticMethod); ok {
					return sm.Func, nil
				}

				// Handle property - call fget with the original instance
				if prop, ok := val.(*PyProperty); ok {
					if prop.Fget == nil {
						return nil, fmt.Errorf("property '%s' has no getter", name)
					}
					return vm.call(prop.Fget, []Value{instance}, nil)
				}

				// Bind method if it's a function - bind to original instance
				if fn, ok := val.(*PyFunction); ok {
					return &PyMethod{Func: fn, Instance: instance}, nil
				}
				return val, nil
			}
		}
		return nil, fmt.Errorf("'super' object has no attribute '%s'", name)

	case *PyInstance:
		// Descriptor protocol: First check class MRO for data descriptors (property with setter)
		// Data descriptors take precedence over instance dict
		for _, cls := range o.Class.Mro {
			if val, ok := cls.Dict[name]; ok {
				if prop, ok := val.(*PyProperty); ok {
					// Property is a data descriptor - call fget
					if prop.Fget == nil {
						return nil, fmt.Errorf("property '%s' has no getter", name)
					}
					return vm.call(prop.Fget, []Value{obj}, nil)
				}
			}
		}

		// Then check instance dict
		if val, ok := o.Dict[name]; ok {
			return val, nil
		}

		// Then check class MRO for methods/attributes (non-data descriptors)
		for _, cls := range o.Class.Mro {
			if val, ok := cls.Dict[name]; ok {
				// Handle classmethod - bind class instead of instance
				if cm, ok := val.(*PyClassMethod); ok {
					if fn, ok := cm.Func.(*PyFunction); ok {
						return &PyMethod{Func: fn, Instance: o.Class}, nil
					}
					// For non-PyFunction callables, return a wrapper
					return &PyBuiltinFunc{
						Name: "bound_classmethod",
						Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
							newArgs := make([]Value, len(args)+1)
							newArgs[0] = o.Class
							copy(newArgs[1:], args)
							return vm.call(cm.Func, newArgs, kwargs)
						},
					}, nil
				}

				// Handle staticmethod - return unwrapped function
				if sm, ok := val.(*PyStaticMethod); ok {
					return sm.Func, nil
				}

				// Bind method if it's a function
				if fn, ok := val.(*PyFunction); ok {
					return &PyMethod{Func: fn, Instance: obj}, nil
				}
				return val, nil
			}
		}
		return nil, fmt.Errorf("'%s' object has no attribute '%s'", o.Class.Name, name)
	case *PyClass:
		// Handle special class attributes
		switch name {
		case "__mro__":
			// Return the MRO as a tuple
			mroItems := make([]Value, len(o.Mro))
			for i, cls := range o.Mro {
				mroItems[i] = cls
			}
			return &PyTuple{Items: mroItems}, nil
		case "__bases__":
			// Return the direct base classes as a tuple
			baseItems := make([]Value, len(o.Bases))
			for i, base := range o.Bases {
				baseItems[i] = base
			}
			return &PyTuple{Items: baseItems}, nil
		case "__name__":
			return &PyString{Value: o.Name}, nil
		case "__dict__":
			// Return a copy of the class dict
			dictCopy := make(map[Value]Value)
			for k, v := range o.Dict {
				dictCopy[&PyString{Value: k}] = v
			}
			return &PyDict{Items: dictCopy}, nil
		}
		// Check class dict and MRO
		for _, cls := range o.Mro {
			if val, ok := cls.Dict[name]; ok {
				// Handle classmethod - bind with class
				if cm, ok := val.(*PyClassMethod); ok {
					if fn, ok := cm.Func.(*PyFunction); ok {
						return &PyMethod{Func: fn, Instance: o}, nil
					}
					// For non-PyFunction callables, return a wrapper
					return &PyBuiltinFunc{
						Name: "bound_classmethod",
						Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
							newArgs := make([]Value, len(args)+1)
							newArgs[0] = o
							copy(newArgs[1:], args)
							return vm.call(cm.Func, newArgs, kwargs)
						},
					}, nil
				}

				// Handle staticmethod - return unwrapped function
				if sm, ok := val.(*PyStaticMethod); ok {
					return sm.Func, nil
				}

				// Handle property on class access - return the property object itself
				// (In Python, accessing a property on the class returns the property object)
				if _, ok := val.(*PyProperty); ok {
					return val, nil
				}

				return val, nil
			}
		}
		return nil, fmt.Errorf("type object '%s' has no attribute '%s'", o.Name, name)
	case *PyDict:
		if name == "get" {
			return &PyBuiltinFunc{Name: "dict.get", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("get() requires at least 1 argument")
				}
				key := args[0]
				def := Value(None)
				if len(args) > 1 {
					def = args[1]
				}
				if val, ok := o.Items[key]; ok {
					return val, nil
				}
				return def, nil
			}}, nil
		}
		if name == "keys" {
			return &PyBuiltinFunc{Name: "dict.keys", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var keys []Value
				for k := range o.Items {
					keys = append(keys, k)
				}
				return &PyList{Items: keys}, nil
			}}, nil
		}
		if name == "values" {
			return &PyBuiltinFunc{Name: "dict.values", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var vals []Value
				for _, v := range o.Items {
					vals = append(vals, v)
				}
				return &PyList{Items: vals}, nil
			}}, nil
		}
		if name == "items" {
			return &PyBuiltinFunc{Name: "dict.items", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				var items []Value
				for k, v := range o.Items {
					items = append(items, &PyTuple{Items: []Value{k, v}})
				}
				return &PyList{Items: items}, nil
			}}, nil
		}
	case *PyList:
		if name == "append" {
			return &PyBuiltinFunc{Name: "list.append", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("append() takes exactly 1 argument")
				}
				o.Items = append(o.Items, args[0])
				return None, nil
			}}, nil
		}
		if name == "pop" {
			return &PyBuiltinFunc{Name: "list.pop", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(o.Items) == 0 {
					return nil, fmt.Errorf("pop from empty list")
				}
				idx := len(o.Items) - 1
				if len(args) > 0 {
					idx = int(vm.toInt(args[0]))
				}
				val := o.Items[idx]
				o.Items = append(o.Items[:idx], o.Items[idx+1:]...)
				return val, nil
			}}, nil
		}
		if name == "extend" {
			return &PyBuiltinFunc{Name: "list.extend", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("extend() takes exactly 1 argument")
				}
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				o.Items = append(o.Items, items...)
				return None, nil
			}}, nil
		}
	case *PyString:
		if name == "upper" {
			return &PyBuiltinFunc{Name: "str.upper", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				result := ""
				for _, ch := range o.Value {
					if ch >= 'a' && ch <= 'z' {
						result += string(ch - 32)
					} else {
						result += string(ch)
					}
				}
				return &PyString{Value: result}, nil
			}}, nil
		}
		if name == "lower" {
			return &PyBuiltinFunc{Name: "str.lower", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				result := ""
				for _, ch := range o.Value {
					if ch >= 'A' && ch <= 'Z' {
						result += string(ch + 32)
					} else {
						result += string(ch)
					}
				}
				return &PyString{Value: result}, nil
			}}, nil
		}
		if name == "split" {
			return &PyBuiltinFunc{Name: "str.split", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				sep := " "
				if len(args) > 0 {
					sep = vm.str(args[0])
				}
				var parts []Value
				current := ""
				// Manually manage loop index to avoid confusion with separator length
				for i := 0; i < len(o.Value); {
					if i+len(sep) <= len(o.Value) && o.Value[i:i+len(sep)] == sep {
						parts = append(parts, &PyString{Value: current})
						current = ""
						i += len(sep)
					} else {
						current += string(o.Value[i])
						i++
					}
				}
				parts = append(parts, &PyString{Value: current})
				return &PyList{Items: parts}, nil
			}}, nil
		}
		if name == "join" {
			return &PyBuiltinFunc{Name: "str.join", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("join() takes exactly 1 argument")
				}
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				var parts []string
				for _, item := range items {
					parts = append(parts, vm.str(item))
				}
				result := ""
				for i, p := range parts {
					if i > 0 {
						result += o.Value
					}
					result += p
				}
				return &PyString{Value: result}, nil
			}}, nil
		}
		if name == "strip" {
			return &PyBuiltinFunc{Name: "str.strip", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				s := o.Value
				start := 0
				end := len(s)
				for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
					start++
				}
				for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
					end--
				}
				return &PyString{Value: s[start:end]}, nil
			}}, nil
		}
		if name == "replace" {
			return &PyBuiltinFunc{Name: "str.replace", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
				if len(args) < 2 {
					return nil, fmt.Errorf("replace() takes at least 2 arguments")
				}
				old := vm.str(args[0])
				new := vm.str(args[1])
				result := ""
				for i := 0; i < len(o.Value); {
					if i+len(old) <= len(o.Value) && o.Value[i:i+len(old)] == old {
						result += new
						i += len(old)
					} else {
						result += string(o.Value[i])
						i++
					}
				}
				return &PyString{Value: result}, nil
			}}, nil
		}
	}
	return nil, fmt.Errorf("'%s' object has no attribute '%s'", vm.typeName(obj), name)
}

func (vm *VM) setAttr(obj Value, name string, val Value) error {
	switch o := obj.(type) {
	case *PyInstance:
		// Check for property with setter in class MRO
		for _, cls := range o.Class.Mro {
			if clsVal, ok := cls.Dict[name]; ok {
				if prop, ok := clsVal.(*PyProperty); ok {
					if prop.Fset == nil {
						return fmt.Errorf("property '%s' has no setter", name)
					}
					// Call the setter with (instance, value)
					_, err := vm.call(prop.Fset, []Value{obj, val}, nil)
					return err
				}
				break // Found in class dict but not a property, fall through to instance assignment
			}
		}
		// Not a property, set on instance dict
		o.Dict[name] = val
		return nil
	case *PyClass:
		o.Dict[name] = val
		return nil
	}
	return fmt.Errorf("'%s' object attribute '%s' is read-only", vm.typeName(obj), name)
}

// Item access

func (vm *VM) getItem(obj Value, index Value) (Value, error) {
	switch o := obj.(type) {
	case *PyList:
		idx := int(vm.toInt(index))
		if idx < 0 {
			idx = len(o.Items) + idx
		}
		if idx < 0 || idx >= len(o.Items) {
			return nil, fmt.Errorf("list index out of range")
		}
		return o.Items[idx], nil
	case *PyTuple:
		idx := int(vm.toInt(index))
		if idx < 0 {
			idx = len(o.Items) + idx
		}
		if idx < 0 || idx >= len(o.Items) {
			return nil, fmt.Errorf("tuple index out of range")
		}
		return o.Items[idx], nil
	case *PyString:
		// Convert to runes for proper UTF-8 character indexing
		runes := []rune(o.Value)
		idx := int(vm.toInt(index))
		if idx < 0 {
			idx = len(runes) + idx
		}
		if idx < 0 || idx >= len(runes) {
			return nil, fmt.Errorf("string index out of range")
		}
		return &PyString{Value: string(runes[idx])}, nil
	case *PyDict:
		// Use hash-based lookup for O(1) average case
		if val, found := o.DictGet(index, vm); found {
			return val, nil
		}
		return nil, fmt.Errorf("KeyError: %v", index)
	case *PyUserData:
		// Check for __getitem__ method in metatable
		if o.Metatable != nil {
			var typeName string
			for k, v := range o.Metatable.Items {
				if ks, ok := k.(*PyString); ok && ks.Value == "__type__" {
					typeName = vm.str(v)
					break
				}
			}
			if typeName != "" {
				mt := typeMetatables[typeName]
				if mt != nil {
					if method, ok := mt.Methods["__getitem__"]; ok {
						// Call __getitem__ with userdata and index
						oldStack := vm.frame.Stack
						oldSP := vm.frame.SP
						vm.frame.Stack = make([]Value, 17)
						vm.frame.Stack[0] = o
						vm.frame.Stack[1] = index
						vm.frame.SP = 2
						n := method(vm)
						var result Value = None
						if n > 0 {
							result = vm.frame.Stack[vm.frame.SP-1]
						}
						vm.frame.Stack = oldStack
						vm.frame.SP = oldSP
						return result, nil
					}
				}
			}
		}
		return nil, fmt.Errorf("'%s' object is not subscriptable", vm.typeName(obj))
	}
	return nil, fmt.Errorf("'%s' object is not subscriptable", vm.typeName(obj))
}

func (vm *VM) setItem(obj Value, index Value, val Value) error {
	switch o := obj.(type) {
	case *PyList:
		idx := int(vm.toInt(index))
		if idx < 0 {
			idx = len(o.Items) + idx
		}
		if idx < 0 || idx >= len(o.Items) {
			return fmt.Errorf("list assignment index out of range")
		}
		o.Items[idx] = val
		return nil
	case *PyDict:
		if !isHashable(index) {
			return fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(index))
		}
		// Use hash-based storage for O(1) average case
		o.DictSet(index, val, vm)
		return nil
	}
	return fmt.Errorf("'%s' object does not support item assignment", vm.typeName(obj))
}

func (vm *VM) delItem(obj Value, index Value) error {
	switch o := obj.(type) {
	case *PyList:
		idx := int(vm.toInt(index))
		if idx < 0 {
			idx = len(o.Items) + idx
		}
		if idx < 0 || idx >= len(o.Items) {
			return fmt.Errorf("list assignment index out of range")
		}
		o.Items = append(o.Items[:idx], o.Items[idx+1:]...)
		return nil
	case *PyDict:
		// Use hash-based deletion for O(1) average case
		o.DictDelete(index, vm)
		return nil
	}
	return fmt.Errorf("'%s' object does not support item deletion", vm.typeName(obj))
}

// Iterator

func (vm *VM) getIter(obj Value) (Value, error) {
	// Generators and coroutines are their own iterators
	switch v := obj.(type) {
	case *PyGenerator:
		return v, nil
	case *PyCoroutine:
		return v, nil
	case *PyIterator:
		return v, nil
	}

	// Try __iter__ method first
	if iterMethod, err := vm.getAttr(obj, "__iter__"); err == nil {
		result, err := vm.call(iterMethod, nil, nil)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	// Fall back to converting to list
	items, err := vm.toList(obj)
	if err != nil {
		return nil, err
	}
	return &PyIterator{Items: items, Index: 0}, nil
}

// iterNext advances an iterator and returns the next value
// Returns (value, done, error) where done is true if the iterator is exhausted
func (vm *VM) iterNext(iter Value) (Value, bool, error) {
	switch it := iter.(type) {
	case *PyIterator:
		if it.Index < len(it.Items) {
			val := it.Items[it.Index]
			it.Index++
			return val, false, nil
		}
		return nil, true, nil

	case *PyGenerator:
		val, done, err := vm.GeneratorSend(it, None)
		if err != nil {
			// StopIteration is not an error for iteration
			if pyErr, ok := err.(*PyException); ok && pyErr.Type() == "StopIteration" {
				return nil, true, nil
			}
			return nil, false, err
		}
		return val, done, nil

	case *PyCoroutine:
		val, done, err := vm.CoroutineSend(it, None)
		if err != nil {
			if pyErr, ok := err.(*PyException); ok && pyErr.Type() == "StopIteration" {
				return nil, true, nil
			}
			return nil, false, err
		}
		return val, done, nil

	default:
		// Try __next__ method
		nextMethod, err := vm.getAttr(iter, "__next__")
		if err != nil {
			return nil, false, fmt.Errorf("'%s' object is not an iterator", vm.typeName(iter))
		}
		val, err := vm.call(nextMethod, nil, nil)
		if err != nil {
			if pyErr, ok := err.(*PyException); ok && pyErr.Type() == "StopIteration" {
				return nil, true, nil
			}
			return nil, false, err
		}
		return val, false, nil
	}
}

// Function calls

func (vm *VM) call(callable Value, args []Value, kwargs map[string]Value) (Value, error) {
	switch fn := callable.(type) {
	case *PyBuiltinFunc:
		return fn.Fn(args, kwargs)

	case *PyGoFunc:
		// Call Go function with gopher-lua style stack-based API
		return vm.callGoFunction(fn, args)

	case *PyFunction:
		return vm.callFunction(fn, args, kwargs)

	case *PyMethod:
		// Prepend instance to args
		allArgs := append([]Value{fn.Instance}, args...)
		return vm.callFunction(fn.Func, allArgs, kwargs)

	case *PyClass:
		// Create instance and call __init__
		instance := &PyInstance{
			Class: fn,
			Dict:  make(map[string]Value),
		}
		// Look for __init__ in class MRO
		for _, cls := range fn.Mro {
			if init, ok := cls.Dict["__init__"]; ok {
				if initFn, ok := init.(*PyFunction); ok {
					allArgs := append([]Value{instance}, args...)
					_, err := vm.callFunction(initFn, allArgs, kwargs)
					if err != nil {
						return nil, err
					}
				}
				break
			}
		}
		return instance, nil
	}
	return nil, fmt.Errorf("'%s' object is not callable", vm.typeName(callable))
}

// callGoFunction calls a Go function with stack-based argument passing
func (vm *VM) callGoFunction(fn *PyGoFunc, args []Value) (Value, error) {
	// Save current frame state
	oldFrame := vm.frame

	// Create a temporary frame for the Go function call
	tempFrame := &Frame{
		Stack:    make([]Value, len(args)+16),
		SP:       0,
		Globals:  vm.Globals,
		Builtins: vm.builtins,
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
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Check if it's a typed PyPanicError for better exception mapping
				if pe, ok := r.(*PyPanicError); ok {
					panicErr = pe
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

	// Create new frame for regular function call
	frame := vm.createFunctionFrame(fn, args, kwargs)

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

	return result, err
}

// createFunctionFrame creates a new frame for a function call without executing it
func (vm *VM) createFunctionFrame(fn *PyFunction, args []Value, kwargs map[string]Value) *Frame {
	code := fn.Code

	// Create new frame
	frame := &Frame{
		Code:     code,
		IP:       0,
		Stack:    make([]Value, code.StackSize+16), // Pre-allocate
		SP:       0,
		Locals:   make([]Value, len(code.VarNames)),
		Globals:  fn.Globals,
		Builtins: vm.builtins,
	}

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

	// Bind arguments to locals
	for i, arg := range args {
		if i < len(frame.Locals) {
			frame.Locals[i] = arg
		}
	}

	// If any arguments are CellVars (captured by inner functions), initialize cells with args
	// CellVars contain names of captured parameters - match by name against VarNames
	for cellIdx, cellName := range code.CellVars {
		// Find if this cell var corresponds to a parameter (in first ArgCount VarNames)
		for argIdx := 0; argIdx < code.ArgCount && argIdx < len(code.VarNames); argIdx++ {
			if code.VarNames[argIdx] == cellName && argIdx < len(args) {
				// This parameter is captured, initialize its cell with the argument
				if cellIdx < len(frame.Cells) && frame.Cells[cellIdx] != nil {
					frame.Cells[cellIdx].Value = args[argIdx]
				}
				break
			}
		}
	}

	// Apply keyword arguments to the appropriate local slots
	if kwargs != nil {
		for name, val := range kwargs {
			// Find the parameter index by name
			for i, varName := range code.VarNames {
				if varName == name && i < code.ArgCount {
					frame.Locals[i] = val
					break
				}
			}
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

	return frame
}

// createGenerator creates a new generator object from a generator function
func (vm *VM) createGenerator(fn *PyFunction, args []Value, kwargs map[string]Value) (*PyGenerator, error) {
	frame := vm.createFunctionFrame(fn, args, kwargs)
	return &PyGenerator{
		Frame: frame,
		Code:  fn.Code,
		Name:  fn.Name,
		State: GenCreated,
	}, nil
}

// createCoroutine creates a new coroutine object from an async function
func (vm *VM) createCoroutine(fn *PyFunction, args []Value, kwargs map[string]Value) (*PyCoroutine, error) {
	frame := vm.createFunctionFrame(fn, args, kwargs)
	return &PyCoroutine{
		Frame: frame,
		Code:  fn.Code,
		Name:  fn.Name,
		State: GenCreated,
	}, nil
}

// GeneratorSend resumes a generator with a value and returns the next yielded value
// Returns (value, done, error) where done is true if the generator finished
func (vm *VM) GeneratorSend(gen *PyGenerator, value Value) (Value, bool, error) {
	if gen.State == GenClosed {
		return nil, true, &PyException{
			TypeName: "StopIteration",
			Message:  "generator already closed",
		}
	}

	if gen.State == GenRunning {
		return nil, false, &PyException{
			TypeName: "ValueError",
			Message:  "generator already executing",
		}
	}

	// For the first call (GenCreated), value must be None
	if gen.State == GenCreated && value != nil && value != None {
		return nil, false, &PyException{
			TypeName: "TypeError",
			Message:  "can't send non-None value to a just-started generator",
		}
	}

	gen.State = GenRunning
	gen.YieldValue = value

	// Save current frame and switch to generator's frame
	oldFrame := vm.frame
	oldFrames := vm.frames

	vm.frame = gen.Frame
	vm.frames = []*Frame{gen.Frame}

	// If resuming from yield, push the sent value onto the stack
	if gen.Frame.IP > 0 {
		vm.push(value)
	}

	// Run until yield or return
	result, yielded, err := vm.runGenerator()

	// Restore old frame
	vm.frame = oldFrame
	vm.frames = oldFrames

	if err != nil {
		gen.State = GenClosed
		return nil, true, err
	}

	if yielded {
		gen.State = GenSuspended
		return result, false, nil
	}

	// Generator returned (finished)
	gen.State = GenClosed
	// Raise StopIteration with the return value
	return nil, true, &PyException{
		TypeName: "StopIteration",
		Message:  "generator finished",
		Args:     &PyTuple{Items: []Value{result}},
	}
}

// GeneratorThrow throws an exception into a generator
func (vm *VM) GeneratorThrow(gen *PyGenerator, excType, excValue Value) (Value, bool, error) {
	if gen.State == GenClosed {
		// Re-raise the exception
		excMsg := "exception"
		if str, ok := excValue.(*PyString); ok {
			excMsg = str.Value
		}
		return nil, true, &PyException{
			TypeName: fmt.Sprintf("%v", excType),
			Message:  excMsg,
		}
	}

	if gen.State == GenRunning {
		return nil, false, &PyException{
			TypeName: "ValueError",
			Message:  "generator already executing",
		}
	}

	if gen.State == GenCreated {
		gen.State = GenClosed
		excMsg := "exception"
		if str, ok := excValue.(*PyString); ok {
			excMsg = str.Value
		}
		return nil, true, &PyException{
			TypeName: fmt.Sprintf("%v", excType),
			Message:  excMsg,
		}
	}

	// Generator is suspended at a yield point - throw the exception there
	// Create the exception to throw
	excMsg := "exception"
	if str, ok := excValue.(*PyString); ok {
		excMsg = str.Value
	} else if excValue != nil && excValue != None {
		excMsg = fmt.Sprintf("%v", excValue)
	}

	exc := &PyException{
		TypeName: fmt.Sprintf("%v", excType),
		Message:  excMsg,
	}

	// Check if excType is a class and set ExcType appropriately
	if cls, ok := excType.(*PyClass); ok {
		exc.ExcType = cls
	}

	// Set the pending exception on the VM for the generator to handle
	vm.generatorThrow = exc
	gen.State = GenRunning

	// Save current frame and switch to generator's frame
	oldFrame := vm.frame
	oldFrames := vm.frames

	vm.frame = gen.Frame
	vm.frames = []*Frame{gen.Frame}

	// Run until yield or return (exception will be handled in runWithYieldSupport)
	result, yielded, err := vm.runGenerator()

	// Restore old frame
	vm.frame = oldFrame
	vm.frames = oldFrames

	if err != nil {
		gen.State = GenClosed
		return nil, true, err
	}

	if yielded {
		gen.State = GenSuspended
		return result, false, nil
	}

	// Generator returned (finished)
	gen.State = GenClosed
	return nil, true, &PyException{
		TypeName: "StopIteration",
		Message:  "generator finished",
		Args:     &PyTuple{Items: []Value{result}},
	}
}

// GeneratorClose closes a generator
func (vm *VM) GeneratorClose(gen *PyGenerator) error {
	if gen.State == GenClosed {
		return nil
	}

	if gen.State == GenCreated {
		gen.State = GenClosed
		return nil
	}

	// Throw GeneratorExit into the generator
	_, _, err := vm.GeneratorThrow(gen, &PyString{Value: "GeneratorExit"}, None)

	// GeneratorExit is expected - if we get it back, ignore it
	if pyErr, ok := err.(*PyException); ok && pyErr.Type() == "GeneratorExit" {
		gen.State = GenClosed
		return nil
	}

	// Any other exception should be propagated
	gen.State = GenClosed
	return err
}

// CoroutineSend resumes a coroutine with a value (same as generator but for coroutines)
func (vm *VM) CoroutineSend(coro *PyCoroutine, value Value) (Value, bool, error) {
	if coro.State == GenClosed {
		return nil, true, &PyException{
			TypeName: "StopIteration",
			Message:  "coroutine already closed",
		}
	}

	if coro.State == GenRunning {
		return nil, false, &PyException{
			TypeName: "ValueError",
			Message:  "coroutine already executing",
		}
	}

	if coro.State == GenCreated && value != nil && value != None {
		return nil, false, &PyException{
			TypeName: "TypeError",
			Message:  "can't send non-None value to a just-started coroutine",
		}
	}

	coro.State = GenRunning
	coro.YieldValue = value

	oldFrame := vm.frame
	oldFrames := vm.frames

	vm.frame = coro.Frame
	vm.frames = []*Frame{coro.Frame}

	if coro.Frame.IP > 0 {
		vm.push(value)
	}

	result, yielded, err := vm.runGenerator()

	vm.frame = oldFrame
	vm.frames = oldFrames

	if err != nil {
		coro.State = GenClosed
		return nil, true, err
	}

	if yielded {
		coro.State = GenSuspended
		return result, false, nil
	}

	coro.State = GenClosed
	return nil, true, &PyException{
		TypeName: "StopIteration",
		Message:  "coroutine finished",
		Args:     &PyTuple{Items: []Value{result}},
	}
}

// runGenerator runs a generator frame until yield or return
// Returns (value, yielded, error) where yielded is true if we hit a yield
func (vm *VM) runGenerator() (Value, bool, error) {
	// Use a wrapper that intercepts yield operations
	// We call runWithYieldSupport which is like run() but returns on yield
	return vm.runWithYieldSupport()
}

// runWithYieldSupport is like run() but returns (value, yielded, error) to support generators
func (vm *VM) runWithYieldSupport() (Value, bool, error) {
	frame := vm.frame

	// Check for pending exception from generator.throw()
	if vm.generatorThrow != nil {
		exc := vm.generatorThrow
		vm.generatorThrow = nil // Clear it

		// Handle the exception - this will look for handlers in the block stack
		_, err := vm.handleException(exc)
		if err != nil {
			// No handler found, propagate the exception
			return nil, false, err
		}
		// Handler found, frame.IP was updated to handler address
		// Continue execution at the handler
		frame = vm.frame // Update frame reference in case it changed
	}

	for frame.IP < len(frame.Code.Code) {
		// Check for timeout/cancellation periodically
		if vm.ctx != nil {
			vm.checkCounter--
			if vm.checkCounter <= 0 {
				vm.checkCounter = vm.checkInterval
				select {
				case <-vm.ctx.Done():
					if vm.ctx.Err() == context.DeadlineExceeded {
						return nil, false, &TimeoutError{}
					}
					return nil, false, &CancelledError{}
				default:
				}
			}
		}

		op := Opcode(frame.Code.Code[frame.IP])
		frame.IP++

		var arg int
		if op.HasArg() {
			arg = int(frame.Code.Code[frame.IP]) | int(frame.Code.Code[frame.IP+1])<<8
			frame.IP += 2
		}

		// Handle yield opcodes specially - these cause suspension
		switch op {
		case OpYieldValue:
			value := vm.pop()
			return value, true, nil

		case OpYieldFrom:
			// Get the iterator from the stack (don't pop it yet)
			iter := vm.top()

			// Try to get next value from iterator
			switch it := iter.(type) {
			case *PyGenerator:
				// For the first iteration, send None
				sendVal := None
				val, done, err := vm.GeneratorSend(it, sendVal)
				if err != nil {
					if pyErr, ok := err.(*PyException); ok && pyErr.Type() == "StopIteration" {
						vm.pop() // Pop the iterator
						if pyErr.Args != nil && len(pyErr.Args.Items) > 0 {
							vm.push(pyErr.Args.Items[0])
						} else {
							vm.push(None)
						}
						continue
					}
					return nil, false, err
				}
				if done {
					vm.pop()
					vm.push(None)
					continue
				}
				// Back up IP so we re-execute OpYieldFrom on resume
				frame.IP -= 1
				if OpYieldFrom.HasArg() {
					frame.IP -= 2
				}
				return val, true, nil

			case *PyCoroutine:
				val, done, err := vm.CoroutineSend(it, None)
				if err != nil {
					if pyErr, ok := err.(*PyException); ok && pyErr.Type() == "StopIteration" {
						vm.pop()
						if pyErr.Args != nil && len(pyErr.Args.Items) > 0 {
							vm.push(pyErr.Args.Items[0])
						} else {
							vm.push(None)
						}
						continue
					}
					return nil, false, err
				}
				if done {
					vm.pop()
					vm.push(None)
					continue
				}
				frame.IP -= 1
				if OpYieldFrom.HasArg() {
					frame.IP -= 2
				}
				return val, true, nil

			case *PyIterator:
				if it.Index >= len(it.Items) {
					vm.pop()
					vm.push(None)
					continue
				}
				val := it.Items[it.Index]
				it.Index++
				frame.IP -= 1
				if OpYieldFrom.HasArg() {
					frame.IP -= 2
				}
				return val, true, nil

			default:
				// Try to get next value from the iterator
				nextVal, done, err := vm.iterNext(iter)
				if err != nil {
					return nil, false, err
				}
				if done {
					vm.pop()
					vm.push(None)
					continue
				}
				frame.IP -= 1
				if OpYieldFrom.HasArg() {
					frame.IP -= 2
				}
				return nextVal, true, nil
			}

		case OpReturn:
			frame.SP--
			result := frame.Stack[frame.SP]
			return result, false, nil

		case OpGenStart:
			// No-op, just marks generator start
			continue

		default:
			// Execute regular opcode using the standard dispatcher
			result, err := vm.executeOpcodeForGenerator(op, arg)
			if err != nil {
				return nil, false, err
			}
			if result != nil {
				// Some opcodes return values (shouldn't happen in generator context)
				return result, false, nil
			}
		}
	}

	return None, false, nil
}

// executeOpcodeForGenerator executes a single opcode in generator context
// Returns (result, error) - result is non-nil only if execution should stop
func (vm *VM) executeOpcodeForGenerator(op Opcode, arg int) (Value, error) {
	frame := vm.frame

	switch op {
	case OpPop:
		vm.pop()
	case OpDup:
		vm.push(vm.top())
	case OpRot2:
		a := vm.pop()
		b := vm.pop()
		vm.push(a)
		vm.push(b)
	case OpRot3:
		a := vm.pop()
		b := vm.pop()
		c := vm.pop()
		vm.push(a)
		vm.push(c)
		vm.push(b)
	case OpLoadConst:
		if frame.SP >= len(frame.Stack) {
			vm.ensureStack(1)
		}
		frame.Stack[frame.SP] = vm.toValue(frame.Code.Constants[arg])
		frame.SP++
	case OpLoadName:
		name := frame.Code.Names[arg]
		if val, ok := frame.Globals[name]; ok {
			vm.push(val)
		} else if frame.EnclosingGlobals != nil {
			if val, ok := frame.EnclosingGlobals[name]; ok {
				vm.push(val)
			} else if val, ok := frame.Builtins[name]; ok {
				vm.push(val)
			} else {
				return nil, fmt.Errorf("name '%s' is not defined", name)
			}
		} else if val, ok := frame.Builtins[name]; ok {
			vm.push(val)
		} else {
			return nil, fmt.Errorf("name '%s' is not defined", name)
		}
	case OpStoreName:
		name := frame.Code.Names[arg]
		frame.Globals[name] = vm.pop()
	case OpLoadFast:
		frame.Stack[frame.SP] = frame.Locals[arg]
		frame.SP++
	case OpStoreFast:
		frame.SP--
		frame.Locals[arg] = frame.Stack[frame.SP]
	case OpLoadGlobal:
		name := frame.Code.Names[arg]
		if val, ok := frame.Globals[name]; ok {
			vm.push(val)
		} else if frame.EnclosingGlobals != nil {
			if val, ok := frame.EnclosingGlobals[name]; ok {
				vm.push(val)
			} else if val, ok := frame.Builtins[name]; ok {
				vm.push(val)
			} else {
				return nil, fmt.Errorf("name '%s' is not defined", name)
			}
		} else if val, ok := frame.Builtins[name]; ok {
			vm.push(val)
		} else {
			return nil, fmt.Errorf("name '%s' is not defined", name)
		}
	case OpStoreGlobal:
		name := frame.Code.Names[arg]
		frame.Globals[name] = vm.pop()

	// Binary operations
	case OpBinaryAdd:
		b := vm.pop()
		a := vm.pop()
		result, err := vm.binaryOp(OpBinaryAdd, a, b)
		if err != nil {
			return nil, err
		}
		vm.push(result)
	case OpBinarySubtract:
		b := vm.pop()
		a := vm.pop()
		result, err := vm.binaryOp(OpBinarySubtract, a, b)
		if err != nil {
			return nil, err
		}
		vm.push(result)
	case OpBinaryMultiply:
		b := vm.pop()
		a := vm.pop()
		result, err := vm.binaryOp(OpBinaryMultiply, a, b)
		if err != nil {
			return nil, err
		}
		vm.push(result)
	case OpBinaryDivide:
		b := vm.pop()
		a := vm.pop()
		result, err := vm.binaryOp(OpBinaryDivide, a, b)
		if err != nil {
			return nil, err
		}
		vm.push(result)
	case OpBinaryFloorDiv:
		b := vm.pop()
		a := vm.pop()
		result, err := vm.binaryOp(OpBinaryFloorDiv, a, b)
		if err != nil {
			return nil, err
		}
		vm.push(result)
	case OpBinaryModulo:
		b := vm.pop()
		a := vm.pop()
		result, err := vm.binaryOp(OpBinaryModulo, a, b)
		if err != nil {
			return nil, err
		}
		vm.push(result)

	// Comparison operations
	case OpCompareEq:
		b := vm.pop()
		a := vm.pop()
		if vm.equal(a, b) {
			vm.push(True)
		} else {
			vm.push(False)
		}
	case OpCompareNe:
		b := vm.pop()
		a := vm.pop()
		if !vm.equal(a, b) {
			vm.push(True)
		} else {
			vm.push(False)
		}
	case OpCompareLt:
		b := vm.pop()
		a := vm.pop()
		vm.push(vm.compareOp(OpCompareLt, a, b))
	case OpCompareLe:
		b := vm.pop()
		a := vm.pop()
		vm.push(vm.compareOp(OpCompareLe, a, b))
	case OpCompareGt:
		b := vm.pop()
		a := vm.pop()
		vm.push(vm.compareOp(OpCompareGt, a, b))
	case OpCompareGe:
		b := vm.pop()
		a := vm.pop()
		vm.push(vm.compareOp(OpCompareGe, a, b))

	// Unary operations
	case OpUnaryNot:
		a := vm.pop()
		if !vm.truthy(a) {
			vm.push(True)
		} else {
			vm.push(False)
		}
	case OpUnaryNegative:
		a := vm.pop()
		result, err := vm.unaryOp(OpUnaryNegative, a)
		if err != nil {
			return nil, err
		}
		vm.push(result)

	// Jump operations
	case OpJump:
		frame.IP = arg
	case OpPopJumpIfFalse:
		cond := vm.pop()
		if !vm.truthy(cond) {
			frame.IP = arg
		}
	case OpPopJumpIfTrue:
		cond := vm.pop()
		if vm.truthy(cond) {
			frame.IP = arg
		}
	case OpJumpIfTrueOrPop:
		if vm.truthy(vm.top()) {
			frame.IP = arg
		} else {
			vm.pop()
		}
	case OpJumpIfFalseOrPop:
		if !vm.truthy(vm.top()) {
			frame.IP = arg
		} else {
			vm.pop()
		}

	// Iteration
	case OpGetIter:
		obj := vm.pop()
		iter, err := vm.getIter(obj)
		if err != nil {
			return nil, err
		}
		vm.push(iter)
	case OpForIter:
		iter := vm.top()
		val, done, err := vm.iterNext(iter)
		if err != nil {
			return nil, err
		}
		if done {
			vm.pop()
			frame.IP = arg
		} else {
			vm.push(val)
		}

	// Function calls
	case OpCall:
		args := make([]Value, arg)
		for i := arg - 1; i >= 0; i-- {
			args[i] = vm.pop()
		}
		callable := vm.pop()
		result, err := vm.call(callable, args, nil)
		if err != nil {
			// Check if exception was already handled in an outer frame
			if err == errExceptionHandledInOuterFrame {
				// Exception handler was found, but it's in the generator's frame
				// Let runWithYieldSupport continue from the handler
				return nil, nil
			}
			// Check if it's a Python exception that can be handled
			if pyExc, ok := err.(*PyException); ok {
				_, handleErr := vm.handleException(pyExc)
				if handleErr != nil {
					return nil, handleErr
				}
				// Handler found - if it's in this frame, continue from handler
				// Otherwise, propagate the sentinel
				if vm.frame != frame {
					return nil, errExceptionHandledInOuterFrame
				}
				return nil, nil
			}
			return nil, err
		}
		vm.push(result)

	case OpMakeFunction:
		name := vm.pop().(*PyString)
		code := vm.pop().(*CodeObject)
		var defaults *PyTuple
		if arg&1 != 0 {
			defaults = vm.pop().(*PyTuple)
		}
		fn := &PyFunction{
			Code:     code,
			Globals:  frame.Globals,
			Defaults: defaults,
			Name:     name.Value,
		}
		// Handle closures
		if len(code.FreeVars) > 0 {
			fn.Closure = make([]*PyCell, len(code.FreeVars))
			for i, freeVar := range code.FreeVars {
				// Find in current frame's cells
				found := false
				for j, cellName := range frame.Code.CellVars {
					if cellName == freeVar && j < len(frame.Cells) {
						fn.Closure[i] = frame.Cells[j]
						found = true
						break
					}
				}
				if !found {
					for j, freeName := range frame.Code.FreeVars {
						if freeName == freeVar {
							idx := len(frame.Code.CellVars) + j
							if idx < len(frame.Cells) {
								fn.Closure[i] = frame.Cells[idx]
								found = true
								break
							}
						}
					}
				}
				if !found {
					fn.Closure[i] = &PyCell{}
				}
			}
		}
		vm.push(fn)

	// Collection building
	case OpBuildList:
		items := make([]Value, arg)
		for i := arg - 1; i >= 0; i-- {
			items[i] = vm.pop()
		}
		vm.push(&PyList{Items: items})
	case OpBuildTuple:
		items := make([]Value, arg)
		for i := arg - 1; i >= 0; i-- {
			items[i] = vm.pop()
		}
		vm.push(&PyTuple{Items: items})
	case OpBuildMap:
		dict := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
		for i := 0; i < arg; i++ {
			val := vm.pop()
			key := vm.pop()
			// Use hash-based storage for O(1) lookup
			dict.DictSet(key, val, vm)
		}
		vm.push(dict)

	// Attribute access
	case OpLoadAttr:
		name := frame.Code.Names[arg]
		obj := vm.pop()
		attr, err := vm.getAttr(obj, name)
		if err != nil {
			return nil, err
		}
		vm.push(attr)
	case OpStoreAttr:
		name := frame.Code.Names[arg]
		obj := vm.pop()
		val := vm.pop()
		if err := vm.setAttr(obj, name, val); err != nil {
			return nil, err
		}

	// Subscript
	case OpBinarySubscr:
		key := vm.pop()
		obj := vm.pop()
		val, err := vm.getItem(obj, key)
		if err != nil {
			return nil, err
		}
		vm.push(val)
	case OpStoreSubscr:
		key := vm.pop()
		obj := vm.pop()
		val := vm.pop()
		if err := vm.setItem(obj, key, val); err != nil {
			return nil, err
		}

	// Closure operations
	case OpLoadDeref:
		idx := arg
		if idx < len(frame.Cells) && frame.Cells[idx] != nil {
			vm.push(frame.Cells[idx].Value)
		} else {
			vm.push(None)
		}
	case OpStoreDeref:
		idx := arg
		val := vm.pop()
		if idx < len(frame.Cells) {
			if frame.Cells[idx] == nil {
				frame.Cells[idx] = &PyCell{}
			}
			frame.Cells[idx].Value = val
		}

	// Specialized ops for common constants
	case OpLoadNone:
		vm.push(None)
	case OpLoadTrue:
		vm.push(True)
	case OpLoadFalse:
		vm.push(False)
	case OpLoadZero:
		vm.push(MakeInt(0))
	case OpLoadOne:
		vm.push(MakeInt(1))

	// Specialized fast loads
	case OpLoadFast0:
		vm.push(frame.Locals[0])
	case OpLoadFast1:
		vm.push(frame.Locals[1])
	case OpLoadFast2:
		vm.push(frame.Locals[2])
	case OpLoadFast3:
		vm.push(frame.Locals[3])
	case OpStoreFast0:
		frame.Locals[0] = vm.pop()
	case OpStoreFast1:
		frame.Locals[1] = vm.pop()
	case OpStoreFast2:
		frame.Locals[2] = vm.pop()
	case OpStoreFast3:
		frame.Locals[3] = vm.pop()

	case OpNop:
		// No operation

	case OpListAppend:
		val := vm.pop()
		listIdx := frame.SP - arg
		if listIdx >= 0 && listIdx < frame.SP {
			if list, ok := frame.Stack[listIdx].(*PyList); ok {
				list.Items = append(list.Items, val)
			}
		}

	case OpLoadMethod:
		name := frame.Code.Names[arg]
		obj := vm.pop()
		method, err := vm.getAttr(obj, name)
		if err != nil {
			return nil, err
		}
		vm.push(obj)
		vm.push(method)

	case OpCallMethod:
		args := make([]Value, arg)
		for i := arg - 1; i >= 0; i-- {
			args[i] = vm.pop()
		}
		method := vm.pop()
		obj := vm.pop()
		var result Value
		var err error
		if _, isBound := method.(*PyMethod); isBound {
			result, err = vm.call(method, args, nil)
		} else {
			allArgs := append([]Value{obj}, args...)
			result, err = vm.call(method, allArgs, nil)
		}
		if err != nil {
			return nil, err
		}
		vm.push(result)

	case OpUnpackSequence:
		seq := vm.pop()
		items, err := vm.toList(seq)
		if err != nil {
			return nil, err
		}
		if len(items) != arg {
			return nil, fmt.Errorf("not enough values to unpack (expected %d, got %d)", arg, len(items))
		}
		for i := len(items) - 1; i >= 0; i-- {
			vm.push(items[i])
		}

	case OpCompareIn:
		container := vm.pop()
		item := vm.pop()
		if vm.contains(container, item) {
			vm.push(True)
		} else {
			vm.push(False)
		}

	case OpCompareNotIn:
		container := vm.pop()
		item := vm.pop()
		if !vm.contains(container, item) {
			vm.push(True)
		} else {
			vm.push(False)
		}

	case OpCompareIs:
		b := vm.pop()
		a := vm.pop()
		if a == b {
			vm.push(True)
		} else {
			vm.push(False)
		}

	case OpCompareIsNot:
		b := vm.pop()
		a := vm.pop()
		if a != b {
			vm.push(True)
		} else {
			vm.push(False)
		}

	case OpPrintExpr:
		val := vm.pop()
		if val != nil && val != None {
			if obj, ok := val.(PyObject); ok {
				fmt.Println(obj.String())
			} else {
				fmt.Println(val)
			}
		}

	// Async/coroutine opcodes
	case OpGetAwaitable:
		obj := vm.pop()
		// If it's already a coroutine, push it back
		if coro, ok := obj.(*PyCoroutine); ok {
			vm.push(coro)
		} else if gen, ok := obj.(*PyGenerator); ok {
			// Generators can also be awaitable
			vm.push(gen)
		} else {
			// Try to get __await__ method
			awaitable, err := vm.getAttr(obj, "__await__")
			if err != nil {
				// Just push the object itself for simple awaitables
				vm.push(obj)
			} else {
				// Call __await__ to get the awaitable iterator
				result, err := vm.call(awaitable, nil, nil)
				if err != nil {
					return nil, err
				}
				vm.push(result)
			}
		}

	case OpGetAIter:
		obj := vm.pop()
		// Try to get __aiter__ method
		aiter, err := vm.getAttr(obj, "__aiter__")
		if err != nil {
			return nil, fmt.Errorf("'%s' object is not async iterable", obj.(PyObject).Type())
		}
		result, err := vm.call(aiter, nil, nil)
		if err != nil {
			return nil, err
		}
		vm.push(result)

	case OpGetANext:
		obj := vm.top() // Don't pop - keep for next iteration
		// Try to get __anext__ method
		anext, err := vm.getAttr(obj, "__anext__")
		if err != nil {
			return nil, fmt.Errorf("'%s' object is not an async iterator", obj.(PyObject).Type())
		}
		result, err := vm.call(anext, nil, nil)
		if err != nil {
			return nil, err
		}
		vm.push(result)

	// In-place fast opcodes
	case OpIncrementFast:
		if v, ok := frame.Locals[arg].(*PyInt); ok {
			frame.Locals[arg] = MakeInt(v.Value + 1)
		} else {
			result, err := vm.binaryOp(OpBinaryAdd, frame.Locals[arg], MakeInt(1))
			if err != nil {
				return nil, err
			}
			frame.Locals[arg] = result
		}

	case OpDecrementFast:
		if v, ok := frame.Locals[arg].(*PyInt); ok {
			frame.Locals[arg] = MakeInt(v.Value - 1)
		} else {
			result, err := vm.binaryOp(OpBinarySubtract, frame.Locals[arg], MakeInt(1))
			if err != nil {
				return nil, err
			}
			frame.Locals[arg] = result
		}

	case OpNegateFast:
		if v, ok := frame.Locals[arg].(*PyInt); ok {
			frame.Locals[arg] = MakeInt(-v.Value)
		} else {
			result, err := vm.unaryOp(OpUnaryNegative, frame.Locals[arg])
			if err != nil {
				return nil, err
			}
			frame.Locals[arg] = result
		}

	case OpAddConstFast:
		localIdx := arg & 0xFF
		constIdx := (arg >> 8) & 0xFF
		constVal := vm.toValue(frame.Code.Constants[constIdx])
		if v, ok := frame.Locals[localIdx].(*PyInt); ok {
			if cv, ok := constVal.(*PyInt); ok {
				frame.Locals[localIdx] = MakeInt(v.Value + cv.Value)
			} else {
				result, err := vm.binaryOp(OpBinaryAdd, frame.Locals[localIdx], constVal)
				if err != nil {
					return nil, err
				}
				frame.Locals[localIdx] = result
			}
		} else {
			result, err := vm.binaryOp(OpBinaryAdd, frame.Locals[localIdx], constVal)
			if err != nil {
				return nil, err
			}
			frame.Locals[localIdx] = result
		}

	case OpAccumulateFast:
		val := vm.pop()
		if v, ok := frame.Locals[arg].(*PyInt); ok {
			if av, ok := val.(*PyInt); ok {
				frame.Locals[arg] = MakeInt(v.Value + av.Value)
			} else {
				result, err := vm.binaryOp(OpBinaryAdd, frame.Locals[arg], val)
				if err != nil {
					return nil, err
				}
				frame.Locals[arg] = result
			}
		} else {
			result, err := vm.binaryOp(OpBinaryAdd, frame.Locals[arg], val)
			if err != nil {
				return nil, err
			}
			frame.Locals[arg] = result
		}

	// Superinstruction opcodes
	case OpLoadFastLoadFast:
		idx1 := arg & 0xFF
		idx2 := (arg >> 8) & 0xFF
		vm.push(frame.Locals[idx1])
		vm.push(frame.Locals[idx2])

	case OpLoadFastLoadConst:
		localIdx := arg & 0xFF
		constIdx := (arg >> 8) & 0xFF
		vm.push(frame.Locals[localIdx])
		vm.push(vm.toValue(frame.Code.Constants[constIdx]))

	case OpStoreFastLoadFast:
		storeIdx := arg & 0xFF
		loadIdx := (arg >> 8) & 0xFF
		frame.Locals[storeIdx] = vm.pop()
		vm.push(frame.Locals[loadIdx])

	case OpLoadConstLoadFast:
		constIdx := (arg >> 8) & 0xFF
		localIdx := arg & 0xFF
		vm.push(vm.toValue(frame.Code.Constants[constIdx]))
		vm.push(frame.Locals[localIdx])

	case OpLoadGlobalLoadFast:
		globalIdx := (arg >> 8) & 0xFF
		localIdx := arg & 0xFF
		name := frame.Code.Names[globalIdx]
		if val, ok := frame.Globals[name]; ok {
			vm.push(val)
		} else if frame.EnclosingGlobals != nil {
			if val, ok := frame.EnclosingGlobals[name]; ok {
				vm.push(val)
			} else if val, ok := frame.Builtins[name]; ok {
				vm.push(val)
			} else {
				return nil, fmt.Errorf("name '%s' is not defined", name)
			}
		} else if val, ok := frame.Builtins[name]; ok {
			vm.push(val)
		} else {
			return nil, fmt.Errorf("name '%s' is not defined", name)
		}
		vm.push(frame.Locals[localIdx])

	case OpBinaryAddInt:
		b := vm.pop().(*PyInt)
		a := vm.pop().(*PyInt)
		vm.push(MakeInt(a.Value + b.Value))

	case OpBinarySubtractInt:
		b := vm.pop().(*PyInt)
		a := vm.pop().(*PyInt)
		vm.push(MakeInt(a.Value - b.Value))

	case OpBinaryMultiplyInt:
		b := vm.pop().(*PyInt)
		a := vm.pop().(*PyInt)
		vm.push(MakeInt(a.Value * b.Value))

	case OpCompareLtInt:
		b := vm.pop().(*PyInt)
		a := vm.pop().(*PyInt)
		if a.Value < b.Value {
			vm.push(True)
		} else {
			vm.push(False)
		}

	case OpCompareLeInt:
		b := vm.pop().(*PyInt)
		a := vm.pop().(*PyInt)
		if a.Value <= b.Value {
			vm.push(True)
		} else {
			vm.push(False)
		}

	case OpCompareGtInt:
		b := vm.pop().(*PyInt)
		a := vm.pop().(*PyInt)
		if a.Value > b.Value {
			vm.push(True)
		} else {
			vm.push(False)
		}

	case OpCompareGeInt:
		b := vm.pop().(*PyInt)
		a := vm.pop().(*PyInt)
		if a.Value >= b.Value {
			vm.push(True)
		} else {
			vm.push(False)
		}

	case OpCompareEqInt:
		b := vm.pop().(*PyInt)
		a := vm.pop().(*PyInt)
		if a.Value == b.Value {
			vm.push(True)
		} else {
			vm.push(False)
		}

	case OpCompareNeInt:
		b := vm.pop().(*PyInt)
		a := vm.pop().(*PyInt)
		if a.Value != b.Value {
			vm.push(True)
		} else {
			vm.push(False)
		}

	// Compare and jump opcodes
	case OpCompareLtJump:
		b := vm.pop()
		a := vm.pop()
		result := false
		if ai, ok := a.(*PyInt); ok {
			if bi, ok := b.(*PyInt); ok {
				result = ai.Value < bi.Value
			} else {
				result = vm.truthy(vm.compareOp(OpCompareLt, a, b))
			}
		} else {
			result = vm.truthy(vm.compareOp(OpCompareLt, a, b))
		}
		if !result {
			frame.IP = arg
		}

	case OpCompareLeJump:
		b := vm.pop()
		a := vm.pop()
		result := false
		if ai, ok := a.(*PyInt); ok {
			if bi, ok := b.(*PyInt); ok {
				result = ai.Value <= bi.Value
			} else {
				result = vm.truthy(vm.compareOp(OpCompareLe, a, b))
			}
		} else {
			result = vm.truthy(vm.compareOp(OpCompareLe, a, b))
		}
		if !result {
			frame.IP = arg
		}

	case OpCompareGtJump:
		b := vm.pop()
		a := vm.pop()
		result := false
		if ai, ok := a.(*PyInt); ok {
			if bi, ok := b.(*PyInt); ok {
				result = ai.Value > bi.Value
			} else {
				result = vm.truthy(vm.compareOp(OpCompareGt, a, b))
			}
		} else {
			result = vm.truthy(vm.compareOp(OpCompareGt, a, b))
		}
		if !result {
			frame.IP = arg
		}

	case OpCompareGeJump:
		b := vm.pop()
		a := vm.pop()
		result := false
		if ai, ok := a.(*PyInt); ok {
			if bi, ok := b.(*PyInt); ok {
				result = ai.Value >= bi.Value
			} else {
				result = vm.truthy(vm.compareOp(OpCompareGe, a, b))
			}
		} else {
			result = vm.truthy(vm.compareOp(OpCompareGe, a, b))
		}
		if !result {
			frame.IP = arg
		}

	case OpCompareEqJump:
		b := vm.pop()
		a := vm.pop()
		result := vm.equal(a, b)
		if !result {
			frame.IP = arg
		}

	case OpCompareNeJump:
		b := vm.pop()
		a := vm.pop()
		result := !vm.equal(a, b)
		if !result {
			frame.IP = arg
		}

	// Exception handling opcodes
	case OpSetupExcept:
		block := Block{
			Type:    BlockExcept,
			Handler: arg,
			Level:   frame.SP,
		}
		frame.BlockStack = append(frame.BlockStack, block)

	case OpSetupFinally:
		block := Block{
			Type:    BlockFinally,
			Handler: arg,
			Level:   frame.SP,
		}
		frame.BlockStack = append(frame.BlockStack, block)

	case OpPopExcept:
		if len(frame.BlockStack) > 0 {
			frame.BlockStack = frame.BlockStack[:len(frame.BlockStack)-1]
		}
		vm.currentException = nil

	case OpClearException:
		vm.currentException = nil

	case OpEndFinally:
		if vm.currentException != nil {
			exc := vm.currentException
			vm.currentException = nil
			_, err := vm.handleException(exc)
			if err != nil {
				return nil, err
			}
		}

	case OpExceptionMatch:
		// Check if exception matches type for except clause
		// Stack: [..., exception, type] -> [..., exception, bool]
		excType := vm.pop()
		exc := vm.top() // Peek, don't pop
		if pyExc, ok := exc.(*PyException); ok {
			if vm.exceptionMatches(pyExc, excType) {
				vm.push(True)
			} else {
				vm.push(False)
			}
		} else {
			vm.push(False)
		}

	case OpRaiseVarargs:
		var exc *PyException
		if arg == 0 {
			// Bare raise - re-raise current exception
			if vm.lastException != nil {
				exc = vm.lastException
			} else {
				return nil, fmt.Errorf("RuntimeError: No active exception to re-raise")
			}
		} else if arg == 1 {
			// raise exc
			excVal := vm.pop()
			exc = vm.createException(excVal, nil)
		} else {
			// raise exc from cause (ignore cause for now)
			cause := vm.pop()
			excVal := vm.pop()
			exc = vm.createException(excVal, cause)
		}
		exc.Traceback = vm.buildTraceback()
		_, err := vm.handleException(exc)
		if err != nil {
			return nil, err
		}
		// Check if handler is in current frame or outer frame
		if vm.frame != frame {
			return nil, errExceptionHandledInOuterFrame
		}

	default:
		return nil, fmt.Errorf("unimplemented opcode in generator: %s", op)
	}

	return nil, nil
}

// callClassBody executes a class body function with a fresh namespace
// and returns the namespace dict (not the function's return value)
func (vm *VM) callClassBody(fn *PyFunction) (map[string]Value, error) {
	code := fn.Code

	// Create a fresh namespace for the class body
	classNamespace := make(map[string]Value)

	// Create new frame with the class namespace as its globals
	// EnclosingGlobals allows the class body to access outer scope variables
	frame := &Frame{
		Code:             code,
		IP:               0,
		Stack:            make([]Value, code.StackSize+16), // Pre-allocate
		SP:               0,
		Locals:           make([]Value, len(code.VarNames)),
		Globals:          classNamespace,
		EnclosingGlobals: vm.Globals,
		Builtins:         vm.builtins,
	}

	// Push frame
	vm.frames = append(vm.frames, frame)
	oldFrame := vm.frame
	vm.frame = frame

	// Execute the class body
	_, err := vm.run()

	// Pop frame
	vm.frame = oldFrame

	if err != nil {
		return nil, err
	}

	return classNamespace, nil
}

// Run executes Python source code
func (vm *VM) Run(code *CodeObject) (Value, error) {
	return vm.Execute(code)
}
