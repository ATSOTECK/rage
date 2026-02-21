package runtime

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// strFormat implements Python's str.format() method
func (vm *VM) strFormat(template string, args []Value, kwargs map[string]Value) (Value, error) {
	var result strings.Builder
	argIdx := 0

	i := 0
	for i < len(template) {
		if template[i] == '{' {
			if i+1 < len(template) && template[i+1] == '{' {
				result.WriteByte('{')
				i += 2
				continue
			}
			// Find closing brace
			j := i + 1
			for j < len(template) && template[j] != '}' {
				j++
			}
			if j >= len(template) {
				return nil, fmt.Errorf("ValueError: Single '{' encountered in format string")
			}
			field := template[i+1 : j]

			// Parse field name and format spec
			var fieldName, formatSpec string
			if colonIdx := strings.Index(field, ":"); colonIdx >= 0 {
				fieldName = field[:colonIdx]
				formatSpec = field[colonIdx+1:]
			} else {
				fieldName = field
			}

			// Get value
			var val Value
			if fieldName == "" {
				if argIdx >= len(args) {
					return nil, fmt.Errorf("IndexError: Replacement index %d out of range", argIdx)
				}
				val = args[argIdx]
				argIdx++
			} else if idx, err := strconv.Atoi(fieldName); err == nil {
				if idx >= len(args) {
					return nil, fmt.Errorf("IndexError: Replacement index %d out of range", idx)
				}
				val = args[idx]
			} else {
				if v, ok := kwargs[fieldName]; ok {
					val = v
				} else {
					return nil, fmt.Errorf("KeyError: '%s'", fieldName)
				}
			}

			// Apply format spec
			formatted, fmtErr := vm.formatValue(val, formatSpec)
			if fmtErr != nil {
				return nil, fmtErr
			}
			result.WriteString(formatted)
			i = j + 1
		} else if template[i] == '}' {
			if i+1 < len(template) && template[i+1] == '}' {
				result.WriteByte('}')
				i += 2
				continue
			}
			return nil, fmt.Errorf("ValueError: Single '}' encountered in format string")
		} else {
			result.WriteByte(template[i])
			i++
		}
	}
	return &PyString{Value: result.String()}, nil
}

// formatValue formats a value with the given format spec, checking __format__ first.
func (vm *VM) formatValue(val Value, spec string) (string, error) {
	if inst, ok := val.(*PyInstance); ok {
		if result, found, err := vm.callDunder(inst, "__format__", &PyString{Value: spec}); found {
			if err != nil {
				return "", err
			}
			if s, ok := result.(*PyString); ok {
				return s.Value, nil
			}
			return "", fmt.Errorf("TypeError: __format__ must return a str, not %s", vm.typeName(result))
		}
	}
	return vm.applyFormatSpec(val, spec), nil
}

// applyFormatSpec applies a format spec like ">10", "<10", "^10", ".2f", "05d"
func (vm *VM) applyFormatSpec(val Value, spec string) string {
	if len(spec) == 0 {
		return vm.str(val)
	}

	// Parse alignment and fill
	fill := " "
	align := byte(0)
	i := 0

	// Check for fill+align or just align
	if len(spec) > 1 && (spec[1] == '<' || spec[1] == '>' || spec[1] == '^') {
		fill = string(spec[0])
		align = spec[1]
		i = 2
	} else if len(spec) > 0 && (spec[0] == '<' || spec[0] == '>' || spec[0] == '^') {
		align = spec[0]
		i = 1
	}

	// Check for sign
	sign := byte(0)
	if i < len(spec) && (spec[i] == '+' || spec[i] == '-' || spec[i] == ' ') {
		sign = spec[i]
		i++
	}

	// Check for zero-fill
	zeroFill := false
	if i < len(spec) && spec[i] == '0' {
		zeroFill = true
		fill = "0"
		if align == 0 {
			align = '>'
		}
		i++
	}

	// Parse width
	width := 0
	for i < len(spec) && spec[i] >= '0' && spec[i] <= '9' {
		width = width*10 + int(spec[i]-'0')
		i++
	}

	// Parse precision
	precision := -1
	if i < len(spec) && spec[i] == '.' {
		i++
		precision = 0
		for i < len(spec) && spec[i] >= '0' && spec[i] <= '9' {
			precision = precision*10 + int(spec[i]-'0')
			i++
		}
	}

	// Parse type
	typeChar := byte(0)
	if i < len(spec) {
		typeChar = spec[i]
	}

	// Format the value
	var s string
	switch typeChar {
	case 'f', 'F':
		f := vm.toFloat(val)
		if precision < 0 {
			precision = 6
		}
		s = strconv.FormatFloat(f, 'f', precision, 64)
	case 'd':
		n := vm.toInt(val)
		s = strconv.FormatInt(n, 10)
	case 'x':
		n := vm.toInt(val)
		s = strconv.FormatInt(n, 16)
	case 'X':
		n := vm.toInt(val)
		s = strings.ToUpper(strconv.FormatInt(n, 16))
	case 'o':
		n := vm.toInt(val)
		s = strconv.FormatInt(n, 8)
	case 'b':
		n := vm.toInt(val)
		s = strconv.FormatInt(n, 2)
	case 'e', 'E':
		f := vm.toFloat(val)
		if precision < 0 {
			precision = 6
		}
		s = strconv.FormatFloat(f, byte(typeChar), precision, 64)
	case 'g', 'G':
		f := vm.toFloat(val)
		if precision < 0 {
			precision = 6
		}
		s = strconv.FormatFloat(f, byte(typeChar), precision, 64)
	case 's', 0:
		s = vm.str(val)
		if precision >= 0 && len(s) > precision {
			s = s[:precision]
		}
	default:
		s = vm.str(val)
	}

	// Apply sign for numeric types
	if sign == '+' && len(s) > 0 && s[0] != '-' {
		s = "+" + s
	} else if sign == ' ' && len(s) > 0 && s[0] != '-' {
		s = " " + s
	}

	// Apply zero-fill for numeric types
	if zeroFill && width > 0 {
		prefix := ""
		if len(s) > 0 && (s[0] == '-' || s[0] == '+' || s[0] == ' ') {
			prefix = string(s[0])
			s = s[1:]
		}
		for len(s) < width-len(prefix) {
			s = "0" + s
		}
		s = prefix + s
		return s
	}

	// Apply width and alignment
	if width > utf8.RuneCountInString(s) {
		padding := width - utf8.RuneCountInString(s)
		switch align {
		case '<':
			s = s + strings.Repeat(fill, padding)
		case '>':
			s = strings.Repeat(fill, padding) + s
		case '^':
			left := padding / 2
			right := padding - left
			s = strings.Repeat(fill, left) + s + strings.Repeat(fill, right)
		default:
			// Default: right-align for numbers, left-align for strings
			switch val.(type) {
			case *PyInt, *PyFloat:
				s = strings.Repeat(fill, padding) + s
			default:
				s = s + strings.Repeat(fill, padding)
			}
		}
	}

	return s
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
