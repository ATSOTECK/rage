package runtime

import (
	"context"
	"fmt"
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
