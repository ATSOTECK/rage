package runtime

import (
	"fmt"
	"math"
	"math/big"
	"strings"
)

// safeIntCount converts an int64 to int, returning an error if the value
// exceeds the platform's int range (relevant on 32-bit systems).
func safeIntCount(v int64) (int, error) {
	if v > math.MaxInt {
		return 0, fmt.Errorf("MemoryError: repetition count too large for platform")
	}
	return int(v), nil
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
			return intNegate(v), nil
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
			if v.BigValue != nil {
				return MakeBigInt(new(big.Int).Not(v.BigValue)), nil
			}
			return MakeInt(^v.Value), nil
		}
	}
	sym := map[Opcode]string{OpUnaryNegative: "-", OpUnaryPositive: "+", OpUnaryInvert: "~"}[op]
	return nil, fmt.Errorf("TypeError: bad operand type for unary %s: '%s'", sym, vm.typeName(a))
}

// inplaceOp implements an augmented assignment (a op= b) for built-in types
// that update in place: list +=, set |= &= -= ^= and dict |=. Everything else
// falls back to the corresponding binary operator.
func (vm *VM) inplaceOp(op Opcode, a, b Value) (Value, error) {
	switch av := a.(type) {
	case *PyList:
		if op == OpInplaceAdd {
			items, err := vm.toList(b)
			if err != nil {
				return nil, err
			}
			if vm.maxCollectionSize > 0 && int64(len(av.Items)+len(items)) > vm.maxCollectionSize {
				return nil, fmt.Errorf("MemoryError: list size limit exceeded (limit is %d)", vm.maxCollectionSize)
			}
			av.Items = append(av.Items, items...)
			return av, nil
		}
	case *PySet:
		if handled, err := vm.setInplaceOp(op, av, b); handled {
			if err != nil {
				return nil, err
			}
			return av, nil
		}
	case *PyDict:
		if op == OpInplaceOr {
			// dict |= accepts anything dict.update() does
			update, err := vm.getAttrDict(av, "update")
			if err != nil {
				return nil, err
			}
			if _, err := update.(*PyBuiltinFunc).Fn([]Value{b}, nil); err != nil {
				return nil, err
			}
			return av, nil
		}
	}
	return vm.binaryOp(op-OpInplaceAdd+OpBinaryAdd, a, b)
}

// binaryOpDunders returns the forward and reflected dunder method names for a
// binary opcode. ok is false for opcodes that have no dunder protocol.
func binaryOpDunders(op Opcode) (forward, reverse string, ok bool) {
	switch op {
	case OpBinaryAdd:
		return "__add__", "__radd__", true
	case OpBinarySubtract:
		return "__sub__", "__rsub__", true
	case OpBinaryMultiply:
		return "__mul__", "__rmul__", true
	case OpBinaryDivide:
		return "__truediv__", "__rtruediv__", true
	case OpBinaryFloorDiv:
		return "__floordiv__", "__rfloordiv__", true
	case OpBinaryModulo:
		return "__mod__", "__rmod__", true
	case OpBinaryPower:
		return "__pow__", "__rpow__", true
	case OpBinaryMatMul:
		return "__matmul__", "__rmatmul__", true
	case OpBinaryAnd:
		return "__and__", "__rand__", true
	case OpBinaryOr:
		return "__or__", "__ror__", true
	case OpBinaryXor:
		return "__xor__", "__rxor__", true
	case OpBinaryLShift:
		return "__lshift__", "__rlshift__", true
	case OpBinaryRShift:
		return "__rshift__", "__rrshift__", true
	}
	return "", "", false
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
	instA, aIsInst := a.(*PyInstance)
	instB, bIsInst := b.(*PyInstance)
	if aIsInst || bIsInst {
		if forward, reverse, ok := binaryOpDunders(op); ok {
			// Try forward method on left operand
			if aIsInst {
				if result, found, err := vm.callDunder(instA, forward, b); found {
					// Check for NotImplemented sentinel (we return nil in that case)
					if result != nil {
						return result, err
					}
				}
			}
			// Try reverse method on right operand
			if bIsInst {
				if result, found, err := vm.callDunder(instB, reverse, a); found {
					if result != nil {
						return result, err
					}
				}
			}
		}
	}

	// int op int, with arbitrary precision
	if ai, ok := a.(*PyInt); ok {
		if bi, ok := b.(*PyInt); ok {
			if result, handled, err := vm.intBinaryOp(op, ai, bi); handled {
				return result, err
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

	// Sequence repetition by an int too large for an index
	if op == OpBinaryMultiply {
		for _, pair := range [2][2]Value{{a, b}, {b, a}} {
			if n, ok := pair[1].(*PyInt); ok && n.BigValue != nil {
				switch seq := pair[0].(type) {
				case *PyString, *PyList, *PyTuple, *PyBytes:
					if n.BigValue.Sign() > 0 {
						return nil, fmt.Errorf("OverflowError: cannot fit 'int' into an index-sized integer")
					}
					return vm.binaryOp(op, seq, MakeInt(0))
				}
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
				count, err := safeIntCount(bi.Value)
				if err != nil {
					return nil, err
				}
				return &PyString{Value: strings.Repeat(as.Value, count)}, nil
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
				count, err := safeIntCount(ai.Value)
				if err != nil {
					return nil, err
				}
				return &PyString{Value: strings.Repeat(as.Value, count)}, nil
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
				count, err := safeIntCount(bi.Value)
				if err != nil {
					return nil, err
				}
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
				count, err := safeIntCount(bi.Value)
				if err != nil {
					return nil, err
				}
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
				count, err := safeIntCount(bi.Value)
				if err != nil {
					return nil, err
				}
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
				count, err := safeIntCount(bi.Value)
				if err != nil {
					return nil, err
				}
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
				count, err := safeIntCount(bi.Value)
				if err != nil {
					return nil, err
				}
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
				count, err := safeIntCount(ai.Value)
				if err != nil {
					return nil, err
				}
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
		if result, ok, err := vm.setBinaryOp(op, a, b); ok {
			return result, err
		}
		if result, ok, err := vm.viewSetOp(op, a, b); ok {
			return result, err
		}
	}

	// PEP 604: type | type creates UnionType
	if op == OpBinaryOr {
		if isUnionableType(a) && isUnionableType(b) {
			return MakeUnionType(a, b), nil
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
		promote := func(v Value) (*PyComplex, bool, error) {
			switch n := v.(type) {
			case *PyInt:
				f, err := intToFloat(n)
				return MakeComplex(f, 0), true, err
			case *PyFloat:
				return MakeComplex(n.Value, 0), true, nil
			}
			return nil, false, nil
		}
		var err error
		if aIsComplex && !bIsComplex {
			if bc, bIsComplex, err = promote(b); err != nil {
				return nil, err
			}
		}
		if bIsComplex && !aIsComplex {
			if ac, aIsComplex, err = promote(a); err != nil {
				return nil, err
			}
		}

		if aIsComplex && bIsComplex {
			ca := complex(ac.Real, ac.Imag)
			cb := complex(bc.Real, bc.Imag)
			switch op {
			case OpBinaryAdd:
				return MakeComplex(ac.Real+bc.Real, ac.Imag+bc.Imag), nil
			case OpBinarySubtract:
				return MakeComplex(ac.Real-bc.Real, ac.Imag-bc.Imag), nil
			case OpBinaryMultiply:
				r := complexProd(ca, cb)
				return MakeComplex(real(r), imag(r)), nil
			case OpBinaryDivide:
				r, err := complexQuot(ca, cb)
				if err != nil {
					return nil, err
				}
				return MakeComplex(real(r), imag(r)), nil
			case OpBinaryPower:
				r, err := complexPow(ca, cb)
				if err != nil {
					return nil, err
				}
				return MakeComplex(real(r), imag(r)), nil
			case OpBinaryFloorDiv, OpBinaryModulo, OpBinaryLShift, OpBinaryRShift, OpBinaryAnd, OpBinaryOr, OpBinaryXor:
				return nil, fmt.Errorf("TypeError: unsupported operand type(s) for %s: '%s' and '%s'", binaryOpSymbol(op), vm.typeName(a), vm.typeName(b))
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
		f, err := intToFloat(ai)
		if err != nil {
			return nil, err
		}
		af = &PyFloat{Value: f}
		aIsFloat = true
	}
	if aIsFloat && bIsInt {
		f, err := intToFloat(bi)
		if err != nil {
			return nil, err
		}
		bf = &PyFloat{Value: f}
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
				return nil, fmt.Errorf("ZeroDivisionError: division by zero")
			}
			return &PyFloat{Value: af.Value / bf.Value}, nil
		case OpBinaryFloorDiv:
			if bf.Value == 0 {
				return nil, fmt.Errorf("ZeroDivisionError: division by zero")
			}
			q, _ := floatDivMod(af.Value, bf.Value)
			return &PyFloat{Value: q}, nil
		case OpBinaryModulo:
			if bf.Value == 0 {
				return nil, fmt.Errorf("ZeroDivisionError: division by zero")
			}
			_, r := floatDivMod(af.Value, bf.Value)
			return &PyFloat{Value: r}, nil
		case OpBinaryPower:
			return floatPow(af.Value, bf.Value)
		}
	}

	// CPython's wording for sequence concatenation and repetition errors
	switch op {
	case OpBinaryAdd:
		switch a.(type) {
		case *PyList, *PyTuple, *PyString:
			return nil, fmt.Errorf("TypeError: can only concatenate %s (not \"%s\") to %s", vm.typeName(a), vm.typeName(b), vm.typeName(a))
		case *PyBytes:
			return nil, fmt.Errorf("TypeError: can't concat %s to bytes", vm.typeName(b))
		}
	case OpBinaryMultiply:
		switch a.(type) {
		case *PyList, *PyTuple, *PyString, *PyBytes:
			if numericFamily(b) > 1 {
				return nil, fmt.Errorf("TypeError: can't multiply sequence by non-int of type '%s'", vm.typeName(b))
			}
		}
		switch b.(type) {
		case *PyList, *PyTuple, *PyString, *PyBytes:
			if numericFamily(a) > 1 {
				return nil, fmt.Errorf("TypeError: can't multiply sequence by non-int of type '%s'", vm.typeName(a))
			}
		}
	}
	return nil, fmt.Errorf("TypeError: unsupported operand type(s) for %s: '%s' and '%s'",
		binaryOpSymbol(op), vm.typeName(a), vm.typeName(b))
}

// binaryOpSymbol returns the Python operator for a binary opcode, as used in
// error messages.
func binaryOpSymbol(op Opcode) string {
	switch op {
	case OpBinaryAdd:
		return "+"
	case OpBinarySubtract:
		return "-"
	case OpBinaryMultiply:
		return "*"
	case OpBinaryDivide:
		return "/"
	case OpBinaryFloorDiv:
		return "//"
	case OpBinaryModulo:
		return "%"
	case OpBinaryPower:
		return "** or pow()"
	case OpBinaryMatMul:
		return "@"
	case OpBinaryAnd:
		return "&"
	case OpBinaryOr:
		return "|"
	case OpBinaryXor:
		return "^"
	case OpBinaryLShift:
		return "<<"
	case OpBinaryRShift:
		return ">>"
	}
	return op.String()
}
