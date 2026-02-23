package runtime

import (
	"math/big"
	"strings"
)

// areBuiltinOrderable checks if two values are built-in types that support ordering.
func (vm *VM) areBuiltinOrderable(a, b Value) bool {
	switch a.(type) {
	case *PyInt, *PyFloat:
		switch b.(type) {
		case *PyInt, *PyFloat:
			return true
		}
	case *PyString:
		_, ok := b.(*PyString)
		return ok
	case *PyList:
		_, ok := b.(*PyList)
		return ok
	case *PyTuple:
		_, ok := b.(*PyTuple)
		return ok
	case *PyBytes:
		_, ok := b.(*PyBytes)
		return ok
	case *PyBool:
		switch b.(type) {
		case *PyBool, *PyInt, *PyFloat:
			return true
		}
	}
	return false
}

// tryRichCompare attempts dunder-based comparison on PyInstance objects.
// It tries a.__dunder__(b) first, then b.__reflected__(a).
// Returns (result, true) if a dunder was found, or (nil, false) to fall back.
func (vm *VM) tryRichCompare(a, b Value, dunder, reflected string) (Value, bool) {
	if inst, ok := a.(*PyInstance); ok {
		result, found, err := vm.callDunder(inst, dunder, b)
		if found && err != nil {
			// Propagate exceptions from dunder methods (e.g. __lt__ raising RuntimeError)
			if pyExc, ok := err.(*PyException); ok {
				vm.currentException = pyExc
			} else {
				vm.currentException = &PyException{TypeName: "RuntimeError", Message: err.Error()}
			}
			return nil, true
		}
		if found && result != NotImplemented {
			return result, true
		}
	}
	if inst, ok := b.(*PyInstance); ok {
		result, found, err := vm.callDunder(inst, reflected, a)
		if found && err != nil {
			if pyExc, ok := err.(*PyException); ok {
				vm.currentException = pyExc
			} else {
				vm.currentException = &PyException{TypeName: "RuntimeError", Message: err.Error()}
			}
			return nil, true
		}
		if found && result != NotImplemented {
			return result, true
		}
	}
	return nil, false
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
		// CPython __ne__ dispatch:
		// When b is a subclass of a, try b first (subclass priority).
		// 1. Try __ne__ on both sides (with subclass priority)
		// 2. If not found, try default __ne__ via __eq__
		// 3. Fall back to identity
		aInst, aIsInst := a.(*PyInstance)
		bInst, bIsInst := b.(*PyInstance)

		// Check if b's class is a subclass of a's class (give b priority)
		bHasPriority := false
		if aIsInst && bIsInst && aInst.Class != bInst.Class {
			for _, base := range bInst.Class.Mro {
				if base == aInst.Class {
					bHasPriority = true
					break
				}
			}
		}

		tried := false
		if bHasPriority && bIsInst {
			// Try b.__ne__(a) first (subclass priority)
			result, found, err := vm.callDunder(bInst, "__ne__", a)
			if found && err == nil && result != NotImplemented {
				return result
			}
			if found {
				tried = true
			}
		}

		if aIsInst {
			result, found, err := vm.callDunder(aInst, "__ne__", b)
			if found && err == nil && result != NotImplemented {
				return result
			}
			if !found {
				// Default __ne__: invert __eq__
				eqResult, eqFound, eqErr := vm.callDunder(aInst, "__eq__", b)
				if eqFound && eqErr == nil && eqResult != NotImplemented {
					if vm.truthy(eqResult) {
						return False
					}
					return True
				}
				if eqFound {
					tried = true
				}
			} else {
				tried = true
			}
		}

		if !bHasPriority && bIsInst {
			result, found, err := vm.callDunder(bInst, "__ne__", a)
			if found && err == nil && result != NotImplemented {
				return result
			}
			if found {
				tried = true
			}
		}

		// If we tried dunders and all returned NotImplemented, use identity
		if tried {
			if a != b {
				return True
			}
			return False
		}

		if !vm.equal(a, b) {
			return True
		}
		return False
	case OpCompareLt:
		if result, ok := vm.tryRichCompare(a, b, "__lt__", "__gt__"); ok {
			return result
		}
		if aIsComplex || bIsComplex || !vm.areBuiltinOrderable(a, b) {
			vm.currentException = &PyException{TypeName: "TypeError", Message: "'<' not supported between instances of '" + vm.typeName(a) + "' and '" + vm.typeName(b) + "'"}
			return nil
		}
		cmp := vm.compare(a, b)
		if vm.currentException != nil {
			return nil
		}
		if cmp < 0 {
			return True
		}
		return False
	case OpCompareLe:
		if result, ok := vm.tryRichCompare(a, b, "__le__", "__ge__"); ok {
			return result
		}
		if aIsComplex || bIsComplex || !vm.areBuiltinOrderable(a, b) {
			vm.currentException = &PyException{TypeName: "TypeError", Message: "'<=' not supported between instances of '" + vm.typeName(a) + "' and '" + vm.typeName(b) + "'"}
			return nil
		}
		cmp := vm.compare(a, b)
		if vm.currentException != nil {
			return nil
		}
		if cmp <= 0 {
			return True
		}
		return False
	case OpCompareGt:
		if result, ok := vm.tryRichCompare(a, b, "__gt__", "__lt__"); ok {
			return result
		}
		if aIsComplex || bIsComplex || !vm.areBuiltinOrderable(a, b) {
			vm.currentException = &PyException{TypeName: "TypeError", Message: "'>' not supported between instances of '" + vm.typeName(a) + "' and '" + vm.typeName(b) + "'"}
			return nil
		}
		cmp := vm.compare(a, b)
		if vm.currentException != nil {
			return nil
		}
		if cmp > 0 {
			return True
		}
		return False
	case OpCompareGe:
		if result, ok := vm.tryRichCompare(a, b, "__ge__", "__le__"); ok {
			return result
		}
		if aIsComplex || bIsComplex || !vm.areBuiltinOrderable(a, b) {
			vm.currentException = &PyException{TypeName: "TypeError", Message: "'>=' not supported between instances of '" + vm.typeName(a) + "' and '" + vm.typeName(b) + "'"}
			return nil
		}
		cmp := vm.compare(a, b)
		if vm.currentException != nil {
			return nil
		}
		if cmp >= 0 {
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
	case *UnionType:
		if bv, ok := b.(*UnionType); ok {
			if len(av.Args) != len(bv.Args) {
				return false
			}
			for i, arg := range av.Args {
				if !vm.equalWithCycleDetection(arg, bv.Args[i], seen) {
					return false
				}
			}
			return true
		}
		return false
	case *PyClass:
		// Classes are compared by identity
		return a == b
	case *PyInstance:
		// Check for __eq__ method
		if result, found, err := vm.callDunder(av, "__eq__", b); found && err == nil {
			if result != NotImplemented {
				return vm.truthy(result)
			}
			// __eq__ returned NotImplemented, try reflected on b
			if bv, ok := b.(*PyInstance); ok {
				if result2, found2, err2 := vm.callDunder(bv, "__eq__", a); found2 && err2 == nil {
					if result2 != NotImplemented {
						return vm.truthy(result2)
					}
				}
			}
			// Both returned NotImplemented, fall back to identity
			return a == b
		}
		// No __eq__ found, try b's __eq__
		if bv, ok := b.(*PyInstance); ok {
			if result, found, err := vm.callDunder(bv, "__eq__", a); found && err == nil {
				if result != NotImplemented {
					return vm.truthy(result)
				}
			}
		}
		// Fall back to identity comparison
		return a == b
	}
	// Check if b is a PyInstance with __eq__
	if bv, ok := b.(*PyInstance); ok {
		if result, found, err := vm.callDunder(bv, "__eq__", a); found && err == nil {
			if result != NotImplemented {
				return vm.truthy(result)
			}
		}
	}
	// Fall back to identity comparison for unhandled types
	return a == b
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
				if vm.currentException != nil {
					return 0
				}
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
				if vm.currentException != nil {
					return 0
				}
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
	// If we reach here with different, non-orderable types, signal a TypeError
	// (identity-equal values like None==None are fine as 0/equal)
	if a != b && !vm.areBuiltinOrderable(a, b) {
		_, aIsInst := a.(*PyInstance)
		_, bIsInst := b.(*PyInstance)
		if !aIsInst && !bIsInst {
			vm.currentException = &PyException{
				TypeName: "TypeError",
				Message:  "'<' not supported between instances of '" + vm.typeName(a) + "' and '" + vm.typeName(b) + "'",
			}
		}
	}
	return 0
}

// containsIdentityOrEqual checks identity first, then equality (CPython behavior for 'in')
func (vm *VM) containsIdentityOrEqual(v, item Value) bool {
	// Identity check first (same pointer)
	if v == item {
		return true
	}
	return vm.equal(v, item)
}

func (vm *VM) contains(container, item Value) bool {
	switch c := container.(type) {
	case *PyString:
		if _, ok := item.(*PyString); ok {
			// Use strings.Contains for optimized O(n+m) substring search
			return strings.Contains(c.Value, item.(*PyString).Value)
		}
		// Non-string item in string raises TypeError
		vm.currentException = &PyException{
			TypeName: "TypeError",
			Message:  "'in <string>' requires string as left operand, not " + vm.typeName(item),
		}
		return false
	case *PyList:
		for _, v := range c.Items {
			if vm.containsIdentityOrEqual(v, item) {
				return true
			}
		}
	case *PyTuple:
		for _, v := range c.Items {
			if vm.containsIdentityOrEqual(v, item) {
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
	case *PyClass:
		// Check for __contains__ in class Dict or MRO
		if method, ok := c.Dict["__contains__"]; ok {
			var result Value
			var err error
			switch fn := method.(type) {
			case *PyFunction:
				result, err = vm.callFunction(fn, []Value{c, item}, nil)
			case *PyBuiltinFunc:
				result, err = fn.Fn([]Value{c, item}, nil)
			}
			if err != nil {
				vm.currentException = &PyException{
					TypeName: "TypeError",
					Message:  err.Error(),
				}
				return false
			}
			if result != nil {
				if boolVal, ok := result.(*PyBool); ok {
					return boolVal.Value
				}
				return vm.truthy(result)
			}
		}
		// Fallback to __iter__-based iteration
		if iterMethod, ok := c.Dict["__iter__"]; ok {
			_ = iterMethod
			if iter, err := vm.getIter(c); err == nil {
				for {
					val, done, err := vm.iterNext(iter)
					if done || err != nil {
						break
					}
					if vm.containsIdentityOrEqual(val, item) {
						return true
					}
				}
				vm.currentException = nil
				return false
			}
		}
	case *PyInstance:
		// Check for __contains__ method in MRO
		containsFound := false
		for _, cls := range c.Class.Mro {
			if cls.Name == "object" {
				continue // skip object base
			}
			if method, ok := cls.Dict["__contains__"]; ok {
				// If __contains__ is explicitly None, raise TypeError
				if _, isNone := method.(*PyNone); isNone {
					vm.currentException = &PyException{
						TypeName: "TypeError",
						Message:  "argument of type '" + c.Class.Name + "' is not iterable",
					}
					return false
				}
				// Call __contains__
				allArgs := []Value{c, item}
				var result Value
				var err error
				switch fn := method.(type) {
				case *PyFunction:
					result, err = vm.callFunction(fn, allArgs, nil)
				case *PyBuiltinFunc:
					result, err = fn.Fn(allArgs, nil)
				}
				if err != nil {
					vm.currentException = &PyException{
						TypeName: "TypeError",
						Message:  err.Error(),
					}
					return false
				}
				if result != nil {
					if boolVal, ok := result.(*PyBool); ok {
						return boolVal.Value
					}
					return vm.truthy(result)
				}
				containsFound = true
				break
			}
		}

		if !containsFound {
			// Fall back to iterating via __iter__
			if iter, err := vm.getIter(c); err == nil {
				for {
					val, done, err := vm.iterNext(iter)
					if done || err != nil {
						break
					}
					if vm.containsIdentityOrEqual(val, item) {
						return true
					}
				}
				// Clear any stale currentException (e.g. StopIteration from iterNext)
				vm.currentException = nil
				return false
			}

			// Fall back to __getitem__ with sequential integer indices
			hasGetitem := false
			for _, cls := range c.Class.Mro {
				if _, ok := cls.Dict["__getitem__"]; ok {
					hasGetitem = true
					break
				}
			}
			if hasGetitem {
				for idx := 0; ; idx++ {
					result, _, err := vm.callDunder(c, "__getitem__", MakeInt(int64(idx)))
					if err != nil {
						// IndexError means we've exhausted the sequence
						if pyExc, ok := err.(*PyException); ok && pyExc.Type() == "IndexError" {
							vm.currentException = nil // Clear stale exception
							return false
						}
						errStr := err.Error()
						if strings.Contains(errStr, "IndexError") {
							vm.currentException = nil // Clear stale exception
							return false
						}
						// Other errors propagate
						vm.currentException = &PyException{
							TypeName: "TypeError",
							Message:  errStr,
						}
						return false
					}
					if vm.containsIdentityOrEqual(result, item) {
						return true
					}
				}
			}

			// No __contains__, __iter__, or __getitem__ - raise TypeError
			vm.currentException = &PyException{
				TypeName: "TypeError",
				Message:  "argument of type '" + c.Class.Name + "' is not iterable",
			}
			return false
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
