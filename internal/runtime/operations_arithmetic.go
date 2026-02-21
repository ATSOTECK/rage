package runtime

import (
	"fmt"
	"math"
	"math/cmplx"
	"strings"
)

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
		case OpUnaryPositive:
			methodName = "__pos__"
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
	return nil, fmt.Errorf("TypeError: bad operand type for unary %s: '%s'", op.String(), vm.typeName(a))
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
				combined := int64(len(al.Items) + len(bl.Items))
				if vm.maxCollectionSize > 0 && combined > vm.maxCollectionSize {
					return nil, fmt.Errorf("MemoryError: list size limit exceeded (limit is %d)", vm.maxCollectionSize)
				}
				items := make([]Value, combined)
				copy(items, al.Items)
				copy(items[len(al.Items):], bl.Items)
				return &PyList{Items: items}, nil
			}
		}
		// Tuple concatenation
		if at, ok := a.(*PyTuple); ok {
			if bt, ok := b.(*PyTuple); ok {
				combined := int64(len(at.Items) + len(bt.Items))
				if vm.maxCollectionSize > 0 && combined > vm.maxCollectionSize {
					return nil, fmt.Errorf("MemoryError: tuple size limit exceeded (limit is %d)", vm.maxCollectionSize)
				}
				items := make([]Value, combined)
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
				if err := vm.trackAlloc(resultSize); err != nil {
					return nil, err
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
				if err := vm.trackAlloc(resultSize); err != nil {
					return nil, err
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
				if vm.maxCollectionSize > 0 && resultItems > vm.maxCollectionSize {
					return nil, fmt.Errorf("MemoryError: list size limit exceeded (limit is %d)", vm.maxCollectionSize)
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
				if vm.maxCollectionSize > 0 && resultItems > vm.maxCollectionSize {
					return nil, fmt.Errorf("MemoryError: list size limit exceeded (limit is %d)", vm.maxCollectionSize)
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
				if vm.maxCollectionSize > 0 && resultItems > vm.maxCollectionSize {
					return nil, fmt.Errorf("MemoryError: tuple size limit exceeded (limit is %d)", vm.maxCollectionSize)
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
				if vm.maxCollectionSize > 0 && resultItems > vm.maxCollectionSize {
					return nil, fmt.Errorf("MemoryError: tuple size limit exceeded (limit is %d)", vm.maxCollectionSize)
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
