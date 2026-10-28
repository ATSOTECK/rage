package runtime

import "fmt"

// Generic special methods for built-in values, so that e.g. d.__setitem__,
// lst.__iadd__ or getattr(s, "__contains__") return callables that behave like
// the corresponding operator.

// numericFamily ranks numeric types: a binary method of a numeric type only
// accepts operands of its own rank or lower (int.__add__(1.5) is
// NotImplemented, float.__add__(1) is not). Non-numbers get 0.
func numericFamily(v Value) int {
	switch v.(type) {
	case *PyBool, *PyInt:
		return 1
	case *PyFloat:
		return 2
	case *PyComplex:
		return 3
	}
	return 0
}

// sameBuiltinKind reports whether a comparison dunder of self should handle
// other rather than return NotImplemented.
func sameBuiltinKind(self, other Value) bool {
	if fs, fo := numericFamily(self), numericFamily(other); fs > 0 || fo > 0 {
		return fs > 0 && fo > 0 && fo <= fs
	}
	if asSetTable(self) != nil {
		return asSetTable(other) != nil
	}
	return builtinValueTypeName(self) == builtinValueTypeName(other)
}

// binaryDunderOps maps binary special method names to operators.
var binaryDunderOps = map[string]Opcode{
	"__add__": OpBinaryAdd, "__sub__": OpBinarySubtract, "__mul__": OpBinaryMultiply,
	"__truediv__": OpBinaryDivide, "__floordiv__": OpBinaryFloorDiv, "__mod__": OpBinaryModulo,
	"__pow__": OpBinaryPower, "__and__": OpBinaryAnd, "__or__": OpBinaryOr,
	"__xor__": OpBinaryXor, "__lshift__": OpBinaryLShift, "__rshift__": OpBinaryRShift,
}

// compareDunderOps maps comparison special method names to operators.
var compareDunderOps = map[string]Opcode{
	"__eq__": OpCompareEq, "__ne__": OpCompareNe, "__lt__": OpCompareLt,
	"__le__": OpCompareLe, "__gt__": OpCompareGt, "__ge__": OpCompareGe,
}

// builtinDunder returns the special method name of a built-in value bound to
// it, or nil if the value's type has no such method. Unhashable types return
// None for __hash__, as in CPython.
func (vm *VM) builtinDunder(obj Value, name string) Value {
	typeName := builtinValueTypeName(obj)
	if typeName == "" {
		return nil
	}
	isSeq := typeName == "list" || typeName == "tuple" || typeName == "str" || typeName == "bytes" || typeName == "range"
	isMapping := typeName == "dict"
	isSet := typeName == "set" || typeName == "frozenset"
	isContainer := isSeq || isMapping || isSet
	isNum := numericFamily(obj) > 0
	mutable := typeName == "list" || isMapping

	bind := func(fn func(args []Value) (Value, error)) Value {
		return &PyBuiltinFunc{Name: typeName + "." + name, Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return fn(args)
		}}
	}
	nargs := func(args []Value, n int) error {
		if len(args) != n {
			return fmt.Errorf("TypeError: %s.%s() takes exactly %d argument(s) (%d given)", typeName, name, n, len(args))
		}
		return nil
	}
	boolValue := func(b bool) Value {
		if b {
			return True
		}
		return False
	}

	switch name {
	case "__len__":
		if !isContainer {
			return nil
		}
		return bind(func(args []Value) (Value, error) {
			n, err := vm.lenOf(obj)
			if err != nil {
				return nil, err
			}
			return MakeInt(n), nil
		})
	case "__iter__":
		if !isContainer {
			return nil
		}
		return bind(func(args []Value) (Value, error) {
			// Not vm.getIter: it looks up __iter__ and would recurse
			if lst, ok := obj.(*PyList); ok {
				return &PyIterator{Source: lst}, nil
			}
			items, err := vm.toList(obj)
			if err != nil {
				return nil, err
			}
			return &PyIterator{Items: items}, nil
		})
	case "__contains__":
		if !isContainer {
			return nil
		}
		return bind(func(args []Value) (Value, error) {
			if err := nargs(args, 1); err != nil {
				return nil, err
			}
			found := vm.contains(obj, args[0])
			if exc := vm.currentException; exc != nil {
				vm.currentException = nil
				return nil, exc
			}
			return boolValue(found), nil
		})
	case "__getitem__":
		if !isSeq && !isMapping {
			return nil
		}
		return bind(func(args []Value) (Value, error) {
			if err := nargs(args, 1); err != nil {
				return nil, err
			}
			return vm.getItem(obj, args[0])
		})
	case "__setitem__":
		if !mutable {
			return nil
		}
		return bind(func(args []Value) (Value, error) {
			if err := nargs(args, 2); err != nil {
				return nil, err
			}
			return None, vm.setItem(obj, args[0], args[1])
		})
	case "__delitem__":
		if !mutable {
			return nil
		}
		return bind(func(args []Value) (Value, error) {
			if err := nargs(args, 1); err != nil {
				return nil, err
			}
			return None, vm.delItem(obj, args[0])
		})
	case "__iadd__", "__imul__":
		if typeName != "list" {
			return nil
		}
		op := OpInplaceAdd
		if name == "__imul__" {
			op = OpInplaceMultiply
		}
		return bind(func(args []Value) (Value, error) {
			if err := nargs(args, 1); err != nil {
				return nil, err
			}
			if op == OpInplaceMultiply {
				lst := obj.(*PyList)
				r, err := vm.binaryOp(OpBinaryMultiply, lst, args[0])
				if err != nil {
					return nil, err
				}
				lst.Items = r.(*PyList).Items
				return lst, nil
			}
			return vm.inplaceOp(op, obj, args[0])
		})
	case "__ior__":
		if !isMapping {
			return nil
		}
		return bind(func(args []Value) (Value, error) {
			if err := nargs(args, 1); err != nil {
				return nil, err
			}
			return vm.inplaceOp(OpInplaceOr, obj, args[0])
		})
	case "__reversed__":
		if typeName != "list" && typeName != "range" && !isMapping {
			return nil
		}
		return bind(func(args []Value) (Value, error) {
			items, err := vm.toList(obj)
			if err != nil {
				return nil, err
			}
			rev := make([]Value, len(items))
			for i, item := range items {
				rev[len(items)-1-i] = item
			}
			return &PyIterator{Items: rev}, nil
		})
	case "__hash__":
		if mutable || typeName == "set" {
			return None
		}
		return bind(func(args []Value) (Value, error) {
			return MakeInt(int64(vm.hashValueVM(obj))), nil
		})
	case "__str__":
		return bind(func(args []Value) (Value, error) {
			return &PyString{Value: vm.str(obj)}, nil
		})
	case "__repr__":
		return bind(func(args []Value) (Value, error) {
			return &PyString{Value: vm.repr(obj)}, nil
		})
	case "__neg__", "__pos__", "__abs__", "__invert__":
		if !isNum || (name == "__invert__" && numericFamily(obj) != 1) {
			return nil
		}
		return bind(func(args []Value) (Value, error) {
			switch name {
			case "__neg__":
				return vm.unaryOp(OpUnaryNegative, obj)
			case "__pos__":
				return vm.unaryOp(OpUnaryPositive, obj)
			case "__invert__":
				return vm.unaryOp(OpUnaryInvert, obj)
			}
			return vm.call(vm.builtins["abs"], []Value{obj}, nil)
		})
	}

	if op, ok := compareDunderOps[name]; ok {
		return bind(func(args []Value) (Value, error) {
			if err := nargs(args, 1); err != nil {
				return nil, err
			}
			if !sameBuiltinKind(obj, args[0]) {
				return NotImplemented, nil
			}
			r := vm.compareOp(op, obj, args[0])
			if exc := vm.currentException; exc != nil {
				vm.currentException = nil
				return nil, exc
			}
			return r, nil
		})
	}

	// Binary arithmetic, including reflected forms (__radd__ etc.)
	reflected := len(name) > 3 && name[:3] == "__r" && name != "__rshift__" && name != "__repr__"
	base := name
	if reflected {
		base = "__" + name[3:]
	}
	op, ok := binaryDunderOps[base]
	if !ok {
		return nil
	}
	switch {
	case isNum:
		if numericFamily(obj) != 1 && (op == OpBinaryAnd || op == OpBinaryOr || op == OpBinaryXor || op == OpBinaryLShift || op == OpBinaryRShift) {
			return nil
		}
	case isSet:
		if op != OpBinaryAnd && op != OpBinaryOr && op != OpBinaryXor && op != OpBinarySubtract {
			return nil
		}
	case isMapping:
		if op != OpBinaryOr {
			return nil
		}
	case typeName == "str" || typeName == "bytes":
		if op != OpBinaryAdd && op != OpBinaryMultiply && op != OpBinaryModulo {
			return nil
		}
		if reflected && op != OpBinaryMultiply {
			return nil
		}
	case typeName == "list" || typeName == "tuple":
		if (op != OpBinaryAdd && op != OpBinaryMultiply) || (reflected && op != OpBinaryMultiply) {
			return nil
		}
	default:
		return nil
	}
	return bind(func(args []Value) (Value, error) {
		if err := nargs(args, 1); err != nil {
			return nil, err
		}
		other := args[0]
		if isNum && (numericFamily(other) == 0 || numericFamily(other) > numericFamily(obj)) {
			return NotImplemented, nil
		}
		if reflected {
			return vm.binaryOp(op, other, obj)
		}
		return vm.binaryOp(op, obj, other)
	})
}
