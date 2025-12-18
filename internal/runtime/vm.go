package runtime

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// Value represents a Python value
type Value interface{}

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

// PyDict represents a Python dictionary
type PyDict struct {
	Items map[Value]Value
}

func (d *PyDict) Type() string { return "dict" }
func (d *PyDict) String() string {
	return fmt.Sprintf("%v", d.Items)
}

// PySet represents a Python set
type PySet struct {
	Items map[Value]struct{}
}

func (s *PySet) Type() string { return "set" }
func (s *PySet) String() string {
	return fmt.Sprintf("%v", s.Items)
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

// NewVM creates a new virtual machine
func NewVM() *VM {
	vm := &VM{
		Globals:       make(map[string]Value),
		builtins:      make(map[string]Value),
		checkInterval: 1000, // Check context every 1000 instructions by default
		checkCounter:  1000, // Initialize counter
	}
	vm.initBuiltins()
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
			return MakeInt(vm.toInt(args[0])), nil
		},
	}

	vm.builtins["float"] = &PyBuiltinFunc{
		Name: "float",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) == 0 {
				return &PyFloat{Value: 0.0}, nil
			}
			return &PyFloat{Value: vm.toFloat(args[0])}, nil
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
			d := &PyDict{Items: make(map[Value]Value)}
			for k, v := range kwargs {
				d.Items[&PyString{Value: k}] = v
			}
			return d, nil
		},
	}

	vm.builtins["set"] = &PyBuiltinFunc{
		Name: "set",
		Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			s := &PySet{Items: make(map[Value]struct{})}
			if len(args) > 0 {
				items, err := vm.toList(args[0])
				if err != nil {
					return nil, err
				}
				for _, item := range items {
					// Check for value equality before adding
					found := false
					for k := range s.Items {
						if vm.equal(k, item) {
							found = true
							break
						}
					}
					if !found {
						s.Items[item] = struct{}{}
					}
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

			// Build MRO (Method Resolution Order) - simple linearization
			class.Mro = []*PyClass{class}
			for _, base := range bases {
				for _, mroClass := range base.Mro {
					// Add if not already in MRO
					found := false
					for _, existing := range class.Mro {
						if existing == mroClass {
							found = true
							break
						}
					}
					if !found {
						class.Mro = append(class.Mro, mroClass)
					}
				}
			}

			return class, nil
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
			// Inline push for constant load
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

		case OpLoadFastLoadFast:
			// Load two locals: arg contains packed indices (low byte = first, high byte = second)
			idx1 := arg & 0xFF
			idx2 := (arg >> 8) & 0xFF
			frame.Stack[frame.SP] = frame.Locals[idx1]
			frame.SP++
			frame.Stack[frame.SP] = frame.Locals[idx2]
			frame.SP++

		case OpLoadFastLoadConst:
			// Load local then const: arg contains packed indices
			localIdx := arg & 0xFF
			constIdx := (arg >> 8) & 0xFF
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
			iter := vm.top().(*PyIterator)
			if iter.Index < len(iter.Items) {
				vm.push(iter.Items[iter.Index])
				iter.Index++
			} else {
				vm.pop() // Pop iterator
				frame.IP = arg
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
			s := &PySet{Items: make(map[Value]struct{})}
			for i := 0; i < arg; i++ {
				val := vm.pop()
				// Check for value equality before adding
				found := false
				for k := range s.Items {
					if vm.equal(k, val) {
						found = true
						break
					}
				}
				if !found {
					s.Items[val] = struct{}{}
				}
			}
			vm.push(s)

		case OpBuildMap:
			d := &PyDict{Items: make(map[Value]Value)}
			for i := 0; i < arg; i++ {
				val := vm.pop()
				key := vm.pop()
				d.Items[key] = val
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
			// Check if an equivalent value already exists (for proper value equality)
			found := false
			for k := range set.Items {
				if vm.equal(k, val) {
					found = true
					break
				}
			}
			if !found {
				set.Items[val] = struct{}{}
			}

		case OpMapAdd:
			val := vm.pop()
			key := vm.pop()
			dict := vm.peek(arg).(*PyDict)
			dict.Items[key] = val

		case OpMakeFunction:
			name := vm.pop().(*PyString).Value
			code := vm.pop().(*CodeObject)
			var defaults *PyTuple
			if arg&1 != 0 {
				defaults = vm.pop().(*PyTuple)
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

		case OpRaiseVarargs:
			if arg == 0 {
				return nil, fmt.Errorf("no active exception to re-raise")
			} else if arg == 1 {
				exc := vm.pop()
				return nil, fmt.Errorf("%v", exc)
			} else {
				cause := vm.pop()
				exc := vm.pop()
				return nil, fmt.Errorf("%v (caused by %v)", exc, cause)
			}

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
	f.Stack[f.SP] = v
	f.SP++
}

func (vm *VM) pop() Value {
	vm.frame.SP--
	return vm.frame.Stack[vm.frame.SP]
}

func (vm *VM) top() Value {
	return vm.frame.Stack[vm.frame.SP-1]
}

func (vm *VM) peek(n int) Value {
	return vm.frame.Stack[vm.frame.SP-1-n]
}

// Type conversions

func (vm *VM) toValue(v interface{}) Value {
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
	switch val := v.(type) {
	case *PyInt:
		return val.Value
	case *PyFloat:
		return int64(val.Value)
	case *PyBool:
		if val.Value {
			return 1
		}
		return 0
	case *PyString:
		var i int64
		fmt.Sscanf(val.Value, "%d", &i)
		return i
	default:
		return 0
	}
}

func (vm *VM) toFloat(v Value) float64 {
	switch val := v.(type) {
	case *PyInt:
		return float64(val.Value)
	case *PyFloat:
		return val.Value
	case *PyBool:
		if val.Value {
			return 1.0
		}
		return 0.0
	case *PyString:
		var f float64
		fmt.Sscanf(val.Value, "%f", &f)
		return f
	default:
		return 0.0
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
				return MakeInt(ai.Value / bi.Value), nil
			case OpBinaryModulo:
				if bi.Value == 0 {
					return nil, fmt.Errorf("integer modulo by zero")
				}
				return MakeInt(ai.Value % bi.Value), nil
			case OpBinaryPower:
				return MakeInt(int64(math.Pow(float64(ai.Value), float64(bi.Value)))), nil
			case OpBinaryLShift:
				return MakeInt(ai.Value << bi.Value), nil
			case OpBinaryRShift:
				return MakeInt(ai.Value >> bi.Value), nil
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
	if op == OpBinaryMultiply {
		if as, ok := a.(*PyString); ok {
			if bi, ok := b.(*PyInt); ok {
				if bi.Value <= 0 {
					return &PyString{Value: ""}, nil
				}
				return &PyString{Value: strings.Repeat(as.Value, int(bi.Value))}, nil
			}
		}
		if as, ok := b.(*PyString); ok {
			if ai, ok := a.(*PyInt); ok {
				if ai.Value <= 0 {
					return &PyString{Value: ""}, nil
				}
				return &PyString{Value: strings.Repeat(as.Value, int(ai.Value))}, nil
			}
		}
		// List repetition - pre-allocate for efficiency
		if al, ok := a.(*PyList); ok {
			if bi, ok := b.(*PyInt); ok {
				if bi.Value <= 0 {
					return &PyList{Items: []Value{}}, nil
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
			for i := range av.Items {
				if !vm.equal(av.Items[i], bv.Items[i]) {
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
				if !vm.equal(av.Items[i], bv.Items[i]) {
					return false
				}
			}
			return true
		}
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
			for j := 0; j <= len(c.Value)-len(i.Value); j++ {
				if c.Value[j:j+len(i.Value)] == i.Value {
					return true
				}
			}
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
		for k := range c.Items {
			if vm.equal(k, item) {
				return true
			}
		}
	case *PyDict:
		for k := range c.Items {
			if vm.equal(k, item) {
				return true
			}
		}
	}
	return false
}

// Attribute access

func (vm *VM) getAttr(obj Value, name string) (Value, error) {
	switch o := obj.(type) {
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
	case *PyInstance:
		// First check instance dict
		if val, ok := o.Dict[name]; ok {
			return val, nil
		}
		// Then check class MRO for methods/attributes
		for _, cls := range o.Class.Mro {
			if val, ok := cls.Dict[name]; ok {
				// Bind method if it's a function
				if fn, ok := val.(*PyFunction); ok {
					return &PyMethod{Func: fn, Instance: obj}, nil
				}
				return val, nil
			}
		}
		return nil, fmt.Errorf("'%s' object has no attribute '%s'", o.Class.Name, name)
	case *PyClass:
		// Check class dict and MRO
		for _, cls := range o.Mro {
			if val, ok := cls.Dict[name]; ok {
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
				for i := 0; i < len(o.Value); i++ {
					if i+len(sep) <= len(o.Value) && o.Value[i:i+len(sep)] == sep {
						parts = append(parts, &PyString{Value: current})
						current = ""
						i += len(sep) - 1
					} else {
						current += string(o.Value[i])
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
		idx := int(vm.toInt(index))
		if idx < 0 {
			idx = len(o.Value) + idx
		}
		if idx < 0 || idx >= len(o.Value) {
			return nil, fmt.Errorf("string index out of range")
		}
		return &PyString{Value: string(o.Value[idx])}, nil
	case *PyDict:
		if val, ok := o.Items[index]; ok {
			return val, nil
		}
		// Try finding equivalent key
		for k, v := range o.Items {
			if vm.equal(k, index) {
				return v, nil
			}
		}
		return nil, fmt.Errorf("KeyError: %v", index)
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
		o.Items[index] = val
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
		delete(o.Items, index)
		return nil
	}
	return fmt.Errorf("'%s' object does not support item deletion", vm.typeName(obj))
}

// Iterator

func (vm *VM) getIter(obj Value) (Value, error) {
	items, err := vm.toList(obj)
	if err != nil {
		return nil, err
	}
	return &PyIterator{Items: items, Index: 0}, nil
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
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Convert panic to error - push error onto stack
				tempFrame.Stack[tempFrame.SP] = NewString(fmt.Sprintf("%v", r))
				tempFrame.SP++
				nResults = -1 // Indicate error
			}
		}()
		nResults = fn.Fn(vm)
	}()

	// Restore frame
	vm.frame = oldFrame

	// Handle error case
	if nResults < 0 {
		errVal := tempFrame.Stack[tempFrame.SP-1]
		return nil, fmt.Errorf("%s", vm.str(errVal))
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

	// Bind arguments to locals
	for i, arg := range args {
		if i < len(frame.Locals) {
			frame.Locals[i] = arg
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

	// Push frame
	vm.frames = append(vm.frames, frame)
	oldFrame := vm.frame
	vm.frame = frame

	// Execute
	result, err := vm.run()

	// Pop frame
	vm.frame = oldFrame

	return result, err
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
