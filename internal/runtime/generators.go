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
// and returns the namespace dict (not the function's return value) and any cells
// (for populating __class__ after class creation)
func (vm *VM) callClassBody(fn *PyFunction) (map[string]Value, []*PyCell, error) {
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
		return nil, nil, err
	}

	return classNamespace, frame.Cells, nil
}
