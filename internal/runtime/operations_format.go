package runtime

import (
	"fmt"
	"strconv"
	"strings"
)

// stringFormat implements Python's "string % args" printf-style formatting
func (vm *VM) stringFormat(format string, args Value) (Value, error) {
	// Check if this is dict-based formatting: %(key)s
	dictArgs, isDict := args.(*PyDict)

	var argList []Value
	if !isDict {
		switch v := args.(type) {
		case *PyTuple:
			argList = v.Items
		default:
			argList = []Value{args}
		}
	}

	var result strings.Builder
	argIdx := 0
	i := 0
	for i < len(format) {
		if format[i] == '%' {
			i++
			if i >= len(format) {
				return nil, fmt.Errorf("TypeError: incomplete format")
			}
			if format[i] == '%' {
				result.WriteByte('%')
				i++
				continue
			}

			// Parse mapping key: %(name)s
			var arg Value
			mappingKey := ""
			hasMappingKey := false
			if format[i] == '(' {
				if !isDict {
					return nil, fmt.Errorf("TypeError: format requires a mapping")
				}
				i++
				start := i
				for i < len(format) && format[i] != ')' {
					i++
				}
				if i >= len(format) {
					return nil, fmt.Errorf("TypeError: incomplete format key")
				}
				mappingKey = format[start:i]
				hasMappingKey = true
				i++ // skip ')'
			}

			// Parse flags: #, 0, -, ' ', +
			leftAlign := false
			zeroPad := false
			signPlus := false
			signSpace := false
			altForm := false
			for i < len(format) {
				switch format[i] {
				case '-':
					leftAlign = true
					i++
				case '0':
					zeroPad = true
					i++
				case '+':
					signPlus = true
					i++
				case ' ':
					signSpace = true
					i++
				case '#':
					altForm = true
					i++
				default:
					goto flagsDone
				}
			}
		flagsDone:

			// Parse width (* means take from args)
			width := 0
			if i < len(format) && format[i] == '*' {
				i++
				if argIdx >= len(argList) {
					return nil, fmt.Errorf("TypeError: not enough arguments for format string")
				}
				width = int(vm.toInt(argList[argIdx]))
				argIdx++
				if width < 0 {
					leftAlign = true
					width = -width
				}
			} else {
				for i < len(format) && format[i] >= '0' && format[i] <= '9' {
					width = width*10 + int(format[i]-'0')
					i++
				}
			}

			// Parse precision
			precision := -1
			if i < len(format) && format[i] == '.' {
				i++
				precision = 0
				if i < len(format) && format[i] == '*' {
					i++
					if argIdx >= len(argList) {
						return nil, fmt.Errorf("TypeError: not enough arguments for format string")
					}
					precision = int(vm.toInt(argList[argIdx]))
					argIdx++
					if precision < 0 {
						precision = 0
					}
				} else {
					for i < len(format) && format[i] >= '0' && format[i] <= '9' {
						precision = precision*10 + int(format[i]-'0')
						i++
					}
				}
			}

			if i >= len(format) {
				return nil, fmt.Errorf("TypeError: incomplete format")
			}

			spec := format[i]
			i++

			// Get the argument
			if hasMappingKey {
				val, found := dictArgs.DictGet(&PyString{Value: mappingKey}, vm)
				if !found {
					return nil, fmt.Errorf("KeyError: '%s'", mappingKey)
				}
				arg = val
			} else {
				if argIdx >= len(argList) {
					return nil, fmt.Errorf("TypeError: not enough arguments for format string")
				}
				arg = argList[argIdx]
				argIdx++
			}

			var s string
			isNumeric := false
			switch spec {
			case 's':
				s = vm.str(arg)
				if precision >= 0 && len(s) > precision {
					s = s[:precision]
				}
			case 'r':
				s = vm.repr(arg)
				if precision >= 0 && len(s) > precision {
					s = s[:precision]
				}
			case 'a':
				s = vm.ascii(arg)
				if precision >= 0 && len(s) > precision {
					s = s[:precision]
				}
			case 'd', 'i', 'u':
				isNumeric = true
				n := vm.toInt(arg)
				s = fmt.Sprintf("%d", n)
			case 'f', 'F':
				isNumeric = true
				f := vm.toFloat(arg)
				if precision < 0 {
					precision = 6
				}
				s = strconv.FormatFloat(f, 'f', precision, 64)
				if spec == 'F' {
					s = strings.ToUpper(s)
				}
			case 'e':
				isNumeric = true
				f := vm.toFloat(arg)
				if precision < 0 {
					precision = 6
				}
				s = strconv.FormatFloat(f, 'e', precision, 64)
			case 'E':
				isNumeric = true
				f := vm.toFloat(arg)
				if precision < 0 {
					precision = 6
				}
				s = strconv.FormatFloat(f, 'E', precision, 64)
			case 'g':
				isNumeric = true
				f := vm.toFloat(arg)
				if precision < 0 {
					precision = 6
				}
				s = strconv.FormatFloat(f, 'g', precision, 64)
			case 'G':
				isNumeric = true
				f := vm.toFloat(arg)
				if precision < 0 {
					precision = 6
				}
				s = strconv.FormatFloat(f, 'G', precision, 64)
			case 'x':
				isNumeric = true
				n := vm.toInt(arg)
				if n < 0 {
					s = fmt.Sprintf("-%x", -n)
				} else {
					s = fmt.Sprintf("%x", n)
				}
				if altForm && n != 0 {
					s = "0x" + s
				}
			case 'X':
				isNumeric = true
				n := vm.toInt(arg)
				if n < 0 {
					s = fmt.Sprintf("-%X", -n)
				} else {
					s = fmt.Sprintf("%X", n)
				}
				if altForm && n != 0 {
					s = "0X" + s
				}
			case 'o':
				isNumeric = true
				n := vm.toInt(arg)
				if n < 0 {
					s = fmt.Sprintf("-%o", -n)
				} else {
					s = fmt.Sprintf("%o", n)
				}
				if altForm && n != 0 {
					s = "0o" + s
				}
			case 'c':
				switch v := arg.(type) {
				case *PyInt:
					s = string(rune(v.Value))
				case *PyString:
					if len(v.Value) != 1 {
						return nil, fmt.Errorf("%%c requires int or char")
					}
					s = v.Value
				default:
					return nil, fmt.Errorf("%%c requires int or char")
				}
			default:
				return nil, fmt.Errorf("TypeError: unsupported format character '%c' (0x%x)", spec, spec)
			}

			// Apply sign for numeric types
			if isNumeric && s != "" && s[0] != '-' {
				if signPlus {
					s = "+" + s
				} else if signSpace {
					s = " " + s
				}
			}

			// Apply width and alignment
			if width > 0 && len(s) < width {
				pad := width - len(s)
				if leftAlign {
					s = s + strings.Repeat(" ", pad)
				} else if zeroPad && isNumeric && !leftAlign {
					// Zero-pad after sign prefix
					prefix := ""
					num := s
					if len(s) > 0 && (s[0] == '-' || s[0] == '+' || s[0] == ' ') {
						prefix = s[:1]
						num = s[1:]
					}
					s = prefix + strings.Repeat("0", width-len(prefix)-len(num)) + num
				} else {
					s = strings.Repeat(" ", pad) + s
				}
			}

			result.WriteString(s)
		} else {
			result.WriteByte(format[i])
			i++
		}
	}

	// Check for too many arguments (only for positional, not dict)
	if !isDict && argIdx < len(argList) {
		return nil, fmt.Errorf("TypeError: not all arguments converted during string formatting")
	}

	return &PyString{Value: result.String()}, nil
}
