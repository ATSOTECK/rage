package runtime

import (
	"fmt"
	"strings"
	"unicode"
)

// getAttrString handles attribute access on *PyString values.
func (vm *VM) getAttrString(str *PyString, name string) (Value, error) {
	if m := vm.strMethod(str, name); m != nil {
		return m, nil
	}
	switch name {
	case "format":
		return &PyBuiltinFunc{Name: "str.format", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return vm.strFormat(str.Value, args, kwargs)
		}}, nil

	case "isascii":
		return &PyBuiltinFunc{Name: "str.isascii", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			// Empty string returns True (CPython behavior)
			for _, r := range str.Value {
				if r > 127 {
					return False, nil
				}
			}
			return True, nil
		}}, nil

	case "isprintable":
		return &PyBuiltinFunc{Name: "str.isprintable", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			// Empty string returns True (CPython behavior)
			for _, r := range str.Value {
				if !unicode.IsPrint(r) {
					return False, nil
				}
			}
			return True, nil
		}}, nil

	case "istitle":
		return &PyBuiltinFunc{Name: "str.istitle", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(str.Value) == 0 {
				return False, nil
			}
			// A titlecased string has uppercase after uncased chars
			// and lowercase after cased chars
			prevCased := false
			hasCased := false
			for _, r := range str.Value {
				if unicode.IsUpper(r) || unicode.IsTitle(r) {
					if prevCased {
						return False, nil
					}
					prevCased = true
					hasCased = true
				} else if unicode.IsLower(r) {
					if !prevCased {
						return False, nil
					}
					prevCased = true
					hasCased = true
				} else {
					prevCased = false
				}
			}
			if !hasCased {
				return False, nil
			}
			return True, nil
		}}, nil

	case "removeprefix":
		return &PyBuiltinFunc{Name: "str.removeprefix", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: removeprefix() takes exactly one argument (%d given)", len(args))
			}
			prefix, ok := args[0].(*PyString)
			if !ok {
				return nil, fmt.Errorf("TypeError: removeprefix arg must be str, not '%s'", vm.typeName(args[0]))
			}
			if strings.HasPrefix(str.Value, prefix.Value) {
				return &PyString{Value: str.Value[len(prefix.Value):]}, nil
			}
			return str, nil
		}}, nil

	case "removesuffix":
		return &PyBuiltinFunc{Name: "str.removesuffix", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: removesuffix() takes exactly one argument (%d given)", len(args))
			}
			suffix, ok := args[0].(*PyString)
			if !ok {
				return nil, fmt.Errorf("TypeError: removesuffix arg must be str, not '%s'", vm.typeName(args[0]))
			}
			if suffix.Value != "" && strings.HasSuffix(str.Value, suffix.Value) {
				return &PyString{Value: str.Value[:len(str.Value)-len(suffix.Value)]}, nil
			}
			return str, nil
		}}, nil

	case "format_map":
		return &PyBuiltinFunc{Name: "str.format_map", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: format_map() takes exactly one argument (%d given)", len(args))
			}
			return vm.strFormatMap(str.Value, args[0])
		}}, nil

	case "maketrans":
		return &PyBuiltinFunc{Name: "str.maketrans", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			return strMaketransImpl(vm, args)
		}}, nil

	case "translate":
		return &PyBuiltinFunc{Name: "str.translate", Fn: func(args []Value, kwargs map[string]Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("TypeError: translate() takes exactly one argument (%d given)", len(args))
			}
			table, ok := args[0].(*PyDict)
			if !ok {
				return nil, fmt.Errorf("TypeError: translate() argument must be a dict")
			}
			var result strings.Builder
			for _, r := range str.Value {
				key := MakeInt(int64(r))
				if val, found := table.DictGet(key, vm); found {
					if val == None {
						// Delete the character
						continue
					}
					switch v := val.(type) {
					case *PyString:
						result.WriteString(v.Value)
					case *PyInt:
						result.WriteRune(rune(v.Value))
					default:
						result.WriteRune(r)
					}
				} else {
					result.WriteRune(r)
				}
			}
			return &PyString{Value: result.String()}, nil
		}}, nil
	}
	return nil, fmt.Errorf("AttributeError: 'str' object has no attribute '%s'", name)
}

// strMaketransImpl implements str.maketrans(x[, y[, z]]).
func strMaketransImpl(vm *VM, args []Value) (Value, error) {
	result := &PyDict{Items: make(map[Value]Value), buckets: make(map[uint64][]dictEntry)}
	if len(args) == 1 {
		d, ok := args[0].(*PyDict)
		if !ok {
			return nil, fmt.Errorf("TypeError: if you give only one argument to maketrans it must be a dict")
		}
		for _, key := range d.Keys(vm) {
			val, _ := d.DictGet(key, vm)
			var intKey Value
			switch k := key.(type) {
			case *PyString:
				runes := []rune(k.Value)
				if len(runes) != 1 {
					return nil, fmt.Errorf("ValueError: string keys in translate table must be of length 1, found %d", len(runes))
				}
				intKey = MakeInt(int64(runes[0]))
			case *PyInt:
				intKey = k
			default:
				return nil, fmt.Errorf("TypeError: keys in translate table must be strings or integers")
			}
			switch v := val.(type) {
			case *PyString:
				runes := []rune(v.Value)
				if len(runes) == 1 {
					result.DictSet(intKey, MakeInt(int64(runes[0])), vm)
				} else {
					result.DictSet(intKey, val, vm)
				}
			case *PyInt:
				result.DictSet(intKey, val, vm)
			case *PyNone:
				result.DictSet(intKey, None, vm)
			default:
				result.DictSet(intKey, val, vm)
			}
		}
		return result, nil
	}
	if len(args) < 2 {
		return nil, fmt.Errorf("TypeError: maketrans requires 1 or 2-3 string arguments")
	}
	x, ok1 := args[0].(*PyString)
	y, ok2 := args[1].(*PyString)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("TypeError: maketrans arguments must be strings")
	}
	xRunes := []rune(x.Value)
	yRunes := []rune(y.Value)
	if len(xRunes) != len(yRunes) {
		return nil, fmt.Errorf("ValueError: the first two maketrans arguments must have equal length")
	}
	for i, r := range xRunes {
		result.DictSet(MakeInt(int64(r)), MakeInt(int64(yRunes[i])), vm)
	}
	if len(args) >= 3 {
		z, ok := args[2].(*PyString)
		if !ok {
			return nil, fmt.Errorf("TypeError: maketrans third argument must be a string")
		}
		for _, r := range z.Value {
			result.DictSet(MakeInt(int64(r)), None, vm)
		}
	}
	return result, nil
}
