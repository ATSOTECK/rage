package runtime

import (
	"fmt"
	"math"
	"math/big"
	"strings"
	"unicode/utf8"
)

// stringFormat implements Python's "string % args" printf-style formatting
func (vm *VM) stringFormat(format string, args Value) (Value, error) {
	// A tuple supplies one argument per conversion; any other object is a
	// single argument. A dict also serves %(name)s lookups.
	var argList []Value
	tupleArgs := false
	if t, ok := args.(*PyTuple); ok {
		argList = t.Items
		tupleArgs = true
	} else {
		argList = []Value{args}
	}
	mapping, isMapping := args.(*PyDict)
	usedMapping := false

	argIdx := 0
	nextArg := func() (Value, error) {
		if (usedMapping && !tupleArgs) || argIdx >= len(argList) {
			return nil, fmt.Errorf("TypeError: not enough arguments for format string")
		}
		v := argList[argIdx]
		argIdx++
		return v, nil
	}

	var result strings.Builder
	i := 0
	for i < len(format) {
		if format[i] != '%' {
			result.WriteByte(format[i])
			i++
			continue
		}
		i++
		if i >= len(format) {
			return nil, fmt.Errorf("ValueError: incomplete format")
		}
		if format[i] == '%' {
			result.WriteByte('%')
			i++
			continue
		}

		// Mapping key: %(name)s
		var arg Value
		hasKey := false
		if format[i] == '(' {
			if !isMapping {
				return nil, fmt.Errorf("TypeError: format requires a mapping")
			}
			depth := 1
			i++
			start := i
			for i < len(format) && depth > 0 {
				switch format[i] {
				case '(':
					depth++
				case ')':
					depth--
				}
				i++
			}
			if depth > 0 {
				return nil, fmt.Errorf("ValueError: incomplete format key")
			}
			val, err := vm.getItem(mapping, &PyString{Value: format[start : i-1]})
			if err != nil {
				return nil, err
			}
			arg = val
			hasKey = true
			usedMapping = true
		}

		// Flags: -, +, space, #, 0
		leftAlign, zeroPad, signPlus, signSpace, altForm := false, false, false, false, false
	flags:
		for i < len(format) {
			switch format[i] {
			case '-':
				leftAlign = true
			case '0':
				zeroPad = true
			case '+':
				signPlus = true
			case ' ':
				signSpace = true
			case '#':
				altForm = true
			default:
				break flags
			}
			i++
		}

		// Width (* takes it from the arguments)
		width := 0
		if i < len(format) && format[i] == '*' {
			i++
			v, err := nextArg()
			if err != nil {
				return nil, err
			}
			n, ok := v.(*PyInt)
			if !ok || n.BigValue != nil {
				return nil, fmt.Errorf("TypeError: * wants int")
			}
			width = int(n.Value)
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

		// Precision
		precision := -1
		if i < len(format) && format[i] == '.' {
			i++
			precision = 0
			if i < len(format) && format[i] == '*' {
				i++
				v, err := nextArg()
				if err != nil {
					return nil, err
				}
				n, ok := v.(*PyInt)
				if !ok || n.BigValue != nil {
					return nil, fmt.Errorf("TypeError: * wants int")
				}
				precision = max(int(n.Value), 0)
			} else {
				for i < len(format) && format[i] >= '0' && format[i] <= '9' {
					precision = precision*10 + int(format[i]-'0')
					i++
				}
			}
		}

		// Length modifiers are accepted and ignored, as in CPython
		for i < len(format) && (format[i] == 'h' || format[i] == 'l' || format[i] == 'L') {
			i++
		}
		if i >= len(format) {
			return nil, fmt.Errorf("ValueError: incomplete format")
		}

		conv, size := utf8.DecodeRuneInString(format[i:])
		i += size
		if !hasKey {
			v, err := nextArg()
			if err != nil {
				return nil, err
			}
			arg = v
		}

		var s string
		numeric := false
		switch conv {
		case 's', 'r', 'a':
			var err error
			switch conv {
			case 's':
				s, err = vm.strValue(arg)
			case 'r':
				s, err = vm.reprValue(arg)
			default:
				s, err = vm.asciiValue(arg)
			}
			if err != nil {
				return nil, err
			}
			if precision >= 0 && utf8.RuneCountInString(s) > precision {
				s = string([]rune(s)[:precision])
			}
		case 'c':
			switch v := arg.(type) {
			case *PyInt:
				if v.BigValue != nil || v.Value < 0 || v.Value > 0x10ffff {
					return nil, fmt.Errorf("OverflowError: %%c arg not in range(0x110000)")
				}
				s = string(rune(v.Value))
			case *PyString:
				if utf8.RuneCountInString(v.Value) != 1 {
					return nil, fmt.Errorf("TypeError: %%c requires an int or a unicode character, not a string of length %d", utf8.RuneCountInString(v.Value))
				}
				s = v.Value
			default:
				return nil, fmt.Errorf("TypeError: %%c requires an int or a unicode character, not %s", vm.typeName(arg))
			}
		case 'd', 'i', 'u', 'x', 'X', 'o':
			numeric = true
			n, err := vm.printfInt(arg, byte(conv))
			if err != nil {
				return nil, err
			}
			base, prefix := 10, ""
			switch conv {
			case 'x':
				base, prefix = 16, "0x"
			case 'X':
				base, prefix = 16, "0X"
			case 'o':
				base, prefix = 8, "0o"
			}
			digits := new(big.Int).Abs(n).Text(base)
			if conv == 'X' {
				digits = strings.ToUpper(digits)
			}
			if precision > len(digits) {
				digits = strings.Repeat("0", precision-len(digits)) + digits
			}
			if !altForm {
				prefix = ""
			}
			s = printfSign(n.Sign() < 0, signPlus, signSpace) + prefix + digits
			if zeroPad && precision >= 0 {
				// C semantics: a precision disables zero padding for integers
				zeroPad = false
			}
		case 'e', 'E', 'f', 'F', 'g', 'G':
			numeric = true
			f, err := vm.printfFloat(arg)
			if err != nil {
				return nil, err
			}
			if precision < 0 {
				precision = 6
			}
			flags := 0
			if altForm {
				flags |= fmtFloatAlt
			}
			text := formatFloat(f, byte(conv), precision, flags)
			negative := strings.HasPrefix(text, "-")
			s = printfSign(negative, signPlus, signSpace) + strings.TrimPrefix(text, "-")
		default:
			return nil, fmt.Errorf("ValueError: unsupported format character %s (0x%x) at index %d",
				strRepr(string(conv)), conv, utf8.RuneCountInString(format[:i])-1)
		}

		// Apply width and alignment
		if n := utf8.RuneCountInString(s); width > n {
			pad := width - n
			switch {
			case leftAlign:
				s += strings.Repeat(" ", pad)
			case zeroPad && numeric:
				// Zero-pad after the sign and any 0x/0o prefix
				lead := 0
				if s != "" && (s[0] == '-' || s[0] == '+' || s[0] == ' ') {
					lead = 1
				}
				if altForm && len(s) >= lead+2 && s[lead] == '0' && strings.ContainsRune("xXo", rune(s[lead+1])) {
					lead += 2
				}
				s = s[:lead] + strings.Repeat("0", pad) + s[lead:]
			default:
				s = strings.Repeat(" ", pad) + s
			}
		}
		result.WriteString(s)
	}

	// Every positional argument must be consumed (a dict argument is exempt)
	if !isMapping && argIdx < len(argList) {
		return nil, fmt.Errorf("TypeError: not all arguments converted during string formatting")
	}

	return &PyString{Value: result.String()}, nil
}

func printfSign(negative, plus, space bool) string {
	switch {
	case negative:
		return "-"
	case plus:
		return "+"
	case space:
		return " "
	}
	return ""
}

// printfInt converts arg for an integer conversion. %d, %i and %u accept any
// real number (floats are truncated); %x, %X and %o require an integer.
func (vm *VM) printfInt(arg Value, conv byte) (*big.Int, error) {
	acceptsReal := conv == 'd' || conv == 'i' || conv == 'u'
	switch v := arg.(type) {
	case *PyBool:
		if v.Value {
			return big.NewInt(1), nil
		}
		return big.NewInt(0), nil
	case *PyInt:
		return v.BigIntValue(), nil
	case *PyFloat:
		if acceptsReal {
			if math.IsInf(v.Value, 0) {
				return nil, fmt.Errorf("OverflowError: cannot convert float infinity to integer")
			}
			if math.IsNaN(v.Value) {
				return nil, fmt.Errorf("ValueError: cannot convert float NaN to integer")
			}
			n, _ := big.NewFloat(math.Trunc(v.Value)).Int(nil)
			return n, nil
		}
	case *PyInstance:
		names := []string{"__index__"}
		if acceptsReal {
			names = append(names, "__int__")
		}
		for _, name := range names {
			if result, found, err := vm.callDunder(v, name); found {
				if err != nil {
					return nil, err
				}
				if n, ok := result.(*PyInt); ok {
					return n.BigIntValue(), nil
				}
				return nil, fmt.Errorf("TypeError: %s returned non-int (type %s)", name, vm.typeName(result))
			}
		}
	}
	if acceptsReal {
		return nil, fmt.Errorf("TypeError: %%%c format: a real number is required, not %s", conv, vm.typeName(arg))
	}
	return nil, fmt.Errorf("TypeError: %%%c format: an integer is required, not %s", conv, vm.typeName(arg))
}

// printfFloat converts arg for a floating-point conversion.
func (vm *VM) printfFloat(arg Value) (float64, error) {
	switch v := arg.(type) {
	case *PyFloat:
		return v.Value, nil
	case *PyInt:
		if v.BigValue != nil {
			f, acc := new(big.Float).SetInt(v.BigValue).Float64()
			if math.IsInf(f, 0) && acc != big.Exact {
				return 0, fmt.Errorf("OverflowError: int too large to convert to float")
			}
			return f, nil
		}
		return float64(v.Value), nil
	case *PyBool:
		if v.Value {
			return 1, nil
		}
		return 0, nil
	case *PyInstance:
		return vm.tryToFloat(v)
	}
	return 0, fmt.Errorf("TypeError: must be real number, not %s", vm.typeName(arg))
}
