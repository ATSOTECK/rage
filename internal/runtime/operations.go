package runtime

import (
	"fmt"
	"math"
	"math/big"
	"math/cmplx"
	"strconv"
	"strings"
)

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

// hasMethod checks if an instance has a method via MRO lookup.
func (vm *VM) hasMethod(instance *PyInstance, name string) bool {
	for _, cls := range instance.Class.Mro {
		if _, ok := cls.Dict[name]; ok {
			return true
		}
	}
	return false
}

// callDunder looks up and calls a dunder method on an instance via MRO.
// Returns (result, found, error) - found is true if the method exists.
func (vm *VM) callDunder(instance *PyInstance, name string, args ...Value) (Value, bool, error) {
	// Search MRO for the method
	if len(instance.Class.Mro) > 0 {
		for _, cls := range instance.Class.Mro {
			if method, ok := cls.Dict[name]; ok {
				// Prepend instance as self
				allArgs := append([]Value{instance}, args...)
				switch fn := method.(type) {
				case *PyFunction:
					result, err := vm.callFunction(fn, allArgs, nil)
					return result, true, err
				case *PyBuiltinFunc:
					result, err := fn.Fn(allArgs, nil)
					return result, true, err
				}
			}
		}
	} else {
		// Fallback: check instance's class dict directly if MRO is empty
		if method, ok := instance.Class.Dict[name]; ok {
			allArgs := append([]Value{instance}, args...)
			switch fn := method.(type) {
			case *PyFunction:
				result, err := vm.callFunction(fn, allArgs, nil)
				return result, true, err
			case *PyBuiltinFunc:
				result, err := fn.Fn(allArgs, nil)
				return result, true, err
			}
		}
	}
	return nil, false, nil
}

// callDel calls __del__ on a PyInstance if the method exists in its MRO.
// Errors from __del__ are silently ignored (matching CPython behavior).
func (vm *VM) callDel(val Value) {
	inst, ok := val.(*PyInstance)
	if !ok {
		return
	}
	for _, cls := range inst.Class.Mro {
		if method, ok := cls.Dict["__del__"]; ok {
			args := []Value{inst}
			switch fn := method.(type) {
			case *PyFunction:
				vm.callFunction(fn, args, nil)
			case *PyBuiltinFunc:
				fn.Fn(args, nil)
			}
			return
		}
	}
}

func (vm *VM) unaryOp(op Opcode, a Value) (Value, error) {
	// Bool is a subclass of int - coerce for unary operations
	if ab, ok := a.(*PyBool); ok {
		if ab.Value {
			a = MakeInt(1)
		} else {
			a = MakeInt(0)
		}
	}

	// Check for dunder methods on instances
	if inst, ok := a.(*PyInstance); ok {
		var methodName string
		switch op {
		case OpUnaryNegative:
			methodName = "__neg__"
		case OpUnaryInvert:
			methodName = "__invert__"
		}
		if methodName != "" {
			if result, found, err := vm.callDunder(inst, methodName); found {
				return result, err
			}
		}
	}

	switch op {
	case OpUnaryNegative:
		switch v := a.(type) {
		case *PyInt:
			return MakeInt(-v.Value), nil
		case *PyFloat:
			return &PyFloat{Value: -v.Value}, nil
		case *PyComplex:
			return MakeComplex(-v.Real, -v.Imag), nil
		}
	case OpUnaryPositive:
		switch v := a.(type) {
		case *PyInt:
			return v, nil
		case *PyFloat:
			return v, nil
		case *PyComplex:
			return MakeComplex(v.Real, v.Imag), nil
		}
	case OpUnaryInvert:
		if v, ok := a.(*PyInt); ok {
			return MakeInt(^v.Value), nil
		}
	}
	return nil, fmt.Errorf("bad operand type for unary %s: '%s'", op.String(), vm.typeName(a))
}

func (vm *VM) binaryOp(op Opcode, a, b Value) (Value, error) {
	// Bool is a subclass of int in Python - coerce bools to ints for arithmetic
	if ab, ok := a.(*PyBool); ok {
		if ab.Value {
			a = MakeInt(1)
		} else {
			a = MakeInt(0)
		}
	}
	if bb, ok := b.(*PyBool); ok {
		if bb.Value {
			b = MakeInt(1)
		} else {
			b = MakeInt(0)
		}
	}

	// Check for dunder methods on instances first
	dunders := map[Opcode]struct{ forward, reverse string }{
		OpBinaryAdd:      {"__add__", "__radd__"},
		OpBinarySubtract: {"__sub__", "__rsub__"},
		OpBinaryMultiply: {"__mul__", "__rmul__"},
		OpBinaryDivide:   {"__truediv__", "__rtruediv__"},
		OpBinaryFloorDiv: {"__floordiv__", "__rfloordiv__"},
		OpBinaryModulo:   {"__mod__", "__rmod__"},
		OpBinaryPower:    {"__pow__", "__rpow__"},
		OpBinaryMatMul:   {"__matmul__", "__rmatmul__"},
		OpBinaryAnd:      {"__and__", "__rand__"},
		OpBinaryOr:       {"__or__", "__ror__"},
		OpBinaryXor:      {"__xor__", "__rxor__"},
		OpBinaryLShift:   {"__lshift__", "__rlshift__"},
		OpBinaryRShift:   {"__rshift__", "__rrshift__"},
	}

	if dunder, ok := dunders[op]; ok {
		// Try forward method on left operand
		if inst, ok := a.(*PyInstance); ok {
			if result, found, err := vm.callDunder(inst, dunder.forward, b); found {
				// Check for NotImplemented sentinel (we return nil in that case)
				if result != nil {
					return result, err
				}
			}
		}
		// Try reverse method on right operand
		if inst, ok := b.(*PyInstance); ok {
			if result, found, err := vm.callDunder(inst, dunder.reverse, a); found {
				if result != nil {
					return result, err
				}
			}
		}
	}

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
					return nil, fmt.Errorf("ZeroDivisionError: division by zero")
				}
				return &PyFloat{Value: float64(ai.Value) / float64(bi.Value)}, nil
			case OpBinaryFloorDiv:
				if bi.Value == 0 {
					return nil, fmt.Errorf("ZeroDivisionError: integer division or modulo by zero")
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
					return nil, fmt.Errorf("ZeroDivisionError: integer division or modulo by zero")
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
		// Tuple concatenation
		if at, ok := a.(*PyTuple); ok {
			if bt, ok := b.(*PyTuple); ok {
				items := make([]Value, len(at.Items)+len(bt.Items))
				copy(items, at.Items)
				copy(items[len(at.Items):], bt.Items)
				return &PyTuple{Items: items}, nil
			}
		}
		// Bytes concatenation
		if ab, ok := a.(*PyBytes); ok {
			if bb, ok := b.(*PyBytes); ok {
				result := make([]byte, len(ab.Value)+len(bb.Value))
				copy(result, ab.Value)
				copy(result[len(ab.Value):], bb.Value)
				return &PyBytes{Value: result}, nil
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
		if al, ok := b.(*PyList); ok {
			if bi, ok := a.(*PyInt); ok {
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
		// Tuple repetition
		const maxTupleRepeatItems = 10 * 1024 * 1024
		if at, ok := a.(*PyTuple); ok {
			if bi, ok := b.(*PyInt); ok {
				if bi.Value <= 0 {
					return &PyTuple{Items: []Value{}}, nil
				}
				resultItems := int64(len(at.Items)) * bi.Value
				if resultItems > maxTupleRepeatItems {
					return nil, fmt.Errorf("MemoryError: tuple repetition result too large")
				}
				count := int(bi.Value)
				items := make([]Value, 0, len(at.Items)*count)
				for i := 0; i < count; i++ {
					items = append(items, at.Items...)
				}
				return &PyTuple{Items: items}, nil
			}
		}
		if at, ok := b.(*PyTuple); ok {
			if bi, ok := a.(*PyInt); ok {
				if bi.Value <= 0 {
					return &PyTuple{Items: []Value{}}, nil
				}
				resultItems := int64(len(at.Items)) * bi.Value
				if resultItems > maxTupleRepeatItems {
					return nil, fmt.Errorf("MemoryError: tuple repetition result too large")
				}
				count := int(bi.Value)
				items := make([]Value, 0, len(at.Items)*count)
				for i := 0; i < count; i++ {
					items = append(items, at.Items...)
				}
				return &PyTuple{Items: items}, nil
			}
		}
		// Bytes repetition
		const maxBytesRepeatSize = 100 * 1024 * 1024
		if ab, ok := a.(*PyBytes); ok {
			if bi, ok := b.(*PyInt); ok {
				if bi.Value <= 0 {
					return &PyBytes{Value: []byte{}}, nil
				}
				resultSize := int64(len(ab.Value)) * bi.Value
				if resultSize > maxBytesRepeatSize {
					return nil, fmt.Errorf("MemoryError: bytes repetition result too large")
				}
				count := int(bi.Value)
				result := make([]byte, 0, len(ab.Value)*count)
				for i := 0; i < count; i++ {
					result = append(result, ab.Value...)
				}
				return &PyBytes{Value: result}, nil
			}
		}
		if ab, ok := b.(*PyBytes); ok {
			if ai, ok := a.(*PyInt); ok {
				if ai.Value <= 0 {
					return &PyBytes{Value: []byte{}}, nil
				}
				resultSize := int64(len(ab.Value)) * ai.Value
				if resultSize > maxBytesRepeatSize {
					return nil, fmt.Errorf("MemoryError: bytes repetition result too large")
				}
				count := int(ai.Value)
				result := make([]byte, 0, len(ab.Value)*count)
				for i := 0; i < count; i++ {
					result = append(result, ab.Value...)
				}
				return &PyBytes{Value: result}, nil
			}
		}
	}

	// String % formatting (printf-style)
	if op == OpBinaryModulo {
		if as, ok := a.(*PyString); ok {
			return vm.stringFormat(as.Value, b)
		}
	}

	// Dict merge operator: d1 | d2
	if op == OpBinaryOr {
		if ad, ok := a.(*PyDict); ok {
			if bd, ok := b.(*PyDict); ok {
				result := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
				for _, k := range ad.Keys(vm) {
					if v, ok := ad.DictGet(k, vm); ok {
						result.DictSet(k, v, vm)
					}
				}
				for _, k := range bd.Keys(vm) {
					if v, ok := bd.DictGet(k, vm); ok {
						result.DictSet(k, v, vm)
					}
				}
				return result, nil
			}
		}
	}

	// Set/FrozenSet operations: |, &, -, ^
	if op == OpBinaryOr || op == OpBinaryAnd || op == OpBinarySubtract || op == OpBinaryXor {
		// Helper to get items from set or frozenset
		getSetItems := func(v Value) (map[Value]struct{}, bool) {
			switch s := v.(type) {
			case *PySet:
				return s.Items, true
			case *PyFrozenSet:
				return s.Items, true
			}
			return nil, false
		}

		aItems, aIsSet := getSetItems(a)
		bItems, bIsSet := getSetItems(b)

		if aIsSet && bIsSet {
			// Determine return type: frozenset if both are frozenset, otherwise set
			_, aIsFrozen := a.(*PyFrozenSet)
			_, bIsFrozen := b.(*PyFrozenSet)
			returnFrozen := aIsFrozen && bIsFrozen

			switch op {
			case OpBinaryOr: // Union
				if returnFrozen {
					result := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
					for k := range aItems {
						result.FrozenSetAdd(k, vm)
					}
					for k := range bItems {
						result.FrozenSetAdd(k, vm)
					}
					return result, nil
				}
				result := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				for k := range aItems {
					result.SetAdd(k, vm)
				}
				for k := range bItems {
					result.SetAdd(k, vm)
				}
				return result, nil

			case OpBinaryAnd: // Intersection
				if returnFrozen {
					result := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
					for k := range aItems {
						for k2 := range bItems {
							if vm.equal(k, k2) {
								result.FrozenSetAdd(k, vm)
								break
							}
						}
					}
					return result, nil
				}
				result := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				for k := range aItems {
					for k2 := range bItems {
						if vm.equal(k, k2) {
							result.SetAdd(k, vm)
							break
						}
					}
				}
				return result, nil

			case OpBinarySubtract: // Difference
				if returnFrozen {
					result := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
					for k := range aItems {
						found := false
						for k2 := range bItems {
							if vm.equal(k, k2) {
								found = true
								break
							}
						}
						if !found {
							result.FrozenSetAdd(k, vm)
						}
					}
					return result, nil
				}
				result := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				for k := range aItems {
					found := false
					for k2 := range bItems {
						if vm.equal(k, k2) {
							found = true
							break
						}
					}
					if !found {
						result.SetAdd(k, vm)
					}
				}
				return result, nil

			case OpBinaryXor: // Symmetric difference
				if returnFrozen {
					result := &PyFrozenSet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
					// Add items from a not in b
					for k := range aItems {
						found := false
						for k2 := range bItems {
							if vm.equal(k, k2) {
								found = true
								break
							}
						}
						if !found {
							result.FrozenSetAdd(k, vm)
						}
					}
					// Add items from b not in a
					for k := range bItems {
						found := false
						for k2 := range aItems {
							if vm.equal(k, k2) {
								found = true
								break
							}
						}
						if !found {
							result.FrozenSetAdd(k, vm)
						}
					}
					return result, nil
				}
				result := &PySet{Items: make(map[Value]struct{}), buckets: make(map[uint64][]setEntry)}
				for k := range aItems {
					found := false
					for k2 := range bItems {
						if vm.equal(k, k2) {
							found = true
							break
						}
					}
					if !found {
						result.SetAdd(k, vm)
					}
				}
				for k := range bItems {
					found := false
					for k2 := range aItems {
						if vm.equal(k, k2) {
							found = true
							break
						}
					}
					if !found {
						result.SetAdd(k, vm)
					}
				}
				return result, nil
			}
		}
	}

	// Complex operations (including promotion from int/float)
	{
		var ac, bc *PyComplex
		aIsComplex := false
		bIsComplex := false

		if c, ok := a.(*PyComplex); ok {
			ac = c
			aIsComplex = true
		}
		if c, ok := b.(*PyComplex); ok {
			bc = c
			bIsComplex = true
		}

		// Promote non-complex operand if the other is complex
		if aIsComplex && !bIsComplex {
			switch v := b.(type) {
			case *PyInt:
				bc = MakeComplex(float64(v.Value), 0)
				bIsComplex = true
			case *PyFloat:
				bc = MakeComplex(v.Value, 0)
				bIsComplex = true
			}
		}
		if bIsComplex && !aIsComplex {
			switch v := a.(type) {
			case *PyInt:
				ac = MakeComplex(float64(v.Value), 0)
				aIsComplex = true
			case *PyFloat:
				ac = MakeComplex(v.Value, 0)
				aIsComplex = true
			}
		}

		if aIsComplex && bIsComplex {
			switch op {
			case OpBinaryAdd:
				return MakeComplex(ac.Real+bc.Real, ac.Imag+bc.Imag), nil
			case OpBinarySubtract:
				return MakeComplex(ac.Real-bc.Real, ac.Imag-bc.Imag), nil
			case OpBinaryMultiply:
				return MakeComplex(
					ac.Real*bc.Real-ac.Imag*bc.Imag,
					ac.Real*bc.Imag+ac.Imag*bc.Real,
				), nil
			case OpBinaryDivide:
				denom := bc.Real*bc.Real + bc.Imag*bc.Imag
				if denom == 0 {
					return nil, fmt.Errorf("ZeroDivisionError: complex division by zero")
				}
				return MakeComplex(
					(ac.Real*bc.Real+ac.Imag*bc.Imag)/denom,
					(ac.Imag*bc.Real-ac.Real*bc.Imag)/denom,
				), nil
			case OpBinaryPower:
				ca := complex(ac.Real, ac.Imag)
				cb := complex(bc.Real, bc.Imag)
				result := cmplx.Pow(ca, cb)
				return MakeComplex(real(result), imag(result)), nil
			case OpBinaryFloorDiv:
				return nil, fmt.Errorf("TypeError: can't take floor of complex number.")
			case OpBinaryModulo:
				return nil, fmt.Errorf("TypeError: can't mod complex numbers.")
			case OpBinaryLShift, OpBinaryRShift, OpBinaryAnd, OpBinaryOr, OpBinaryXor:
				return nil, fmt.Errorf("TypeError: unsupported operand type(s) for %s: 'complex' and 'complex'", op.String())
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
				return nil, fmt.Errorf("ZeroDivisionError: float division by zero")
			}
			return &PyFloat{Value: af.Value / bf.Value}, nil
		case OpBinaryFloorDiv:
			if bf.Value == 0 {
				return nil, fmt.Errorf("ZeroDivisionError: float floor division by zero")
			}
			return &PyFloat{Value: math.Floor(af.Value / bf.Value)}, nil
		case OpBinaryModulo:
			if bf.Value == 0 {
				return nil, fmt.Errorf("ZeroDivisionError: float modulo")
			}
			// Python modulo: result has same sign as divisor
			r := math.Mod(af.Value, bf.Value)
			if r != 0 && (r < 0) != (bf.Value < 0) {
				r += bf.Value
			}
			return &PyFloat{Value: r}, nil
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

	// Set comparison operators: <=, <, >=, > mean subset/superset
	if as, aOk := a.(*PySet); aOk {
		if bs, bOk := b.(*PySet); bOk {
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
			case OpCompareLt: // proper subset
				if len(as.Items) >= len(bs.Items) {
					return False
				}
				for k := range as.Items {
					if !bs.SetContains(k, vm) {
						return False
					}
				}
				return True
			case OpCompareLe: // subset
				for k := range as.Items {
					if !bs.SetContains(k, vm) {
						return False
					}
				}
				return True
			case OpCompareGt: // proper superset
				if len(as.Items) <= len(bs.Items) {
					return False
				}
				for k := range bs.Items {
					if !as.SetContains(k, vm) {
						return False
					}
				}
				return True
			case OpCompareGe: // superset
				for k := range bs.Items {
					if !as.SetContains(k, vm) {
						return False
					}
				}
				return True
			}
		}
	}

	// Complex numbers support == and != but not ordering
	_, aIsComplex := a.(*PyComplex)
	_, bIsComplex := b.(*PyComplex)

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
		if aIsComplex || bIsComplex {
			vm.currentException = &PyException{TypeName: "TypeError", Message: "'<' not supported between instances of '" + vm.typeName(a) + "' and '" + vm.typeName(b) + "'"}
			return nil
		}
		if vm.compare(a, b) < 0 {
			return True
		}
		return False
	case OpCompareLe:
		if aIsComplex || bIsComplex {
			vm.currentException = &PyException{TypeName: "TypeError", Message: "'<=' not supported between instances of '" + vm.typeName(a) + "' and '" + vm.typeName(b) + "'"}
			return nil
		}
		if vm.compare(a, b) <= 0 {
			return True
		}
		return False
	case OpCompareGt:
		if aIsComplex || bIsComplex {
			vm.currentException = &PyException{TypeName: "TypeError", Message: "'>' not supported between instances of '" + vm.typeName(a) + "' and '" + vm.typeName(b) + "'"}
			return nil
		}
		if vm.compare(a, b) > 0 {
			return True
		}
		return False
	case OpCompareGe:
		if aIsComplex || bIsComplex {
			vm.currentException = &PyException{TypeName: "TypeError", Message: "'>=' not supported between instances of '" + vm.typeName(a) + "' and '" + vm.typeName(b) + "'"}
			return nil
		}
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
		switch bv := b.(type) {
		case *PyBool:
			return av.Value == bv.Value
		case *PyInt:
			ai := int64(0)
			if av.Value {
				ai = 1
			}
			return ai == bv.Value
		case *PyFloat:
			ai := 0.0
			if av.Value {
				ai = 1.0
			}
			return ai == bv.Value
		}
	case *PyInt:
		switch bv := b.(type) {
		case *PyInt:
			if av.BigValue != nil || bv.BigValue != nil {
				return av.BigIntValue().Cmp(bv.BigIntValue()) == 0
			}
			return av.Value == bv.Value
		case *PyFloat:
			if av.BigValue != nil {
				bf := new(big.Float).SetInt(av.BigIntValue())
				return bf.Cmp(big.NewFloat(bv.Value)) == 0
			}
			return float64(av.Value) == bv.Value
		case *PyBool:
			bi := int64(0)
			if bv.Value {
				bi = 1
			}
			return av.Value == bi
		case *PyComplex:
			return bv.Imag == 0 && float64(av.Value) == bv.Real
		}
	case *PyFloat:
		switch bv := b.(type) {
		case *PyFloat:
			return av.Value == bv.Value
		case *PyInt:
			return av.Value == float64(bv.Value)
		case *PyComplex:
			return bv.Imag == 0 && av.Value == bv.Real
		}
	case *PyComplex:
		switch bv := b.(type) {
		case *PyComplex:
			return av.Real == bv.Real && av.Imag == bv.Imag
		case *PyInt:
			return av.Imag == 0 && av.Real == float64(bv.Value)
		case *PyFloat:
			return av.Imag == 0 && av.Real == bv.Value
		case *PyBool:
			bi := int64(0)
			if bv.Value {
				bi = 1
			}
			return av.Imag == 0 && av.Real == float64(bi)
		}
	case *PyString:
		if bv, ok := b.(*PyString); ok {
			return av.Value == bv.Value
		}
	case *PyBytes:
		if bv, ok := b.(*PyBytes); ok {
			if len(av.Value) != len(bv.Value) {
				return false
			}
			for i := range av.Value {
				if av.Value[i] != bv.Value[i] {
					return false
				}
			}
			return true
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
	case *PySet:
		// Set can equal another set or frozenset
		switch bv := b.(type) {
		case *PySet:
			if len(av.Items) != len(bv.Items) {
				return false
			}
			for k := range av.Items {
				if !bv.SetContains(k, vm) {
					return false
				}
			}
			return true
		case *PyFrozenSet:
			if len(av.Items) != len(bv.Items) {
				return false
			}
			for k := range av.Items {
				if !bv.FrozenSetContains(k, vm) {
					return false
				}
			}
			return true
		}
	case *PyFrozenSet:
		// FrozenSet can equal another frozenset or set
		switch bv := b.(type) {
		case *PyFrozenSet:
			if len(av.Items) != len(bv.Items) {
				return false
			}
			for k := range av.Items {
				if !bv.FrozenSetContains(k, vm) {
					return false
				}
			}
			return true
		case *PySet:
			if len(av.Items) != len(bv.Items) {
				return false
			}
			for k := range av.Items {
				if !bv.SetContains(k, vm) {
					return false
				}
			}
			return true
		}
	case *PyRange:
		if bv, ok := b.(*PyRange); ok {
			// Two ranges are equal if they produce the same sequence
			aLen := rangeLen(av)
			bLen := rangeLen(bv)
			if aLen != bLen {
				return false
			}
			if aLen == 0 {
				return true
			}
			if av.Start != bv.Start {
				return false
			}
			if aLen == 1 {
				return true
			}
			return av.Step == bv.Step
		}
	case *PyClass:
		// Classes are compared by identity
		return a == b
	case *PyInstance:
		// Check for __eq__ method
		if result, found, err := vm.callDunder(av, "__eq__", b); found && err == nil {
			if bv, ok := result.(*PyBool); ok {
				return bv.Value
			}
		}
		// Fall back to identity comparison
		return a == b
	}
	// Check if b is a PyInstance with __eq__
	if bv, ok := b.(*PyInstance); ok {
		if result, found, err := vm.callDunder(bv, "__eq__", a); found && err == nil {
			if boolVal, ok := result.(*PyBool); ok {
				return boolVal.Value
			}
		}
	}
	return false
}

func (vm *VM) compare(a, b Value) int {
	// Bool is a subclass of int - coerce for comparison
	if ab, ok := a.(*PyBool); ok {
		if ab.Value {
			a = MakeInt(1)
		} else {
			a = MakeInt(0)
		}
	}
	if bb, ok := b.(*PyBool); ok {
		if bb.Value {
			b = MakeInt(1)
		} else {
			b = MakeInt(0)
		}
	}

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
	case *PyBytes:
		if bv, ok := b.(*PyBytes); ok {
			minLen := len(av.Value)
			if len(bv.Value) < minLen {
				minLen = len(bv.Value)
			}
			for i := 0; i < minLen; i++ {
				if av.Value[i] < bv.Value[i] {
					return -1
				} else if av.Value[i] > bv.Value[i] {
					return 1
				}
			}
			if len(av.Value) < len(bv.Value) {
				return -1
			} else if len(av.Value) > len(bv.Value) {
				return 1
			}
			return 0
		}
	case *PyList:
		if bv, ok := b.(*PyList); ok {
			minLen := len(av.Items)
			if len(bv.Items) < minLen {
				minLen = len(bv.Items)
			}
			for i := 0; i < minLen; i++ {
				c := vm.compare(av.Items[i], bv.Items[i])
				if c != 0 {
					return c
				}
			}
			if len(av.Items) < len(bv.Items) {
				return -1
			} else if len(av.Items) > len(bv.Items) {
				return 1
			}
			return 0
		}
	case *PyTuple:
		if bv, ok := b.(*PyTuple); ok {
			minLen := len(av.Items)
			if len(bv.Items) < minLen {
				minLen = len(bv.Items)
			}
			for i := 0; i < minLen; i++ {
				c := vm.compare(av.Items[i], bv.Items[i])
				if c != 0 {
					return c
				}
			}
			if len(av.Items) < len(bv.Items) {
				return -1
			} else if len(av.Items) > len(bv.Items) {
				return 1
			}
			return 0
		}
	case *PyInstance:
		// Try __lt__ first
		if result, found, err := vm.callDunder(av, "__lt__", b); found && err == nil {
			if boolVal, ok := result.(*PyBool); ok && boolVal.Value {
				return -1
			}
		}
		// Try __gt__
		if result, found, err := vm.callDunder(av, "__gt__", b); found && err == nil {
			if boolVal, ok := result.(*PyBool); ok && boolVal.Value {
				return 1
			}
		}
		// Try __eq__ for equality
		if result, found, err := vm.callDunder(av, "__eq__", b); found && err == nil {
			if boolVal, ok := result.(*PyBool); ok && boolVal.Value {
				return 0
			}
		}
		return 0
	}
	// Check if b is a PyInstance
	if bv, ok := b.(*PyInstance); ok {
		// Try __gt__ on b (means a < b)
		if result, found, err := vm.callDunder(bv, "__gt__", a); found && err == nil {
			if boolVal, ok := result.(*PyBool); ok && boolVal.Value {
				return -1
			}
		}
		// Try __lt__ on b (means a > b)
		if result, found, err := vm.callDunder(bv, "__lt__", a); found && err == nil {
			if boolVal, ok := result.(*PyBool); ok && boolVal.Value {
				return 1
			}
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
	case *PyFrozenSet:
		// Use hash-based lookup for O(1) average case
		return c.FrozenSetContains(item, vm)
	case *PyDict:
		// Use hash-based lookup for O(1) average case
		return c.DictContains(item, vm)
	case *PyRange:
		if i, ok := item.(*PyInt); ok && i.BigValue == nil {
			return c.Contains(i.Value)
		}
		if b, ok := item.(*PyBool); ok {
			v := int64(0)
			if b.Value {
				v = 1
			}
			return c.Contains(v)
		}
	case *PyBytes:
		if i, ok := item.(*PyInt); ok && i.BigValue == nil {
			for _, b := range c.Value {
				if int64(b) == i.Value {
					return true
				}
			}
		}
		if sub, ok := item.(*PyBytes); ok {
			if len(sub.Value) == 0 {
				return true
			}
			return bytesContains(c.Value, sub.Value)
		}
	case *PyInstance:
		// Check for __contains__ method
		if result, found, err := vm.callDunder(c, "__contains__", item); found && err == nil {
			if boolVal, ok := result.(*PyBool); ok {
				return boolVal.Value
			}
			// Python's __contains__ can return truthy values
			return vm.truthy(result)
		}
		// Fall back to iterating via __iter__
		if iter, err := vm.getIter(c); err == nil {
			for {
				val, done, err := vm.iterNext(iter)
				if done || err != nil {
					break
				}
				if vm.equal(val, item) {
					return true
				}
			}
		}
	}
	return false
}

// bytesContains checks if sub is a subsequence of data
func bytesContains(data, sub []byte) bool {
	if len(sub) > len(data) {
		return false
	}
	for i := 0; i <= len(data)-len(sub); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
			if data[i+j] != sub[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// stringFormat implements Python's "string % args" printf-style formatting
func (vm *VM) stringFormat(format string, args Value) (Value, error) {
	// Check if this is dict-based formatting: %(key)s
	dictArgs, isDict := args.(*PyDict)

	var argList []Value
	if !isDict {
		switch v := args.(type) {
		case *PyTuple:
			argList = v.Items
		default:
			argList = []Value{args}
		}
	}

	var result strings.Builder
	argIdx := 0
	i := 0
	for i < len(format) {
		if format[i] == '%' {
			i++
			if i >= len(format) {
				return nil, fmt.Errorf("TypeError: incomplete format")
			}
			if format[i] == '%' {
				result.WriteByte('%')
				i++
				continue
			}

			// Parse mapping key: %(name)s
			var arg Value
			mappingKey := ""
			hasMappingKey := false
			if format[i] == '(' {
				if !isDict {
					return nil, fmt.Errorf("TypeError: format requires a mapping")
				}
				i++
				start := i
				for i < len(format) && format[i] != ')' {
					i++
				}
				if i >= len(format) {
					return nil, fmt.Errorf("TypeError: incomplete format key")
				}
				mappingKey = format[start:i]
				hasMappingKey = true
				i++ // skip ')'
			}

			// Parse flags: #, 0, -, ' ', +
			leftAlign := false
			zeroPad := false
			signPlus := false
			signSpace := false
			altForm := false
			for i < len(format) {
				switch format[i] {
				case '-':
					leftAlign = true
					i++
				case '0':
					zeroPad = true
					i++
				case '+':
					signPlus = true
					i++
				case ' ':
					signSpace = true
					i++
				case '#':
					altForm = true
					i++
				default:
					goto flagsDone
				}
			}
		flagsDone:

			// Parse width (* means take from args)
			width := 0
			if i < len(format) && format[i] == '*' {
				i++
				if argIdx >= len(argList) {
					return nil, fmt.Errorf("TypeError: not enough arguments for format string")
				}
				width = int(vm.toInt(argList[argIdx]))
				argIdx++
				if width < 0 {
					leftAlign = true
					width = -width
				}
			} else {
				for i < len(format) && format[i] >= '0' && format[i] <= '9' {
					width = width*10 + int(format[i]-'0')
					i++
				}
			}

			// Parse precision
			precision := -1
			if i < len(format) && format[i] == '.' {
				i++
				precision = 0
				if i < len(format) && format[i] == '*' {
					i++
					if argIdx >= len(argList) {
						return nil, fmt.Errorf("TypeError: not enough arguments for format string")
					}
					precision = int(vm.toInt(argList[argIdx]))
					argIdx++
					if precision < 0 {
						precision = 0
					}
				} else {
					for i < len(format) && format[i] >= '0' && format[i] <= '9' {
						precision = precision*10 + int(format[i]-'0')
						i++
					}
				}
			}

			if i >= len(format) {
				return nil, fmt.Errorf("TypeError: incomplete format")
			}

			spec := format[i]
			i++

			// Get the argument
			if hasMappingKey {
				val, found := dictArgs.DictGet(&PyString{Value: mappingKey}, vm)
				if !found {
					return nil, fmt.Errorf("KeyError: '%s'", mappingKey)
				}
				arg = val
			} else {
				if argIdx >= len(argList) {
					return nil, fmt.Errorf("TypeError: not enough arguments for format string")
				}
				arg = argList[argIdx]
				argIdx++
			}

			var s string
			isNumeric := false
			switch spec {
			case 's':
				s = vm.str(arg)
				if precision >= 0 && len(s) > precision {
					s = s[:precision]
				}
			case 'r':
				s = vm.repr(arg)
				if precision >= 0 && len(s) > precision {
					s = s[:precision]
				}
			case 'a':
				s = vm.repr(arg) // approximate ascii() with repr()
				if precision >= 0 && len(s) > precision {
					s = s[:precision]
				}
			case 'd', 'i', 'u':
				isNumeric = true
				n := vm.toInt(arg)
				s = fmt.Sprintf("%d", n)
			case 'f', 'F':
				isNumeric = true
				f := vm.toFloat(arg)
				if precision < 0 {
					precision = 6
				}
				s = strconv.FormatFloat(f, 'f', precision, 64)
				if spec == 'F' {
					s = strings.ToUpper(s)
				}
			case 'e':
				isNumeric = true
				f := vm.toFloat(arg)
				if precision < 0 {
					precision = 6
				}
				s = strconv.FormatFloat(f, 'e', precision, 64)
			case 'E':
				isNumeric = true
				f := vm.toFloat(arg)
				if precision < 0 {
					precision = 6
				}
				s = strconv.FormatFloat(f, 'E', precision, 64)
			case 'g':
				isNumeric = true
				f := vm.toFloat(arg)
				if precision < 0 {
					precision = 6
				}
				s = strconv.FormatFloat(f, 'g', precision, 64)
			case 'G':
				isNumeric = true
				f := vm.toFloat(arg)
				if precision < 0 {
					precision = 6
				}
				s = strconv.FormatFloat(f, 'G', precision, 64)
			case 'x':
				isNumeric = true
				n := vm.toInt(arg)
				if n < 0 {
					s = fmt.Sprintf("-%x", -n)
				} else {
					s = fmt.Sprintf("%x", n)
				}
				if altForm && n != 0 {
					s = "0x" + s
				}
			case 'X':
				isNumeric = true
				n := vm.toInt(arg)
				if n < 0 {
					s = fmt.Sprintf("-%X", -n)
				} else {
					s = fmt.Sprintf("%X", n)
				}
				if altForm && n != 0 {
					s = "0X" + s
				}
			case 'o':
				isNumeric = true
				n := vm.toInt(arg)
				if n < 0 {
					s = fmt.Sprintf("-%o", -n)
				} else {
					s = fmt.Sprintf("%o", n)
				}
				if altForm && n != 0 {
					s = "0o" + s
				}
			case 'c':
				switch v := arg.(type) {
				case *PyInt:
					s = string(rune(v.Value))
				case *PyString:
					if len(v.Value) != 1 {
						return nil, fmt.Errorf("%%c requires int or char")
					}
					s = v.Value
				default:
					return nil, fmt.Errorf("%%c requires int or char")
				}
			default:
				return nil, fmt.Errorf("TypeError: unsupported format character '%c' (0x%x)", spec, spec)
			}

			// Apply sign for numeric types
			if isNumeric && s != "" && s[0] != '-' {
				if signPlus {
					s = "+" + s
				} else if signSpace {
					s = " " + s
				}
			}

			// Apply width and alignment
			if width > 0 && len(s) < width {
				pad := width - len(s)
				if leftAlign {
					s = s + strings.Repeat(" ", pad)
				} else if zeroPad && isNumeric && !leftAlign {
					// Zero-pad after sign prefix
					prefix := ""
					num := s
					if len(s) > 0 && (s[0] == '-' || s[0] == '+' || s[0] == ' ') {
						prefix = s[:1]
						num = s[1:]
					}
					s = prefix + strings.Repeat("0", width-len(prefix)-len(num)) + num
				} else {
					s = strings.Repeat(" ", pad) + s
				}
			}

			result.WriteString(s)
		} else {
			result.WriteByte(format[i])
			i++
		}
	}

	// Check for too many arguments (only for positional, not dict)
	if !isDict && argIdx < len(argList) {
		return nil, fmt.Errorf("TypeError: not all arguments converted during string formatting")
	}

	return &PyString{Value: result.String()}, nil
}
