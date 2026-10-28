package runtime

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// strFormatter implements str.format() and str.format_map() (PEP 3101).
type strFormatter struct {
	vm      *VM
	args    []Value
	kwargs  map[string]Value
	mapping Value // non-nil for format_map: named fields are looked up here

	autoIndex int
	numbering byte // 0 until the first field, then 'a' (automatic) or 'm' (manual)
}

// strFormat implements Python's str.format() method
func (vm *VM) strFormat(template string, args []Value, kwargs map[string]Value) (Value, error) {
	f := &strFormatter{vm: vm, args: args, kwargs: kwargs}
	s, err := f.format(template, 2)
	if err != nil {
		return nil, err
	}
	return &PyString{Value: s}, nil
}

// strFormatMap implements Python's str.format_map() method
func (vm *VM) strFormatMap(template string, mapping Value) (Value, error) {
	f := &strFormatter{vm: vm, mapping: mapping}
	s, err := f.format(template, 2)
	if err != nil {
		return nil, err
	}
	return &PyString{Value: s}, nil
}

// format expands template. depth limits how deeply format specs may nest
// replacement fields ("{x:{width}}").
func (f *strFormatter) format(template string, depth int) (string, error) {
	if depth <= 0 {
		return "", fmt.Errorf("ValueError: Max string recursion exceeded")
	}
	var b strings.Builder
	i, n := 0, len(template)
	for i < n {
		c := template[i]
		if c == '}' {
			if i+1 < n && template[i+1] == '}' {
				b.WriteByte('}')
				i += 2
				continue
			}
			return "", fmt.Errorf("ValueError: Single '}' encountered in format string")
		}
		if c != '{' {
			b.WriteByte(c)
			i++
			continue
		}
		if i+1 < n && template[i+1] == '{' {
			b.WriteByte('{')
			i += 2
			continue
		}
		if i+1 >= n {
			return "", fmt.Errorf("ValueError: Single '{' encountered in format string")
		}

		fieldName, conversion, spec, end, err := parseReplacementField(template, i+1)
		if err != nil {
			return "", err
		}
		i = end
		s, err := f.renderField(fieldName, conversion, spec, depth)
		if err != nil {
			return "", err
		}
		b.WriteString(s)
	}
	return b.String(), nil
}

// parseReplacementField parses the field starting at template[start] (just
// after the opening brace) and returns its parts and the index just past the
// closing brace.
func parseReplacementField(template string, start int) (fieldName string, conversion rune, spec string, end int, err error) {
	n := len(template)
	i := start
	var c byte
	terminated := false
	for i < n && !terminated {
		c = template[i]
		i++
		switch c {
		case '{':
			return "", 0, "", 0, fmt.Errorf("ValueError: unexpected '{' in field name")
		case '[':
			// An index may contain any character except ']'.
			for i < n && template[i] != ']' {
				i++
			}
		case '}', ':', '!':
			terminated = true
		}
	}
	if !terminated {
		return "", 0, "", 0, fmt.Errorf("ValueError: expected '}' before end of string")
	}
	fieldName = template[start : i-1]
	if c == '}' {
		return fieldName, 0, "", i, nil
	}
	if c == '!' {
		if i >= n {
			return "", 0, "", 0, fmt.Errorf("ValueError: end of string while looking for conversion specifier")
		}
		r, size := utf8.DecodeRuneInString(template[i:])
		conversion = r
		i += size
		if i < n {
			next := template[i]
			i++
			if next == '}' {
				return fieldName, conversion, "", i, nil
			}
			if next != ':' {
				return "", 0, "", 0, fmt.Errorf("ValueError: expected ':' after conversion specifier")
			}
		}
	}
	specStart := i
	count := 1
	for i < n {
		switch template[i] {
		case '{':
			count++
		case '}':
			count--
			if count == 0 {
				return fieldName, conversion, template[specStart:i], i + 1, nil
			}
		}
		i++
	}
	return "", 0, "", 0, fmt.Errorf("ValueError: unmatched '{' in format spec")
}

// renderField looks up, converts and formats a single replacement field.
func (f *strFormatter) renderField(fieldName string, conversion rune, spec string, depth int) (string, error) {
	val, err := f.lookupField(fieldName)
	if err != nil {
		return "", err
	}
	vm := f.vm
	switch conversion {
	case 0:
	case 'r':
		s, err := vm.reprValue(val)
		if err != nil {
			return "", err
		}
		val = &PyString{Value: s}
	case 's':
		s, err := vm.strValue(val)
		if err != nil {
			return "", err
		}
		val = &PyString{Value: s}
	case 'a':
		s, err := vm.asciiValue(val)
		if err != nil {
			return "", err
		}
		val = &PyString{Value: s}
	default:
		if conversion < utf8.RuneSelf && conversion > ' ' {
			return "", fmt.Errorf("ValueError: Unknown conversion specifier %c", conversion)
		}
		return "", fmt.Errorf("ValueError: Unknown conversion specifier \\x%x", conversion)
	}
	if strings.ContainsRune(spec, '{') {
		spec, err = f.format(spec, depth-1)
		if err != nil {
			return "", err
		}
	}
	return vm.formatValue(val, spec)
}

// lookupField resolves a field name such as "", "0", "name", "0.attr" or
// "name[key][0]" against the positional and keyword arguments.
func (f *strFormatter) lookupField(fieldName string) (Value, error) {
	vm := f.vm
	first := len(fieldName)
	if i := strings.IndexAny(fieldName, ".["); i >= 0 {
		first = i
	}
	argName, rest := fieldName[:first], fieldName[first:]

	var val Value
	if argName == "" || isDecimalDigits(argName) {
		var index int
		if argName == "" {
			if f.numbering == 'm' {
				return nil, fmt.Errorf("ValueError: cannot switch from manual field specification to automatic field numbering")
			}
			f.numbering = 'a'
			index = f.autoIndex
			f.autoIndex++
		} else {
			if f.numbering == 'a' {
				return nil, fmt.Errorf("ValueError: cannot switch from automatic field numbering to manual field specification")
			}
			f.numbering = 'm'
			n, err := strconv.Atoi(argName)
			if err != nil {
				return nil, fmt.Errorf("ValueError: Too many decimal digits in format string")
			}
			index = n
		}
		if f.mapping != nil {
			return nil, fmt.Errorf("ValueError: Format string contains positional fields")
		}
		if index >= len(f.args) {
			return nil, fmt.Errorf("IndexError: Replacement index %d out of range for positional args tuple", index)
		}
		val = f.args[index]
	} else if f.mapping != nil {
		v, err := vm.getItem(f.mapping, &PyString{Value: argName})
		if err != nil {
			return nil, err
		}
		val = v
	} else {
		v, ok := f.kwargs[argName]
		if !ok {
			return nil, vm.newKeyError(&PyString{Value: argName})
		}
		val = v
	}

	// Attribute and index accessors.
	for rest != "" {
		switch rest[0] {
		case '.':
			rest = rest[1:]
			end := len(rest)
			if i := strings.IndexAny(rest, ".["); i >= 0 {
				end = i
			}
			name := rest[:end]
			if name == "" {
				return nil, fmt.Errorf("ValueError: Empty attribute in format string")
			}
			rest = rest[end:]
			v, err := vm.getAttr(val, name)
			if err != nil {
				return nil, err
			}
			val = v
		case '[':
			end := strings.IndexByte(rest, ']')
			if end < 0 {
				return nil, fmt.Errorf("ValueError: Missing ']' in format string")
			}
			key := rest[1:end]
			if key == "" {
				return nil, fmt.Errorf("ValueError: Empty attribute in format string")
			}
			rest = rest[end+1:]
			if rest != "" && rest[0] != '.' && rest[0] != '[' {
				return nil, fmt.Errorf("ValueError: Only '.' or '[' may follow ']' in format field specifier")
			}
			var index Value = &PyString{Value: key}
			if isDecimalDigits(key) {
				if n, err := strconv.ParseInt(key, 10, 64); err == nil {
					index = MakeInt(n)
				}
			}
			v, err := vm.getItem(val, index)
			if err != nil {
				return nil, err
			}
			val = v
		default:
			return nil, fmt.Errorf("ValueError: Only '.' or '[' may follow ']' in format field specifier")
		}
	}
	return val, nil
}

func isDecimalDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
