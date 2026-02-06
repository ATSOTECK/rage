package runtime

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"
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

	vm.frames = append(vm.frames, frame)
	vm.frame = frame
	vm.ctx = context.Background()
	vm.checkCounter = vm.checkInterval

	_, err := vm.run()
	return err
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

		case OpDup2:
			// Duplicate top two stack items: [a, b] -> [a, b, a, b]
			b := vm.top()
			a := vm.peek(1)
			vm.push(a)
			vm.push(b)

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
				length = int64(utf8.RuneCountInString(v.Value))
			case *PyList:
				length = int64(len(v.Items))
			case *PyTuple:
				length = int64(len(v.Items))
			case *PyDict:
				length = int64(len(v.Items))
			case *PySet:
				length = int64(len(v.Items))
			case *PyFrozenSet:
				length = int64(len(v.Items))
			case *PyBytes:
				length = int64(len(v.Value))
			case *PyInstance:
				// Check for __len__ method
				if result, found, err := vm.callDunder(v, "__len__"); found {
					if err != nil {
						return nil, err
					}
					if i, ok := result.(*PyInt); ok {
						length = i.Value
					} else {
						return nil, fmt.Errorf("__len__() should return an integer")
					}
				} else {
					return nil, fmt.Errorf("object of type '%s' has no len()", vm.typeName(obj))
				}
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
			// Check for __pos__ on instances
			if inst, ok := a.(*PyInstance); ok {
				if result, found, err := vm.callDunder(inst, "__pos__"); found {
					if err != nil {
						return nil, err
					}
					vm.push(result)
					break
				}
			}
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
				length = int64(utf8.RuneCountInString(s.Value))
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
			levelVal := vm.pop() // level (for relative imports)

			// Get level as int
			level := 0
			if levelInt, ok := levelVal.(*PyInt); ok {
				level = int(levelInt.Value)
			}

			// Resolve relative imports
			moduleName := name
			if level > 0 {
				// Get __package__ from globals for relative import resolution
				packageName := ""
				if pkgVal, ok := frame.Globals["__package__"]; ok {
					if pkgStr, ok := pkgVal.(*PyString); ok {
						packageName = pkgStr.Value
					}
				}
				// If __package__ is not set, try to derive from __name__
				if packageName == "" {
					if nameVal, ok := frame.Globals["__name__"]; ok {
						if nameStr, ok := nameVal.(*PyString); ok {
							// For a module like "pkg.sub.module", the package is "pkg.sub"
							// For a package like "pkg.sub", the package is "pkg.sub" itself
							packageName = nameStr.Value
						}
					}
				}

				resolved, err := ResolveRelativeImport(name, level, packageName)
				if err != nil {
					return nil, err
				}
				moduleName = resolved
			}

			// Handle dotted imports (e.g., "import outer.inner")
			// Need to import each part of the path and return the appropriate module
			var rootMod, targetMod *PyModule
			parts := splitModuleName(moduleName)

			for i := range parts {
				partialName := joinModuleName(parts[:i+1])
				mod, err := vm.ImportModule(partialName)
				if err != nil {
					return nil, err
				}
				if i == 0 {
					rootMod = mod
				}
				targetMod = mod
			}

			// Determine which module to push:
			// - For "import X.Y.Z": push X (the root), stored as "X" in namespace
			// - For "from X.Y import Z": push X.Y (the target), for IMPORT_FROM to use
			hasFromlist := false
			if fromlist != nil && fromlist != None {
				if list, ok := fromlist.(*PyList); ok && len(list.Items) > 0 {
					hasFromlist = true
				} else if tuple, ok := fromlist.(*PyTuple); ok && len(tuple.Items) > 0 {
					hasFromlist = true
				} else if strList, ok := fromlist.([]string); ok && len(strList) > 0 {
					hasFromlist = true
				}
			}

			if hasFromlist {
				vm.push(targetMod)
			} else {
				vm.push(rootMod)
			}

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

// Run executes Python source code
func (vm *VM) Run(code *CodeObject) (Value, error) {
	return vm.Execute(code)
}
