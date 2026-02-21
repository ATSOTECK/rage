package runtime

import (
	"context"
	"fmt"
	"time"
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
	currentException *PyException   // Currently active exception being handled
	lastException    *PyException   // Last raised exception (for bare raise)
	excHandlerStack  []*PyException // Stack of exceptions being handled (for __context__)

	// Generator exception injection
	generatorThrow *PyException // Exception to throw into generator on resume

	// Generator return/continue-through-finally state
	generatorPendingReturn    Value // Return value pending while finally block executes
	generatorHasPendingReturn bool  // Whether a return is pending
	generatorPendingJump      int   // Jump target pending while finally block executes (continue/break)
	generatorHasPendingJump   bool  // Whether a jump is pending

	// except* state stack
	exceptStarStack []ExceptStarState

	// Resource limits (0 = unlimited)
	maxRecursionDepth int64
	maxMemoryBytes    int64
	allocatedBytes    int64
	maxCollectionSize int64

	// Filesystem module imports
	SearchPaths  []string                              // Directories to search for .py modules
	FileImporter func(filename string) (*CodeObject, error) // Callback to compile a .py file (avoids circular dep)
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

// SetMaxRecursionDepth sets the maximum call stack depth. 0 means unlimited.
func (vm *VM) SetMaxRecursionDepth(n int64) {
	if n < 0 {
		n = 0
	}
	vm.maxRecursionDepth = n
}

// MaxRecursionDepth returns the current maximum recursion depth. 0 means unlimited.
func (vm *VM) MaxRecursionDepth() int64 {
	return vm.maxRecursionDepth
}

// SetMaxMemoryBytes sets the approximate memory limit in bytes. 0 means unlimited.
func (vm *VM) SetMaxMemoryBytes(n int64) {
	if n < 0 {
		n = 0
	}
	vm.maxMemoryBytes = n
}

// MaxMemoryBytes returns the current memory limit. 0 means unlimited.
func (vm *VM) MaxMemoryBytes() int64 {
	return vm.maxMemoryBytes
}

// AllocatedBytes returns the approximate number of bytes currently tracked.
func (vm *VM) AllocatedBytes() int64 {
	return vm.allocatedBytes
}

// SetMaxCollectionSize sets the maximum number of elements in a single collection.
// 0 means unlimited.
func (vm *VM) SetMaxCollectionSize(n int64) {
	if n < 0 {
		n = 0
	}
	vm.maxCollectionSize = n
}

// MaxCollectionSize returns the current collection size limit. 0 means unlimited.
func (vm *VM) MaxCollectionSize() int64 {
	return vm.maxCollectionSize
}

// trackAlloc adds n bytes to the allocation counter and returns a MemoryError if over limit.
func (vm *VM) trackAlloc(n int64) error {
	vm.allocatedBytes += n
	if vm.maxMemoryBytes > 0 && vm.allocatedBytes > vm.maxMemoryBytes {
		return fmt.Errorf("MemoryError: memory limit exceeded (%d bytes allocated, limit is %d)", vm.allocatedBytes, vm.maxMemoryBytes)
	}
	return nil
}

// TrackAlloc is the exported wrapper of trackAlloc for stdlib use.
func (vm *VM) TrackAlloc(n int64) error {
	return vm.trackAlloc(n)
}

// checkCollectionSize returns a MemoryError if current >= maxCollectionSize (when set).
func (vm *VM) checkCollectionSize(current int64, collType string) error {
	if vm.maxCollectionSize > 0 && current >= vm.maxCollectionSize {
		return fmt.Errorf("MemoryError: %s size limit exceeded (limit is %d)", collType, vm.maxCollectionSize)
	}
	return nil
}

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

	// Track top-level frame memory
	if vm.maxMemoryBytes > 0 {
		frameBytes := int64(len(frame.Stack))*16 + int64(len(frame.Locals))*16
		if err := vm.trackAlloc(frameBytes); err != nil {
			return nil, err
		}
	}

	vm.frames = append(vm.frames, frame)
	vm.frame = frame
	vm.ctx = ctx
	vm.checkCounter = vm.checkInterval // Reset counter for new execution

	return vm.run()
}

// ExecuteInModule runs bytecode with a module's dict as the global namespace.
// This is used to populate a module's namespace when registering Python modules.
func (vm *VM) ExecuteInModule(code *CodeObject, mod *PyModule) error {
	frame := &Frame{
		Code:     code,
		IP:       0,
		Stack:    make([]Value, code.StackSize+16),
		SP:       0,
		Locals:   make([]Value, len(code.VarNames)),
		Globals:  mod.Dict,
		Builtins: vm.builtins,
	}

	// Track module frame memory
	if vm.maxMemoryBytes > 0 {
		frameBytes := int64(len(frame.Stack))*16 + int64(len(frame.Locals))*16
		if err := vm.trackAlloc(frameBytes); err != nil {
			return err
		}
	}

	vm.frames = append(vm.frames, frame)
	vm.frame = frame
	vm.ctx = context.Background()
	vm.checkCounter = vm.checkInterval

	_, err := vm.run()
	return err
}

// tryHandleError attempts to handle a Go error as a Python exception.
// Returns (true, nil) if handler found in current frame (caller should continue).
// Returns (false, errExceptionHandledInOuterFrame) if handled in outer frame.
// Returns (false, err) if no handler found (caller should return nil, err).
func (vm *VM) tryHandleError(err error, frame *Frame) (bool, error) {
	if err == errExceptionHandledInOuterFrame {
		if vm.frame == frame {
			return true, nil
		}
		return false, errExceptionHandledInOuterFrame
	}
	var pyExc *PyException
	if pe, ok := err.(*PyException); ok {
		pyExc = pe
	} else {
		pyExc = vm.wrapGoError(err)
	}
	// Set implicit __context__ from the exception handler stack
	if len(vm.excHandlerStack) > 0 {
		handledException := vm.excHandlerStack[len(vm.excHandlerStack)-1]
		if pyExc != handledException {
			pyExc.Context = handledException
		}
	}
	_, handleErr := vm.handleException(pyExc)
	if handleErr != nil {
		return false, handleErr
	}
	if vm.frame != frame {
		return false, errExceptionHandledInOuterFrame
	}
	return true, nil
}

// Stack operations - using stack pointer with pre-allocated slice

func (vm *VM) push(v Value) {
	f := vm.frame
	// Grow stack if needed
	if f.SP >= len(f.Stack) {
		oldSize := len(f.Stack)
		newStack := make([]Value, oldSize*2)
		copy(newStack, f.Stack)
		f.Stack = newStack
		// Best-effort memory tracking (no error return)
		vm.trackAlloc(int64(oldSize) * 16)
	}
	f.Stack[f.SP] = v
	f.SP++
}

// ensureStack ensures the stack has at least n additional slots available
func (vm *VM) ensureStack(n int) {
	f := vm.frame
	needed := f.SP + n
	if needed > len(f.Stack) {
		oldSize := len(f.Stack)
		newSize := oldSize * 2
		if newSize < needed {
			newSize = needed + 16
		}
		newStack := make([]Value, newSize)
		copy(newStack, f.Stack)
		f.Stack = newStack
		// Best-effort memory tracking
		vm.trackAlloc(int64(newSize-oldSize) * 16)
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

// unboundLocalError returns an UnboundLocalError for the given local variable index.
func unboundLocalError(frame *Frame, index int) error {
	varName := ""
	if index < len(frame.Code.VarNames) {
		varName = frame.Code.VarNames[index]
	}
	return fmt.Errorf("UnboundLocalError: cannot access local variable '%s' referenced before assignment", varName)
}

// checkCurrentException handles the current exception if present.
// Returns (handled, result, err):
//   - handled=false: no exception was pending, continue normally
//   - handled=true, err!=nil: exception not caught, caller should return nil, err
//   - handled=true, result!=nil: exception caught in outer frame, caller should return result, nil
//   - handled=true, result==nil, err==nil: exception caught in current frame, caller should break
func (vm *VM) checkCurrentException() (bool, Value, error) {
	if vm.currentException == nil {
		return false, nil, nil
	}
	exc := vm.currentException
	vm.currentException = nil
	result, err := vm.handleException(exc)
	if err != nil {
		return true, nil, err
	}
	return true, result, nil
}

// Run executes Python source code
func (vm *VM) Run(code *CodeObject) (Value, error) {
	return vm.Execute(code)
}
