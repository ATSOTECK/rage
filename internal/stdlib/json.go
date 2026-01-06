package stdlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// InitJSONModule registers the json module
func InitJSONModule() {
	runtime.NewModuleBuilder("json").
		Doc("JSON encoder and decoder.").
		Func("dumps", jsonDumps).
		Func("loads", jsonLoads).
		Func("dump", jsonDump).
		Func("load", jsonLoad).
		Register()
}

// jsonDumps serializes a Python object to a JSON string.
// dumps(obj, *, indent=None, separators=None, sort_keys=False) -> str
func jsonDumps(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("dumps() missing required argument: 'obj'")
		return 0
	}

	obj := vm.Get(1)

	// Parse optional keyword arguments
	indent := ""
	itemSep := ", "
	keySep := ": "
	sortKeys := false

	// Check for indent argument (position 2)
	if vm.GetTop() >= 2 {
		indentVal := vm.Get(2)
		if !runtime.IsNone(indentVal) {
			switch v := indentVal.(type) {
			case *runtime.PyInt:
				if v.Value >= 0 {
					indent = strings.Repeat(" ", int(v.Value))
				}
			case *runtime.PyString:
				indent = v.Value
			}
		}
	}

	// Check for separators argument (position 3)
	if vm.GetTop() >= 3 {
		sepVal := vm.Get(3)
		if !runtime.IsNone(sepVal) {
			if tuple, ok := sepVal.(*runtime.PyTuple); ok && len(tuple.Items) == 2 {
				if s, ok := tuple.Items[0].(*runtime.PyString); ok {
					itemSep = s.Value
				}
				if s, ok := tuple.Items[1].(*runtime.PyString); ok {
					keySep = s.Value
				}
			} else if list, ok := sepVal.(*runtime.PyList); ok && len(list.Items) == 2 {
				if s, ok := list.Items[0].(*runtime.PyString); ok {
					itemSep = s.Value
				}
				if s, ok := list.Items[1].(*runtime.PyString); ok {
					keySep = s.Value
				}
			}
		}
	}

	// Check for sort_keys argument (position 4)
	if vm.GetTop() >= 4 {
		sortKeysVal := vm.Get(4)
		if !runtime.IsNone(sortKeysVal) {
			sortKeys = vm.ToBool(4)
		}
	}

	result, err := encodeJSON(obj, indent, itemSep, keySep, sortKeys, 0)
	if err != nil {
		vm.RaiseError("TypeError: Object of type '%s' is not JSON serializable", jsonTypeName(obj))
		return 0
	}

	vm.Push(runtime.NewString(result))
	return 1
}

// jsonLoads deserializes a JSON string to a Python object.
// loads(s) -> object
func jsonLoads(vm *runtime.VM) int {
	if vm.GetTop() < 1 {
		vm.RaiseError("loads() missing required argument: 's'")
		return 0
	}

	s := vm.CheckString(1)

	result, err := decodeJSON(s)
	if err != nil {
		vm.RaiseError("json.JSONDecodeError: %s", err.Error())
		return 0
	}

	vm.Push(result)
	return 1
}

// jsonDump writes JSON to a file object (requires file I/O support)
func jsonDump(vm *runtime.VM) int {
	vm.RaiseError("json.dump() is not yet implemented (requires file I/O)")
	return 0
}

// jsonLoad reads JSON from a file object (requires file I/O support)
func jsonLoad(vm *runtime.VM) int {
	vm.RaiseError("json.load() is not yet implemented (requires file I/O)")
	return 0
}

// encodeJSON converts a Python value to a JSON string
func encodeJSON(v runtime.Value, indent, itemSep, keySep string, sortKeys bool, depth int) (string, error) {
	switch val := v.(type) {
	case *runtime.PyNone:
		return "null", nil

	case *runtime.PyBool:
		if val.Value {
			return "true", nil
		}
		return "false", nil

	case *runtime.PyInt:
		return strconv.FormatInt(val.Value, 10), nil

	case *runtime.PyFloat:
		// Handle special float values
		if val.Value != val.Value { // NaN check
			return "", fmt.Errorf("NaN is not JSON serializable")
		}
		if val.Value > 1e308 || val.Value < -1e308 {
			return "", fmt.Errorf("Infinity is not JSON serializable")
		}
		return strconv.FormatFloat(val.Value, 'g', -1, 64), nil

	case *runtime.PyString:
		return encodeString(val.Value), nil

	case *runtime.PyBytes:
		return "", fmt.Errorf("bytes is not JSON serializable")

	case *runtime.PyList:
		return encodeArray(val.Items, indent, itemSep, keySep, sortKeys, depth)

	case *runtime.PyTuple:
		return encodeArray(val.Items, indent, itemSep, keySep, sortKeys, depth)

	case *runtime.PyDict:
		return encodeObject(val, indent, itemSep, keySep, sortKeys, depth)

	default:
		return "", fmt.Errorf("not JSON serializable")
	}
}

// encodeString encodes a string with proper JSON escaping
func encodeString(s string) string {
	var buf bytes.Buffer
	buf.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			if r < 0x20 {
				buf.WriteString(fmt.Sprintf(`\u%04x`, r))
			} else {
				buf.WriteRune(r)
			}
		}
	}
	buf.WriteByte('"')
	return buf.String()
}

// encodeArray encodes a list/tuple as a JSON array
func encodeArray(items []runtime.Value, indent, itemSep, keySep string, sortKeys bool, depth int) (string, error) {
	if len(items) == 0 {
		return "[]", nil
	}

	var buf bytes.Buffer
	buf.WriteByte('[')

	if indent != "" {
		buf.WriteByte('\n')
	}

	for i, item := range items {
		if i > 0 {
			buf.WriteString(strings.TrimRight(itemSep, " "))
			if indent != "" {
				buf.WriteByte('\n')
			}
		}

		if indent != "" {
			buf.WriteString(strings.Repeat(indent, depth+1))
		}

		encoded, err := encodeJSON(item, indent, itemSep, keySep, sortKeys, depth+1)
		if err != nil {
			return "", err
		}
		buf.WriteString(encoded)
	}

	if indent != "" {
		buf.WriteByte('\n')
		buf.WriteString(strings.Repeat(indent, depth))
	}

	buf.WriteByte(']')
	return buf.String(), nil
}

// encodeObject encodes a dict as a JSON object
func encodeObject(d *runtime.PyDict, indent, itemSep, keySep string, sortKeys bool, depth int) (string, error) {
	if len(d.Items) == 0 {
		return "{}", nil
	}

	// Collect and validate keys
	type keyValue struct {
		key   string
		value runtime.Value
	}
	pairs := make([]keyValue, 0, len(d.Items))

	for k, v := range d.Items {
		keyStr, err := keyToString(k)
		if err != nil {
			return "", err
		}
		pairs = append(pairs, keyValue{key: keyStr, value: v})
	}

	// Sort keys if requested
	if sortKeys {
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].key < pairs[j].key
		})
	}

	var buf bytes.Buffer
	buf.WriteByte('{')

	if indent != "" {
		buf.WriteByte('\n')
	}

	for i, pair := range pairs {
		if i > 0 {
			buf.WriteString(strings.TrimRight(itemSep, " "))
			if indent != "" {
				buf.WriteByte('\n')
			}
		}

		if indent != "" {
			buf.WriteString(strings.Repeat(indent, depth+1))
		}

		buf.WriteString(encodeString(pair.key))
		buf.WriteString(keySep)

		encoded, err := encodeJSON(pair.value, indent, itemSep, keySep, sortKeys, depth+1)
		if err != nil {
			return "", err
		}
		buf.WriteString(encoded)
	}

	if indent != "" {
		buf.WriteByte('\n')
		buf.WriteString(strings.Repeat(indent, depth))
	}

	buf.WriteByte('}')
	return buf.String(), nil
}

// keyToString converts a dict key to a string for JSON encoding
func keyToString(k runtime.Value) (string, error) {
	switch v := k.(type) {
	case *runtime.PyString:
		return v.Value, nil
	case *runtime.PyInt:
		return strconv.FormatInt(v.Value, 10), nil
	case *runtime.PyFloat:
		return strconv.FormatFloat(v.Value, 'g', -1, 64), nil
	case *runtime.PyBool:
		if v.Value {
			return "true", nil
		}
		return "false", nil
	case *runtime.PyNone:
		return "null", nil
	default:
		return "", fmt.Errorf("keys must be str, int, float, bool or None")
	}
}

// jsonTypeName returns the Python type name of a value for JSON error messages
func jsonTypeName(v runtime.Value) string {
	switch v.(type) {
	case *runtime.PyNone:
		return "NoneType"
	case *runtime.PyBool:
		return "bool"
	case *runtime.PyInt:
		return "int"
	case *runtime.PyFloat:
		return "float"
	case *runtime.PyString:
		return "str"
	case *runtime.PyBytes:
		return "bytes"
	case *runtime.PyList:
		return "list"
	case *runtime.PyTuple:
		return "tuple"
	case *runtime.PyDict:
		return "dict"
	case *runtime.PyFunction:
		return "function"
	case *runtime.PyClass:
		return "type"
	case *runtime.PyInstance:
		return "object"
	default:
		return "unknown"
	}
}

// decodeJSON parses a JSON string into a Python value
func decodeJSON(s string) (runtime.Value, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return nil, fmt.Errorf("Expecting value: line 1 column 1 (char 0)")
	}

	// Use Go's json package to decode, then convert to Python values
	var goVal any
	decoder := json.NewDecoder(strings.NewReader(s))
	decoder.UseNumber() // Preserve number precision
	if err := decoder.Decode(&goVal); err != nil {
		return nil, err
	}

	return goValueToPython(goVal)
}

// goValueToPython converts a Go value (from json.Unmarshal) to a Python value
func goValueToPython(v any) (runtime.Value, error) {
	if v == nil {
		return runtime.None, nil
	}

	switch val := v.(type) {
	case bool:
		return runtime.NewBool(val), nil

	case json.Number:
		// Try to parse as int first, then float
		if i, err := val.Int64(); err == nil {
			return runtime.NewInt(i), nil
		}
		if f, err := val.Float64(); err == nil {
			return runtime.NewFloat(f), nil
		}
		return nil, fmt.Errorf("invalid number: %s", val)

	case float64:
		// Check if it's actually an integer
		if val == float64(int64(val)) {
			return runtime.NewInt(int64(val)), nil
		}
		return runtime.NewFloat(val), nil

	case string:
		return runtime.NewString(val), nil

	case []any:
		items := make([]runtime.Value, len(val))
		for i, item := range val {
			pyVal, err := goValueToPython(item)
			if err != nil {
				return nil, err
			}
			items[i] = pyVal
		}
		return runtime.NewList(items), nil

	case map[string]any:
		d := runtime.NewDict()
		for k, item := range val {
			pyVal, err := goValueToPython(item)
			if err != nil {
				return nil, err
			}
			d.Items[runtime.NewString(k)] = pyVal
		}
		return d, nil

	default:
		return nil, fmt.Errorf("unsupported JSON type: %T", v)
	}
}
