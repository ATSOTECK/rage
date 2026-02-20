package runtime

import (
	"context"
	"fmt"
	"unicode/utf8"
)

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

	// Save current frame and VM state, switch to generator's frame
	oldFrame := vm.frame
	oldFrames := vm.frames
	oldCurrentException := vm.currentException
	oldLastException := vm.lastException
	oldExcHandlerStack := vm.excHandlerStack
	oldPendingReturn := vm.generatorPendingReturn
	oldHasPendingReturn := vm.generatorHasPendingReturn
	oldPendingJump := vm.generatorPendingJump
	oldHasPendingJump := vm.generatorHasPendingJump

	vm.frame = gen.Frame
	vm.frames = []*Frame{gen.Frame}
	vm.currentException = gen.SavedCurrentException
	vm.lastException = gen.SavedLastException
	vm.excHandlerStack = gen.SavedExcHandlerStack
	vm.generatorPendingReturn = gen.SavedPendingReturn
	vm.generatorHasPendingReturn = gen.SavedHasPendingReturn
	vm.generatorPendingJump = gen.SavedPendingJump
	vm.generatorHasPendingJump = gen.SavedHasPendingJump

	// If resuming from yield, push the sent value onto the stack
	if gen.Frame.IP > 0 {
		vm.push(value)
	}

	// Run until yield or return
	result, yielded, err := vm.runGenerator()

	// Save generator's state
	gen.SavedCurrentException = vm.currentException
	gen.SavedLastException = vm.lastException
	gen.SavedExcHandlerStack = vm.excHandlerStack
	gen.SavedPendingReturn = vm.generatorPendingReturn
	gen.SavedHasPendingReturn = vm.generatorHasPendingReturn
	gen.SavedPendingJump = vm.generatorPendingJump
	gen.SavedHasPendingJump = vm.generatorHasPendingJump

	// Restore caller's frame and VM state
	vm.frame = oldFrame
	vm.frames = oldFrames
	vm.currentException = oldCurrentException
	vm.lastException = oldLastException
	vm.excHandlerStack = oldExcHandlerStack
	vm.generatorPendingReturn = oldPendingReturn
	vm.generatorHasPendingReturn = oldHasPendingReturn
	vm.generatorPendingJump = oldPendingJump
	vm.generatorHasPendingJump = oldHasPendingJump

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
		if pyExc, ok := excType.(*PyException); ok {
			return nil, true, pyExc
		}
		if inst, ok := excType.(*PyInstance); ok && vm.isExceptionClass(inst.Class) {
			return nil, true, vm.createException(inst, nil)
		}
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
		if pyExc, ok := excType.(*PyException); ok {
			return nil, true, pyExc
		}
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
	var exc *PyException

	// Check if excType is already a PyException
	if pyExc, ok := excType.(*PyException); ok {
		exc = pyExc
	} else if inst, ok := excType.(*PyInstance); ok && vm.isExceptionClass(inst.Class) {
		// Already-instantiated exception instance (e.g., ValueError("xyz"))
		exc = vm.createException(inst, nil)
	} else {
		excMsg := "exception"
		if str, ok := excValue.(*PyString); ok {
			excMsg = str.Value
		} else if excValue != nil && excValue != None {
			excMsg = fmt.Sprintf("%v", excValue)
		}

		exc = &PyException{
			TypeName: fmt.Sprintf("%v", excType),
			Message:  excMsg,
		}

		// Check if excType is a class and set ExcType appropriately
		if cls, ok := excType.(*PyClass); ok {
			exc.ExcType = cls
		}
	}

	// Set the pending exception on the VM for the generator to handle
	vm.generatorThrow = exc
	gen.State = GenRunning

	// Save current frame and VM state, switch to generator's frame
	oldFrame := vm.frame
	oldFrames := vm.frames
	oldCurrentException := vm.currentException
	oldLastException := vm.lastException
	oldExcHandlerStack := vm.excHandlerStack
	oldPendingReturn := vm.generatorPendingReturn
	oldHasPendingReturn := vm.generatorHasPendingReturn
	oldPendingJump := vm.generatorPendingJump
	oldHasPendingJump := vm.generatorHasPendingJump

	vm.frame = gen.Frame
	vm.frames = []*Frame{gen.Frame}
	vm.currentException = gen.SavedCurrentException
	vm.lastException = gen.SavedLastException
	vm.excHandlerStack = gen.SavedExcHandlerStack
	vm.generatorPendingReturn = gen.SavedPendingReturn
	vm.generatorHasPendingReturn = gen.SavedHasPendingReturn
	vm.generatorPendingJump = gen.SavedPendingJump
	vm.generatorHasPendingJump = gen.SavedHasPendingJump

	// Run until yield or return (exception will be handled in runWithYieldSupport)
	result, yielded, err := vm.runGenerator()

	// Save generator's state
	gen.SavedCurrentException = vm.currentException
	gen.SavedLastException = vm.lastException
	gen.SavedExcHandlerStack = vm.excHandlerStack
	gen.SavedPendingReturn = vm.generatorPendingReturn
	gen.SavedHasPendingReturn = vm.generatorHasPendingReturn
	gen.SavedPendingJump = vm.generatorPendingJump
	gen.SavedHasPendingJump = vm.generatorHasPendingJump

	// Restore caller's frame and VM state
	vm.frame = oldFrame
	vm.frames = oldFrames
	vm.currentException = oldCurrentException
	vm.lastException = oldLastException
	vm.excHandlerStack = oldExcHandlerStack
	vm.generatorPendingReturn = oldPendingReturn
	vm.generatorHasPendingReturn = oldHasPendingReturn
	vm.generatorPendingJump = oldPendingJump
	vm.generatorHasPendingJump = oldHasPendingJump
	vm.generatorPendingReturn = oldPendingReturn
	vm.generatorHasPendingReturn = oldHasPendingReturn

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
	genExitClass, _ := vm.builtins["GeneratorExit"]
	genExitExc := &PyException{
		TypeName: "GeneratorExit",
		Message:  "generator exit",
	}
	if cls, ok := genExitClass.(*PyClass); ok {
		genExitExc.ExcType = cls
	}
	_, _, err := vm.GeneratorThrow(gen, genExitExc, None)

	// GeneratorExit is expected - if we get it back, ignore it
	if err != nil {
		if pyErr, ok := err.(*PyException); ok {
			if pyErr.Type() == "GeneratorExit" {
				gen.State = GenClosed
				return nil
			}
		}
		// StopIteration means the generator returned normally after catching
		// GeneratorExit and not re-raising - this should raise RuntimeError
		// But for now, just suppress StopIteration from close()
		if pyErr, ok := err.(*PyException); ok && pyErr.Type() == "StopIteration" {
			gen.State = GenClosed
			return nil
		}
	}

	// No error means generator yielded - this is illegal after close
	if err == nil {
		gen.State = GenClosed
		return &PyException{
			TypeName: "RuntimeError",
			Message:  "generator ignored GeneratorExit",
		}
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

// CoroutineThrow throws an exception into a coroutine
func (vm *VM) CoroutineThrow(coro *PyCoroutine, excType, excValue Value) (Value, bool, error) {
	if coro.State == GenClosed {
		// Re-raise the exception
		excMsg := "exception"
		if str, ok := excValue.(*PyString); ok {
			excMsg = str.Value
		}
		exc := &PyException{
			TypeName: fmt.Sprintf("%v", excType),
			Message:  excMsg,
		}
		if cls, ok := excType.(*PyClass); ok {
			exc.ExcType = cls
		}
		return nil, true, exc
	}

	if coro.State == GenRunning {
		return nil, false, &PyException{
			TypeName: "ValueError",
			Message:  "coroutine already executing",
		}
	}

	if coro.State == GenCreated {
		coro.State = GenClosed
		excMsg := "exception"
		if str, ok := excValue.(*PyString); ok {
			excMsg = str.Value
		}
		exc := &PyException{
			TypeName: fmt.Sprintf("%v", excType),
			Message:  excMsg,
		}
		if cls, ok := excType.(*PyClass); ok {
			exc.ExcType = cls
		}
		return nil, true, exc
	}

	// Coroutine is suspended at an await point - throw the exception there
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

	// Set the pending exception on the VM for the coroutine to handle
	vm.generatorThrow = exc
	coro.State = GenRunning

	// Save current frame and switch to coroutine's frame
	oldFrame := vm.frame
	oldFrames := vm.frames

	vm.frame = coro.Frame
	vm.frames = []*Frame{coro.Frame}

	// Run until yield or return (exception will be handled in runWithYieldSupport)
	result, yielded, err := vm.runGenerator()

	// Restore old frame
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

	// Coroutine returned (finished)
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
			// On resume after a yield, the sent value is pushed on top of the iterator.
			// We need to pop the sent value and use the iterator beneath it.
			// On first execution, only the iterator is on the stack.
			var sendVal Value = None
			iter := vm.top()

			// Check if the top of stack is the sent value (not an iterator)
			// This happens on resume: stack is [..., iterator, sent_value]
			switch iter.(type) {
			case *PyGenerator, *PyCoroutine, *PyIterator:
				// Top is already the iterator
			default:
				// Top is the sent value, pop it and get the iterator below
				sendVal = vm.pop()
				iter = vm.top()
			}

			// Try to get next value from iterator
			switch it := iter.(type) {
			case *PyGenerator:
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
				items := it.Items
				if it.Source != nil {
					items = it.Source.Items
				}
				if it.Index >= len(items) {
					vm.pop()
					vm.push(None)
					continue
				}
				val := items[it.Index]
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
			// Check for finally blocks on the block stack that need to run before returning
			foundFinally := false
			for len(frame.BlockStack) > 0 {
				block := frame.BlockStack[len(frame.BlockStack)-1]
				if block.Type == BlockFinally {
					frame.BlockStack = frame.BlockStack[:len(frame.BlockStack)-1]
					frame.SP = block.Level
					frame.IP = block.Handler
					// Push a "return sentinel" so EndFinally knows to complete the return
					vm.generatorPendingReturn = result
					vm.generatorHasPendingReturn = true
					// Push None as the "exception" for finally (no exception, just cleanup)
					vm.push(None)
					foundFinally = true
					break
				}
				frame.BlockStack = frame.BlockStack[:len(frame.BlockStack)-1]
			}
			if foundFinally {
				continue // Continue the outer opcode execution loop to run the finally block
			}
			return result, false, nil

		case OpContinueLoop:
			// Continue loop - check for finally blocks that need to run first
			foundFinally := false
			for len(frame.BlockStack) > 0 {
				block := frame.BlockStack[len(frame.BlockStack)-1]
				if block.Type == BlockFinally {
					frame.BlockStack = frame.BlockStack[:len(frame.BlockStack)-1]
					frame.SP = block.Level
					frame.IP = block.Handler
					// Set pending jump so EndFinally knows to jump to the loop target
					vm.generatorPendingJump = arg
					vm.generatorHasPendingJump = true
					foundFinally = true
					break
				}
				if block.Type == BlockLoop {
					break // Stop at the loop block
				}
				frame.BlockStack = frame.BlockStack[:len(frame.BlockStack)-1]
			}
			if foundFinally {
				continue
			}
			// No finally block, just jump directly
			frame.IP = arg

		case OpGenStart:
			// No-op, just marks generator start
			continue

		default:
			// Execute regular opcode using the standard dispatcher
			result, err := vm.executeOpcodeForGenerator(op, arg)
			if err != nil {
				// Try to handle the exception with try/except handlers in the generator
				var pyExc *PyException
				if pe, ok := err.(*PyException); ok {
					pyExc = pe
				} else {
					pyExc = vm.wrapGoError(err)
				}
				_, handleErr := vm.handleException(pyExc)
				if handleErr != nil {
					// No handler found — propagate
					return nil, false, handleErr
				}
				// Handler found — continue execution at handler address
				frame = vm.frame
				continue
			}
			if result != nil {
				// Some opcodes return values (e.g., pending return from EndFinally)
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
		val := frame.Locals[arg]
		if val == nil {
			return nil, unboundLocalError(frame, arg)
		}
		frame.Stack[frame.SP] = val
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
		vm.push(vm.compareOp(OpCompareNe, a, b))
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
	case OpDeleteAttr:
		name := frame.Code.Names[arg]
		obj := vm.pop()
		if err := vm.delAttr(obj, name); err != nil {
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
		val := frame.Locals[0]
		if val == nil {
			return nil, unboundLocalError(frame, 0)
		}
		vm.push(val)
	case OpLoadFast1:
		val := frame.Locals[1]
		if val == nil {
			return nil, unboundLocalError(frame, 1)
		}
		vm.push(val)
	case OpLoadFast2:
		val := frame.Locals[2]
		if val == nil {
			return nil, unboundLocalError(frame, 2)
		}
		vm.push(val)
	case OpLoadFast3:
		val := frame.Locals[3]
		if val == nil {
			return nil, unboundLocalError(frame, 3)
		}
		vm.push(val)
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

	case OpSetupAnnotations:
		if _, ok := frame.Globals["__annotations__"]; !ok {
			frame.Globals["__annotations__"] = &PyDict{Items: make(map[Value]Value)}
		}

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
		alreadyBound := false
		if _, isBound := method.(*PyMethod); isBound {
			alreadyBound = true
		} else if bf, ok := method.(*PyBuiltinFunc); ok && bf.Bound {
			alreadyBound = true
		}
		if alreadyBound {
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

	case OpUnpackEx:
		countBefore := arg & 0xFF
		countAfter := (arg >> 8) & 0xFF
		seq := vm.pop()
		items, err := vm.toList(seq)
		if err != nil {
			return nil, err
		}
		totalRequired := countBefore + countAfter
		if len(items) < totalRequired {
			return nil, fmt.Errorf("ValueError: not enough values to unpack (expected at least %d, got %d)", totalRequired, len(items))
		}
		for i := len(items) - 1; i >= len(items)-countAfter; i-- {
			vm.push(items[i])
		}
		starItems := make([]Value, len(items)-totalRequired)
		copy(starItems, items[countBefore:len(items)-countAfter])
		vm.push(&PyList{Items: starItems})
		for i := countBefore - 1; i >= 0; i-- {
			vm.push(items[i])
		}

	case OpCompareIn:
		container := vm.pop()
		item := vm.pop()
		if vm.contains(container, item) {
			vm.push(True)
		} else {
			if vm.currentException != nil {
				exc := vm.currentException
				vm.currentException = nil
				return nil, exc
			}
			vm.push(False)
		}

	case OpCompareNotIn:
		container := vm.pop()
		item := vm.pop()
		if !vm.contains(container, item) {
			if vm.currentException != nil {
				exc := vm.currentException
				vm.currentException = nil
				return nil, exc
			}
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
		if frame.Locals[arg] == nil {
			return nil, unboundLocalError(frame, arg)
		}
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
		if frame.Locals[arg] == nil {
			return nil, unboundLocalError(frame, arg)
		}
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
		val1 := frame.Locals[idx1]
		if val1 == nil {
			return nil, unboundLocalError(frame, idx1)
		}
		val2 := frame.Locals[idx2]
		if val2 == nil {
			return nil, unboundLocalError(frame, idx2)
		}
		vm.push(val1)
		vm.push(val2)

	case OpLoadFastLoadConst:
		localIdx := arg & 0xFF
		constIdx := (arg >> 8) & 0xFF
		val := frame.Locals[localIdx]
		if val == nil {
			return nil, unboundLocalError(frame, localIdx)
		}
		vm.push(val)
		vm.push(vm.toValue(frame.Code.Constants[constIdx]))

	case OpStoreFastLoadFast:
		storeIdx := arg & 0xFF
		loadIdx := (arg >> 8) & 0xFF
		frame.Locals[storeIdx] = vm.pop()
		val := frame.Locals[loadIdx]
		if val == nil {
			return nil, unboundLocalError(frame, loadIdx)
		}
		vm.push(val)

	case OpLoadConstLoadFast:
		constIdx := (arg >> 8) & 0xFF
		localIdx := arg & 0xFF
		val := frame.Locals[localIdx]
		if val == nil {
			return nil, unboundLocalError(frame, localIdx)
		}
		vm.push(vm.toValue(frame.Code.Constants[constIdx]))
		vm.push(val)

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
		localVal := frame.Locals[localIdx]
		if localVal == nil {
			return nil, unboundLocalError(frame, localIdx)
		}
		vm.push(localVal)

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
		neResult := vm.compareOp(OpCompareNe, a, b)
		if neResult == nil || !vm.truthy(neResult) {
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

	case OpPopBlock:
		if len(frame.BlockStack) > 0 {
			frame.BlockStack = frame.BlockStack[:len(frame.BlockStack)-1]
		}

	case OpPopExceptHandler:
		vm.currentException = nil
		if len(vm.excHandlerStack) > 0 {
			vm.excHandlerStack = vm.excHandlerStack[:len(vm.excHandlerStack)-1]
		}

	case OpClearException:
		if vm.currentException != nil {
			vm.excHandlerStack = append(vm.excHandlerStack, vm.currentException)
		}
		vm.currentException = nil

	case OpSetupWith:
		block := Block{
			Type:    BlockWith,
			Handler: arg,
			Level:   frame.SP,
		}
		frame.BlockStack = append(frame.BlockStack, block)

	case OpWithCleanup:
		exc := vm.pop()
		cm := vm.pop()

		exitMethod, err := vm.getAttr(cm, "__exit__")
		if err != nil {
			return nil, fmt.Errorf("AttributeError: __exit__: %w", err)
		}

		var excType, excVal, excTb Value
		if pyExc, ok := exc.(*PyException); ok {
			if pyExc.ExcType != nil {
				excType = pyExc.ExcType
			} else {
				excType = &PyString{Value: pyExc.Type()}
			}
			excVal = pyExc
			excTb = None
		} else {
			excType = None
			excVal = None
			excTb = None
		}

		var result Value
		switch fn := exitMethod.(type) {
		case *PyMethod:
			result, err = vm.callFunction(fn.Func, []Value{fn.Instance, excType, excVal, excTb}, nil)
		case *PyFunction:
			result, err = vm.callFunction(fn, []Value{cm, excType, excVal, excTb}, nil)
		case *PyBuiltinFunc:
			result, err = fn.Fn([]Value{cm, excType, excVal, excTb}, nil)
		default:
			return nil, fmt.Errorf("TypeError: __exit__ is not callable")
		}
		if err != nil {
			return nil, err
		}

		if vm.truthy(result) {
			vm.currentException = nil
		}

	case OpEndFinally:
		// Pop the handler stack entry that was pushed when entering finally
		if len(vm.excHandlerStack) > 0 {
			vm.excHandlerStack = vm.excHandlerStack[:len(vm.excHandlerStack)-1]
		}
		if vm.currentException != nil {
			exc := vm.currentException
			vm.currentException = nil
			_, err := vm.handleException(exc)
			if err != nil {
				return nil, err
			}
		}
		// Check for pending return from OpReturn through finally block
		if vm.generatorHasPendingReturn {
			vm.generatorHasPendingReturn = false
			result := vm.generatorPendingReturn
			vm.generatorPendingReturn = nil
			return result, nil
		}
		// Check for pending jump from OpContinueLoop through finally block
		if vm.generatorHasPendingJump {
			vm.generatorHasPendingJump = false
			frame.IP = vm.generatorPendingJump
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
			// Bare raise - re-raise the exception from the current handler
			if len(vm.excHandlerStack) > 0 {
				exc = vm.excHandlerStack[len(vm.excHandlerStack)-1]
			} else if vm.lastException != nil {
				exc = vm.lastException
			} else {
				return nil, fmt.Errorf("RuntimeError: No active exception to re-raise")
			}
		} else if arg == 1 {
			// raise exc
			excVal := vm.pop()
			exc = vm.createException(excVal, nil)
		} else {
			// raise exc from cause
			cause := vm.pop()
			excVal := vm.pop()
			exc = vm.createException(excVal, cause)
		}
		// Set implicit __context__ from the exception handler stack
		if arg != 0 && len(vm.excHandlerStack) > 0 {
			handledException := vm.excHandlerStack[len(vm.excHandlerStack)-1]
			if exc != handledException {
				exc.Context = handledException
			}
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

	// ==========================================
	// Additional opcodes for generator support
	// ==========================================

	case OpInplaceAdd, OpInplaceSubtract, OpInplaceMultiply, OpInplaceDivide,
		OpInplaceFloorDiv, OpInplaceModulo, OpInplacePower, OpInplaceMatMul,
		OpInplaceLShift, OpInplaceRShift, OpInplaceAnd, OpInplaceOr, OpInplaceXor:
		b := vm.pop()
		a := vm.pop()

		// Try inplace dunder method on PyInstance first
		var result Value
		var err error
		if inst, ok := a.(*PyInstance); ok {
			var inplaceDunders = [...]string{
				OpInplaceAdd - OpInplaceAdd:     "__iadd__",
				OpInplaceSubtract - OpInplaceAdd: "__isub__",
				OpInplaceMultiply - OpInplaceAdd: "__imul__",
				OpInplaceDivide - OpInplaceAdd:   "__itruediv__",
				OpInplaceFloorDiv - OpInplaceAdd: "__ifloordiv__",
				OpInplaceModulo - OpInplaceAdd:   "__imod__",
				OpInplacePower - OpInplaceAdd:    "__ipow__",
				OpInplaceMatMul - OpInplaceAdd:   "__imatmul__",
				OpInplaceLShift - OpInplaceAdd:   "__ilshift__",
				OpInplaceRShift - OpInplaceAdd:   "__irshift__",
				OpInplaceAnd - OpInplaceAdd:      "__iand__",
				OpInplaceOr - OpInplaceAdd:       "__ior__",
				OpInplaceXor - OpInplaceAdd:      "__ixor__",
			}
			dunder := inplaceDunders[op-OpInplaceAdd]
			var found bool
			result, found, err = vm.callDunder(inst, dunder, b)
			if err != nil {
				return nil, err
			}
			if found && result != nil {
				vm.push(result)
				return nil, nil
			}
		}

		// Fall back to regular binary op
		binOp := op - OpInplaceAdd + OpBinaryAdd
		result, err = vm.binaryOp(binOp, a, b)
		if err != nil {
			return nil, err
		}
		vm.push(result)

	case OpLoadEmptyList:
		frame.Stack[frame.SP] = &PyList{Items: []Value{}}
		frame.SP++

	case OpLoadEmptyTuple:
		frame.Stack[frame.SP] = &PyTuple{Items: []Value{}}
		frame.SP++

	case OpLoadEmptyDict:
		frame.Stack[frame.SP] = &PyDict{Items: make(map[Value]Value)}
		frame.SP++

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

	case OpBinaryPower:
		b := vm.pop()
		a := vm.pop()
		result, err := vm.binaryOp(op, a, b)
		if err != nil {
			return nil, err
		}
		vm.push(result)

	case OpBinaryLShift, OpBinaryRShift, OpBinaryAnd, OpBinaryOr, OpBinaryXor:
		b := vm.pop()
		a := vm.pop()
		result, err := vm.binaryOp(op, a, b)
		if err != nil {
			return nil, err
		}
		vm.push(result)

	case OpUnaryPositive:
		a := vm.pop()
		result, err := vm.unaryOp(op, a)
		if err != nil {
			return nil, err
		}
		vm.push(result)

	case OpUnaryInvert:
		a := vm.pop()
		result, err := vm.unaryOp(op, a)
		if err != nil {
			return nil, err
		}
		vm.push(result)

	case OpDup2:
		b := vm.top()
		a := vm.peek(1)
		vm.push(a)
		vm.push(b)

	case OpDeleteFast:
		vm.callDel(frame.Locals[arg])
		frame.Locals[arg] = nil

	case OpDeleteGlobal:
		name := frame.Code.Names[arg]
		vm.callDel(frame.Globals[name])
		delete(frame.Globals, name)

	case OpDeleteName:
		name := frame.Code.Names[arg]
		vm.callDel(frame.Globals[name])
		delete(frame.Globals, name)

	case OpDeleteSubscr:
		index := vm.pop()
		obj := vm.pop()
		err := vm.delItem(obj, index)
		if err != nil {
			return nil, err
		}

	case OpJumpIfTrue:
		if vm.truthy(vm.top()) {
			frame.IP = arg
		}

	case OpJumpIfFalse:
		if !vm.truthy(vm.top()) {
			frame.IP = arg
		}

	case OpCallKw:
		kwNames, ok := vm.pop().(*PyTuple)
		if !ok {
			return nil, fmt.Errorf("TypeError: internal error: expected keyword names tuple")
		}
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

	case OpCallEx:
		// Call with *args/**kwargs unpacking in generator context
		var kwargs map[string]Value
		if arg&1 != 0 {
			kwargsVal := vm.pop()
			if kwargsDict, ok := kwargsVal.(*PyDict); ok {
				kwargs = make(map[string]Value)
				for _, key := range kwargsDict.Keys(vm) {
					if ks, ok := key.(*PyString); ok {
						val, _ := kwargsDict.DictGet(key, vm)
						kwargs[ks.Value] = val
					}
				}
			}
		}
		argsTuple := vm.pop()
		callable := vm.pop()
		var callArgs []Value
		switch at := argsTuple.(type) {
		case *PyTuple:
			callArgs = at.Items
		case *PyList:
			callArgs = at.Items
		default:
			callArgs = []Value{}
		}
		result, err := vm.call(callable, callArgs, kwargs)
		if err != nil {
			return nil, err
		}
		vm.push(result)

	case OpBuildSet:
		s := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
		for i := 0; i < arg; i++ {
			val := vm.pop()
			s.SetAdd(val, vm)
		}
		vm.push(s)

	case OpSetAdd:
		val := vm.pop()
		set := vm.peek(arg).(*PySet)
		set.SetAdd(val, vm)

	case OpMapAdd:
		val := vm.pop()
		key := vm.pop()
		dict := vm.peek(arg).(*PyDict)
		dict.DictSet(key, val, vm)

	case OpCopyDict:
		keyCount := int(vm.pop().(*PyInt).Value)
		keysToRemove := make([]Value, keyCount)
		for i := keyCount - 1; i >= 0; i-- {
			keysToRemove[i] = vm.pop()
		}
		subject := vm.top()
		dict := subject.(*PyDict)
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

	case OpLoadBuildClass:
		vm.push(vm.builtins["__build_class__"])

	case OpLoadClosure:
		if arg < len(frame.Cells) {
			vm.push(frame.Cells[arg])
		} else {
			return nil, fmt.Errorf("closure cell index %d out of range", arg)
		}

	case OpLoadLocals:
		locals := &PyDict{Items: make(map[Value]Value)}
		for i, name := range frame.Code.VarNames {
			if frame.Locals[i] != nil {
				locals.Items[&PyString{Value: name}] = frame.Locals[i]
			}
		}
		vm.push(locals)

	case OpImportName:
		name := frame.Code.Names[arg]
		fromlist := vm.pop()
		levelVal := vm.pop()
		level := 0
		if levelInt, ok := levelVal.(*PyInt); ok {
			level = int(levelInt.Value)
		}
		moduleName := name
		if level > 0 {
			packageName := ""
			if pkgVal, ok := frame.Globals["__package__"]; ok {
				if pkgStr, ok := pkgVal.(*PyString); ok {
					packageName = pkgStr.Value
				}
			}
			if packageName == "" {
				if nameVal, ok := frame.Globals["__name__"]; ok {
					if nameStr, ok := nameVal.(*PyString); ok {
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
		hasFromlist := false
		if fromlist != nil && fromlist != None {
			if list, ok := fromlist.(*PyList); ok && len(list.Items) > 0 {
				hasFromlist = true
			} else if tuple, ok := fromlist.(*PyTuple); ok && len(tuple.Items) > 0 {
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
		mod := vm.top()
		pyMod, ok := mod.(*PyModule)
		if !ok {
			return nil, fmt.Errorf("import from requires a module, got %s", vm.typeName(mod))
		}
		value, exists := pyMod.Get(name)
		if !exists {
			return nil, fmt.Errorf("cannot import name '%s' from '%s'", name, pyMod.Name)
		}
		vm.push(value)

	default:
		return nil, fmt.Errorf("unimplemented opcode in generator: %s", op)
	}

	return nil, nil
}

// callClassBody executes a class body function with a fresh namespace
// and returns the namespace dict (not the function's return value) and any cells
// (for populating __class__ after class creation)
func (vm *VM) callClassBody(fn *PyFunction) (map[string]Value, []*PyCell, []string, error) {
	code := fn.Code

	// Create a fresh namespace for the class body
	classNamespace := make(map[string]Value)

	// Create new frame with the class namespace as its globals
	// EnclosingGlobals allows the class body to access outer scope variables
	frame := &Frame{
		Code:              code,
		IP:                0,
		Stack:             make([]Value, code.StackSize+16), // Pre-allocate
		SP:                0,
		Locals:            make([]Value, len(code.VarNames)),
		Globals:           classNamespace,
		EnclosingGlobals:  vm.Globals,
		Builtins:          vm.builtins,
		OrderedGlobalKeys: make([]string, 0, 8),
	}

	// Set up cells for the class body (for __class__ cell and other captured variables)
	numCells := len(code.CellVars) + len(code.FreeVars)
	if numCells > 0 || len(fn.Closure) > 0 {
		frame.Cells = make([]*PyCell, numCells)
		// CellVars are new cells for our locals that will be captured
		for i := 0; i < len(code.CellVars); i++ {
			frame.Cells[i] = &PyCell{}
		}
		// FreeVars come from the function's closure (if any)
		for i, cell := range fn.Closure {
			frame.Cells[len(code.CellVars)+i] = cell
		}
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
		return nil, nil, nil, err
	}

	return classNamespace, frame.Cells, frame.OrderedGlobalKeys, nil
}
