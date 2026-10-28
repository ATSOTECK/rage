package runtime

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// maxIntStringDigits is the maximum number of digits allowed when converting
// a string to a big integer. Matches CPython 3.12+ default (sys.int_info.default_max_str_digits).
const maxIntStringDigits = 4300

// Type conversions

func (vm *VM) toValue(v any) Value {
	if v == nil {
		return None
	}
	switch val := v.(type) {
	case bool:
		if val {
			return True
		}
		return False
	case int:
		return MakeInt(int64(val))
	case int64:
		return MakeInt(val)
	case float64:
		return &PyFloat{Value: val}
	case string:
		return &PyString{Value: val}
	case []byte:
		return &PyBytes{Value: val}
	case []string:
		items := make([]Value, len(val))
		for i, s := range val {
			items[i] = &PyString{Value: s}
		}
		return &PyTuple{Items: items}
	case *big.Int:
		return MakeBigInt(val)
	case *CodeObject:
		return val
	case Value:
		return val
	default:
		return &PyString{Value: fmt.Sprintf("%v", v)}
	}
}

func (vm *VM) toInt(v Value) int64 {
	i, _ := vm.tryToInt(v)
	return i
}

// tryToInt converts a value to int64, returning an error if conversion fails.
// Use this for Python's int() builtin where ValueError should be raised on failure.
func (vm *VM) tryToInt(v Value) (int64, error) {
	switch val := v.(type) {
	case *PyInt:
		if val.BigValue != nil {
			// Big int doesn't fit in int64
			return 0, fmt.Errorf("OverflowError: Python int too large to convert to int64")
		}
		return val.Value, nil
	case *PyComplex:
		_ = val
		return 0, fmt.Errorf("TypeError: int() argument must be a string, a bytes-like object or a real number, not 'complex'")
	case *PyFloat:
		return int64(val.Value), nil
	case *PyBool:
		if val.Value {
			return 1, nil
		}
		return 0, nil
	case *PyString:
		s := strings.TrimSpace(val.Value)
		if s == "" {
			return 0, fmt.Errorf("ValueError: invalid literal for int() with base 10: %s", strRepr(val.Value))
		}
		// Remove underscores
		s = strings.ReplaceAll(s, "_", "")
		// Try parsing as integer
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("ValueError: invalid literal for int() with base 10: %s", strRepr(val.Value))
		}
		return i, nil
	case *PyInstance:
		// Check for __int__ method
		if result, found, err := vm.callDunder(val, "__int__"); found {
			if err != nil {
				return 0, err
			}
			if i, ok := result.(*PyInt); ok {
				if i.BigValue != nil {
					return 0, fmt.Errorf("OverflowError: Python int too large to convert to int64")
				}
				return i.Value, nil
			}
			return 0, fmt.Errorf("TypeError: __int__ returned non-int")
		}
		return 0, fmt.Errorf("TypeError: int() argument must be a string or a number, not '%s'", vm.typeName(v))
	default:
		return 0, fmt.Errorf("TypeError: int() argument must be a string or a number, not '%s'", vm.typeName(v))
	}
}

// tryToIntValue converts a value to a PyInt (possibly big), returning an error if conversion fails.
func (vm *VM) tryToIntValue(v Value) (Value, error) {
	switch val := v.(type) {
	case *PyInt:
		return val, nil
	case *PyString:
		s := strings.TrimSpace(val.Value)
		if s == "" {
			return nil, fmt.Errorf("ValueError: invalid literal for int() with base 10: %s", strRepr(val.Value))
		}
		s = strings.ReplaceAll(s, "_", "")
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			if len(s) > maxIntStringDigits {
				return nil, fmt.Errorf("ValueError: Exceeds the limit (%d digits) for integer string conversion: value has %d digits", maxIntStringDigits, len(s))
			}
			// Try big.Int for overflow
			bi := new(big.Int)
			_, ok := bi.SetString(s, 10)
			if !ok {
				return nil, fmt.Errorf("ValueError: invalid literal for int() with base 10: %s", strRepr(val.Value))
			}
			return MakeBigInt(bi), nil
		}
		return MakeInt(i), nil
	default:
		i, err := vm.tryToInt(v)
		if err != nil {
			return nil, err
		}
		return MakeInt(i), nil
	}
}

// getIntIndex gets an integer index value, supporting __index__ protocol
func (vm *VM) getIntIndex(v Value) (int64, error) {
	switch val := v.(type) {
	case *PyInt:
		if val.BigValue != nil {
			return 0, fmt.Errorf("OverflowError: Python int too large to convert to index")
		}
		return val.Value, nil
	case *PyBool:
		if val.Value {
			return 1, nil
		}
		return 0, nil
	case *PyInstance:
		if result, found, err := vm.callDunder(val, "__index__"); found {
			if err != nil {
				return 0, err
			}
			if i, ok := result.(*PyInt); ok {
				return i.Value, nil
			}
		}
		return 0, fmt.Errorf("TypeError: '%s' object cannot be interpreted as an integer", vm.typeName(v))
	default:
		return 0, fmt.Errorf("TypeError: '%s' object cannot be interpreted as an integer", vm.typeName(v))
	}
}

// intFromStringBase converts a string to int with a given base
func (vm *VM) intFromStringBase(s string, base int64) (Value, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("ValueError: invalid literal for int() with base %d: ''", base)
	}

	// Validate base range
	if base != 0 && (base < 2 || base > 36) {
		return nil, fmt.Errorf("ValueError: int() base must be >= 2 and <= 36, or 0")
	}

	// Remove underscores
	s = strings.ReplaceAll(s, "_", "")

	// Handle sign
	negative := false
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		negative = s[0] == '-'
		s = s[1:]
		if s == "" {
			return nil, fmt.Errorf("ValueError: invalid literal for int() with base %d: %s", base, strRepr(s))
		}
	}

	// Handle base 0 (auto-detect)
	if base == 0 {
		if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
			base = 16
			s = s[2:]
		} else if strings.HasPrefix(s, "0o") || strings.HasPrefix(s, "0O") {
			base = 8
			s = s[2:]
		} else if strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B") {
			base = 2
			s = s[2:]
		} else if len(s) > 1 && s[0] == '0' {
			// Leading zeros in base 0 are not allowed (except "0" itself)
			allZeros := true
			for _, c := range s {
				if c != '0' {
					allZeros = false
					break
				}
			}
			if !allZeros {
				return nil, fmt.Errorf("ValueError: invalid literal for int() with base 0: '0%s'", s[1:])
			}
			base = 10
		} else {
			base = 10
		}
	} else {
		// Strip prefix if it matches the base
		if base == 16 && (strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X")) {
			s = s[2:]
		} else if base == 8 && (strings.HasPrefix(s, "0o") || strings.HasPrefix(s, "0O")) {
			s = s[2:]
		} else if base == 2 && (strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B")) {
			s = s[2:]
		}
	}

	if s == "" {
		return nil, fmt.Errorf("ValueError: invalid literal for int() with base %d: ''", base)
	}

	i, err := strconv.ParseInt(s, int(base), 64)
	if err != nil {
		if len(s) > maxIntStringDigits {
			return nil, fmt.Errorf("ValueError: Exceeds the limit (%d digits) for integer string conversion: value has %d digits", maxIntStringDigits, len(s))
		}
		// Try big.Int for overflow
		bi := new(big.Int)
		_, ok := bi.SetString(s, int(base))
		if !ok {
			return nil, fmt.Errorf("ValueError: invalid literal for int() with base %d: %s", base, strRepr(s))
		}
		if negative {
			bi.Neg(bi)
		}
		return MakeBigInt(bi), nil
	}
	if negative {
		i = -i
	}
	return MakeInt(i), nil
}

func (vm *VM) toFloat(v Value) float64 {
	f, _ := vm.tryToFloat(v)
	return f
}

// tryToFloat converts a value to float64, returning an error if conversion fails.
// Use this for Python's float() builtin where ValueError should be raised on failure.
func (vm *VM) tryToFloat(v Value) (float64, error) {
	switch val := v.(type) {
	case *PyInt:
		if val.BigValue != nil {
			f, _ := new(big.Float).SetInt(val.BigValue).Float64()
			return f, nil
		}
		return float64(val.Value), nil
	case *PyComplex:
		_ = val
		return 0, fmt.Errorf("TypeError: float() argument must be a string or a real number, not 'complex'")
	case *PyFloat:
		return val.Value, nil
	case *PyBool:
		if val.Value {
			return 1.0, nil
		}
		return 0.0, nil
	case *PyString:
		s := strings.TrimSpace(val.Value)
		if s == "" {
			return 0, fmt.Errorf("ValueError: could not convert string to float: %s", strRepr(val.Value))
		}
		// Handle Python-style special values (case-insensitive)
		lower := strings.ToLower(s)
		switch lower {
		case "inf", "+inf", "infinity", "+infinity":
			return math.Inf(1), nil
		case "-inf", "-infinity":
			return math.Inf(-1), nil
		case "nan", "+nan", "-nan":
			return math.NaN(), nil
		}
		// Remove underscores (Python numeric literal syntax)
		s = strings.ReplaceAll(s, "_", "")
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			// ParseFloat returns ±Inf for overflow (ErrRange), which Python accepts
			if math.IsInf(f, 0) {
				return f, nil
			}
			return 0, fmt.Errorf("ValueError: could not convert string to float: %s", strRepr(val.Value))
		}
		return f, nil
	case *PyInstance:
		// Check for __float__ method
		if result, found, err := vm.callDunder(val, "__float__"); found {
			if err != nil {
				return 0, err
			}
			if f, ok := result.(*PyFloat); ok {
				return f.Value, nil
			}
			return 0, fmt.Errorf("TypeError: __float__ returned non-float (type %s)", vm.typeName(result))
		}
		return 0, fmt.Errorf("TypeError: float() argument must be a string or a number, not '%s'", vm.typeName(v))
	default:
		return 0, fmt.Errorf("TypeError: float() argument must be a string or a number, not '%s'", vm.typeName(v))
	}
}

func (vm *VM) toList(v Value) ([]Value, error) {
	switch val := v.(type) {
	case *PyList:
		items := make([]Value, len(val.Items))
		copy(items, val.Items)
		return items, nil
	case *PyTuple:
		items := make([]Value, len(val.Items))
		copy(items, val.Items)
		return items, nil
	case *PyString:
		runes := []rune(val.Value)
		items := make([]Value, len(runes))
		for i, ch := range runes {
			items[i] = &PyString{Value: string(ch)}
		}
		return items, nil
	case *PyBytes:
		items := make([]Value, len(val.Value))
		for i, b := range val.Value {
			items[i] = MakeInt(int64(b))
		}
		return items, nil
	case *PyRange:
		var items []Value
		for i := val.Start; (val.Step > 0 && i < val.Stop) || (val.Step < 0 && i > val.Stop); i += val.Step {
			items = append(items, MakeInt(i))
		}
		return items, nil
	case *PySet:
		var items []Value
		for k := range val.Items {
			items = append(items, k)
		}
		return items, nil
	case *PyFrozenSet:
		var items []Value
		for k := range val.Items {
			items = append(items, k)
		}
		return items, nil
	case *PyDict:
		keys := val.Keys(vm)
		items := make([]Value, len(keys))
		copy(items, keys)
		return items, nil
	case *PyIterator:
		items := val.Items
		if val.Source != nil {
			items = val.Source.Items
		}
		return items[val.Index:], nil
	case *PyGenerator:
		var items []Value
		for {
			value, done, err := vm.GeneratorSend(val, None)
			if done || err != nil {
				if err != nil {
					if pyExc, ok := err.(*PyException); ok && pyExc.Type() == "StopIteration" {
						vm.currentException = nil // Clear so it doesn't propagate
						break
					}
					return nil, err
				}
				break
			}
			items = append(items, value)
		}
		return items, nil
	case *PyClass:
		// Check for __iter__ in class dict or MRO
		if iterMethod, err := vm.getAttr(val, "__iter__"); err == nil {
			iterResult, err := vm.call(iterMethod, nil, nil)
			if err != nil {
				return nil, err
			}
			return vm.toList(iterResult)
		}
		return nil, fmt.Errorf("'%s' object is not iterable", vm.typeName(v))
	case *PyInstance:
		// Check for __iter__ method
		if iterResult, found, err := vm.callDunder(val, "__iter__"); found {
			if err != nil {
				return nil, err
			}
			// The result of __iter__ should be an iterator with __next__
			return vm.iteratorToList(iterResult)
		}
		return nil, fmt.Errorf("'%s' object is not iterable", vm.typeName(v))
	default:
		return nil, fmt.Errorf("'%s' object is not iterable", vm.typeName(v))
	}
}

// iteratorToList collects all items from an iterator (object with __next__) into a list
func (vm *VM) iteratorToList(iterator Value) ([]Value, error) {
	inst, ok := iterator.(*PyInstance)
	if !ok {
		// If __iter__ returned a known iterable type, just toList it
		return vm.toList(iterator)
	}
	var items []Value
	for {
		val, found, err := vm.callDunder(inst, "__next__")
		if !found {
			return nil, fmt.Errorf("iterator has no __next__ method")
		}
		if err != nil {
			// StopIteration means we're done
			if pyExc, ok := err.(*PyException); ok && pyExc.Type() == "StopIteration" {
				vm.currentException = nil // Clear so it doesn't propagate
				break
			}
			return nil, err
		}
		items = append(items, val)
	}
	return items, nil
}

func (vm *VM) truthy(v Value) bool {
	switch val := v.(type) {
	case *PyNone:
		return false
	case *PyBool:
		return val.Value
	case *PyInt:
		return val.Value != 0
	case *PyFloat:
		return val.Value != 0.0
	case *PyComplex:
		return val.Real != 0 || val.Imag != 0
	case *PyString:
		return len(val.Value) > 0
	case *PyList:
		return len(val.Items) > 0
	case *PyTuple:
		return len(val.Items) > 0
	case *PyDict:
		return len(val.Items) > 0
	case *PySet:
		return len(val.Items) > 0
	case *PyFrozenSet:
		return len(val.Items) > 0
	case *PyRange:
		return rangeLen(val) > 0
	case *PyBytes:
		return len(val.Value) > 0
	case *PyInstance:
		// Check __bool__ first
		if result, found, err := vm.callDunder(val, "__bool__"); found {
			if err != nil {
				// Propagate exceptions from __bool__ via currentException
				vm.currentException = vm.wrapGoError(err)
				return false
			}
			if b, ok := result.(*PyBool); ok {
				return b.Value
			}
		}
		// Fall back to __len__ (truthy if non-zero)
		if result, found, err := vm.callDunder(val, "__len__"); found && err == nil {
			if i, ok := result.(*PyInt); ok {
				return i.Value != 0
			}
		}
		// Default to true for instances
		return true
	default:
		return true
	}
}

// str returns str(v) for internal callers that cannot propagate errors (for
// example when building error messages). Python-visible conversions use
// strValue, which reports exceptions raised by __str__/__repr__.
func (vm *VM) str(v Value) string {
	s, err := vm.strValue(v)
	if err != nil {
		return vm.textFallback(v)
	}
	return s
}

// repr returns repr(v) for internal callers that cannot propagate errors.
// Python-visible conversions use reprValue.
func (vm *VM) repr(v Value) string {
	s, err := vm.reprValue(v)
	if err != nil {
		return vm.textFallback(v)
	}
	return s
}

// ascii returns ascii(v) for internal callers that cannot propagate errors.
func (vm *VM) ascii(v Value) string {
	return asciiEscape(vm.repr(v))
}

// textFallback is the text used when a user-defined __repr__/__str__ fails
// and the caller has no way to report the error.
func (vm *VM) textFallback(v Value) string {
	return fmt.Sprintf("<%s object>", vm.typeName(v))
}

func (vm *VM) typeName(v Value) string {
	switch val := v.(type) {
	case *PyNone:
		return "NoneType"
	case *PyNotImplementedType:
		_ = val
		return "NotImplementedType"
	case *PyBool:
		return "bool"
	case *PyInt:
		return "int"
	case *PyFloat:
		return "float"
	case *PyComplex:
		return "complex"
	case *PyString:
		return "str"
	case *PyBytes:
		return "bytes"
	case *PyList:
		return "list"
	case *PyTuple:
		return "tuple"
	case *PyDict:
		return "dict"
	case *PySet:
		return "set"
	case *PyFrozenSet:
		return "frozenset"
	case *PyFunction:
		return "function"
	case *PyBuiltinFunc:
		return "builtin_function_or_method"
	case *PyGoFunc:
		return "builtin_function_or_method"
	case *PyClass:
		return "type"
	case *PyInstance:
		return val.Class.Name
	case *PyException:
		return val.Type()
	case *PyEllipsisType:
		return "ellipsis"
	case *PyRange:
		return "range"
	case *PyIterator:
		return "iterator"
	case *PyUserData:
		if val.Metatable != nil {
			// Find __type__ key in metatable (iterate because Value keys use pointers)
			for k, v := range val.Metatable.Items {
				if ks, ok := k.(*PyString); ok && ks.Value == "__type__" {
					return vm.str(v)
				}
			}
		}
		return "userdata"
	case *PyModule:
		return "module"
	case *GenericAlias:
		return "GenericAlias"
	case *UnionType:
		return "types.UnionType"
	default:
		return "object"
	}
}
