package runtime

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"
)

func (vm *VM) run() (Value, error) {
	frame := vm.frame

	for frame.IP < len(frame.Code.Code) {
		// Check for pending memory error from stack growth
		if vm.pendingMemError {
			vm.pendingMemError = false
			memExc := &PyException{
				ExcType:  vm.builtinClass("MemoryError"),
				Args:     &PyTuple{Items: []Value{&PyString{Value: "memory limit exceeded during stack growth"}}},
				Message:  "MemoryError: memory limit exceeded during stack growth",
				TypeName: "MemoryError",
			}
			if _, err := vm.handleException(memExc); err != nil {
				return nil, err
			}
			frame = vm.frame
			continue
		}

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
			if _, exists := frame.Globals[name]; !exists && frame.OrderedGlobalKeys != nil {
				frame.OrderedGlobalKeys = append(frame.OrderedGlobalKeys, name)
			}
			frame.Globals[name] = vm.pop()

		case OpDeleteName:
			name := frame.Code.Names[arg]
			vm.callDel(frame.Globals[name])
			delete(frame.Globals, name)

		case OpLoadFast:
			// Inline push for local variable load
			val := frame.Locals[arg]
			if val == nil {
				return nil, unboundLocalError(frame, arg)
			}
			frame.Stack[frame.SP] = val
			frame.SP++

		case OpStoreFast:
			// Inline pop for local variable store
			frame.SP--
			frame.Locals[arg] = frame.Stack[frame.SP]

		case OpDeleteFast:
			vm.callDel(frame.Locals[arg])
			frame.Locals[arg] = nil

		case OpLoadDeref:
			// Load from closure cell
			// arg indexes into cells: first CellVars, then FreeVars
			if arg < len(frame.Cells) {
				cell := frame.Cells[arg]
				if cell != nil {
					if cell.Value == nil {
						// Cell exists but value is nil — variable deleted or not yet assigned
						varName := frame.Code.CellOrFreeName(arg)
						if arg < len(frame.Code.CellVars) {
							// CellVar (our own local) → UnboundLocalError
							err = fmt.Errorf("UnboundLocalError: cannot access local variable '%s' referenced before assignment", varName)
						} else {
							// FreeVar (from enclosing scope) → NameError
							err = fmt.Errorf("NameError: free variable '%s' referenced before assignment in enclosing scope", varName)
						}
						if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
							return nil, handleErr
						} else if handled {
							continue
						}
						return nil, err
					}
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

		case OpDeleteDeref:
			// Delete cell variable (set to nil) - for 'del' on closure variables
			if arg < len(frame.Cells) && frame.Cells[arg] != nil {
				frame.Cells[arg].Value = nil
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
			val := frame.Locals[0]
			if val == nil {
				return nil, unboundLocalError(frame, 0)
			}
			frame.Stack[frame.SP] = val
			frame.SP++

		case OpLoadFast1:
			val := frame.Locals[1]
			if val == nil {
				return nil, unboundLocalError(frame, 1)
			}
			frame.Stack[frame.SP] = val
			frame.SP++

		case OpLoadFast2:
			val := frame.Locals[2]
			if val == nil {
				return nil, unboundLocalError(frame, 2)
			}
			frame.Stack[frame.SP] = val
			frame.SP++

		case OpLoadFast3:
			val := frame.Locals[3]
			if val == nil {
				return nil, unboundLocalError(frame, 3)
			}
			frame.Stack[frame.SP] = val
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
			if frame.Locals[arg] == nil {
				return nil, unboundLocalError(frame, arg)
			}
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
			val1 := frame.Locals[idx1]
			if val1 == nil {
				return nil, unboundLocalError(frame, idx1)
			}
			val2 := frame.Locals[idx2]
			if val2 == nil {
				return nil, unboundLocalError(frame, idx2)
			}
			vm.ensureStack(2) // Ensure space for two pushes
			frame.Stack[frame.SP] = val1
			frame.SP++
			frame.Stack[frame.SP] = val2
			frame.SP++

		case OpLoadFastLoadConst:
			// Load local then const: arg contains packed indices
			localIdx := arg & 0xFF
			constIdx := (arg >> 8) & 0xFF
			val := frame.Locals[localIdx]
			if val == nil {
				return nil, unboundLocalError(frame, localIdx)
			}
			vm.ensureStack(2) // Ensure space for two pushes
			frame.Stack[frame.SP] = val
			frame.SP++
			frame.Stack[frame.SP] = vm.toValue(frame.Code.Constants[constIdx])
			frame.SP++

		case OpStoreFastLoadFast:
			// Store to local then load another: arg contains packed indices
			storeIdx := arg & 0xFF
			loadIdx := (arg >> 8) & 0xFF
			frame.SP--
			frame.Locals[storeIdx] = frame.Stack[frame.SP]
			val := frame.Locals[loadIdx]
			if val == nil {
				return nil, unboundLocalError(frame, loadIdx)
			}
			frame.Stack[frame.SP] = val
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
						divErr := fmt.Errorf("ZeroDivisionError: division by zero")
						if handled, handleErr := vm.tryHandleError(divErr, frame); handleErr != nil {
							return nil, handleErr
						} else if handled {
							continue
						}
					}
					frame.Stack[frame.SP] = &PyFloat{Value: float64(ai.Value) / float64(bi.Value)}
					frame.SP++
					break
				}
				if bf, ok := b.(*PyFloat); ok {
					if bf.Value == 0 {
						divErr := fmt.Errorf("ZeroDivisionError: float division by zero")
						if handled, handleErr := vm.tryHandleError(divErr, frame); handleErr != nil {
							return nil, handleErr
						} else if handled {
							continue
						}
					}
					frame.Stack[frame.SP] = &PyFloat{Value: float64(ai.Value) / bf.Value}
					frame.SP++
					break
				}
			}
			if af, ok := a.(*PyFloat); ok {
				if bi, ok := b.(*PyInt); ok {
					if bi.Value == 0 {
						divErr := fmt.Errorf("ZeroDivisionError: float division by zero")
						if handled, handleErr := vm.tryHandleError(divErr, frame); handleErr != nil {
							return nil, handleErr
						} else if handled {
							continue
						}
					}
					frame.Stack[frame.SP] = &PyFloat{Value: af.Value / float64(bi.Value)}
					frame.SP++
					break
				}
				if bf, ok := b.(*PyFloat); ok {
					if bf.Value == 0 {
						divErr := fmt.Errorf("ZeroDivisionError: float division by zero")
						if handled, handleErr := vm.tryHandleError(divErr, frame); handleErr != nil {
							return nil, handleErr
						} else if handled {
							continue
						}
					}
					frame.Stack[frame.SP] = &PyFloat{Value: af.Value / bf.Value}
					frame.SP++
					break
				}
			}
			// Fallback
			result, err := vm.binaryOp(OpBinaryDivide, a, b)
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
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

		case OpCompareLtInt, OpCompareLeInt, OpCompareGtInt, OpCompareGeInt, OpCompareEqInt, OpCompareNeInt:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					var cmp bool
					switch op {
					case OpCompareLtInt:
						cmp = ai.Value < bi.Value
					case OpCompareLeInt:
						cmp = ai.Value <= bi.Value
					case OpCompareGtInt:
						cmp = ai.Value > bi.Value
					case OpCompareGeInt:
						cmp = ai.Value >= bi.Value
					case OpCompareEqInt:
						cmp = ai.Value == bi.Value
					case OpCompareNeInt:
						cmp = ai.Value != bi.Value
					}
					if cmp {
						frame.Stack[frame.SP] = True
					} else {
						frame.Stack[frame.SP] = False
					}
					frame.SP++
					break
				}
			}
			// Map specialized opcode back to base compare opcode
			var baseOp Opcode
			switch op {
			case OpCompareLtInt:
				baseOp = OpCompareLt
			case OpCompareLeInt:
				baseOp = OpCompareLe
			case OpCompareGtInt:
				baseOp = OpCompareGt
			case OpCompareGeInt:
				baseOp = OpCompareGe
			case OpCompareEqInt:
				baseOp = OpCompareEq
			case OpCompareNeInt:
				baseOp = OpCompareNe
			}
			frame.Stack[frame.SP] = vm.compareOp(baseOp, a, b)
			if handled, result, err := vm.checkCurrentException(); handled {
				if err != nil {
					return nil, err
				}
				if result != nil {
					return result, nil
				}
				break
			}
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

		case OpCompareLtJump, OpCompareLeJump, OpCompareGtJump, OpCompareGeJump:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if ai, ok := a.(*PyInt); ok {
				if bi, ok := b.(*PyInt); ok {
					var cmp bool
					switch op {
					case OpCompareLtJump:
						cmp = ai.Value < bi.Value
					case OpCompareLeJump:
						cmp = ai.Value <= bi.Value
					case OpCompareGtJump:
						cmp = ai.Value > bi.Value
					case OpCompareGeJump:
						cmp = ai.Value >= bi.Value
					}
					if !cmp {
						frame.IP = arg
					}
					break
				}
			}
			// Fallback: map jump opcode to base compare opcode
			baseOp := OpCompareLt + Opcode(op-OpCompareLtJump)
			cmpResult := vm.compareOp(baseOp, a, b)
			if handled, r, err := vm.checkCurrentException(); handled {
				if err != nil {
					return nil, err
				}
				if r != nil {
					return r, nil
				}
				break
			}
			if !vm.truthy(cmpResult) {
				frame.IP = arg
			}

		case OpCompareEqJump:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			if !vm.equal(a, b) {
				frame.IP = arg
			}

		case OpCompareNeJump:
			frame.SP--
			b := frame.Stack[frame.SP]
			frame.SP--
			a := frame.Stack[frame.SP]
			neResult := vm.compareOp(OpCompareNe, a, b)
			if neResult == nil || !vm.truthy(neResult) {
				frame.IP = arg
			}

		case OpCompareLtLocalJump:
			// Ultra-optimized: compare two locals and jump if false
			// arg format: bits 0-7 = local1, bits 8-15 = local2, bits 16+ = jump offset
			local1 := arg & 0xFF
			local2 := (arg >> 8) & 0xFF
			jumpOffset := arg >> 16
			a := frame.Locals[local1]
			if a == nil {
				return nil, unboundLocalError(frame, local1)
			}
			b := frame.Locals[local2]
			if b == nil {
				return nil, unboundLocalError(frame, local2)
			}
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
			if handled, r, err := vm.checkCurrentException(); handled {
				if err != nil {
					return nil, err
				}
				if r != nil {
					return r, nil
				}
				break
			}
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
			val := frame.Locals[localIdx]
			if val == nil {
				return nil, unboundLocalError(frame, localIdx)
			}
			vm.ensureStack(2) // Ensure space for two pushes
			frame.Stack[frame.SP] = vm.toValue(frame.Code.Constants[constIdx])
			frame.SP++
			frame.Stack[frame.SP] = val
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
			localVal := frame.Locals[localIdx]
			if localVal == nil {
				return nil, unboundLocalError(frame, localIdx)
			}
			frame.Stack[frame.SP] = localVal
			frame.SP++

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
			val := vm.pop()
			if frame.EnclosingGlobals != nil {
				// In class body, explicit 'global' declaration writes to module globals
				frame.EnclosingGlobals[name] = val
			} else {
				frame.Globals[name] = val
			}

		case OpDeleteGlobal:
			name := frame.Code.Names[arg]
			if frame.EnclosingGlobals != nil {
				vm.callDel(frame.EnclosingGlobals[name])
				delete(frame.EnclosingGlobals, name)
			} else {
				vm.callDel(frame.Globals[name])
				delete(frame.Globals, name)
			}

		case OpLoadAttr:
			name := frame.Code.Names[arg]
			obj := vm.pop()
			val, err := vm.getAttr(obj, name)
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			vm.push(val)

		case OpStoreAttr:
			name := frame.Code.Names[arg]
			obj := vm.pop()
			val := vm.pop()
			err = vm.setAttr(obj, name, val)
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}

		case OpDeleteAttr:
			name := frame.Code.Names[arg]
			obj := vm.pop()
			err = vm.delAttr(obj, name)
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}

		case OpBinarySubscr:
			index := vm.pop()
			obj := vm.pop()
			val, err := vm.getItem(obj, index)
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			vm.push(val)

		case OpStoreSubscr:
			index := vm.pop()
			obj := vm.pop()
			val := vm.pop()
			// Check collection size limit for dict insertions of new keys
			if vm.maxCollectionSize > 0 {
				if d, ok := obj.(*PyDict); ok {
					if _, exists := d.DictGet(index, vm); !exists {
						if err := vm.checkCollectionSize(int64(d.DictLen()), "dict"); err != nil {
							if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
								return nil, handleErr
							} else if handled {
								continue
							}
						}
					}
				}
			}
			err = vm.setItem(obj, index, val)
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}

		case OpDeleteSubscr:
			index := vm.pop()
			obj := vm.pop()
			err = vm.delItem(obj, index)
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}

		case OpUnaryPositive:
			a := vm.pop()
			result, err := vm.unaryOp(op, a)
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			vm.push(result)

		case OpUnaryNegative:
			a := vm.pop()
			result, err := vm.unaryOp(op, a)
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			vm.push(result)

		case OpUnaryNot:
			a := vm.pop()
			result := vm.truthy(a)
			if handled, r, err := vm.checkCurrentException(); handled {
				if err != nil {
					return nil, err
				}
				if r != nil {
					return r, nil
				}
				break
			}
			if result {
				vm.push(False)
			} else {
				vm.push(True)
			}

		case OpUnaryInvert:
			a := vm.pop()
			result, err := vm.unaryOp(op, a)
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
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
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
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
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
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
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			frame.Stack[frame.SP] = result
			frame.SP++

		case OpBinaryDivide, OpBinaryFloorDiv, OpBinaryModulo, OpBinaryPower, OpBinaryMatMul,
			OpBinaryLShift, OpBinaryRShift, OpBinaryAnd, OpBinaryOr, OpBinaryXor:
			b := vm.pop()
			a := vm.pop()
			result, err := vm.binaryOp(op, a, b)
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			vm.push(result)

		case OpInplaceAdd, OpInplaceSubtract, OpInplaceMultiply, OpInplaceDivide,
			OpInplaceFloorDiv, OpInplaceModulo, OpInplacePower, OpInplaceMatMul,
			OpInplaceLShift, OpInplaceRShift, OpInplaceAnd, OpInplaceOr, OpInplaceXor:
			b := vm.pop()
			a := vm.pop()

			// Handle list += (extend in place) and list *= (repeat in place)
			if lst, ok := a.(*PyList); ok {
				if op == OpInplaceAdd {
					items, err := vm.toList(b)
					if err != nil {
						if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
							return nil, handleErr
						} else if handled {
							continue
						}
					}
					lst.Items = append(lst.Items, items...)
					vm.push(lst)
					continue
				}
				if op == OpInplaceMultiply {
					if n, ok := b.(*PyInt); ok {
						count := int(n.Value)
						if count <= 0 {
							lst.Items = []Value{}
						} else {
							original := make([]Value, len(lst.Items))
							copy(original, lst.Items)
							for i := 1; i < count; i++ {
								lst.Items = append(lst.Items, original...)
							}
						}
						vm.push(lst)
						continue
					}
					if n, ok := b.(*PyBool); ok {
						count := 0
						if n.Value {
							count = 1
						}
						if count <= 0 {
							lst.Items = []Value{}
						} else {
							// count == 1, no change
						}
						vm.push(lst)
						continue
					}
				}
			}

			// Try inplace dunder method on PyInstance first
			var result Value
			var err error
			if inst, ok := a.(*PyInstance); ok {
				var inplaceDunders = [...]string{
					OpInplaceAdd - OpInplaceAdd:      "__iadd__",
					OpInplaceSubtract - OpInplaceAdd:  "__isub__",
					OpInplaceMultiply - OpInplaceAdd:  "__imul__",
					OpInplaceDivide - OpInplaceAdd:    "__itruediv__",
					OpInplaceFloorDiv - OpInplaceAdd:  "__ifloordiv__",
					OpInplaceModulo - OpInplaceAdd:    "__imod__",
					OpInplacePower - OpInplaceAdd:     "__ipow__",
					OpInplaceMatMul - OpInplaceAdd:    "__imatmul__",
					OpInplaceLShift - OpInplaceAdd:    "__ilshift__",
					OpInplaceRShift - OpInplaceAdd:    "__irshift__",
					OpInplaceAnd - OpInplaceAdd:       "__iand__",
					OpInplaceOr - OpInplaceAdd:        "__ior__",
					OpInplaceXor - OpInplaceAdd:       "__ixor__",
				}
				dunder := inplaceDunders[op-OpInplaceAdd]

				// Check if the dunder is explicitly set to None (blocks inheritance and fallback)
				dunderIsNone := false
				for _, cls := range inst.Class.Mro {
					if method, ok := cls.Dict[dunder]; ok {
						if _, isNone := method.(*PyNone); isNone {
							dunderIsNone = true
						}
						break
					}
				}
				if dunderIsNone {
					err = fmt.Errorf("TypeError: unsupported operand type(s) for %s: '%s' and '%s'",
						dunder, vm.typeName(a), vm.typeName(b))
					if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
						return nil, handleErr
					} else if handled {
						continue
					}
				}

				var found bool
				result, found, err = vm.callDunder(inst, dunder, b)
				if err != nil {
					if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
						return nil, handleErr
					} else if handled {
						continue
					}
				}
				if found && result != nil {
					// Check if result is NotImplemented - if so, fall through to __add__
					if _, isNotImpl := result.(*PyNotImplementedType); !isNotImpl {
						vm.push(result)
						continue
					}
				}
			}

			// Fall back to regular binary op
			binOp := op - OpInplaceAdd + OpBinaryAdd
			result, err = vm.binaryOp(binOp, a, b)
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
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
			if handled, r, err := vm.checkCurrentException(); handled {
				if err != nil {
					return nil, err
				}
				if r != nil {
					return r, nil
				}
				break
			}
			frame.Stack[frame.SP] = result
			frame.SP++

		case OpCompareEq, OpCompareNe, OpCompareLe,
			OpCompareGt, OpCompareGe, OpCompareIs, OpCompareIsNot,
			OpCompareIn, OpCompareNotIn:
			b := vm.pop()
			a := vm.pop()
			result := vm.compareOp(op, a, b)
			if handled, r, err := vm.checkCurrentException(); handled {
				if err != nil {
					return nil, err
				}
				if r != nil {
					return r, nil
				}
				break
			}
			vm.push(result)

		case OpJump:
			frame.IP = arg

		case OpContinueLoop:
			// Continue loop - same as jump in regular VM, handled specially in generators
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
				vm.pop() // Pop iterator before handling error
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
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
			if vm.maxCollectionSize > 0 && int64(arg) > vm.maxCollectionSize {
				err = vm.checkCollectionSize(int64(arg), "list")
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			items := make([]Value, arg)
			for i := arg - 1; i >= 0; i-- {
				items[i] = vm.pop()
			}
			vm.push(&PyList{Items: items})

		case OpBuildSet:
			if vm.maxCollectionSize > 0 && int64(arg) > vm.maxCollectionSize {
				err = vm.checkCollectionSize(int64(arg), "set")
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			s := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
			var buildSetErr error
			for i := 0; i < arg; i++ {
				val := vm.pop()
				if !isHashable(val) {
					buildSetErr = fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(val))
					break
				}
				// Use hash-based storage for O(1) lookup
				s.SetAdd(val, vm)
			}
			if buildSetErr != nil {
				if handled, handleErr := vm.tryHandleError(buildSetErr, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			vm.push(s)

		case OpBuildMap:
			if vm.maxCollectionSize > 0 && int64(arg) > vm.maxCollectionSize {
				err = vm.checkCollectionSize(int64(arg), "dict")
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			d := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
			// Pop all key-value pairs first (they come off stack in reverse order)
			type kvPair struct {
				key, val Value
			}
			pairs := make([]kvPair, arg)
			for i := arg - 1; i >= 0; i-- {
				val := vm.pop()
				key := vm.pop()
				pairs[i] = kvPair{key, val}
			}
			var buildMapErr error
			for _, p := range pairs {
				if !isHashable(p.key) {
					buildMapErr = fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(p.key))
					break
				}
				d.DictSet(p.key, p.val, vm)
			}
			if buildMapErr != nil {
				if handled, handleErr := vm.tryHandleError(buildMapErr, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			vm.push(d)

		case OpUnpackSequence:
			seq := vm.pop()
			items, err := vm.toList(seq)
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			if len(items) != arg {
				err = fmt.Errorf("ValueError: not enough values to unpack (expected %d, got %d)", arg, len(items))
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			for i := len(items) - 1; i >= 0; i-- {
				vm.push(items[i])
			}

		case OpUnpackEx:
			// arg = countBefore | (countAfter << 8)
			countBefore := arg & 0xFF
			countAfter := (arg >> 8) & 0xFF
			seq := vm.pop()
			items, err := vm.toList(seq)
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			totalRequired := countBefore + countAfter
			if len(items) < totalRequired {
				err = fmt.Errorf("ValueError: not enough values to unpack (expected at least %d, got %d)", totalRequired, len(items))
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			// Push in reverse order: after items, starred list, before items
			// After items (from end)
			for i := len(items) - 1; i >= len(items)-countAfter; i-- {
				vm.push(items[i])
			}
			// Starred items as a list
			starItems := make([]Value, len(items)-totalRequired)
			copy(starItems, items[countBefore:len(items)-countAfter])
			vm.push(&PyList{Items: starItems})
			// Before items
			for i := countBefore - 1; i >= 0; i-- {
				vm.push(items[i])
			}

		case OpListAppend:
			val := vm.pop()
			list := vm.peek(arg).(*PyList)
			if vm.maxCollectionSize > 0 {
				if err := vm.checkCollectionSize(int64(len(list.Items)), "list"); err != nil {
					if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
						return nil, handleErr
					} else if handled {
						continue
					}
				}
			}
			list.Items = append(list.Items, val)

		case OpSetAdd:
			val := vm.pop()
			set := vm.peek(arg).(*PySet)
			if vm.maxCollectionSize > 0 {
				if err := vm.checkCollectionSize(int64(set.SetLen()), "set"); err != nil {
					if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
						return nil, handleErr
					} else if handled {
						continue
					}
				}
			}
			if !isHashable(val) {
				err = fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(val))
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			// Use hash-based storage for O(1) lookup
			set.SetAdd(val, vm)

		case OpMapAdd:
			val := vm.pop()
			key := vm.pop()
			dict := vm.peek(arg).(*PyDict)
			if vm.maxCollectionSize > 0 {
				if err := vm.checkCollectionSize(int64(dict.DictLen()), "dict"); err != nil {
					if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
						return nil, handleErr
					} else if handled {
						continue
					}
				}
			}
			if !isHashable(key) {
				err = fmt.Errorf("TypeError: unhashable type: '%s'", vm.typeName(key))
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
			}
			// Use hash-based storage for O(1) lookup
			dict.DictSet(key, val, vm)

		case OpMakeFunction:
			name := vm.pop().(*PyString).Value
			code := vm.pop().(*CodeObject)
			var defaults *PyTuple
			var kwDefaults map[string]Value
			var closure []*PyCell
			if arg&8 != 0 {
				// Has closure - pop tuple of cells
				closureTuple := vm.pop().(*PyTuple)
				closure = make([]*PyCell, len(closureTuple.Items))
				for i, item := range closureTuple.Items {
					closure[i] = item.(*PyCell)
				}
			}
			if arg&2 != 0 {
				// Has kwonly defaults dict
				kwDefaultsDict := vm.pop().(*PyDict)
				kwDefaults = make(map[string]Value)
				for _, key := range kwDefaultsDict.Keys(vm) {
					if ks, ok := key.(*PyString); ok {
						kwDefaults[ks.Value], _ = kwDefaultsDict.DictGet(key, vm)
					}
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
				Code:       code,
				Globals:    fnGlobals,
				Defaults:   defaults,
				KwDefaults: kwDefaults,
				Closure:    closure,
				Name:       name,
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
				// Convert Go error to Python exception so try/except can catch it
				pyExc := vm.wrapGoError(err)
				_, handleErr := vm.handleException(pyExc)
				if handleErr != nil {
					return nil, handleErr
				}
				if vm.frame != frame {
					return nil, errExceptionHandledInOuterFrame
				}
				continue
			}
			vm.push(result)

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
				// Convert Go error to Python exception so try/except can catch it
				pyExc := vm.wrapGoError(err)
				_, handleErr := vm.handleException(pyExc)
				if handleErr != nil {
					return nil, handleErr
				}
				if vm.frame != frame {
					return nil, errExceptionHandledInOuterFrame
				}
				continue
			}
			vm.push(result)

		case OpCallEx:
			// Call with *args/**kwargs unpacking
			// Stack: callable, args_tuple [, kwargs_dict if arg&1]
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
			var args []Value
			switch at := argsTuple.(type) {
			case *PyTuple:
				args = at.Items
			case *PyList:
				args = at.Items
			default:
				// Try to iterate
				iter, iterErr := vm.getIter(argsTuple)
				if iterErr != nil {
					err = fmt.Errorf("TypeError: argument after * must be an iterable")
					pyExc := vm.wrapGoError(err)
					_, handleErr := vm.handleException(pyExc)
					if handleErr != nil {
						return nil, handleErr
					}
					if vm.frame != frame {
						return nil, errExceptionHandledInOuterFrame
					}
					continue
				}
				for {
					val, ok, _ := vm.iterNext(iter)
					if !ok {
						break
					}
					args = append(args, val)
				}
			}
			result, err := vm.call(callable, args, kwargs)
			if err != nil {
				if err == errExceptionHandledInOuterFrame {
					if vm.frame == frame {
						continue
					}
					return nil, err
				}
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
				pyExc := vm.wrapGoError(err)
				_, handleErr := vm.handleException(pyExc)
				if handleErr != nil {
					return nil, handleErr
				}
				if vm.frame != frame {
					return nil, errExceptionHandledInOuterFrame
				}
				continue
			}
			vm.push(result)

		case OpReturn:
			// Inline pop for return
			frame.SP--
			result := frame.Stack[frame.SP]
			if len(vm.frames) > 0 {
				vm.frames = vm.frames[:len(vm.frames)-1]
				if len(vm.frames) > 0 {
					vm.frame = vm.frames[len(vm.frames)-1]
				}
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
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
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
			// Check if method is already bound (PyMethod or bound PyBuiltinFunc)
			alreadyBound := false
			if _, isBound := method.(*PyMethod); isBound {
				alreadyBound = true
			} else if bf, ok := method.(*PyBuiltinFunc); ok && bf.Bound {
				alreadyBound = true
			}
			if alreadyBound {
				// Bound method already has self, don't prepend again
				result, err = vm.call(method, args, nil)
			} else {
				// Raw function, need to prepend self
				allArgs := append([]Value{obj}, args...)
				result, err = vm.call(method, allArgs, nil)
			}
			if err != nil {
				if handled, handleErr := vm.tryHandleError(err, frame); handleErr != nil {
					return nil, handleErr
				} else if handled {
					continue
				}
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
				Type:          BlockExcept,
				Handler:       arg,
				Level:         frame.SP,
				ExcStackLevel: len(vm.excHandlerStack),
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
			// No-exception path: pop BlockExcept from block stack
			if len(frame.BlockStack) > 0 {
				frame.BlockStack = frame.BlockStack[:len(frame.BlockStack)-1]
			}
			vm.currentException = nil

		case OpPopBlock:
			// Pop top block from block stack (used before normal finally entry)
			if len(frame.BlockStack) > 0 {
				frame.BlockStack = frame.BlockStack[:len(frame.BlockStack)-1]
			}

		case OpPopExceptHandler:
			// End of except handler body: pop excHandlerStack entry
			vm.currentException = nil
			if len(vm.excHandlerStack) > 0 {
				vm.excHandlerStack = vm.excHandlerStack[:len(vm.excHandlerStack)-1]
			}

		case OpClearException:
			// Clear the current exception state (when handler catches exception)
			// Does NOT pop the block stack (block was already popped by handleException)
			// Push onto handler stack so new exceptions in this handler get __context__
			if vm.currentException != nil {
				vm.excHandlerStack = append(vm.excHandlerStack, vm.currentException)
			}
			vm.currentException = nil

		case OpSetupExceptStar:
			// Push except* handler block onto block stack
			block := Block{
				Type:    BlockExceptStar,
				Handler: arg,
				Level:   frame.SP,
			}
			frame.BlockStack = append(frame.BlockStack, block)

		case OpExceptStarMatch:
			// Stack: [eg, eg_dup, type] → [eg, subgroup_or_None]
			// Pop type and the DUP'd eg (state is tracked in exceptStarStack)
			excType := vm.pop()
			vm.pop() // Discard DUP'd eg
			if len(vm.exceptStarStack) == 0 {
				vm.push(None)
				break
			}
			state := &vm.exceptStarStack[len(vm.exceptStarStack)-1]
			var matched, remaining []*PyException
			for _, exc := range state.Remaining {
				if vm.exceptionMatches(exc, excType) {
					matched = append(matched, exc)
				} else {
					remaining = append(remaining, exc)
				}
			}
			state.Remaining = remaining
			if len(matched) > 0 {
				// Build a new ExceptionGroup with matched exceptions
				subgroup, err := vm.buildExceptionGroup(state.Message, matched, state.IsBase)
				if err != nil {
					return nil, err
				}
				vm.push(subgroup)
			} else {
				vm.push(None)
			}

		case OpExceptStarReraise:
			// Pop eg from stack
			vm.pop()
			if len(vm.exceptStarStack) == 0 {
				break
			}
			state := vm.exceptStarStack[len(vm.exceptStarStack)-1]
			vm.exceptStarStack = vm.exceptStarStack[:len(vm.exceptStarStack)-1]
			if len(state.Remaining) > 0 {
				// Re-raise unmatched as new ExceptionGroup
				newGroup, err := vm.buildExceptionGroup(state.Message, state.Remaining, state.IsBase)
				if err != nil {
					return nil, err
				}
				exc := vm.createException(newGroup, nil)
				_, herr := vm.handleException(exc)
				if herr != nil {
					return nil, herr
				}
			} else {
				// All matched — clear exception
				vm.currentException = nil
			}

		case OpSetupWith:
			// Push with-cleanup block onto block stack
			block := Block{
				Type:    BlockWith,
				Handler: arg,
				Level:   frame.SP,
			}
			frame.BlockStack = append(frame.BlockStack, block)

		case OpWithCleanup:
			// Exception path: stack has [..., cm, exception]
			// Pop exception, pop context manager, call __exit__(exc_type, exc_val, exc_tb)
			exc := vm.pop() // the exception
			cm := vm.pop()  // the context manager

			// Get __exit__ method
			exitMethod, err := vm.getAttr(cm, "__exit__")
			if err != nil {
				return nil, fmt.Errorf("AttributeError: __exit__: %w", err)
			}

			// Build args: exc_type, exc_val, exc_tb
			var excType, excVal, excTb Value
			if pyExc, ok := exc.(*PyException); ok {
				// exc_type: the exception type (class or string name)
				if pyExc.ExcType != nil {
					excType = pyExc.ExcType
				} else {
					excType = &PyString{Value: pyExc.Type()}
				}
				excVal = pyExc
				excTb = None // traceback not implemented as object
			} else {
				excType = None
				excVal = None
				excTb = None
			}

			// Call __exit__(exc_type, exc_val, exc_tb)
			var result Value
			switch fn := exitMethod.(type) {
			case *PyMethod:
				// Bound method: instance already captured, pass as self
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

			// If __exit__ returns truthy, suppress the exception
			if vm.truthy(result) {
				vm.currentException = nil
			}

		case OpEndFinally:
			// End finally block - re-raise exception if one was active
			// Pop the handler stack entry that was pushed when entering finally
			if len(vm.excHandlerStack) > 0 {
				vm.excHandlerStack = vm.excHandlerStack[:len(vm.excHandlerStack)-1]
			}
			if _, _, err := vm.checkCurrentException(); err != nil {
				return nil, err
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
			} else if arg >= 2 {
				// raise exc from cause
				cause := vm.pop()
				excVal := vm.pop()
				exc = vm.createException(excVal, cause)
			}

			// Set implicit __context__ from the exception handler stack.
			// When we're inside an except block, the handled exception is on
			// the excHandlerStack. Bare raise re-raises, so skip for arg==0.
			if arg != 0 && len(vm.excHandlerStack) > 0 {
				handledException := vm.excHandlerStack[len(vm.excHandlerStack)-1]
				if exc != handledException {
					exc.Context = handledException
				}
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

		case OpSetupAnnotations:
			// Ensure __annotations__ dict exists in current namespace
			if _, ok := frame.Globals["__annotations__"]; !ok {
				frame.Globals["__annotations__"] = &PyDict{Items: make(map[Value]Value)}
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
