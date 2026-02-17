package runtime

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

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
			return 0, fmt.Errorf("ValueError: invalid literal for int() with base 10: %q", val.Value)
		}
		// Remove underscores
		s = strings.ReplaceAll(s, "_", "")
		// Try parsing as integer
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("ValueError: invalid literal for int() with base 10: %q", val.Value)
		}
		return i, nil
	case *PyInstance:
		// Check for __int__ method
		if result, found, err := vm.callDunder(val, "__int__"); found {
			if err != nil {
				return 0, err
			}
			if i, ok := result.(*PyInt); ok {
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
			return nil, fmt.Errorf("ValueError: invalid literal for int() with base 10: %q", val.Value)
		}
		s = strings.ReplaceAll(s, "_", "")
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			// Try big.Int for overflow
			bi := new(big.Int)
			_, ok := bi.SetString(s, 10)
			if !ok {
				return nil, fmt.Errorf("ValueError: invalid literal for int() with base 10: %q", val.Value)
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
			return nil, fmt.Errorf("ValueError: invalid literal for int() with base %d: %q", base, s)
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
		// Try big.Int for overflow
		bi := new(big.Int)
		_, ok := bi.SetString(s, int(base))
		if !ok {
			return nil, fmt.Errorf("ValueError: invalid literal for int() with base %d: %q", base, s)
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
			return 0, fmt.Errorf("ValueError: could not convert string to float: %q", val.Value)
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
			return 0, fmt.Errorf("ValueError: could not convert string to float: %q", val.Value)
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
		return val.Items, nil
	case *PyTuple:
		return val.Items, nil
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
		return val.Items[val.Index:], nil
	case *PyGenerator:
		var items []Value
		for {
			value, done, err := vm.GeneratorSend(val, None)
			if done || err != nil {
				if err != nil {
					if pyExc, ok := err.(*PyException); ok && pyExc.Type() == "StopIteration" {
						break
					}
					return nil, err
				}
				break
			}
			items = append(items, value)
		}
		return items, nil
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
		if result, found, err := vm.callDunder(val, "__bool__"); found && err == nil {
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

func (vm *VM) str(v Value) string {
	switch val := v.(type) {
	case *PyNone:
		return "None"
	case *PyNotImplementedType:
		_ = val
		return "NotImplemented"
	case *PyBool:
		if val.Value {
			return "True"
		}
		return "False"
	case *PyInt:
		return fmt.Sprintf("%d", val.Value)
	case *PyFloat:
		s := strconv.FormatFloat(val.Value, 'g', -1, 64)
		// Python always shows at least one decimal for floats
		if !strings.ContainsAny(s, ".eEn") {
			s += ".0"
		}
		return s
	case *PyComplex:
		return formatComplex(val.Real, val.Imag)
	case *PyString:
		return val.Value
	case *PyBytes:
		return bytesRepr(val.Value)
	case *PyList:
		parts := make([]string, len(val.Items))
		for i, item := range val.Items {
			parts[i] = vm.repr(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *PyTuple:
		parts := make([]string, len(val.Items))
		for i, item := range val.Items {
			parts[i] = vm.repr(item)
		}
		if len(parts) == 1 {
			return "(" + parts[0] + ",)"
		}
		return "(" + strings.Join(parts, ", ") + ")"
	case *PyDict:
		orderedKeys := val.Keys(vm)
		parts := make([]string, 0, len(orderedKeys))
		for _, k := range orderedKeys {
			if v, ok := val.DictGet(k, vm); ok {
				parts = append(parts, vm.repr(k)+": "+vm.repr(v))
			}
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *PySet:
		if len(val.Items) == 0 {
			return "set()"
		}
		parts := make([]string, 0, len(val.Items))
		for k := range val.Items {
			parts = append(parts, vm.repr(k))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *PyFrozenSet:
		if len(val.Items) == 0 {
			return "frozenset()"
		}
		parts := make([]string, 0, len(val.Items))
		for k := range val.Items {
			parts = append(parts, vm.repr(k))
		}
		return "frozenset({" + strings.Join(parts, ", ") + "})"
	case *PyFunction:
		return fmt.Sprintf("<function %s>", val.Name)
	case *PyBuiltinFunc:
		return fmt.Sprintf("<built-in function %s>", val.Name)
	case *PyGoFunc:
		return fmt.Sprintf("<go function %s>", val.Name)
	case *PyUserData:
		return fmt.Sprintf("<userdata %T>", val.Value)
	case *PyModule:
		return fmt.Sprintf("<module '%s'>", val.Name)
	case *PyInstance:
		// Check if this is an exception instance
		if vm.isExceptionClass(val.Class) {
			return vm.formatExceptionInstance(val)
		}
		// Check for __str__ method via MRO
		if result, found, err := vm.callDunder(val, "__str__"); found && err == nil {
			if s, ok := result.(*PyString); ok {
				return s.Value
			}
		}
		// Check for __repr__ method via MRO
		if result, found, err := vm.callDunder(val, "__repr__"); found && err == nil {
			if s, ok := result.(*PyString); ok {
				return s.Value
			}
		}
		return fmt.Sprintf("<%s object>", val.Class.Name)
	case *PyClass:
		return fmt.Sprintf("<class '%s'>", val.Name)
	case *GenericAlias:
		return val.formatRepr()
	case *PyException:
		return vm.formatException(val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatExceptionInstance formats an exception instance for str().
// In CPython, str(e) returns just the message, not "Type: message".
func (vm *VM) formatExceptionInstance(inst *PyInstance) string {
	// Get args from the instance
	if args, ok := inst.Dict["args"]; ok {
		if t, ok := args.(*PyTuple); ok && len(t.Items) > 0 {
			if len(t.Items) == 1 {
				return vm.str(t.Items[0])
			}
			// Multiple arguments - show as tuple
			parts := make([]string, len(t.Items))
			for i, item := range t.Items {
				parts[i] = vm.repr(item)
			}
			return fmt.Sprintf("(%s)", strings.Join(parts, ", "))
		}
	}

	// No args
	return ""
}

// formatException formats a PyException for str().
// In CPython, str(e) returns just the message, not "Type: message".
func (vm *VM) formatException(exc *PyException) string {
	if exc.Args != nil && len(exc.Args.Items) > 0 {
		if len(exc.Args.Items) == 1 {
			return vm.str(exc.Args.Items[0])
		}
		parts := make([]string, len(exc.Args.Items))
		for i, item := range exc.Args.Items {
			parts[i] = vm.repr(item)
		}
		return fmt.Sprintf("(%s)", strings.Join(parts, ", "))
	}

	if exc.Message != "" {
		return exc.Message
	}

	return ""
}

// bytesRepr produces the Python repr for a bytes object
func bytesRepr(data []byte) string {
	var b strings.Builder
	b.WriteString("b'")
	for _, c := range data {
		switch {
		case c == '\\':
			b.WriteString("\\\\")
		case c == '\'':
			b.WriteString("\\'")
		case c == '\t':
			b.WriteString("\\t")
		case c == '\n':
			b.WriteString("\\n")
		case c == '\r':
			b.WriteString("\\r")
		case c >= 32 && c < 127:
			b.WriteByte(c)
		default:
			fmt.Fprintf(&b, "\\x%02x", c)
		}
	}
	b.WriteByte('\'')
	return b.String()
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
	default:
		return "object"
	}
}

func (vm *VM) repr(v Value) string {
	switch val := v.(type) {
	case *PyComplex:
		return formatComplex(val.Real, val.Imag)
	case *PyString:
		return fmt.Sprintf("'%s'", val.Value)
	case *PyBytes:
		return bytesRepr(val.Value)
	case *PyNone:
		return "None"
	case *PyBool:
		if val.Value {
			return "True"
		}
		return "False"
	case *PyList:
		parts := make([]string, len(val.Items))
		for i, item := range val.Items {
			parts[i] = vm.repr(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *PyTuple:
		parts := make([]string, len(val.Items))
		for i, item := range val.Items {
			parts[i] = vm.repr(item)
		}
		if len(parts) == 1 {
			return "(" + parts[0] + ",)"
		}
		return "(" + strings.Join(parts, ", ") + ")"
	case *PyDict:
		orderedKeys := val.Keys(vm)
		parts := make([]string, 0, len(orderedKeys))
		for _, k := range orderedKeys {
			if v, ok := val.DictGet(k, vm); ok {
				parts = append(parts, vm.repr(k)+": "+vm.repr(v))
			}
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *PySet:
		if len(val.Items) == 0 {
			return "set()"
		}
		parts := make([]string, 0, len(val.Items))
		for k := range val.Items {
			parts = append(parts, vm.repr(k))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *PyInstance:
		if result, found, err := vm.callDunder(val, "__repr__"); found && err == nil {
			if s, ok := result.(*PyString); ok {
				return s.Value
			}
		}
		return fmt.Sprintf("<%s object>", val.Class.Name)
	case *GenericAlias:
		return val.formatRepr()
	default:
		return vm.str(v)
	}
}
