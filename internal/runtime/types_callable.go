package runtime

import (
	"fmt"
	"strings"
)

// PyFunction represents a Python function
type PyFunction struct {
	Code       *CodeObject
	Globals    map[string]Value
	Defaults   *PyTuple
	KwDefaults map[string]Value // Keyword-only argument defaults
	Closure    []*PyCell
	Name       string
	IsAbstract bool             // Set by @abstractmethod decorator
	Dict       map[string]Value // Custom attributes (e.g. func._name)
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
	Name  string
	Fn    func(args []Value, kwargs map[string]Value) (Value, error)
	Bound bool // true when self is already captured (bound method wrapper)
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
	ExcType        *PyClass         // Exception class (e.g., ValueError, TypeError)
	TypeName       string           // Exception type name (used when ExcType is nil)
	Args           *PyTuple         // Exception arguments
	Message        string           // String representation
	Cause          *PyException     // __cause__ for chained exceptions (raise X from Y)
	Context        *PyException     // __context__ for implicit chaining
	SuppressContext bool            // __suppress_context__ - set True when __cause__ is assigned
	Traceback      []TracebackEntry // Traceback frames
	Instance       *PyInstance      // non-nil for ExceptionGroup instances
	Notes          *PyList          // __notes__ list (nil until add_note is called)
}

// ExceptStarState tracks the remaining unmatched exceptions during except* handling
type ExceptStarState struct {
	Remaining []*PyException
	Message   string
	IsBase    bool
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
	OrderedGlobalKeys []string        // Insertion-ordered global names (for class bodies)
}

// Block represents a control flow block
type Block struct {
	Type          BlockType
	Handler       int // Handler address
	Level         int // Stack level
	ExcStackLevel int // excHandlerStack level at block setup time
}

// BlockType identifies the type of block
type BlockType int

const (
	BlockLoop BlockType = iota
	BlockExcept
	BlockFinally
	BlockWith
	BlockExceptStar
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
	Items  []Value
	Index  int
	Source *PyList // Optional: if set, Items is dynamically read from Source.Items (for live mutation visibility)
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

	// Saved VM exception state (isolated per-generator)
	SavedCurrentException    *PyException
	SavedLastException       *PyException
	SavedExcHandlerStack     []*PyException
	SavedPendingReturn       Value
	SavedHasPendingReturn    bool
	SavedPendingJump         int
	SavedHasPendingJump      bool
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
