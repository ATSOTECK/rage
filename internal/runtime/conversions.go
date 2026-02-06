package runtime

import (
	"fmt"
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
		return val.Value, nil
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
		// Try parsing as integer
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			// Try parsing as float then converting
			f, ferr := strconv.ParseFloat(s, 64)
			if ferr != nil {
				return 0, fmt.Errorf("ValueError: invalid literal for int() with base 10: %q", val.Value)
			}
			return int64(f), nil
		}
		return i, nil
	default:
		return 0, fmt.Errorf("TypeError: int() argument must be a string or a number, not '%s'", vm.typeName(v))
	}
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
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, fmt.Errorf("ValueError: could not convert string to float: %q", val.Value)
		}
		return f, nil
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
		var items []Value
		for k := range val.Items {
			items = append(items, k)
		}
		return items, nil
	case *PyIterator:
		return val.Items[val.Index:], nil
	default:
		return nil, fmt.Errorf("'%s' object is not iterable", vm.typeName(v))
	}
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
	case *PyBool:
		if val.Value {
			return "True"
		}
		return "False"
	case *PyInt:
		return fmt.Sprintf("%d", val.Value)
	case *PyFloat:
		return fmt.Sprintf("%g", val.Value)
	case *PyString:
		return val.Value
	case *PyBytes:
		return fmt.Sprintf("b'%s'", string(val.Value))
	case *PyList:
		return fmt.Sprintf("%v", val.Items)
	case *PyTuple:
		return fmt.Sprintf("%v", val.Items)
	case *PyDict:
		return fmt.Sprintf("%v", val.Items)
	case *PySet:
		return fmt.Sprintf("%v", val.Items)
	case *PyFrozenSet:
		if len(val.Items) == 0 {
			return "frozenset()"
		}
		return fmt.Sprintf("frozenset(%v)", val.Items)
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
	case *PyException:
		return vm.formatException(val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatExceptionInstance formats an exception instance for display
func (vm *VM) formatExceptionInstance(inst *PyInstance) string {
	className := inst.Class.Name

	// Get args from the instance
	if args, ok := inst.Dict["args"]; ok {
		if t, ok := args.(*PyTuple); ok && len(t.Items) > 0 {
			if len(t.Items) == 1 {
				// Single argument - just show the message
				return fmt.Sprintf("%s: %s", className, vm.str(t.Items[0]))
			}
			// Multiple arguments - show as tuple
			parts := make([]string, len(t.Items))
			for i, item := range t.Items {
				parts[i] = vm.str(item)
			}
			return fmt.Sprintf("%s: (%s)", className, strings.Join(parts, ", "))
		}
	}

	// No args, just show the class name
	return className
}

// formatException formats a PyException for display
func (vm *VM) formatException(exc *PyException) string {
	typeName := exc.Type()

	if exc.Args != nil && len(exc.Args.Items) > 0 {
		if len(exc.Args.Items) == 1 {
			return fmt.Sprintf("%s: %s", typeName, vm.str(exc.Args.Items[0]))
		}
		parts := make([]string, len(exc.Args.Items))
		for i, item := range exc.Args.Items {
			parts[i] = vm.str(item)
		}
		return fmt.Sprintf("%s: (%s)", typeName, strings.Join(parts, ", "))
	}

	if exc.Message != "" && exc.Message != typeName {
		return fmt.Sprintf("%s: %s", typeName, exc.Message)
	}

	return typeName
}

func (vm *VM) typeName(v Value) string {
	switch val := v.(type) {
	case *PyNone:
		return "NoneType"
	case *PyBool:
		return "bool"
	case *PyInt:
		return "int"
	case *PyFloat:
		return "float"
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
	default:
		return "object"
	}
}
