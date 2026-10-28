package runtime

import (
	"fmt"
	"unicode/utf8"
)

// Structural pattern matching opcodes, shared by the main dispatch loop and
// the generator dispatcher.
//
// Stack contract: each MATCH_* test peeks at the subject on TOS, leaves it in
// place, and pushes a bool. On success some opcodes also push extracted values
// beneath the bool (MATCH_KEYS, MATCH_CLASS); on failure nothing but the bool
// is pushed. The compiler relies on the subject surviving both outcomes.

// opMatchSequence implements MATCH_SEQUENCE: is TOS a list/tuple of length
// arg? (arg 0xFFFF or -1 means any length, used with star patterns.)
func (vm *VM) opMatchSequence(arg int) {
	var length int
	isSequence := false
	switch s := vm.top().(type) {
	case *PyList:
		length = len(s.Items)
		isSequence = true
	case *PyTuple:
		length = len(s.Items)
		isSequence = true
	}
	anyLength := arg == 65535 || arg == -1
	if isSequence && (anyLength || length == arg) {
		vm.push(True)
	} else {
		vm.push(False)
	}
}

// opMatchStar implements MATCH_STAR: does the sequence on TOS have at least
// arg (= elements before + after the star) items?
func (vm *VM) opMatchStar(arg int) {
	var length int
	switch s := vm.top().(type) {
	case *PyList:
		length = len(s.Items)
	case *PyTuple:
		length = len(s.Items)
	default:
		vm.push(False)
		return
	}
	if length < arg {
		vm.push(False)
	} else {
		vm.push(True)
	}
}

// opExtractStar implements EXTRACT_STAR: pushes the list captured by a star
// pattern, items[before : len-after], with arg = before<<8 | after.
func (vm *VM) opExtractStar(arg int) {
	beforeStar := arg >> 8
	afterStar := arg & 0xFF

	var items []Value
	switch s := vm.top().(type) {
	case *PyList:
		items = s.Items
	case *PyTuple:
		items = s.Items
	default:
		vm.push(&PyList{Items: nil})
		return
	}

	end := len(items) - afterStar
	if end < beforeStar {
		end = beforeStar
	}
	slice := make([]Value, end-beforeStar)
	copy(slice, items[beforeStar:end])
	vm.push(&PyList{Items: slice})
}

// opMatchMapping implements MATCH_MAPPING: is TOS a dict?
func (vm *VM) opMatchMapping() {
	if _, ok := vm.top().(*PyDict); ok {
		vm.push(True)
	} else {
		vm.push(False)
	}
}

// opMatchKeys implements MATCH_KEYS.
// Stack: [subject, key1, ..., keyN, N] -> [subject, valueN, ..., value1, True]
// or [subject, False] if the subject lacks any key. value1 ends up on top
// (after the bool is consumed) so the first value pattern sees it first.
func (vm *VM) opMatchKeys() {
	keyCount := int(vm.pop().(*PyInt).Value)
	keys := make([]Value, keyCount)
	for i := keyCount - 1; i >= 0; i-- {
		keys[i] = vm.pop()
	}

	dict, ok := vm.top().(*PyDict)
	if !ok {
		vm.push(False)
		return
	}

	values := make([]Value, keyCount)
	for i, key := range keys {
		v, found := dict.DictGet(key, vm)
		if !found {
			vm.push(False)
			return
		}
		values[i] = v
	}
	for i := len(values) - 1; i >= 0; i-- {
		vm.push(values[i])
	}
	vm.push(True)
}

// opCopyDict implements COPY_DICT for `**rest` in mapping patterns.
// Stack: [subject, key1, ..., keyN, N] -> [subject, rest], where rest is a
// copy of the subject dict, in insertion order, without the matched keys.
func (vm *VM) opCopyDict() {
	keyCount := int(vm.pop().(*PyInt).Value)
	keysToRemove := make([]Value, keyCount)
	for i := keyCount - 1; i >= 0; i-- {
		keysToRemove[i] = vm.pop()
	}

	dict := vm.top().(*PyDict)
	newDict := &PyDict{Items: make(map[Value]Value)}
	for _, k := range dict.Keys(vm) {
		shouldRemove := false
		for _, removeKey := range keysToRemove {
			if vm.equal(k, removeKey) {
				shouldRemove = true
				break
			}
		}
		if !shouldRemove {
			if v, ok := dict.DictGet(k, vm); ok {
				newDict.DictSet(k, v, vm)
			}
		}
	}
	vm.push(newDict)
}

// opMatchClass implements MATCH_CLASS with arg positional sub-patterns.
// Stack: [subject, cls] -> [subject, attrN, ..., attr1, True] or
// [subject, False]. Positional attributes come from cls.__match_args__,
// except for self-matching builtins (int, str, ...) where a single positional
// pattern receives the subject itself.
func (vm *VM) opMatchClass(arg int) {
	cls := vm.pop()
	subject := vm.top()
	positionalCount := arg

	isInstance := false
	var matchedClass *PyClass
	if c, ok := cls.(*PyClass); ok {
		matchedClass = c
		if inst, ok := subject.(*PyInstance); ok {
			isInstance = vm.isInstanceOf(inst, c)
		}
	}
	if !isInstance {
		// Builtin types (int, str, list, ...) and non-instance subjects:
		// defer to isinstance() semantics.
		if isinstFn, ok := vm.builtins["isinstance"].(*PyBuiltinFunc); ok {
			if r, err := isinstFn.Fn([]Value{subject, cls}, nil); err == nil && vm.truthy(r) {
				isInstance = true
			}
		}
	}
	if !isInstance {
		vm.push(False)
		return
	}

	if positionalCount == 1 && vm.isSelfMatchingClass(cls) {
		vm.push(subject)
		vm.push(True)
		return
	}

	var matchArgs []string
	if matchedClass != nil && positionalCount > 0 {
		var items []Value
		for _, c := range matchedClass.Mro {
			if maVal, ok := c.Dict["__match_args__"]; ok {
				switch ma := maVal.(type) {
				case *PyTuple:
					items = ma.Items
				case *PyList:
					items = ma.Items
				}
				break
			}
		}
		for _, item := range items {
			if s, ok := item.(*PyString); ok {
				matchArgs = append(matchArgs, s.Value)
			}
		}
	}
	if positionalCount > len(matchArgs) {
		vm.push(False)
		return
	}

	// getattr semantics: properties and class attributes count
	attrs := make([]Value, positionalCount)
	for i := 0; i < positionalCount; i++ {
		val, err := vm.getAttr(subject, matchArgs[i])
		if err != nil {
			vm.push(False)
			return
		}
		attrs[i] = val
	}
	for i := len(attrs) - 1; i >= 0; i-- {
		vm.push(attrs[i])
	}
	vm.push(True)
}

// opGetLen implements GET_LEN: pushes len(TOS) without consuming it.
func (vm *VM) opGetLen() error {
	var length int64
	switch s := vm.top().(type) {
	case *PyList:
		length = int64(len(s.Items))
	case *PyTuple:
		length = int64(len(s.Items))
	case *PyString:
		length = int64(utf8.RuneCountInString(s.Value))
	case *PyDict:
		length = int64(s.DictLen())
	default:
		return fmt.Errorf("TypeError: object of type '%s' has no len()", vm.typeName(s))
	}
	vm.push(MakeInt(length))
	return nil
}
